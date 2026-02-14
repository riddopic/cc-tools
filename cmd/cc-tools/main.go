// Package main implements the cc-tools CLI application.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/riddopic/cc-tools/internal/compact"
	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/hookcmd"
	"github.com/riddopic/cc-tools/internal/hooks"
	"github.com/riddopic/cc-tools/internal/notify"
	"github.com/riddopic/cc-tools/internal/observe"
	"github.com/riddopic/cc-tools/internal/output"
	"github.com/riddopic/cc-tools/internal/pkgmanager"
	"github.com/riddopic/cc-tools/internal/session"
	"github.com/riddopic/cc-tools/internal/shared"
	"github.com/riddopic/cc-tools/internal/superpowers"
)

const (
	minArgs             = 2
	helpFlag            = "--help"
	helpCommand         = "help"
	minSessionArgs      = 3
	minSessionAliasArgs = 4
	sessionListLimit    = 10
)

// Build-time variables.
var version = "dev"

func needsStdin(cmd string) bool {
	return cmd == "validate" || cmd == "hook"
}

func main() {
	// Read stdin once for commands that need it.
	var stdinData []byte
	if len(os.Args) > 1 && needsStdin(os.Args[1]) {
		if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinData, _ = io.ReadAll(os.Stdin)
		}
	}

	out := output.NewTerminal(os.Stdout, os.Stderr)

	// Debug logging - log all invocations to a file
	writeDebugLog(os.Args, stdinData)

	if len(os.Args) < minArgs {
		printUsage(out)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		runValidate(stdinData)
	case "hook":
		runHookCommand(stdinData)
	case "session":
		runSessionCommand()
	case "skip":
		runSkipCommand()
	case "unskip":
		runUnskipCommand()
	case "debug":
		runDebugCommand()
	case "mcp":
		runMCPCommand()
	case "config":
		runConfigCommand()
	case "version":
		// Print version to stdout as intended output
		_ = out.Raw(fmt.Sprintf("cc-tools %s\n", version))
	case helpCommand, "-h", helpFlag:
		printUsage(out)
	default:
		_ = out.Error("Unknown command: %s", os.Args[1])
		printUsage(out)
		os.Exit(1)
	}
}

func printUsage(out *output.Terminal) {
	_ = out.RawError(`cc-tools - Claude Code Tools

Usage:
  cc-tools <command> [arguments]

Commands:
  validate      Run smart validation (lint and test in parallel)
  hook          Handle Claude Code hook events
  session       Manage Claude Code sessions
  skip          Configure skip settings for directories
  unskip        Remove skip settings from directories
  debug         Configure debug logging for directories
  mcp           Manage Claude MCP servers
  config        Manage configuration settings
  version       Print version information
  help          Show this help message

Examples:
  echo '{"file_path": "main.go"}' | cc-tools validate
  echo '{"hook_event_name":"SessionStart"}' | cc-tools hook session-start
  cc-tools session list
  cc-tools mcp list
  cc-tools mcp enable jira
`)
}

// hookEventMap maps CLI subcommand names to Claude Code hook event names.
var hookEventMap = map[string]string{ //nolint:gochecknoglobals // static lookup table
	"session-start":         "SessionStart",
	"session-end":           "SessionEnd",
	"pre-tool-use":          "PreToolUse",
	"post-tool-use":         "PostToolUse",
	"post-tool-use-failure": "PostToolUseFailure",
	"pre-compact":           "PreCompact",
	"stop":                  "Stop",
	"notification":          "Notification",
}

const minHookArgs = 3

func runHookCommand(stdinData []byte) {
	out := output.NewTerminal(os.Stdout, os.Stderr)

	if len(os.Args) < minHookArgs {
		_ = out.Error("Usage: cc-tools hook <event-type>")
		os.Exit(1)
	}

	subCmd := os.Args[2]
	eventName, ok := hookEventMap[subCmd]
	if !ok {
		// Accept unknown events gracefully (future-proofing)
		eventName = subCmd
	}

	input, err := hookcmd.ParseInput(bytes.NewReader(stdinData))
	if err != nil {
		_ = out.Error("error parsing hook input: %v", err)
		os.Exit(0) // still exit 0 — hooks must not block
	}
	input.HookEventName = eventName

	registry := buildHookRegistry()
	exitCode := hookcmd.Dispatch(context.Background(), input, os.Stdout, os.Stderr, registry)
	os.Exit(exitCode)
}

// handlerFunc adapts a name and function into a hookcmd.Handler.
type handlerFunc struct {
	name string
	fn   func(ctx context.Context, input *hookcmd.HookInput, out, errOut io.Writer) error
}

func (h *handlerFunc) Name() string { return h.name }

func (h *handlerFunc) Run(ctx context.Context, input *hookcmd.HookInput, out, errOut io.Writer) error {
	return h.fn(ctx, input, out, errOut)
}

func buildHookRegistry() map[string][]hookcmd.Handler {
	cfg := loadHookConfig()

	return map[string][]hookcmd.Handler{
		"SessionStart":       {superpowersHandler(), pkgManagerHandler(), sessionContextHandler()},
		"PreToolUse":         {suggestCompactHandler(cfg), observeHandler(cfg, "pre")},
		"PostToolUse":        {observeHandler(cfg, "post")},
		"PostToolUseFailure": {observeHandler(cfg, "failure")},
		"PreCompact":         {logCompactionHandler()},
		"Notification":       {notifyAudioHandler(cfg), notifyDesktopHandler(cfg)},
	}
}

func loadHookConfig() *config.Values {
	mgr := config.NewManager()
	cfg, err := mgr.GetConfig(context.Background())
	if err != nil {
		return nil
	}
	return cfg
}

func superpowersHandler() *handlerFunc {
	return &handlerFunc{
		name: "superpowers",
		fn: func(ctx context.Context, input *hookcmd.HookInput, out, _ io.Writer) error {
			return superpowers.NewInjector(input.Cwd).Run(ctx, out)
		},
	}
}

func pkgManagerHandler() *handlerFunc {
	return &handlerFunc{
		name: "pkg-manager",
		fn: func(_ context.Context, input *hookcmd.HookInput, _, _ io.Writer) error {
			manager := pkgmanager.Detect(input.Cwd)
			envDir := filepath.Join(input.Cwd, ".claude")
			if mkErr := os.MkdirAll(envDir, 0o750); mkErr != nil {
				return fmt.Errorf("create .claude directory: %w", mkErr)
			}
			envFile := filepath.Join(envDir, ".env")
			return pkgmanager.WriteToEnvFile(envFile, manager)
		},
	}
}

func suggestCompactHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "suggest-compact",
		fn: func(_ context.Context, input *hookcmd.HookInput, _, errOut io.Writer) error {
			if cfg == nil {
				return nil
			}
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return fmt.Errorf("get home directory: %w", homeErr)
			}
			stateDir := filepath.Join(homeDir, ".cache", "cc-tools", "compact")
			s := compact.NewSuggestor(stateDir, cfg.Compact.Threshold, cfg.Compact.ReminderInterval)
			s.RecordCall(input.SessionID, errOut)
			return nil
		},
	}
}

func observeHandler(cfg *config.Values, phase string) *handlerFunc {
	return &handlerFunc{
		name: "observe-" + phase,
		fn: func(_ context.Context, input *hookcmd.HookInput, _, _ io.Writer) error {
			if cfg == nil || !cfg.Observe.Enabled {
				return nil
			}
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return fmt.Errorf("get home directory: %w", homeErr)
			}
			dir := filepath.Join(homeDir, ".cache", "cc-tools", "observations")
			obs := observe.NewObserver(dir, cfg.Observe.MaxFileSizeMB)
			return obs.Record(observe.Event{
				Timestamp: time.Now(),
				Phase:     phase,
				ToolName:  input.ToolName,
				ToolInput: input.ToolInput,
				SessionID: input.SessionID,
			})
		},
	}
}

func logCompactionHandler() *handlerFunc {
	return &handlerFunc{
		name: "log-compaction",
		fn: func(_ context.Context, _ *hookcmd.HookInput, _, _ io.Writer) error {
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return fmt.Errorf("get home directory: %w", homeErr)
			}
			logDir := filepath.Join(homeDir, ".cache", "cc-tools")
			return compact.LogCompaction(logDir)
		},
	}
}

func sessionContextHandler() *handlerFunc {
	return &handlerFunc{
		name: "session-context",
		fn: func(_ context.Context, _ *hookcmd.HookInput, out, errOut io.Writer) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}

			storeDir := filepath.Join(homeDir, ".claude", "sessions")
			store := session.NewStore(storeDir)

			sessions, _ := store.List(1)
			if len(sessions) == 0 {
				return nil
			}

			latest := sessions[0]
			if latest.Summary != "" {
				_, _ = fmt.Fprintf(out, "Previous session (%s): %s\n", latest.Date, latest.Summary)
			}

			aliasFile := filepath.Join(homeDir, ".claude", "session-aliases.json")
			aliases := session.NewAliasManager(aliasFile)
			aliasList, aliasErr := aliases.List()
			if aliasErr == nil && len(aliasList) > 0 {
				names := make([]string, 0, len(aliasList))
				for name := range aliasList {
					names = append(names, name)
				}
				_, _ = fmt.Fprintf(errOut, "[session-context] %d alias(es): %s\n",
					len(aliasList), strings.Join(names, ", "))
			}

			return nil
		},
	}
}

func notifyAudioHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "notify-audio",
		fn: func(_ context.Context, _ *hookcmd.HookInput, _, _ io.Writer) error {
			if cfg == nil || !cfg.Notify.Audio.Enabled {
				return nil
			}

			player := &afPlayer{}
			qh := notify.QuietHours{
				Enabled: cfg.Notify.QuietHours.Enabled,
				Start:   cfg.Notify.QuietHours.Start,
				End:     cfg.Notify.QuietHours.End,
			}
			audio := notify.NewAudio(player, cfg.Notify.Audio.Directory, qh, nil)
			return audio.PlayRandom()
		},
	}
}

// afPlayer implements notify.AudioPlayer using macOS afplay.
type afPlayer struct{}

func (a *afPlayer) Play(filepath string) error {
	return exec.CommandContext(context.Background(), "afplay", filepath).Run()
}

func notifyDesktopHandler(cfg *config.Values) *handlerFunc {
	return &handlerFunc{
		name: "notify-desktop",
		fn: func(_ context.Context, input *hookcmd.HookInput, _, _ io.Writer) error {
			if cfg == nil || !cfg.Notify.Desktop.Enabled {
				return nil
			}

			qh := notify.QuietHours{
				Enabled: cfg.Notify.QuietHours.Enabled,
				Start:   cfg.Notify.QuietHours.Start,
				End:     cfg.Notify.QuietHours.End,
			}
			if qh.IsActive(time.Now()) {
				return nil
			}

			runner := &osascriptRunner{}
			desktop := notify.NewDesktop(runner)

			title := "Claude Code"
			message := "Task completed"
			if input.Title != "" {
				title = input.Title
			}
			if input.Message != "" {
				message = input.Message
			}

			return desktop.Send(title, message)
		},
	}
}

// osascriptRunner implements notify.CmdRunner.
type osascriptRunner struct{}

func (o *osascriptRunner) Run(name string, args ...string) error {
	return exec.CommandContext(context.Background(), name, args...).Run()
}

func runSessionCommand() {
	out := output.NewTerminal(os.Stdout, os.Stderr)

	if len(os.Args) < minSessionArgs {
		_ = out.Error("Usage: cc-tools session <list|info|alias|aliases|search> [args]")
		os.Exit(1)
	}

	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		_ = out.Error("get home directory: %v", homeErr)
		os.Exit(1)
	}

	storeDir := filepath.Join(homeDir, ".claude", "sessions")
	aliasFile := filepath.Join(homeDir, ".claude", "session-aliases.json")
	store := session.NewStore(storeDir)
	aliases := session.NewAliasManager(aliasFile)

	switch os.Args[2] {
	case listCommand:
		runSessionList(store, out)
	case "info":
		runSessionInfo(store, aliases, out)
	case "alias":
		runSessionAlias(aliases, out)
	case "aliases":
		runSessionAliasList(aliases, out)
	case "search":
		runSessionSearch(store, out)
	default:
		_ = out.Error("Unknown session subcommand: %s", os.Args[2])
		os.Exit(1)
	}
}

func runSessionList(store *session.Store, out *output.Terminal) {
	limit := sessionListLimit

	for i := minSessionArgs; i < len(os.Args); i++ {
		if os.Args[i] == "--limit" && i+1 < len(os.Args) {
			parsed, parseErr := strconv.Atoi(os.Args[i+1])
			if parseErr != nil {
				_ = out.Error("invalid limit value: %s", os.Args[i+1])
				os.Exit(1)
			}
			limit = parsed
		}
	}

	sessions, err := store.List(limit)
	if err != nil {
		_ = out.Error("list sessions: %v", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(os.Stdout, "No sessions found.")
		return
	}

	fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "DATE", "ID", "TITLE")
	fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "----", "--", "-----")
	for _, s := range sessions {
		fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", s.Date, s.ID, s.Title)
	}
}

func runSessionInfo(store *session.Store, aliases *session.AliasManager, out *output.Terminal) {
	if len(os.Args) < minSessionAliasArgs {
		_ = out.Error("Usage: cc-tools session info <id-or-alias>")
		os.Exit(1)
	}

	idOrAlias := os.Args[3]

	// Try resolving as alias first.
	resolvedID, resolveErr := aliases.Resolve(idOrAlias)
	if resolveErr == nil {
		idOrAlias = resolvedID
	}

	sess, err := store.Load(idOrAlias)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			_ = out.Error("session not found: %s", idOrAlias)
			os.Exit(1)
		}
		_ = out.Error("load session: %v", err)
		os.Exit(1)
	}

	data, marshalErr := json.MarshalIndent(sess, "", "  ")
	if marshalErr != nil {
		_ = out.Error("marshal session: %v", marshalErr)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, string(data))
}

func runSessionAlias(aliases *session.AliasManager, out *output.Terminal) {
	if len(os.Args) < minSessionAliasArgs {
		_ = out.Error("Usage: cc-tools session alias <set|remove|list>")
		os.Exit(1)
	}

	switch os.Args[3] {
	case "set":
		runSessionAliasSet(aliases, out)
	case "remove":
		runSessionAliasRemove(aliases, out)
	case listCommand:
		runSessionAliasList(aliases, out)
	default:
		_ = out.Error("Unknown alias subcommand: %s", os.Args[3])
		os.Exit(1)
	}
}

const minAliasSetArgs = 6

func runSessionAliasSet(aliases *session.AliasManager, out *output.Terminal) {
	if len(os.Args) < minAliasSetArgs {
		_ = out.Error("Usage: cc-tools session alias set <name> <session-id>")
		os.Exit(1)
	}

	name := os.Args[4]
	sessionID := os.Args[5]

	if err := aliases.Set(name, sessionID); err != nil {
		_ = out.Error("set alias: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Alias %q set to session %s\n", name, sessionID)
}

const minAliasRemoveArgs = 5

func runSessionAliasRemove(aliases *session.AliasManager, out *output.Terminal) {
	if len(os.Args) < minAliasRemoveArgs {
		_ = out.Error("Usage: cc-tools session alias remove <name>")
		os.Exit(1)
	}

	name := os.Args[4]

	if err := aliases.Remove(name); err != nil {
		_ = out.Error("remove alias: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Alias %q removed\n", name)
}

func runSessionAliasList(aliases *session.AliasManager, out *output.Terminal) {
	aliasList, err := aliases.List()
	if err != nil {
		_ = out.Error("list aliases: %v", err)
		os.Exit(1)
	}

	if len(aliasList) == 0 {
		fmt.Fprintln(os.Stdout, "No aliases defined.")
		return
	}

	fmt.Fprintf(os.Stdout, "%-20s  %s\n", "ALIAS", "SESSION ID")
	fmt.Fprintf(os.Stdout, "%-20s  %s\n", "-----", "----------")
	for name, id := range aliasList {
		fmt.Fprintf(os.Stdout, "%-20s  %s\n", name, id)
	}
}

func runSessionSearch(store *session.Store, out *output.Terminal) {
	if len(os.Args) < minSessionAliasArgs {
		_ = out.Error("Usage: cc-tools session search <query>")
		os.Exit(1)
	}

	query := strings.Join(os.Args[3:], " ")

	sessions, err := store.Search(query)
	if err != nil {
		_ = out.Error("search sessions: %v", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(os.Stdout, "No matching sessions found.")
		return
	}

	fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "DATE", "ID", "TITLE")
	fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", "----", "--", "-----")
	for _, s := range sessions {
		fmt.Fprintf(os.Stdout, "%-12s  %-36s  %s\n", s.Date, s.ID, s.Title)
	}
}

func loadValidateConfig() (int, int) {
	timeoutSecs := 60
	cooldownSecs := 5

	// Load configuration
	cfg, _ := config.Load()
	if cfg != nil {
		if cfg.Hooks.Validate.TimeoutSeconds > 0 {
			timeoutSecs = cfg.Hooks.Validate.TimeoutSeconds
		}
		if cfg.Hooks.Validate.CooldownSeconds > 0 {
			cooldownSecs = cfg.Hooks.Validate.CooldownSeconds
		}
	}

	// Environment variables override config
	if envTimeout := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_TIMEOUT_SECONDS"); envTimeout != "" {
		if val, err := strconv.Atoi(envTimeout); err == nil && val > 0 {
			timeoutSecs = val
		}
	}
	if envCooldown := os.Getenv("CC_TOOLS_HOOKS_VALIDATE_COOLDOWN_SECONDS"); envCooldown != "" {
		if val, err := strconv.Atoi(envCooldown); err == nil && val >= 0 {
			cooldownSecs = val
		}
	}

	return timeoutSecs, cooldownSecs
}

func runValidate(stdinData []byte) {
	timeoutSecs, cooldownSecs := loadValidateConfig()
	debug := os.Getenv("CLAUDE_HOOKS_DEBUG") == "1"

	exitCode := hooks.ValidateWithSkipCheck(
		context.Background(),
		stdinData,
		os.Stdout,
		os.Stderr,
		debug,
		timeoutSecs,
		cooldownSecs,
	)
	os.Exit(exitCode)
}

func writeDebugLog(args []string, stdinData []byte) {
	debugFile := getDebugLogPath()

	f, err := os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	_, _ = fmt.Fprintf(f, "\n========================================\n")
	_, _ = fmt.Fprintf(f, "[%s] cc-tools invoked\n", timestamp)
	_, _ = fmt.Fprintf(f, "Args: %v\n", args)
	_, _ = fmt.Fprintf(f, "  CLAUDE_HOOKS_DEBUG: %s\n", os.Getenv("CLAUDE_HOOKS_DEBUG"))

	if wd, wdErr := os.Getwd(); wdErr == nil {
		_, _ = fmt.Fprintf(f, "  Working Dir: %s\n", wd)
	}

	if len(stdinData) > 0 {
		_, _ = fmt.Fprintf(f, "Stdin: %s\n", string(stdinData))
	} else {
		_, _ = fmt.Fprintf(f, "Stdin: (no data)\n")
	}
}

// getDebugLogPath returns the debug log path for the current directory.
func getDebugLogPath() string {
	wd, err := os.Getwd()
	if err != nil {
		// Fallback to generic log if we can't get working directory
		return "/tmp/cc-tools.debug"
	}
	return shared.GetDebugLogPathForDir(wd)
}
