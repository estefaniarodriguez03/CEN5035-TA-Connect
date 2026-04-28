package queue

// QueueStatus is persisted on queues.status (open | paused | closed).
type QueueStatus string

const (
	QueueStatusOpen   QueueStatus = "open"
	QueueStatusPaused QueueStatus = "paused"
	QueueStatusClosed QueueStatus = "closed"
)

// Valid reports whether s is a supported queue status value.
func (s QueueStatus) Valid() bool {
	switch s {
	case QueueStatusOpen, QueueStatusPaused, QueueStatusClosed:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether setting status to `to` is valid when the queue
// is currently `from`. A transition to the same value is always allowed (no-op).
//
// Valid edges: open↔paused, open/closed, paused/closed, closed→open only
// (a closed queue must be reopened to open before it can be paused again).
func CanTransitionTo(from, to QueueStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case QueueStatusOpen:
		return to == QueueStatusPaused || to == QueueStatusClosed
	case QueueStatusPaused:
		return to == QueueStatusOpen || to == QueueStatusClosed
	case QueueStatusClosed:
		return to == QueueStatusOpen
	default:
		return false
	}
}

// AllowedTargetStatuses returns valid target statuses for PATCH /state (including
// the current status for idempotent requests).
func AllowedTargetStatuses(from QueueStatus) []string {
	if !from.Valid() {
		return nil
	}
	var out []string
	for _, c := range []QueueStatus{QueueStatusOpen, QueueStatusPaused, QueueStatusClosed} {
		if CanTransitionTo(from, c) {
			out = append(out, string(c))
		}
	}
	return out
}
