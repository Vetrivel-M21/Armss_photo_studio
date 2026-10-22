// Wire types mirroring the Go backend's internal/http/dto and internal/agents
// packages exactly (field names, optionality) — see cmd/server for the source
// of truth. Kept in one file since they're pure data shapes, not behavior.

export type Role = "staff" | "admin";

export interface User {
  id: string;
  email: string;
  role: Role;
}

export interface PaletteHex {
  hex: string;
  role: string;
}

export interface StyleProfile {
  summary: string;
  palette: PaletteHex[];
  typography: string;
  layoutPatterns: string;
  imagery: string;
  finish: string;
  tone: string;
  logoTreatment: string;
  fixedElements: string[];
  consistency: string;
}

export interface Brief {
  objective: string;
  audience: string;
  keyMessage: string;
  occasion: string;
  eventDate: string;
  tone: string[];
  format: string;
  dimensions: string;
  palette: string[];
  mustInclude: string[];
  avoid: string[];
  assumptions: string[];
  questions: string[];
}

export interface Concept {
  name: string;
  bigIdea: string;
  visualMetaphor: string;
  mood: string;
  palette: string;
}

export interface ConceptsPayload {
  concepts: Concept[];
}

export interface Copy {
  eyebrow?: string;
  headline: string;
  subhead: string;
  cta: string;
  valueWords: string[];
  dateText?: string;
  tagline?: string;
  alternates: string[];
}

export interface Zone {
  id: "logo" | "tagline" | "footer" | string;
  x: number;
  y: number;
  w: number;
  h: number;
  fill: string;
  note: string;
}

export type SceneObjectCategory = "hero" | "cultural" | "nature" | "supporting" | "decorative";

export interface SceneObject {
  name: string;
  category: SceneObjectCategory;
}

export interface SectionAllocation {
  section: string;
  percentage: number;
  role: string;
}

export interface CompositionPlan {
  strategy: string;
  description: string;
  framing?: string;
  sections: SectionAllocation[];
}

export interface VisualHierarchy {
  primaryElement: string;
  secondaryElement: string;
  tertiaryElement: string;
  decorativeElements: string[];
}

export interface DepthPlanes {
  foreground: string;
  midground: string;
  background: string;
  focusDepth: string;
}

export interface RelationalIntegration {
  interactions: string[];
  shadows: string;
  lightSpill: string;
}

export interface MaterialRealism {
  metals?: string;
  fabrics?: string;
  surfaces?: string;
  organic?: string;
}

export interface LightingSetup {
  primaryLight: string;
  secondaryLight: string;
  rimLight: string;
  atmosphere: string;
}

export interface ColorDistribution {
  dominant: string;
  secondary: string;
  neutral: string;
  accents: string;
}

export interface TypographyPlan {
  heroTitleTreatment: string;
  eyebrowStyle: string;
  dateAndBadges?: string;
  supportingBody: string;
}

export interface PosterSpec {
  artDirectionStyle: string;
  composition: CompositionPlan;
  visualHierarchy: VisualHierarchy;
  depth: DepthPlanes;
  integration: RelationalIntegration;
  materials: MaterialRealism;
  lighting: LightingSetup;
  colors: ColorDistribution;
  typography: TypographyPlan;
  negativeRules?: string[];
}

export interface ArtDirection {
  spec?: PosterSpec;
  layout: string;
  focalPoint: string;
  typography: string;
  colour: string;
  textToRender: string[];
  fullPosterPrompt: string;
  backgroundPrompt: string;
  negativePrompt: string;
  aspectRatio: string;
  reservedZones: Zone[];
  compositeSteps: string[];
  sceneObjects: SceneObject[];
}

export interface SelectionPayload {
  selectedIndex: number;
  concept: Concept;
}

export type StyleProfileStatus = "ready" | "pending_manual";

export interface Template {
  id: string;
  title: string;
  imageUrl: string;
  thumbUrl?: string;
  occasionTags: string[];
  season?: string;
  styleProfile?: StyleProfile;
  styleProfileStatus: StyleProfileStatus;
  isActive: boolean;
  createdAt: string;
}

export interface BrandAsset {
  id: string;
  kind: "logo" | "tagline" | "footer";
  imageUrl: string;
  meta?: Record<string, unknown>;
  version: number;
  isActive: boolean;
  createdAt: string;
}

export interface UploadResponse {
  id: string;
  objectKey?: string;
  contentType: string;
  sizeBytes: number;
}

export interface RecommendationView {
  template: Template;
  reason: string;
}

export type ProjectState =
  | "draft"
  | "briefing"
  | "concepts_ready"
  | "developing"
  | "render_ready"
  | "composited"
  | "approved";

export interface ProjectSummary {
  id: string;
  title?: string;
  inputText: string;
  templateId?: string;
  state: ProjectState;
  createdAt: string;
  updatedAt: string;
}

// The stage key on the wire ("art_direction") differs from the DB token
// ("art") — see internal/http/handlers/projects.go's wireStageName.
export type StageKey =
  | "style"
  | "brief"
  | "concepts"
  | "concept_selection"
  | "copy"
  | "art_direction";

export interface StageOutputView<TPayload = unknown> {
  id: string;
  payload: TPayload;
  model?: string;
  tokensIn?: number;
  tokensOut?: number;
  createdAt: string;
}

export interface RenderView {
  id: string;
  kind: "raw" | "composited";
  provider?: string;
  url: string;
  prompt?: string;
  costCents?: number;
  stale: boolean;
  createdAt: string;
}

export interface ProjectDetail {
  id: string;
  title?: string;
  inputText: string;
  templateId?: string;
  logoUrl?: string;
  footerText?: string;
  state: ProjectState;
  createdAt: string;
  updatedAt: string;
  stages: Partial<{
    style: StageOutputView<StyleProfile>;
    brief: StageOutputView<Brief>;
    concepts: StageOutputView<ConceptsPayload>;
    concept_selection: StageOutputView<SelectionPayload>;
    copy: StageOutputView<Copy>;
    art_direction: StageOutputView<ArtDirection>;
  }>;
  renders: RenderView[];
}

export interface ZoneOverride {
  x?: number;
  y?: number;
  w?: number;
  h?: number;
}

export type SSEStatus = "running" | "done" | "error";

export interface PipelineEvent {
  stage: string;
  status: SSEStatus;
  payload?: unknown;
  tokensIn: number;
  tokensOut: number;
}
