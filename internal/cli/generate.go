package cli

import (
	"fmt"
	"os"

	"github.com/kipenv/kip/internal/envfile"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "generate [file]",
		Short: "Generate .env.example from a .env file",
		Long:  "Reads a .env file, strips values, and outputs a .env.example with placeholder types.",
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
				return fmt.Errorf("no entries found in %s", inputFile)
			}

			example := envfile.GenerateExample(entries)

			if output == "" {
				fmt.Fprint(cmd.OutOrStdout(), example)
				return nil
			}

			// 0644 by design: .env.example holds keys with placeholder values —
			// no secrets — and exists to be committed and read by the whole team.
			if err := os.WriteFile(output, []byte(example), 0o644); err != nil { //nolint:gosec // see above
				return fmt.Errorf("write file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated %s (%d keys)\n", output, len(entries))
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")

	return cmd
}
