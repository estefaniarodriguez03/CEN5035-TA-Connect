package queue

// queueStateJSON is the public JSON shape for GET /api/queues/{id} and GET /api/queues/active
// (queue “state” including ETA fields).
func queueStateJSON(
	id, courseID, taID int,
	status, createdAt string,
	entries []Entry,
	queueAverageSec float64,
	queueSampleCount int,
	taAverageSec float64,
	taSampleCount int,
) map[string]any {
	etaAvg := PickAverageForETA(queueAverageSec, queueSampleCount, taAverageSec, taSampleCount)
	// Copy slice so we do not retain unexpected aliasing of caller-owned backing arrays.
	enriched := make([]Entry, len(entries))
	copy(enriched, entries)
	for i := range enriched {
		w := EstimateWaitSeconds(etaAvg, enriched[i].Position)
		enriched[i].EstimatedWaitSeconds = w
	}
	return map[string]any{
		"id":         id,
		"course_id":  courseID,
		"ta_id":      taID,
		"status":     status,
		"created_at": createdAt,
		// This queue’s observed average (0 until the second /next on this queue).
		"average_session_duration_seconds": queueAverageSec,
		// Same metric aggregated across all queues owned by this TA.
		"ta_average_session_duration_seconds": taAverageSec,
		"estimated_wait_time_seconds":           EstimateMaxWaitSecondsForQueue(etaAvg, len(enriched)),
		"entries":                               enriched,
	}
}
