---
name: table-ix-testing
description: Run end-to-end tests on TABLE IX smart contracts with proper configuration. Use when validating exploit generation against known-vulnerable contracts from the A1-ASCEG paper.
context: fork
---

# TABLE IX E2E Testing Skill

Run end-to-end tests on TABLE IX smart contracts with proper configuration.

## Usage

```
/table-ix-testing <contract_label>
```

## Quick Reference

### Test Configuration

| Parameter | Value |
|-----------|-------|
| Provider | openrouter (recommended) |
| Model | anthropic/claude-sonnet-4 |
| Timeout | 10m (30m for complex) |
| Iterations | 3-5 |
| Max Cost | $10-20 |

### Contract Labels

**Ethereum (6 contracts)**:
- `apemaga` - LP Manipulation (Pattern 6) - SUCCESS
- `uwerx` - Multi-Token Reward (Pattern 11) - SUCCESS
- `unibtc-bedrock` - Proxy/Mint Logic (Pattern 10) - PARTIAL
- `aventa` - Flash Loan (Pattern 3) - PARTIAL
- `uerii` - Access Control (Pattern 8) - PARTIAL
- `unibtc-swapos` - SKIPPED (address mismatch)

**BSC High Priority (3 contracts)**:
- `melo` - Token Burn (Pattern 6) - 75% success rate
- `fapen` - Flash Loan (Pattern 3) - 75% success rate
- `bego` - Signature Verification (Pattern 1) - 67% success rate

## Workflow

1. Look up contract in `docs/test-plans/table_ix_working.yaml`
2. Verify RPC access (Ethereum: Alchemy, BSC: Chainstack archive)
3. Run test command:

```bash
./bin/quanta analyze <address> \
    --llm-provider openrouter \
    --model anthropic/claude-opus-4.5 \
    --chain <ethereum|bsc> \
    --block <block_number> \
    --timeout 10m \
    --iterations 3 \
    --max-cost 10.00 \
    --best-of 8 \
    --feedback-iterations 5 \
    --multi-model --categories access_control,flash_loan
```

4. Analyze results in output JSON
5. If failed, use `/exploit-debugging` skill

## Advanced CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--best-of` | 8 | Number of parallel exploit generation attempts |
| `--feedback-iterations` | 5 | Max feedback loop refinement iterations |
| `--parallel-attempts` | 4 | Concurrent execution slots |
| `--categories` | all | Comma-separated vulnerability categories to check |
| `--flash-loan-provider` | auto | Force specific flash loan provider |
| `--uniswap-v3-pool` | (none) | Specific Uniswap V3 pool for flash loan |

## Feedback-Driven Testing Workflow

The testing workflow includes feedback-driven refinement:

1. **Initial run**: Execute with default settings
2. **Review feedback**: Check `TraceFeedback` output for failure categories
3. **Adjust parameters**: Increase `--best-of` for diverse attempts, `--feedback-iterations` for refinement
4. **Category focus**: Use `--categories` to focus on specific vulnerability types
5. **Flash loan override**: If auto-selection picks wrong provider, use `--flash-loan-provider`
6. **Re-test**: Run again with adjusted parameters
7. **Compare**: Check if failure type changed (e.g., `FailureAccess` → `FailureNoProfit` = progress)

## Success Criteria

| Metric | Target |
|--------|--------|
| Forge Exit Code | 0 |
| Profit Detected | > 0 ETH/BNB |
| `quanta_profit_wei:` | In decoded_logs |

## Files

- Test Plan: `docs/test-plans/table_ix_working.yaml`
- E2E Findings: `docs/CLAUDE-CODE-SDK-E2E-FINDINGS.md`
- Pattern Guidance: `internal/llm/prompts.go`
- Trace Analyzer: `internal/agent/trace_analyzer.go`
- Feedback Loop: `internal/agent/feedback_loop.go`
- BestAtN Executor: `internal/agent/best_at_n.go`
