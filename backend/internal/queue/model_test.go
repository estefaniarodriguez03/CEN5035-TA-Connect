package queue

import "testing"

func TestCanTransitionTo(t *testing.T) {
	t.Parallel()
	if !CanTransitionTo(QueueStatusOpen, QueueStatusOpen) {
		t.Error("open->open")
	}
	if !CanTransitionTo(QueueStatusOpen, QueueStatusPaused) || !CanTransitionTo(QueueStatusOpen, QueueStatusClosed) {
		t.Error("open->paused/closed")
	}
	if !CanTransitionTo(QueueStatusPaused, QueueStatusOpen) || !CanTransitionTo(QueueStatusPaused, QueueStatusClosed) {
		t.Error("paused->open/closed")
	}
	if !CanTransitionTo(QueueStatusClosed, QueueStatusOpen) || !CanTransitionTo(QueueStatusClosed, QueueStatusClosed) {
		t.Error("closed->open (reopen) or idempotent closed")
	}
	if CanTransitionTo(QueueStatusClosed, QueueStatusPaused) {
		t.Error("closed->paused must be invalid")
	}
}

func TestAllowedTargetStatuses(t *testing.T) {
	t.Parallel()
	open := AllowedTargetStatuses(QueueStatusOpen)
	if len(open) != 3 {
		t.Fatalf("open: got %v", open)
	}
	closed := AllowedTargetStatuses(QueueStatusClosed)
	if len(closed) != 2 {
		t.Fatalf("closed: got %v", closed)
	}
}
