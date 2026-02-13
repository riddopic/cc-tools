# Regression Run Workflow

**Extracted:** 2026-02-08
**Context:** Running full-mode regression baselines for exploit generation validation

## Problem

Every session that runs regression tests reinvents the same sequence of steps, hitting the same pitfalls:

- Forgetting to build to `bin/quanta` (subprocesses use that path, not `./quanta`)
- Background Bash tasks timing out for long-running regression (30-60 min)
- Not using `--warm-cache` and getting Etherscan rate-limit failures
- Using quick mode (BestOf=1) to evaluate prompt changes (too much variance)

## Solution: Standard Regression Run Sequence

```bash
# 1. Build binary (make build outputs to bin/quanta directly)
make build

# 2. Run full-mode regression with nohup (survives tool timeouts)
nohup bin/quanta regression baseline create --warm-cache \
  > /path/to/scratchpad/regression-run.log 2>&1 &
echo "PID: $!"

# 3. Monitor progress
grep "Test passed\|Test failed" /path/to/scratchpad/regression-run.log
grep "Progress" /path/to/scratchpad/regression-run.log | tail -3
ps aux | grep "quanta.*analyze" | grep -v grep | wc -l  # active workers

# 4. Check final results
ls -lt out/regression/summary_*.json | head -1

# 5. Compare against checked-in baseline
bin/quanta regression compare
```

## Key Facts

- `DefaultBinaryPath = "bin/quanta"` — subprocesses run `bin/quanta`, NOT `./quanta`
- `make build` outputs directly to `bin/quanta` — no need for `cp`
- Full mode defaults: BestOf=3, Iterations=3, claude-opus-4.6, claude-code provider
- `--warm-cache` populates `~/.quanta/cache/source/{chainID}/{address}.sol` sequentially
- Base chain tests are always skipped (no RPC URL configured)
- Expected 3 warm-cache warnings: 2 Base chain, 1 unverified contract (0xf340)
- Full run takes 30-60 minutes for 28 tests
- DO NOT use Bash `run_in_background` for regression — it times out; use `nohup` instead
- Quick mode (BestOf=1) has too much stochastic variance for A/B comparisons

## Provider Options

```bash
# Claude (default, best for exploit gen)
bin/quanta regression baseline create --warm-cache

# Gemini
bin/quanta regression baseline create --warm-cache --llm-provider gemini --model gemini-3-pro

# Codex (NOT suitable for exploit gen — 0/28)
bin/quanta regression baseline create --warm-cache --llm-provider codex --model gpt-5.3-codex
```

## Debugging Regression Runs

```bash
# Verbose + raw logs — captures subprocess stderr and LLM conversations
bin/quanta regression baseline create --warm-cache --verbose --raw-logs

# Inspect per-test stderr logs after the run
cat out/regression/testname_*.log | head -50

# Inspect raw LLM conversations
cat out/regression/testname_*_raw.json | jq '.[].content' | head -20

# Debug log level — maximum output from runner and subprocesses
bin/quanta regression baseline create --warm-cache --log-level debug --raw-logs
```

## Reading Results

```bash
# Check individual test profit from subprocess output
grep "max_profit_wei" out/regression/testname_*.json | tail -1

# The last max_profit_wei line is authoritative (multi-iteration output)
```

## When to Use

- Before merging prompt changes to validate no regression
- After fixing infrastructure bugs (source cache, session IDs, etc.)
- When establishing new performance baselines
- Provider comparison runs
