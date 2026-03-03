package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type EventType string

const (
	EventStudentJoined EventType = "STUDENT_JOINED"
	EventStudentLeft   EventType = "STUDENT_LEFT"
	EventQueueUpdated  EventType = "QUEUE_UPDATED"
	EventStudentServed EventType = "STUDENT_SERVED"
)

type QueueEvent struct {
	Type    EventType `json:"type"`
	QueueID int       `json:"queue_id"`
	Payload any       `json:"payload,omitempty"`
}

// Hub keeps a list of subscribers per queue.
type Hub struct {
	mu     sync.RWMutex
	queues map[int]map[chan QueueEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{
		queues: make(map[int]map[chan QueueEvent]struct{}),
	}
}

// DefaultHub is the global hub instance used by queue handlers.
var DefaultHub = NewHub()

// Subscribe registers a new subscriber for a queue and returns a channel of events.
// When ctx is done, the subscriber is removed.
func (h *Hub) Subscribe(ctx context.Context, queueID int) <-chan QueueEvent {
	ch := make(chan QueueEvent, 8)

	h.mu.Lock()
	if h.queues[queueID] == nil {
		h.queues[queueID] = make(map[chan QueueEvent]struct{})
	}
	h.queues[queueID][ch] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.mu.Lock()
		defer h.mu.Unlock()
		if subs, ok := h.queues[queueID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(h.queues, queueID)
			}
		}
		close(ch)
	}()

	return ch
}

// Publish sends an event to all subscribers of a queue.
func (h *Hub) Publish(queueID int, evt QueueEvent) {
	h.mu.RLock()
	subs := h.queues[queueID]
	h.mu.RUnlock()

	for ch := range subs {
		select {
		case ch <- evt:
		default:
			// drop if subscriber is too slow
		}
	}
}

// StreamQueueEvents is the SSE endpoint: GET /api/queues/{id}/events.
func StreamQueueEvents(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueID, err := parseQueueID(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid queue id"})
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		events := hub.Subscribe(ctx, queueID)

		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-events:
				if !ok {
					return
				}
				data, err := json.Marshal(evt.Payload)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: %s\n", evt.Type)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}