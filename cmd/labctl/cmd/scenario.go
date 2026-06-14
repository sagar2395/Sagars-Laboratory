// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sagars-lab/labctl/pkg/checks"
	scenariopkg "github.com/sagars-lab/labctl/internal/scenario"
	"github.com/spf13/cobra"
)

var scenarioCmd = &cobra.Command{
	Use:   "scenario",
	Short: "Manage lab scenarios",
}

var scenarioListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scenarios",
	RunE: func(cmd *cobra.Command, args []string) error {
		scenarios := scenes.List()
		if len(scenarios) == 0 {
			fmt.Println("No scenarios found in scenarios/ directory.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tDISPLAY NAME\tCATEGORY\tSOURCE\tSTATUS")
		for _, s := range scenarios {
			status := "inactive"
			if s.Active {
				status = "active"
			}
			source := s.Source
			if source == "" {
				source = "repo"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Name, s.DisplayName, s.Category, source, status)
		}
		w.Flush()

		if errs := scenes.LoadErrors(); len(errs) > 0 {
			fmt.Fprintf(os.Stderr, "\nWarning: %d scenario(s) failed to load (run with -v for details):\n", len(errs))
			for key, err := range errs {
				fmt.Fprintf(os.Stderr, "  %s: %v\n", key, err)
			}
		}
		return nil
	},
}

var scenarioUpCmd = &cobra.Command{
	Use:   "up [scenario-name]",
	Short: "Activate a scenario",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := scenes.Up(args[0], exec)
		if errors.Is(err, scenariopkg.ErrAlreadyActive) {
			fmt.Fprintf(os.Stderr, "Scenario %s is already active. Use --force to reinstall.\n", args[0])
			return nil
		}
		return err
	},
}

var scenarioDownCmd = &cobra.Command{
	Use:   "down [scenario-name]",
	Short: "Deactivate a scenario",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return scenes.Down(args[0], exec)
	},
}

var scenarioStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show scenario status",
	RunE: func(cmd *cobra.Command, args []string) error {
		statuses := scenes.Status()
		if len(statuses) == 0 {
			fmt.Println("No scenarios found.")
			return nil
		}

		hasActive := false
		for _, s := range statuses {
			if s.Active {
				hasActive = true
				break
			}
		}

		if !hasActive {
			fmt.Println("No scenarios are currently active.")
			fmt.Println("Use 'labctl scenario list' to see available scenarios.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tSTATUS")
		for _, s := range statuses {
			if s.Active {
				fmt.Fprintf(w, "%s\t%s\tactive\n", s.Name, s.Category)
			}
		}
		w.Flush()
		return nil
	},
}

var (
	verifyWatch        bool
	verifyInterval     time.Duration
	verifyTimeout      time.Duration
	verifyCheckTimeout time.Duration
)

var scenarioVerifyCmd = &cobra.Command{
	Use:   "verify [scenario-name]",
	Short: "Run a scenario's checks and report pass/fail",
	Long: `Runs the machine-verifiable checks declared in the scenario's
checks block (scenario format v2) and reports each result. Exits non-zero
if any check fails, so it is safe to use in CI and scripts.

With --watch, checks are re-run every --interval until they all pass or
--timeout elapses — useful right after 'scenario up' while pods settle.`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true, // a failing check is a result, not a usage error
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := newCheckRunner()
		ctx := context.Background()
		deadline := time.Now().Add(verifyTimeout)

		for {
			results, err := scenes.Verify(ctx, args[0], runner)
			if err != nil {
				return err
			}
			printCheckResults(results)
			if checks.AllPass(results) {
				fmt.Printf("All %d checks passed.\n", len(results))
				return nil
			}
			failed := 0
			for _, r := range results {
				if !r.Pass {
					failed++
				}
			}
			if !verifyWatch || time.Now().After(deadline) {
				return fmt.Errorf("%d of %d checks failed", failed, len(results))
			}
			fmt.Printf("\n%d of %d checks failing — retrying in %s (until %s)...\n\n",
				failed, len(results), verifyInterval, deadline.Format("15:04:05"))
			time.Sleep(verifyInterval)
		}
	},
}

// newCheckRunner builds a check runner wired to the lab's config: the
// Prometheus endpoint (PROMETHEUS_URL env override, else the ingress
// hostname) and the standard script environment.
func newCheckRunner() *checks.Runner {
	r := checks.NewRunner()
	r.DefaultTimeout = verifyCheckTimeout
	promURL := os.Getenv("PROMETHEUS_URL")
	if promURL == "" {
		promURL = "http://prometheus." + cfg.DomainSuffix
	}
	r.PrometheusURL = promURL
	r.Env = []string{
		"DOMAIN_SUFFIX=" + cfg.DomainSuffix,
		"MONITORING_NAMESPACE=" + cfg.MonitoringNamespace,
		"PROJECT_ROOT=" + cfg.ProjectRoot,
	}
	return r
}

func printCheckResults(results []checks.Result) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, r := range results {
		mark := "PASS"
		if !r.Pass {
			mark = "FAIL"
		}
		detail := ""
		if r.Got != "" || r.Want != "" {
			detail = fmt.Sprintf("got: %s, want: %s", orDash(r.Got), orDash(r.Want))
		}
		if r.Error != "" {
			detail = "error: " + r.Error
		}
		fmt.Fprintf(w, "%s\t%s\t(%s)\t%s\t%dms\n", mark, r.Name, r.Type, detail, r.DurationMS)
	}
	w.Flush()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

var (
	packName  string
	packForce bool
)

var scenarioInstallCmd = &cobra.Command{
	Use:   "install <git-url>[@ref]",
	Short: "Install a scenario pack from a git repository",
	Long: `Clones a scenario pack into .labctl/catalog/ and validates every
scenario in it against the schema before it becomes visible. A pack is a
git repo containing one scenario (scenario.yaml at the root) or a
collection of scenario directories.

SECURITY: packs run scripts and apply manifests on your cluster — only
install sources you trust. Packs never auto-update; reinstall with --force
to upgrade.`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pack, err := scenes.InstallPack(args[0], packName, packForce, nil)
		if err != nil {
			return err
		}
		fmt.Printf("Installed pack %q with %d scenario(s):\n", pack.Name, len(pack.Scenarios))
		for _, s := range pack.Scenarios {
			fmt.Printf("  - %s\n", s)
		}
		fmt.Printf("\nActivate with: labctl scenario up <name>\nRemove with:   labctl scenario uninstall %s\n", pack.Name)
		return nil
	},
}

var scenarioUninstallCmd = &cobra.Command{
	Use:   "uninstall <pack-name>",
	Short: "Remove an installed scenario pack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scenes.UninstallPack(args[0]); err != nil {
			return err
		}
		fmt.Printf("Pack %q removed.\n", args[0])
		return nil
	},
}

var scenarioPacksCmd = &cobra.Command{
	Use:   "packs",
	Short: "List installed scenario packs",
	RunE: func(cmd *cobra.Command, args []string) error {
		packs := scenes.Packs()
		if len(packs) == 0 {
			fmt.Println("No packs installed. Install one with: labctl scenario install <git-url>")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "PACK\tSCENARIOS")
		for _, p := range packs {
			fmt.Fprintf(w, "%s\t%s\n", p.Name, strings.Join(p.Scenarios, ", "))
		}
		w.Flush()
		return nil
	},
}

var scenarioInfoCmd = &cobra.Command{
	Use:   "info [scenario-name]",
	Short: "Show detailed information about a scenario",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := scenes.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Name:        %s\n", s.Name)
		fmt.Printf("Display:     %s\n", s.DisplayName)
		fmt.Printf("Category:    %s\n", s.Category)
		fmt.Printf("Description: %s\n", s.Description)

		status := "inactive"
		if s.Active {
			status = "active"
		}
		fmt.Printf("Status:      %s\n", status)

		if len(s.Prerequisites.Platform) > 0 {
			fmt.Printf("\nPrerequisites (platform):\n")
			for _, p := range s.Prerequisites.Platform {
				fmt.Printf("  - %s\n", p)
			}
		}
		if len(s.Prerequisites.Apps) > 0 {
			fmt.Printf("\nPrerequisites (apps):\n")
			for _, a := range s.Prerequisites.Apps {
				fmt.Printf("  - %s\n", a)
			}
		}

		if len(s.Objectives) > 0 {
			fmt.Printf("\nObjectives:\n")
			for _, o := range s.Objectives {
				fmt.Printf("  - %s\n", o)
			}
		}

		printComponent := func(c scenariopkg.Component, indent string) {
			fmt.Printf("%s- %s [%s]", indent, c.Name, c.Type)
			if c.Chart != "" {
				fmt.Printf(" chart=%s", c.Chart)
			}
			if c.Namespace != "" {
				fmt.Printf(" ns=%s", c.Namespace)
			}
			fmt.Println()
		}

		if len(s.Stages) > 0 {
			fmt.Printf("\nStages (%d):\n", len(s.Stages))
			for _, st := range s.Stages {
				fmt.Printf("  %s:\n", st.Name)
				for _, c := range st.Components {
					printComponent(c, "    ")
				}
			}
		} else {
			fmt.Printf("\nComponents (%d):\n", len(s.Components))
			for _, c := range s.Components {
				printComponent(c, "  ")
			}
		}

		if len(s.Checks) > 0 {
			fmt.Printf("\nChecks (%d) — run 'labctl scenario verify %s':\n", len(s.Checks), s.Name)
			for _, c := range s.Checks {
				fmt.Printf("  - %s [%s]\n", c.Name, c.Type)
			}
		}

		if len(s.Explore.URLs) > 0 || len(s.Explore.Commands) > 0 {
			fmt.Println("\nExplore:")
			for _, u := range s.Explore.URLs {
				fmt.Printf("  URL: %-25s %s\n", u.Label, scenes.ResolveTemplate(u.URL))
			}
			for _, c := range s.Explore.Commands {
				fmt.Printf("  CMD: %s\n       %s\n", c.Label, scenes.ResolveTemplate(c.Command))
			}
		}

		return nil
	},
}

func init() {
	scenarioVerifyCmd.Flags().BoolVar(&verifyWatch, "watch", false, "re-run checks until they all pass or --timeout elapses")
	scenarioVerifyCmd.Flags().DurationVar(&verifyInterval, "interval", 10*time.Second, "delay between re-runs in --watch mode")
	scenarioVerifyCmd.Flags().DurationVar(&verifyTimeout, "timeout", 5*time.Minute, "overall deadline in --watch mode")
	scenarioVerifyCmd.Flags().DurationVar(&verifyCheckTimeout, "check-timeout", 30*time.Second, "per-check timeout")

	scenarioInstallCmd.Flags().StringVar(&packName, "name", "", "pack name (default: repository basename)")
	scenarioInstallCmd.Flags().BoolVar(&packForce, "force", false, "replace the pack if it is already installed")

	scenarioCmd.AddCommand(scenarioListCmd)
	scenarioCmd.AddCommand(scenarioUpCmd)
	scenarioCmd.AddCommand(scenarioDownCmd)
	scenarioCmd.AddCommand(scenarioStatusCmd)
	scenarioCmd.AddCommand(scenarioInfoCmd)
	scenarioCmd.AddCommand(scenarioVerifyCmd)
	scenarioCmd.AddCommand(scenarioInstallCmd)
	scenarioCmd.AddCommand(scenarioUninstallCmd)
	scenarioCmd.AddCommand(scenarioPacksCmd)
	rootCmd.AddCommand(scenarioCmd)
}
