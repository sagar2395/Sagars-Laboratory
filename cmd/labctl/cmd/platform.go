package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Manage platform components",
}

// providerForCategory returns the configured provider name for a platform
// category (matching the env-var selection), or "" if none is selected.
func providerForCategory(category string) string {
	switch category {
	case "ingress":
		return cfg.IngressProvider
	case "monitoring/metrics":
		return cfg.MetricsProvider
	case "logging":
		return cfg.LoggingProvider
	case "tracing":
		return cfg.TracingProvider
	case "gitops":
		return cfg.GitOpsProvider
	case "chaos":
		return cfg.ChaosProvider
	case "security/policy":
		return cfg.PolicyProvider
	case "security/secrets":
		return cfg.SecretsProvider
	case "mesh":
		return cfg.MeshProvider
	default:
		return ""
	}
}

// resolveProvider picks the provider to act on for a single category. It prefers
// the configured selection; if none is set it falls back to the registry — using
// the sole provider when a category has exactly one, or erroring with the choices
// when it has several (e.g. mesh: istio|linkerd).
func resolveProvider(category string) (string, error) {
	if p := providerForCategory(category); p != "" {
		return p, nil
	}
	providers := reg.GetProviders(category)
	if len(providers) == 0 {
		return "", fmt.Errorf("unknown platform category %q", category)
	}
	if len(providers) == 1 {
		return providers[0].Name, nil
	}
	var names []string
	for _, p := range providers {
		names = append(names, p.Name)
	}
	envVar := strings.ToUpper(strings.ReplaceAll(category, "/", "_")) + "_PROVIDER"
	return "", fmt.Errorf("category %q has multiple providers (%s); select one with %s",
		category, strings.Join(names, ", "), envVar)
}

func platformUpRun(cmd *cobra.Command, args []string) error {
	// Per-category install: `labctl platform up <category>`.
	if len(args) == 1 {
		category := args[0]
		provider, err := resolveProvider(category)
		if err != nil {
			return err
		}
		fmt.Printf("Installing %s (%s)...\n", category, provider)
		if err := reg.Install(category, provider, exec); err != nil {
			return fmt.Errorf("%s install failed: %w", category, err)
		}
		fmt.Printf("\n%s installed successfully.\n", category)
		return nil
	}

	// Install ingress
	if cfg.IngressProvider != "" {
		fmt.Printf("Installing ingress (%s)...\n", cfg.IngressProvider)
		if err := reg.Install("ingress", cfg.IngressProvider, exec); err != nil {
			return fmt.Errorf("ingress install failed: %w", err)
		}
	}

	// Install monitoring (metrics)
	if cfg.MetricsProvider != "" {
		fmt.Printf("Installing metrics (%s)...\n", cfg.MetricsProvider)
		if err := reg.Install("monitoring/metrics", cfg.MetricsProvider, exec); err != nil {
			fmt.Printf("Warning: metrics install: %v\n", err)
		}
	}

	// Install grafana (visualization)
	fmt.Println("Installing grafana...")
	if err := reg.Install("monitoring", "grafana", exec); err != nil {
		fmt.Printf("Warning: grafana install: %v\n", err)
	}

	fmt.Println("\nPlatform installed successfully.")
	return nil
}

func platformDownRun(cmd *cobra.Command, args []string) error {
	// Per-category uninstall: `labctl platform down <category>`.
	if len(args) == 1 {
		category := args[0]
		provider, err := resolveProvider(category)
		if err != nil {
			return err
		}
		fmt.Printf("Uninstalling %s (%s)...\n", category, provider)
		if err := reg.Uninstall(category, provider, exec); err != nil {
			return fmt.Errorf("%s uninstall failed: %w", category, err)
		}
		fmt.Printf("\n%s uninstalled.\n", category)
		return nil
	}

	// Uninstall in reverse order
	fmt.Println("Uninstalling grafana...")
	_ = reg.Uninstall("monitoring", "grafana", exec)

	if cfg.MetricsProvider != "" {
		fmt.Printf("Uninstalling metrics (%s)...\n", cfg.MetricsProvider)
		_ = reg.Uninstall("monitoring/metrics", cfg.MetricsProvider, exec)
	}

	if cfg.IngressProvider != "" {
		fmt.Printf("Uninstalling ingress (%s)...\n", cfg.IngressProvider)
		_ = reg.Uninstall("ingress", cfg.IngressProvider, exec)
	}

	fmt.Println("\nPlatform uninstalled.")
	return nil
}

var platformUpCmd = &cobra.Command{
	Use:   "up [category]",
	Short: "Install all platform components, or a single category (e.g. mesh)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  platformUpRun,
}

var platformDownCmd = &cobra.Command{
	Use:   "down [category]",
	Short: "Uninstall all platform components, or a single category (e.g. mesh)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  platformDownRun,
}

var platformStatusCmd = &cobra.Command{
	Use:   "status [category]",
	Short: "Show platform component status (all, or a single category)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		categories := reg.Categories()
		if len(categories) == 0 {
			fmt.Println("No platform components found.")
			return nil
		}

		// Per-category status: `labctl platform status <category>`.
		if len(args) == 1 {
			cat := args[0]
			providers := reg.GetProviders(cat)
			if len(providers) == 0 {
				return fmt.Errorf("unknown platform category %q", cat)
			}
			for _, p := range providers {
				if p.HasScript("status.sh") {
					fmt.Printf("--- %s/%s ---\n", cat, p.Name)
					_ = reg.Status(cat, p.Name, exec)
					fmt.Println()
				}
			}
			return nil
		}

		for _, cat := range categories {
			providers := reg.GetProviders(cat)
			for _, p := range providers {
				if p.HasScript("status.sh") {
					fmt.Printf("--- %s/%s ---\n", cat, p.Name)
					_ = reg.Status(cat, p.Name, exec)
					fmt.Println()
				}
			}
		}
		return nil
	},
}

func init() {
	platformCmd.AddCommand(platformUpCmd)
	platformCmd.AddCommand(platformDownCmd)
	platformCmd.AddCommand(platformStatusCmd)
	rootCmd.AddCommand(platformCmd)
}
