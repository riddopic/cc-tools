# Bubbles Components

Build recipes for `charmbracelet/bubbles` components. Each recipe covers initialization, wiring into Update, and handling `WindowSizeMsg`.

## Integration Checklist

For every bubbles component:

- [ ] Initialize with styles from `*styles.Styles`
- [ ] Set dimensions on creation and in `SetSize`
- [ ] Wire `Update` — forward messages and collect `tea.Cmd`
- [ ] Handle `tea.WindowSizeMsg` to resize

## High-Priority Components

### viewport — Scrollable content area

```go
import "charm.land/bubbles/v2/viewport"

type LogView struct {
    vp     viewport.Model
    styles *styles.Styles
}

func NewLogView(s *styles.Styles) LogView {
    vp := viewport.New()
    vp.Style = s.Panel
    return LogView{vp: vp, styles: s}
}

func (lv *LogView) SetSize(w, h int) {
    lv.vp.Width = w
    lv.vp.Height = h
}

func (lv *LogView) SetContent(content string) {
    lv.vp.SetContent(content)
}

func (lv *LogView) Update(msg tea.Msg) (*LogView, tea.Cmd) {
    vp, cmd := lv.vp.Update(msg)
    lv.vp = vp
    return lv, cmd
}

func (lv *LogView) View() string { return lv.vp.View() }
```

Viewport handles scroll keys (up/down/pgup/pgdn) automatically.

### spinner — Activity indicator

```go
import "charm.land/bubbles/v2/spinner"

type LoadingView struct {
    spinner spinner.Model
    styles  *styles.Styles
}

func NewLoadingView(s *styles.Styles) LoadingView {
    sp := spinner.New()
    sp.Spinner = spinner.Dot          // or MiniDot, Line, Jump, Pulse, Points, Globe, Moon, Monkey, Meter
    sp.Style = lipgloss.NewStyle().Foreground(s.styles.StatusRunning.GetForeground())
    return LoadingView{spinner: sp, styles: s}
}

// Init must return the spinner's tick command.
func (lv LoadingView) Init() tea.Cmd {
    return lv.spinner.Tick
}

func (lv LoadingView) Update(msg tea.Msg) (LoadingView, tea.Cmd) {
    sp, cmd := lv.spinner.Update(msg)
    lv.spinner = sp
    return lv, cmd
}

func (lv LoadingView) View() string {
    return lv.spinner.View() + " Loading..."
}
```

### progress — Progress bar

```go
import "charm.land/bubbles/v2/progress"

type ProgressPanel struct {
    bar    progress.Model
    pct    float64
    width  int
    styles *styles.Styles
}

func NewProgressPanel(s *styles.Styles) ProgressPanel {
    bar := progress.New(
        progress.WithDefaultGradient(),
        progress.WithWidth(40),
    )
    return ProgressPanel{bar: bar, styles: s}
}

func (pp *ProgressPanel) SetPercent(pct float64) tea.Cmd {
    return pp.bar.SetPercent(pct)
}

func (pp *ProgressPanel) SetWidth(w int) {
    pp.width = w
    pp.bar.Width = w - 4 // padding
}

func (pp *ProgressPanel) Update(msg tea.Msg) (*ProgressPanel, tea.Cmd) {
    bar, cmd := pp.bar.Update(msg)
    pp.bar = bar.(progress.Model)
    return pp, cmd
}

func (pp *ProgressPanel) View() string {
    return pp.bar.View()
}
```

### table (bubbles) — Simple data table

```go
import "charm.land/bubbles/v2/table"

cols := []table.Column{
    {Title: "Contract", Width: 20},
    {Title: "Chain", Width: 10},
    {Title: "Status", Width: 12},
    {Title: "Profit", Width: 15},
}

rows := []table.Row{
    {"Uniswap V2", "ethereum", "completed", "$1,234.56"},
}

t := table.New(
    table.WithColumns(cols),
    table.WithRows(rows),
    table.WithFocused(true),
    table.WithHeight(20),
)

// Style the table
s := table.DefaultStyles()
s.Header = s.Header.
    BorderStyle(lipgloss.NormalBorder()).
    BorderBottom(true).
    Bold(true)
s.Selected = s.Selected.
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("57"))
t.SetStyles(s)
```

**When to use bubbles/table vs Quanta's `components.TableModel`**: Use Quanta's `TableModel` when you need live row updates by ID (`AddRow`/`UpdateRow` with O(1) lookup), status-based row styling, and integration with Quanta's styles system. Use `bubbles/table` for simpler static or read-only tables.

### list — Filterable item list

```go
import "charm.land/bubbles/v2/list"

items := []list.Item{
    listItem{title: "Uniswap V2", desc: "Ethereum DEX"},
    listItem{title: "PancakeSwap", desc: "BSC DEX"},
}

l := list.New(items, list.NewDefaultDelegate(), 0, 0)
l.Title = "Contracts"
l.SetShowStatusBar(true)
l.SetFilteringEnabled(true)
```

Implement `list.Item` interface:

```go
type listItem struct {
    title, desc string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }
```

### help — Key binding help

```go
import "charm.land/bubbles/v2/help"

h := help.New()
h.Width = width

// In View:
helpView := h.View(keyMap) // keyMap implements help.KeyMap
```

Quanta's `KeyMap` already implements `help.KeyMap` with `ShortHelp()` and `FullHelp()`.

### textinput — Single-line text input

```go
import "charm.land/bubbles/v2/textinput"

ti := textinput.New()
ti.Placeholder = "Enter contract address..."
ti.CharLimit = 42
ti.Width = 50
ti.Focus() // Start with focus

// In Update:
ti, cmd := ti.Update(msg)

// Get value:
address := ti.Value()
```

### textarea — Multi-line text input

```go
import "charm.land/bubbles/v2/textarea"

ta := textarea.New()
ta.Placeholder = "Paste Solidity source..."
ta.SetWidth(80)
ta.SetHeight(20)
ta.Focus()

// In Update:
ta, cmd := ta.Update(msg)
```

## Medium-Priority Components

### paginator — Page navigation

```go
import "charm.land/bubbles/v2/paginator"

p := paginator.New()
p.Type = paginator.Dots // or Arabic
p.PerPage = 10
p.TotalPages = (totalItems + 9) / 10

// Get current page slice:
start, end := p.GetSliceBounds(len(items))
pageItems := items[start:end]
```

### filepicker — File selection

```go
import "charm.land/bubbles/v2/filepicker"

fp := filepicker.New()
fp.AllowedTypes = []string{".sol", ".json"}
fp.CurrentDirectory = "."
```

### timer / stopwatch — Time tracking

```go
import "charm.land/bubbles/v2/timer"
import "charm.land/bubbles/v2/stopwatch"

// Countdown:
t := timer.NewWithInterval(5*time.Minute, time.Second)

// Elapsed:
sw := stopwatch.NewWithInterval(time.Second)
```

## Composition Patterns

### spinner + viewport (loading then content)

```go
func (m Model) View() string {
    if m.loading {
        return m.spinner.View() + " Fetching data..."
    }
    return m.viewport.View()
}
```

### textinput + list (search/filter)

```go
func (m Model) View() string {
    return lipgloss.JoinVertical(lipgloss.Left,
        m.searchInput.View(),
        m.filteredList.View(),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    ti, cmd := m.searchInput.Update(msg)
    m.searchInput = ti
    cmds = append(cmds, cmd)

    // Filter list based on input
    m.filteredList.SetFilterText(m.searchInput.Value())

    return m, tea.Batch(cmds...)
}
```

### progress + table (batch operation tracking)

```go
func (m Model) View() string {
    return lipgloss.JoinVertical(lipgloss.Left,
        m.progressBar.View(),
        fmt.Sprintf(" %d/%d completed", m.done, m.total),
        "",
        m.resultsTable.View(),
    )
}
```
