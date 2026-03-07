package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Manage platform components",
}

func platformUpRun(cmd *cobra.Command, args []string) error {
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
		if err := reg.Install("monitoring", cfg.MetricsProvider, exec); err != nil {
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
	// Uninstall in reverse order
	fmt.Println("Uninstalling grafana...")
	_ = reg.Uninstall("monitoring", "grafana", exec)

	if cfg.MetricsProvider != "" {
		fmt.Printf("Uninstalling metrics (%s)...\n", cfg.MetricsProvider)
		_ = reg.Uninstall("monitoring", cfg.MetricsProvider, exec)
	}

	if cfg.IngressProvider != "" {
		fmt.Printf("Uninstalling ingress (%s)...\n", cfg.IngressProvider)
		_ = reg.Uninstall("ingress", cfg.IngressProvider, exec)
	}

	fmt.Println("\nPlatform uninstalled.")
	return nil
}

var platformUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Install all platform components",
	RunE:  platformUpRun,
}

var platformDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Uninstall all platform components",
	RunE:  platformDownRun,
}

var platformStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show platform component status",
	RunE: func(cmd *cobra.Command, args []string) error {
		categories := reg.Categories()
		if len(categories) == 0 {
			fmt.Println("No platform components found.")
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
