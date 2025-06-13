package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"log"
	"net/http"
	"time"

	"github.com/tliefheid/go-ils/internal/backend"
	"github.com/tliefheid/go-ils/internal/repository/postgres"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	HTTP       string
}

func LoadConfig() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "library"),
		HTTP:       getEnv("HTTP", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := LoadConfig()

	db, err := postgres.NewStore(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()

	s, err := backend.New(backend.Config{
		Repository: db,
	})
	if err != nil {
		log.Fatalf("Failed to initialize backend service: %v", err)
	}

	err = db.Migrate("migrations.sql")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Library ILS Backend - Go API running on ", cfg.HTTP)

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
