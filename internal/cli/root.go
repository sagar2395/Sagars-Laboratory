// SPDX-License-Identifier: Apache-2.0
package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"go.flightdeck.dev/flightdeck/internal/config"
	"go.flightdeck.dev/flightdeck/internal/executor"
	"go.flightdeck.dev/flightdeck/internal/incident"
	"go.flightdeck.dev/flightdeck/internal/platform"
	"go.flightdeck.dev/flightdeck/internal/runtime"
	"go.flightdeck.dev/flightdeck/internal/scenario"
	"go.flightdeck.dev/flightdeck/internal/services"
)

var (
	projectDir string
	verbose    bool

	cfg    *config.Config
	exec   *executor.Executor
	reg    *platform.Registry
	scenes *scenario.Engine
	incEng *incident.Engine
	svcReg *services.Registry
	rtm    *runtime.Manager
)

var rootCmd = &cobra.Command{
	Use:   "labctl",
	Short: "Flightdeck — platform engineering simulator control plane",
	Long:  `labctl is the CLI, web UI, and API for Flightdeck, a Kubernetes platform engineering simulator.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip init for commands that must work when the environment is
		// broken. doctor in particular exists to diagnose exactly the
		// situations that would make this initialisation fail.
		switch cmd.Name() {
		case "completion", "help", "doctor", "runs", "list", "logs", "cancel":
			return nil
		}

		// Configure log level before doing anything else so debug output is visible.
		logLevel := slog.LevelWarn
		if verbose {
			logLevel = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

		slog.Debug("loading config", "projectDir", projectDir)
		var err error
		cfg, err = config.Load(projectDir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		slog.Debug("config loaded", "root", cfg.ProjectRoot, "profile", cfg.Profile, "cluster", cfg.ClusterName)

		exec = executor.New(cfg.ProjectRoot)
		// Propagate resolved config values so all child scripts inherit them.
		exec.SetEnv("CLUSTER_NAME", cfg.ClusterName)
		exec.SetEnv("DOMAIN_SUFFIX", cfg.DomainSuffix)
		exec.SetEnv("HTTP_PORT", cfg.HTTPPort)
		exec.SetEnv("HTTPS_PORT", cfg.HTTPSPort)
		exec.SetEnv("INGRESS_CLASS", cfg.IngressClass)
		exec.SetEnv("STORAGE_CLASS", cfg.StorageClass)
		exec.SetEnv("PROFILE", cfg.Profile)
		exec.SetEnv("MONITORING_NAMESPACE", cfg.MonitoringNamespace)
		reg = platform.NewRegistryWithNamespace(cfg.ProjectRoot, cfg.MonitoringNamespace)
		scenes = scenario.NewEngine(cfg.ProjectRoot, cfg.DomainSuffix, cfg.Profile)
		scenes.MonitoringNamespace = cfg.MonitoringNamespace
		incEng = incident.NewEngine(cfg.ProjectRoot, cfg.DomainSuffix)
		incEng.AlertmanagerURL = os.Getenv("ALERTMANAGER_URL")
		if incEng.AlertmanagerURL == "" {
			incEng.AlertmanagerURL = "http://alertmanager." + cfg.DomainSuffix
		}
		svcReg = services.NewRegistry(cfg.ProjectRoot)
		rtm = runtime.NewManager(cfg.ProjectRoot, cfg.ClusterName)
		slog.Debug("registries initialised", "runtimes", rtm.Names())
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&projectDir, "project-dir", "", "project root directory (auto-detected if not set)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug-level logging (config load, script exec, API calls)")

	rootCmd.AddCommand(learnCmd())
	rootCmd.AddCommand(challengeCmd())
}
