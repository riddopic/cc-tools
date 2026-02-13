# Detect and Resolve Prompt Contradictions

**Extracted:** 2026-02-08
**Context:** When LLM prompts contain conflicting directives that cause unpredictable behavior

## Problem

Large LLM prompts (300+ lines) accumulate sections from different improvement rounds. Sections added at different times may directly contradict each other, causing the LLM to randomly follow one or the other, producing inconsistent and degraded output.

In Round 4, the ChainOfThoughtTemplate had:

- **State Pre-Validation** section that taught try/catch patterns for testing contract state
- **COMMIT TO ATTACK** directive that explicitly banned try/catch as "reconnaissance"

These contradicted each other. The LLM would sometimes generate try/catch (following State Pre-Validation) and sometimes avoid it (following COMMIT TO ATTACK), producing inconsistent exploit code.

## Solution

### Before Adding to a Prompt

1. **Search for antonyms of your new directive** in the existing prompt
2. **Check for behavioral conflicts**: if section A says "do X" and section B says "never do X", one must go
3. **Test with the specific pattern**: generate output and check if the LLM oscillates between behaviors

### Systematic Contradiction Scan

For each new directive you add, ask:

- Does any existing section teach the opposite behavior?
- Does any existing section use the pattern I'm banning (or ban the pattern I'm teaching)?
- Could the LLM interpret any existing section as conflicting with this one?

### Resolution Strategy

- **Remove the weaker section** entirely (don't just add exceptions)
- **Never add "except when" clauses** to resolve contradictions — they make prompts longer and more confusing
- **Prefer the section that aligns with the desired behavior** for the majority of cases

## Examples

| Contradiction | Resolution |
|--------------|------------|
| "Use try/catch to test state" vs "Never use try/catch" | Remove try/catch teaching (State Pre-Validation) since committed attacks are more important |
| "Generate defensive code" vs "Minimize code complexity" | Remove defensive code guidance — simpler exploits are more likely to compile |
| "Explore all functions" vs "Target only the vulnerable function" | Keep targeted approach — exploration wastes tokens |

## When to Use

- Before adding any new section to an LLM prompt over 200 lines
- When a prompt produces inconsistent output across runs
- When reducing prompt length — check if removed sections contradict remaining ones
- During prompt audits (discovery-oriented compliance checks)
