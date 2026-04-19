package cli

import (
	"fmt"

	"github.com/kipenv/kip/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage kip configuration",
	}

	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigGetCmd())

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if serverURL != "" {
				cfg.ServerURL = serverURL
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Configuration saved.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "server URL")

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "server_url: %s\n", cfg.ServerURL)
			return nil
		},
	}
}
