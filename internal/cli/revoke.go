package cli

import (
	"fmt"
	"strings"

	"github.com/kipenv/kip/internal/client"
	"github.com/kipenv/kip/internal/config"
	"github.com/spf13/cobra"
)

func newRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke <link-or-id>",
		Short: "Revoke a shared secret so it can no longer be read",
		Long:  "Deletes a secret from the server immediately. Accepts a full kip link or just the secret ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := extractID(args[0])

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			c := client.New(cfg.ServerURL)
			if err := c.DeleteSecret(id); err != nil {
				return fmt.Errorf("revoke: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Secret %s has been revoked.\n", id)
			return nil
		},
	}

	return cmd
}

// extractID returns the secret ID from a full kip link or a bare ID string.
func extractID(input string) string {
	// Strip any fragment (key after #)
	input = strings.ReplaceAll(input, `\#`, "#")
	if idx := strings.Index(input, "#"); idx != -1 {
		input = input[:idx]
	}

	// If it contains /s/, extract the ID from the path
	if idx := strings.LastIndex(input, "/s/"); idx != -1 {
		return input[idx+3:]
	}

	return input
}
