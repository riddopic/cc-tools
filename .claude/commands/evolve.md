---
description: Cluster related instincts into skills, commands, or agents
allowed-tools:
  - Bash
  - Read
  - Write
argument-hint: "[--domain <name>]"
---

# /evolve - Evolve Instincts into Higher-Level Structures

Analyzes instincts and clusters related ones into higher-level structures:

- **Skills**: 3+ related instincts in the same domain
- **Commands**: High-confidence (>=0.7) workflow instincts
- **Agents**: Large clusters (3+) with >=0.75 average confidence

## Arguments

`$ARGUMENTS` can be:

- `--domain <name>` - Only evolve instincts in specified domain
- (no argument) - Analyze all instincts and suggest evolutions

## Implementation

Run the cc-tools instinct CLI:

```bash
cc-tools instinct evolve $ARGUMENTS
```

## What to Do

1. Run the command above
2. Display cluster analysis to the user
3. If candidates are found, discuss which ones to create
4. Requires 3+ instincts to perform analysis

## Storage

Instincts are read from `~/.config/cc-tools/instincts/`.
