package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/concurrency"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

// Orchestrator wires the agent client to storage and publishes SSE progress.
// It is the only place that calls agents.Client — handlers never call agents
// directly, so every stage transition goes through the same supersede/state
// logic (spec §4: "do not collapse [the pipeline] into one prompt").
type Orchestrator struct {
	Agents       *agents.Client
	Store        *storage.Store
	Projects     *mysql.ProjectsRepo
	StageOutputs *mysql.StageOutputsRepo
	ProjectRefs  *mysql.ProjectReferencesRepo
	Templates    *mysql.TemplatesRepo
	Uploads      *mysql.UploadsRepo
	BrandAssets  *mysql.BrandAssetsRepo
	Bus          *Bus
	Logger       *slog.Logger
	Locks        *concurrency.Registry
}

// CreateProject implements the start of Flow A/B/C (spec §3): create the
// project, analyse any attached reference images (Flow C), resolve the style
// profile (from the selected template or the freshly-analysed uploads), then
// run Intake and the Creative Director back to back.
func (o *Orchestrator) CreateProject(ctx context.Context, userID, inputText string, templateID *string, uploadIDs []string, logoUploadID *string, footerText *string) (*mysql.Project, error) {
	if !o.Agents.Enabled() {
		return nil, agents.ErrAgentsUnavailable
	}

	var logoObjectKey *string
	if logoUploadID != nil && *logoUploadID != "" {
		up, err := o.Uploads.GetByID(ctx, *logoUploadID)
		if err != nil {
			return nil, fmt.Errorf("load logo upload %s: %w", *logoUploadID, err)
		}
		logoObjectKey = &up.ObjectKey
	}

	project, err := o.Projects.Create(ctx, userID, inputText, templateID, logoObjectKey, footerText)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	style, err := o.analyzeUploadsIfAny(ctx, project.ID, uploadIDs)
	if err != nil {
		return nil, fmt.Errorf("analyze reference uploads: %w", err)
	}
	if style == nil {
		style, err = o.resolveTemplateStyle(ctx, templateID)
		if err != nil {
			return nil, fmt.Errorf("resolve template style: %w", err)
		}
	}

	if err := o.Projects.SetState(ctx, project.ID, mysql.StateBriefing); err != nil {
		return nil, err
	}
	if _, err := o.runBriefAndConcepts(ctx, project.ID, inputText, style); err != nil {
		return nil, err
	}
	if err := o.Projects.SetState(ctx, project.ID, mysql.StateConceptsReady); err != nil {
		return nil, err
	}

	return o.Projects.GetByID(ctx, project.ID)
}

// RerunBrief re-runs Intake with extraDetail appended to the original input,
// cascading a supersede of everything downstream (spec: brief/rerun re-runs
// intake and downstream).
func (o *Orchestrator) RerunBrief(ctx context.Context, projectID, extraDetail string) (*mysql.Project, error) {
	if o.Locks != nil {
		unlock, ok := o.Locks.TryLock(projectID, "pipeline")
		if !ok {
			return nil, concurrency.ErrAlreadyInProgress
		}
		defer unlock()
	}
	if !o.Agents.Enabled() {
		return nil, agents.ErrAgentsUnavailable
	}
	project, err := o.Projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	style, err := o.currentStyleProfile(ctx, project)
	if err != nil {
		return nil, err
	}

	combinedText := project.InputText
	if extraDetail != "" {
		combinedText = combinedText + "\n\nAdditional detail: " + extraDetail
	}

	if err := SupersedeCascade(ctx, o.StageOutputs, projectID,
		mysql.StageConcepts, mysql.StageConceptSelection, mysql.StageCopy, mysql.StageArt); err != nil {
		return nil, err
	}
	if _, err := o.runBriefAndConcepts(ctx, projectID, combinedText, style); err != nil {
		return nil, err
	}
	if err := o.Projects.SetState(ctx, projectID, mysql.StateConceptsReady); err != nil {
		return nil, err
	}
	return o.Projects.GetByID(ctx, projectID)
}

// SelectConcept records the chosen concept, then runs the Copywriter and Art
// Director in sequence (spec: concepts/:n/select "runs agents 4-6" — 6/Proof
// is out of scope for this build, so this runs 4 and 5).
func (o *Orchestrator) SelectConcept(ctx context.Context, projectID string, index int) (*mysql.Project, error) {
	if o.Locks != nil {
		unlock, ok := o.Locks.TryLock(projectID, "pipeline")
		if !ok {
			return nil, concurrency.ErrAlreadyInProgress
		}
		defer unlock()
	}
	if !o.Agents.Enabled() {
		return nil, agents.ErrAgentsUnavailable
	}

	latestConcepts, err := o.StageOutputs.GetLatest(ctx, projectID, mysql.StageConcepts)
	if err != nil {
		return nil, fmt.Errorf("no concepts to select from: %w", err)
	}
	var cp ConceptsPayload
	if err := json.Unmarshal(latestConcepts.Payload, &cp); err != nil {
		return nil, fmt.Errorf("unmarshal concepts: %w", err)
	}
	if index < 0 || index >= len(cp.Concepts) {
		return nil, fmt.Errorf("concept index %d is out of range (have %d concepts)", index, len(cp.Concepts))
	}
	concept := cp.Concepts[index]

	project, err := o.Projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	brief, err := o.currentBrief(ctx, projectID)
	if err != nil {
		return nil, err
	}
	style, err := o.currentStyleProfile(ctx, project)
	if err != nil {
		return nil, err
	}

	if err := SupersedeCascade(ctx, o.StageOutputs, projectID, mysql.StageCopy, mysql.StageArt); err != nil {
		return nil, err
	}
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageConceptSelection,
		SelectionPayload{SelectedIndex: index, Concept: concept}, "", 0, 0); err != nil {
		return nil, err
	}
	o.Bus.Publish(projectID, Event{Stage: "concept_selection", Status: StatusDone, Payload: concept})

	if err := o.runCopyAndArt(ctx, project, brief, concept, style); err != nil {
		return nil, err
	}
	if err := o.Projects.SetState(ctx, projectID, mysql.StateDeveloping); err != nil {
		return nil, err
	}
	return o.Projects.GetByID(ctx, projectID)
}

// PatchCopy applies a manual copy edit (model="" marks it human-authored, not
// agent-authored) and, since art direction is built from copy's text, also
// regenerates art direction from the edited copy when AI is available — this
// is what makes "re-run only the stages downstream of it" (spec §4) concrete
// for a copy edit rather than just leaving art permanently stale.
func (o *Orchestrator) PatchCopy(ctx context.Context, projectID string, newCopy agents.Copy) (*mysql.Project, error) {
	if o.Locks != nil {
		unlock, ok := o.Locks.TryLock(projectID, "pipeline")
		if !ok {
			return nil, concurrency.ErrAlreadyInProgress
		}
		defer unlock()
	}
	project, err := o.Projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageCopy, newCopy, "", 0, 0); err != nil {
		return nil, err
	}
	o.Bus.Publish(projectID, Event{Stage: "copy", Status: StatusDone, Payload: newCopy})

	if !o.Agents.Enabled() {
		if err := SupersedeCascade(ctx, o.StageOutputs, projectID, mysql.StageArt); err != nil {
			return nil, err
		}
		return o.Projects.GetByID(ctx, projectID)
	}

	brief, err := o.currentBrief(ctx, projectID)
	if err != nil {
		return nil, err
	}
	concept, err := o.currentConcept(ctx, projectID)
	if err != nil {
		return nil, err
	}
	style, err := o.currentStyleProfile(ctx, project)
	if err != nil {
		return nil, err
	}
	if err := o.runArtDirection(ctx, project, brief, concept, newCopy, style, nil); err != nil {
		return nil, err
	}
	return o.Projects.GetByID(ctx, projectID)
}

// PatchArtDirection applies a manual art-direction edit. No agent call, no
// cascade — it's the last agent stage before rendering (spec: "PATCH
// art-direction: manual prompt edits").
func (o *Orchestrator) PatchArtDirection(ctx context.Context, projectID string, art agents.ArtDirection) (*mysql.Project, error) {
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageArt, art, "", 0, 0); err != nil {
		return nil, err
	}
	o.Bus.Publish(projectID, Event{Stage: "art_direction", Status: StatusDone, Payload: art})
	return o.Projects.GetByID(ctx, projectID)
}

// --- internal helpers ---

func (o *Orchestrator) runBriefAndConcepts(ctx context.Context, projectID, inputText string, style *agents.StyleProfile) ([]agents.Concept, error) {
	o.Bus.Publish(projectID, Event{Stage: "brief", Status: StatusRunning})
	brief, usage, err := o.Agents.RunIntake(ctx, inputText, style)
	if err != nil {
		o.Bus.Publish(projectID, Event{Stage: "brief", Status: StatusError, Payload: errPayload()})
		return nil, fmt.Errorf("run intake: %w", err)
	}
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageBrief, brief, o.Agents.ModelName(), usage.TokensIn, usage.TokensOut); err != nil {
		return nil, err
	}
	o.Bus.Publish(projectID, Event{Stage: "brief", Status: StatusDone, Payload: brief, TokensIn: usage.TokensIn, TokensOut: usage.TokensOut})

	o.Bus.Publish(projectID, Event{Stage: "concepts", Status: StatusRunning})
	concepts, usage2, err := o.Agents.RunCreativeDirector(ctx, brief, style)
	if err != nil {
		o.Bus.Publish(projectID, Event{Stage: "concepts", Status: StatusError, Payload: errPayload()})
		return nil, fmt.Errorf("run creative director: %w", err)
	}
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageConcepts, ConceptsPayload{Concepts: concepts}, o.Agents.ModelName(), usage2.TokensIn, usage2.TokensOut); err != nil {
		return nil, err
	}
	o.Bus.Publish(projectID, Event{Stage: "concepts", Status: StatusDone, Payload: concepts, TokensIn: usage2.TokensIn, TokensOut: usage2.TokensOut})

	return concepts, nil
}

func (o *Orchestrator) runCopyAndArt(ctx context.Context, project *mysql.Project, brief agents.Brief, concept agents.Concept, style *agents.StyleProfile) error {
	projectID := project.ID
	o.Bus.Publish(projectID, Event{Stage: "copy", Status: StatusRunning})
	cp, usage, err := o.Agents.RunCopywriter(ctx, brief, concept)
	if err != nil {
		o.Bus.Publish(projectID, Event{Stage: "copy", Status: StatusError, Payload: errPayload()})
		return fmt.Errorf("run copywriter: %w", err)
	}
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageCopy, cp, o.Agents.ModelName(), usage.TokensIn, usage.TokensOut); err != nil {
		return err
	}
	o.Bus.Publish(projectID, Event{Stage: "copy", Status: StatusDone, Payload: cp, TokensIn: usage.TokensIn, TokensOut: usage.TokensOut})

	return o.runArtDirection(ctx, project, brief, concept, cp, style, nil)
}

// runArtDirection only asks the Art Director to reserve zones for brand
// elements that will actually be composited in later (a per-project or
// global default logo, a global tagline, project-supplied footer text) —
// otherwise the reserved zone still gets baked into the generated artwork as
// a visible empty gap (its "fill" color) with nothing ever placed there.
func (o *Orchestrator) runArtDirection(ctx context.Context, project *mysql.Project, brief agents.Brief, concept agents.Concept, cp agents.Copy, style *agents.StyleProfile, excludedObjects []string) error {
	projectID := project.ID
	availableZones, err := o.resolveZoneAvailability(ctx, project)
	if err != nil {
		return fmt.Errorf("resolve zone availability: %w", err)
	}

	o.Bus.Publish(projectID, Event{Stage: "art_direction", Status: StatusRunning})
	art, usage, err := o.Agents.RunArtDirector(ctx, brief, concept, cp, style, availableZones, excludedObjects)
	if err != nil {
		o.Bus.Publish(projectID, Event{Stage: "art_direction", Status: StatusError, Payload: errPayload()})
		return fmt.Errorf("run art director: %w", err)
	}
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageArt, art, o.Agents.ModelName(), usage.TokensIn, usage.TokensOut); err != nil {
		return err
	}
	o.Bus.Publish(projectID, Event{Stage: "art_direction", Status: StatusDone, Payload: art, TokensIn: usage.TokensIn, TokensOut: usage.TokensOut})
	return nil
}

// RerunArtDirection re-runs just the Art Director from the project's existing
// brief/concept/copy/style — used when the user edits the sceneObjects
// checklist (spec-consistent with RerunBrief: edit + re-run, downstream
// stages that aren't stage_outputs — i.e. renders — are picked up as stale by
// the existing createdAt comparison, no separate cascade needed here).
func (o *Orchestrator) RerunArtDirection(ctx context.Context, projectID string, excludedObjects []string) (*mysql.Project, error) {
	if o.Locks != nil {
		unlock, ok := o.Locks.TryLock(projectID, "pipeline")
		if !ok {
			return nil, concurrency.ErrAlreadyInProgress
		}
		defer unlock()
	}
	if !o.Agents.Enabled() {
		return nil, agents.ErrAgentsUnavailable
	}
	project, err := o.Projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	brief, err := o.currentBrief(ctx, projectID)
	if err != nil {
		return nil, err
	}
	concept, err := o.currentConcept(ctx, projectID)
	if err != nil {
		return nil, err
	}
	cp, err := o.currentCopy(ctx, projectID)
	if err != nil {
		return nil, err
	}
	style, err := o.currentStyleProfile(ctx, project)
	if err != nil {
		return nil, err
	}

	if err := o.runArtDirection(ctx, project, brief, concept, cp, style, excludedObjects); err != nil {
		return nil, err
	}
	return o.Projects.GetByID(ctx, projectID)
}

// resolveZoneAvailability decides which of logo/tagline/footer will actually
// have something composited into them for this project: a logo if the
// project has its own upload or a global default exists, a tagline if a
// global default exists, a footer if the project has footer text set.
func (o *Orchestrator) resolveZoneAvailability(ctx context.Context, project *mysql.Project) ([]string, error) {
	activeAssets, err := o.BrandAssets.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active brand assets: %w", err)
	}
	hasGlobalLogo, hasGlobalTagline := false, false
	for _, a := range activeAssets {
		switch a.Kind {
		case mysql.BrandAssetLogo:
			hasGlobalLogo = true
		case mysql.BrandAssetTagline:
			hasGlobalTagline = true
		}
	}

	var zones []string
	if project.LogoObjectKey.Valid || hasGlobalLogo {
		zones = append(zones, agents.ZoneLogo)
	}
	if hasGlobalTagline {
		zones = append(zones, agents.ZoneTagline)
	}
	if project.FooterText.Valid && project.FooterText.String != "" {
		zones = append(zones, agents.ZoneFooter)
	}
	return zones, nil
}

func (o *Orchestrator) analyzeUploadsIfAny(ctx context.Context, projectID string, uploadIDs []string) (*agents.StyleProfile, error) {
	if len(uploadIDs) == 0 {
		return nil, nil
	}

	var images []agents.ImageInput
	var objectKeys []string
	for _, id := range uploadIDs {
		up, err := o.Uploads.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("load upload %s: %w", id, err)
		}
		data, err := o.Store.Get(ctx, up.ObjectKey)
		if err != nil {
			return nil, fmt.Errorf("download upload %s: %w", id, err)
		}
		images = append(images, agents.ImageInput{Data: data, MediaType: up.ContentType})
		objectKeys = append(objectKeys, up.ObjectKey)
	}

	o.Bus.Publish(projectID, Event{Stage: "style", Status: StatusRunning})
	profile, usage, err := o.Agents.AnalyzeStyle(ctx, images)
	if err != nil {
		o.Bus.Publish(projectID, Event{Stage: "style", Status: StatusError, Payload: errPayload()})
		return nil, fmt.Errorf("analyze style: %w", err)
	}
	if _, err := InsertSuperseding(ctx, o.StageOutputs, projectID, mysql.StageStyle, profile, o.Agents.ModelName(), usage.TokensIn, usage.TokensOut); err != nil {
		return nil, err
	}
	o.Bus.Publish(projectID, Event{Stage: "style", Status: StatusDone, Payload: profile, TokensIn: usage.TokensIn, TokensOut: usage.TokensOut})

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal analysed style profile: %w", err)
	}
	for _, key := range objectKeys {
		if err := o.ProjectRefs.Create(ctx, projectID, key, profileJSON); err != nil {
			return nil, err
		}
	}

	return &profile, nil
}

func (o *Orchestrator) currentBrief(ctx context.Context, projectID string) (agents.Brief, error) {
	row, err := o.StageOutputs.GetLatest(ctx, projectID, mysql.StageBrief)
	if err != nil {
		return agents.Brief{}, fmt.Errorf("no brief found for project: %w", err)
	}
	var brief agents.Brief
	if err := json.Unmarshal(row.Payload, &brief); err != nil {
		return agents.Brief{}, fmt.Errorf("unmarshal brief: %w", err)
	}
	return brief, nil
}

func (o *Orchestrator) currentConcept(ctx context.Context, projectID string) (agents.Concept, error) {
	row, err := o.StageOutputs.GetLatest(ctx, projectID, mysql.StageConceptSelection)
	if err != nil {
		return agents.Concept{}, fmt.Errorf("no concept selected for project: %w", err)
	}
	var sel SelectionPayload
	if err := json.Unmarshal(row.Payload, &sel); err != nil {
		return agents.Concept{}, fmt.Errorf("unmarshal concept selection: %w", err)
	}
	return sel.Concept, nil
}

func (o *Orchestrator) currentCopy(ctx context.Context, projectID string) (agents.Copy, error) {
	row, err := o.StageOutputs.GetLatest(ctx, projectID, mysql.StageCopy)
	if err != nil {
		return agents.Copy{}, fmt.Errorf("no copy found for project: %w", err)
	}
	var cp agents.Copy
	if err := json.Unmarshal(row.Payload, &cp); err != nil {
		return agents.Copy{}, fmt.Errorf("unmarshal copy: %w", err)
	}
	return cp, nil
}

func (o *Orchestrator) currentStyleProfile(ctx context.Context, project *mysql.Project) (*agents.StyleProfile, error) {
	if project.TemplateID.Valid {
		id := project.TemplateID.String
		return o.resolveTemplateStyle(ctx, &id)
	}
	refs, err := o.ProjectRefs.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("list project references: %w", err)
	}
	for _, ref := range refs {
		if len(ref.StyleProfile) == 0 {
			continue
		}
		var sp agents.StyleProfile
		if err := json.Unmarshal(ref.StyleProfile, &sp); err != nil {
			return nil, fmt.Errorf("unmarshal reference style profile: %w", err)
		}
		return &sp, nil
	}
	return nil, nil
}

func (o *Orchestrator) resolveTemplateStyle(ctx context.Context, templateID *string) (*agents.StyleProfile, error) {
	if templateID == nil {
		return nil, nil
	}
	tmpl, err := o.Templates.GetByID(ctx, *templateID)
	if err != nil {
		return nil, fmt.Errorf("load template: %w", err)
	}
	if len(tmpl.StyleProfile) == 0 {
		return nil, nil
	}
	var sp agents.StyleProfile
	if err := json.Unmarshal(tmpl.StyleProfile, &sp); err != nil {
		return nil, fmt.Errorf("unmarshal template style profile: %w", err)
	}
	return &sp, nil
}

// errPayload is deliberately generic — never leak a raw provider error to the
// client over SSE (spec §13); the full error is logged by the caller instead.
func errPayload() map[string]string {
	return map[string]string{"message": "This step failed. You can retry it."}
}
