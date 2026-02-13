# Clarity Gate Procedures

**Version:** 1.0
**Last Updated:** 2026-01-13
**Specification:** CLARITY_GATE_FORMAT_SPEC.md

---

## Purpose

This document describes **HOW to use** Clarity Gate for pre-ingestion verification and CGD/SOT creation. For **WHAT the format is**, see [CLARITY_GATE_FORMAT_SPEC.md](CLARITY_GATE_FORMAT_SPEC.md).

**Key distinction:**
- **Format Specification** (v2.1) — Normative rules for file structure and validation
- **Procedures** (this doc) — Informative guidance on verification workflows

---

## Table of Contents

1. [The 9 Verification Points](#the-9-verification-points)
2. [Verification Hierarchy](#verification-hierarchy)
3. [Two-Round HITL Process](#two-round-hitl-process)
4. [Command Reference](#command-reference)
5. [Practical Workflows](#practical-workflows)
6. [Examples](#examples)

---

## The 9 Verification Points

For the full 9-point checklist with examples and fix guidance, see [ARCHITECTURE.md](ARCHITECTURE.md#the-9-verification-points).

**Summary:** The 9 points guide semantic review of document content. Findings are recorded in CGD state fields (`clarity-status`, `hitl-status`, `hitl-pending-count`), and state consistency is enforced by structural rules (C7-C10 in the spec).

| # | Point | Focus |
|---|-------|-------|
| 1-4 | Epistemic Checks | Hypothesis labeling, uncertainty markers, assumption visibility, authoritative data |
| 5-7 | Data Quality | Consistency, implicit causation, future as present |
| 8-9 | Verification Routing | Temporal coherence, externally verifiable claims |

---

## Verification Hierarchy

```
Claim Extracted --> Does Source of Truth Exist?
                           |
           +---------------+---------------+
           YES                             NO
           |                               |
     Tier 1: Automated              Tier 2: HITL
     Verification                   (Last Resort)
           |                               |
     +-----+-----+                   Human reviews:
     |           |                   - Add markers
   Tier 1A    Tier 1B               - Provide source
   Internal   External              - Reject claim
           |           |                   |
     PASS/BLOCK   PASS/BLOCK        APPROVE/REJECT
```

### Tier 1A: Internal Consistency (Ready Now)

Checks for contradictions *within* a document. No external systems required.

**Capabilities:**

| Check | Description | Status |
|-------|-------------|--------|
| Figure vs. Text | Cross-reference numerical claims | Ready |
| Abstract vs. Body | Verify summary matches content | Ready |
| Table vs. Prose | Ensure counts/lists are consistent | Ready |
| Duplicate values | Flag conflicting repeated claims | Ready |

**Implementation:** The Claude skill handles Tier 1A checks through:
1. Extracting claims from document
2. Cross-referencing numerical values
3. Flagging discrepancies with specific locations

---

### Tier 1B: External Verification (Extension Interface)

For claims verifiable against structured sources. **Users implement connectors.**

**Example Connectors (User-Implemented):**

| Claim Type | Source | Connector |
|------------|--------|-----------|
| "Q3 revenue was $47M" | Financial system | `FinancialDataConnector` |
| "Feature deployed Oct 15" | Git commits | `GitHistoryConnector` |
| "Customer count is 1,247" | CRM | `CRMConnector` |
| "API latency is 50ms" | Monitoring | `MetricsConnector` |

**Honest Limitation:** External verification requires bespoke integration for each data source. This is **not out-of-the-box functionality**. Clarity Gate provides the interface; users provide implementations.

---

### Tier 2: HITL Verification (Last Resort)

When automated verification cannot resolve a claim, it routes to human review.

**Purpose:** Intelligent routing—detect *which specific claims* need human review, and *what kind of review* each needs.

---

## Two-Round HITL Process

### Why Two Rounds?

Different claims need different types of verification:

| Claim Type | What Human Checks | Cognitive Load |
|------------|-------------------|----------------|
| LLM found source, human witnessed | "Did I interpret correctly?" | Low (quick scan) |
| Human's own data | "Is this actually true?" | High (real verification) |
| No source found | "Is this actually true?" | High (real verification) |

Mixing these in one table creates checkbox fatigue—human rubber-stamps everything instead of focusing attention where it matters.

---

### Round A: Derived Data Confirmation

Claims where LLM found a source AND human was present in the session.

**Purpose:** Confirm interpretation, not truth. Human already saw the source.

**Format:** Simple list (lighter visual weight for quick scan)

```markdown
## Round A: Derived Data Confirmation

These claims came from sources found in this session:

- o3 prices cut 80% June 2025 (OpenAI blog)
- Opus 4.5 is $5/$25 (Anthropic pricing page)

Reply "confirmed" or flag any I misread.
```

---

### Round B: True HITL Verification

Claims where:
- No source was found
- Source is human's own data/experiment
- LLM is extrapolating or inferring
- Conflicting sources found

**Purpose:** Verify truth. Human may NOT have seen this or it may not exist.

**Format:** Full table with True/False confirmation

```markdown
## Round B: HITL Verification Required

| # | Claim | Why HITL Needed | Human Confirms |
|---|-------|-----------------|----------------|
| 1 | Benchmark scores (100%, 75%→100%) | Your experiment data | [ ] True / [ ] False |
```

---

### Classification Logic

```
Claim Extracted
      │
      ▼
Was source found in THIS session?
      │
      ├─── YES ────► Was human present/active?
      │                    │
      │              ├─ YES ──► ROUND A (Derived)
      │              │
      │              └─ NO/UNCLEAR ──► ROUND B (True HITL)
      │
      └─── NO ─────► Is this human's own data?
                           │
                     ├─ YES ──► ROUND B with note "your data"
                     │
                     └─ NO ──► ROUND B with note "no source found"
```

**Default behavior:** When uncertain, assign to Round B.

---

### Human Review Options

When a claim is routed to Round B, the human must:

1. **Provide Source of Truth** — Point to authoritative source that was missed
2. **Add Epistemic Markers** — Mark as [PROJECTION], [HYPOTHESIS], [UNVERIFIED]
3. **Reject Claim** — Remove or rewrite the claim entirely

---

## Command Reference

> **Note:** CLI commands below describe the planned external tool interface, not Quanta commands. Currently, use the Claude skill directly for verification.

### verify

**Purpose:** Run verification checks on a document and produce a findings report.

**Usage:**
```bash
clarity-gate verify <input-file>
```

**Output:** Verification report with issues grouped by severity (Critical, Warning, Temporal, Verifiable)

**Does NOT modify the input file.**

---

### process

**Purpose:** Generate a CGD file with inline epistemic markers applied.

**Usage:**
```bash
clarity-gate process <input-file> -o <output-file>
```

**Output:** CGD file (`.cgd.md`) with:
- YAML frontmatter with all required fields
- Epistemic markers applied to problematic claims
- HITL verification record (if applicable)
- End marker: `<!-- CLARITY_GATE_END -->` followed by `Clarity Gate: <status> | <hitl-status>`

---

### promote

**Purpose:** Add `tier:` block to a CGD file, promoting it to SOT (Source of Truth) status.

**Usage:**
```bash
clarity-gate promote <cgd-file> --owner "Team" --version 1.0
```

**Effect:**
- Validates document is CGD (no tier block)
- Checks for existing `## Verified Claims` table or generates one from `hitl-claims`
- Adds `tier:` block to YAML frontmatter
- Recomputes `document-sha256`
- Updates `processed-date` to current date

**When to use:** When a CGD has verified claims ready to become a canonical reference document.

---

### demote

**Purpose:** Remove `tier:` block from a SOT file, demoting it back to CGD status.

**Usage:**
```bash
clarity-gate demote <sot-file> --reason "Superseded by v2"
```

**Effect:**
- Validates document is SOT (has tier block)
- Removes `tier:` block from YAML frontmatter
- Archives `## Verified Claims` section with comment marker for later restoration
- Recomputes `document-sha256`
- Updates `processed-date` to current date

**When to use:** When a SOT is superseded, needs rework, or should no longer be a canonical reference.

---

### apply-hitl

**Purpose:** Apply human responses from HITL verification to update document status.

**Usage:**
```bash
clarity-gate apply-hitl <cgd-file> --responses <responses-file>
```

**Effect:**
- Updates `hitl-claims` with confirmed claims
- Updates `hitl-status` based on responses
- Updates `clarity-status` if all issues resolved
- Recomputes `rag-ingestable`

---

### status

**Purpose:** Show tier and status summary for one or more documents.

**Usage:**
```bash
clarity-gate status <files...>
```

**Output:** Summary table showing document states:

```
File                    Status          HITL Status    RAG Ingestable  Tier
──────────────────────  ──────────────  ─────────────  ──────────────  ────
api-docs.cgd.md         CLEAR           REVIEWED       true            CGD
project-data.cgd.md     CLEAR           REVIEWED       true            SOT
draft.cgd.md            UNCLEAR         PENDING        false           CGD
```

**Use cases:**
- Quick overview of document verification states
- CI/CD checks before ingestion
- Batch status reporting

---

## Practical Workflows

### Workflow 1: New Document Creation

**Goal:** Create a CGD from scratch.

**Steps:**

1. **Write initial content** with uncertainty markers where appropriate
2. **Run verify:**
   ```bash
   clarity-gate verify draft.md
   ```
3. **Review findings** and fix critical issues
4. **Run process:**
   ```bash
   clarity-gate process draft.md -o output.cgd.md
   ```
5. **Complete HITL verification** (Round A and Round B)
6. **Apply HITL responses:**
   ```bash
   clarity-gate apply-hitl output.cgd.md --responses responses.yaml
   ```
7. **Ingest if `rag-ingestable: true`**

---

### Workflow 2: Iterative Refinement

**Goal:** Improve an existing CGD with new information.

**Steps:**

1. **Start with existing CGD** (`CLEAR | REVIEWED`)
2. **Add new content** (introduces unverified claims)
3. **Re-process the document:**
   ```bash
   clarity-gate process existing.cgd.md -o updated.cgd.md
   ```
4. **Complete HITL verification** for new claims
5. **Apply HITL responses:**
   ```bash
   clarity-gate apply-hitl updated.cgd.md --responses responses.yaml
   ```
6. **Optionally promote to SOT** if it should become a canonical reference:
   ```bash
   clarity-gate promote updated.cgd.md --owner "Team" --version 2.0
   ```

---

### Workflow 3: Team Handoff

**Goal:** Document verification state for handoff between team members or LLM sessions.

**Steps:**

1. **Original author creates CGD** with pending claims in `hitl-claims` (status: `UNCLEAR | PENDING`)
2. **Commit to version control**
3. **Reviewer verifies claims:**
   - Confirms each claim via HITL verification
   - Updates `hitl-claims` with `confirmed-by` and `confirmed-date`
   - Resolves all pending claims
4. **When all claims verified**, document becomes `CLEAR | REVIEWED`
5. **Optionally promote to SOT** if it should be a canonical reference:
   ```bash
   clarity-gate promote doc.cgd.md --owner "Team" --version 1.0
   ```
6. **Handoff complete** with full verification trail in `hitl-claims`

---

### Workflow 4: CI/CD Integration

**Goal:** Block ingestion of documents that fail verification.

**Pipeline:**

```yaml
- name: Verify document
  run: clarity-gate verify doc.md --strict

- name: Check RAG ingestability
  run: |
    if [ "$(clarity-gate check-ingestable doc.cgd.md)" = "false" ]; then
      echo "Document not RAG-ingestable"
      exit 1
    fi
```

---

## Examples

### Example 1: Complete Verification Flow

**Input document (draft.md):**
```markdown
# Product Roadmap

Revenue will be $50M by Q4.
The new feature reduces churn.
Our system handles 10,000 requests per second.
```

**Step 1: Run verify**
```bash
clarity-gate verify draft.md
```

**Output:**
```
## Clarity Gate Results

**Document:** draft.md
**Issues Found:** 3

### Critical (will cause hallucination)
- Line 3: "Revenue will be $50M by Q4" — Future state lacks uncertainty marker
  FIX: "Revenue is **projected** to be $50M by Q4"

- Line 4: "The new feature reduces churn" — Unvalidated causal claim
  FIX: "The new feature is **expected** to reduce churn [hypothesis]"

- Line 5: "Our system handles 10,000 requests per second" — Future capability as present
  FIX: "Our system is **designed to** handle 10,000 requests per second"

---

## Round B: HITL Verification Required

| # | Claim | Why HITL Needed | Human Confirms |
|---|-------|-----------------|----------------|
| 1 | Revenue $50M by Q4 | No source found | [ ] True / [ ] False |
| 2 | Feature reduces churn | No source found | [ ] True / [ ] False |
| 3 | 10K requests/second | No source found | [ ] True / [ ] False |

**Verdict:** PENDING CONFIRMATION
```

**Step 2: Human responds**
```yaml
responses:
  - claim_id: 1
    status: confirmed
    note: "From Q3 planning doc, page 7"
  - claim_id: 2
    status: rejected
    note: "Not tested yet, hypothesis only"
  - claim_id: 3
    status: confirmed
    note: "Design spec, not measured performance"
```

**Step 3: Apply HITL**
```bash
clarity-gate apply-hitl output.cgd.md --responses responses.yaml
```

**Final CGD (output.cgd.md):**
```yaml
---
clarity-gate-version: 2.0
processed-date: 2026-01-13
processed-by: Claude + Francesco
clarity-status: CLEAR
hitl-status: REVIEWED
hitl-pending-count: 0
points-passed: 1-9
rag-ingestable: true
document-sha256: [computed-hash]
hitl-claims:
  - text: "Revenue projected to be $50M by Q4"
    round: B
    id: claim-1a2b3c4d
    location: "product-roadmap/1"
    source: "Q3 planning doc, page 7"
    confirmed-date: 2026-01-13
    confirmed-by: Francesco
  - text: "System designed to handle 10K rps"
    round: B
    id: claim-5e6f7g8h
    location: "product-roadmap/3"
    source: "Design spec"
    confirmed-date: 2026-01-13
    confirmed-by: Francesco
---

# Product Roadmap

Revenue is **projected** to be $50M by Q4 [Q3 planning doc, page 7].
The new feature is **expected** to reduce churn *(hypothesis, not yet validated)*.
Our system is **designed to** handle 10,000 requests per second [design spec].

---

## HITL Verification Record

### Round B: True HITL Verification
| # | Claim | Status | Verified By | Date |
|---|-------|--------|-------------|------|
| 1 | Revenue projected at $50M | ✓ Confirmed | Francesco | 2026-01-13 |
| 2 | Feature expected to reduce churn | ✓ Marked as hypothesis | Francesco | 2026-01-13 |
| 3 | System designed for 10K rps | ✓ Confirmed (design spec) | Francesco | 2026-01-13 |

<!-- CLARITY_GATE_END -->
Clarity Gate: CLEAR | REVIEWED
```

---

### Example 2: CGD to SOT Promotion

**Input CGD (project-data.cgd.md):**

A CGD with verified claims, ready to become a canonical Source of Truth.

```yaml
---
clarity-gate-version: 2.0
processed-date: 2026-01-12
processed-by: Claude + Francesco
clarity-status: CLEAR
hitl-status: REVIEWED
hitl-pending-count: 0
points-passed: 1-9
rag-ingestable: true
document-sha256: [computed-hash]
hitl-claims:
  - text: "User count is 1,247"
    round: B
    id: claim-a1b2c3d4
    location: "project-data/1"
    value: "1,247"
    source: "CRM export"
    confirmed-date: 2026-01-12
    confirmed-by: Francesco
  - text: "Revenue was $47M in Q3"
    round: B
    id: claim-e5f6g7h8
    location: "project-data/2"
    value: "$47M"
    source: "Financial system"
    confirmed-date: 2026-01-12
    confirmed-by: Francesco
---

# Project Data

User count is 1,247 [CRM export, Dec 15].
Revenue was $47M in Q3 [Financial system].

<!-- CLARITY_GATE_END -->
Clarity Gate: CLEAR | REVIEWED
```

**Step 1: Promote to SOT**
```bash
clarity-gate promote project-data.cgd.md --owner "Data Team" --version 1.0
```

**Output SOT:**
```yaml
---
clarity-gate-version: 2.0
processed-date: 2026-01-13
processed-by: Claude + Francesco
clarity-status: CLEAR
hitl-status: REVIEWED
hitl-pending-count: 0
points-passed: 1-9
rag-ingestable: true
document-sha256: [computed-hash]
hitl-claims:
  - text: "User count is 1,247"
    round: B
    id: claim-a1b2c3d4
    location: "project-data/1"
    value: "1,247"
    source: "CRM export"
    confirmed-date: 2026-01-12
    confirmed-by: Francesco
  - text: "Revenue was $47M in Q3"
    round: B
    id: claim-e5f6g7h8
    location: "project-data/2"
    value: "$47M"
    source: "Financial system"
    confirmed-date: 2026-01-12
    confirmed-by: Francesco
tier:
  level: SOT
  owner: Data Team
  version: 1.0
  promoted-date: 2026-01-13
  promoted-by: Data Team
---

# Project Data

User count is 1,247 [CRM export, Dec 15].
Revenue was $47M in Q3 [Financial system].

## Verified Claims

| Claim | Value | Source | Verified |
|-------|-------|--------|----------|
| User count is 1,247 | 1,247 | CRM export | 2026-01-12 |
| Revenue was $47M in Q3 | $47M | Financial system | 2026-01-12 |

<!-- CLARITY_GATE_END -->
Clarity Gate: CLEAR | REVIEWED
```

*(Note: `tier:` block added, `## Verified Claims` table auto-generated)*

---

## Quick Scan Checklist

| Pattern | Action |
|---------|--------|
| Specific percentages (89%, 73%) | Add source or mark as estimate |
| Comparison tables | Add "PROJECTED" header |
| "Achieves", "delivers", "provides" | Use "designed to", "intended to" if not validated |
| Checkmarks | Verify these are confirmed |
| "100%" anything | Almost always needs qualification |
| "Last Updated: [date]" | Check against current date |
| Version numbers with dates | Verify chronological order |
| "$X.XX" or "~$X" (pricing) | Flag for external verification |
| "averages", "typically" | Flag for source/citation |
| Competitor capability claims | Flag for external verification |

---

## Critical Limitation

> **Clarity Gate verifies FORM, not TRUTH.**

This system checks whether claims are properly marked as uncertain—it cannot verify if claims are actually true.

### The Risk

An LLM can hallucinate facts INTO a document, then "pass" Clarity Gate by adding source markers to false claims.

Example:
```
FAIL: "Revenue will be $50M"
PASS: "Revenue is projected to be $50M [source: Q3 planning doc]"
```

The second passes Clarity Gate even if the "Q3 planning doc" doesn't exist or says something different.

### The Mitigation

HITL Fact Verification is **MANDATORY** before declaring PASS. The human must:
1. Spot-check that cited sources actually exist
2. Verify cited sources actually support the claims
3. Flag any suspicious attribution patterns

---

## What This Does NOT Do

- Does not classify document types (use Stream Coding for that)
- Does not restructure documents
- Does not add deep links or references
- Does not evaluate writing quality
- **Does not check factual accuracy autonomously** (requires HITL)

---

## Related Documents

- [CLARITY_GATE_FORMAT_SPEC.md](CLARITY_GATE_FORMAT_SPEC.md) — Format specification (normative)
- [ARCHITECTURE.md](ARCHITECTURE.md) — 9-point verification system
- [SKILL.md](../SKILL.md) — Claude skill implementation
