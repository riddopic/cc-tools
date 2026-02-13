# Subprocess Mixed Output JSON Extraction

**Extracted:** 2026-02-09
**Context:** Parsing JSON results from Go subprocesses that mix structured log lines with a final result JSON blob

## Problem

Go subprocesses (e.g., `exec.CommandContext`) often write structured log lines (one JSON object per line) followed by a final result JSON object. Using `json.Unmarshal` on the combined stdout fails silently because the output contains multiple JSON objects concatenated with newlines — not a single valid JSON document.

This caused the `CalledTargetFunctions` classification bug in Quanta's regression system: every test showed `false` because the result JSON was never parsed.

## Solution

Scan lines in reverse to find the last parseable JSON object, with a fallback to full-output parse for single multi-line JSON blocks:

```go
func extractLastJSONLine(output string, target any) bool {
    trimmed := strings.TrimSpace(output)
    if trimmed == "" {
        return false
    }
    // Try line-by-line in reverse
    lines := strings.Split(trimmed, "\n")
    for i := len(lines) - 1; i >= 0; i-- {
        line := strings.TrimSpace(lines[i])
        if len(line) == 0 || line[0] != '{' {
            continue
        }
        if err := json.Unmarshal([]byte(line), target); err == nil {
            return true
        }
    }
    // Fallback: try entire output as single JSON block
    return json.Unmarshal([]byte(trimmed), target) == nil
}
```

Key design choices:

- **Reverse scan**: The result blob is typically the last line — scanning backward finds it in O(1) for the common case
- **Fallback**: Handles test fixtures and single-JSON outputs that may be pretty-printed across multiple lines
- **Generic target**: Uses `any` parameter so the same helper works for different struct types

## When to Use

- Parsing output from `exec.CommandContext` or `exec.Command` subprocesses
- Any scenario where stdout contains mixed structured logs + a result object
- When `json.Unmarshal` on full output returns an error but you know JSON is present
- When multiple callers need to extract different fields from the same mixed output (share the helper)

## Anti-patterns

Never assume subprocess stdout is a single JSON object:

```go
// BAD: fails silently on mixed output
var result MyStruct
json.Unmarshal([]byte(output), &result) // err != nil, result stays zero-valued
```

Never assume the JSON structure matches your test fixtures without verifying against real runtime output. In Quanta, `extractExploitCodeFromJSONMap` looked for `exploit_code` and `best_exploit.code` but the actual subprocess output used `attempts[].exploit.code`. Always inspect real output when a "fixed" bug persists:

```go
// BAD: only checks two known paths
if v, ok := m["exploit_code"].(string); ok { return v }
if best, ok := m["best_exploit"].(map[string]any); ok { ... }

// GOOD: also checks attempts[].exploit.code (actual subprocess output)
if attempts, ok := m["attempts"].([]any); ok {
    for i := len(attempts) - 1; i >= 0; i-- {
        // ... extract from attempt["exploit"]["code"]
    }
}
```
