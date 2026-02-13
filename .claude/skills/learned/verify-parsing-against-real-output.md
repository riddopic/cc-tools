# Verify Parsing Fixes Against Real Runtime Output

**Extracted:** 2026-02-09
**Context:** When fixing JSON/data parsing bugs, always verify the fix works against actual runtime data, not just test fixtures

## Problem

A parsing fix can pass all unit tests but fail in production because the test fixtures don't match the real data structure. This happened twice in the CalledTargetFunctions classification bug:

1. **R5 fix**: Added `extractLastJSONSuffix` to parse mixed log+JSON output. Tests passed. But the parsed JSON didn't contain the expected `exploit_code` field — the real subprocess output uses `attempts[].exploit.code` instead.

2. **Result**: `called_target_functions` stayed `false` for ALL tests despite the parsing fix being correct. The field extraction was wrong, not the JSON parsing.

## Solution

After implementing a parsing fix, add a verification step that inspects real runtime data:

```bash
# 1. Run the system to produce real output
bin/quanta regression baseline create --warm-cache

# 2. Inspect the actual output structure
python3 -c "
import json
with open('out/regression/<test>_<timestamp>.json', 'r') as f:
    content = f.read()
# Find the last JSON block (same logic as Go code)
for i in range(len(content)-1, -1, -1):
    if content[i] == '{' and (i == 0 or content[i-1] in '\n\r'):
        cand = content[i:].strip()
        try:
            obj = json.loads(cand)
            print('Keys:', list(obj.keys()))
            # Drill into nested structures
            if 'attempts' in obj:
                att = obj['attempts'][0]
                print('Attempt keys:', list(att.keys()))
            break
        except: continue
"
```

The key principle: **test fixtures show you what you THINK the data looks like; real output shows you what it ACTUALLY looks like.**

## Checklist

When fixing any parsing/extraction bug:

1. Write the fix with unit tests (TDD as normal)
2. Run the system end-to-end to produce real output
3. Inspect the real output to verify the field paths match your assumptions
4. If they don't match, update the extraction logic AND add a test case using the real structure
5. Run again to confirm the fix works on real data

## When to Use

- After fixing any JSON/data extraction bug
- When a "fixed" bug persists in production but passes in tests
- When parsing subprocess output, API responses, or any external data
- When multiple layers of extraction are involved (parse output, then extract field, then validate)
