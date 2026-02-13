# Glamour Markdown Rendering

## Current Usage

Quanta uses Glamour in `internal/output/render.go` via `TerminalRenderer`:

```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),       // auto-detect light/dark
    glamour.WithWordWrap(wordWrap), // wrap to terminal width
)
rendered, err := renderer.Render(content)
```

Key design decisions in the existing code:
- Creates a **new renderer per operation** for thread safety (no shared mutable state)
- Graceful degradation: returns raw content if rendering fails
- Supports `NO_COLOR` and `TERM=dumb` environments

## Rendering Markdown in Bubble Tea

### Create renderer once, not in View()

Glamour rendering is expensive. Create the renderer in Init or when content changes, store the result:

```go
type DocView struct {
    rawContent      string
    renderedContent string
    width           int
    styles          *styles.Styles
}

func (dv *DocView) SetContent(md string) tea.Cmd {
    return func() tea.Msg {
        renderer, err := glamour.NewTermRenderer(
            glamour.WithAutoStyle(),
            glamour.WithWordWrap(dv.width),
        )
        if err != nil {
            return markdownRenderedMsg{content: md} // fallback to raw
        }
        rendered, err := renderer.Render(md)
        if err != nil {
            return markdownRenderedMsg{content: md}
        }
        return markdownRenderedMsg{content: rendered}
    }
}

type markdownRenderedMsg struct {
    content string
}

func (dv *DocView) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case markdownRenderedMsg:
        dv.renderedContent = msg.content
    }
    return nil
}

func (dv *DocView) View() string {
    return dv.renderedContent // pre-rendered, no work here
}
```

### Custom styles mapped to Quanta theme

```go
func quantaGlamourStyle(theme styles.Theme) glamour.TermRendererOption {
    // Start with auto style and customize
    return glamour.WithStylesOnDark(ansi.StyleConfig{
        Heading: ansi.StyleBlock{
            StylePrimitive: ansi.StylePrimitive{
                Color: ptrString(theme.Primary.Dark),
                Bold:  ptrBool(true),
            },
        },
        Code: ansi.StyleBlock{
            StylePrimitive: ansi.StylePrimitive{
                Color: ptrString(theme.Secondary.Dark),
            },
        },
        Link: ansi.StylePrimitive{
            Color: ptrString(theme.Accent2.Dark),
        },
    })
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }
```

## Recipe: Markdown Viewport Component

Combine Glamour with `bubbles/viewport` for scrollable markdown:

```go
package components

import (
    "charm.land/bubbles/v2/viewport"
    "charm.land/glamour/v2"
    tea "charm.land/bubbletea/v2"
    "github.com/riddopic/quanta/internal/tui/styles"
)

type MarkdownViewport struct {
    viewport viewport.Model
    raw      string
    width    int
    height   int
    styles   *styles.Styles
}

func NewMarkdownViewport(s *styles.Styles) MarkdownViewport {
    vp := viewport.New()
    vp.Style = s.Panel
    return MarkdownViewport{
        viewport: vp,
        styles:   s,
    }
}

// SetContent renders markdown and loads into viewport.
// Call as a tea.Cmd — rendering is expensive.
func (mv *MarkdownViewport) SetContent(md string) tea.Cmd {
    return func() tea.Msg {
        renderer, err := glamour.NewTermRenderer(
            glamour.WithAutoStyle(),
            glamour.WithWordWrap(mv.width - 4), // account for viewport padding
        )
        if err != nil {
            return mdViewportContentMsg{content: md}
        }
        rendered, err := renderer.Render(md)
        if err != nil {
            return mdViewportContentMsg{content: md}
        }
        return mdViewportContentMsg{content: rendered}
    }
}

type mdViewportContentMsg struct {
    content string
}

func (mv *MarkdownViewport) Update(msg tea.Msg) (*MarkdownViewport, tea.Cmd) {
    switch msg := msg.(type) {
    case mdViewportContentMsg:
        mv.viewport.SetContent(msg.content)
        return mv, nil
    }

    vp, cmd := mv.viewport.Update(msg)
    mv.viewport = vp
    return mv, cmd
}

func (mv *MarkdownViewport) SetSize(width, height int) {
    mv.width = width
    mv.height = height
    mv.viewport.Width = width
    mv.viewport.Height = height
}

func (mv *MarkdownViewport) View() string {
    return mv.viewport.View()
}
```

## Streaming Markdown for LLM Output

For the Analyze view, LLM output arrives in chunks. Re-render incrementally:

```go
type AnalyzeOutput struct {
    buffer   strings.Builder
    rendered string
    width    int
    dirty    bool
}

func (ao *AnalyzeOutput) AppendChunk(chunk string) {
    ao.buffer.WriteString(chunk)
    ao.dirty = true
}

// RerenderCmd re-renders the accumulated buffer.
// Call periodically (e.g., on tick) when dirty, not on every chunk.
func (ao *AnalyzeOutput) RerenderCmd() tea.Cmd {
    if !ao.dirty {
        return nil
    }
    ao.dirty = false
    content := ao.buffer.String()
    width := ao.width

    return func() tea.Msg {
        renderer, err := glamour.NewTermRenderer(
            glamour.WithAutoStyle(),
            glamour.WithWordWrap(width),
        )
        if err != nil {
            return analyzeRenderMsg{content: content}
        }
        rendered, _ := renderer.Render(content)
        return analyzeRenderMsg{content: rendered}
    }
}
```

Throttle re-renders to avoid performance issues — render at most once per tick interval, not per chunk.

## Width-Responsive Re-rendering

When the terminal resizes, markdown must be re-rendered with the new width:

```go
case tea.WindowSizeMsg:
    mv.SetSize(msg.Width, msg.Height)
    // Re-render with new width
    return mv, mv.SetContent(mv.raw)
```

Store the raw markdown so it can be re-rendered at any width.
