package cli

import (
	"fmt"
	"os"

	"github.com/kipenv/kip/internal/envfile"
	"github.com/kipenv/kip/internal/scanner"
	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan [file]",
		Short: "Scan a .env file for leaked credentials and weak secrets",
		Long:  "Checks each entry against known credential patterns (AWS keys, Stripe keys, private keys, etc.) and flags weak or short secrets.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := ".env"
			if len(args) > 0 {
				inputFile = args[0]
			}

			entries, err := envfile.Parse(inputFile)
			if err != nil {
				return fmt.Errorf("parse %s: %w", inputFile, err)
			}

			if len(entries) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No entries found in %s\n", inputFile)
				return nil
			}

			findings := scanner.Scan(entries)

			warnCount := printFindings(cmd, findings)

			if warnCount > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s %d warning(s) found in %s\n",
					colorRed("✗"), warnCount, inputFile)
				os.Exit(1)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n%s No issues found in %s\n",
				colorGreen("✓"), inputFile)
			return nil
		},
	}

	return cmd
}

func printFindings(cmd *cobra.Command, findings []scanner.Finding) int {
	warnCount := 0

	for _, f := range findings {
		switch f.Severity {
		case scanner.SeverityWarn:
			warnCount++
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s — %s\n",
				colorYellow("WARN"), f.Key, f.Message)
		case scanner.SeverityOK:
			fmt.Fprintf(cmd.OutOrStdout(), "  %s    %s — %s\n",
				colorGreen("OK"), f.Key, f.Message)
		}
	}

	return warnCount
}

// ANSI color helpers.

func colorRed(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func colorYellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func colorGreen(s string) string {
	return "\033[32m" + s + "\033[0m"
}
