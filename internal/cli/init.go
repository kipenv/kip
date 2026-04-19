package cli

import (
	"fmt"
	"os"

	"github.com/kipenv/kip/internal/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var useExclude bool

	cmd := &cobra.Command{
		Use:   "init <team-name>",
		Short: "Link current directory to a team (creates .kip)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]

			teams, err := config.LoadTeams()
			if err != nil {
				return fmt.Errorf("load teams config: %w", err)
			}

			entry := teams.FindByName(teamName)
			if entry == nil {
				return fmt.Errorf("team %q not found. Join or create it first", teamName)
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			projCfg := config.ProjectConfig{
				Team:   teamName,
				Server: cfg.ServerURL,
			}

			path, err := config.SaveProject(dir, projCfg)
			if err != nil {
				return fmt.Errorf("save .kip: %w", err)
			}

			if useExclude {
				if err := config.AddToGitExclude(dir); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Warning: could not add to .git/info/exclude: %v\n", err)
				}
			} else {
				if err := config.AddToGitignore(dir); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Warning: could not add to .gitignore: %v\n", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized kip for team %q\n", teamName)
			fmt.Fprintf(cmd.OutOrStdout(), "Config: %s\n", path)

			return nil
		},
	}

	cmd.Flags().BoolVar(&useExclude, "git-exclude", false, "add .kip to .git/info/exclude instead of .gitignore")

	return cmd
}
