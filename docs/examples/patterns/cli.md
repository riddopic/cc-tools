# CLI Development Patterns for cc-tools

## Command Structure with Cobra

### Root Command Setup

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/riddopic/cc-tools/internal/config"
)

var (
    cfgFile string
    verbose bool
    rootCmd = &cobra.Command{
        Use:   "cc-tools",
        Short: "Claude Code integration tools for hooks, validation, and debugging",
        Long: `cc-tools provides utilities for managing Claude Code hooks,
running validation checks, debugging configurations, and managing
MCP servers with comprehensive logging and testing capabilities.

Examples:
  cc-tools validate                 # Run all validation checks
  cc-tools hooks list               # List all configured hooks
  cc-tools config show              # Show current configuration
  cc-tools debug                    # Start debug mode`,
        Version: "1.0.0",
    }
)

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)

    // Global flags
    rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
        "config file (default: $HOME/.cc-tools.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
        "verbose output")

    // Bind flags to viper
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

    // Set up custom help and usage templates
    rootCmd.SetHelpTemplate(customHelpTemplate)
    rootCmd.SetUsageTemplate(customUsageTemplate)
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Warning: Could not find home directory: %v\n", err)
        }

        // Search config in home directory and current directory
        viper.AddConfigPath(home)
        viper.AddConfigPath(".")
        viper.SetConfigName(".cc-tools")
        viper.SetConfigType("yaml")
    }

    // Environment variables
    viper.SetEnvPrefix("CC_TOOLS")
    viper.AutomaticEnv()

    // Read config file
    if err := viper.ReadInConfig(); err == nil {
        if verbose {
            fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
        }
    }
}
```

### Subcommand Pattern

```go
// cmd/start.go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/riddopic/cc-tools/internal/statusline"
)

var (
    theme           string
    refreshInterval int
    position        string
)

var startCmd = &cobra.Command{
    Use:   "start [flags]",
    Short: "Start the statusline display",
    Long:  `Start displaying the Claude Code statusline in your terminal.`,
    Example: `  cc-tools start
  cc-tools start --theme powerline
  cc-tools start --theme minimal --refresh 2
  cc-tools start --position bottom`,
    PreRunE: validateStartFlags,
    RunE:    runStart,
}

func init() {
    rootCmd.AddCommand(startCmd)

    // Command-specific flags
    startCmd.Flags().StringVarP(&theme, "theme", "t", "default",
        "statusline theme (default|powerline|minimal|classic)")
    startCmd.Flags().IntVarP(&refreshInterval, "refresh", "r", 1,
        "refresh interval in seconds (1-60)")
    startCmd.Flags().StringVarP(&position, "position", "p", "bottom",
        "statusline position (top|bottom)")

    // Register flag completion
    startCmd.RegisterFlagCompletionFunc("theme", themeCompletion)
    startCmd.RegisterFlagCompletionFunc("position", positionCompletion)
}

func validateStartFlags(cmd *cobra.Command, args []string) error {
    // Validate theme
    validThemes := []string{"default", "powerline", "minimal", "classic"}
    if !contains(validThemes, theme) {
        return fmt.Errorf("invalid theme: %s. Valid themes: %v", theme, validThemes)
    }

    // Validate refresh interval
    if refreshInterval < 1 || refreshInterval > 60 {
        return fmt.Errorf("refresh interval must be between 1 and 60 seconds")
    }

    // Validate position
    validPositions := []string{"top", "bottom"}
    if !contains(validPositions, position) {
        return fmt.Errorf("invalid position: %s. Valid positions: %v", position, validPositions)
    }

    return nil
}

func runStart(cmd *cobra.Command, args []string) error {
    // Create configuration
    cfg := &statusline.Config{
        Theme:           theme,
        RefreshInterval: refreshInterval,
        Position:        position,
    }

    // Create statusline
    sl, err := statusline.New(cfg)
    if err != nil {
        return fmt.Errorf("failed to create statusline: %w", err)
    }

    // Set up signal handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("\nShutting down...")
        cancel()
    }()

    // Start statusline
    fmt.Printf("Starting statusline with theme '%s'...\n", theme)
    return sl.Run(ctx)
}

func themeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"default", "powerline", "minimal", "classic"}, cobra.ShellCompDirectiveNoFileComp
}

func positionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"top", "bottom"}, cobra.ShellCompDirectiveNoFileComp
}
```

### Configuration Management Commands

```go
// cmd/config.go
package cmd

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage cc-tools configuration",
    Long:  `View, edit, and validate cc-tools configuration.`,
}

var configShowCmd = &cobra.Command{
    Use:   "show",
    Short: "Show current configuration",
    Long:  `Display the current configuration including defaults and overrides.`,
    RunE:  runConfigShow,
}

var configInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize configuration file",
    Long:  `Create a new configuration file with default settings.`,
    RunE:  runConfigInit,
}

var configValidateCmd = &cobra.Command{
    Use:   "validate [file]",
    Short: "Validate configuration file",
    Long:  `Check if a configuration file is valid.`,
    Args:  cobra.MaximumNArgs(1),
    RunE:  runConfigValidate,
}

func init() {
    rootCmd.AddCommand(configCmd)
    configCmd.AddCommand(configShowCmd)
    configCmd.AddCommand(configInitCmd)
    configCmd.AddCommand(configValidateCmd)

    // Flags for config show
    configShowCmd.Flags().StringP("format", "f", "yaml", "output format (yaml|json)")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
    format, _ := cmd.Flags().GetString("format")

    // Get all settings
    settings := viper.AllSettings()

    switch format {
    case "json":
        encoder := json.NewEncoder(os.Stdout)
        encoder.SetIndent("", "  ")
        return encoder.Encode(settings)
    case "yaml":
        encoder := yaml.NewEncoder(os.Stdout)
        encoder.SetIndent(2)
        return encoder.Encode(settings)
    default:
        return fmt.Errorf("unsupported format: %s", format)
    }
}

func runConfigInit(cmd *cobra.Command, args []string) error {
    home, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("could not find home directory: %w", err)
    }

    configPath := filepath.Join(home, ".cc-tools.yaml")

    // Check if file already exists
    if _, err := os.Stat(configPath); err == nil {
        fmt.Printf("Configuration file already exists at %s\n", configPath)
        fmt.Print("Overwrite? (y/N): ")

        var response string
        fmt.Scanln(&response)
        if response != "y" && response != "Y" {
            fmt.Println("Cancelled.")
            return nil
        }
    }

    // Create default configuration
    defaultConfig := `# cc-tools configuration
theme: default
refresh_interval: 1
position: bottom

# Theme colors
colors:
  background: "#1e1e1e"
  foreground: "#ffffff"
  accent: "#007acc"
  success: "#4caf50"
  warning: "#ff9800"
  error: "#f44336"

# Metrics to display
metrics:
  cpu: true
  memory: true
  network: false
  disk: false

# Display settings
display:
  width: 0  # 0 for auto-detect
  show_time: true
  show_session_id: true
  time_format: "15:04:05"

# Claude Code integration
claude:
  api_endpoint: "https://api.claude.ai"
  poll_interval: 5

# Logging
logging:
  level: info
  file: ""  # Empty for stdout
`

    if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }

    fmt.Printf("Configuration file created at %s\n", configPath)
    return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
    var configPath string

    if len(args) > 0 {
        configPath = args[0]
    } else if cfgFile != "" {
        configPath = cfgFile
    } else {
        configPath = viper.ConfigFileUsed()
    }

    if configPath == "" {
        return fmt.Errorf("no configuration file specified")
    }

    // Try to load and validate the config
    v := viper.New()
    v.SetConfigFile(configPath)

    if err := v.ReadInConfig(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }

    // Validate specific fields
    cfg := &config.Config{}
    if err := v.Unmarshal(cfg); err != nil {
        return fmt.Errorf("configuration validation failed: %w", err)
    }

    if err := cfg.Validate(); err != nil {
        return fmt.Errorf("configuration validation failed: %w", err)
    }

    fmt.Printf("✓ Configuration file %s is valid\n", configPath)
    return nil
}
```

## Output Formatting Patterns

### Structured Output

```go
// internal/output/formatter.go
package output

import (
    "encoding/json"
    "fmt"
    "io"
    "text/tabwriter"

    "gopkg.in/yaml.v3"
)

type Format string

const (
    FormatText  Format = "text"
    FormatJSON  Format = "json"
    FormatYAML  Format = "yaml"
    FormatTable Format = "table"
)

type Formatter interface {
    Format(data interface{}) error
}

func NewFormatter(w io.Writer, format Format) Formatter {
    switch format {
    case FormatJSON:
        return &JSONFormatter{w: w}
    case FormatYAML:
        return &YAMLFormatter{w: w}
    case FormatTable:
        return &TableFormatter{w: w}
    default:
        return &TextFormatter{w: w}
    }
}

type JSONFormatter struct {
    w io.Writer
}

func (f *JSONFormatter) Format(data interface{}) error {
    encoder := json.NewEncoder(f.w)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}

type YAMLFormatter struct {
    w io.Writer
}

func (f *YAMLFormatter) Format(data interface{}) error {
    encoder := yaml.NewEncoder(f.w)
    encoder.SetIndent(2)
    return encoder.Encode(data)
}

type TableFormatter struct {
    w io.Writer
}

func (f *TableFormatter) Format(data interface{}) error {
    tw := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)

    // Format data as table (implementation depends on data structure)
    switch v := data.(type) {
    case [][]string:
        for _, row := range v {
            for i, col := range row {
                if i > 0 {
                    fmt.Fprint(tw, "\t")
                }
                fmt.Fprint(tw, col)
            }
            fmt.Fprintln(tw)
        }
    default:
        return fmt.Errorf("unsupported data type for table format")
    }

    return tw.Flush()
}

type TextFormatter struct {
    w io.Writer
}

func (f *TextFormatter) Format(data interface{}) error {
    _, err := fmt.Fprintln(f.w, data)
    return err
}
```

### Color and Style Support

```go
// internal/display/colors.go
package display

import (
    "fmt"
    "os"

    "github.com/fatih/color"
)

var (
    InfoColor    = color.New(color.FgCyan)
    SuccessColor = color.New(color.FgGreen)
    WarningColor = color.New(color.FgYellow)
    ErrorColor   = color.New(color.FgRed)
    BoldColor    = color.New(color.Bold)
)

func init() {
    // Respect NO_COLOR environment variable
    if os.Getenv("NO_COLOR") != "" {
        color.NoColor = true
    }

    // Force color output if requested
    if os.Getenv("FORCE_COLOR") != "" {
        color.NoColor = false
    }
}

func Info(format string, a ...interface{}) {
    InfoColor.Printf(format, a...)
}

func Success(format string, a ...interface{}) {
    SuccessColor.Printf(format, a...)
}

func Warning(format string, a ...interface{}) {
    WarningColor.Printf(format, a...)
}

func Error(format string, a ...interface{}) {
    ErrorColor.Printf(format, a...)
}

func Bold(format string, a ...interface{}) {
    BoldColor.Printf(format, a...)
}
```

## Interactive Elements

### Progress Indicators

```go
// internal/display/progress.go
package display

import (
    "fmt"
    "time"

    "github.com/briandowns/spinner"
    "github.com/schollz/progressbar/v3"
)

type ProgressIndicator interface {
    Start()
    Stop()
    Update(message string)
}

type Spinner struct {
    spinner *spinner.Spinner
}

func NewSpinner(message string) *Spinner {
    s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
    s.Suffix = " " + message
    return &Spinner{spinner: s}
}

func (s *Spinner) Start() {
    s.spinner.Start()
}

func (s *Spinner) Stop() {
    s.spinner.Stop()
}

func (s *Spinner) Update(message string) {
    s.spinner.Suffix = " " + message
}

type ProgressBar struct {
    bar *progressbar.ProgressBar
}

func NewProgressBar(total int, description string) *ProgressBar {
    bar := progressbar.NewOptions(total,
        progressbar.OptionEnableColorCodes(true),
        progressbar.OptionShowBytes(false),
        progressbar.OptionSetDescription(description),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "[green]█[reset]",
            SaucerHead:    "[green]█[reset]",
            SaucerPadding: " ",
            BarStart:      "[",
            BarEnd:        "]",
        }),
    )
    return &ProgressBar{bar: bar}
}

func (p *ProgressBar) Add(n int) {
    p.bar.Add(n)
}

func (p *ProgressBar) Finish() {
    p.bar.Finish()
}
```

### User Prompts

```go
// internal/display/prompt.go
package display

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/manifoldco/promptui"
)

func Confirm(message string) bool {
    prompt := promptui.Prompt{
        Label:     message,
        IsConfirm: true,
    }

    _, err := prompt.Run()
    return err == nil
}

func Select(label string, items []string) (string, error) {
    prompt := promptui.Select{
        Label: label,
        Items: items,
    }

    _, result, err := prompt.Run()
    return result, err
}

func Input(label string, defaultValue string) (string, error) {
    prompt := promptui.Prompt{
        Label:   label,
        Default: defaultValue,
    }

    return prompt.Run()
}

func Password(label string) (string, error) {
    prompt := promptui.Prompt{
        Label: label,
        Mask:  '*',
    }

    return prompt.Run()
}

// Simple confirmation without external dependencies
func SimpleConfirm(message string) bool {
    reader := bufio.NewReader(os.Stdin)
    fmt.Printf("%s (y/N): ", message)

    response, err := reader.ReadString('\n')
    if err != nil {
        return false
    }

    response = strings.ToLower(strings.TrimSpace(response))
    return response == "y" || response == "yes"
}
```

## Shell Completion

### Bash Completion

```go
// cmd/completion.go
package cmd

import (
    "os"

    "github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate shell completion script",
    Long: `To load completions:

Bash:
  $ source <(cc-tools completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ cc-tools completion bash > /etc/bash_completion.d/cc-tools
  # macOS:
  $ cc-tools completion bash > /usr/local/etc/bash_completion.d/cc-tools

Zsh:
  $ source <(cc-tools completion zsh)

  # To load completions for each session, execute once:
  $ cc-tools completion zsh > "${fpath[1]}/_cc-tools"

Fish:
  $ cc-tools completion fish | source

  # To load completions for each session, execute once:
  $ cc-tools completion fish > ~/.config/fish/completions/cc-tools.fish

PowerShell:
  PS> cc-tools completion powershell | Out-String | Invoke-Expression

  # To load completions for each session, add to PowerShell profile:
  PS> cc-tools completion powershell >> $PROFILE
`,
    DisableFlagsInUseLine: true,
    ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
    Args:                  cobra.ExactValidArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        switch args[0] {
        case "bash":
            cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            cmd.Root().GenZshCompletion(os.Stdout)
        case "fish":
            cmd.Root().GenFishCompletion(os.Stdout, true)
        case "powershell":
            cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
        }
    },
}

func init() {
    rootCmd.AddCommand(completionCmd)
}
```

## Error Handling and User Feedback

### User-Friendly Error Messages

```go
// internal/errors/user_errors.go
package errors

import (
    "fmt"
    "strings"
)

type UserError struct {
    Message     string
    Suggestion  string
    ExitCode    int
}

func (e *UserError) Error() string {
    var b strings.Builder
    b.WriteString(e.Message)
    if e.Suggestion != "" {
        b.WriteString("\n\nSuggestion: ")
        b.WriteString(e.Suggestion)
    }
    return b.String()
}

func NewUserError(message, suggestion string, exitCode int) *UserError {
    return &UserError{
        Message:    message,
        Suggestion: suggestion,
        ExitCode:   exitCode,
    }
}

// Common error constructors
func ConfigNotFound(path string) *UserError {
    return NewUserError(
        fmt.Sprintf("Configuration file not found: %s", path),
        "Run 'cc-tools config init' to create a default configuration",
        2,
    )
}

func InvalidTheme(theme string) *UserError {
    return NewUserError(
        fmt.Sprintf("Invalid theme: %s", theme),
        "Available themes: default, powerline, minimal, classic",
        2,
    )
}

func ConnectionFailed(endpoint string) *UserError {
    return NewUserError(
        fmt.Sprintf("Failed to connect to %s", endpoint),
        "Check your internet connection and Claude Code session",
        3,
    )
}
```

## Testing CLI Commands

### Command Testing

```go
// cmd/start_test.go
package cmd

import (
    "bytes"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestStartCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid theme",
            args:    []string{"start", "--theme", "powerline"},
            wantErr: false,
        },
        {
            name:    "invalid theme",
            args:    []string{"start", "--theme", "invalid"},
            wantErr: true,
            errMsg:  "invalid theme",
        },
        {
            name:    "invalid refresh interval",
            args:    []string{"start", "--refresh", "100"},
            wantErr: true,
            errMsg:  "refresh interval must be between",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewRootCommand()
            buf := new(bytes.Buffer)
            cmd.SetOut(buf)
            cmd.SetErr(buf)
            cmd.SetArgs(tt.args)

            err := cmd.Execute()

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Best Practices Summary

1. **Clear Command Hierarchy**: Organize commands logically
2. **Comprehensive Help**: Provide examples and detailed descriptions
3. **Input Validation**: Validate flags and arguments early
4. **Error Messages**: Make errors actionable with suggestions
5. **Color Support**: Respect NO_COLOR environment variable
6. **Shell Completion**: Generate completion scripts for all shells
7. **Configuration**: Support files, environment variables, and flags
8. **Progress Feedback**: Show progress for long operations
9. **Graceful Shutdown**: Handle signals properly
10. **Testing**: Test commands, not just functions
