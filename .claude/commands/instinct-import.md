---
description: Import instincts from teammates, Skill Creator, or other sources
allowed-tools:
  - Bash
  - Read
argument-hint: "<file> [--dry-run] [--force] [--min-confidence <n>]"
---

# /instinct-import - Import Instincts

Import instincts from:

- Teammates' exports
- Community collections
- Previous machine backups

## Arguments

`$ARGUMENTS` should be a file path plus optional flags:

- `<file>` - Required. Local path to instinct YAML file
- `--dry-run` - Preview without importing
- `--force` - Overwrite existing instincts with same ID
- `--min-confidence <n>` - Only import instincts above threshold

## Implementation

Run the cc-tools instinct CLI:

```bash
cc-tools instinct import $ARGUMENTS
```

## What to Do

1. Run the command above
2. Display import results (added, skipped, updated counts)
3. Suggest running `/instinct-status` to see all instincts

## Storage

Imported instincts are saved to `~/.config/cc-tools/instincts/inherited/`.
