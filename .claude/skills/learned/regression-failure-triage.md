# Regression Failure Triage Methodology

**Extracted:** 2026-02-08
**Context:** Analyzing regression run results to identify the highest-impact improvement opportunities

## Problem

After a regression run, the summary JSON contains aggregated failure counts, but understanding what to fix requires drilling into per-test results and tracing the classification logic back to code. Without a systematic approach, effort gets directed at the wrong failure bucket.

## Solution

### Step 1: Read the Summary Failure Analysis

```bash
python3 -c "
import json
with open('out/regression/summary_*.json') as f:
    data = json.load(f)
print(json.dumps(data['failure_analysis'], indent=2))
"
```

Key buckets (from `internal/regression/types.go:CategorizeFailure()`):

- **generic_exploit**: Compiled, ran, 0 profit, did NOT call target functions (DEX-only code)
- **compilation**: Code didn't compile
- **exploit_no_profit**: Compiled, called target functions, but 0 profit
- **timeout**: Process killed (signal: killed)
- **source_code**: Couldn't fetch contract source
- **unknown**: Uncategorized errors

**FIXED (2026-02-09)**: `ValidateSubprocessTargetCalls` now uses `extractLastJSONLine` to parse mixed log+JSON output. Classification is accurate — see `subprocess-json-extraction.md` for the pattern.

### Step 2: Compare Against Baseline

```bash
bin/quanta regression compare
```

Look for:

- Statistical significance (p < 0.05)
- Which tests flipped pass/fail (stochastic variance vs real regression)
- Whether failure categories shifted (e.g., compilation up, generic down)

### Step 3: Per-Test Drill-Down

```bash
python3 -c "
import json
with open('out/regression/summary_*.json') as f:
    data = json.load(f)
for r in data['results']:
    name = r['test']['name']
    status = r['status']
    compiled = r.get('compiled', False)
    err = r.get('error', '')[:80]
    print(f'{name:20s} {status:5s} compiled={compiled} err={err}')
"
```

### Step 4: Identify Highest-Impact Bucket

Prioritize by count AND fixability:

1. **generic_exploit** (8/19 in R5 run) = LLM generates DEX-only code, never calls target. Fix: `writeGenericExploitGuidance` now injects target address + affected function names into refinement prompt.
2. **timeout** (6/19) = Process killed at 45min. Fix: reduce iteration count or timeout budget.
3. **compilation** (4/19) = Generated code has type/import errors. Fix: pragma matching, interface extraction.
4. **no_source** (1/19) = Contract source not verified (0xf340). Not fixable.

### Step 5: Verify Mitigations Are Wired

Before adding fixes, check existing mitigations (see `verify-before-implementing` skill):

- PreExtractedInterface → already in gate4.go:101 and feedback_loop.go:948
- ExtractPragma → already in feedback_loop.go:949
- validateTargetCalls → runs on every iteration in feedback_loop.go:474-479

## Classification Logic Reference

From `internal/regression/types.go:CategorizeFailure()`:

```
compiled=false                          → compilation
compiled=true, profit=0, funcs=false    → generic_exploit
compiled=true, profit=0, funcs=true     → exploit_no_profit
signal: killed                          → timeout (or unknown)
source fetch failed                     → source_code
```

## When to Use

- After any `make gt-baseline` run
- Before planning the next round of prompt improvements
- When deciding which failure bucket to target
- Before creating improvement plan items
