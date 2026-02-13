# quanta Coding Guidelines

> Comprehensive coding guidelines for the quanta project - a CLI tool that displays Claude Code session status in the terminal.

## Table of Contents

1. [Project Philosophy](#project-philosophy)
2. [Project Structure](#project-structure)
3. [Go Language Standards](#go-language-standards)
4. [Naming Conventions](#naming-conventions)
5. [Error Handling](#error-handling)
6. [Documentation Standards](#documentation-standards)
7. [Testing Standards](#testing-standards)
8. [CLI Development](#cli-development)
9. [Performance Guidelines](#performance-guidelines)
10. [Security Practices](#security-practices)

## Project Philosophy

### Core Principles

- **Simplicity First**: Favor simple, obvious solutions over clever ones
- **Explicit Over Implicit**: Make intentions clear in code
- **Composition Over Inheritance**: Use interfaces and embedding
- **Early Returns**: Reduce nesting with guard clauses
- **Small Functions**: Keep functions focused and under 50 lines
- **Testability**: Design code to be easily testable

### Go Idioms We Follow

- Errors are values, not exceptions
- Don't communicate by sharing memory; share memory by communicating
- The zero value should be useful
- Accept interfaces, return concrete types
- Make the zero value useful
- A little copying is better than a little dependency

## Project Structure

```text
quanta/
├── go.mod                   # Module definition
├── go.sum                   # Dependency checksums
├── main.go                  # Application entry point
├── cmd/                     # Command implementations (Cobra)
│   ├── root.go              # Root command setup
│   ├── start.go             # Start statusline command
│   ├── stop.go              # Stop statusline command
│   ├── config.go            # Configuration management command
│   └── version.go           # Version information command
├── internal/                # Private application code
│   ├── statusline/          # Core statusline logic
│   │   ├── statusline.go    # Main statusline implementation
│   │   ├── renderer.go      # Rendering engine
│   │   ├── themes.go        # Theme definitions
│   │   └── metrics.go       # Metrics collection
│   ├── config/              # Configuration management
│   │   ├── config.go        # Configuration structures
│   │   ├── loader.go        # Configuration loading logic
│   │   └── validator.go     # Configuration validation
│   ├── claude/              # Claude Code integration
│   │   ├── client.go        # Claude Code API client
│   │   ├── monitor.go       # Session monitoring
│   │   └── types.go         # Claude-specific types
│   ├── display/             # Terminal display
│   │   ├── terminal.go      # Terminal manipulation
│   │   ├── colors.go        # Color management
│   │   └── layout.go        # Layout calculations
│   └── utils/               # Utility functions
│       ├── format.go        # Formatting helpers
│       └── system.go        # System utilities
├── pkg/                     # Public packages (if any)
├── configs/                 # Configuration files
│   ├── default.yaml         # Default configuration
│   └── themes/              # Theme configurations
├── docs/examples/           # Usage examples
├── docs/                    # Documentation
├── scripts/                 # Build and utility scripts
├── testdata/                # Test fixtures
└── .github/                 # GitHub configuration
    └── workflows/           # CI/CD workflows
```

### Directory Guidelines

- **`/internal`**: Use for all private application code
- **`/cmd`**: Keep command files small, delegate to internal packages
- **`/pkg`**: Only create if you explicitly want to share code
- **`/testdata`**: Store all test fixtures here
- **Never create empty directories** - add them when needed

## Go Language Standards

### Language Version

- **Target Go 1.25** for development
- **Use Go 1.23** for production builds (latest stable)
- Enable Go modules: `GO111MODULE=on`

### Code Formatting

```bash
# Install required tools
task tools-install

# Format code
task fmt          # Runs gofmt and goimports
task polish       # Format, auto-fix lint issues, clean backup files
```

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"
    "io"

    // Third-party packages
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "go.uber.org/zap"

    // Internal packages
    "github.com/user/quanta/internal/config"
    "github.com/user/quanta/internal/statusline"
)
```

## Naming Conventions

### Packages

- Use singular nouns: `user` not `users`
- Short, lowercase: `config` not `configuration`
- No underscores or mixed caps

### Files

- Lowercase with underscores: `status_line.go`
- Test files: `status_line_test.go`
- Group related functionality

### Go Identifiers

```go
// Exported (public)
type StatusLine struct {}
func NewStatusLine() *StatusLine {}
const MaxRetries = 3

// Unexported (private)
type renderer struct {}
func parseConfig() error {}
const defaultTimeout = 30
```

### Interfaces

```go
// Good - behavior focused
type Reader interface {}
type Validator interface {}
type Renderer interface {}

// Bad - noun focused or Hungarian notation
type IReader interface {}
type ReaderInterface interface {}
type DataInterface interface {}
```

## Error Handling

### Error Creation and Wrapping

```go
// Define quanta errors
var (
    ErrConfigNotFound = errors.New("configuration file not found")
    ErrInvalidTheme   = errors.New("invalid theme specified")
)

// Wrap errors with context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &cfg, nil
}
```

### Error Checking

```go
// Check and handle immediately
resp, err := client.Get(url)
if err != nil {
    return fmt.Errorf("failed to fetch data: %w", err)
}
defer resp.Body.Close()

// Check specific errors
if errors.Is(err, ErrConfigNotFound) {
    // Handle missing config
}

// Check error types
var syntaxErr *json.SyntaxError
if errors.As(err, &syntaxErr) {
    // Handle syntax error
}
```

### Error Messages

- Start with lowercase (unless proper noun)
- No punctuation at the end
- Include context about what failed
- Be actionable when possible

```go
// Good
"failed to connect to database"
"invalid configuration: theme 'dark' not found"
"timeout waiting for response"

// Bad
"Error connecting to database."
"Failed"
"Something went wrong"
```

## Documentation Standards

### Package Documentation

```go
// Package statusline provides a customizable terminal statusline
// for displaying Claude Code session information.
//
// The statusline supports multiple themes, real-time metrics,
// and various display modes. It can be configured through
// configuration files or command-line flags.
package statusline
```

### Function Documentation

```go
// NewStatusLine creates a new statusline instance with the given configuration.
// It validates the configuration and initializes all required components.
//
// The configuration must specify a valid theme and display settings.
// If the configuration is invalid, an error is returned.
//
// Example:
//
//  cfg := &Config{Theme: "powerline", Width: 80}
//  sl, err := NewStatusLine(cfg)
//  if err != nil {
//      log.Fatal(err)
//  }
func NewStatusLine(cfg *Config) (*StatusLine, error) {
    // Implementation
}
```

### Type Documentation

```go
// StatusLine represents an active terminal statusline display.
// It manages the rendering loop, metrics collection, and theme application.
//
// A StatusLine must be started with Start() before it begins displaying,
// and should be stopped with Stop() to clean up resources.
type StatusLine struct {
    config   *Config
    renderer *renderer
    metrics  *metricsCollector
    stop     chan struct{}
}
```

### Field Documentation

```go
type Config struct {
    // Theme specifies the visual theme for the statusline.
    // Valid values are: "default", "powerline", "minimal", "classic"
    Theme string `json:"theme" yaml:"theme"`

    // RefreshInterval is the time between statusline updates in seconds.
    // Must be between 1 and 60. Default is 1.
    RefreshInterval int `json:"refresh_interval" yaml:"refresh_interval"`

    // Width is the maximum width of the statusline in characters.
    // Use 0 for automatic terminal width detection.
    Width int `json:"width" yaml:"width"`
}
```

## Testing Standards

### Table-Driven Tests

```go
func TestRenderer_Render(t *testing.T) {
    tests := []struct {
        name    string
        input   *StatusData
        theme   Theme
        want    string
        wantErr bool
    }{
        {
            name: "default theme with basic data",
            input: &StatusData{
                SessionID: "abc123",
                Status:    "active",
            },
            theme: DefaultTheme,
            want:  "Session: abc123 | Status: active",
        },
        {
            name:    "nil input returns error",
            input:   nil,
            theme:   DefaultTheme,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := NewRenderer(tt.theme)
            got, err := r.Render(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Mock Setup

```go
func TestClaudeClient_GetSession(t *testing.T) {
    tests := []struct {
        name      string
        sessionID string
        setupMock func(*mocks.MockHTTPClient)
        want      *Session
        wantErr   bool
    }{
        {
            name:      "successful fetch",
            sessionID: "test-123",
            setupMock: func(m *mocks.MockHTTPClient) {
                response := &http.Response{
                    StatusCode: 200,
                    Body:       io.NopCloser(strings.NewReader(`{"id":"test-123","status":"active"}`)),
                }
                m.On("Get", "https://api.claude.ai/sessions/test-123").Return(response, nil)
            },
            want: &Session{
                ID:     "test-123",
                Status: "active",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(mocks.MockHTTPClient)
            tt.setupMock(mockClient)

            client := NewClaudeClient(mockClient)
            got, err := client.GetSession(tt.sessionID)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }

            mockClient.AssertExpectations(t)
        })
    }
}
```

### Test Organization

- Keep unit tests next to code
- Use `testdata/` for fixtures
- Group integration tests separately
- Run tests in parallel when possible
- Use `t.Helper()` for test utilities

### Database Isolation

Test isolation prevents state leakage between tests. Follow these patterns:

**Unit Tests - Use In-Memory SQLite**
```go
func setupTestStorage(t *testing.T) interfaces.Storage {
    t.Helper()
    storage, err := NewSQLiteStorage(":memory:")
    require.NoError(t, err)

    t.Cleanup(func() {
        _ = storage.Close()
    })

    return storage
}
```

**Integration Tests - Use Temporary Directories**
```go
func setupIntegrationDB(t *testing.T) (*Storage, string) {
    t.Helper()
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    storage, err := NewSQLiteStorage(dbPath)
    require.NoError(t, err)

    t.Cleanup(func() {
        _ = storage.Close()
    })

    return storage, dbPath
}
```

**Viper State Isolation**
```go
t.Run("test case", func(t *testing.T) {
    viper.Reset()  // Clear global state
    viper.Set("analyze.db", ":memory:")  // Use in-memory DB
    // ... test code
})
```

**Rules**:
- Never reference `~/.quanta/quanta.db` in tests
- Always use `t.Cleanup()` or `defer` for resource cleanup
- Use `t.TempDir()` for file-based tests (auto-cleanup)
- Reset Viper state when tests modify configuration

## CLI Development

### Command Structure (Cobra)

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "quanta",
    Short: "Display Claude Code session status",
    Long: `quanta provides a customizable terminal statusline
for monitoring your Claude Code sessions with real-time metrics
and multiple theme options.`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Initialize configuration
        return initConfig()
    },
}

// cmd/start.go
var startCmd = &cobra.Command{
    Use:   "start [flags]",
    Short: "Start the statusline display",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.Load()
        if err != nil {
            return fmt.Errorf("failed to load config: %w", err)
        }

        sl, err := statusline.New(cfg)
        if err != nil {
            return fmt.Errorf("failed to create statusline: %w", err)
        }

        return sl.Start(cmd.Context())
    },
}
```

### Configuration Management (Viper)

```go
func initConfig() error {
    // Set defaults
    viper.SetDefault("theme", "default")
    viper.SetDefault("refresh_interval", 1)
    viper.SetDefault("colors.background", "#000000")
    viper.SetDefault("colors.foreground", "#ffffff")

    // Set config search paths
    viper.SetConfigName(".quanta")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("$HOME")
    viper.AddConfigPath(".")

    // Environment variables
    viper.SetEnvPrefix("CC_STATUSLINE")
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    viper.AutomaticEnv()

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return fmt.Errorf("failed to read config: %w", err)
        }
    }

    return nil
}
```

### User Interaction

```go
// Respect NO_COLOR environment variable
if os.Getenv("NO_COLOR") != "" {
    color.NoColor = true
}

// Provide helpful error messages
if err := cmd.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    fmt.Fprintf(os.Stderr, "Run 'quanta --help' for usage.\n")
    os.Exit(1)
}

// Use appropriate exit codes
const (
    ExitSuccess = 0
    ExitError   = 1
    ExitUsage   = 2
)
```

## Performance Guidelines

### Optimization Principles

1. **Measure First**: Profile before optimizing
2. **Optimize Hot Paths**: Focus on frequently executed code
3. **Reduce Allocations**: Minimize heap allocations
4. **Use Buffering**: Buffer I/O operations
5. **Concurrent When Beneficial**: Use goroutines for parallel work

### Common Optimizations

```go
// Pre-allocate slices
data := make([]string, 0, expectedSize)

// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

// Buffer channels appropriately
ch := make(chan Event, 100) // Buffered channel

// Use strings.Builder for string concatenation
var b strings.Builder
b.WriteString("Session: ")
b.WriteString(sessionID)
result := b.String()
```

### Profiling

```go
// Enable profiling in development
import _ "net/http/pprof"

func init() {
    if os.Getenv("ENABLE_PROFILING") == "true" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
}

// Profile with: go tool pprof http://localhost:6060/debug/pprof/profile
```

## Security Practices

### Input Validation

```go
// Validate all user input
func SetTheme(theme string) error {
    validThemes := []string{"default", "powerline", "minimal", "classic"}

    for _, valid := range validThemes {
        if theme == valid {
            return nil
        }
    }

    return fmt.Errorf("invalid theme: %s", theme)
}

// Sanitize file paths
func LoadConfigFile(path string) error {
    // Prevent directory traversal
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return errors.New("invalid path: directory traversal detected")
    }

    // Check file exists and is readable
    info, err := os.Stat(cleanPath)
    if err != nil {
        return fmt.Errorf("cannot access config file: %w", err)
    }

    if info.IsDir() {
        return errors.New("path is a directory, not a file")
    }

    return nil
}
```

### Network Client Creation

All HTTP clients MUST use the centralized Tor-aware factory to ensure
traffic routes through Tor when enabled:

- Use `cmd.CreateHTTPClient()` for creating HTTP clients
- Accept `*http.Client` in config structs for dependency injection
- Never use `http.DefaultClient`, `http.Get()`, or `&http.Client{}`
- Use `socks5h://` scheme for SOCKS5 proxies (prevents DNS leaks)

```go
// ✅ DO: Use centralized factory
client, err := cmd.CreateHTTPClient()

// ✅ DO: Accept injected client
type Config struct {
    HTTPClient *http.Client
}

// ❌ DON'T: Create bare clients
client := &http.Client{Timeout: 30 * time.Second}

// ❌ DON'T: Use default client
resp, err := http.Get(url)
```

### Sensitive Data

- Never log sensitive information
- Use environment variables for secrets
- Clear sensitive data from memory after use
- Don't commit secrets to version control

## Code Quality Checklist

### Before Committing

- [ ] Code passes formatting checks (`task fmt`)
- [ ] No linter warnings (`task lint`)
- [ ] All tests pass (`task test`)
- [ ] Race detector passes (`task test-race`)
- [ ] New code has appropriate test coverage (`task coverage`)
- [ ] Documentation is updated for public APIs
- [ ] Error messages are clear and actionable
- [ ] No commented-out code
- [ ] No TODO comments without issue references

**Quick check:** Run `task check` to verify all pre-commit checks pass

### Code Review Focus

- [ ] Follows project structure conventions
- [ ] Uses appropriate error handling
- [ ] Has comprehensive test coverage
- [ ] Documentation is clear and complete
- [ ] No unnecessary complexity
- [ ] Follows naming conventions
- [ ] No security vulnerabilities
- [ ] Performance is acceptable

## Quick Reference

### Common Commands

```bash
# Development
task fmt           # Format code (gofmt and goimports)
task lint          # Run golangci-lint
task test          # Run fast unit tests (-short)
task watch         # Auto-run tests on file changes (TDD essential!)
task test-race     # Run tests with race detector
task coverage      # Generate test coverage report
task build         # Build binary with version info

# Benchmarking
task bench         # Run benchmarks with memory stats

# Environment & Tools
task doctor        # Check development environment
task tools-install # Install all required tools

# Before commit
task check         # Run all checks (fmt + lint + test-race)

# For profiling (still use go commands directly)
go test -cpuprofile=cpu.prof ./...    # CPU profile
go test -memprofile=mem.prof ./...    # Memory profile
go tool pprof cpu.prof                # Analyze profile
```

### Useful Links

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Viper Documentation](https://github.com/spf13/viper)

## References

For detailed information on specific topics, see:

- [docs/research/go-best-practices-2025.md](docs/research/go-best-practices-2025.md)
- [docs/research/go-project-structure.md](docs/research/go-project-structure.md)
- [docs/research/go-cli-development.md](docs/research/go-cli-development.md)
- [docs/research/go-testing-practices.md](docs/research/go-testing-practices.md)
- [docs/examples/](docs/examples/) - Example code and patterns
