package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
			return listSessions(os.Stdout, store, limit)
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
			return showSessionInfo(os.Stdout, store, aliases, args[0])
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
			return setSessionAlias(os.Stdout, aliases, args[0], args[1])
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
			return removeSessionAlias(os.Stdout, aliases, args[0])
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
			return listSessionAliases(os.Stdout, aliases)
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
			return searchSessions(os.Stdout, store, strings.Join(args, " "))
		},
	}
}

// listSessions writes a formatted table of recent sessions to w.
func listSessions(w io.Writer, store *session.Store, limit int) error {
	sessions, err := store.List(limit)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(w, "No sessions found.")
		return nil
	}

	fmt.Fprintf(w, "%-12s  %-36s  %s\n", "DATE", "ID", "TITLE")
	fmt.Fprintf(w, "%-12s  %-36s  %s\n", "----", "--", "-----")
	for _, s := range sessions {
		fmt.Fprintf(w, "%-12s  %-36s  %s\n", s.Date, s.ID, s.Title)
	}
	return nil
}

// showSessionInfo resolves an ID or alias and writes session details as JSON to w.
func showSessionInfo(w io.Writer, store *session.Store, aliases *session.AliasManager, idOrAlias string) error {
	if resolved, resolveErr := aliases.Resolve(idOrAlias); resolveErr == nil {
		idOrAlias = resolved
	}

	sess, err := store.Load(idOrAlias)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			return fmt.Errorf("session not found: %s", idOrAlias)
		}
		return fmt.Errorf("load session: %w", err)
	}

	data, marshalErr := json.MarshalIndent(sess, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("marshal session: %w", marshalErr)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

// setSessionAlias creates or overwrites a named alias for a session ID.
func setSessionAlias(w io.Writer, aliases *session.AliasManager, name, sessionID string) error {
	if err := aliases.Set(name, sessionID); err != nil {
		return fmt.Errorf("set alias: %w", err)
	}
	fmt.Fprintf(w, "Alias %q set to session %s\n", name, sessionID)
	return nil
}

// removeSessionAlias deletes a named alias.
func removeSessionAlias(w io.Writer, aliases *session.AliasManager, name string) error {
	if err := aliases.Remove(name); err != nil {
		return fmt.Errorf("remove alias: %w", err)
	}
	fmt.Fprintf(w, "Alias %q removed\n", name)
	return nil
}

// listSessionAliases writes all aliases as a formatted table to w.
func listSessionAliases(w io.Writer, aliases *session.AliasManager) error {
	aliasList, err := aliases.List()
	if err != nil {
		return fmt.Errorf("list aliases: %w", err)
	}
	if len(aliasList) == 0 {
		fmt.Fprintln(w, "No aliases defined.")
		return nil
	}
	fmt.Fprintf(w, "%-20s  %s\n", "ALIAS", "SESSION ID")
	fmt.Fprintf(w, "%-20s  %s\n", "-----", "----------")
	for name, id := range aliasList {
		fmt.Fprintf(w, "%-20s  %s\n", name, id)
	}
	return nil
}

// searchSessions searches sessions by query and writes matches as a formatted table to w.
func searchSessions(w io.Writer, store *session.Store, query string) error {
	sessions, err := store.Search(query)
	if err != nil {
		return fmt.Errorf("search sessions: %w", err)
	}
	if len(sessions) == 0 {
		fmt.Fprintln(w, "No matching sessions found.")
		return nil
	}
	fmt.Fprintf(w, "%-12s  %-36s  %s\n", "DATE", "ID", "TITLE")
	fmt.Fprintf(w, "%-12s  %-36s  %s\n", "----", "--", "-----")
	for _, s := range sessions {
		fmt.Fprintf(w, "%-12s  %-36s  %s\n", s.Date, s.ID, s.Title)
	}
	return nil
}
