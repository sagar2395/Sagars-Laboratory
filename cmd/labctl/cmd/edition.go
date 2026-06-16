// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.flightdeck.dev/labctl/pkg/edition"
)

func editionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edition",
		Short: "Show the active Flightdeck edition and feature availability",
		Long: `Print the active Flightdeck edition (Community, Professional, or Enterprise)
and the associated feature availability table. The edition is set at deployment
time via the LABCTL_EDITION environment variable; it defaults to Community (CE).

Editions differ ONLY in entitled content and optional services — the engine
binary is identical for all editions. See docs/strategy/OSS-COMMERCIAL-STRATEGY.md.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ed := edition.Current()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Edition: %s\n", ed.DisplayName())
			fmt.Fprintf(out, "         %s\n\n", ed.Summary())

			features := ed.FeatureLine()

			// Sort feature names for stable output
			names := make([]string, 0, len(features))
			for k := range features {
				names = append(names, k)
			}
			sort.Strings(names)

			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "Feature\tAvailable")
			fmt.Fprintln(tw, "───────\t─────────")
			for _, name := range names {
				avail := "yes"
				if !features[name] {
					avail = "no  (premium/enterprise content)"
				}
				fmt.Fprintf(tw, "%s\t%s\n", name, avail)
			}
			tw.Flush()

			fmt.Fprintf(out, "\nSet LABCTL_EDITION=professional|enterprise to unlock premium content.\n")
			fmt.Fprintf(out, "CE features are permanent — no existing CE feature will ever move behind a paywall.\n")
			return nil
		},
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(editionCmd())
}
