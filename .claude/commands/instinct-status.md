---
description: Show all learned instincts with their confidence levels
allowed-tools:
  - Bash
---

# /instinct-status - Show Instinct Status

## Arguments

`$ARGUMENTS` can be:

- `--domain <name>` - Filter by domain (code-style, testing, git, etc.)
- `--min-confidence <n>` - Only show instincts above threshold
- (no argument) - Show all instincts grouped by domain

## Implementation

Run the cc-tools instinct CLI:

```bash
cc-tools instinct status $ARGUMENTS
```

## What to Do

1. Run the command above
2. Display the output to the user
3. If no instincts are found, suggest running `/learn` or `/learn-eval` to start building instincts

## Storage

Instincts are stored at:
- Personal: `~/.config/cc-tools/instincts/personal/`
- Inherited: `~/.config/cc-tools/instincts/inherited/`
