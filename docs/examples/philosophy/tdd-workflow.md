# Test-Driven Development (TDD) Workflow Examples

This document provides concrete examples of the TDD workflow following the Red-Green-Refactor cycle as outlined in the [docs/CODING_GUIDELINES.md](../../docs/CODING_GUIDELINES.md).

## Core TDD Principles

**REMEMBER**: No production code without a failing test first. This is the fundamental practice that enables all other principles.

## Example 1: StatusLine Renderer Feature

Let's build a statusline renderer feature from scratch using TDD.

### Step 1: Red - Write a Failing Test

Start with the simplest behavior:

```go
// renderer_test.go
package statusline

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestStatusRenderer(t *testing.T) {
    t.Run("should render a basic status", func(t *testing.T) {
        renderer := NewRenderer()
        status := Status{
            SessionID: "abc123",
            Active:    true,
            Theme:     "default",
        }

        result := renderer.Render(context.Background(), status)

        assert.Contains(t, result, "abc123")
        assert.Contains(t, result, "●") // Active indicator
    })
}
```

**Test fails**: `NewRenderer` and `Render` don't exist yet.

### Step 2: Green - Minimal Implementation

Write ONLY what's needed to make the test pass:

```go
// renderer.go
package statusline

import "context"

type Status struct {
    SessionID string
    Active    bool
    Theme     string
}

type Renderer struct{}

func NewRenderer() *Renderer {
    return &Renderer{}
}

func (r *Renderer) Render(ctx context.Context, status Status) string {
    return status.SessionID + " ●"
}
```

**Test passes!** ✅

### Step 3: Commit Working Code

```bash
git add .
git commit -m "feat: implement basic status rendering"
```

### Step 4: Red - Add Test for Inactive Status

```go
t.Run("should render inactive status differently", func(t *testing.T) {
    renderer := NewRenderer()
    status := Status{
        SessionID: "xyz789",
        Active:    false,
        Theme:     "default",
    }

    result := renderer.Render(context.Background(), status)

    assert.Contains(t, result, "xyz789")
    assert.Contains(t, result, "○") // Inactive indicator
    assert.NotContains(t, result, "●") // Should not contain active indicator
})
```

**Test fails**: Current implementation always shows active indicator.

### Step 5: Green - Add Status Logic

```go
func (r *Renderer) Render(ctx context.Context, status Status) string {
    indicator := "○"
    if status.Active {
        indicator = "●"
    }
    return status.SessionID + " " + indicator
}
```

**Both tests pass!** ✅

### Step 6: Refactor - Assess and Improve

Now that tests are green, assess if refactoring would add value:

```go
// Extract indicator logic for clarity
const (
    ActiveIndicator   = "●"
    InactiveIndicator = "○"
)

func (r *Renderer) getStatusIndicator(active bool) string {
    if active {
        return ActiveIndicator
    }
    return InactiveIndicator
}

func (r *Renderer) Render(ctx context.Context, status Status) string {
    indicator := r.getStatusIndicator(status.Active)
    return status.SessionID + " " + indicator
}
```

**Tests still pass!** ✅

### Step 7: Commit Refactoring

```bash
git add .
git commit -m "refactor: extract status indicator logic"
```

## Example 2: Configuration Loader with TDD

### Step 1: Red - Test Configuration Loading

```go
// config_test.go
package config

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConfigLoader(t *testing.T) {
    t.Run("should load configuration from file", func(t *testing.T) {
        // Create temporary config file
        tempDir := t.TempDir()
        configPath := filepath.Join(tempDir, "config.yaml")
        configContent := `
theme: "powerline"
refresh_rate: "2s"
position: "top"
`
        require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

        loader := NewLoader()
        config, err := loader.Load(configPath)

        assert.NoError(t, err)
        assert.Equal(t, "powerline", config.Theme)
        assert.Equal(t, "2s", config.RefreshRate)
        assert.Equal(t, "top", config.Position)
    })
}
```

**Test fails**: Components don't exist.

### Step 2: Green - Minimal Implementation

```go
// config.go
package config

import (
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Theme       string `yaml:"theme"`
    RefreshRate string `yaml:"refresh_rate"`
    Position    string `yaml:"position"`
}

type Loader struct{}

func NewLoader() *Loader {
    return &Loader{}
}

func (l *Loader) Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

**Test passes!** ✅

### Step 3: Red - Test Error Handling

```go
t.Run("should handle missing file gracefully", func(t *testing.T) {
    loader := NewLoader()
    config, err := loader.Load("nonexistent.yaml")

    assert.Error(t, err)
    assert.Nil(t, config)
    assert.Contains(t, err.Error(), "no such file")
})
```

**Test passes!** ✅ (Go's os.ReadFile already handles this)

### Step 4: Red - Test Invalid YAML

```go
t.Run("should handle invalid YAML", func(t *testing.T) {
    tempDir := t.TempDir()
    configPath := filepath.Join(tempDir, "invalid.yaml")
    invalidContent := `
theme: "powerline
refresh_rate: [invalid yaml
`
    require.NoError(t, os.WriteFile(configPath, []byte(invalidContent), 0644))

    loader := NewLoader()
    config, err := loader.Load(configPath)

    assert.Error(t, err)
    assert.Nil(t, config)
    assert.Contains(t, err.Error(), "yaml")
})
```

**Test passes!** ✅ (yaml.Unmarshal handles this)

### Step 5: Red - Test Default Values

```go
t.Run("should provide default values for missing fields", func(t *testing.T) {
    tempDir := t.TempDir()
    configPath := filepath.Join(tempDir, "minimal.yaml")
    minimalContent := `theme: "simple"`
    require.NoError(t, os.WriteFile(configPath, []byte(minimalContent), 0644))

    loader := NewLoader()
    config, err := loader.Load(configPath)

    assert.NoError(t, err)
    assert.Equal(t, "simple", config.Theme)
    assert.Equal(t, "1s", config.RefreshRate) // Default value
    assert.Equal(t, "bottom", config.Position) // Default value
})
```

**Test fails**: No default values implemented.

### Step 6: Green - Add Default Values

```go
func (l *Loader) Load(path string) (*Config, error) {
    // Set defaults first
    config := Config{
        Theme:       "default",
        RefreshRate: "1s",
        Position:    "bottom",
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

**All tests pass!** ✅

### Step 7: Refactor - Extract Defaults

```go
var DefaultConfig = Config{
    Theme:       "default",
    RefreshRate: "1s",
    Position:    "bottom",
}

func (l *Loader) Load(path string) (*Config, error) {
    config := DefaultConfig // Start with defaults

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }

    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("parsing config YAML: %w", err)
    }

    return &config, nil
}
```

**Tests still pass!** ✅

## Example 3: StatusLine Service with Async Operations

### Step 1: Red - Test Service Startup

```go
// service_test.go
package statusline

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestStatusLineService(t *testing.T) {
    t.Run("should start service successfully", func(t *testing.T) {
        mockRenderer := new(MockRenderer)
        service := NewService(mockRenderer)

        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()

        err := service.Start(ctx)

        assert.NoError(t, err)
    })
}
```

**Test fails**: Components don't exist.

### Step 2: Green - Minimal Implementation

```go
// service.go
package statusline

import "context"

type Service struct {
    renderer Renderer
}

type Renderer interface {
    Render(ctx context.Context, status Status) string
}

func NewService(renderer Renderer) *Service {
    return &Service{renderer: renderer}
}

func (s *Service) Start(ctx context.Context) error {
    return nil // Minimal implementation
}
```

```go
// mocks.go (or use mockery to generate)
package statusline

import (
    "context"

    "github.com/stretchr/testify/mock"
)

type MockRenderer struct {
    mock.Mock
}

func (m *MockRenderer) Render(ctx context.Context, status Status) string {
    args := m.Called(ctx, status)
    return args.String(0)
}
```

**Test passes!** ✅

### Step 3: Red - Test Status Updates

```go
t.Run("should update status periodically", func(t *testing.T) {
    mockRenderer := new(MockRenderer)
    mockRenderer.On("Render", mock.Anything, mock.AnythingOfType("Status")).
        Return("rendered status")

    service := NewService(mockRenderer)

    ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
    defer cancel()

    go service.Start(ctx)

    // Wait for at least one update
    time.Sleep(120 * time.Millisecond)

    mockRenderer.AssertCalled(t, "Render", mock.Anything, mock.AnythingOfType("Status"))
})
```

**Test fails**: No periodic updates implemented.

### Step 4: Green - Add Periodic Updates

```go
func (s *Service) Start(ctx context.Context) error {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            status := Status{
                SessionID: "test-session",
                Active:    true,
                Theme:     "default",
            }
            s.renderer.Render(ctx, status)
        }
    }
}
```

**Test passes!** ✅

### Step 5: Red - Test Context Cancellation

```go
t.Run("should stop gracefully when context is cancelled", func(t *testing.T) {
    mockRenderer := new(MockRenderer)
    service := NewService(mockRenderer)

    ctx, cancel := context.WithCancel(context.Background())

    done := make(chan error, 1)
    go func() {
        done <- service.Start(ctx)
    }()

    // Cancel after short delay
    time.Sleep(10 * time.Millisecond)
    cancel()

    // Should return quickly after cancellation
    select {
    case err := <-done:
        assert.Equal(t, context.Canceled, err)
    case <-time.After(50 * time.Millisecond):
        t.Fatal("service did not stop within timeout")
    }
})
```

**Test passes!** ✅ (Already implemented in the Start method)

### Step 6: Refactor - Extract Status Creation

```go
const DefaultUpdateInterval = 100 * time.Millisecond

func (s *Service) Start(ctx context.Context) error {
    return s.StartWithInterval(ctx, DefaultUpdateInterval)
}

func (s *Service) StartWithInterval(ctx context.Context, interval time.Duration) error {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            status := s.getCurrentStatus()
            s.renderer.Render(ctx, status)
        }
    }
}

func (s *Service) getCurrentStatus() Status {
    return Status{
        SessionID: "test-session", // TODO: Get from actual session
        Active:    true,
        Theme:     "default",
    }
}
```

**Tests still pass!** ✅

## Common TDD Anti-Patterns to Avoid

### ❌ Writing Multiple Tests Before Making First One Pass

```go
// WRONG: Don't write all tests upfront
func TestCalculator(t *testing.T) {
    t.Run("should add two numbers", func(t *testing.T) { /* pending */ })
    t.Run("should subtract two numbers", func(t *testing.T) { /* pending */ })
    t.Run("should multiply two numbers", func(t *testing.T) { /* pending */ })
    t.Run("should divide two numbers", func(t *testing.T) { /* pending */ })
    t.Run("should handle division by zero", func(t *testing.T) { /* pending */ })
}

// Then implement everything at once
```

### ✅ Correct Approach: One Test at a Time

```go
// RIGHT: Write one test, make it pass, then next
func TestCalculator(t *testing.T) {
    t.Run("should add two numbers", func(t *testing.T) {
        calc := NewCalculator()
        result := calc.Add(2, 3)
        assert.Equal(t, 5, result)
    })
}

// Implement Add(), verify test passes
// THEN write the next test
```

### ❌ Writing More Code Than Needed

```go
// Test only requires returning a user with an ID
func TestCreateUser(t *testing.T) {
    t.Run("should create user with ID", func(t *testing.T) {
        service := NewUserService()
        user, err := service.CreateUser(CreateUserRequest{Name: "John"})
        assert.NoError(t, err)
        assert.NotEmpty(t, user.ID)
    })
}

// WRONG: Over-implementing
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    // Don't add these until tests require them!
    if err := s.validateUser(req); err != nil {
        return nil, err
    }
    if err := s.checkDuplicateEmail(req.Email); err != nil {
        return nil, err
    }
    if err := s.hashPassword(req.Password); err != nil {
        return nil, err
    }
    if err := s.sendWelcomeEmail(req.Email); err != nil {
        return nil, err
    }
    s.logUserCreation(req)

    return &User{
        ID:        generateID(),
        Name:      req.Name,
        Email:     req.Email,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}
```

### ✅ Correct Approach: Minimal Implementation

```go
// RIGHT: Just enough to pass the test
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    return &User{
        ID:   "test-id", // Hard-coded is fine until test requires otherwise
        Name: req.Name,
    }, nil
}
```

### ❌ Skipping Refactoring Step

```go
// After making several tests pass
func ProcessOrder(order *Order) (*ProcessedOrder, error) {
    // 200 lines of nested if/else
    // Duplicated validation logic
    // Magic numbers everywhere
    // No clear structure
}

// "It works, ship it!" - WRONG
```

### ✅ Correct Approach: Always Assess Refactoring

```go
// After each green test, ask:
// - Can I extract a function to clarify intent?
// - Are there magic numbers to name?
// - Is there duplication to remove?
// - Would restructuring improve readability?

// If yes to any, refactor while tests stay green
```

## StatusLine-Specific TDD Examples

### Theme Management

```go
func TestThemeManager(t *testing.T) {
    t.Run("should load default theme", func(t *testing.T) {
        manager := NewThemeManager()
        theme, err := manager.GetTheme("default")

        assert.NoError(t, err)
        assert.Equal(t, "default", theme.Name)
        assert.NotEmpty(t, theme.Colors)
    })
}
```

### Session Monitoring

```go
func TestSessionMonitor(t *testing.T) {
    t.Run("should detect active Claude Code session", func(t *testing.T) {
        monitor := NewSessionMonitor()

        session, err := monitor.GetActiveSession(context.Background())

        if err != nil {
            // No active session is also a valid state
            assert.Nil(t, session)
        } else {
            assert.NotEmpty(t, session.ID)
        }
    })
}
```

### Terminal Integration

```go
func TestTerminalRenderer(t *testing.T) {
    t.Run("should render to terminal buffer", func(t *testing.T) {
        var output strings.Builder
        renderer := NewTerminalRenderer(&output)

        status := Status{
            SessionID: "abc123",
            Active:    true,
        }

        err := renderer.RenderToTerminal(context.Background(), status)

        assert.NoError(t, err)
        assert.Contains(t, output.String(), "abc123")
    })
}
```

## TDD Checklist

Before writing any production code, ask yourself:

- [ ] Do I have a failing test that demands this code?
- [ ] Am I writing the minimum code to make the test pass?
- [ ] Have I committed my working code before refactoring?
- [ ] Have I assessed whether refactoring would improve the code?
- [ ] Do all tests still pass after refactoring?
- [ ] Are my tests focused on behavior, not implementation?
- [ ] Am I using table-driven tests where appropriate?
- [ ] Are my test names descriptive and behavior-focused?

## StatusLine Development Workflow

1. **Start with the smallest testable unit** (e.g., status indicator formatting)
2. **Write a failing test** that describes the desired behavior
3. **Implement minimal code** to make the test pass
4. **Commit the working code** with a clear message
5. **Assess refactoring opportunities** (extract functions, rename variables, etc.)
6. **Refactor if it adds value** while keeping tests green
7. **Commit the refactored code** with a descriptive message
8. **Move to the next smallest increment**

## Integration with CI/CD

```bash
# Run tests before every commit
task test

# Run tests with race detection
task test-race

# Run tests with coverage report
task coverage

# Run all pre-commit checks (fmt + lint + test-race)
task check

# Run specific test patterns (still use go test directly)
go test -run TestStatusRenderer ./...
```

## Summary

TDD is not about testing - it's about design. The tests are a beneficial side effect of the design process. By following Red-Green-Refactor strictly, you ensure:

1. **Every line of code has a purpose** (demanded by a test)
2. **Design emerges from actual requirements** (not speculation)
3. **Refactoring is safe** (tests catch regressions)
4. **Documentation is built-in** (tests show how to use the code)
5. **Bugs are caught immediately** (at the moment of creation)

Remember: If you're writing production code without a failing test, you're not doing TDD. Stop and write the test first!
