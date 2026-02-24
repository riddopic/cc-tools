---
name: continuous-learning-v2
description: Instinct-based learning system that observes sessions via hooks, creates atomic instincts with confidence scoring, and evolves them into skills/commands/agents.
version: 2.0.0
---

# Continuous Learning v2 - Instinct-Based Architecture

An advanced learning system that turns your Claude Code sessions into reusable knowledge through atomic "instincts" - small learned behaviors with confidence scoring.

## When to Activate

- Setting up automatic learning from Claude Code sessions
- Reviewing, exporting, or importing instinct libraries
- Tuning confidence thresholds for learned behaviors
- Evolving instincts into full skills, commands, or agents

## The Instinct Model

An instinct is a small learned behavior stored as a YAML frontmatter file:

```yaml
---
id: prefer-functional-style
trigger: "when writing new functions"
confidence: 0.7
domain: "code-style"
source: "session-observation"
---

# Prefer Functional Style

## Action
Use functional patterns over classes when appropriate.

## Evidence
- Observed 5 instances of functional pattern preference
- User corrected class-based approach to functional on 2025-01-15
```

**Properties:**
- **Atomic** -- one trigger, one action
- **Confidence-weighted** -- 0.3 = tentative, 0.9 = near certain
- **Domain-tagged** -- code-style, testing, git, debugging, workflow, etc.
- **Evidence-backed** -- tracks what observations created it

## Quick Start

### 1. Observation

Observation is handled automatically by `cc-tools hook`. The ObserveHandler captures tool usage events (tool name, input, output, errors) to `~/.cache/cc-tools/observations/observations.jsonl`.

No manual hook configuration is needed -- `cc-tools hook` dispatches to the ObserveHandler as part of its standard handler registry.

### 2. Use the Instinct Commands

```bash
cc-tools instinct status               # Show learned instincts with confidence scores
cc-tools instinct evolve               # Cluster related instincts into skills/commands
cc-tools instinct export               # Export instincts as YAML to stdout
cc-tools instinct import <file>        # Import instincts from others
```

Or via slash commands:

```
/instinct-status                       # Show instincts with confidence bars
/evolve                                # Cluster and suggest evolutions
/instinct-export                       # Export for sharing
/instinct-import <file>                # Import from file
```

## Commands

| Command | Description |
|---------|-------------|
| `/instinct-status` | Show all learned instincts grouped by domain with confidence bars |
| `/evolve` | Cluster related instincts, suggest skill/command/agent candidates |
| `/instinct-export` | Export instincts as YAML or JSON for sharing |
| `/instinct-import <file>` | Import instincts with dedup, dry-run, force options |
| `/learn` | Extract patterns from current session into instinct files |
| `/learn-eval` | Extract with quality gate (scoring rubric) before saving |

## Configuration

Managed through `cc-tools config`:

| Key | Default | Description |
|-----|---------|-------------|
| `instinct.personal_path` | `~/.config/cc-tools/instincts/personal` | Personal instincts directory |
| `instinct.inherited_path` | `~/.config/cc-tools/instincts/inherited` | Imported instincts directory |
| `instinct.min_confidence` | `0.3` | Minimum confidence threshold |
| `instinct.auto_approve` | `0.7` | Auto-approve threshold |
| `instinct.decay_rate` | `0.02` | Weekly confidence decay rate |
| `instinct.max_instincts` | `100` | Maximum stored instincts |
| `instinct.cluster_threshold` | `3` | Minimum instincts to form a cluster |

## File Structure

```
~/.config/cc-tools/instincts/
├── personal/              # Auto-learned instincts
│   ├── prefer-functional.md
│   └── always-test-first.md
└── inherited/             # Imported from others
    └── team-conventions.md

~/.cache/cc-tools/observations/
└── observations.jsonl     # Raw session observations
```

## Privacy

- Observations stay **local** on your machine
- Only **instincts** (patterns) can be exported
- Exports include triggers, actions, confidence scores, and domains
- No actual code, file paths, or conversation content is shared
