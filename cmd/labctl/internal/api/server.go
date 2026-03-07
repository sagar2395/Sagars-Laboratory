package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/sagars-lab/labctl/internal/config"
	"github.com/sagars-lab/labctl/internal/executor"
	"github.com/sagars-lab/labctl/internal/platform"
	"github.com/sagars-lab/labctl/internal/scenario"
	"github.com/sagars-lab/labctl/internal/services"
)

// Server is the API server that backs the web UI.
type Server struct {
	cfg      *config.Config
	exec     *executor.Executor
	registry *platform.Registry
	scenes   *scenario.Engine
	svcs     *services.Registry
	router   *mux.Router
	upgrader websocket.Upgrader
	uiFS     fs.FS
}

// NewServer creates a new API server. The embeddedUI parameter should be the
// embedded ui/dist filesystem (from go:embed). If nil or empty, the server
// falls back to serving UI files from the project's ui/dist/ directory.
func NewServer(cfg *config.Config, exec *executor.Executor, registry *platform.Registry, scenes *scenario.Engine, svcs *services.Registry, embeddedUI fs.FS) *Server {
	s := &Server{
		cfg:      cfg,
		exec:     exec,
		registry: registry,
		scenes:   scenes,
		svcs:     svcs,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		uiFS: embeddedUI,
	}
	s.setupRoutes()
	return s
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(corsMiddleware)
	api.Use(jsonMiddleware)

	api.HandleFunc("/status", s.handleStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/apps", s.handleListApps).Methods("GET", "OPTIONS")
	api.HandleFunc("/apps/{name}/deploy", s.handleAppDeploy).Methods("POST", "OPTIONS")
	api.HandleFunc("/apps/{name}/destroy", s.handleAppDestroy).Methods("POST", "OPTIONS")
	api.HandleFunc("/platform", s.handlePlatformStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/platform/up", s.handlePlatformUp).Methods("POST", "OPTIONS")
	api.HandleFunc("/platform/down", s.handlePlatformDown).Methods("POST", "OPTIONS")
	api.HandleFunc("/scenarios", s.handleListScenarios).Methods("GET", "OPTIONS")
	api.HandleFunc("/scenarios/{name}", s.handleScenarioInfo).Methods("GET", "OPTIONS")
	api.HandleFunc("/scenarios/{name}/up", s.handleScenarioUp).Methods("POST", "OPTIONS")
	api.HandleFunc("/scenarios/{name}/down", s.handleScenarioDown).Methods("POST", "OPTIONS")
	api.HandleFunc("/services", s.handleListServices).Methods("GET", "OPTIONS")
	api.HandleFunc("/services/{name}/up", s.handleServiceUp).Methods("POST", "OPTIONS")
	api.HandleFunc("/services/{name}/down", s.handleServiceDown).Methods("POST", "OPTIONS")
	api.HandleFunc("/ws", s.handleWebSocket)

	// Serve UI — use embedded FS if available, fall back to filesystem for dev
	var uiHandler http.Handler
	if s.uiFS != nil {
		if _, err := fs.Stat(s.uiFS, "index.html"); err == nil {
			uiHandler = http.FileServer(http.FS(s.uiFS))
		}
	}
	if uiHandler == nil {
		uiHandler = http.FileServer(http.Dir(s.cfg.ProjectRoot + "/ui/dist"))
	}
	s.router.PathPrefix("/").Handler(uiHandler)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
