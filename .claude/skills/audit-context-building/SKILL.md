---
name: audit-context-building
description: Enables ultra-granular, line-by-line code analysis to build deep architectural context before vulnerability or bug finding.
context: fork
---

# Deep Context Builder (Ultra-Granular Pure Context Mode)

Build deep, accurate understanding of a codebase through line-by-line analysis **before** vulnerability hunting. This is pure context building -- no findings, no fixes, no exploits.

## When to Use

- Deep comprehension needed before bug or vulnerability discovery
- Bottom-up understanding instead of high-level guessing
- Reducing hallucinations, contradictions, and context loss
- Preparing for security auditing, architecture review, or threat modeling

**Not for:** vulnerability findings, fix recommendations, exploit reasoning, severity rating.

## Phase 1 -- Initial Orientation

Before deep analysis, perform a minimal mapping:

1. Identify major modules/files/contracts
2. Note public/external entrypoints
3. Identify actors (users, owners, relayers, oracles, other contracts)
4. Identify important storage variables, state structs, or cells
5. Build preliminary structure without assuming behavior

## Phase 2 -- Ultra-Granular Function Analysis

Every non-trivial function receives full micro analysis.

### Per-Function Checklist

For each function:

1. **Purpose** -- Why the function exists and its role in the system

2. **Inputs & Assumptions**
   - Parameters and implicit inputs (state, sender, env)
   - Preconditions and constraints

3. **Outputs & Effects**
   - Return values, state/storage writes
   - Events/messages, external interactions

4. **Block-by-Block / Line-by-Line Analysis**
   For each logical block:
   - What it does
   - Why it appears here (ordering logic)
   - What assumptions it relies on
   - What invariants it establishes or maintains
   - What later logic depends on it

   Apply per-block: **First Principles**, **5 Whys**, **5 Hows**

### Cross-Function & External Flow

When encountering calls, continue the same micro-first analysis across boundaries.

**Internal calls:**
- Jump into the callee immediately
- Perform block-by-block analysis of relevant code
- Track data, assumptions, and invariants: caller -> callee -> return -> caller

**External calls with available code:**
- Treat as internal -- jump in, continue micro-analysis
- Consider edge cases based on actual code, not black-box guesses

**External calls without available code (black box):**
- Describe payload/parameters sent
- Identify assumptions about the target
- Consider all outcomes: revert, incorrect returns, unexpected state changes, reentrancy

**Continuity rule:** Treat the entire call chain as one continuous execution flow. Never reset context. All invariants and dependencies propagate across calls.

### Completeness Check

Before concluding analysis of a function, verify:
- All sections present (Purpose, Inputs, Outputs, Block-by-Block, Dependencies)
- Minimum 3 invariants, 5 assumptions, 3 risk considerations for external interactions
- At least 1 First Principles and 3 combined 5 Whys/5 Hows applications
- Cross-references to related functions and shared state documented

See [COMPLETENESS_CHECKLIST.md](resources/COMPLETENESS_CHECKLIST.md) and [FUNCTION_MICRO_ANALYSIS_EXAMPLE.md](resources/FUNCTION_MICRO_ANALYSIS_EXAMPLE.md) for detailed format and walkthrough.

## Phase 3 -- Global System Understanding

After sufficient micro-analysis:

1. **State & Invariant Reconstruction** -- Map reads/writes of each state variable; derive multi-function invariants
2. **Workflow Reconstruction** -- Identify end-to-end flows; track state transforms; record persistent assumptions
3. **Trust Boundary Mapping** -- Actor -> entrypoint -> behavior; untrusted input paths; privilege changes
4. **Complexity & Fragility Clustering** -- Functions with many assumptions, high branching, multi-step dependencies, coupled state changes

These clusters guide the vulnerability-hunting phase.

## Subagent Usage

Use the **`function-analyzer`** agent for per-function deep analysis. It follows the full microstructure checklist and quality thresholds defined here.

For large codebases, apply `recursive-decomposition` to partition analysis across subagents rather than loading everything into a single context.

## Non-Goals

While active, do NOT: identify vulnerabilities, propose fixes, generate proofs-of-concept, model exploits, or assign severity. This is **pure context building** only.
