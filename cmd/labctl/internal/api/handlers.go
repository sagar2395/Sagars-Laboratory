package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sagars-lab/labctl/internal/config"
	"github.com/sagars-lab/labctl/internal/executor"
	"github.com/sagars-lab/labctl/internal/k8s"
)

// StatusResponse represents the overall lab status.
type StatusResponse struct {
	DomainSuffix string             `json:"domainSuffix"`
	Cluster      *k8s.ClusterInfo   `json:"cluster"`
	Platform     PlatformStatusResp `json:"platform"`
	Apps         []AppStatusResp    `json:"apps"`
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

	resp := StatusResponse{
		DomainSuffix: s.cfg.DomainSuffix,
	}

	// Cluster info
	clusterInfo, _ := k8s.GetClusterInfo(ctx)
	resp.Cluster = clusterInfo

	// Platform status — derive namespace from the registry so we don't
	// assume namespace == provider name.
	ingressActive := false
	if p, err := s.registry.GetProvider("ingress", s.cfg.IngressProvider); err == nil {
		ingressActive = k8s.NamespaceExists(ctx, p.Namespace())
	}
	metricsActive := false
	if p, err := s.registry.GetProvider("monitoring/metrics", s.cfg.MetricsProvider); err == nil {
		metricsActive = k8s.NamespaceExists(ctx, p.Namespace())
	}
	loggingActive := false
	if p, err := s.registry.GetProvider("logging", s.cfg.LoggingProvider); err == nil {
		loggingActive = k8s.ServiceExists(ctx, p.Namespace(), "loki-gateway")
	}
	tracingActive := false
	if p, err := s.registry.GetProvider("tracing", s.cfg.TracingProvider); err == nil {
		tracingActive = k8s.ServiceExists(ctx, p.Namespace(), "tempo")
	}
	resp.Platform = PlatformStatusResp{
		Ingress: ComponentStatus{
			Provider: s.cfg.IngressProvider,
			Active:   ingressActive,
		},
		Metrics: ComponentStatus{
			Provider: s.cfg.MetricsProvider,
			Active:   metricsActive,
		},
		Logging: ComponentStatus{
			Provider: s.cfg.LoggingProvider,
			Active:   loggingActive,
		},
		Tracing: ComponentStatus{
			Provider: s.cfg.TracingProvider,
			Active:   tracingActive,
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
	go func() {
		s.exec.RunScriptStreamed(fmt.Sprintf("Deploy %s", name), "engine/deploy.sh", "deploy", name)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "deploy", "app": name})
}

func (s *Server) handleAppDestroy(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		s.exec.RunScriptStreamed(fmt.Sprintf("Destroy %s", name), "engine/deploy.sh", "destroy", name)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "destroy", "app": name})
}

func (s *Server) handleAppBuild(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		s.exec.RunScriptStreamed(fmt.Sprintf("Build %s", name), "engine/build.sh", name)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "build", "app": name})
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
				"installed": k8s.NamespaceExists(ctx, p.Namespace()),
			}
			result[cat] = append(result[cat], entry)
		}
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handlePlatformUp(w http.ResponseWriter, r *http.Request) {
	go func() {
		// Install ingress
		if s.cfg.IngressProvider != "" {
			s.registry.InstallStreamed("ingress", s.cfg.IngressProvider, s.exec)
		}
		// Install metrics
		if s.cfg.MetricsProvider != "" {
			s.registry.InstallStreamed("monitoring/metrics", s.cfg.MetricsProvider, s.exec)
		}
		// Install grafana
		s.registry.InstallStreamed("monitoring", "grafana", s.exec)
		// Install logging
		if s.cfg.LoggingProvider != "" {
			s.registry.InstallStreamed("logging", s.cfg.LoggingProvider, s.exec)
		}
		// Install tracing
		if s.cfg.TracingProvider != "" {
			s.registry.InstallStreamed("tracing", s.cfg.TracingProvider, s.exec)
		}
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "platform-up"})
}

func (s *Server) handlePlatformDown(w http.ResponseWriter, r *http.Request) {
	go func() {
		// Uninstall in reverse order
		if s.cfg.TracingProvider != "" {
			s.registry.UninstallStreamed("tracing", s.cfg.TracingProvider, s.exec)
		}
		if s.cfg.LoggingProvider != "" {
			s.registry.UninstallStreamed("logging", s.cfg.LoggingProvider, s.exec)
		}
		s.registry.UninstallStreamed("monitoring", "grafana", s.exec)
		if s.cfg.MetricsProvider != "" {
			s.registry.UninstallStreamed("monitoring/metrics", s.cfg.MetricsProvider, s.exec)
		}
		if s.cfg.IngressProvider != "" {
			s.registry.UninstallStreamed("ingress", s.cfg.IngressProvider, s.exec)
		}
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "platform-down"})
}

func (s *Server) handleComponentUp(w http.ResponseWriter, r *http.Request) {
	category := mux.Vars(r)["category"]
	name := mux.Vars(r)["name"]
	go func() {
		s.registry.InstallStreamed(category, name, s.exec)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": fmt.Sprintf("install %s/%s", category, name)})
}

func (s *Server) handleComponentDown(w http.ResponseWriter, r *http.Request) {
	category := mux.Vars(r)["category"]
	name := mux.Vars(r)["name"]
	go func() {
		s.registry.UninstallStreamed(category, name, s.exec)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": fmt.Sprintf("uninstall %s/%s", category, name)})
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

	// Resolve template variables in explore URLs and commands
	for i := range sc.Explore.URLs {
		sc.Explore.URLs[i].URL = s.scenes.ResolveTemplate(sc.Explore.URLs[i].URL)
	}
	for i := range sc.Explore.Commands {
		sc.Explore.Commands[i].Command = s.scenes.ResolveTemplate(sc.Explore.Commands[i].Command)
	}

	respondJSON(w, http.StatusOK, sc)
}

func (s *Server) handleScenarioUp(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		if err := s.scenes.Up(name, s.exec); err != nil {
			exitCode := 1
			s.exec.Broadcast.Send(executor.ActionEvent{
				ID:        fmt.Sprintf("scenario-up-%s", name),
				Type:      "action_end",
				Action:    fmt.Sprintf("Activate scenario: %s", name),
				ExitCode:  &exitCode,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
		} else {
			exitCode := 0
			s.exec.Broadcast.Send(executor.ActionEvent{
				ID:        fmt.Sprintf("scenario-up-%s", name),
				Type:      "action_end",
				Action:    fmt.Sprintf("Activate scenario: %s", name),
				ExitCode:  &exitCode,
				Timestamp: time.Now(),
			})
		}
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "scenario-up", "scenario": name})
}

func (s *Server) handleScenarioDown(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		if err := s.scenes.Down(name, s.exec); err != nil {
			exitCode := 1
			s.exec.Broadcast.Send(executor.ActionEvent{
				ID:        fmt.Sprintf("scenario-down-%s", name),
				Type:      "action_end",
				Action:    fmt.Sprintf("Deactivate scenario: %s", name),
				ExitCode:  &exitCode,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
		} else {
			exitCode := 0
			s.exec.Broadcast.Send(executor.ActionEvent{
				ID:        fmt.Sprintf("scenario-down-%s", name),
				Type:      "action_end",
				Action:    fmt.Sprintf("Deactivate scenario: %s", name),
				ExitCode:  &exitCode,
				Timestamp: time.Now(),
			})
		}
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "scenario-down", "scenario": name})
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
	go func() {
		s.svcs.Install(name, s.exec)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "service-up", "service": name})
}

func (s *Server) handleServiceDown(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		s.svcs.Uninstall(name, s.exec)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "service-down", "service": name})
}

// DashboardURL represents a link to a platform dashboard.
type DashboardURL struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	URL       string `json:"url"`
	Available bool   `json:"available"`
	Category  string `json:"category"`
}

func (s *Server) handleDashboardURLs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	domain := s.cfg.DomainSuffix
	var dashboards []DashboardURL

	if k8s.NamespaceExists(ctx, "monitoring") {
		dashboards = append(dashboards, DashboardURL{
			Name: "grafana", Label: "Grafana",
			URL: fmt.Sprintf("http://grafana.%s", domain), Available: true, Category: "monitoring",
		})
		dashboards = append(dashboards, DashboardURL{
			Name: "prometheus", Label: "Prometheus",
			URL: fmt.Sprintf("http://prometheus.%s", domain), Available: true, Category: "monitoring",
		})
	}

	if s.cfg.IngressProvider == "traefik" && k8s.NamespaceExists(ctx, "traefik") {
		dashboards = append(dashboards, DashboardURL{
			Name: "traefik", Label: "Traefik Dashboard",
			URL: fmt.Sprintf("http://traefik.%s/dashboard/", domain), Available: true, Category: "ingress",
		})
	}

	if k8s.NamespaceExists(ctx, "kubernetes-dashboard") {
		dashboards = append(dashboards, DashboardURL{
			Name: "kubernetes-dashboard", Label: "Kubernetes Dashboard",
			URL: fmt.Sprintf("http://dashboard.%s", domain), Available: true, Category: "cluster",
		})
	}

	if k8s.NamespaceExists(ctx, "argocd") {
		dashboards = append(dashboards, DashboardURL{
			Name: "argocd", Label: "ArgoCD",
			URL: fmt.Sprintf("http://argocd.%s", domain), Available: true, Category: "gitops",
		})
	}

	if k8s.NamespaceExists(ctx, "chaos-mesh") {
		dashboards = append(dashboards, DashboardURL{
			Name: "chaos-mesh", Label: "Chaos Mesh",
			URL: "http://localhost:2333", Available: true, Category: "chaos",
		})
	}

	if k8s.ServiceExists(ctx, "monitoring", "loki-gateway") {
		dashboards = append(dashboards, DashboardURL{
			Name: "loki", Label: "Logs (Loki)",
			URL:       fmt.Sprintf("http://grafana.%s/explore?orgId=1&left=%%7B%%22datasource%%22:%%22Loki%%22%%7D", domain),
			Available: true, Category: "monitoring",
		})
	}

	if k8s.ServiceExists(ctx, "monitoring", "tempo") {
		dashboards = append(dashboards, DashboardURL{
			Name: "tempo", Label: "Traces (Tempo)",
			URL:       fmt.Sprintf("http://grafana.%s/explore?orgId=1&left=%%7B%%22datasource%%22:%%22Tempo%%22%%7D", domain),
			Available: true, Category: "monitoring",
		})
	}

	respondJSON(w, http.StatusOK, dashboards)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Subscribe to action events
	actionCh := s.exec.Broadcast.Subscribe()
	defer s.exec.Broadcast.Unsubscribe(actionCh)

	// Periodic status updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Read pump — discard incoming messages, detect close
	closeCh := make(chan struct{})
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(closeCh)
				return
			}
		}
	}()

	for {
		select {
		case <-closeCh:
			return
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
		case event := <-actionCh:
			if err := conn.WriteJSON(map[string]interface{}{
				"type": "action",
				"data": event,
			}); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleListRuntimes(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, s.runtimes.List())
}

func (s *Server) handleRuntimeActivate(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		s.runtimes.Activate(name, s.exec)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "runtime-activate", "runtime": name})
}

func (s *Server) handleRuntimeDeactivate(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	go func() {
		s.runtimes.Deactivate(name, s.exec)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"status": "started", "action": "runtime-deactivate", "runtime": name})
}
