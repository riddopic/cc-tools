# Huh Forms

## Status

Huh (`charm.land/huh/v2`) is **not yet in go.mod**. Quanta currently uses a custom `components.FormModel` in `internal/tui/components/form.go`.

## Adoption Steps

```bash
go get charm.land/huh/v2@latest
```

Verify compatibility with the existing Bubble Tea version in `go.mod`.

## Huh vs Custom Form

| Feature | Custom `FormModel` | Huh |
|---------|-------------------|-----|
| Field types | Text, Int, Float, Bool, Select, MultiSelect, Duration | Input, Text, Select, MultiSelect, Confirm, Note, FilePicker |
| Validation | Manual per-field | Built-in per-field + cross-field |
| Grouping | Manual `FormGroup` | `huh.NewGroup()` with pages |
| Theming | Uses `*styles.Styles` | `huh.Theme` (map to Quanta theme) |
| Accessibility | Manual | Built-in accessible mode |
| Two-column layout | Custom `splitGroups` | Not built-in (wrap in lipgloss layout) |

**Migration strategy**: Keep the custom form for the regression config (38 fields, two-column layout, tight integration). Use Huh for new, simpler forms (settings, single-page wizards).

## Embedding Huh in Bubble Tea

**Critical**: Never call `form.Run()` inside a Bubble Tea program. That starts its own `tea.Program` and conflicts. Instead, embed via the Update loop:

```go
package views

import (
    "charm.land/huh/v2"
    tea "charm.land/bubbletea/v2"
)

type SettingsView struct {
    form   *huh.Form
    width  int
    height int
}

func NewSettingsView() *SettingsView {
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Theme").
                Options(
                    huh.NewOption("Default", "default"),
                    huh.NewOption("Dark", "dark"),
                    huh.NewOption("Light", "light"),
                ).
                Value(&theme),

            huh.NewInput().
                Title("API Key").
                Placeholder("sk-...").
                Value(&apiKey),

            huh.NewConfirm().
                Title("Enable Tor?").
                Value(&useTor),
        ),
    )

    return &SettingsView{form: form}
}

// Init delegates to the form's Init.
func (sv *SettingsView) Init() tea.Cmd {
    return sv.form.Init()
}

// Update delegates to the form's Update.
func (sv *SettingsView) Update(msg tea.Msg) (*SettingsView, tea.Cmd) {
    form, cmd := sv.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        sv.form = f
    }
    return sv, cmd
}

// View delegates to the form's View.
func (sv *SettingsView) View() string {
    if sv.form.State == huh.StateCompleted {
        return "Settings saved."
    }
    return sv.form.View()
}
```

## Field Types

### Input (single-line text)

```go
var name string
huh.NewInput().
    Title("Contract Name").
    Placeholder("e.g., Uniswap V2").
    Value(&name).
    Validate(func(s string) error {
        if s == "" {
            return fmt.Errorf("name is required")
        }
        return nil
    })
```

### Text (multi-line)

```go
var notes string
huh.NewText().
    Title("Notes").
    Placeholder("Additional context...").
    CharLimit(500).
    Value(&notes)
```

### Select

```go
var chain string
huh.NewSelect[string]().
    Title("Chain").
    Options(
        huh.NewOption("Ethereum", "ethereum"),
        huh.NewOption("BSC", "bsc"),
        huh.NewOption("Base", "base"),
    ).
    Value(&chain)
```

### MultiSelect

```go
var features []string
huh.NewMultiSelect[string]().
    Title("Features").
    Options(
        huh.NewOption("RAG", "rag"),
        huh.NewOption("Multi-Model", "multi-model"),
        huh.NewOption("Multi-Actor", "multi-actor"),
    ).
    Value(&features)
```

### Confirm

```go
var proceed bool
huh.NewConfirm().
    Title("Run regression?").
    Affirmative("Yes").
    Negative("No").
    Value(&proceed)
```

### Note (read-only text)

```go
huh.NewNote().
    Title("Warning").
    Description("This will consume API credits.")
```

## Validation

### Per-field

```go
huh.NewInput().
    Title("Workers").
    Value(&workers).
    Validate(func(s string) error {
        n, err := strconv.Atoi(s)
        if err != nil {
            return fmt.Errorf("must be a number")
        }
        if n < 1 || n > 32 {
            return fmt.Errorf("must be between 1 and 32")
        }
        return nil
    })
```

### Cross-field (group-level)

```go
huh.NewGroup(fields...).
    WithValidate(func() error {
        if useTor && apiKey == "" {
            return fmt.Errorf("API key required when Tor is enabled")
        }
        return nil
    })
```

## Multi-Page Forms (Groups)

Each `huh.NewGroup()` becomes a page. Users navigate between pages:

```go
form := huh.NewForm(
    huh.NewGroup(basicFields...).Title("Basic Settings"),
    huh.NewGroup(advancedFields...).Title("Advanced Settings"),
    huh.NewGroup(confirmFields...).Title("Confirm"),
)
```

## Theming to Match Quanta

Map Quanta's `styles.Theme` to a Huh theme:

```go
func quantaHuhTheme(theme styles.Theme) *huh.Theme {
    t := huh.ThemeBase()

    t.Focused.Title = t.Focused.Title.Foreground(theme.Primary)
    t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(theme.Success)
    t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(theme.Text)
    t.Focused.Base = t.Focused.Base.BorderForeground(theme.Border)

    t.Blurred.Title = t.Blurred.Title.Foreground(theme.TextDim)

    return t
}

// Apply:
form := huh.NewForm(groups...).WithTheme(quantaHuhTheme(styles.GetTheme()))
```

## Accessible Mode

For screen readers and reduced-motion environments:

```go
form := huh.NewForm(groups...).WithAccessible(true)
```

This disables animations and uses simpler rendering.

## Value Extraction

Huh writes directly to the pointer variables passed to `.Value()`:

```go
var chain string
var workers int

// After form completes:
if form.State == huh.StateCompleted {
    fmt.Println(chain)   // value is already set
    fmt.Println(workers) // value is already set
}
```

For dynamic access, use `form.GetString("key")` if keys are assigned via `.Key("chain")`.
