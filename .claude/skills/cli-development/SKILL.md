---
name: cli-development
description: Apply CLI development patterns with Cobra and Viper. Use when creating CLI commands, adding flags, implementing subcommands, managing configuration, or working on terminal UI.
---

# CLI Development Patterns for Quanta

## Command Structure with Cobra

### Root Command Setup

```go
var rootCmd = &cobra.Command{
    Use:   "quanta",
    Short: "Short description",
    Long:  `Detailed description with examples`,
    Version: "1.0.0",
}

func init() {
    cobra.OnInitialize(initConfig)

    // Global flags
    rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
        "config file (default: $HOME/.quanta.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
        "verbose output")

    // Bind flags to viper
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}
```

### Subcommand Pattern

```go
var startCmd = &cobra.Command{
    Use:     "start [flags]",
    Short:   "Start the service",
    Example: `  quanta start --theme powerline`,
    PreRunE: validateFlags,  // Validate before running
    RunE:    runStart,       // Main logic
}

func init() {
    rootCmd.AddCommand(startCmd)

    // Command-specific flags
    startCmd.Flags().StringVarP(&theme, "theme", "t", "default",
        "theme (default|powerline|minimal)")

    // Register completions
    startCmd.RegisterFlagCompletionFunc("theme", themeCompletion)
}
```

## Configuration with Viper

```go
func initConfig() {
    // Set defaults
    viper.SetDefault("theme", "default")
    viper.SetDefault("refresh_interval", 1)

    // Config file locations
    viper.SetConfigName(".quanta")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("$HOME")
    viper.AddConfigPath(".")

    // Environment variables
    viper.SetEnvPrefix("QUANTA")
    viper.AutomaticEnv()

    viper.ReadInConfig()
}
```

## User Feedback Patterns

```go
// Respect NO_COLOR
if os.Getenv("NO_COLOR") != "" {
    color.NoColor = true
}

// User-friendly errors
type UserError struct {
    Message    string
    Suggestion string
    ExitCode   int
}

func (e *UserError) Error() string {
    if e.Suggestion != "" {
        return fmt.Sprintf("%s\n\nSuggestion: %s", e.Message, e.Suggestion)
    }
    return e.Message
}
```

## Signal Handling

```go
func runStart(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("\nShutting down...")
        cancel()
    }()

    return service.Run(ctx)
}
```

## Testing CLI Commands

```go
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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewRootCommand()
            cmd.SetArgs(tt.args)
            err := cmd.Execute()
            // assertions...
        })
    }
}
```

## Advanced CLI Patterns

### Best-Of Flag (Integer with Validation)

```go
var bestOf int

func init() {
    analyzeCmd.Flags().IntVar(&bestOf, "best-of", 1,
        "Run N independent attempts, return best result (1-10)")
}

func validateBestOf(cmd *cobra.Command, args []string) error {
    if bestOf < 1 || bestOf > 10 {
        return fmt.Errorf("best-of must be between 1 and 10, got %d", bestOf)
    }
    return nil
}
```

### Gate Pipeline Flag (Boolean, Default True)

```go
analyzeCmd.Flags().Bool(
    "gate-pipeline",
    true,
    "Enable multi-gate analysis pipeline (Gate 1→2→2.5→3→4)",
)
```

### Early-Termination-Profit Flag (Float with Default)

```go
var earlyTerminationProfit float64

func init() {
    analyzeCmd.Flags().Float64Var(&earlyTerminationProfit, "early-termination-profit", 0.1,
        "Stop Best@N early when profit exceeds threshold in ETH (0 to disable)")
}

func validateEarlyTermination(cmd *cobra.Command, args []string) error {
    if earlyTerminationProfit < 0 {
        return fmt.Errorf("early-termination-profit cannot be negative")
    }
    return nil
}
```

### Combining Flags in PreRunE

```go
var analyzeCmd = &cobra.Command{
    Use:     "analyze [address]",
    PreRunE: func(cmd *cobra.Command, args []string) error {
        if err := validateBestOf(cmd, args); err != nil {
            return err
        }
        if err := validateEarlyTermination(cmd, args); err != nil {
            return err
        }
        // Early termination only makes sense with best-of > 1
        if earlyTerminationProfit > 0 && bestOf == 1 {
            fmt.Println("Note: --early-termination-profit has no effect without --best-of > 1")
        }
        return nil
    },
    RunE: runAnalyze,
}
```

## Detailed CLI Patterns

For complete CLI patterns, see [cli.md](../../../docs/examples/patterns/cli.md)
