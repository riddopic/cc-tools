---
name: regression-baseline-tracking
description: Track regression performance over time using baseline management. Use when validating improvements to exploit generation, before merging significant changes, or when running regression comparisons.
---

# Regression Baseline Tracking

Track exploit generation performance over time using statistical comparison.

## Quick Reference

| Action | Command |
| -------- | --------- |
| Create baseline | `make gt-baseline` |
| Update baseline | `make gt-update` |
| Compare vs baseline | `make gt-check` |
| Clean outputs | `make gt-clean` |

## When to Use

### Create Baseline (`make gt-baseline`)

- After significant improvements to exploit generation
- When establishing a new performance benchmark
- After fixing critical bugs affecting success rate

### Update Baseline (`make gt-update`)

- When current results should become the new standard
- After validated improvements are merged to main
- When baseline is outdated (new contracts added)

### Compare (`make gt-check`)

- Before merging changes affecting exploit generation
- During PR review for LLM prompt changes
- To validate feedback loop improvements
- **Requires** a prior run in `out/regression/` — run `make gt-baseline` first

## Workflow

### Tracking Improvements

1. Run a baseline to generate output:

   ```bash
   make gt-baseline
   ```

2. Make improvements to code (prompts, feedback loops, etc.)

3. Run another baseline and compare against checked-in baseline:

   ```bash
   make gt-baseline
   make gt-check
   ```

4. If improvement is significant, update baseline:

   ```bash
   make gt-update
   ```

5. Commit the updated baseline:

   ```bash
   git add baselines/ground-truth-latest.json
   git commit -m "Update ground truth baseline"
   ```

### Before Merging to Main

1. Ensure baseline is current on main branch
2. Run `make gt-check` on your feature branch
3. Verify no statistically significant regression
4. If regression detected, investigate before merging

## Interpreting Results

### Statistical Test

The comparison uses Wilcoxon signed-rank test on paired contract scores.

| Metric | Interpretation |
| -------- | ---------------- |
| **p-value < 0.05** | Statistically significant difference |
| **p-value ≥ 0.05** | No significant difference detected |
| **p-value = -1.0** | Test not performed (< 6 paired contracts) |

### Direction

| Direction | Meaning |
| ----------- | --------- |
| `improvement` | Candidate scores higher (good) |
| `regression` | Candidate scores lower (bad) |
| `neutral` | No significant difference |

### Minimum Samples

At least **6 paired contracts** are required for the statistical test.
Fewer samples will show `p-value: -1.0` and `direction: neutral`.

## Key Files

| File | Purpose |
| ------ | --------- |
| `baselines/ground-truth-latest.json` | Current checked-in baseline |
| `out/regression/summary_*.json` | Run outputs (not committed) |
| `internal/regression/comparison.go` | Comparison logic |
| `internal/regression/runner.go` | Test runner |

## CLI Commands

Direct CLI usage (equivalent to make targets):

```bash
# Create baseline candidate
quanta regression baseline create

# Update baseline from latest run
quanta regression baseline update

# Compare against baseline
quanta regression compare
```
