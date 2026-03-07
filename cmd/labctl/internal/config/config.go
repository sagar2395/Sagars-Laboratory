package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the resolved project configuration.
type Config struct {
	// Project root directory
	ProjectRoot string

	// Cluster/Runtime
	Profile     string
	ClusterName string
	HTTPPort    string
	HTTPSPort   string

	// Runtime-specific
	IngressClass string
	StorageClass string
	DomainSuffix string
	RegistryType string

	// Provider selections
	IngressProvider string
	MetricsProvider string
	LoggingProvider string
	TracingProvider string
	GitOpsProvider  string
	ChaosProvider   string
	PolicyProvider  string
	SecretsProvider string

	// App defaults
	AppName         string
	HelmReleaseName string
	HelmValues      string
}

// AppConfig holds per-app configuration from app.env.
type AppConfig struct {
	AppName        string
	BuildStrategy  string
	DeployStrategy string
	HelmRelease    string
	HelmValues     string
	Namespace      string
}

// Load reads the project configuration from .env, versions.env, and runtime.env.
func Load(projectRoot string) (*Config, error) {
	if projectRoot == "" {
		var err error
		projectRoot, err = findProjectRoot()
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		ProjectRoot: projectRoot,
	}

	// Load .env (optional)
	loadEnvFile(filepath.Join(projectRoot, ".env"))

	// Load runtime.env based on profile
	profile := getEnvOrDefault("PROFILE", "k3d")
	runtimeEnv := filepath.Join(projectRoot, "runtimes", profile, "runtime.env")
	loadEnvFile(runtimeEnv)

	// Populate config from environment
	cfg.Profile = profile
	cfg.ClusterName = getEnvOrDefault("CLUSTER_NAME", "sagars-cluster")
	cfg.HTTPPort = getEnvOrDefault("HTTP_PORT", "80")
	cfg.HTTPSPort = getEnvOrDefault("HTTPS_PORT", "443")

	cfg.IngressClass = getEnvOrDefault("INGRESS_CLASS", "traefik")
	cfg.StorageClass = getEnvOrDefault("STORAGE_CLASS", "local-path")
	cfg.DomainSuffix = getEnvOrDefault("DOMAIN_SUFFIX", "k3d.local")
	cfg.RegistryType = getEnvOrDefault("REGISTRY_TYPE", "k3d-import")

	cfg.IngressProvider = getEnvOrDefault("INGRESS_PROVIDER", "traefik")
	cfg.MetricsProvider = getEnvOrDefault("METRICS_PROVIDER", "prometheus")
	cfg.LoggingProvider = getEnvOrDefault("LOGGING_PROVIDER", "")
	cfg.TracingProvider = getEnvOrDefault("TRACING_PROVIDER", "")
	cfg.GitOpsProvider = getEnvOrDefault("GITOPS_PROVIDER", "")
	cfg.ChaosProvider = getEnvOrDefault("CHAOS_PROVIDER", "")
	cfg.PolicyProvider = getEnvOrDefault("POLICY_PROVIDER", "")
	cfg.SecretsProvider = getEnvOrDefault("SECRETS_PROVIDER", "")

	cfg.AppName = getEnvOrDefault("APP_NAME", "go-api")
	cfg.HelmReleaseName = getEnvOrDefault("HELM_RELEASE_NAME", "go-api")
	cfg.HelmValues = getEnvOrDefault("HELM_VALUES", "values-dev.yaml")

	return cfg, nil
}

// LoadAppConfig reads app-specific config from apps/<name>/app.env.
func LoadAppConfig(projectRoot, appName string) (*AppConfig, error) {
	appEnv := filepath.Join(projectRoot, "apps", appName, "app.env")
	if _, err := os.Stat(appEnv); os.IsNotExist(err) {
		return nil, fmt.Errorf("app config not found: %s", appEnv)
	}

	v := viper.New()
	v.SetConfigFile(appEnv)
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading app config: %w", err)
	}

	return &AppConfig{
		AppName:        v.GetString("APP_NAME"),
		BuildStrategy:  v.GetString("BUILD_STRATEGY"),
		DeployStrategy: v.GetString("DEPLOY_STRATEGY"),
		HelmRelease:    v.GetString("HELM_RELEASE_NAME"),
		HelmValues:     v.GetString("HELM_VALUES"),
		Namespace:      v.GetString("NAMESPACE"),
	}, nil
}

// ListApps returns a list of app names from the apps/ directory.
func ListApps(projectRoot string) ([]string, error) {
	appsDir := filepath.Join(projectRoot, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, fmt.Errorf("reading apps directory: %w", err)
	}

	var apps []string
	for _, e := range entries {
		if e.IsDir() {
			appEnv := filepath.Join(appsDir, e.Name(), "app.env")
			if _, err := os.Stat(appEnv); err == nil {
				apps = append(apps, e.Name())
			}
		}
	}
	return apps, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "engine")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (looked for Makefile + engine/)")
		}
		dir = parent
	}
}

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip inline comments
		if idx := strings.Index(val, " #"); idx >= 0 {
			val = strings.TrimSpace(val[:idx])
		}
		// Don't override existing env vars
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return defaultVal
}
