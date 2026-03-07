package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sagars-lab/labctl/internal/config"
	"github.com/sagars-lab/labctl/internal/k8s"
)

// StatusResponse represents the overall lab status.
type StatusResponse struct {
	Cluster  *k8s.ClusterInfo   `json:"cluster"`
	Platform PlatformStatusResp `json:"platform"`
	Apps     []AppStatusResp    `json:"apps"`
}

type PlatformStatusResp struct {
	Ingress ComponentStatus `json:"ingress"`
	Metrics ComponentStatus `json:"metrics"`
	Logging ComponentStatus `json:"logging"`
	Tracing ComponentStatus `json:"tracing"`
}

type ComponentStatus struct {
	Provider string `json:"provider"`
	Active   bool   `json:"active"`
}

type AppStatusResp struct {
	Name     string `json:"name"`
	Build    string `json:"buildStrategy"`
	Deploy   string `json:"deployStrategy"`
	Deployed bool   `json:"deployed"`
	Replicas string `json:"replicas,omitempty"`
	Ready    string `json:"ready,omitempty"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp := StatusResponse{}

	// Cluster info
	clusterInfo, _ := k8s.GetClusterInfo(ctx)
	resp.Cluster = clusterInfo

	// Platform status
	resp.Platform = PlatformStatusResp{
		Ingress: ComponentStatus{
			Provider: s.cfg.IngressProvider,
			Active:   k8s.NamespaceExists(ctx, "traefik"),
		},
		Metrics: ComponentStatus{
			Provider: s.cfg.MetricsProvider,
			Active:   k8s.NamespaceExists(ctx, "monitoring"),
		},
		Logging: ComponentStatus{
			Provider: s.cfg.LoggingProvider,
		},
		Tracing: ComponentStatus{
			Provider: s.cfg.TracingProvider,
		},
	}

	// Apps
	apps, _ := config.ListApps(s.cfg.ProjectRoot)
	for _, appName := range apps {
		appCfg, _ := config.LoadAppConfig(s.cfg.ProjectRoot, appName)
		appResp := AppStatusResp{Name: appName}
		if appCfg != nil {
			appResp.Build = appCfg.BuildStrategy
			appResp.Deploy = appCfg.DeployStrategy
			ns := appName
			if appCfg.Namespace != "" {
				ns = appCfg.Namespace
			}
			status, _ := k8s.GetAppStatus(ctx, appName, ns)
			if status != nil {
				appResp.Deployed = status.Deployed
				appResp.Replicas = status.Replicas
				appResp.Ready = status.Ready
			}
		}
		resp.Apps = append(resp.Apps, appResp)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	apps, err := config.ListApps(s.cfg.ProjectRoot)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var result []AppStatusResp
	for _, appName := range apps {
		appCfg, _ := config.LoadAppConfig(s.cfg.ProjectRoot, appName)
		appResp := AppStatusResp{Name: appName}
		if appCfg != nil {
			appResp.Build = appCfg.BuildStrategy
			appResp.Deploy = appCfg.DeployStrategy
			ns := appName
			if appCfg.Namespace != "" {
				ns = appCfg.Namespace
			}
			status, _ := k8s.GetAppStatus(ctx, appName, ns)
			if status != nil {
				appResp.Deployed = status.Deployed
				appResp.Replicas = status.Replicas
				appResp.Ready = status.Ready
			}
		}
		result = append(result, appResp)
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleAppDeploy(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	err := s.exec.RunScript("engine/deploy.sh", "deploy", name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("deploy failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deployed", "app": name})
}

func (s *Server) handleAppDestroy(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	err := s.exec.RunScript("engine/deploy.sh", "destroy", name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("destroy failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "destroyed", "app": name})
}

func (s *Server) handlePlatformStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	categories := s.registry.Categories()
	result := make(map[string][]map[string]interface{})

	for _, cat := range categories {
		providers := s.registry.GetProviders(cat)
		for _, p := range providers {
			entry := map[string]interface{}{
				"name":      p.Name,
				"category":  cat,
				"installed": k8s.NamespaceExists(ctx, p.Name),
			}
			result[cat] = append(result[cat], entry)
		}
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handlePlatformUp(w http.ResponseWriter, r *http.Request) {
	// Install ingress
	if s.cfg.IngressProvider != "" {
		s.registry.Install("ingress", s.cfg.IngressProvider, s.exec)
	}
	// Install metrics
	if s.cfg.MetricsProvider != "" {
		s.registry.Install("monitoring", s.cfg.MetricsProvider, s.exec)
	}
	// Install grafana
	s.registry.Install("monitoring", "grafana", s.exec)

	respondJSON(w, http.StatusOK, map[string]string{"status": "platform installed"})
}

func (s *Server) handlePlatformDown(w http.ResponseWriter, r *http.Request) {
	s.registry.Uninstall("monitoring", "grafana", s.exec)
	if s.cfg.MetricsProvider != "" {
		s.registry.Uninstall("monitoring", s.cfg.MetricsProvider, s.exec)
	}
	if s.cfg.IngressProvider != "" {
		s.registry.Uninstall("ingress", s.cfg.IngressProvider, s.exec)
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "platform removed"})
}

func (s *Server) handleListScenarios(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, s.scenes.Status())
}

func (s *Server) handleScenarioInfo(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	sc, err := s.scenes.Get(name)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, sc)
}

func (s *Server) handleScenarioUp(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.scenes.Up(name, s.exec); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("scenario up failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "activated", "scenario": name})
}

func (s *Server) handleScenarioDown(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.scenes.Down(name, s.exec); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("scenario down failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deactivated", "scenario": name})
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	svcs := s.svcs.List()
	type svcResp struct {
		Name string `json:"name"`
	}
	var result []svcResp
	for _, svc := range svcs {
		result = append(result, svcResp{Name: svc.Name})
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleServiceUp(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.svcs.Install(name, s.exec); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("service install failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "installed", "service": name})
}

func (s *Server) handleServiceDown(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.svcs.Uninstall(name, s.exec); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("service uninstall failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "uninstalled", "service": name})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Send periodic status updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			info, _ := k8s.GetClusterInfo(ctx)
			cancel()

			if err := conn.WriteJSON(map[string]interface{}{
				"type": "status",
				"data": info,
			}); err != nil {
				return
			}
		}
	}
}
