# Instinct System Migration: continuous-learning-v2 → cc-tools

**Date:** 2026-02-24
**Status:** Approved
**Approach:** Full Go rewrite (Approach A)

## Problem

The `continuous-learning-v2` skill implements an instinct-based learning system
using Python (`instinct-cli.py`) and Bash (`observe.sh`) scripts. These scripts
duplicate functionality already present in the cc-tools Go binary:

- `observe.sh` captures tool usage to JSONL — `ObserveHandler` in Go already
  does this on every PreToolUse, PostToolUse, and PostToolUseFailure event.
- `config.json` manages observation settings — `config.Values` already manages
  these under `observe.*` keys.

This creates two observation systems running in parallel, a Python dependency in
an otherwise pure-Go tool chain, and configuration split across two files.

## Solution

Rewrite the instinct management system in Go as `cc-tools instinct` subcommands.
Remove the redundant Python/Bash observation hooks. Consolidate config. Move
instinct storage to `~/.config/cc-tools/instincts/` (user-global). Drop the
experimental background observer process.

## Section 1: Observation Layer Changes

### Remove duplicate hooks

Remove from `.claude/settings.json`:

```json
// PreToolUse — REMOVE:
{ "type": "command", "command": ".claude/skills/continuous-learning-v2/hooks/observe.sh pre" }

// PostToolUse — REMOVE:
{ "type": "command", "command": ".claude/skills/continuous-learning-v2/hooks/observe.sh post" }
```

The `cc-tools hook` entries already run `ObserveHandler` on these events.

### Extend observe.Event

Add `ToolOutput` and `Error` fields to capture richer data for instinct analysis:

```go
type Event struct {
    Timestamp  time.Time       `json:"timestamp"`
    Phase      string          `json:"phase"`      // "pre", "post", "failure"
    ToolName   string          `json:"tool_name"`
    ToolInput  json.RawMessage `json:"tool_input,omitempty"`
    ToolOutput json.RawMessage `json:"tool_output,omitempty"` // NEW
    Error      string          `json:"error,omitempty"`       // NEW (failure phase)
    SessionID  string          `json:"session_id"`
}
```

Update `ObserveHandler` to pass `input.ToolOutput` for post events and
`input.Error` for failure events.

## Section 2: internal/instinct Package

### Package structure

```
internal/instinct/
├── instinct.go       # Core types: Instinct, Domain, Source
├── store.go          # File-based storage: Load, Save, List, Delete
├── parser.go         # YAML frontmatter parser
├── confidence.go     # Confidence scoring, decay calculations
├── cluster.go        # Clustering algorithm for evolve
├── export.go         # Export to YAML/JSON formats
├── import.go         # Import with dedup/merge logic
└── mocks/            # Generated mockery mocks
```

### Core types

```go
type Instinct struct {
    ID         string    `json:"id" yaml:"id"`
    Trigger    string    `json:"trigger" yaml:"trigger"`
    Confidence float64   `json:"confidence" yaml:"confidence"`
    Domain     string    `json:"domain" yaml:"domain"`
    Source     string    `json:"source" yaml:"source"`
    SourceRepo string    `json:"source_repo,omitempty" yaml:"source_repo,omitempty"`
    Content    string    `json:"content,omitempty" yaml:"-"`
    CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" yaml:"updated_at"`
}

type Store interface {
    List(opts ListOptions) ([]Instinct, error)
    Get(id string) (*Instinct, error)
    Save(inst Instinct) error
    Delete(id string) error
}
```

### Storage layout

```
~/.config/cc-tools/instincts/
├── personal/         # Auto-learned and manually created
│   ├── prefer-functional-style.yaml
│   └── use-table-driven-tests.yaml
└── inherited/        # Imported from others
    └── team-go-conventions.yaml
```

### File format

Same YAML frontmatter format as Python system for backward compatibility:

```yaml
---
id: prefer-functional-style
trigger: "when writing new functions"
confidence: 0.75
domain: "code-style"
source: "session-observation"
created_at: "2025-01-22T10:30:00Z"
updated_at: "2025-06-15T14:00:00Z"
---

## Action
Use functional patterns over classes when appropriate.

## Evidence
- Observed 5 instances of functional pattern preference
```

### Confidence scoring

| Observations | Confidence |
|---|---|
| 1-2 | 0.3 (tentative) |
| 3-5 | 0.5 (moderate) |
| 6-10 | 0.7 (strong) |
| 11+ | 0.85 (very strong) |

Adjustments:
- +0.05 per confirming observation
- -0.10 per contradicting observation
- -0.02 per week without observation (decay)
- Clamped to [0.3, 0.9] range

### Clustering for evolve

Groups instincts by normalized trigger text (strip stop words like "when",
"creating", "writing"). Requires 3+ instincts per cluster. Produces candidates:

- **Skills:** Clusters of 3+ related instincts in the same domain
- **Commands:** High-confidence (≥0.7) workflow instincts
- **Agents:** Large clusters (3+) with ≥0.75 average confidence

## Section 3: cc-tools instinct Commands

Register in `cmd/cc-tools/main.go` alongside existing commands:

```go
root.AddCommand(instinctCmd())
```

### Subcommands

**`cc-tools instinct status`**
- Lists instincts from personal/ and inherited/ directories
- Groups by domain with confidence bar chart (█░ format)
- Flags: `--domain`, `--min-confidence`

**`cc-tools instinct export`**
- Writes instincts to stdout or file
- Flags: `--output`, `--domain`, `--min-confidence`, `--format yaml|json`

**`cc-tools instinct import <source>`**
- Reads from file path or URL
- Deduplicates by ID, updates if new confidence > old
- Writes to inherited/ directory
- Flags: `--dry-run`, `--force`, `--min-confidence`

**`cc-tools instinct evolve`**
- Analyzes instinct clusters
- Reports candidates for skills/commands/agents
- Requires 3+ instincts
- Flag: `--generate` (future: auto-generate evolved structures)

## Section 4: Configuration

Add to `internal/config/values.go`:

```go
type InstinctValues struct {
    PersonalPath     string  `json:"personal_path"`
    InheritedPath    string  `json:"inherited_path"`
    MinConfidence    float64 `json:"min_confidence"`
    AutoApprove      float64 `json:"auto_approve"`
    DecayRate        float64 `json:"decay_rate"`
    MaxInstincts     int     `json:"max_instincts"`
    ClusterThreshold int     `json:"cluster_threshold"`
}
```

| Key | Type | Default |
|---|---|---|
| `instinct.personal_path` | string | `~/.config/cc-tools/instincts/personal` |
| `instinct.inherited_path` | string | `~/.config/cc-tools/instincts/inherited` |
| `instinct.min_confidence` | float64 | 0.3 |
| `instinct.auto_approve` | float64 | 0.7 |
| `instinct.decay_rate` | float64 | 0.02 |
| `instinct.max_instincts` | int | 100 |
| `instinct.cluster_threshold` | int | 3 |

Eliminates the separate `continuous-learning-v2/config.json`.

## Section 5: File Changes

### settings.json

Remove the two `observe.sh` hook entries from PreToolUse and PostToolUse arrays.

### Files to remove

| File | Reason |
|---|---|
| `.claude/skills/continuous-learning-v2/hooks/observe.sh` | Replaced by ObserveHandler |
| `.claude/skills/continuous-learning-v2/scripts/instinct-cli.py` | Replaced by cc-tools instinct |
| `.claude/skills/continuous-learning-v2/scripts/test_parse_instinct.py` | Tests move to Go |
| `.claude/skills/continuous-learning-v2/config.json` | Merged into cc-tools config |
| `.claude/skills/continuous-learning-v2/agents/start-observer.sh` | Observer dropped |
| `.claude/skills/continuous-learning-v2/agents/observer.md` | Observer dropped |

### Files to update

| File | Change |
|---|---|
| `.claude/skills/continuous-learning-v2/SKILL.md` | Rewrite to reference cc-tools commands and new paths |
| `.claude/commands/instinct-status.md` | Call `cc-tools instinct status` |
| `.claude/commands/instinct-export.md` | Call `cc-tools instinct export` |
| `.claude/commands/instinct-import.md` | Call `cc-tools instinct import` |
| `.claude/commands/evolve.md` | Call `cc-tools instinct evolve` |
| `.claude/commands/learn.md` | Update paths to `~/.config/cc-tools/instincts/personal/` |
| `.claude/commands/learn-eval.md` | Update paths to `~/.config/cc-tools/instincts/personal/` |
| `.claude/rules/hooks.md` | Remove observe.sh references |
| `.claude/rules/agents.md` | Update continuous-learning-v2 description |

## Section 6: Data Migration

Optional `cc-tools instinct migrate` subcommand:

1. Scans `.claude/homunculus/instincts/personal/` and `inherited/`
2. Parses each YAML frontmatter file
3. Copies to `~/.config/cc-tools/instincts/personal/` and `inherited/`
4. Does NOT delete originals

Not critical for MVP — manual copy works for small instinct collections.

## Testing Strategy

| Package | Test Focus |
|---|---|
| `internal/instinct/parser_test.go` | YAML frontmatter parsing, malformed input, multi-instinct files |
| `internal/instinct/store_test.go` | List, Get, Save, Delete on temp directories |
| `internal/instinct/confidence_test.go` | Scoring, decay, range clamping |
| `internal/instinct/cluster_test.go` | Trigger normalization, grouping, threshold |
| `internal/instinct/export_test.go` | YAML and JSON output formats |
| `internal/instinct/import_test.go` | Dedup, merge, URL fetch, dry-run |
| `internal/observe/observe_test.go` | Extended Event with ToolOutput/Error |
| `cmd/cc-tools/instinct_test.go` | Cobra command integration |

All table-driven tests, testify assertions, `-tags=testmode`, mockery for
interfaces.

## Implementation Phases

1. **Observation cleanup:** Remove observe.sh, extend Event struct, update
   settings.json
2. **Core package:** `internal/instinct/` types, parser, store, confidence
3. **CLI commands:** `cc-tools instinct status|export|import|evolve`
4. **Config integration:** Add instinct keys to config manager
5. **Command/skill updates:** Rewire .claude/commands/ and SKILL.md
6. **File cleanup:** Remove Python/Bash scripts, update rules and agents

## Decisions

- **Background observer dropped:** Manual `/learn` and `/learn-eval` provide
  equivalent functionality with more user control and lower complexity.
- **User-global storage:** Instincts at `~/.config/cc-tools/instincts/` persist
  across projects, matching cc-tools conventions.
- **YAML frontmatter preserved:** Same file format as Python system for backward
  compatibility and human readability.
- **Confidence clamped to [0.3, 0.9]:** Prevents instincts from reaching 0
  (useless) or 1.0 (overconfident).
