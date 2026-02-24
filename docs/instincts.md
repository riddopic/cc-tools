# Understanding Instincts

Instincts are atomic learned behaviors that cc-tools extracts from your Claude Code sessions. Each instinct represents a single pattern -- a trigger condition paired with an action -- that the system has observed you performing repeatedly. Instincts carry a confidence score that increases with reinforcement and decays without it.

## Overview

The instinct system turns repetitive session behavior into persistent, reusable knowledge. Rather than requiring you to manually codify every pattern, cc-tools watches what you do, identifies recurring actions, and stores them as instincts that can influence future sessions.

An instinct answers one question: "When I see X, what should I do?" The system handles the rest -- scoring, pruning, and eventually promoting instincts into higher-level constructs like skills, commands, or agents.

## The Learning Lifecycle

Instincts move through six stages: observation, extraction, confidence scoring, storage, sharing, and evolution. Each stage is independently configurable.

### 1. Observation

The `ObserveHandler` captures tool usage events during every session. It runs automatically as part of `cc-tools hook` dispatch -- you do not need to configure a separate hook for it.

Each observation records:

- Tool name and input parameters
- Tool output and errors
- Phase of the event: `pre`, `post`, or `failure`

Observations are appended to `~/.cache/cc-tools/observations/observations.jsonl`. The file is a newline-delimited JSON log, one event per line.

| Config Key | Type | Default | Description |
|------------|------|---------|-------------|
| `observe.enabled` | bool | `true` | Enable observation logging |
| `observe.max_file_size_mb` | int | `10` | Max observation file size before rotation |

### 2. Instinct Structure

An instinct is a YAML frontmatter file containing these fields:

| Field | Description | Example |
|-------|-------------|---------|
| `id` | Unique identifier | `use-guard-clauses` |
| `trigger` | Condition that activates the instinct | `when reviewing nested conditionals` |
| `confidence` | Score from 0.0 to 1.0 | `0.72` |
| `domain` | Category | `testing`, `workflow`, `code-review` |
| `source` | Origin type | `personal` or `inherited` |
| `source_repo` | Repository the instinct came from (optional) | `github.com/team/project` |
| `created_at` | Creation timestamp (RFC 3339) | `2026-01-15T10:30:00Z` |
| `updated_at` | Last update timestamp (RFC 3339) | `2026-02-20T14:00:00Z` |

The file may also contain free-form content after the closing `---` delimiter, which stores the action or additional context for the instinct.

### 3. Confidence Scoring

Confidence determines how strongly an instinct influences behavior. It is bounded between 0.3 (minimum) and 0.9 (maximum).

**Thresholds that control activation:**

| Config Key | Type | Default | Effect |
|------------|------|---------|--------|
| `instinct.min_confidence` | float | `0.3` | Below this, instincts are inactive |
| `instinct.auto_approve` | float | `0.7` | Above this, instincts apply without confirmation |

**How confidence changes:**

- **Reinforcement** -- when the system observes the pattern again, confidence increases. The base confidence depends on observation count: 3+ observations yield 0.5, 6+ yield 0.7, 11+ yield 0.85.
- **Decay** -- confidence decreases by `instinct.decay_rate` (default 0.02) for each week without reinforcement. This ensures stale instincts gradually lose influence.
- **Pruning** -- the system retains at most `instinct.max_instincts` (default 100) instincts. When the limit is reached, the lowest-confidence instincts are removed first.

### 4. Storage

Instincts are stored as individual YAML frontmatter files in two directories:

| Directory | Purpose | Default Path |
|-----------|---------|--------------|
| Personal | Learned from your sessions | `~/.config/cc-tools/instincts/personal/` |
| Inherited | Imported from teammates or other projects | `~/.config/cc-tools/instincts/inherited/` |

Each file is named `{id}.yaml`. Personal instincts take precedence over inherited ones when both exist with the same ID. Only personal instincts can be deleted through the store; inherited instincts are managed through import.

Paths are configurable via `instinct.personal_path` and `instinct.inherited_path`.

### 5. Import and Export

You can share instincts between projects and teammates through export and import.

**Export** writes instincts to a portable file:

```bash
cc-tools instinct export --format yaml --output my-instincts.yaml
cc-tools instinct export --domain testing --min-confidence 0.5
cc-tools instinct export --format json --output instincts.json
```

The `--format` flag accepts `yaml` (default) or `json`. When `--output` is omitted, instincts are written to stdout.

**Import** brings instincts into your inherited store:

```bash
cc-tools instinct import teammates-instincts.yaml --dry-run
cc-tools instinct import teammates-instincts.yaml --force
cc-tools instinct import shared.yaml --min-confidence 0.4
```

Each instinct in the source file is classified into one of three categories:

| Category | Condition | Behavior |
|----------|-----------|----------|
| New | Instinct does not exist locally | Added to inherited store |
| Overwrite | Instinct exists locally | Updated only with `--force` flag |
| Skip | Duplicate or below `--min-confidence` | Not imported |

The `--dry-run` flag shows what would happen without writing any files.

### 6. Evolution

When enough related instincts accumulate, the evolve command analyzes clusters and suggests promotions to higher-level constructs:

```bash
cc-tools instinct evolve
```

The command groups instincts by shared trigger keywords and evaluates each cluster against three candidate types:

| Candidate Type | Criteria | What It Suggests |
|----------------|----------|------------------|
| Skill | 3+ related instincts in the same domain | A new skill file in `.claude/skills/` |
| Command | Confidence >= 0.7, workflow domain | A new slash command |
| Agent | 3+ instincts in cluster, average confidence >= 0.75 | A specialized agent configuration |

The cluster threshold is configurable via `instinct.cluster_threshold` (default: 3). Evolution is advisory -- it prints suggestions but does not create files automatically.

## CLI Commands

| Command | Description |
|---------|-------------|
| `cc-tools instinct status` | List instincts grouped by domain with confidence bars |
| `cc-tools instinct export` | Export instincts to YAML or JSON |
| `cc-tools instinct import <source>` | Import instincts from a file |
| `cc-tools instinct evolve` | Analyze clusters and suggest evolution candidates |

The `status` command accepts `--domain` and `--min-confidence` flags for filtering. The `export` command accepts `--format`, `--output`, `--domain`, and `--min-confidence`. The `import` command accepts `--dry-run`, `--force`, and `--min-confidence`.

## Configuration Reference

All instinct-related configuration keys, their types, defaults, and descriptions:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `observe.enabled` | bool | `true` | Enable observation logging |
| `observe.max_file_size_mb` | int | `10` | Max observation file size |
| `instinct.personal_path` | string | `~/.config/cc-tools/instincts/personal` | Personal instinct storage directory |
| `instinct.inherited_path` | string | `~/.config/cc-tools/instincts/inherited` | Inherited instinct storage directory |
| `instinct.min_confidence` | float | `0.3` | Minimum confidence for activation |
| `instinct.auto_approve` | float | `0.7` | Auto-approve threshold |
| `instinct.decay_rate` | float | `0.02` | Confidence decay per week without reinforcement |
| `instinct.max_instincts` | int | `100` | Maximum instincts retained |
| `instinct.cluster_threshold` | int | `3` | Minimum cluster size for evolve candidates |

Set any value with `cc-tools config set`:

```bash
cc-tools config set instinct.max_instincts 200
cc-tools config set instinct.decay_rate 0.01
```
