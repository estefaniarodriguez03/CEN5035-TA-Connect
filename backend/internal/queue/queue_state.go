package queue

// queueStateJSON is the public JSON shape for GET /api/queues/{id} and GET /api/queues/active
// (queue “state” including ETA fields).
func queueStateJSON(id, courseID, taID int, status, createdAt string, entries []Entry, averageSessionSeconds float64) map[string]any {
	// Copy slice so we do not retain unexpected aliasing of caller-owned backing arrays.
	enriched := make([]Entry, len(entries))
	copy(enriched, entries)
	for i := range enriched {
		w := EstimateWaitSeconds(averageSessionSeconds, enriched[i].Position)
		enriched[i].EstimatedWaitSeconds = w
	}
	return map[string]any{
		"id":                               id,
		"course_id":                        courseID,
		"ta_id":                            taID,
		"status":                           status,
		"created_at":                       createdAt,
		"average_session_duration_seconds": averageSessionSeconds,
		"estimated_wait_time_seconds":      EstimateMaxWaitSecondsForQueue(averageSessionSeconds, len(enriched)),
		"entries":                          enriched,
	}
}
