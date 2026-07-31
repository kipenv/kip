package cli

import (
	"fmt"

	"github.com/kipenv/kip/internal/client"
	"github.com/kipenv/kip/internal/config"
	"github.com/spf13/cobra"
)

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage teams",
	}

	cmd.AddCommand(
		newTeamCreateCmd(),
		newTeamJoinCmd(),
		newTeamLeaveCmd(),
		newTeamListCmd(),
		newTeamMembersCmd(),
	)

	return cmd
}

func newTeamCreateCmd() *cobra.Command {
	var username string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if username == "" {
				return fmt.Errorf("--username is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			c := client.New(cfg.ServerURL)
			resp, err := c.CreateTeam(cmd.Context(), client.CreateTeamRequest{
				Name:     name,
				Username: username,
			})
			if err != nil {
				return fmt.Errorf("create team: %w", err)
			}

			// Save token locally
			teams, err := config.LoadTeams()
			if err != nil {
				return fmt.Errorf("load teams config: %w", err)
			}
			teams.AddTeam(config.TeamEntry{
				ID:         resp.ID,
				Name:       resp.Name,
				Token:      resp.MemberToken,
				InviteCode: resp.InviteCode,
				Username:   username,
			})
			if err := config.SaveTeams(teams); err != nil {
				return fmt.Errorf("save teams config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Team created: %s\n", resp.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Invite code: %s\n", resp.InviteCode)
			fmt.Fprintf(cmd.OutOrStdout(), "\nShare the invite code so others can join.\n")

			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "your username in this team")
	_ = cmd.MarkFlagRequired("username")

	return cmd
}

func newTeamJoinCmd() *cobra.Command {
	var username string

	cmd := &cobra.Command{
		Use:   "join <invite-code>",
		Short: "Join a team with an invite code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inviteCode := args[0]
			if username == "" {
				return fmt.Errorf("--username is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			c := client.New(cfg.ServerURL)
			resp, err := c.JoinTeam(cmd.Context(), client.JoinTeamRequest{
				InviteCode: inviteCode,
				Username:   username,
			})
			if err != nil {
				return fmt.Errorf("join team: %w", err)
			}

			teams, err := config.LoadTeams()
			if err != nil {
				return fmt.Errorf("load teams config: %w", err)
			}
			teams.AddTeam(config.TeamEntry{
				ID:       resp.ID,
				Name:     resp.Name,
				Token:    resp.MemberToken,
				Username: username,
			})
			if err := config.SaveTeams(teams); err != nil {
				return fmt.Errorf("save teams config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Joined team: %s\n", resp.Name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "your username in this team")
	_ = cmd.MarkFlagRequired("username")

	return cmd
}

func newTeamLeaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "leave <team-name>",
		Short: "Leave a team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]

			teams, err := config.LoadTeams()
			if err != nil {
				return fmt.Errorf("load teams config: %w", err)
			}

			entry := teams.FindByName(teamName)
			if entry == nil {
				return fmt.Errorf("team %q not found in your config", teamName)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			c := client.New(cfg.ServerURL)
			if err := c.LeaveTeam(cmd.Context(), entry.ID, entry.Token); err != nil {
				return fmt.Errorf("leave team: %w", err)
			}

			teams.RemoveTeam(entry.ID)
			if err := config.SaveTeams(teams); err != nil {
				return fmt.Errorf("save teams config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Left team: %s\n", teamName)

			return nil
		},
	}
}

func newTeamListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List your teams",
		RunE: func(cmd *cobra.Command, _ []string) error {
			teams, err := config.LoadTeams()
			if err != nil {
				return fmt.Errorf("load teams config: %w", err)
			}

			if len(teams.Teams) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No teams. Create one with: kip team create <name> -u <username>")
				return nil
			}

			for _, t := range teams.Teams {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s (as %s)\n", t.Name, t.Username)
			}

			return nil
		},
	}
}

func newTeamMembersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "members <team-name>",
		Short: "List members of a team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]

			teams, err := config.LoadTeams()
			if err != nil {
				return fmt.Errorf("load teams config: %w", err)
			}

			entry := teams.FindByName(teamName)
			if entry == nil {
				return fmt.Errorf("team %q not found in your config", teamName)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			c := client.New(cfg.ServerURL)
			members, err := c.GetTeamMembers(cmd.Context(), entry.ID, entry.Token)
			if err != nil {
				return fmt.Errorf("get members: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Members of %s:\n", teamName)
			for _, m := range members {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s (joined %s)\n", m.Username, m.JoinedAt)
			}

			return nil
		},
	}
}
