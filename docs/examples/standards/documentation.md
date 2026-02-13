# Documentation Standards

This document defines Go documentation standards including godoc comments, error handling documentation, and general commenting practices.

## Package Documentation

Every package must have a package comment that describes its purpose and functionality:

```go
// Package statusline provides a customizable terminal statusline
// for displaying Claude Code session information.
//
// The statusline supports multiple themes, real-time metrics,
// and various display modes. It can be configured through
// configuration files or command-line flags.
package statusline
```

## Function and Method Documentation

### Exported Functions

All exported functions must be documented starting with the function name:

```go
// NewStatusLine creates a new statusline instance with the given configuration.
// It validates the configuration and initializes all required components.
//
// The configuration must specify a valid theme and display settings.
// If the configuration is invalid, an error is returned.
//
// Example:
//
//	cfg := &Config{Theme: "powerline", Width: 80}
//	sl, err := NewStatusLine(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewStatusLine(cfg *Config) (*StatusLine, error) {
    // Implementation
}
```

### Parameters and Returns

Document complex functions with parameter and return descriptions:

```go
// ValidateResource validates a GCP resource against compliance policies.
//
// Parameters:
//   - ctx: Context for cancellation and tracing
//   - resource: The resource to validate
//   - constraintPath: File path to policy constraints
//
// Returns:
//   - A slice of Violation objects if policies are violated
//   - An error if validation fails
func ValidateResource(ctx context.Context, resource *Resource, constraintPath string) ([]*Violation, error) {
    // Implementation
}
```

## Type Documentation

### Structs

Document the purpose and usage of types:

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

### Struct Fields

All exported struct fields need documentation:

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

## Interface Documentation

Document interface purpose and expected behavior:

```go
// Renderer defines the interface for statusline rendering engines.
// Implementations must be thread-safe as Render may be called
// concurrently from multiple goroutines.
type Renderer interface {
    // Render formats the status data according to the configured theme.
    // Returns an error if the data cannot be rendered.
    Render(data *StatusData) (string, error)

    // SetTheme updates the rendering theme.
    // The theme change takes effect on the next Render call.
    SetTheme(theme Theme)
}
```

## Constants and Variables

Document the purpose and usage of exported constants:

```go
// Control type constants define the categories of violations that can be detected.
// These types are used to categorize findings sent to Security Command Center.
const (
    // CaiControlType represents Cloud Asset Inventory based policy controls
    CaiControlType = "CAI"

    // IamPolicyAnalyserControlType represents IAM policy analysis controls
    IamPolicyAnalyserControlType = "IAM_POLICY_ANALYZER"
)

// SeverityToString maps severity enum values to their string representations.
var SeverityToString = map[Severity]string{
    SeverityUnspecified: "UNSPECIFIED",
    SeverityLow:         "LOW",
    SeverityMedium:      "MEDIUM",
    SeverityHigh:        "HIGH",
    SeverityCritical:    "CRITICAL",
}
```

## Error Documentation

### Error Variables

Document cc-tools errors clearly:

```go
var (
    // ErrConfigNotFound is returned when the configuration file cannot be located.
    ErrConfigNotFound = errors.New("configuration file not found")

    // ErrInvalidTheme is returned when an unsupported theme is specified.
    ErrInvalidTheme = errors.New("invalid theme specified")

    // ErrSessionExpired is returned when the Claude Code session has expired.
    ErrSessionExpired = errors.New("session expired")
)
```

### Error Return Documentation

Document error conditions in function comments:

```go
// ProcessResource processes a resource for compliance validation.
//
// Returns:
//   - ErrInvalidResource: If the resource data is malformed
//   - ErrConstraintViolation: If the resource violates constraints
//   - ErrContextCanceled: If the operation was canceled through context
func ProcessResource(ctx context.Context, resource *Resource) error {
    // Implementation
}
```

## Implementation Comments

### Complex Logic

Document non-obvious algorithms or approaches:

```go
func calculateMetrics(data []float64) Statistics {
    // The algorithm uses Welford's online algorithm for calculating
    // variance in a single pass. This approach is numerically stable
    // and memory efficient for streaming data.
    //
    // Reference: https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance

    var mean, m2 float64
    for i, x := range data {
        delta := x - mean
        mean += delta / float64(i+1)
        m2 += delta * (x - mean)
    }

    return Statistics{
        Mean:     mean,
        Variance: m2 / float64(len(data)-1),
    }
}
```

### Warnings and Important Notes

Use standard comment markers for important information:

```go
// TODO(username): Implement retry logic for transient failures
// FIXME: This assumes UTC timezone, need to handle user's local time
// NOTE: This function is not thread-safe and must be called with a lock held
// DEPRECATED: Use NewRenderer instead. Will be removed in v2.0.0
```

## Test Documentation

Document test purpose and verification points:

```go
// TestValidateResource tests the ValidateResource function with different
// types of resources and constraint templates. It verifies that:
//   - Compliant resources produce no violations
//   - Non-compliant resources produce the expected violations
//   - Invalid inputs result in appropriate errors
func TestValidateResource(t *testing.T) {
    // Test implementation
}
```

## Example Documentation

Provide executable examples for complex functionality:

```go
// This example demonstrates how to create and start a statusline
// with a custom theme and refresh interval.
func ExampleStatusLine_Start() {
    cfg := &Config{
        Theme:           "powerline",
        RefreshInterval: 2,
    }

    sl, err := NewStatusLine(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := sl.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Output:
    // Statusline started with powerline theme
}
```

## Documentation Style Guide

### General Rules

1. **Complete sentences**: Use proper grammar and punctuation
2. **Start with the name**: Begin with the element being documented
3. **Present tense**: Use "returns" not "will return"
4. **Be concise**: Avoid unnecessary verbosity
5. **Be precise**: Include specific requirements and constraints

### What to Document

- **Why**: Explain the purpose and reasoning
- **What**: Describe behavior and effects
- **How**: Include usage examples for complex features
- **When**: Note any timing or ordering requirements
- **Errors**: Document error conditions and types

### What NOT to Document

- **Obvious code**: Don't document simple getters/setters
- **Implementation details**: Focus on behavior, not internals
- **Redundant information**: Don't repeat what the code clearly shows

## Godoc Best Practices

### Package Organization

- Place package comment immediately before `package` declaration
- No blank lines between comment and package statement
- Include a blank line after the package statement

### Formatting

- Use indentation for code examples
- Use blank lines to separate paragraphs
- URLs are automatically converted to links
- Use backticks for inline code references

### Links and References

```go
// See [Config] for configuration options.
// This implements the [io.Reader] interface.
// For more details, visit https://example.com/docs
```

## Maintenance

### When to Update Documentation

- **API changes**: Update when function signatures change
- **Behavior changes**: Document new behavior or constraints
- **Bug fixes**: Note any behavior corrections
- **Deprecations**: Mark deprecated items clearly

### Documentation Reviews

- Review documentation during code reviews
- Ensure examples still compile and run
- Update outdated references
- Remove documentation for deleted code
