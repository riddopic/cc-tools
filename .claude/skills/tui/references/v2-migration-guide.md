# Charmbracelet v1 to v2 Migration Guide

Reference for migrating TUI code from Charmbracelet v1 libraries to v2 (charm.land) imports and APIs.

## Import Mapping

When upgrading dependencies, update all imports according to this table:

| v1 Import | v2 Import |
|-----------|-----------|
| `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2` |
| `github.com/charmbracelet/lipgloss/table` | `charm.land/lipgloss/v2/table` |
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` |
| `github.com/charmbracelet/bubbles/viewport` | `charm.land/bubbles/v2/viewport` |
| `github.com/charmbracelet/bubbles/textinput` | `charm.land/bubbles/v2/textinput` |
| `github.com/charmbracelet/bubbles/list` | `charm.land/bubbles/v2/list` |
| `github.com/charmbracelet/glamour` | `charm.land/glamour/v2` |
| `github.com/charmbracelet/huh` | `charm.land/huh/v2` |

## Color Type Changes

### Color system migration

Colors in v2 use the standard library `color.Color` interface instead of `lipgloss.TerminalColor`:

```go
// v1
style := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF5F87")).
    Background(lipgloss.Color("240"))

// v2
import "image/color"
style := lipgloss.NewStyle().
    Foreground(color.RGBA{R: 255, G: 95, B: 135, A: 255}).
    Background(color.Gray16{Y: 240})
```

### AdaptiveColor moved to compat

`lipgloss.AdaptiveColor` is now in the compat package. Import from `charm.land/lipgloss/v2/compat`:

```go
import (
    "charm.land/lipgloss/v2"
    "charm.land/lipgloss/v2/compat"
)

style := lipgloss.NewStyle().
    Foreground(compat.AdaptiveColor{
        Light:  "#000000",
        Dark:   "#FFFFFF",
    })
```

## Output Changes

### Use lipgloss.Println for color downsampling

v2 requires explicit downsampling for terminal color support. Replace `fmt.Println(style.Render(...))`:

```go
// v1
fmt.Println(style.Render("Hello"))

// v2 - explicit downsampling
import "charm.land/lipgloss/v2"
lipgloss.Println(style.Render("Hello"))
```

### Use lipgloss.Fprint for writers

For output to custom writers, use `lipgloss.Fprint` instead of `fmt.Fprint`:

```go
// v2
lipgloss.Fprint(os.Stderr, style.Render("Error"))
```

## BubbleTea v2 Changes

### KeyMsg → KeyPressMsg

Key events renamed from `tea.KeyMsg` to `tea.KeyPressMsg`:

```go
// v1
case tea.KeyMsg:
    if msg.Type == tea.KeyEnter {
        // handle enter
    }

// v2
case tea.KeyPressMsg:
    if msg.String() == "enter" {
        // handle enter
    }
```

### Key matching via String()

Match keys by their string representation instead of `KeyType`:

```go
// v2 key matching
switch msg.String() {
case "enter":
    return m, handleEnter()
case "space":
    return m, handleSpace()
case "esc":
    return m, handleEscape()
case "tab":
    return m, handleTab()
case "up":
    return m, m.scrollUp()
case "down":
    return m, m.scrollDown()
}
```

### Alt screen mode

Alt screen is now configured via the `View()` return value:

```go
// v1
tea.WithAltScreen()

// v2
func (m Model) View() string {
    return tea.View{
        Content:   "Your UI content",
        AltScreen: true,
    }.Render()
}
```

### Viewport initialization

Viewport no longer takes width/height in constructor. Use setters instead:

```go
// v1
vp := viewport.New(width, height)

// v2
vp := viewport.New()
vp.SetWidth(width)
vp.SetHeight(height)

// Or chain them:
vp := viewport.New()
vp.Width = width
vp.Height = height
```

## ANSI-aware Truncation

When truncating styled strings, use the ANSI-aware truncator from `charm.land/x/ansi`:

```go
import ansi "github.com/charmbracelet/x/ansi"

// ❌ Don't: byte-based truncation removes ANSI codes
truncated := styledText[:maxLen]

// ✅ Do: ANSI-aware truncation preserves styling
truncated := ansi.Truncate(styledText, maxLen, "...")
```

This is critical for styled output—simple string slicing corrupts color codes and formatting.

## Charmtone Palette

For semantic color naming, use Charmtone from `charm.land/x/exp/charmtone`:

```go
import "charm.land/x/exp/charmtone"

palette := charmtone.Default()

// Primary accent colors
primaryColor := palette.Primary.Light()

// Backgrounds and foregrounds
bgColor := palette.Backgrounds.Default()
fgColor := palette.Foregrounds.Default()

// Semantic colors
errorColor := palette.Semantic.Error()
warningColor := palette.Semantic.Warning()
infoColor := palette.Semantic.Info()
successColor := palette.Semantic.Success()
```

Use semantic colors for consistent error/warning/success messaging throughout the TUI.

## Known v2 Issues

### Table width calculation (issue #574)

`lipgloss.Table` doesn't account for column borders in width calculations:

```go
// Table may overflow expected width due to border rendering
// Workaround: add padding or reduce content width
table := lipgloss.NewTable().
    WithRows(rows).
    WithPaddingRight(2)  // compensate for borders
```

### Styling edge cases (issue #520)

Some edge cases exist with style composition and color merging. Test styling thoroughly when upgrading.

## Migration Checklist

- [ ] Update all imports to charm.land v2
- [ ] Replace AdaptiveColor imports with compat package
- [ ] Replace fmt.Println with lipgloss.Println
- [ ] Update KeyMsg to KeyPressMsg with String() matching
- [ ] Update viewport initialization to use setters
- [ ] Review all string truncation to use ansi.Truncate
- [ ] Test color rendering in light/dark terminals
- [ ] Run full test suite with race detector
