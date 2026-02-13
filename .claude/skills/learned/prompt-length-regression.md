# Prompt Length Has Diminishing Returns

**Extracted:** 2026-02-08
**Context:** When modifying LLM prompt templates for exploit generation (ChainOfThoughtTemplate, refinement prompts)

## Problem

Adding more guidance text to an already-long prompt template (~340 lines) can WORSEN overall results by diluting LLM attention. In Round 3, adding 36 lines / 7 guidance blocks caused:

- Compilation failures: 38% -> 88%
- Pass rate: 7/28 -> 2/28
- Individual test improvements coexisted with massive overall regression

## Solution

Before adding text to `ChainOfThoughtTemplate` or refinement prompts:

1. Count current template lines. If >300 lines, DO NOT add more text
2. Consider REPLACING existing guidance rather than appending
3. Use structured few-shot examples or separate chain-specific prompt files instead
4. Always run full-mode regression (BestOf=3) to validate — quick mode has too much variance

If a prompt change causes >20% increase in compilation failures, REVERT immediately.

## Anti-Pattern: "More Guidance = Better Results"

```go
// BAD: Adding yet another guidance block to a 340-line template
"**NEW GUIDANCE BLOCK**\n" +
"Extra text about burn patterns...\n" +
"More text about flash loans...\n"

// GOOD: Replace weak guidance with stronger, more concise guidance
// Or move specialized guidance to chain-specific context injection
```

## When to Use

- Before modifying `internal/llm/prompts.go` (ChainOfThoughtTemplate)
- Before modifying `internal/llm/prompts/refinement.go`
- When regression results show compilation failures > 40%
- When considering adding "just one more" guidance section
