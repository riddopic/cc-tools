# Quanta TUI Architecture

## Architecture Overview

```
cmd/tui.go
  └─ tui.NewApp()                    # App model (tea.Model)
       ├─ components.HeaderModel     # Top bar
       ├─ components.StatusBarModel  # Bottom bar
       └─ viewHandlers map           # ViewType → ViewHandler
            ├─ ViewDashboard → (placeholder)
            ├─ ViewRegression → RegressionContainer
            ├─ ViewAnalyze → (placeholder)
            ├─ ViewHistory → (placeholder)
            └─ ViewSettings → (placeholder)

RegressionContainer (ViewHandler)
  ├─ phase: Config | Running | Results
  ├─ RegressionConfig (sub-view)
  │    └─ components.FormModel
  ├─ Regression (sub-view, results table)
  └─ components.ProfitListModel
```

**Key files:**
- `internal/tui/app.go` — App model, message routing, layout
- `internal/tui/messages.go` — ViewHandler interface, all message types
- `internal/tui/keys.go` — KeyMap with help.KeyMap interface
- `internal/tui/styles/theme.go` — Theme (AdaptiveColor palette)
- `internal/tui/styles/styles.go` — Styles (lipgloss.Style collection)
- `cmd/tui.go` — Entry point, view registration, tea.NewProgram

## Recipe: Creating a New View

### Step 1: Define message types

Add to `internal/tui/messages.go`:

```go
// AnalyzeStreamMsg delivers streaming LLM output.
type AnalyzeStreamMsg struct {
    Chunk string
}

// AnalyzeCompleteMsg signals analysis completion.
type AnalyzeCompleteMsg struct {
    Success bool
    Error   error
}
```

Name messages as `<Noun><Verb>Msg`. Include all data the view needs — messages are the only way to communicate.

### Step 2: Add ViewType (if needed)

ViewTypes are already defined in `messages.go`:

```go
const (
    ViewDashboard ViewType = iota
    ViewRegression
    ViewAnalyze
    // ...
)
```

### Step 3: Implement ViewHandler

Create `internal/tui/views/analyze_container.go`:

```go
package views

import (
    tea "charm.land/bubbletea/v2"
    "github.com/riddopic/quanta/internal/tui"
    "github.com/riddopic/quanta/internal/tui/styles"
)

// Compile-time check.
var _ tui.ViewHandler = (*AnalyzeContainer)(nil)

type AnalyzeContainer struct {
    width  int
    height int
    styles *styles.Styles
    // sub-views and state...
}

func NewAnalyzeContainer(s *styles.Styles) *AnalyzeContainer {
    return &AnalyzeContainer{
        styles: s,
    }
}

func (ac *AnalyzeContainer) HandleMessage(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        ac.SetSize(msg.Width, msg.Height)
    case tui.AnalyzeStreamMsg:
        // handle streaming chunk
    case tui.AnalyzeCompleteMsg:
        // handle completion
    case tea.KeyMsg:
        return ac.handleKeys(msg)
    }
    return nil
}

func (ac *AnalyzeContainer) View() string {
    if ac.width == 0 {
        return "Loading..."
    }
    // render using ac.styles
    return ""
}

func (ac *AnalyzeContainer) SetSize(width, height int) {
    ac.width = width
    ac.height = height
    // propagate to sub-components
}
```

### Step 4: Register in cmd/tui.go

```go
analyzeContainer := views.NewAnalyzeContainer(&s)
app.RegisterView(tui.ViewAnalyze, analyzeContainer)
```

The App handles message forwarding to the active view's `HandleMessage`.

## Recipe: Creating a New Component

Components are reusable UI elements that don't implement `ViewHandler`. They follow a simpler pattern:

```go
package components

import (
    tea "charm.land/bubbletea/v2"
    "charm.land/lipgloss/v2"
    "github.com/riddopic/quanta/internal/tui/styles"
)

type MetricsPanel struct {
    width  int
    height int
    styles *styles.Styles
    data   MetricsData
}

func NewMetricsPanel(s *styles.Styles) MetricsPanel {
    return MetricsPanel{styles: s}
}

func (m *MetricsPanel) SetWidth(w int)          { m.width = w }
func (m *MetricsPanel) SetHeight(h int)         { m.height = h }
func (m *MetricsPanel) SetData(d MetricsData)   { m.data = d }

func (m *MetricsPanel) Update(msg tea.Msg) (*MetricsPanel, tea.Cmd) {
    // handle messages relevant to this component
    return m, nil
}

func (m *MetricsPanel) View() string {
    panel := m.styles.Panel.Width(m.width)
    header := m.styles.PanelHeader.Render("METRICS")
    content := m.renderContent()
    return panel.Render(lipgloss.JoinVertical(lipgloss.Left, header, content))
}
```

Parent views call `SetWidth`/`SetHeight` when they receive `tea.WindowSizeMsg`, and call `View()` in their own `View()`.

## Multi-Phase Views

The `RegressionContainer` demonstrates the multi-phase pattern:

```go
type RegressionContainer struct {
    phase   tui.RegressionPhase  // Config | Running | Results
    config  *RegressionConfig    // Phase 1 sub-view
    results *Regression          // Phase 2-3 sub-view
}

func (rc *RegressionContainer) View() string {
    switch rc.phase {
    case tui.RegressionPhaseConfig:
        return rc.config.View()
    case tui.RegressionPhaseRunning, tui.RegressionPhaseResults:
        return rc.results.View()
    default:
        return rc.config.View()
    }
}
```

Phase transitions happen via messages:

```go
case tui.RegressionStartMsg:
    rc.phase = tui.RegressionPhaseRunning
    rc.results = NewRegression(rc.styles)
    rc.results.SetSize(rc.width, rc.height)
    if rc.runnerFactory != nil {
        cfg := rc.config.BuildConfig()
        return rc, rc.runnerFactory(cfg)
    }

case tui.RegressionCompleteMsg:
    rc.phase = tui.RegressionPhaseResults
```

## Async Integration

The `RegressionRunner` adapter bridges domain logic to TUI messages:

```go
// RunnerFactory creates a tea.Cmd from config.
type RunnerFactory func(cfg *regression.Config) tea.Cmd

// Container stores the factory.
container.SetRunnerFactory(func(cfg *regression.Config) tea.Cmd {
    runner := tui.NewRegressionRunner(cfg, nil)
    ctx := context.Background()
    return runner.Run(ctx)
})
```

For progress updates, use `send func(tea.Msg)`:

```go
func (rr *RegressionRunner) RunWithProgress(ctx context.Context, send func(tea.Msg)) tea.Cmd {
    return func() tea.Msg {
        // Send incremental updates
        for _, test := range tests {
            send(TestStartedMsg{ContractName: test.Name})
        }

        runner.SetProgressCallback(func(completed, total int, cost, profit float64) {
            send(RegressionProgressMsg{
                Completed: completed,
                Total:     total,
                Cost:      cost,
                Profit:    profit,
            })
        })

        // Final message returned from the Cmd
        return RegressionCompleteMsg{...}
    }
}
```

The `send` function is `p.Send()` from `tea.Program` — it safely sends messages from any goroutine.

## Message Design

- **Naming**: `<Noun><Verb>Msg` — `RegressionUpdateMsg`, `AnalyzeCompleteMsg`, `TestStartedMsg`
- **Data completeness**: Include everything the handler needs. Messages are the single source of truth.
- **Phase messages**: Use `<Feature>PhaseChangeMsg` for state machine transitions.
- **Simple signals**: Use empty structs for events with no data: `type RegressionStartMsg struct{}`

## Testing TUI Components

Test via message injection, not visual output:

```go
func TestRegressionContainer_PhaseTransition(t *testing.T) {
    s := styles.NewStyles(styles.DefaultTheme())
    container := views.NewRegressionContainer(&s)

    // Verify initial phase
    assert.Equal(t, tui.RegressionPhaseConfig, container.Phase())

    // Simulate form submission
    container.HandleMessage(components.FormSubmitMsg{})
    assert.Equal(t, tui.RegressionPhaseRunning, container.Phase())

    // Simulate completion
    container.HandleMessage(tui.RegressionCompleteMsg{TotalTests: 5})
    assert.Equal(t, tui.RegressionPhaseResults, container.Phase())
}

func TestRegressionContainer_View(t *testing.T) {
    s := styles.NewStyles(styles.DefaultTheme())
    container := views.NewRegressionContainer(&s)
    container.SetSize(120, 40)

    view := container.View()
    assert.Contains(t, view, "REGRESSION CONFIGURATION")
}
```
