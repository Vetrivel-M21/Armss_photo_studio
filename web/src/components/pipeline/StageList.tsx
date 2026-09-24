import type { ProjectDetail } from "../../api/types";
import type { StageStatusMap } from "../../hooks/useSSE";
import { BriefStage } from "./BriefStage";
import { ConceptsStage } from "./ConceptsStage";
import { CopyStage } from "./CopyStage";
import { ArtDirectionStage } from "./ArtDirectionStage";

interface Props {
  project: ProjectDetail;
  liveStatuses: StageStatusMap;
}

// A vertical run of stages (spec §12), each showing its agent's output as it
// streams in over SSE. Layout Proof and Review are out of scope for this
// build (spec §15 steps 1-8) so there's no placeholder section for them.
export function StageList({ project, liveStatuses }: Props) {
  return (
    <div className="space-y-4">
      <BriefStage projectId={project.id} brief={project.stages.brief?.payload} liveStatus={liveStatuses["brief"]} />
      <ConceptsStage
        projectId={project.id}
        concepts={project.stages.concepts?.payload.concepts}
        selectedIndex={project.stages.concept_selection?.payload.selectedIndex}
        liveStatus={liveStatuses["concepts"]}
      />
      <CopyStage projectId={project.id} copy={project.stages.copy?.payload} liveStatus={liveStatuses["copy"]} />
      <ArtDirectionStage
        projectId={project.id}
        art={project.stages.art_direction?.payload}
        liveStatus={liveStatuses["art_direction"]}
      />
    </div>
  );
}
