package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// ----- Global DB/Redis clients ----

var db *sql.DB
var redisClient *redis.Client

func main() {
	// Load env for DB/Redis.
	dbDsn := getenv("DB_DSN", "postgres://postgres:postgres@localhost:5432/api_db?sslmode=disable")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")

	var err error
	db, err = sql.Open("postgres", dbDsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	// Test DB connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})
	if err = redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Could not ping redis at startup: %v", err)
	}

	r := mux.NewRouter()

	// Actual API endpoints would go here
	r.HandleFunc("/", helloHandler).Methods("GET")

	// Health/Readiness endpoints
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/ready", readinessHandler).Methods("GET")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

go func() {
		<-shutdownCh
		log.Println("Shutdown signal received, gracefully stopping server...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
		if db != nil {
			_ = db.Close()
		}
		if redisClient != nil {
			_ = redisClient.Close()
		}
		log.Println("Server shutdown complete.")
	}()

	log.Printf("API server starting on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello, world!")
}

// /health: checks DB and Redis connectivity, returns 200 OK or 503
func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	code := http.StatusOK
	problems := make([]string, 0)

	if err := db.PingContext(ctx); err != nil {
		problems = append(problems, "db")
		code = http.StatusServiceUnavailable
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		problems = append(problems, "redis")
		code = http.StatusServiceUnavailable
	}
	if code == http.StatusOK {
		fmt.Fprintln(w, "healthy")
	} else {
		w.WriteHeader(code)
		fmt.Fprintf(w, "unhealthy subsystems: %v\n", problems)
	}
}

// /ready: liveness/readiness probe
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// May be enhanced; here we reuse health check
	healthHandler(w, r)
}

// Helpers
func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
