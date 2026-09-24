package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"armss-poster-studio/internal/pipeline"
)

type SSEHandler struct {
	Bus *pipeline.Bus
}

func NewSSEHandler(bus *pipeline.Bus) *SSEHandler {
	return &SSEHandler{Bus: bus}
}

// Stream implements GET /projects/:id/events (spec §9): server-sent events
// with a 15s heartbeat, so the UI shows each agent completing instead of one
// long spinner.
func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events, unsubscribe := h.Bus.Subscribe(projectID)
	defer unsubscribe()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, open := <-events:
			if !open {
				return
			}
			payload, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
