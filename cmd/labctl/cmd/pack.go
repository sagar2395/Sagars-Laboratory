// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	Long: `Manage scenario packs — versioned bundles of scenarios distributed from a git
source or an OCI registry, and discoverable through the community registry index.

A pack may carry a pack.yaml manifest (apiVersion packs.flightdeck.dev/v1) with
its name, version, publisher, license, tier, and engine compatibility. Packs run
scripts and apply manifests on your cluster — only install sources you trust.`,
}

var packAddCmd = &cobra.Command{
	Use:   "add <name> | <git-url>[@ref] | oci://<registry>/<repo>[:tag]",
	Short: "Install a scenario pack by registry name, git source, or OCI registry",
	Long: `Install a scenario pack. The argument is resolved as, in order:

  - an OCI reference:   labctl pack add oci://ghcr.io/snowops/kafka-drills:1.4.2
  - a git source:       labctl pack add https://github.com/snowops/kafka-drills@v1
  - a registry name:    labctl pack add kafka-drills   (resolved via the index)

A bare name (no scheme) is looked up in the registry index (see 'labctl pack
search') and resolved to its latest version's ref. OCI packs are content-
addressed; add --require-signature to verify a cosign signature before any
content is extracted (premium/verified packs).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resolved, entry, err := resolveAddSource(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		name := packAddName
		if name == "" && entry != nil {
			name = bareName(entry.Name)
		}

		var p *scenario.Pack
		if pack.IsOCIRef(resolved) {
			p, err = addOCIPack(cmd.Context(), resolved, name)
		} else {
			p, err = scenes.InstallPack(resolved, name, packAddForce, nil)
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

// resolveAddSource turns the add argument into an installable source. OCI and
// git references (anything with a scheme or scp-like git@ prefix) pass straight
// through. A bare name is resolved against the registry index to its newest
// version's ref, also returning the matched entry so the caller can default the
// pack name.
func resolveAddSource(ctx context.Context, arg string) (string, *pack.IndexEntry, error) {
	if pack.IsOCIRef(arg) || strings.Contains(arg, "://") || strings.HasPrefix(arg, "git@") {
		return arg, nil, nil
	}
	idx, _, err := newIndexClient().Load(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("resolving %q via the registry index: %w", arg, err)
	}
	entry, ok := idx.Find(arg)
	if !ok {
		return "", nil, fmt.Errorf("pack %q not found in the registry index (try 'labctl pack search %s' or pass a git/oci ref)", arg, arg)
	}
	return entry.ResolveRef(), entry, nil
}

// bareName drops an optional "publisher/" prefix so an index name maps to a
// valid catalog directory name.
func bareName(name string) string {
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		return name[i+1:]
	}
	return name
}

// newIndexClient builds a registry index client from the resolved config and the
// project's local cache dir.
func newIndexClient() *pack.IndexClient {
	return &pack.IndexClient{
		URL:       cfg.RegistryIndexURL,
		CachePath: filepath.Join(cfg.ProjectRoot, ".labctl", "cache", "registry-index.json"),
		VerifyKey: cfg.RegistryIndexKey,
	}
}

// addOCIPack pulls an OCI pack into a temp dir (verifying signature first when
// required), then installs it through the same validate-then-rename path as git.
func addOCIPack(ctx context.Context, ref, name string) (*scenario.Pack, error) {
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

var packSearchCmd = &cobra.Command{
	Use:   "search [term]",
	Short: "Search the registry index for packs",
	Long: `Search the community registry index — a static, optionally signed JSON catalog
fetched from PACK_REGISTRY_INDEX and cached locally. With no term, lists every
pack. Install a result with 'labctl pack add <name>'.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		term := ""
		if len(args) == 1 {
			term = args[0]
		}
		idx, stale, err := newIndexClient().Load(cmd.Context())
		if err != nil {
			return err
		}
		if stale {
			fmt.Fprintln(os.Stderr, "warning: using a cached registry index (refresh failed)")
		}
		results := idx.Search(term)
		if len(results) == 0 {
			fmt.Printf("No packs match %q in the registry index.\n", term)
			return nil
		}
		fmt.Printf("%-26s %-9s %-11s %-9s %s\n", "NAME", "VERSION", "TIER", "VERIFIED", "DESCRIPTION")
		for _, e := range results {
			verified := "-"
			if e.Verified {
				verified = "yes"
			}
			tier := e.Tier
			if tier == "" {
				tier = "community"
			}
			fmt.Printf("%-26s %-9s %-11s %-9s %s\n", e.Name, e.Version, tier, verified, e.Description)
		}
		fmt.Printf("\nInstall with: labctl pack add <name>\n")
		return nil
	},
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

var packValidateIndexCmd = &cobra.Command{
	Use:   "validate-index <index.json>",
	Short: "Validate a registry index file (schema + semantics)",
	Long: `Parse and validate a registry index file the way the CLI does before trusting
it. Intended as the PR gate in the registry repo's validate-index CI.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		idx, err := pack.ParseIndex(data)
		if err != nil {
			return err
		}
		if err := idx.Validate(); err != nil {
			return err
		}
		fmt.Printf("%s is valid: %d entr%s.\n", args[0], len(idx.Entries), plural(len(idx.Entries), "y", "ies"))
		return nil
	},
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
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
	Short: "Show details about an installed pack or a registry entry",
	Long: `Show pack details. An installed pack is shown from its local manifest; otherwise
the name is looked up in the registry index.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, p := range scenes.Packs() {
			if p.Name != args[0] && bareName(p.Name) != args[0] {
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

		// Not installed — try the registry index.
		idx, _, err := newIndexClient().Load(cmd.Context())
		if err != nil {
			return fmt.Errorf("pack %q is not installed and the registry index is unavailable: %w", args[0], err)
		}
		e, ok := idx.Find(args[0])
		if !ok {
			return fmt.Errorf("pack %q is not installed and not in the registry index (see 'labctl pack list' / 'labctl pack search')", args[0])
		}
		fmt.Printf("Pack:      %s  (registry — not installed)\n", e.Name)
		fmt.Printf("Version:   %s\n", e.Version)
		if e.Publisher != "" {
			fmt.Printf("Publisher: %s\n", e.Publisher)
		}
		tier := e.Tier
		if tier == "" {
			tier = "community"
		}
		fmt.Printf("Tier:      %s", tier)
		if e.Verified {
			fmt.Print("  (verified publisher)")
		}
		fmt.Println()
		if e.License != "" {
			fmt.Printf("License:   %s\n", e.License)
		}
		if e.Description != "" {
			fmt.Printf("About:     %s\n", e.Description)
		}
		fmt.Printf("Ref:       %s\n", e.Ref)
		fmt.Printf("\nInstall with: labctl pack add %s\n", bareName(e.Name))
		return nil
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

	packCmd.AddCommand(packAddCmd, packSearchCmd, packListCmd, packInfoCmd, packRemoveCmd, packPublishCmd, packValidateIndexCmd)
	rootCmd.AddCommand(packCmd)
}
