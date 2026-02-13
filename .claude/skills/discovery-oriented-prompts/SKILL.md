---
name: discovery-oriented-prompts
description: Maintain LLM prompts that enable autonomous vulnerability discovery. Use when writing or reviewing pattern detection functions, templates, or any guidance that helps the LLM find exploits. Ensures prompts guide investigation rather than prescribe specific solutions.
---

# Discovery-Oriented Prompts Skill

Ensures LLM prompts enable autonomous vulnerability discovery rather than prescribing specific exploits.

## Core Principle

**The LLM should discover exploits autonomously.** Prompts should guide investigation, NOT tell the LLM what specific exploit to use for specific contracts. The goal is that the prompts enable discovery of exploits for OTHER contracts, not just the ones we've seen before.

## Anti-Patterns to Avoid

### 1. Specific Contract References

```go
// ❌ WRONG - Names specific contract and dollar amount
Guidance: "TrueBit lost $26.6M because SafeMath protected multiplication but NOT addition."

// ✅ CORRECT - Describes pattern without naming contracts
Guidance: "INVESTIGATE: Does SafeMath protect ALL arithmetic operations? Look for raw +, -, *, / operators."
```

### 2. Protocol-Specific Pattern Names

```go
// ❌ WRONG - Names protocol in pattern
PatternName: "Balancer V2-style Precision Loss"
PatternName: "SAFEMOON-style Public Burn"

// ✅ CORRECT - Generic descriptive names
PatternName: "Precision Loss (Inconsistent Rounding)"
PatternName: "Public Burn Function"
```

### 3. Explicit Attack Flows

```go
// ❌ WRONG - Prescribes exact attack steps
Guidance: "Attack flow: Flash-mint BPT → drain reserves → execute precision-loss loop → profit"

// ✅ CORRECT - Investigation questions
Guidance: "INVESTIGATE: (1) What happens to precision at small reserve values? " +
    "(2) Are rounding directions consistent between operations?"
```

### 4. Real Examples Section

```go
// ❌ WRONG - Lists specific exploits
"**Real examples:** SAFEMOON, SHADOWFI, BAMBOO had unprotected burn targeting LP."

// ✅ CORRECT - Describes what to look for
"**What to look for:** Unprotected burn functions that accept arbitrary address parameters."
```

## Correct Pattern Structure

```go
func detectPatternN(source, sourceLower string, isProxy bool) *PatternHint {
    // Detection triggers (what code patterns to look for)
    if !hasRequiredSignature(source) {
        return nil
    }

    return &PatternHint{
        PatternNum:  PatternN,
        PatternName: "Generic Descriptive Name",  // No protocol names
        Confidence:  "high",
        Guidance: "CRITICAL: Found [pattern indicator]. " +
            "INVESTIGATE: " +
            "(1) Investigation question about code structure? " +
            "(2) Investigation question about state impact? " +
            "(3) Investigation question about exploitability? " +
            "KEY INSIGHT: [What makes this pattern exploitable without naming examples]",
    }
}
```

## Investigation Question Categories

When writing guidance, use numbered "INVESTIGATE:" questions:

1. **Code Structure**: "Does SafeMath protect ALL operations?"
2. **State Impact**: "What happens when reserves approach zero?"
3. **Access Control**: "Can ANY address call this function?"
4. **Economic Impact**: "What is the price impact of this manipulation?"
5. **Exploitability**: "Can this be triggered in a single transaction?"

## Key Insight Format

End guidance with a KEY INSIGHT that explains the vulnerability mechanism WITHOUT naming specific contracts:

```go
// ✅ Good KEY INSIGHT
"KEY INSIGHT: Partial SafeMath creates false confidence. One unprotected operation breaks everything."

// ❌ Bad KEY INSIGHT
"KEY INSIGHT: Like TrueBit, contracts using SafeMath may miss .add() protection."
```

## Review Checklist

When reviewing prompts across all prompt files (see Files section):

- [ ] No specific contract names (TrueBit, SAFEMOON, Balancer, etc.)
- [ ] No specific dollar amounts ($26.6M, $9M loss, etc.)
- [ ] No protocol-specific pattern names
- [ ] Uses "INVESTIGATE:" numbered questions
- [ ] Has "KEY INSIGHT:" without naming examples
- [ ] "What to look for:" instead of "Real examples:"
- [ ] Pattern names are generic and descriptive

## Auditor Phase Prompts (Gate Pipeline)

When `--gate-pipeline` is enabled (default), Gate 2 produces structured JSON output before any code generation.

### AnalysisPlan JSON Schema

```json
{
  "vulnerability_type": "string",
  "confidence": 0.0-1.0,
  "affected_functions": ["function1", "function2"],
  "attack_vector": "string description",
  "preconditions": ["condition1", "condition2"],
  "estimated_profit_potential": "low|medium|high"
}
```

### Auditor Prompt Principles

1. **JSON Output Only**: Auditor prompts must request structured JSON, not Solidity code
2. **No Code Generation**: The Auditor phase analyzes; the Developer phase generates
3. **Confidence Scoring**: Require explicit confidence (0.0-1.0) for validation gate
4. **Function Identification**: Must identify specific affected functions for Developer context

### Validation Gate

The orchestrator validates Auditor output before proceeding to Developer phase:

- `confidence >= 0.3` required (configurable)
- `affected_functions` must not be empty
- If validation fails, execution stops early (token savings)

### Example Auditor Prompt Structure

```
Analyze this contract for vulnerabilities. Return ONLY valid JSON:
{
  "vulnerability_type": "identified vulnerability or 'none'",
  "confidence": <0.0-1.0>,
  "affected_functions": ["list", "of", "functions"],
  "attack_vector": "how the vulnerability can be exploited",
  "preconditions": ["required conditions for exploit"]
}

INVESTIGATE the contract to determine:
(1) Are there unprotected state-changing functions?
(2) Can external calls be manipulated?
(3) Are there arithmetic operations without SafeMath?
```

## RAG Context Injection Hygiene

Patterns retrieved from vector store must maintain discovery-oriented principles.

### Safety Gate Rules (internal/knowledge/safety/gate.go)

- Protocol blocklist: Uniswap, Aave, Curve, Balancer, Compound, etc.
- Regex patterns: Dollar amounts, step sequences, tx hashes
- Metadata filtering: Remove chain/date/profit data

### RAG Context Checklist

- [ ] No specific protocol names in injected patterns
- [ ] No dollar amounts in "KEY INSIGHT"
- [ ] INVESTIGATE questions are generic
- [ ] Pattern metadata doesn't reveal exploit prevalence
- [ ] Safety gate sanitized before injection

### Files

- Safety Gate: `internal/knowledge/safety/gate.go`
- Sanitizer: `internal/knowledge/indexer/sanitizer.go`
- Retriever: `internal/knowledge/retriever/retriever.go`

## Structured JSON AttackPlan Prompts

When the structured planner generates attack plans via LLM, the prompts must follow discovery-oriented principles.

### Key Functions

- `AttackPlanPromptTemplate` — Template requesting JSON `AttackPlan` output from analysis
- `BuildAttackPlanPrompt()` — Constructs the full prompt with contract context
- `BuildAnalysisPlan()` — Generates the pre-planning analysis prompt

### Discovery Principles for Plan Prompts

```go
// ❌ WRONG - Hardcoded addresses in template
"Use Aave V3 at 0x7d2768... for the flash loan"

// ✅ CORRECT - Generic provider selection
"Select an appropriate flash loan provider based on chain and liquidity"

// ❌ WRONG - Protocol-specific step sequences
"Step 1: Borrow from dYdX, Step 2: Swap on Uniswap V2"

// ✅ CORRECT - Generic action descriptions
"INVESTIGATE: What flash loan providers are available? What DEX has the most liquidity for this token?"
```

- Plan prompt templates must NOT contain hardcoded contract addresses
- Plan prompt templates must NOT prescribe protocol-specific step sequences
- Actor role descriptions must be generic (e.g., "flash loan receiver" not "Aave callback handler")

## Refinement Prompts from Feedback

The `FeedbackOrchestrator` sends trace analysis feedback back to the LLM for exploit refinement.

### Feedback → Prompt Cycle

1. `TraceAnalyzer` produces `TraceFeedback` with `FailureType` and `SuggestedFix`
2. Refinement prompt incorporates failure context
3. LLM adjusts exploit based on feedback

### FailureType-Specific Adjustments

| FailureType | Prompt Adjustment |
|-------------|-------------------|
| `FailureRevert` | "The exploit reverted. INVESTIGATE: What require() condition failed?" |
| `FailureBalance` | "Insufficient balance. INVESTIGATE: Does the flash loan amount cover all transfers?" |
| `FailureAccess` | "Access denied. INVESTIGATE: What access control modifier blocks this call?" |
| `FailureNoProfit` | "Exploit executed but no profit. INVESTIGATE: Are extraction fees reducing net profit to zero?" |
| `FailureRepayment` | "Flash loan repayment failed. INVESTIGATE: Are extracted tokens routed back to the lender?" |

### Discovery Compliance

- Refinement suggestions must remain **generic** — describe the failure category, not a specific fix
- Must use "INVESTIGATE:" format for suggested remediation
- Must NOT inject specific protocol names or addresses into refinement prompts

## Body/State-Aware RAG Context in Prompts

Enhanced RAG now injects function body content and state variable patterns into prompts.

### Function Body Context

- Chunked function bodies appear as investigation targets
- The prompt frames them as "INVESTIGATE this function body for [pattern indicators]"
- Body content is sanitized through the safety gate before injection

### State Variable Context

- State variable access patterns (reads/writes) provide cross-function analysis context
- Prompt format: "This function reads `balanceOf` and writes `totalSupply` — INVESTIGATE: Is the state update order safe?"

### Category-Specific Ensemble Predictions

- When multi-model ensemble provides category confidence, it appears as context: "Ensemble confidence: access_control (0.92)"
- The prompt uses this to focus investigation without prescribing the exploit

## Review Checklist

When reviewing prompts across the prompt file tree:

- [ ] No specific contract names (TrueBit, SAFEMOON, Balancer, etc.)
- [ ] No specific dollar amounts ($26.6M, $9M loss, etc.)
- [ ] No protocol-specific pattern names
- [ ] Uses "INVESTIGATE:" numbered questions
- [ ] Has "KEY INSIGHT:" without naming examples
- [ ] "What to look for:" instead of "Real examples:"
- [ ] Pattern names are generic and descriptive
- [ ] Plan prompt templates have no hardcoded addresses
- [ ] Plan actor roles use generic descriptions
- [ ] Refinement prompts use INVESTIGATE format (not specific fixes)
- [ ] Body/state RAG context passes through safety gate
- [ ] CLI client MD constants follow same discovery-oriented principles

## Files

Gate Pipeline Prompts:
- `internal/llm/prompts.go` — Core gate prompts (1-4) + pattern detection functions
- `internal/llm/prompts_gate2a.go` — Gate 2 Auditor perspective prompt
- `internal/llm/prompts_gate2b.go` — Gate 2 Attacker perspective prompt

Prompt Template Library:
- `internal/llm/prompts/refinement.go` — Feedback loop refinement templates
- `internal/llm/prompts/monetization.go` — Monetization refinement prompt

Agentic & Prediction:
- `internal/agent/agentic/prompt.go` — Agentic loop system prompt
- `internal/llm/predictor.go` — Vulnerability prediction prompt
- `internal/writeup/prompts.go` — Writeup documentation prompt

CLI Client Prompts:
- `internal/llm/gemini_client.go` — GEMINI.md content constants
- `internal/llm/kimi_client.go` — KIMI.md content constants
- `internal/llm/codex_client.go` — AGENTS.md content constants
- `internal/llm/claudecode_client.go` — Hardcoded system prompt

Supporting:
- Analysis Plan Types: `internal/interfaces/analysis_plan.go`
- RAG Safety Gate: `internal/knowledge/safety/gate.go`
- Structured Planner: `internal/coordinator/structured_planner_backend.go`
- Feedback Loop: `internal/agent/feedback_loop.go`

## Related Skills

- `/pattern-management` - Add or update patterns (use this skill for HOW to write them)
- `/a1-asceg-research` - Research exploit patterns from papers
- `/rag-knowledge-system` - RAG implementation and vector store patterns
