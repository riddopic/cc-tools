---
name: a1-asceg-research
description: Deep research workflow for analyzing exploit success rate gaps between Quanta and A1-ASCEG. Use when investigating why exploits fail, comparing methodology with the A1-ASCEG paper, or improving pattern detection.
context: fork
---

# A1-ASCEG Research Skill

Deep research workflow for analyzing exploit success rate gaps between Quanta and A1-ASCEG.

## Usage

```
/a1-asceg-research [focus_area]
```

Focus areas: `gap-analysis`, `pattern-coverage`, `methodology`, `benchmarking`

## Quick Reference

### Current Status

| Metric | Quanta | A1-ASCEG | Gap |
|--------|--------|----------|-----|
| Ethereum Success Rate | 33% (2/6) | 42-63% | -9% to -30% |
| Single Iteration | 60% | 25-51% | +9% to +35% |
| With Refinement | 100% (tested) | 62-88% | +12% to +38% |

### Key Gaps Identified

1. **Contract Variant Mismatch**: Our test addresses may differ from A1-ASCEG targets
2. **Block Selection**: Liquidity may not exist at our test blocks
3. **Pattern Selection**: LLM sometimes picks wrong pattern for contract
4. **Profit Threshold**: A1-ASCEG uses > 0, we use > 0.1 ETH

### Research Resources

- **A1-ASCEG Paper**: arXiv:2507.05558
- **SCONE-bench**: Anthropic's smart contract benchmark
- **DeFiHackLabs**: Real-world PoC collection
- **TABLE IX**: 36 contracts, 20 with verified bytecode

## Workflow

### Gap Analysis

1. **Compare Success Rates**:
   - Run batch test on all TABLE IX contracts
   - Calculate success rate by pattern type
   - Identify underperforming patterns
   - When running batch analysis across all TABLE IX contracts, use `recursive-decomposition` to partition contracts into batches for parallel sub-agent processing

2. **Root Cause Analysis**:
   - For each failed contract, use `/exploit-debugging`
   - Categorize failures (variant, liquidity, pattern, output)
   - Prioritize by fix complexity

3. **Document Findings**:
   - Update `docs/CLAUDE-CODE-SDK-E2E-FINDINGS.md`
   - Update `docs/test-plans/table_ix_working.yaml`

### Pattern Coverage

1. **Map A1-ASCEG Patterns**:
   | A1 Pattern | Quanta Pattern | Coverage |
   |------------|----------------|----------|
   | Reentrancy | 1 | Full |
   | Flash Loan | 3 | Full |
   | Access Control | 8 | Full |
   | LP Manipulation | 6 | Full |
   | Oracle | 9 | Partial |
   | Proxy/Mint | 10 | Partial |
   | Multi-Token | 11 | New |

2. **Identify Missing Patterns**:
   - Review A1-ASCEG TABLE IX vulnerability types
   - Check if Quanta detects each type
   - Add missing patterns with `/pattern-management add`

### Methodology Comparison

1. **Time Budget**:
   - A1-ASCEG: 5 hours per contract
   - Quanta: 10-30 minutes
   - Impact: Complex contracts may need more iterations

2. **Initial Funding**:
   - A1-ASCEG: 100 ETH/BNB seeded
   - Quanta: Harness provides 1M ETH + 10M stablecoins
   - Impact: Should be sufficient

3. **Static Analysis**:
   - A1-ASCEG: Slither pre-analysis
   - Quanta: LLM-only analysis
   - Impact: Could miss complex vulnerabilities

### Benchmarking

1. **Reproducibility Test**:
   - For each A1-ASCEG success, verify we can reproduce
   - Document any configuration differences
   - Track cost and time metrics

2. **Success Rate Tracking**:
   ```
   | Session | Contracts | Success | Partial | Failed | Rate |
   |---------|-----------|---------|---------|--------|------|
   | Jan 4 AM | 3 | 2 | 1 | 0 | 67% |
   | Jan 4 PM | 3 | 0 | 3 | 0 | 0% |
   | Total | 6 | 2 | 4 | 0 | 33% |
   ```

## Research Agents

Use these specialized agents for deep research:

- `crypto-smart-contracts-researcher`: Vulnerability patterns, DeFi exploits
- `deep-research-specialist`: Multi-source validation, A1-ASCEG paper analysis
- `exploit-generator`: Generate PoC code for identified vulnerabilities
- `exploit-validator`: Test exploits on forked blockchain state

## Files

- Gap Analysis: `docs/research/A1-ASCEG-GAP-ANALYSIS.md`
- Methodology: `docs/research/A1-ASCEG-METHODOLOGY-RESEARCH.md`
- E2E Findings: `docs/CLAUDE-CODE-SDK-E2E-FINDINGS.md`
- Test Plan: `docs/test-plans/table_ix_working.yaml`
