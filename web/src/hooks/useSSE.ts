import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { PipelineEvent, SSEStatus } from "../api/types";

export type StageStatusMap = Record<string, SSEStatus>;

/**
 * Subscribes to GET /projects/:id/events (spec §9). Drives an immediate local
 * reducer for per-stage running/done/error pills, and invalidates the
 * project query on done/error so the canonical state (with computed render
 * staleness etc.) refetches from the server. The browser's EventSource
 * auto-reconnects on drop — no manual resume logic needed for this build.
 */
export function useSSE(projectId: string | undefined) {
  const [statuses, setStatuses] = useState<StageStatusMap>({});
  const queryClient = useQueryClient();
  const lastEventRef = useRef<PipelineEvent | null>(null);

  useEffect(() => {
    if (!projectId) return;
    setStatuses({});

    const source = new EventSource(`/api/v1/projects/${projectId}/events`, { withCredentials: true });

    source.onmessage = (msg) => {
      let evt: PipelineEvent;
      try {
        evt = JSON.parse(msg.data);
      } catch {
        return;
      }
      lastEventRef.current = evt;
      setStatuses((prev) => ({ ...prev, [evt.stage]: evt.status }));
      if (evt.status === "done" || evt.status === "error") {
        queryClient.invalidateQueries({ queryKey: ["project", projectId] });
      }
    };

    return () => source.close();
  }, [projectId, queryClient]);

  return { statuses, lastEvent: lastEventRef.current };
}
