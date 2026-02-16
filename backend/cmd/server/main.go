package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/db"
	"backend/internal/routes"

	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file loaded:", err)
	}

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatalf("database migrate: %v", err)
	}

	defer database.Close()

	router := routes.SetupRoutes(database)
	port := getEnv("PORT", defaultPort)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	log.Println("server stopped")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
