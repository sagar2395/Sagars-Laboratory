// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sagars-lab/labctl/internal/scenario"
	"github.com/sagars-lab/labctl/pkg/pack"
	"github.com/spf13/cobra"
)

// `labctl pack` is the first-class command group for scenario packs — content
// add-ons that bundle scenarios (and later platform modules / incidents). It
// wraps the same engine machinery as `labctl scenario install|packs|uninstall`,
// which remain as aliases.

var (
	packAddName  string
	packAddForce bool

	packAddRequireSig bool
	packAddCosignKey  string
	packAddCertID     string
	packAddCertIssuer string

	packPublishSign      bool
	packPublishCosignKey string
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Manage scenario packs (content add-ons)",
	Long: `Manage scenario packs — versioned bundles of scenarios distributed from a
git source (OCI registries and a searchable index land in later releases).

A pack may carry a pack.yaml manifest (apiVersion packs.flightdeck.dev/v1) with
its name, version, publisher, license, tier, and engine compatibility. Packs run
scripts and apply manifests on your cluster — only install sources you trust.`,
}

var packAddCmd = &cobra.Command{
	Use:   "add <git-url>[@ref] | oci://<registry>/<repo>[:tag]",
	Short: "Install a scenario pack from a git source or OCI registry",
	Long: `Install a scenario pack from a git source or an OCI registry.

Git:  labctl pack add https://github.com/snowops/kafka-drills@v1
OCI:  labctl pack add oci://ghcr.io/snowops/kafka-drills:1.4.2

OCI packs are content-addressed; oras verifies layer digests against the
manifest on pull. Add --require-signature to additionally verify a cosign
signature before any content is extracted (premium/verified packs).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		var p *scenario.Pack
		var err error
		if pack.IsOCIRef(src) {
			p, err = addOCIPack(cmd.Context(), src)
		} else {
			p, err = scenes.InstallPack(src, packAddName, packAddForce, nil)
		}
		if err != nil {
			return err
		}
		if p.Manifest == nil {
			fmt.Printf("Installed pack %q (no pack.yaml manifest — legacy/community pack).\n", p.Name)
		} else {
			m := p.Manifest.Metadata
			fmt.Printf("Installed pack %q — %s v%s (tier: %s, license: %s).\n",
				p.Name, displayOr(m.DisplayName, m.Name), m.Version, p.Manifest.Tier(), m.License)
		}
		fmt.Printf("Scenarios: %s\n", strings.Join(p.Scenarios, ", "))
		fmt.Printf("\nActivate with: labctl scenario up <name>\nRemove with:   labctl pack remove %s\n", p.Name)
		return nil
	},
}

// addOCIPack pulls an OCI pack into a temp dir (verifying signature first when
// required), then installs it through the same validate-then-rename path as git.
func addOCIPack(ctx context.Context, ref string) (*scenario.Pack, error) {
	name := packAddName
	if name == "" {
		name = pack.PackNameFromOCIRef(ref)
	}
	tmp, err := os.MkdirTemp("", "labctl-pack-oci-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	if _, err := pack.Pull(ctx, ref, tmp, pack.PullOptions{
		RequireSignature: packAddRequireSig,
		CosignKey:        packAddCosignKey,
		CertIdentity:     packAddCertID,
		CertOIDCIssuer:   packAddCertIssuer,
	}); err != nil {
		return nil, err
	}
	return scenes.InstallPackFromDir(tmp, name, packAddForce)
}

var packPublishCmd = &cobra.Command{
	Use:   "publish <dir> oci://<registry>/<repo>[:tag]",
	Short: "Publish a scenario pack to an OCI registry",
	Long: `Package a pack directory into a single content-addressed OCI artifact and
push it to a registry with oras. The directory must contain a valid pack.yaml.

  labctl pack publish ./packs/examples/hello-pack oci://ghcr.io/snowops/hello:0.1.0

Add --sign to sign the pushed artifact with cosign (keyless by default, or
--cosign-key for a key pair).`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, ref := args[0], args[1]
		if !pack.IsOCIRef(ref) {
			return fmt.Errorf("destination %q must be an oci:// reference", ref)
		}
		digest, err := pack.Publish(cmd.Context(), dir, ref, pack.PublishOptions{
			Sign:      packPublishSign,
			CosignKey: packPublishCosignKey,
		})
		if err != nil {
			return err
		}
		fmt.Printf("Published %s\n  layer digest: %s\n", ref, digest)
		if packPublishSign {
			fmt.Println("  signed: yes (cosign)")
		}
		return nil
	},
}

var packListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed scenario packs",
	RunE: func(cmd *cobra.Command, args []string) error {
		packs := scenes.Packs()
		if len(packs) == 0 {
			fmt.Println("No packs installed. Install one with: labctl pack add <git-url>")
			return nil
		}
		fmt.Printf("%-24s %-10s %-12s %s\n", "PACK", "VERSION", "TIER", "SCENARIOS")
		for _, p := range packs {
			version, tier := "-", "-"
			if p.Manifest != nil {
				version = p.Manifest.Metadata.Version
				tier = p.Manifest.Tier()
			}
			fmt.Printf("%-24s %-10s %-12s %s\n", p.Name, version, tier, strings.Join(p.Scenarios, ", "))
		}
		return nil
	},
}

var packInfoCmd = &cobra.Command{
	Use:   "info <pack-name>",
	Short: "Show details about an installed pack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, p := range scenes.Packs() {
			if p.Name != args[0] {
				continue
			}
			fmt.Printf("Pack:      %s\n", p.Name)
			if p.Manifest != nil {
				m := p.Manifest.Metadata
				fmt.Printf("Name:      %s\n", m.Name)
				fmt.Printf("Version:   %s\n", m.Version)
				if m.Publisher != "" {
					fmt.Printf("Publisher: %s\n", m.Publisher)
				}
				fmt.Printf("Tier:      %s\n", p.Manifest.Tier())
				if m.License != "" {
					fmt.Printf("License:   %s\n", m.License)
				}
				if m.Description != "" {
					fmt.Printf("About:     %s\n", m.Description)
				}
				if req := p.Manifest.Spec.Requires; len(req.Platform) > 0 || len(req.Packs) > 0 {
					fmt.Printf("Requires:  platform=%s\n", strings.Join(req.Platform, ", "))
				}
			} else {
				fmt.Println("Manifest:  (none — legacy git pack without pack.yaml)")
			}
			fmt.Printf("Scenarios: %s\n", strings.Join(p.Scenarios, ", "))
			return nil
		}
		return fmt.Errorf("pack %q is not installed (see 'labctl pack list')", args[0])
	},
}

var packRemoveCmd = &cobra.Command{
	Use:   "remove <pack-name>",
	Short: "Remove an installed scenario pack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scenes.UninstallPack(args[0]); err != nil {
			return err
		}
		fmt.Printf("Removed pack %q.\n", args[0])
		return nil
	},
}

func displayOr(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func init() {
	packAddCmd.Flags().StringVar(&packAddName, "name", "", "override the pack name (default: derived from the URL/ref)")
	packAddCmd.Flags().BoolVar(&packAddForce, "force", false, "replace the pack if it is already installed")
	packAddCmd.Flags().BoolVar(&packAddRequireSig, "require-signature", false, "verify a cosign signature before extracting (OCI only)")
	packAddCmd.Flags().StringVar(&packAddCosignKey, "cosign-key", "", "cosign public key for signature verification (default: keyless)")
	packAddCmd.Flags().StringVar(&packAddCertID, "certificate-identity", "", "expected keyless signer identity (with --require-signature)")
	packAddCmd.Flags().StringVar(&packAddCertIssuer, "certificate-oidc-issuer", "", "expected keyless OIDC issuer (with --require-signature)")

	packPublishCmd.Flags().BoolVar(&packPublishSign, "sign", false, "sign the pushed artifact with cosign")
	packPublishCmd.Flags().StringVar(&packPublishCosignKey, "cosign-key", "", "cosign private key (default: keyless OIDC signing)")

	packCmd.AddCommand(packAddCmd, packListCmd, packInfoCmd, packRemoveCmd, packPublishCmd)
	rootCmd.AddCommand(packCmd)
}
