// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"go.flightdeck.dev/labctl/pkg/credential"
)

// credentialCmd is the parent for `labctl credential` subcommands.
// Credentials are verifiable achievement records issued when a user completes
// a certification track (task 078). The OSS engine provides issuing and
// verification; certification content lives in premium packs (private repo).
func credentialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credential",
		Short: "Manage Flightdeck achievement credentials",
		Long: `Issue, list, and verify Flightdeck certification credentials.

Credentials are self-contained verifiable records (HMAC-SHA256 signed) awarded
when a learner completes a certification track. The OSS engine handles the
grading and record format; certification content and exam delivery live in
premium packs (private repo).

Set LABCTL_SIGNING_KEY (hex-encoded 32 bytes) to sign issued credentials.
Without it, credentials are issued unsigned and still verifiable within the
same lab instance.`,
	}
	cmd.AddCommand(credentialListCmd())
	cmd.AddCommand(credentialIssueCmd())
	cmd.AddCommand(credentialVerifyCmd())
	return cmd
}

func credentialStorePath() string {
	return filepath.Join(cfg.ProjectRoot, ".labctl", "credentials")
}

func credentialListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List issued credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store := credential.NewStore(credentialStorePath())
			certs, err := store.All()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(certs) == 0 {
				fmt.Fprintln(out, "No credentials issued yet.")
				fmt.Fprintln(out, "Earn one by completing a certification track: labctl learn start <track>")
				return nil
			}
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tSubject\tAchievement\tGrade\tScore\tIssued")
			for _, c := range certs {
				id := c.ID
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(tw, "%s…\t%s\t%s\t%s\t%d\t%s\n",
					id, c.Subject, c.Achievement, c.Grade, c.Score,
					c.IssuedAt.Format("2006-01-02"))
			}
			tw.Flush()
			return nil
		},
	}
}

var (
	credIssueSubject     string
	credIssueScore       int
	credIssueAchievement string
	credIssueIssuer      string
	credIssueOutput      string
)

func credentialIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a credential for a completed achievement",
		Long: `Issue a verifiable credential for a completed certification achievement.
The credential is stored in .labctl/credentials/ and printed as JSON.

Sign issued credentials with LABCTL_SIGNING_KEY (hex-encoded 32 bytes).
Without a key the credential is still valid — just unsigned.

Example:
  labctl credential issue --achievement kubernetes-foundations --subject alice --score 88`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if credIssueAchievement == "" {
				return fmt.Errorf("--achievement is required")
			}
			if credIssueSubject == "" {
				user := os.Getenv("USER")
				if user == "" {
					user = "unknown"
				}
				credIssueSubject = user
			}
			if credIssueScore < 0 || credIssueScore > 100 {
				return fmt.Errorf("--score must be 0-100, got %d", credIssueScore)
			}

			grade := credential.GradeForScore(credIssueScore)
			issuer := credIssueIssuer
			if issuer == "" {
				issuer = "Flightdeck"
			}

			cert, err := credential.Issue(credential.IssueRequest{
				Subject:     credIssueSubject,
				Achievement: credIssueAchievement,
				Grade:       grade,
				Score:       credIssueScore,
				Issuer:      issuer,
				SigningKey:  signingKeyFromEnv(),
			})
			if err != nil {
				return fmt.Errorf("issuing credential: %w", err)
			}

			// Persist
			store := credential.NewStore(credentialStorePath())
			if err := store.Save(cert); err != nil {
				return fmt.Errorf("saving credential: %w", err)
			}

			out := cmd.OutOrStdout()
			if credIssueOutput == "json" {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(cert)
			}

			fmt.Fprintf(out, "Credential issued!\n\n")
			fmt.Fprintf(out, "  ID:          %s\n", cert.ID)
			fmt.Fprintf(out, "  Subject:     %s\n", cert.Subject)
			fmt.Fprintf(out, "  Achievement: %s\n", cert.Achievement)
			fmt.Fprintf(out, "  Grade:       %s\n", cert.Grade)
			fmt.Fprintf(out, "  Score:       %d / 100\n", cert.Score)
			fmt.Fprintf(out, "  Issued at:   %s\n", cert.IssuedAt.Format(time.RFC3339))
			fmt.Fprintf(out, "  Issuer:      %s\n", cert.Issuer)
			if cert.Signature != "" {
				fmt.Fprintf(out, "  Signed:      yes (HMAC-SHA256)\n")
			} else {
				fmt.Fprintf(out, "  Signed:      no  (set LABCTL_SIGNING_KEY to sign)\n")
			}
			fmt.Fprintf(out, "\nStore: %s/%s.json\n", credentialStorePath(), cert.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&credIssueSubject, "subject", "", "who earned the credential (default: $USER)")
	cmd.Flags().StringVar(&credIssueAchievement, "achievement", "", "achievement / certification track name (required)")
	cmd.Flags().IntVar(&credIssueScore, "score", 75, "numeric score 0-100")
	cmd.Flags().StringVar(&credIssueIssuer, "issuer", "", "issuer name (default: Flightdeck)")
	cmd.Flags().StringVar(&credIssueOutput, "output", "text", "output format: text | json")
	return cmd
}

func credentialVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <credential-file>",
		Short: "Verify a credential's integrity",
		Long: `Verify the HMAC-SHA256 signature of a credential file. Requires the same
LABCTL_SIGNING_KEY that was used at issuance.

Without a signing key, unsigned credentials from the same issuer are accepted
(no signature to verify). Provide the key to confirm the credential has not been
tampered with.

Example:
  labctl credential verify .labctl/credentials/<id>.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("reading credential: %w", err)
			}
			var cert credential.Certificate
			if err := json.Unmarshal(data, &cert); err != nil {
				return fmt.Errorf("parsing credential: %w", err)
			}
			key := signingKeyFromEnv()
			if err := cert.Verify(key); err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Credential verified.")
			fmt.Fprintf(out, "  Subject:     %s\n", cert.Subject)
			fmt.Fprintf(out, "  Achievement: %s\n", cert.Achievement)
			fmt.Fprintf(out, "  Grade:       %s\n", cert.Grade)
			fmt.Fprintf(out, "  Score:       %d\n", cert.Score)
			fmt.Fprintf(out, "  Issued:      %s\n", cert.IssuedAt.Format(time.RFC3339))
			fmt.Fprintf(out, "  Issuer:      %s\n", cert.Issuer)
			if cert.Signature != "" {
				fmt.Fprintln(out, "  Signature:   valid (HMAC-SHA256)")
			} else {
				fmt.Fprintln(out, "  Signature:   none (unsigned)")
			}
			return nil
		},
	}
}

// signingKeyFromEnv decodes LABCTL_SIGNING_KEY (hex string) to bytes.
// Returns nil when the env var is unset, which means unsigned mode.
func signingKeyFromEnv() []byte {
	hex := os.Getenv("LABCTL_SIGNING_KEY")
	if hex == "" {
		return nil
	}
	key, err := hexDecode(hex)
	if err != nil {
		return nil
	}
	return key
}

// hexDecode converts a hex string to bytes.
func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("odd-length hex string")
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		v, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex at position %d: %w", i, err)
		}
		b[i/2] = byte(v)
	}
	return b, nil
}

func init() {
	rootCmd.AddCommand(credentialCmd())
}
