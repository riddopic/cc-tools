# Validate Metrics Before Optimizing Based on Them

**Extracted:** 2026-02-09
**Context:** When regression run metrics drive optimization decisions, always verify the classification code is correct before trusting the numbers

## Problem

R5 regression summary reported: compilation=6, generic_exploit=8, timeout=1, unknown=4. This led to prioritizing generic_exploit fixes (8 tests). But the classification was **massively wrong**:

- **Actual**: timeout=8, generic_exploit=6, compilation=4, source_code=1
- 8 "signal: killed" timeouts were misclassified as 6 compilation + 2 generic_exploit
- The 1 reported "timeout" was a false positive from `--attempt-timeout` flag substring match
- 4 "unknown" tests were really compilation failures (compiled=false, empty error string)

Three classification bugs conspired:

1. "signal: killed" check existed in code but wasn't in the R5 binary
2. "no source code found" missing from source_code patterns (only "source code not found")
3. `!compiled && fail && empty error` fell through to unknown instead of compilation

## Solution

Before optimizing based on failure category counts:

1. **Spot-check 3-5 individual test results** against their reported category
2. **Read the classification function** (`CategorizeFailure`) and trace each branch
3. **Check for "impossible" categories** (e.g., 0 timeouts when tests run 45+ minutes)
4. **Verify the binary version** matches the code you're reading

```bash
# Quick sanity check: do durations match categories?
python3 -c "
import json, glob
for f in sorted(glob.glob('out/regression/summary_*.json'))[-1:]:
    data = json.load(open(f))
    for r in data['results']:
        if r['status'] != 'pass':
            name = r['test']['name']
            dur = r.get('duration_seconds', 0)
            cat = r.get('failure_category', '?')
            err = r.get('error', '')[:60]
            print(f'{name:20s} {cat:20s} {dur:6.0f}s err={err}')
"
```

If a test ran for exactly the timeout duration but isn't categorized as timeout, the classification is wrong.

## When to Use

- Before planning any round of regression improvements
- When failure category counts seem surprising or don't match intuition
- When "unknown" category has more than 1-2 tests (usually means classification gaps)
- After any binary rebuild, before trusting old run summaries
