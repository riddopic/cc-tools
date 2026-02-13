# Ultraviolet Primitives

Low-level terminal rendering below the Bubble Tea abstraction layer.

## When to Use vs Bubble Tea

| Use Case | Tool |
|----------|------|
| Standard views, forms, lists | Bubble Tea + Bubbles |
| Custom animated components | `charmbracelet/x/cellbuf` + Bubble Tea |
| ANSI string manipulation | `charmbracelet/x/ansi` |
| Particle systems, splash screens | Grid-based rendering (Quanta pattern) |
| Full terminal control (rare) | Ultraviolet (emerging) |

**Default to Bubble Tea.** Drop to lower levels only when you need:
- Per-cell character placement
- Physics-based animations
- Custom drawing that doesn't fit Model/Update/View

## charmbracelet/x/ansi — ANSI Utilities

String width, truncation, and escape sequence handling:

```go
import "github.com/charmbracelet/x/ansi"

// Correct string width (handles CJK, emoji, escape sequences)
w := ansi.StringWidth("Hello 世界")  // 11

// Truncate with awareness of ANSI sequences
truncated := ansi.Truncate("bold text", 8, "...")

// Strip ANSI escape sequences
plain := ansi.Strip("\x1b[1mBold\x1b[0m")  // "Bold"
```

## charmbracelet/x/cellbuf — Cell-Based Buffers

For building custom renderers that need per-cell control:

```go
import "github.com/charmbracelet/x/cellbuf"

// Create a screen buffer
buf := cellbuf.NewBuffer(width, height)

// Set individual cells
buf.SetCell(x, y, cellbuf.Cell{
    Rune:  '█',
    Style: cellbuf.Style{Foreground: cellbuf.Color(196)},
})

// Render to string
output := buf.Render()
```

## Neural Animation Pattern

Quanta's splash screen (`internal/tui/splash/neural.go`) demonstrates the grid-based animation pattern within Bubble Tea:

### Architecture

```
NeuralModel (tea.Model)
  ├─ particles []Particle    // animated entities
  ├─ phase NeuralPhase       // Drifting → Converging → Locked
  ├─ canvasW, canvasH int    // fixed canvas dimensions
  └─ startTime time.Time     // for phase transitions
```

### Grid rendering

The model uses a 2D string grid instead of cellbuf for simplicity:

```go
func makeGrid(w, h int) [][]string {
    grid := make([][]string, h)
    for y := range h {
        grid[y] = make([]string, w)
        for x := range w {
            grid[y][x] = " "
        }
    }
    return grid
}

func (m NeuralModel) View() string {
    grid := makeGrid(m.canvasW, m.canvasH)

    // Draw connections between close particles
    if m.phase == PhaseDrifting {
        drawConnections(grid, m.particles, m.canvasW, m.canvasH)
    }

    // Draw particles on top
    drawParticles(grid, m.particles, m.phase, m.canvasW, m.canvasH)

    // Flatten grid to string
    var sb strings.Builder
    for y := range m.canvasH {
        for x := range m.canvasW {
            sb.WriteString(grid[y][x])
        }
        sb.WriteRune('\n')
    }
    return sb.String()
}
```

### Physics simulation

Each particle has position, velocity, and a target:

```go
type Particle struct {
    x, y     float64       // current position
    vx, vy   float64       // velocity
    targetX  float64       // final position
    targetY  float64       // final position
    targetCh string        // character to display when locked
    color    lipgloss.Color
}
```

Phase-based physics:
- **Drifting**: `p.x += p.vx` with wall bouncing
- **Converging**: Spring force `p.vx += (target - p.x) * springForce` with damping
- **Locked**: No movement, display target character

### Tick-based animation

```go
const tickInterval = 50 * time.Millisecond  // ~20 FPS

func (m NeuralModel) Init() tea.Cmd {
    return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

func (m NeuralModel) Update(msg tea.Msg) (NeuralModel, tea.Cmd) {
    switch msg := msg.(type) {
    case TickMsg:
        // Update physics
        for i := range m.particles {
            updatePhysics(&m.particles[i])
        }

        // Check for phase transition
        if m.phase == PhaseConverging && m.allParticlesLocked() {
            m.phase = PhaseLocked
            return m, func() tea.Msg { return LockedMsg{} }
        }

        if m.phase == PhaseLocked {
            return m, nil  // stop ticking
        }

        // Re-subscribe for next tick
        return m, tea.Tick(tickInterval, func(t time.Time) tea.Msg {
            return TickMsg(t)
        })
    }
    return m, nil
}
```

### Styled cell placement

Apply lipgloss styles to individual grid cells:

```go
func drawParticles(grid [][]string, particles []Particle, phase NeuralPhase, w, h int) {
    for _, p := range particles {
        ix, iy := int(math.Round(p.x)), int(math.Round(p.y))
        if !inBounds(ix, iy, w, h) {
            continue
        }
        style := lipgloss.NewStyle().Foreground(p.color)
        char := selectChar(p, phase)
        grid[iy][ix] = style.Render(char)
    }
}
```

## Building Custom Animated Components

Recipe for a new animation within Bubble Tea:

1. **Define a model** with position data, phase enum, and tick interval
2. **Init** returns `tea.Tick` at your target FPS
3. **Update** handles `TickMsg`: update physics, check transitions, re-tick
4. **View** builds a grid, places styled characters, flattens to string
5. **Signal completion** with a custom message (e.g., `LockedMsg{}`)

Keep the canvas size fixed (don't recompute on every `WindowSizeMsg`) and center the output in the available space using `lipgloss.Place`.

## Ultraviolet Status

Ultraviolet is Charm's emerging low-level terminal framework. As of now:
- **Not stable** — API may change
- Use `charmbracelet/x/cellbuf` and `charmbracelet/x/ansi` for production needs
- Prototype with cellbuf, upgrade to Ultraviolet when it stabilizes
- Monitor `github.com/charmbracelet/ultraviolet` for releases
