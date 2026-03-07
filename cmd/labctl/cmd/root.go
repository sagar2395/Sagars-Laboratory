package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sagars-lab/labctl/internal/config"
	"github.com/sagars-lab/labctl/internal/executor"
	"github.com/sagars-lab/labctl/internal/platform"
	"github.com/sagars-lab/labctl/internal/scenario"
	"github.com/sagars-lab/labctl/internal/services"
)

var (
	cfgFile    string
	projectDir string
	verbose    bool

	cfg     *config.Config
	exec    *executor.Executor
	reg     *platform.Registry
	scenes  *scenario.Engine
	svcReg  *services.Registry
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
		reg = platform.NewRegistry(cfg.ProjectRoot)
		scenes = scenario.NewEngine(cfg.ProjectRoot, cfg.DomainSuffix)
		svcReg = services.NewRegistry(cfg.ProjectRoot)
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
