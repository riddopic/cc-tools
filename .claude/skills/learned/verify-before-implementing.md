# Verify Existing Mitigations Before Implementing New Ones

**Extracted:** 2026-02-08
**Context:** When working from an improvement plan that lists features to add

## Problem

Improvement plans may list features as "TODO" that were actually already implemented in a prior session or by a different code path. Implementing a duplicate creates dead code, wasted effort, and can introduce bugs.

In this project, the regression improvement plan listed:

- "2.1: Wire interface auto-extraction to Gate 4" — already done in gate4.go:101 and feedback_loop.go:948
- "2.2: Wire pragma extraction to Gate 4" — already done in feedback_loop.go:949

Both were already fully wired. Without verification, hours would have been spent re-implementing existing functionality.

## Solution

Before implementing any item from an improvement plan:

1. **Search for the feature in code** — use Grep/Explore to find where the function is already called
2. **Trace the data flow** — verify the extracted data actually reaches the prompt template
3. **Check both code paths** — Quanta has two exploit generation paths:
   - `executor.go` → `ChainOfThoughtTemplate` (legacy, used by in-process mode)
   - `feedback_loop.go` → `InitialExploitInput`/`RefinementInput` (new, used by subprocess mode)
4. **Check which path the regression uses** — subprocess mode uses feedback_loop.go, which is the path that matters for regression

## Key Code Locations

| Feature | Initial (iter 0) | Refinement (iter 1+) |
|---------|------------------|---------------------|
| PreExtractedInterface | gate4.go:101, feedback_loop.go:948 | feedback_loop.go:998 |
| ExtractPragma | feedback_loop.go:949 | feedback_loop.go:1035 |
| ChainDEXInfo | feedback_loop.go:950 | feedback_loop.go:1036 |

## When to Use

- Before implementing any item from `docs/plans/*.md`
- When an improvement plan says "wire X to Y" or "add X support"
- When a feature "should" exist based on the improvement plan but you're not sure
- Before adding new template parameters to prompts
