// Package cli defines the CLI commands for kip.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for kip. version and commit are stamped
// in at build time by the release pipeline; they fall back to placeholders in
// a plain `go build`.
func NewRootCmd(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kip",
		Short:   "Secure .env file sharing with self-destructing links",
		Long:    "kip — the missing CLI between pasting in Slack and deploying Vault.",
		Version: version,
		// A failed upload is not a usage mistake: don't dump the help text over
		// it. Errors are printed once, by main.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Report the commit too when the build stamped one in, so a bug report
	// pins down an exact build rather than just a tag.
	if commit != "" && commit != "none" {
		cmd.SetVersionTemplate(fmt.Sprintf("kip %s (%s)\n", version, commit))
	} else {
		cmd.SetVersionTemplate(fmt.Sprintf("kip %s\n", version))
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
