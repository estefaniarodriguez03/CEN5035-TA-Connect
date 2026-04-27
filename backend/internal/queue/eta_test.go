package queue

import "testing"

func TestEstimateWaitSeconds(t *testing.T) {
	t.Parallel()
	if g := EstimateWaitSeconds(60, 1); g != 0 {
		t.Errorf("pos 1: got %d want 0", g)
	}
	if g := EstimateWaitSeconds(60, 2); g != 60 {
		t.Errorf("pos 2: got %d want 60", g)
	}
	want := int64(2 * DefaultAverageSessionSeconds)
	if g := EstimateWaitSeconds(0, 3); g != want {
		t.Errorf("default avg pos 3: got %d want %d", g, want)
	}
	if g := EstimateWaitSeconds(60, 0); g != 0 {
		t.Errorf("pos 0: got %d want 0", g)
	}
}

func TestEstimateMaxWaitSecondsForQueue(t *testing.T) {
	t.Parallel()
	if g := EstimateMaxWaitSecondsForQueue(120, 0); g != 0 {
		t.Errorf("empty: got %d", g)
	}
	if g := EstimateMaxWaitSecondsForQueue(60, 3); g != 120 {
		t.Errorf("last of 3: got %d want 120", g)
	}
}
