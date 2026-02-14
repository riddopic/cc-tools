package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/session"
)

const (
	defaultSessionLimit = 10
	sessionAliasSetArgs = 2
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage Claude Code sessions",
	}
	cmd.AddCommand(
		newSessionListCmd(),
		newSessionInfoCmd(),
		newSessionAliasCmd(),
		newSessionSearchCmd(),
	)
	return cmd
}

func newSessionListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List recent sessions",
		Example: "  cc-tools session list --limit 20",
		RunE: func(_ *cobra.Command, _ []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))

			sessions, listErr := store.List(limit)
			if listErr != nil {
				return fmt.Errorf("list sessions: %w", listErr)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(os.Stdout, "No sessions found.")
				return nil
			}

			fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "DATE", "ID", "TITLE")
			fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "----", "--", "-----")
			for _, s := range sessions {
				fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", s.Date, s.ID, s.Title)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", defaultSessionLimit, "maximum number of sessions to show")
	return cmd
}

func newSessionInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "info <id-or-alias>",
		Short:   "Show session details",
		Args:    cobra.ExactArgs(1),
		Example: "  cc-tools session info abc123",
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))
			aliases := session.NewAliasManager(filepath.Join(homeDir, ".claude", "session-aliases.json"))

			idOrAlias := args[0]
			if resolved, resolveErr := aliases.Resolve(idOrAlias); resolveErr == nil {
				idOrAlias = resolved
			}

			sess, loadErr := store.Load(idOrAlias)
			if loadErr != nil {
				if errors.Is(loadErr, session.ErrNotFound) {
					return fmt.Errorf("session not found: %s", idOrAlias)
				}
				return fmt.Errorf("load session: %w", loadErr)
			}

			data, marshalErr := json.MarshalIndent(sess, "", "  ")
			if marshalErr != nil {
				return fmt.Errorf("marshal session: %w", marshalErr)
			}
			fmt.Fprintln(os.Stdout, string(data))
			return nil
		},
	}
}

func newSessionAliasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage session aliases",
	}
	cmd.AddCommand(
		newSessionAliasSetCmd(),
		newSessionAliasRemoveCmd(),
		newSessionAliasListCmd(),
	)
	return cmd
}

func newSessionAliasSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set <name> <session-id>",
		Short:   "Create a session alias",
		Args:    cobra.ExactArgs(sessionAliasSetArgs),
		Example: "  cc-tools session alias set mywork abc123",
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			aliases := session.NewAliasManager(filepath.Join(homeDir, ".claude", "session-aliases.json"))
			if setErr := aliases.Set(args[0], args[1]); setErr != nil {
				return fmt.Errorf("set alias: %w", setErr)
			}
			fmt.Fprintf(os.Stdout, "Alias %q set to session %s\n", args[0], args[1])
			return nil
		},
	}
}

func newSessionAliasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Short:   "Remove a session alias",
		Args:    cobra.ExactArgs(1),
		Example: "  cc-tools session alias remove mywork",
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			aliases := session.NewAliasManager(filepath.Join(homeDir, ".claude", "session-aliases.json"))
			if rmErr := aliases.Remove(args[0]); rmErr != nil {
				return fmt.Errorf("remove alias: %w", rmErr)
			}
			fmt.Fprintf(os.Stdout, "Alias %q removed\n", args[0])
			return nil
		},
	}
}

func newSessionAliasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all session aliases",
		RunE: func(_ *cobra.Command, _ []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			aliases := session.NewAliasManager(filepath.Join(homeDir, ".claude", "session-aliases.json"))
			aliasList, listErr := aliases.List()
			if listErr != nil {
				return fmt.Errorf("list aliases: %w", listErr)
			}
			if len(aliasList) == 0 {
				fmt.Fprintln(os.Stdout, "No aliases defined.")
				return nil
			}
			fmt.Fprintf(os.Stdout, "%-20s  %s\n", "ALIAS", "SESSION ID")
			fmt.Fprintf(os.Stdout, "%-20s  %s\n", "-----", "----------")
			for name, id := range aliasList {
				fmt.Fprintf(os.Stdout, "%-20s  %s\n", name, id)
			}
			return nil
		},
	}
}

func newSessionSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "search <query>",
		Short:   "Search sessions",
		Args:    cobra.MinimumNArgs(1),
		Example: "  cc-tools session search refactor",
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			store := session.NewStore(filepath.Join(homeDir, ".claude", "sessions"))
			query := strings.Join(args, " ")
			sessions, searchErr := store.Search(query)
			if searchErr != nil {
				return fmt.Errorf("search sessions: %w", searchErr)
			}
			if len(sessions) == 0 {
				fmt.Fprintln(os.Stdout, "No matching sessions found.")
				return nil
			}
			fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "DATE", "ID", "TITLE")
			fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "----", "--", "-----")
			for _, s := range sessions {
				fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", s.Date, s.ID, s.Title)
			}
			return nil
		},
	}
}
