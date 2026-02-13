# Golines Multiline Lint Cascade

**Extracted:** 2026-02-10
**Context:** Fixing long-line lint errors in Go code triggers cascading lint failures

## Problem

When a `golines` lint error forces you to break a long string/call across multiple lines, two additional lint issues commonly appear:

1. **gofumpt alignment** — If the broken line was in a `var` block with aligned `=` signs, the alignment padding (extra spaces) becomes invalid after one entry goes multiline
2. **perfsprint** — When breaking `fmt.Errorf("static string")` across lines, the linter notices there are no format verbs and demands `errors.New()` instead

## Solution

When fixing a golines violation:

1. Break the line across multiple lines
2. Check if the line was in an aligned `var`/`const` block — if so, remove extra alignment spaces from sibling entries
3. Check if the call was `fmt.Errorf` with no `%` verbs — if so, switch to `errors.New`
4. Verify `"errors"` is in the import block when switching to `errors.New`

## Example

```go
// Before (golines violation):
var (
    errNoRPCURLs = errors.New("no RPC URLs configured (set ETH_RPC_URL, BSC_RPC_URL, ARBITRUM_RPC_URL, BASE_RPC_URL, and/or SEI_RPC_URL)")
    errNoAPIKey  = errors.New("no LLM API key found")
)

// After (all three fixes applied):
var (
    errNoRPCURLs = errors.New(
        "no RPC URLs configured (set ETH_RPC_URL, BSC_RPC_URL, ARBITRUM_RPC_URL, BASE_RPC_URL, and/or SEI_RPC_URL)",
    )
    errNoAPIKey = errors.New("no LLM API key found")  // removed extra space
)
```

## When to Use

When breaking long Go lines to fix golines lint errors, especially in `var` blocks or `fmt.Errorf` calls.
