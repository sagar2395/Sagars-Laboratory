package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
		fmt.Fprintln(w, "NAME\tDISPLAY NAME\tCATEGORY\tSTATUS")
		for _, s := range scenarios {
			status := "inactive"
			if s.Active {
				status = "active"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.DisplayName, s.Category, status)
		}
		w.Flush()
		return nil
	},
}

var scenarioUpCmd = &cobra.Command{
	Use:   "up [scenario-name]",
	Short: "Activate a scenario",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return scenes.Up(args[0], exec)
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

		fmt.Printf("\nComponents (%d):\n", len(s.Components))
		for _, c := range s.Components {
			fmt.Printf("  - %s [%s]", c.Name, c.Type)
			if c.Chart != "" {
				fmt.Printf(" chart=%s", c.Chart)
			}
			if c.Namespace != "" {
				fmt.Printf(" ns=%s", c.Namespace)
			}
			fmt.Println()
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
	scenarioCmd.AddCommand(scenarioListCmd)
	scenarioCmd.AddCommand(scenarioUpCmd)
	scenarioCmd.AddCommand(scenarioDownCmd)
	scenarioCmd.AddCommand(scenarioStatusCmd)
	scenarioCmd.AddCommand(scenarioInfoCmd)
	rootCmd.AddCommand(scenarioCmd)
}
