package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port                = getEnv("PORT", "8080")
	readinessFailure    = getEnv("READINESS_FAILURE", "false") == "true"
	shutdownTimeout     = getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second)
	simulateFailure     bool
	httpRequestsTotal   prometheus.Counter
	httpRequestDuration prometheus.Histogram
)

func init() {
	// Register Prometheus metrics
	httpRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	})
	prometheus.MustRegister(httpRequestsTotal)

	httpRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})
	prometheus.MustRegister(httpRequestDuration)
}

func main() {
	flag.BoolVar(&simulateFailure, "failure", readinessFailure, "Simulate readiness check failure")
	flag.Parse()

	// Create router
	mux := http.NewServeMux()

	// Health endpoint - always OK
	mux.HandleFunc("/health", handleHealth)

	// Readiness endpoint - can simulate failure
	mux.HandleFunc("/ready", handleReady)

	// Metrics endpoint - Prometheus format
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting server on %s\n", server.Addr)
		fmt.Printf("Simulate failure: %v\n", simulateFailure)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v\n", sig)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped")
}

// handleHealth returns 200 OK always
func handleHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		httpRequestsTotal.Inc()
		httpRequestDuration.Observe(time.Since(start).Seconds())
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleReady can simulate readiness failure based on flag
func handleReady(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		httpRequestsTotal.Inc()
		httpRequestDuration.Observe(time.Since(start).Seconds())
	}()

	w.Header().Set("Content-Type", "application/json")

	if simulateFailure {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not_ready","reason":"simulated failure"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// getEnv returns environment variable value or default
func getEnv(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultVal
}

// getDurationEnv returns environment variable as duration or default
func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}
