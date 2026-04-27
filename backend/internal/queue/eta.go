package queue

import "math"

// DefaultAverageSessionSeconds is used when there is no measured average yet (minSamples == 0).
// Matches the client-side placeholder of 4 minutes per student.
const DefaultAverageSessionSeconds = 240.0

// PickAverageForETA uses this queue’s rolling average if any samples exist; otherwise the TA’s
// cross-queue average; otherwise 0 (callers use DefaultAverageSessionSeconds via EstimateWaitSeconds).
func PickAverageForETA(queueAvg float64, queueSampleCount int, taAvg float64, taSampleCount int) float64 {
	if queueSampleCount > 0 {
		return queueAvg
	}
	if taSampleCount > 0 {
		return taAvg
	}
	return 0
}

// EstimateWaitSeconds is the estimated time until a student at the given 1-based queue
// position is seen, assuming all ahead-of-line sessions use avgSessionDurationSeconds.
// Position 1 is the front of the line (only people ahead in line count, not the
// current student being served, who is not in queue_entries).
// If avgSessionDurationSeconds is <= 0, DefaultAverageSessionSeconds is used.
func EstimateWaitSeconds(avgSessionDurationSeconds float64, position int) int64 {
	avg := avgSessionDurationSeconds
	if avg <= 0 {
		avg = DefaultAverageSessionSeconds
	}
	if position < 1 {
		return 0
	}
	// No one ahead of position 1 in the listed queue.
	ahead := position - 1
	return int64(math.Round(avg * float64(ahead)))
}

// EstimateMaxWaitSecondsForQueue is the wait for the last student in line
// (or 0 if the queue is empty). Uses DefaultAverageSessionSeconds when avg is unknown.
func EstimateMaxWaitSecondsForQueue(avgSessionDurationSeconds float64, numWaiting int) int64 {
	if numWaiting < 1 {
		return 0
	}
	return EstimateWaitSeconds(avgSessionDurationSeconds, numWaiting)
}
