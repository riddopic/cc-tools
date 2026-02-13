# Timeout Fixes Unmask Hidden Failures

**Extracted:** 2026-02-09
**Context:** When increasing timeouts resolves killed processes, the newly-completing tests often reveal a different dominant failure mode

## Problem

R5 had 8 tests killed at the 45-minute timeout boundary. These were misclassified as compilation/generic_exploit failures. When R7 extended the timeout to 60 minutes, 6 of those 8 tests completed — but they completed with **compilation failures**, not passes.

The result: timeout dropped from 8→2 (win!) but compilation jumped from 4→10 (surprise!). The net pass count stayed at 7/28 — the timeout fix didn't directly produce more passes, it just let tests run long enough to fail at a more informative stage.

One notable exception: **bebop** ran for 44.5 minutes and passed with 3.56 ETH profit. At the old 45-minute timeout, it would have been killed with ~30 seconds to spare. This single test validates the timeout extension.

## Lesson

When fixing a "coarse" failure mode (timeout, OOM, crash), expect to unmask a "finer" failure mode underneath. Plan for two rounds:

1. **Round N**: Fix the coarse failure (timeout extension, memory increase)
2. **Round N+1**: Address the newly-visible finer failures (compilation errors, wrong-chain addresses)

Don't count the coarse-failure tests as "potential passes" — most will fail at the next stage. Budget your expectations accordingly.

## When to Use

- After extending timeouts or resource limits in batch test runs
- When planning regression improvement rounds
- When estimating the impact of infrastructure fixes vs. prompt/logic fixes
- Any time a "gate" is removed and previously-blocked items flow through
