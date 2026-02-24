---
description: Export instincts for sharing with teammates or other projects
allowed-tools:
  - Bash
  - Write
argument-hint: "[--domain <name>] [--min-confidence <n>] [--output <file>] [--format yaml|json]"
---

# /instinct-export - Export Instincts

Exports instincts to a shareable format. Use for:

- Sharing with teammates
- Transferring to a new machine
- Contributing to project conventions

## Arguments

`$ARGUMENTS` can be:

- `--domain <name>` - Export only specified domain
- `--min-confidence <n>` - Minimum confidence threshold (default: 0.3)
- `--output <file>` - Output file path (default: stdout)
- `--format <yaml|json>` - Output format (default: yaml)
- (no argument) - Export all instincts as YAML to stdout

## Implementation

Run the cc-tools instinct CLI:

```bash
cc-tools instinct export $ARGUMENTS
```

## Storage

Instincts are read from:
- Personal: `~/.config/cc-tools/instincts/personal/`
- Inherited: `~/.config/cc-tools/instincts/inherited/`

## Privacy

Exports include triggers, actions, confidence scores, and domains. They do NOT include actual code snippets, file paths, or session transcripts.
