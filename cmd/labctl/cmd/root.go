package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sagars-lab/labctl/internal/config"
	"github.com/sagars-lab/labctl/internal/executor"
	"github.com/sagars-lab/labctl/internal/platform"
	"github.com/sagars-lab/labctl/internal/runtime"
	"github.com/sagars-lab/labctl/internal/scenario"
	"github.com/sagars-lab/labctl/internal/services"
)

var (
	projectDir string
	verbose    bool

	cfg    *config.Config
	exec   *executor.Executor
	reg    *platform.Registry
	scenes *scenario.Engine
	svcReg *services.Registry
	rtm    *runtime.Manager
)

var rootCmd = &cobra.Command{
	Use:   "labctl",
	Short: "Sagars-Laboratory controller",
	Long:  `labctl is the CLI and web UI for managing your Platform Engineering homelab.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip init for completion and help commands
		if cmd.Name() == "completion" || cmd.Name() == "help" {
			return nil
		}

		var err error
		cfg, err = config.Load(projectDir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		exec = executor.New(cfg.ProjectRoot)
		// Propagate resolved config values so all child scripts inherit them.
		exec.SetEnv("CLUSTER_NAME", cfg.ClusterName)
		exec.SetEnv("DOMAIN_SUFFIX", cfg.DomainSuffix)
		exec.SetEnv("HTTP_PORT", cfg.HTTPPort)
		exec.SetEnv("HTTPS_PORT", cfg.HTTPSPort)
		exec.SetEnv("INGRESS_CLASS", cfg.IngressClass)
		exec.SetEnv("STORAGE_CLASS", cfg.StorageClass)
		exec.SetEnv("PROFILE", cfg.Profile)
		reg = platform.NewRegistry(cfg.ProjectRoot)
		scenes = scenario.NewEngine(cfg.ProjectRoot, cfg.DomainSuffix)
		svcReg = services.NewRegistry(cfg.ProjectRoot)
		rtm = runtime.NewManager(cfg.ProjectRoot, cfg.ClusterName)
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
