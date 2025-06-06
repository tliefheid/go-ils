package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/library-ils-backend/internal/frontend"
)

//go:embed assets/*.css assets/*.gohtml
var resources embed.FS

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := struct {
		HTTP string
	}{
		HTTP: ":4000",
	}
	s, err := frontend.New(frontend.Config{
		BackendUri: "http://localhost:8182",
	})

	if err != nil {
		log.Fatalf("Failed to initialize backend service: %v", err)
	}

	s.Mux().Handle("/assets/*", http.FileServer(http.FS(resources)))

	fmt.Println("Library ILS Frontend - Running on ", cfg.HTTP)

	srv := &http.Server{
		Addr:    cfg.HTTP,
		Handler: s.Mux(),
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	go func() {
		fmt.Println("Starting server on", cfg.HTTP)

		if err := http.ListenAndServe(cfg.HTTP, srv.Handler); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	fmt.Println("wait for shutdown signal...")
	<-ctx.Done()
	fmt.Println("shutting down...")
}
