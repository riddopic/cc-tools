---
description: Consult your AI team (Gemini, Codex, Claude and Kimi) for second opinions and research
argument-hint: "<claude|gemini|codex|kimi|all> <query>"
skills:
  - consult
---

## Argument Routing

Parse `$ARGUMENTS` to determine which model(s) to consult:

1. Extract the **first word** of `$ARGUMENTS` (case-insensitive)
2. If it matches a known target (`claude`, `gemini`, `codex`, `kimi`, `all`), use that as the target and the **rest** as the query
3. If it does NOT match a known target, treat the **entire** `$ARGUMENTS` as the query and default target to `all`

### Target → Action

| Target | Action |
|--------|--------|
| `claude` | Run `claude -p "$QUERY" --model opus --output-format json` as a background task |
| `gemini` | Run `gemini "$QUERY" -m gemini-3-pro -o json` as a background task |
| `codex` | Run `codex exec --json "$QUERY"` as a background task |
| `kimi` | Run `kimi-cli -p "$QUERY" --quiet` as a background task |
| `all` | Run **all four** commands above in parallel as background tasks |

### Output Parsing

After each background task completes, parse the output:

- **Claude:** `jq -r '.result'`
- **Gemini:** `jq -r '.response'`
- **Codex:** `jq -rs 'map(select(.item.type? == "agent_message")) | last | .item.text'`
- **Kimi:** Plain text (no parsing needed)

### Execution

Save all outputs to the session scratchpad directory. After all targeted models complete, synthesize the results into a unified response highlighting:

- Points of **agreement** across models
- **Unique insights** from individual models
- Any **contradictions** or **disagreements**

## Input

$ARGUMENTS
