---
description: Audit LLM prompts to ensure they enable autonomous vulnerability discovery rather than prescribing specific exploits. Validates prompts follow discovery-oriented patterns.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Task
  - Write
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "[all|templates|patterns|function:NAME]"
model: opus
---

# Audit LLM Prompts for Discovery-Oriented Compliance

**IMPORTANT**: Think carefully about this audit. Enter plan mode to systematically analyze all prompts before executing checks.

## Core Principle

**The LLM should discover exploits autonomously.** Prompts must guide investigation, NOT tell the LLM what specific exploit to use for specific contracts. The goal is that prompts enable discovery of exploits for OTHER contracts, not just ones we've seen before.

## Scope: $ARGUMENTS

Parse the argument to determine audit scope:

- `all` (default) - Audit all prompt files, templates, AND pattern functions
- `templates` - Audit template strings across all prompt files (prompts.go, prompts_gate2a/b.go, prompts/refinement.go, prompts/monetization.go, prompts/personas.go, agentic/prompt.go, predictor.go, writeup/prompts.go)
- `patterns` - Audit detectPattern* functions in prompts.go AND gate 2a/2b perspective prompts
- `clients` - Audit CLI client MD content constants (gemini_client.go, kimi_client.go, codex_client.go, claudecode_client.go)
- `function:NAME` - Audit a specific function (e.g., `function:detectPattern8`)

## Target Files

Primary:

- `internal/llm/prompts.go` — Core gate prompts (1-4) + pattern detection functions
- `internal/llm/prompts_gate2a.go` — Gate 2 Auditor perspective prompt
- `internal/llm/prompts_gate2b.go` — Gate 2 Attacker perspective prompt

Prompt Template Library:

- `internal/llm/prompts/personas.go` — Ensemble analysis personas (Auditor, Attacker, Invariant)
- `internal/llm/prompts/refinement.go` — Feedback loop refinement templates
- `internal/llm/prompts/monetization.go` — Monetization refinement prompt

Agentic & Prediction:

- `internal/agent/agentic/prompt.go` — Agentic loop system prompt
- `internal/llm/predictor.go` — Vulnerability prediction prompt
- `internal/writeup/prompts.go` — Writeup documentation prompt

RAG Components:

- `internal/knowledge/safety/gate.go` — Safety gate for RAG context injection
- `internal/knowledge/indexer/sanitizer.go` — Pattern sanitizer for indexed content
- `internal/knowledge/retriever/retriever.go` — RAG retrieval with context injection

Advanced Pipeline Components:

- `internal/coordinator/structured_planner_backend.go` — Structured plan prompt generation
- `internal/agent/feedback_loop.go` — Feedback-driven refinement prompts

CLI Client Prompts:

- `internal/llm/gemini_client.go` — GEMINI.md content (geminiMDHeader + constants)
- `internal/llm/kimi_client.go` — KIMI.md content (kimiMDHeader + constants)
- `internal/llm/codex_client.go` — AGENTS.md content (agentsMDHeader + constants)
- `internal/llm/claudecode_client.go` — Hardcoded system prompt (line ~680)

## Phase 1: Automated Anti-Pattern Scanning

Execute these Grep checks to find potential violations:

### Check 1: Specific Contract Names

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/prompts.go --context=2
```

### Check 2: Dollar Amounts with Context

```bash
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/prompts.go
```

### Check 3: Protocol-Specific Pattern Names

```bash
rg 'PatternName:.*".*V[0-9]|PatternName:.*style' internal/llm/prompts.go
```

### Check 4: Explicit Attack Flows

```bash
rg -in "(attack flow:|attack sequence:|step 1:.*flash|→.*→.*→)" internal/llm/prompts.go
```

### Check 5: Real Examples Sections

```bash
rg -in "(real examples:|real-world:|actual exploit:)" internal/llm/prompts.go
```

### Check 6: RAG Safety Gate Compliance

```bash
rg -in "(protocol|uniswap|aave|curve|balancer)" internal/knowledge/safety/gate.go
rg -in "(protocol|uniswap|aave|curve|balancer)" internal/knowledge/indexer/sanitizer.go
```

### Check 7: RAG Retriever Context Injection

```bash
rg -in "(\$[0-9]+[MKB]|step 1:|attack flow:)" internal/knowledge/retriever/retriever.go
```

### Check 8: Structured Plan Prompt Compliance

```bash
rg -in "(0x[0-9a-fA-F]{40}|aave.*0x|uniswap.*0x|balancer.*0x)" internal/coordinator/structured_planner_backend.go
```

### Check 9: Plan Template Discovery Violations

```bash
rg -in "(step 1:.*flash|step 1:.*borrow|attack flow:)" internal/coordinator/structured_planner_backend.go
```

### Check 10: Refinement Prompt Leakage

```bash
rg -in "(specific fix:|do this:|use aave|use uniswap|use balancer)" internal/agent/feedback_loop.go
```

### Check 11: Gate 2a/2b Dual Perspective Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/prompts_gate2a.go internal/llm/prompts_gate2b.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/prompts_gate2a.go internal/llm/prompts_gate2b.go
rg -in "(real examples:|real-world:|actual exploit:)" internal/llm/prompts_gate2a.go internal/llm/prompts_gate2b.go
```

### Check 12: Persona Prompt Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/prompts/personas.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/prompts/personas.go
rg -in "(real examples:|real-world:|actual exploit:)" internal/llm/prompts/personas.go
```

### Check 13: Refinement Template Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/prompts/refinement.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/prompts/refinement.go
rg -in "(attack flow:|attack sequence:|step 1:.*flash|→.*→.*→)" internal/llm/prompts/refinement.go
```

### Check 14: Monetization Prompt Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/prompts/monetization.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/prompts/monetization.go
rg -in "(specific fix:|do this:|use aave|use uniswap|use balancer)" internal/llm/prompts/monetization.go
```

### Check 15: Agentic Loop Prompt Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/agent/agentic/prompt.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/agent/agentic/prompt.go
rg -in "(real examples:|real-world:|actual exploit:)" internal/agent/agentic/prompt.go
```

### Check 16: CLI Client MD Content Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo)" internal/llm/gemini_client.go internal/llm/kimi_client.go internal/llm/codex_client.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/gemini_client.go internal/llm/kimi_client.go internal/llm/codex_client.go
rg -in "(real examples:|real-world:|actual exploit:)" internal/llm/gemini_client.go internal/llm/kimi_client.go internal/llm/codex_client.go
```

### Check 17: ClaudeCode Hardcoded System Prompt

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/claudecode_client.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/claudecode_client.go
```

### Check 18: Predictor and Writeup Prompt Anti-Patterns

```bash
rg -in "(truebit|safemoon|shadowfi|bamboo|balancer|curve)" internal/llm/predictor.go internal/writeup/prompts.go --context=2
rg '\$[0-9]+(\.[0-9]+)?[MKB]' internal/llm/predictor.go internal/writeup/prompts.go
rg -in "(real examples:|real-world:|actual exploit:)" internal/llm/predictor.go internal/writeup/prompts.go
```

## Phase 2: Sub-Agent Deep Analysis

### Agent 1: prompt-engineering-expert

Task: Analyze prompt templates for discovery enablement

- Does the prompt guide investigation or prescribe solutions?
- Are there prescriptive sequences that limit exploration?
- Do prompts enable transfer learning to unseen contracts?

### Agent 2: security-threat-analyst

Task: Assess information leakage

- Do prompts reveal specific exploit methodologies?
- Are specific attack vectors being handed to the LLM?
- Could a competitor extract our exploit knowledge from prompts?

### Agent 3: code-analyzer-debugger

Task: Trace pattern detection functions

- Extract all PatternName values
- Extract all Guidance strings
- Check for hardcoded contract addresses in detection logic

## Phase 3: Per-Function Audit

For each `detectPattern*()` function, evaluate:

| Criterion | Weight | Pass/Fail |
| --------- | ------ | --------- |
| No contract names in Guidance | 25% | |
| Uses INVESTIGATE: format | 20% | |
| Generic PatternName (no protocols) | 15% | |
| KEY INSIGHT without victims | 15% | |
| No dollar amounts | 10% | |
| Detection based on code signals | 10% | |
| Actionable investigation questions | 5% | |

## Anti-Pattern Checklist

From `.claude/skills/discovery-oriented-prompts/SKILL.md`:

### VIOLATIONS (Must Not Contain)

- [ ] Specific contract names (TrueBit, SAFEMOON, Balancer, etc.)
- [ ] Specific dollar amounts ($26.6M, $9M loss, etc.)
- [ ] Protocol-specific pattern names
- [ ] "Real examples:" sections
- [ ] Explicit multi-step attack flows
- [ ] Copy-paste exploit recipes

### REQUIRED (Must Contain)

- [ ] "INVESTIGATE:" numbered questions
- [ ] "KEY INSIGHT:" without naming examples
- [ ] "What to look for:" instead of "Real examples:"
- [ ] Generic, descriptive pattern names
- [ ] First-principles investigation methodology

## Scoring Rubric

| Score | Grade | Action |
| ----- | ----- | ------ |
| 90-100 | A | Approved for production |
| 80-89 | B | Minor improvements recommended |
| 70-79 | C | Review before production use |
| 60-69 | D | Significant remediation required |
| <60 | F | Reject - major rewrite needed |

**Automatic Failure Triggers:**

- Named contract + dollar amount in same guidance
- "Real examples:" section present
- Protocol-specific pattern name (e.g., "Balancer V2-style")
- Explicit attack flow prescription

## Phase 4: Generate Audit Report

Create output file at: `docs/audits/prompt-audit-{YYYYMMDD-HHMMSS}.md`

### Report Structure

```markdown
# LLM Prompt Audit Report

**Audit Date:** {timestamp}
**Scope:** {argument}
**Overall Score:** {score}/100 ({grade})

## Executive Summary
- Total items audited: {count}
- Violations found: {count}
- Compliance rate: {percentage}%

## Anti-Pattern Scan Results

### Critical Violations
{findings with line numbers}

### Warnings
{findings requiring manual review}

## Template Analysis

### SystemPromptTemplate
- Score: X/100
- Findings: {details}

### DiscoveryInvestigationTemplate
- Score: X/100
- Findings: {details}

[...for each template]

## Pattern Function Analysis

### detectPattern8 (Public Burn Function)
- PatternName: {extracted}
- Compliance: PASS/FAIL
- Issues: {list}

[...for each pattern function]

## Recommendations

### Immediate Actions (Critical)
1. {specific remediation}

### Before Next Release
1. {improvements}

## Appendix: Files Analyzed

Gate Pipeline Prompts:
- internal/llm/prompts.go
- internal/llm/prompts_gate2a.go
- internal/llm/prompts_gate2b.go

Prompt Template Library:
- internal/llm/prompts/personas.go
- internal/llm/prompts/refinement.go
- internal/llm/prompts/monetization.go

Agentic & Prediction:
- internal/agent/agentic/prompt.go
- internal/llm/predictor.go
- internal/writeup/prompts.go

RAG Components:
- internal/knowledge/safety/gate.go
- internal/knowledge/indexer/sanitizer.go
- internal/knowledge/retriever/retriever.go

Advanced Pipeline:
- internal/coordinator/structured_planner_backend.go
- internal/agent/feedback_loop.go

CLI Client Prompts:
- internal/llm/gemini_client.go
- internal/llm/kimi_client.go
- internal/llm/codex_client.go
- internal/llm/claudecode_client.go

Pattern functions: {count}
Templates: {count}
```

## Structured Plan Prompt Audit

When auditing structured planner components, evaluate:

### BuildAttackPlanPrompt()

- [ ] No hardcoded contract addresses in the template
- [ ] No protocol-specific step sequences (e.g., "borrow from dYdX")
- [ ] Actor role descriptions are generic (not provider-specific)
- [ ] Uses INVESTIGATE format for investigation context
- [ ] Flash loan provider selection is parameterized, not hardcoded

### BuildAnalysisPlan()

- [ ] Analysis prompts guide investigation, not prescribe attack vectors
- [ ] Category predictions from ensemble are presented as context, not directives
- [ ] Function body context from enhanced RAG passes through safety gate

### Refinement Prompts (feedback_loop.go)

- [ ] FailureType-specific adjustments use INVESTIGATE format
- [ ] No specific protocol names injected into refinement context
- [ ] Suggested fixes describe the problem category, not a specific solution

## Remediation Guidance

When violations are found, transform prescriptive to investigative:

| Bad (Prescriptive) | Good (Investigative) |
| ------------------ | -------------------- |
| "TrueBit lost $26.6M because SafeMath protected multiplication but NOT addition" | "INVESTIGATE: Does SafeMath protect ALL arithmetic operations? Look for raw +, -, *, / operators" |
| "Real examples: SAFEMOON, SHADOWFI, BAMBOO had unprotected burn" | "What to look for: Unprotected burn functions that accept arbitrary address parameters" |
| "Attack flow: Flash-mint BPT → drain reserves" | "INVESTIGATE: (1) Can pool tokens be flash-minted? (2) What happens to reserves at extreme values?" |
| "Balancer V2-style Precision Loss" | "Precision Loss (Inconsistent Rounding)" |

## Acceptable Infrastructure References

DEX addresses ARE acceptable (infrastructure, not exploit templates):

- Uniswap/PancakeSwap Router addresses
- Factory addresses
- WETH/WBNB addresses
- Flash loan provider addresses

## RAG Safety Gate Audit

When auditing RAG components, verify:

### Safety Gate Rules (internal/knowledge/safety/gate.go)

- [ ] Protocol blocklist includes: Uniswap, Aave, Curve, Balancer, Compound, MakerDAO
- [ ] Regex patterns block: Dollar amounts ($1M, 1000 ETH), step sequences, tx hashes
- [ ] Metadata filtering removes: chain/date/profit data from indexed patterns

### Sanitizer Rules (internal/knowledge/indexer/sanitizer.go)

- [ ] Protocol names stripped before indexing
- [ ] Dollar amounts anonymized (e.g., "$26.6M" → "[AMOUNT]")
- [ ] Contract addresses not included in searchable content

### Retriever Context Injection (internal/knowledge/retriever/retriever.go)

- [ ] Retrieved patterns pass through safety gate before injection
- [ ] No specific protocol names in injected context
- [ ] INVESTIGATE questions remain generic after injection

## Integration

This command integrates with:

- `.claude/skills/discovery-oriented-prompts/SKILL.md` - Source of truth for anti-patterns
- `.claude/skills/rag-knowledge-system/SKILL.md` - RAG implementation patterns
- `.claude/skills/multi-actor-coordination/SKILL.md` - Structured plan types
- `.claude/skills/exploit-debugging/SKILL.md` - Feedback loop debugging
- `.claude/commands/validate-a1-asceg.md` - Similar validation pattern
