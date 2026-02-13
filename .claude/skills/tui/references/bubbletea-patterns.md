# Advanced Bubble Tea Patterns

## Program Options

Configure `tea.NewProgram` in `cmd/tui.go`:

```go
p := tea.NewProgram(app,
    tea.WithMouseCellMotion(),    // Mouse support
    tea.WithFPS(60),              // Higher FPS for animations
)
// Note: In v2, alt screen is enabled via view.AltScreen = true in View()
```

For testing, redirect I/O:

```go
var buf bytes.Buffer
p := tea.NewProgram(model,
    tea.WithInput(&bytes.Buffer{}),
    tea.WithOutput(&buf),
)
```

## Commands (tea.Cmd)

A `tea.Cmd` is a function that returns a `tea.Msg`. The runtime executes it asynchronously.

### Basic command

```go
func loadDataCmd(id string) tea.Cmd {
    return func() tea.Msg {
        data, err := fetchData(id)
        if err != nil {
            return errMsg{err}
        }
        return dataLoadedMsg{data}
    }
}
```

### Batch — run multiple commands concurrently

```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        loadConfigCmd(),
        startTickCmd(),
    )
}
// Note: In v2, alt screen is enabled via view.AltScreen = true in View()
```

### Sequence — run commands in order

```go
cmd := tea.Sequence(
    validateCmd(),   // runs first
    saveCmd(),       // runs after validate completes
    notifyCmd(),     // runs after save completes
)
```

### Returning multiple commands from Update

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        cmds = append(cmds, m.updateLayout())
    case dataMsg:
        m.data = msg.Data
        cmds = append(cmds, m.refreshView())
    }

    // Forward to child
    child, cmd := m.child.Update(msg)
    m.child = child
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}
```

## Subscriptions

### Tick — periodic updates

Used by Quanta's App for refresh:

```go
const tickInterval = time.Second

type TickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

// In Update, re-subscribe:
case TickMsg:
    cmds = append(cmds, tickCmd())
```

### Channel listening

Bridge external events (goroutines, domain callbacks) to the TUI:

```go
func listenForUpdates(ch <-chan Update) tea.Cmd {
    return func() tea.Msg {
        update := <-ch  // blocks until message available
        return updateMsg(update)
    }
}

// In Update, re-subscribe after each message:
case updateMsg:
    m.handleUpdate(msg)
    return m, listenForUpdates(m.updateCh)
```

### Using tea.Every for fixed intervals

```go
type everyMsg time.Time

func (m Model) Init() tea.Cmd {
    return tea.Every(5*time.Second, func(t time.Time) tea.Msg {
        return everyMsg(t)
    })
}
```

## Progress Callbacks from Domain Logic

The `RegressionRunner.RunWithProgress` pattern uses `send func(tea.Msg)`:

```go
// In cmd/tui.go setup:
container.SetRunnerFactory(func(cfg *regression.Config) tea.Cmd {
    runner := tui.NewRegressionRunner(cfg, nil)
    ctx := context.Background()
    // p.Send is safe to call from any goroutine
    return runner.RunWithProgress(ctx, p.Send)
})

// In the runner:
func (rr *RegressionRunner) RunWithProgress(ctx context.Context, send func(tea.Msg)) tea.Cmd {
    return func() tea.Msg {
        // send() delivers messages mid-execution
        send(TestStartedMsg{ContractName: "Foo"})

        runner.SetProgressCallback(func(completed, total int, cost, profit float64) {
            send(RegressionProgressMsg{Completed: completed, Total: total})
        })

        summary, err := runner.Run(ctx, tests)
        // Final return is also a tea.Msg
        return RegressionCompleteMsg{...}
    }
}
```

## Window Management

### Alt screen

In v2, alt screen is controlled via the View return value, not commands:

```go
func (m Model) View() tea.View {
    view := tea.NewView(m.renderContent())
    view.AltScreen = true
    return view
}
```

### Responsive layout

Always handle `tea.WindowSizeMsg`:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.ready = true
    m.updateLayout()
```

Propagate to children via `SetSize`:

```go
func (a *App) updateLayout() {
    a.header.SetWidth(a.width)
    a.statusBar.SetWidth(a.width)
    contentHeight := a.height - appContentOffset
    for _, handler := range a.viewHandlers {
        handler.SetSize(a.width, contentHeight)
    }
}
```

### Cursor management

```go
// Hide cursor (useful for non-input views)
return m, tea.HideCursor

// Show cursor (for text input views)
return m, tea.ShowCursor
```

## Error Handling

Define an error message type and handle gracefully:

```go
type OperationError struct {
    Err     error
    Context string
}

// In Update:
case OperationError:
    m.lastError = msg
    m.showingError = true
    return m, nil

// In View:
if m.showingError {
    errStyle := m.styles.StatusFail
    return errStyle.Render("Error: " + m.lastError.Context + ": " + m.lastError.Err.Error())
}
```

## Nested tea.Program

For standalone screens like splash animations:

```go
func runSplashThenApp() error {
    // Phase 1: Splash screen
    splashModel := splash.NewNeuralModel()
    splashProgram := tea.NewProgram(splashModel)
    if _, err := splashProgram.Run(); err != nil {
        return err
    }

    // Phase 2: Main app
    app := tui.NewApp()
    mainProgram := tea.NewProgram(app)
    _, err := mainProgram.Run()
    return err
}
```

Or embed as a phase within a single program (preferred — avoids screen flash):

```go
type AppWithSplash struct {
    phase  int // 0=splash, 1=app
    splash splash.NeuralModel
    app    *tui.App
}

func (m AppWithSplash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.phase == 0 {
        if _, ok := msg.(splash.LockedMsg); ok {
            m.phase = 1
            return m, nil
        }
        updated, cmd := m.splash.Update(msg)
        m.splash = updated
        return m, cmd
    }
    return m.app.Update(msg)
}
```
