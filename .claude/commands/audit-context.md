---
description: Builds deep architectural context before vulnerability hunting
allowed-tools:
  - Read
  - Grep
  - Glob
  - Task
argument-hint: "<codebase-path> [--focus <module>]"
model: opus
---

# Build Audit Context

**Arguments:** $ARGUMENTS

Parse arguments:

1. **Codebase path** (required): Path to codebase to analyze
2. **Focus** (optional): `--focus <module>` for specific module analysis

## Workflow

Execute the `audit-context-building` skill through these phases:

### Phase 1 — Initial Orientation

Perform a bottom-up scan to establish structural anchors:

1. Use Glob to identify major modules, files, and contracts in the target path.
2. Use Grep to locate public/external entrypoints (`external`, `public`, exported functions).
3. Identify actors (users, owners, relayers, oracles, other contracts).
4. Identify important storage variables, state structs, or mappings.
5. Build a preliminary module map without assuming behavior.

If `--focus <module>` is specified, narrow the orientation to that module and its immediate dependencies.

### Phase 2 — Ultra-Granular Micro-Analysis

For each non-trivial function in scope:

1. Read the function source with the Read tool.
2. Apply the per-function microstructure checklist from the skill:
   - Purpose, Inputs & Assumptions, Outputs & Effects
   - Block-by-block / line-by-line analysis
   - First Principles, 5 Whys, 5 Hows per logical block
3. For internal calls: jump into the callee and continue micro-analysis.
4. For external calls without source: model as adversarial (all outcomes).
5. Maintain continuity — never reset context across call chains.

Use Task with subagents for dense or complex functions, cryptographic logic, or long call chains.

### Phase 3 — Global System Understanding

After sufficient micro-analysis, synthesize:

1. **State & Invariant Reconstruction** — map reads/writes, derive cross-function invariants.
2. **Workflow Reconstruction** — trace end-to-end flows (deposit, withdraw, lifecycle, upgrades).
3. **Trust Boundary Mapping** — actor to entrypoint to behavior; untrusted input paths.
4. **Complexity & Fragility Clustering** — functions with many assumptions, high branching, coupled state.

## Output

Save the context analysis report to: `docs/audits/context-audit-{YYYYMMDD-HHMMSS}.md`

Structure the report as:

```markdown
# Context Audit: {target}

**Date:** {timestamp}
**Scope:** {codebase-path} {focus if specified}

## Module Map
{orientation results}

## Function Micro-Analyses
{per-function analysis results}

## Global Understanding
### State & Invariants
### Workflows
### Trust Boundaries
### Fragility Clusters

## Open Questions
{unresolved items requiring further investigation}
```

## Integration

This command produces context consumed by downstream phases:

- **Vulnerability discovery** — uses invariants and trust boundaries to find violations
- **Exploit reasoning** — uses workflows and state maps to construct attack paths
- **Threat modeling** — uses actor mapping and fragility clusters to prioritize
