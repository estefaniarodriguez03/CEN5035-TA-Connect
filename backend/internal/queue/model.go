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
