package routes

import (
	"database/sql"
	"net/http"

	"backend/internal/auth"
	"backend/internal/studentschedule"
	"backend/internal/course"
	"backend/internal/httperr"
	"backend/internal/officehour"
	"backend/internal/queue"
	"backend/internal/sessionlog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRoutes returns a configured chi.Mux with health check, auth, and CORS.
func SetupRoutes(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// --- Public routes (no auth) ---
	r.Get("/health", healthHandler(db))
	r.Post("/api/login", auth.Login(db))
	r.Post("/api/register", auth.Register(db))

	r.Get("/api/office-hours/ta/{ta_id}", officehour.ListByTA(db))
	r.Get("/api/office-hours/course/{course_id}", officehour.ListByCourse(db))

	r.Get("/api/queues/active", queue.GetActiveQueueByCourse(db))

	r.Get("/api/courses", course.ListAll(db))

	// SSE is public so unauthenticated browsers can subscribe.
	r.Get("/api/queues/{id}/events", queue.StreamQueueEvents(queue.DefaultHub))
	r.Get("/api/queues/{id}", queue.GetQueue(db))

	// --- TA-only routes ---
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Use(auth.RequireRole("ta"))

		r.Put("/api/users/{id}/profile", auth.UpdateProfile(db))
		r.Post("/api/office-hours", officehour.Create(db))
		r.Put("/api/office-hours/{id}", officehour.Update(db))
		r.Delete("/api/office-hours/{id}", officehour.Delete(db))

		r.Post("/api/queues", queue.CreateQueue(db))
		r.Patch("/api/queues/{id}/state", queue.UpdateQueueState(db))
		r.Post("/api/queues/{id}/status", queue.UpdateStatus(db))
		r.Post("/api/queues/{id}/next", queue.Next(db))
		r.Post("/api/queues/{id}/announcement", queue.PostAnnouncement(db))
		r.Post("/api/queues/{id}/session", queue.StartSession(db))
		r.Post("/api/ta/courses", course.AddForTA(db))
		r.Get("/api/ta/courses", course.ListForTA(db))
		r.Delete("/api/ta/courses/{id}", course.DeleteForTA(db))
	})

	// --- Student-only routes ---
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Use(auth.RequireRole("student"))

		r.Post("/api/queues/{id}/join", queue.Join(db))
		r.Post("/api/queues/{id}/leave", queue.Leave(db))

		r.Post("/api/student/schedule", studentschedule.Add(db))
		r.Get("/api/student/schedule", studentschedule.List(db))
		r.Delete("/api/student/schedule/{id}", studentschedule.Remove(db))
	})

	// --- Authenticated: TA and student (own session history) ---
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Get("/api/session-history", sessionlog.ListHistory(db))
	})

	return r
}

// healthHandler returns 200 if the database connection is alive, 503 otherwise.
func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			httperr.Write(w, http.StatusServiceUnavailable, "service_unavailable", "database unavailable", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

// corsMiddleware adds basic CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
