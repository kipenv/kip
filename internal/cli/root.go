// Package cli defines the CLI commands for kip.
package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for kip.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kip",
		Short: "Secure .env file sharing with self-destructing links",
		Long:  "kip — the missing CLI between pasting in Slack and deploying Vault.",
		// A failed upload is not a usage mistake: don't dump the help text over
		// it. Errors are printed once, by main.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newPushCmd(),
		newPullCmd(),
		newRevokeCmd(),
		newConfigCmd(),
		newTeamCmd(),
		newInitCmd(),
		newGenerateCmd(),
		newScanCmd(),
	)

	return cmd
}
