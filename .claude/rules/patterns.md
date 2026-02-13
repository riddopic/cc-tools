---
paths:
  - "**/*.go"
---

# Go Patterns

Common patterns used in the Quanta codebase.

## Interface Composition

```go
// ✅ DO: Compose interfaces from smaller ones
type Reader interface {
    Read(ctx context.Context, key string) ([]byte, error)
}

type Writer interface {
    Write(ctx context.Context, key string, data []byte) error
}

type Storage interface {
    Reader
    Writer
}

// Lifecycle interfaces (from internal/interfaces/lifecycle.go)
type Shutdowner interface {
    Shutdown(ctx context.Context) error
}

type LRUCacheWithShutdown interface {
    LRUCache
    Shutdowner
}
```

## Cobra Command Structure

```go
// ✅ DO: Use RunE for error handling, PreRunE for validation
var startCmd = &cobra.Command{
    Use:     "start [flags]",
    Short:   "Start the statusline display",
    Example: "  quanta start --theme powerline",
    PreRunE: validateStartFlags,
    RunE:    runStart,
}

func init() {
    rootCmd.AddCommand(startCmd)

    startCmd.Flags().StringVarP(&theme, "theme", "t", "default",
        "statusline theme (default|powerline|minimal)")

    // Enable shell completion
    startCmd.RegisterFlagCompletionFunc("theme", themeCompletion)
}

func validateStartFlags(cmd *cobra.Command, args []string) error {
    validThemes := []string{"default", "powerline", "minimal"}
    if !slices.Contains(validThemes, theme) {
        return fmt.Errorf("invalid theme: %s", theme)
    }
    return nil
}
```

## Viper Configuration

```go
// ✅ DO: Set defaults, support env vars, read config files
func initConfig() error {
    viper.SetDefault("theme", "default")
    viper.SetDefault("refresh_interval", 1)

    viper.SetConfigName(".quanta")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("$HOME")
    viper.AddConfigPath(".")

    viper.SetEnvPrefix("QUANTA")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return fmt.Errorf("config error: %w", err)
        }
    }
    return nil
}
```

## Constructor Pattern

```go
// ✅ DO: Use New* constructors with options
type Service struct {
    client   Client
    timeout  time.Duration
    logger   *zap.Logger
}

func NewService(client Client, opts ...Option) *Service {
    s := &Service{
        client:  client,
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

type Option func(*Service)

func WithTimeout(d time.Duration) Option {
    return func(s *Service) {
        s.timeout = d
    }
}

func WithLogger(l *zap.Logger) Option {
    return func(s *Service) {
        s.logger = l
    }
}
```

## Context Propagation

```go
// ✅ DO: Pass context as first parameter, respect cancellation
func (s *Service) Process(ctx context.Context, data *Data) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    result, err := s.client.Fetch(ctx, data.ID)
    if err != nil {
        return fmt.Errorf("fetch failed: %w", err)
    }

    return s.store.Save(ctx, result)
}

// ✅ DO: Create child contexts for timeouts
func (s *Service) FetchWithTimeout(ctx context.Context, id string) (*Result, error) {
    ctx, cancel := context.WithTimeout(ctx, s.timeout)
    defer cancel()

    return s.client.Fetch(ctx, id)
}
```

## Graceful Shutdown

```go
// ✅ DO: Handle signals and shutdown gracefully
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("\nShutting down...")
        cancel()
    }()

    if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Compile-Time Interface Checks

```go
// ✅ DO: Verify interface implementation at compile time
var _ Runner = (*StatusLine)(nil)
var _ io.Closer = (*StatusLine)(nil)

// Multiple checks
var (
    _ interfaces.ForgeExecutor = (*ForgeRunner)(nil)
    _ interfaces.Shutdowner    = (*ForgeRunner)(nil)
)
```

## Adapter Placement

When `cmd/` code needs to satisfy an `internal/` interface, place the adapter in the appropriate `internal/` package — NOT in `cmd/`.

```go
// ✅ DO: Reusable adapter in internal/
// internal/logger/sugared.go
type SugaredLogger struct { sugar *zap.SugaredLogger }
func (l *SugaredLogger) Info(msg string, args ...any) { l.sugar.Infow(msg, args...) }

// cmd/regression.go — one-line usage
regLogger := applogger.NewSugaredLogger(config.LogLevel, silent)

// ✅ DO: Accept the interface, inject from cmd/
type Service struct {
    logger Logger  // interface from internal/
}

// ❌ DON'T: Define adapter structs in cmd/
// cmd/foo.go
type myLogger struct { sugar *zap.SugaredLogger }
func (l *myLogger) Info(msg string, args ...any) { ... }
```

**Detection checklist** — an adapter is suspect if:

- It lives in `cmd/` and wraps a third-party type
- It implements an `internal/` interface
- It has a factory function only called from one place
- Moving it to `internal/` would make it reusable

## Interface Deduplication

Never define a second interface with the same method set. Reuse the canonical one.

```go
// ✅ DO: Single canonical interface
// internal/interfaces/logger.go
type Logger interface {
    Info(msg string, args ...any)
    Debug(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}

// ✅ DO: Import and reuse
import "github.com/riddopic/quanta/internal/interfaces"
type Config struct {
    Logger interfaces.Logger
}

// ❌ DON'T: Redefine the same interface in another package
// internal/regression/types.go
type Logger interface {
    Info(msg string, args ...any)  // Same signature!
    Debug(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

**Before defining a new interface**, search for existing ones:

```bash
rg "type.*interface" internal/ --glob "*.go" | grep -i <name>
```

## Error Types

```go
// ✅ DO: Define sentinel errors for comparison
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrTimeout      = errors.New("operation timed out")
)

// ✅ DO: Create custom error types for context
type ValidationError struct {
    Field string
    Value interface{}
    Msg   string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Msg)
}

// Check errors with errors.Is and errors.As
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

var valErr ValidationError
if errors.As(err, &valErr) {
    // Handle validation error
}
```

## Worker Pool Pattern

```go
// ✅ DO: Use worker pools for concurrent processing
func ProcessItems(ctx context.Context, items []Item, workers int) error {
    jobs := make(chan Item, len(items))
    results := make(chan error, len(items))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range jobs {
                select {
                case <-ctx.Done():
                    results <- ctx.Err()
                    return
                default:
                    results <- processItem(ctx, item)
                }
            }
        }()
    }

    // Send jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    // Wait and collect errors
    go func() {
        wg.Wait()
        close(results)
    }()

    for err := range results {
        if err != nil {
            return err
        }
    }
    return nil
}
```
