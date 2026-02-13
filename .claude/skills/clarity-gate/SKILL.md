---
name: clarity-gate
description: >
  Pre-ingestion verification for epistemic quality in RAG systems.
  Ensures documents are properly qualified before entering knowledge bases.
  Use when preparing documents for RAG ingestion, checking for hallucination
  risks, reviewing for equivocation, or validating clarity-gated documents.
context: fork
---

# Clarity Gate

**Core Question:** "If another LLM reads this document, will it mistake assumptions for facts?"

**Core Principle:** Detection finds what is; enforcement ensures what should be. Find the missing uncertainty markers before they become confident hallucinations.

## The Key Distinction

| Tool Type | Question | Example |
|-----------|----------|---------|
| **Detection** | "Does this text contain hedges?" | UnScientify/HedgeHunter find "may", "possibly" |
| **Enforcement** | "Should this claim be hedged but isn't?" | Clarity Gate flags "Revenue will be $50M" |

---

## When to Use

- Before ingesting documents into RAG systems
- Before sharing documents with other AI systems
- After writing specifications, state docs, or methodology descriptions
- When a document contains projections, estimates, or hypotheses
- Before publishing claims that haven't been validated
- When handing off documentation between LLM sessions

---

## Critical Limitation

> **Clarity Gate verifies FORM, not TRUTH.**
>
> This skill checks whether claims are properly marked as uncertain — it cannot verify if claims are actually true.
>
> **Risk:** An LLM can hallucinate facts INTO a document, then "pass" Clarity Gate by adding source markers to false claims.
>
> **Solution:** HITL (Human-In-The-Loop) verification is **MANDATORY** before declaring PASS.

---

## The 9 Verification Points

Semantic review checks that require judgment. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed examples.

| # | Point | Question |
|---|-------|----------|
| 1 | Hypothesis vs Fact | Is this claim marked as validated or hypothetical? |
| 2 | Uncertainty Markers | Do forward-looking statements have qualifiers? |
| 3 | Assumption Visibility | Are implicit assumptions made explicit? |
| 4 | Authoritative Data | Do tables with specific numbers have sources? |
| 5 | Data Consistency | Are there conflicting numbers within the document? |
| 6 | Implicit Causation | Does the claim imply causation without evidence? |
| 7 | Future as Present | Are planned outcomes described as already achieved? |
| 8 | Temporal Coherence | Are dates coherent with each other and the present? |
| 9 | Verifiable Claims | Are specific claims sourced or flagged for verification? |

**Connection to CGD format:**
1. Semantic findings (9 points) determine what issues exist
2. Issues are recorded in CGD state fields (`clarity-status`, `hitl-status`, `hitl-pending-count`)
3. State consistency is enforced by structural rules (C7-C10)

---

## CGD Output Format

Clarity-Gated Documents use `.cgd.md` extension per [docs/CLARITY_GATE_FORMAT_SPEC.md](docs/CLARITY_GATE_FORMAT_SPEC.md).

```yaml
---
clarity-gate-version: 2.1
processed-date: 2026-01-12
processed-by: Claude + Human Review
clarity-status: CLEAR
hitl-status: REVIEWED
hitl-pending-count: 0
points-passed: 1-9
rag-ingestable: true          # computed by validator - do not set manually
document-sha256: 7d865e...
hitl-claims:
  - id: claim-75fb137a
    text: "Revenue projection is $50M"
    value: "$50M"
    source: "Q3 planning doc"
    location: "revenue-projections/1"
    round: B
    confirmed-by: reviewer
    confirmed-date: 2026-01-12
---

# Document Title

[Document body with epistemic markers applied]

---

## HITL Verification Record

### Round A: Derived Data Confirmation
- Claim 1 (source) confirmed

### Round B: True HITL Verification
| # | Claim | Status | Verified By | Date |
|---|-------|--------|-------------|------|
| 1 | [claim] | Confirmed | [name] | [date] |

<!-- CLARITY_GATE_END -->
Clarity Gate: CLEAR | REVIEWED
```

### Claim Completion Status

Status is determined by field **presence**, not an explicit status field:

| State | `confirmed-by` | `confirmed-date` | Meaning |
|-------|----------------|------------------|----------|
| **PENDING** | absent | absent | Awaiting human verification |
| **VERIFIED** | present | present | Human has confirmed |

---

## Go Integration Example

```go
// Check document clarity before vector store ingestion.
func (idx *Indexer) IngestDocument(ctx context.Context, doc Document) error {
    if doc.ClarityStatus != "CLEAR" || doc.HITLStatus != "REVIEWED" {
        return fmt.Errorf("document %s failed clarity gate: status=%s hitl=%s",
            doc.ID, doc.ClarityStatus, doc.HITLStatus)
    }
    if doc.HITLPendingCount > 0 {
        return fmt.Errorf("document %s has %d pending HITL claims",
            doc.ID, doc.HITLPendingCount)
    }
    // Proceed with vector store ingestion...
    return idx.store.Add(ctx, doc.Chunks())
}
```

---

## Severity Levels

| Level | Definition | Action |
|-------|------------|--------|
| **CRITICAL** | LLM will likely treat hypothesis as fact | Must fix before use |
| **WARNING** | LLM might misinterpret | Should fix |
| **TEMPORAL** | Date/time inconsistency detected | Verify and update |
| **VERIFIABLE** | Specific claim that could be fact-checked | Route to HITL or external search |

---

## Quick Scan Checklist

| Pattern | Action |
|---------|--------|
| Specific percentages (89%, 73%) | Add source or mark as estimate |
| Comparison tables | Add "PROJECTED" header |
| "Achieves", "delivers", "provides" | Use "designed to", "intended to" if not validated |
| "100%" anything | Almost always needs qualification |
| "$X.XX" or "~$X" (pricing) | Flag for external verification |
| Competitor capability claims | Flag for external verification |

---

## Output Format

After running Clarity Gate, report:

```
## Clarity Gate Results

**Document:** [filename]
**Issues Found:** [number]

### Critical (will cause hallucination)
- [issue + location + fix]

### Warning (could cause equivocation)
- [issue + location + fix]

### Externally Verifiable Claims
| # | Claim | Type | Suggested Verification |
|---|-------|------|------------------------|
| 1 | [claim] | Pricing | [where to verify] |

---

## Round A: Derived Data Confirmation
- [claim] ([source])

Reply "confirmed" or flag any I misread.

---

## Round B: HITL Verification Required
| # | Claim | Why HITL Needed | Human Confirms |
|---|-------|-----------------|----------------|
| 1 | [claim] | [reason] | [ ] True / [ ] False |

---

**Verdict:** PENDING CONFIRMATION
```

---

## Bundled Scripts

Reference implementations for deterministic computations per FORMAT_SPEC.

> **Requires Python 3.6+** to run.

- `scripts/claim_id.py` — Computes stable, hash-based claim IDs for HITL tracking (per FORMAT_SPEC section 1.3.4)
- `scripts/document_hash.py` — Computes document SHA-256 hash excluding the `document-sha256` line itself (per FORMAT_SPEC section 2.2)

---

## What This Skill Does NOT Do

- Does not classify document types
- Does not restructure documents
- Does not add deep links or references
- Does not evaluate writing quality
- **Does not check factual accuracy autonomously** (requires HITL)
