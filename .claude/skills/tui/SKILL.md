---
name: tui
description: Build and review TUI views, components, forms, and terminal UI features using the Charmbracelet ecosystem. Use when creating new views, reviewing BubbleTea/Lipgloss code, or any terminal UI work in the Quanta TUI.
---

# TUI Development & Review

**Note:** Quanta uses **Charmbracelet v2** (`charm.land/*/v2` imports).

## Part 1: Building TUI Components

### Quick Reference

| Task | Reference |
|------|-----------|
| New view or component | [references/quanta-tui-architecture.md](references/quanta-tui-architecture.md) |
| Advanced Bubble Tea (commands, subs, program opts) | [references/bubbletea-patterns.md](references/bubbletea-patterns.md) |
| Forms with Huh or custom fields | [references/huh-forms.md](references/huh-forms.md) |
| Markdown rendering in TUI | [references/glamour-markdown.md](references/glamour-markdown.md) |
| Bubbles components (viewport, spinner, table, etc.) | [references/bubbles-components.md](references/bubbles-components.md) |
| Styling, layout, responsive design | [references/lipgloss-theme-system.md](references/lipgloss-theme-system.md) |
| Low-level cell-based rendering | [references/ultraviolet-primitives.md](references/ultraviolet-primitives.md) |

### Charmbracelet v2 Migration

| v1 Pattern | v2 Pattern |
|------------|------------|
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` |
| `tea.KeyMsg` | `tea.KeyPressMsg` |
| `msg.Type == tea.KeyEnter` | `msg.String() == "enter"` |
| `msg.Type == tea.KeyRunes && msg.Runes[0] == ' '` | `msg.String() == "space"` |
| `tea.WithAltScreen()` program option | `view.AltScreen = true` in `View()` |
| `viewport.New(width, height)` | `viewport.New()` then `SetWidth()`/`SetHeight()` |
| `lipgloss.AdaptiveColor` | `compat.AdaptiveColor` from `charm.land/lipgloss/v2/compat` |

**Key string constants** (use these instead of raw strings):
```go
const (
    keyUp    = "up"
    keyDown  = "down"
    keyEnter = "enter"
    keyEsc   = "esc"
    keySpace = "space"
)
```

### Critical Quanta Conventions

| Convention | Details |
|------------|---------|
| **ViewHandler interface** | `HandleMessage(tea.Msg) tea.Cmd`, `View() string`, `SetSize(w, h int)` in `internal/tui/messages.go` |
| **RegisterView pattern** | `app.RegisterView(tui.ViewRegression, container)` in `cmd/tui.go` |
| **Styles singleton** | `styles.GetStyles()` or inject `*styles.Styles` — never create ad-hoc styles |
| **Messages** | Define in `internal/tui/messages.go`, name as `<Noun><Verb>Msg` |
| **Keys** | Define in `internal/tui/keys.go`, use `key.Binding` with `key.WithHelp` |
| **Compile-time check** | `var _ tui.ViewHandler = (*MyContainer)(nil)` |
| **Theme colors** | Use `styles.Theme` `AdaptiveColor` fields, never hardcode ANSI codes |
| **Key events** | Use `tea.KeyPressMsg` and `msg.String()` for key matching (v2) |

### Build Checklist

When creating a new TUI feature:

- [ ] Define message types in `internal/tui/messages.go`
- [ ] Implement `ViewHandler` interface (or component with `SetSize`/`Update`/`View`)
- [ ] Add compile-time interface check
- [ ] Inject `*styles.Styles` — do not call `lipgloss.NewStyle()` ad-hoc in `View()`
- [ ] Handle `tea.WindowSizeMsg` (propagate via `SetSize`)
- [ ] Wire into `cmd/tui.go` via `RegisterView` (for views)
- [ ] Use `tea.Cmd` for all async/IO — never block in `Update`
- [ ] Test via message-driven assertions, not UI pixel matching

### DO / DON'T

| DO | DON'T |
|----|-------|
| Return `tea.Cmd` from `Update` for async work | Call `os.ReadFile`, `http.Get`, or `time.Sleep` in `Update` |
| Define styles in `styles/styles.go` or component init | Create `lipgloss.NewStyle()` inside `View()` |
| Use `lipgloss.JoinVertical`/`JoinHorizontal` for layout | Concatenate strings with `\n` for multi-section layouts |
| Propagate `SetSize` to all child components | Ignore `tea.WindowSizeMsg` |
| Use `tea.Batch` for multiple commands | Return only one command when several are needed |
| Send messages via `func() tea.Msg { return FooMsg{} }` | Call methods directly across component boundaries |
| Use `AdaptiveColor` for light/dark support | Hardcode ANSI color codes |
| Embed domain runner via `send func(tea.Msg)` callback | Run domain logic synchronously in `Update` |

---

## Part 2: Reviewing TUI Code

### CRITICAL: Avoid False Positives

**Read [elm-architecture.md](references/elm-architecture.md) first!** The most common review mistake is flagging correct patterns as bugs.

### NOT Issues (Do NOT Flag These)

| Pattern | Why It's Correct |
|---------|------------------|
| `return m, m.loadData()` | `tea.Cmd` is returned immediately; runtime executes async |
| Value receiver on `Update()` | Standard BubbleTea pattern; model returned by value |
| Nested `m.child, cmd = m.child.Update(msg)` | Normal component composition |
| Helper functions returning `tea.Cmd` | Creates command descriptor, no I/O in Update |
| `tea.Batch(cmd1, cmd2)` | Commands execute concurrently by runtime |

### ACTUAL Issues (DO Flag These)

| Pattern | Why It's Wrong |
|---------|----------------|
| `os.ReadFile()` in Update | Blocks UI thread |
| `http.Get()` in Update | Network I/O blocks |
| `time.Sleep()` in Update | Freezes UI |
| `<-channel` in Update (blocking) | May block indefinitely |
| `huh.Form.Run()` in Update | Blocking call |

### Review Checklist

**Architecture:**
- [ ] **No blocking I/O in Update()** (file, network, sleep)
- [ ] Helper functions returning `tea.Cmd` are NOT flagged as blocking
- [ ] Commands used for all async operations

**Model & Update:**
- [ ] Model is immutable (Update returns new model, not mutates)
- [ ] Init returns proper initial command (or nil)
- [ ] Update handles all expected message types
- [ ] WindowSizeMsg handled for responsive layout
- [ ] tea.Batch used for multiple commands
- [ ] tea.Quit used correctly for exit

**View & Styling:**
- [ ] View is a pure function (no side effects)
- [ ] Lipgloss styles defined once, not in View
- [ ] Key bindings use key.Matches with help.KeyMap

**Components:**
- [ ] Sub-component updates propagated correctly
- [ ] Bubbles components initialized with dimensions
- [ ] Huh forms embedded via Update loop (not Run())

### Critical Patterns

**Model Must Be Immutable:**

```go
// BAD - mutates model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.items = append(m.items, newItem)  // mutation!
    return m, nil
}

// GOOD - returns new model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newItems := make([]Item, len(m.items)+1)
    copy(newItems, m.items)
    newItems[len(m.items)] = newItem
    m.items = newItems
    return m, nil
}
```

**Commands for Async/IO:**

```go
// BAD - blocking in Update
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    data, _ := os.ReadFile("config.json")  // blocks UI!
    m.config = parse(data)
    return m, nil
}

// GOOD - use commands
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, loadConfigCmd()
}

func loadConfigCmd() tea.Cmd {
    return func() tea.Msg {
        data, err := os.ReadFile("config.json")
        if err != nil {
            return errMsg{err}
        }
        return configLoadedMsg{parse(data)}
    }
}
```

**Styles Defined Once:**

```go
// BAD - creates new style each render
func (m Model) View() string {
    style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
    return style.Render("Hello")
}

// GOOD - define styles at package level or in model
var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

func (m Model) View() string {
    return titleStyle.Render("Hello")
}
```

### When to Load References

| You are... | Load |
|------------|------|
| First time reviewing BubbleTea | [elm-architecture.md](references/elm-architecture.md) (prevents false positives) |
| Creating a new view or component | [quanta-tui-architecture.md](references/quanta-tui-architecture.md) |
| Reviewing Update function logic | [model-update.md](references/model-update.md) |
| Reviewing View function, styling | [view-styling.md](references/view-styling.md) |
| Reviewing component hierarchy | [composition.md](references/composition.md) |
| Working with commands, subscriptions | [bubbletea-patterns.md](references/bubbletea-patterns.md) |
| Building a form (Huh or custom) | [huh-forms.md](references/huh-forms.md) |
| Rendering markdown content | [glamour-markdown.md](references/glamour-markdown.md) |
| Using viewport, spinner, table, etc. | [bubbles-components.md](references/bubbles-components.md) |
| Styling, theming, responsive layout | [lipgloss-theme-system.md](references/lipgloss-theme-system.md) |
| Building custom animated rendering | [ultraviolet-primitives.md](references/ultraviolet-primitives.md) |

### Review Questions

1. Is Update() free of blocking I/O? (NOT: "is the cmd helper blocking?")
2. Is the model immutable in Update?
3. Are Lipgloss styles defined once, not in View?
4. Is WindowSizeMsg handled for resizing?
5. Are key bindings documented with help.KeyMap?
6. Are Bubbles components sized correctly?
