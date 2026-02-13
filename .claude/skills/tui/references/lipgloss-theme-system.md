# Lipgloss Theme System

## Quanta's Theme/Styles Architecture

### Theme — Color palette

Defined in `internal/tui/styles/theme.go`:

```go
type Theme struct {
    Primary    lipgloss.AdaptiveColor  // Cyan brand
    Secondary  lipgloss.AdaptiveColor  // Purple brand
    Success    lipgloss.AdaptiveColor  // Green (profit, pass)
    Warning    lipgloss.AdaptiveColor  // Yellow (running)
    Error      lipgloss.AdaptiveColor  // Red (fail)
    Muted      lipgloss.AdaptiveColor  // Gray (dimmed)
    Background lipgloss.AdaptiveColor  // Panel backgrounds
    Border     lipgloss.AdaptiveColor  // Border color
    Text       lipgloss.AdaptiveColor  // Primary text
    TextDim    lipgloss.AdaptiveColor  // Secondary text
    Accent1    lipgloss.AdaptiveColor  // Gold (profit values)
    Accent2    lipgloss.AdaptiveColor  // Blue (info)
}
```

`AdaptiveColor` provides separate values for light and dark terminals:

```go
Primary: compat.AdaptiveColor{Light: lipgloss.Color("27"), Dark: lipgloss.Color("86")},
```

### Styles — Pre-built lipgloss.Style collection

Defined in `internal/tui/styles/styles.go`:

```go
type Styles struct {
    Title, Subtitle, Header             lipgloss.Style
    Selected, Cursor                    lipgloss.Style
    StatusPass, StatusFail, StatusRunning, StatusSkip lipgloss.Style
    Profit, Cost                        lipgloss.Style
    Border, Separator                   lipgloss.Style
    Muted, Help, Label, Value           lipgloss.Style
    Tier1, Tier2, Tier3                 lipgloss.Style
    Panel, PanelHeader                  lipgloss.Style
}
```

### Singleton access

```go
// Thread-safe global access (sync.Once)
theme := styles.GetTheme()
s := styles.GetStyles()

// Or inject via constructor (preferred for testability)
container := views.NewRegressionContainer(&s)
```

## Recipe: Adding New Styles

1. Add the style field to `Styles` struct:

```go
type Styles struct {
    // ... existing fields ...
    ProgressBar lipgloss.Style
    Badge       lipgloss.Style
}
```

2. Initialize in `NewStyles`:

```go
func NewStyles(theme Theme) Styles {
    return Styles{
        // ... existing styles ...
        ProgressBar: lipgloss.NewStyle().
            Foreground(theme.Primary),
        Badge: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Text).
            Background(theme.Secondary).
            Padding(0, 1),
    }
}
```

3. Use via injected `*styles.Styles`:

```go
func (m *MyComponent) View() string {
    return m.styles.Badge.Render("NEW")
}
```

## Layout

### JoinVertical — Stack elements top to bottom

```go
lipgloss.JoinVertical(lipgloss.Left,
    header,
    content,
    statusBar,
)
```

Alignment options: `lipgloss.Left`, `lipgloss.Center`, `lipgloss.Right`.

### JoinHorizontal — Place elements side by side

```go
lipgloss.JoinHorizontal(lipgloss.Top,
    leftPanel,
    rightPanel,
)
```

Alignment options: `lipgloss.Top`, `lipgloss.Center`, `lipgloss.Bottom`.

### Place — Position content within a box

```go
// Center text in a fixed-width area
lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center,
    styles.Title.Render("REGRESSION CONFIGURATION"),
)
```

Used in `RegressionConfig.View()` for centered headers.

### Two-column layout (Quanta pattern)

```go
colWidth := width / 2
leftCol := renderColumn(leftGroups, colWidth)
rightCol := renderColumn(rightGroups, colWidth)
columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
```

## Responsive Design

### WindowSizeMsg propagation

```go
// App propagates to all view handlers
func (a *App) updateLayout() {
    a.header.SetWidth(a.width)
    a.statusBar.SetWidth(a.width)
    contentHeight := a.height - appContentOffset
    for _, handler := range a.viewHandlers {
        handler.SetSize(a.width, contentHeight)
    }
}

// Container propagates to sub-views
func (rc *RegressionContainer) SetSize(width, height int) {
    rc.width = width
    rc.height = height
    rc.config.SetSize(width, height)
    rc.results.SetSize(width, height)
    rc.profit.SetWidth(width)
}
```

### Breakpoints

Define layout constants and adapt:

```go
const (
    narrowWidth = 80
    wideWidth   = 120
)

func (m *MyView) View() string {
    if m.width < narrowWidth {
        // Single column, compact
        return lipgloss.JoinVertical(lipgloss.Left, m.panel1.View(), m.panel2.View())
    }
    // Side by side
    return lipgloss.JoinHorizontal(lipgloss.Top, m.panel1.View(), m.panel2.View())
}
```

## Borders and Panels

### Border types

```go
lipgloss.NormalBorder()    // ┌─┐│ │└─┘
lipgloss.RoundedBorder()   // ╭─╮│ │╰─╯  (used by Quanta's Panel)
lipgloss.ThickBorder()     // ┏━┓┃ ┃┗━┛
lipgloss.DoubleBorder()    // ╔═╗║ ║╚═╝
lipgloss.HiddenBorder()    // Invisible (useful for alignment)
```

### Frame size calculation

Borders consume space. Account for it:

```go
panel := styles.Panel  // has RoundedBorder + Padding(0, 1)

// Get the frame dimensions (border + padding + margin)
frameW := panel.GetHorizontalFrameSize()  // left+right border + padding
frameH := panel.GetVerticalFrameSize()    // top+bottom border + padding

// Set inner content width
innerWidth := totalWidth - frameW
panel = panel.Width(innerWidth)
```

### Separator lines

```go
sep := styles.Border.Render(strings.Repeat("─", width))
doubleSep := styles.Border.Render(strings.Repeat("═", width))
```

## Color System

### AdaptiveColor — light/dark terminal support

```go
// Automatically uses Light or Dark based on terminal
color := compat.AdaptiveColor{Light: lipgloss.Color("27"), Dark: lipgloss.Color("86")}
style := lipgloss.NewStyle().Foreground(color)
```

### Color — single ANSI color

```go
color := lipgloss.Color("205")       // ANSI 256
color := lipgloss.Color("#FF00FF")   // Hex
```

### CompleteColor — full color specification

```go
color := lipgloss.CompleteColor{
    TrueColor: "#FF00FF",
    ANSI256:   "205",
    ANSI:      "5",
}
```

### Never hardcode colors in components

```go
// BAD
style := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

// GOOD — use theme
style := lipgloss.NewStyle().Foreground(theme.Error)

// BEST — use pre-built style
rendered := m.styles.StatusFail.Render("FAILED")
```

## Text Handling

### Width measurement

```go
import "charm.land/lipgloss/v2"

w := lipgloss.Width("Hello, 世界")  // Correct width including CJK
```

### Truncation

```go
import "github.com/muesli/reflow/truncate"

truncated := truncate.StringWithTail(longText, uint(maxWidth), "...")
```

### String builder for complex views

```go
var sb strings.Builder
sb.WriteString(m.styles.Title.Render("Header"))
sb.WriteRune('\n')
sb.WriteString(m.styles.Border.Render(strings.Repeat("─", m.width)))
sb.WriteRune('\n')
sb.WriteString(content)
return sb.String()
```

Use `strings.Builder` for assembling multi-line content within a single section. Use `lipgloss.JoinVertical` for combining distinct visual sections.
