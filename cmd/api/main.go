package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"practice-one/internal/handlers"
	"practice-one/internal/middleware"
	"practice-one/internal/router"
	"practice-one/internal/store"
)

func main() {
	taskStore := store.NewTaskStore()

	taskHandler := handlers.NewTaskHandler(taskStore)

	r := router.NewRouter()

	r.GET("/v1/tasks", taskHandler.GetTask)
	r.POST("/v1/tasks", taskHandler.CreateTask)
	r.PATCH("/v1/tasks", taskHandler.UpdateTask)
	r.DELETE("/v1/tasks", taskHandler.DeleteTask)

	r.GET("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	r.PrintRoutes()

	validAPIKeys := map[string]bool{
		"secret12345":      true,
		"dev-key-001":      true,
		"production-key-1": true,
	}

	rateLimiter := middleware.NewRateLimiter(10)

	handler := middleware.Chain(
		middleware.Logger,
		middleware.RequestID,
		rateLimiter.Limit,
		middleware.APIKeyAuth(validAPIKeys),
	)(r)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		log.Println("Shutting down server gracefully...")
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("Starting server on %s", srv.Addr)
	log.Printf("Swagger documentation available at http://localhost:8080/swagger")
	log.Printf("API v1 endpoints available at /v1/tasks")
	log.Printf("Valid API keys: secret12345, dev-key-001, production-key-1")

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
	log.Println("Server stopped gracefully")
}
