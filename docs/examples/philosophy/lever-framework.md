# The LEVER Framework for Go Development

The LEVER Framework is our guiding philosophy for building maintainable, scalable Go software. Each letter represents a core principle that drives our development decisions.

## Overview

**Leverage** existing patterns
**Extend** before creating
**Verify** through reactivity
**Eliminate** duplication
**Reduce** complexity

## L - Leverage Existing Patterns

Always look for established patterns and solutions before creating new ones.

### Examples

```go
// ❌ Avoid: Creating custom context management
type CustomContext struct {
    data   map[string]interface{}
    cancel chan struct{}
}

func (c *CustomContext) Set(key string, value interface{}) {
    c.data[key] = value
}

func (c *CustomContext) Get(key string) interface{} {
    return c.data[key]
}

// ✅ Good: Leverage existing patterns (context.Context)
import "context"

func ProcessRequest(ctx context.Context, userID string) error {
    ctx = context.WithValue(ctx, "userID", userID)
    return doWork(ctx)
}
```

### Leveraging Standard Library

```go
// ❌ Avoid: Reimplementing standard functionality
func CustomDebounce(f func(), delay time.Duration) func() {
    var timer *time.Timer
    return func() {
        if timer != nil {
            timer.Stop()
        }
        timer = time.AfterFunc(delay, f)
    }
}

// ✅ Good: Use context.Context with timeout for debouncing
func SearchWithDebounce(ctx context.Context, query string, delay time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, delay)
    defer cancel()

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(delay):
        return performSearch(query)
    }
}
```

### Leveraging Go Patterns

```go
// ❌ Avoid: Complex custom validation
func IsValidUser(obj interface{}) bool {
    v := reflect.ValueOf(obj)
    if v.Kind() != reflect.Struct {
        return false
    }

    // Complex reflection-based validation...
    return true
}

// ✅ Good: Leverage interfaces and type assertions
type Validator interface {
    Validate() error
}

type User struct {
    ID    string `validate:"required"`
    Email string `validate:"required,email"`
    Age   int    `validate:"min=0,max=120"`
}

func (u User) Validate() error {
    return validator.New().Struct(u)
}

func IsValidUser(u User) error {
    return u.Validate()
}
```

## E - Extend Before Creating

When existing solutions don't quite fit, extend them rather than starting from scratch.

### Extending Existing Interfaces

```go
// ❌ Avoid: Creating a new interface from scratch
type CustomLogger interface {
    Info(msg string)
    Error(msg string)
    Debug(msg string)
    Warn(msg string)
    SetLevel(level string)
    AddContext(key, value string)
}

// ✅ Good: Extend existing standard interfaces
import "log/slog"

type EnhancedLogger interface {
    slog.Logger // Embed standard logger
    WithContext(ctx context.Context) EnhancedLogger
    WithComponent(component string) EnhancedLogger
}

type enhancedLogger struct {
    *slog.Logger
    ctx       context.Context
    component string
}

func (e *enhancedLogger) WithContext(ctx context.Context) EnhancedLogger {
    return &enhancedLogger{
        Logger:    e.Logger,
        ctx:       ctx,
        component: e.component,
    }
}

func (e *enhancedLogger) WithComponent(component string) EnhancedLogger {
    return &enhancedLogger{
        Logger:    e.Logger.With("component", component),
        ctx:       e.ctx,
        component: component,
    }
}
```

### Extending Structs with Embedding

```go
// ✅ Good: Extend base functionality through embedding
type BaseEntity struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
    BaseEntity        // Embedded struct
    Email      string `json:"email"`
    Name       string `json:"name"`
}

type Session struct {
    BaseEntity        // Same base functionality
    UserID     string `json:"user_id"`
    Status     string `json:"status"`
    ExpiresAt  time.Time `json:"expires_at"`
}

// Extend error types
type AppError struct {
    error                 // Embed standard error
    Code    string       `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (e AppError) WithDetail(key string, value interface{}) AppError {
    if e.Details == nil {
        e.Details = make(map[string]interface{})
    }
    e.Details[key] = value
    return e
}
```

## V - Verify Through Reactivity

Make systems self-verifying through reactive patterns and observable state.

### Reactive Validation with Channels

```go
// ✅ Good: Reactive form validation using channels
type FormValidator struct {
    errors   chan ValidationError
    fields   map[string]interface{}
    watchers map[string][]chan<- ValidationResult
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ValidationResult struct {
    Field string `json:"field"`
    Valid bool   `json:"valid"`
    Error string `json:"error,omitempty"`
}

func NewFormValidator() *FormValidator {
    return &FormValidator{
        errors:   make(chan ValidationError, 10),
        fields:   make(map[string]interface{}),
        watchers: make(map[string][]chan<- ValidationResult),
    }
}

func (fv *FormValidator) SetField(name string, value interface{}) {
    fv.fields[name] = value

    // Validate and notify watchers
    go func() {
        result := fv.validateField(name, value)
        for _, watcher := range fv.watchers[name] {
            select {
            case watcher <- result:
            default: // Non-blocking
            }
        }
    }()
}

func (fv *FormValidator) Watch(field string) <-chan ValidationResult {
    ch := make(chan ValidationResult, 1)
    fv.watchers[field] = append(fv.watchers[field], ch)
    return ch
}
```

### Reactive State Management with Context

```go
// ✅ Good: Self-verifying state with context cancellation
type StatusMonitor struct {
    ctx      context.Context
    cancel   context.CancelFunc
    status   chan Status
    watchers []chan<- Status
    mu       sync.RWMutex
}

type Status struct {
    SessionID string    `json:"session_id"`
    Active    bool      `json:"active"`
    LastSeen  time.Time `json:"last_seen"`
    Theme     string    `json:"theme"`
}

func NewStatusMonitor(ctx context.Context) *StatusMonitor {
    ctx, cancel := context.WithCancel(ctx)
    return &StatusMonitor{
        ctx:    ctx,
        cancel: cancel,
        status: make(chan Status, 10),
    }
}

func (sm *StatusMonitor) Start() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-sm.ctx.Done():
            return
        case <-ticker.C:
            status := sm.getCurrentStatus()
            sm.notifyWatchers(status)
        case newStatus := <-sm.status:
            sm.notifyWatchers(newStatus)
        }
    }
}

func (sm *StatusMonitor) Subscribe() <-chan Status {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    ch := make(chan Status, 1)
    sm.watchers = append(sm.watchers, ch)
    return ch
}
```

### Reactive Effects with Goroutines

```go
// ✅ Good: Side effects that react to state changes
func (sl *StatusLine) StartReactiveUpdates(ctx context.Context) {
    configChanges := sl.config.Watch()
    statusChanges := sl.monitor.Subscribe()

    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case config := <-configChanges:
                sl.handleConfigChange(config)
            case status := <-statusChanges:
                sl.handleStatusChange(status)
            }
        }
    }()
}

// Self-verifying API client with circuit breaker
type APIClient struct {
    client     *http.Client
    baseURL    string
    failures   int64
    lastError  time.Time
    circuitOpen bool
    mu         sync.RWMutex
}

func (c *APIClient) Get(ctx context.Context, path string) (*http.Response, error) {
    if c.isCircuitOpen() {
        return nil, errors.New("circuit breaker is open")
    }

    resp, err := c.client.Get(c.baseURL + path)
    if err != nil {
        c.recordFailure()
        return nil, fmt.Errorf("request failed: %w", err)
    }

    c.recordSuccess()
    return resp, nil
}

func (c *APIClient) recordFailure() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.failures++
    c.lastError = time.Now()

    if c.failures >= 5 {
        c.circuitOpen = true
        // Auto-recover after 30 seconds
        time.AfterFunc(30*time.Second, func() {
            c.mu.Lock()
            defer c.mu.Unlock()
            c.circuitOpen = false
            c.failures = 0
        })
    }
}
```

## E - Eliminate Duplication

Remove duplication of knowledge, not just code. Remember: duplicate code is cheaper than the wrong abstraction.

### Knowledge Duplication (Bad)

```go
// ❌ Same business rule in multiple places
type OrderService struct{}

func (os *OrderService) CalculateShipping(order Order) float64 {
    if order.Total > 50.0 { // Free shipping rule
        return 0
    }
    return 5.99
}

func (oc *OrderComponent) RenderShipping(order Order) string {
    if order.Total > 50.0 { // Same rule duplicated
        return "Free shipping!"
    }
    return "Shipping: $5.99"
}

func (es *EmailService) SendOrderConfirmation(order Order) error {
    shipping := 5.99
    if order.Total > 50.0 { // Rule duplicated again
        shipping = 0
    }
    // Send email...
    return nil
}

// ✅ Good: Single source of truth
const (
    FreeShippingThreshold = 50.0
    StandardShippingCost  = 5.99
)

type ShippingCalculator struct{}

func (sc *ShippingCalculator) Calculate(orderTotal float64) float64 {
    if orderTotal > FreeShippingThreshold {
        return 0
    }
    return StandardShippingCost
}

func (sc *ShippingCalculator) IsFree(orderTotal float64) bool {
    return orderTotal > FreeShippingThreshold
}

// Now all services use the same calculator
var shipping = &ShippingCalculator{}
```

### Acceptable Code Duplication

```go
// ✅ These look similar but represent different knowledge
func validateAge(age int) error {
    if age < 18 || age > 100 { // Legal age requirement
        return errors.New("invalid age")
    }
    return nil
}

func validateExperience(years int) error {
    if years < 0 || years > 50 { // Career experience range
        return errors.New("invalid experience")
    }
    return nil
}

func validateRating(rating int) error {
    if rating < 1 || rating > 5 { // Star rating system
        return errors.New("invalid rating")
    }
    return nil
}

// Don't abstract these - they're different business rules
// that happen to look similar
```

## R - Reduce Complexity

Always strive for the simplest solution that works.

### Simplify Control Flow

```go
// ❌ Complex nested conditions
func processUser(user *User) string {
    if user != nil {
        if user.IsActive {
            if user.Subscription != nil {
                if user.Subscription.Status == "active" {
                    if user.Subscription.Plan == "premium" {
                        return "premium-features"
                    } else {
                        return "basic-features"
                    }
                } else {
                    return "renew-subscription"
                }
            } else {
                return "create-subscription"
            }
        } else {
            return "activate-account"
        }
    } else {
        return "login-required"
    }
}

// ✅ Good: Early returns and clear logic
func processUser(user *User) string {
    if user == nil {
        return "login-required"
    }
    if !user.IsActive {
        return "activate-account"
    }
    if user.Subscription == nil {
        return "create-subscription"
    }
    if user.Subscription.Status != "active" {
        return "renew-subscription"
    }

    if user.Subscription.Plan == "premium" {
        return "premium-features"
    }
    return "basic-features"
}
```

### Simplify Data Structures

```go
// ❌ Complex nested structure
type ComplexAppState struct {
    UI struct {
        Modals struct {
            Settings struct {
                IsOpen bool
                Tabs   struct {
                    General  struct{ IsDirty bool }
                    Advanced struct{ IsDirty bool }
                }
            }
        }
    }
}

// ✅ Good: Flattened, focused structures
type SettingsModal struct {
    IsOpen            bool   `json:"is_open"`
    ActiveTab         string `json:"active_tab"`
    HasUnsavedChanges bool   `json:"has_unsaved_changes"`
}

type AppState struct {
    SettingsModal SettingsModal `json:"settings_modal"`
    // Other state slices...
}
```

### Simplify Functions

```go
// ❌ Complex multi-purpose function
func handleUserAction(action string, data interface{}, options map[string]interface{}) (interface{}, error) {
    switch action {
    case "create":
        // 20 lines of code
    case "update":
        // 30 lines of code
    case "delete":
        // 15 lines of code
        // ... more cases
    }
    return nil, nil
}

// ✅ Good: Simple, focused functions
func createUser(ctx context.Context, data CreateUserRequest) (*User, error) {
    if err := data.Validate(); err != nil {
        return nil, fmt.Errorf("invalid user data: %w", err)
    }
    return userRepo.Create(ctx, data)
}

func updateUser(ctx context.Context, id string, data UpdateUserRequest) (*User, error) {
    if err := data.Validate(); err != nil {
        return nil, fmt.Errorf("invalid update data: %w", err)
    }
    return userRepo.Update(ctx, id, data)
}

func deleteUser(ctx context.Context, id string) error {
    return userRepo.Delete(ctx, id)
}
```

## Applying LEVER in Practice

### Example: Building a StatusLine System

```go
// L - Leverage: Use existing patterns (context, channels, interfaces)
import (
    "context"
    "fmt"
    "log/slog"
    "time"
)

// E - Extend: Build on standard interfaces
type StatusRenderer interface {
    fmt.Stringer  // Embed standard interface
    Render(ctx context.Context, status Status) string
    SetTheme(theme string) error
}

// V - Verify: Reactive monitoring
type StatusLineService struct {
    logger   *slog.Logger
    renderer StatusRenderer
    monitor  *StatusMonitor
    updates  chan Status
}

func NewStatusLineService(logger *slog.Logger, renderer StatusRenderer) *StatusLineService {
    return &StatusLineService{
        logger:   logger,
        renderer: renderer,
        monitor:  NewStatusMonitor(context.Background()),
        updates:  make(chan Status, 10),
    }
}

func (sls *StatusLineService) Start(ctx context.Context) error {
    // Emit events for reactive systems
    sls.logger.Info("starting statusline service")

    go func() {
        for {
            select {
            case <-ctx.Done():
                sls.logger.Info("statusline service stopped")
                return
            case status := <-sls.monitor.Subscribe():
                sls.handleStatusUpdate(ctx, status)
            }
        }
    }()

    return nil
}

// E - Eliminate: Single source of status logic
func (sls *StatusLineService) handleStatusUpdate(ctx context.Context, status Status) {
    rendered := sls.renderer.Render(ctx, status)
    sls.logger.Debug("status updated", "output", rendered)

    // Notify other parts of the system
    select {
    case sls.updates <- status:
    default: // Non-blocking
    }
}

// R - Reduce: Simple, focused API
func (sls *StatusLineService) GetCurrentStatus(ctx context.Context) (Status, error) {
    // Simple implementation - delegate to monitor
    return sls.monitor.GetCurrent(), nil
}

func (sls *StatusLineService) UpdateTheme(theme string) error {
    // Simple validation and update
    if theme == "" {
        return errors.New("theme cannot be empty")
    }
    return sls.renderer.SetTheme(theme)
}
```

## StatusLine-Specific Examples

### Configuration Management

```go
// L - Leverage: Use Viper for configuration
// E - Extend: Add reactive configuration watching
type ConfigManager struct {
    *viper.Viper           // Embed Viper
    watchers []chan<- Config
    mu       sync.RWMutex
}

// V - Verify: Reactive configuration updates
func (cm *ConfigManager) Watch() <-chan Config {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    ch := make(chan Config, 1)
    cm.watchers = append(cm.watchers, ch)

    // Watch for file changes
    cm.WatchConfig()
    cm.OnConfigChange(func(e fsnotify.Event) {
        config := cm.unmarshalConfig()
        for _, watcher := range cm.watchers {
            select {
            case watcher <- config:
            default:
            }
        }
    })

    return ch
}
```

## Key Takeaways

1. **Leverage**: Don't reinvent the wheel. Use standard library, established patterns, and existing interfaces.
2. **Extend**: When existing solutions are close but not perfect, extend them rather than starting over.
3. **Verify**: Build systems that validate themselves through reactive patterns and observable state.
4. **Eliminate**: Remove duplication of business knowledge, but don't be afraid of code that looks similar but represents different concepts.
5. **Reduce**: Always choose the simpler solution. Complex code is a liability.

Remember: "The best code is no code. The second best code is code that already exists and works."
