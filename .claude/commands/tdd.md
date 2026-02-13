---
description: Enforce test-driven development workflow. Scaffold interfaces, generate tests FIRST, then implement minimal code to pass. Ensure 80%+ coverage.
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Edit
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
skills:
  - tdd-workflow
  - testing-patterns
  - go-coding-standards
---

# TDD Command

This command enforces test-driven development methodology for Go code.

**🚨 CRITICAL — TEST BEHAVIOR, NOT IMPLEMENTATION!** Tests should read like business requirements documentation and remain valid even if the implementation changes completely.

## What This Command Does

1. **Scaffold Interfaces** - Define types/interfaces first
2. **Generate Tests First** - Write failing tests (RED)
3. **Implement Minimal Code** - Write just enough to pass (GREEN)
4. **Refactor** - Improve code while keeping tests green (REFACTOR)
5. **Verify Coverage** - Ensure 80%+ test coverage

## When to Use

Use `/tdd` when:
- Implementing new features
- Adding new functions/components
- Fixing bugs (write test that reproduces bug first)
- Refactoring existing code
- Building critical business logic

## Required Standards

Follow the coding guidelines in `docs/CODING_GUIDELINES.md`:
- Table-driven tests for all code
- 80%+ coverage target
- Tests co-located with source (*_test.go)

Reference patterns in:
- `docs/examples/patterns/testing.md` - Test structure patterns
- `docs/examples/patterns/mocking.md` - Mock patterns with Mockery

## TDD Cycle

```
RED → GREEN → REFACTOR → REPEAT

RED:      Write a failing test
GREEN:    Write minimal code to pass
REFACTOR: Improve code, keep tests passing
REPEAT:   Next feature/scenario
```

## Example Usage

```
User: /tdd I need a function to calculate market liquidity score

Agent:
# TDD Session: Market Liquidity Score Calculator

## Step 1: Define Interface (SCAFFOLD)
```

```go
// internal/liquidity/calculator.go
package liquidity

import "time"

// MarketData represents market metrics for liquidity calculation.
type MarketData struct {
    TotalVolume   float64
    BidAskSpread  float64
    ActiveTraders int
    LastTradeTime time.Time
}

// Calculator calculates liquidity scores for markets.
type Calculator interface {
    CalculateScore(market MarketData) float64
}

// calculator implements Calculator interface.
type calculator struct{}

// Compile-time interface check
var _ Calculator = (*calculator)(nil)

// NewCalculator creates a new liquidity calculator.
func NewCalculator() Calculator {
    return &calculator{}
}

// CalculateScore returns a liquidity score from 0-100.
func (c *calculator) CalculateScore(market MarketData) float64 {
    // TODO: Implementation
    panic("not implemented")
}
```

## Step 2: Write Failing Test (RED)

```go
// internal/liquidity/calculator_test.go
package liquidity_test

import (
    "testing"
    "time"

    "github.com/riddopic/quanta/internal/liquidity"
    "github.com/stretchr/testify/assert"
)

func TestCalculateScore(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        market   liquidity.MarketData
        wantMin  float64
        wantMax  float64
    }{
        {
            name: "high liquidity market returns high score",
            market: liquidity.MarketData{
                TotalVolume:   100000,
                BidAskSpread:  0.01,
                ActiveTraders: 500,
                LastTradeTime: time.Now(),
            },
            wantMin: 80,
            wantMax: 100,
        },
        {
            name: "low liquidity market returns low score",
            market: liquidity.MarketData{
                TotalVolume:   100,
                BidAskSpread:  0.5,
                ActiveTraders: 2,
                LastTradeTime: time.Now().Add(-24 * time.Hour),
            },
            wantMin: 0,
            wantMax: 30,
        },
        {
            name: "zero volume market returns zero score",
            market: liquidity.MarketData{
                TotalVolume:   0,
                BidAskSpread:  0,
                ActiveTraders: 0,
                LastTradeTime: time.Now(),
            },
            wantMin: 0,
            wantMax: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            calc := liquidity.NewCalculator()
            got := calc.CalculateScore(tt.market)

            assert.GreaterOrEqual(t, got, tt.wantMin)
            assert.LessOrEqual(t, got, tt.wantMax)
        })
    }
}
```

## Step 3: Run Tests - Verify FAIL

```bash
make test

--- FAIL: TestCalculateScore (0.00s)
    --- FAIL: TestCalculateScore/high_liquidity_market_returns_high_score (0.00s)
        panic: not implemented

FAIL
```

✅ Tests fail as expected. Ready to implement.

## Step 4: Implement Minimal Code (GREEN)

```go
// CalculateScore returns a liquidity score from 0-100.
func (c *calculator) CalculateScore(market MarketData) float64 {
    if market.TotalVolume == 0 {
        return 0
    }

    // Calculate component scores (0-100 scale)
    volumeScore := min(market.TotalVolume/1000, 100)
    spreadScore := max(100-(market.BidAskSpread*1000), 0)
    traderScore := min(float64(market.ActiveTraders)/10, 100)

    // Recent activity bonus
    hoursSinceLastTrade := time.Since(market.LastTradeTime).Hours()
    recencyScore := max(100-(hoursSinceLastTrade*10), 0)

    // Weighted average
    score := volumeScore*0.4 + spreadScore*0.3 + traderScore*0.2 + recencyScore*0.1

    return min(max(score, 0), 100) // Clamp to 0-100
}
```

## Step 5: Run Tests - Verify PASS

```bash
make test

=== RUN   TestCalculateScore
=== RUN   TestCalculateScore/high_liquidity_market_returns_high_score
=== RUN   TestCalculateScore/low_liquidity_market_returns_low_score
=== RUN   TestCalculateScore/zero_volume_market_returns_zero_score
--- PASS: TestCalculateScore (0.00s)

PASS
```

✅ All tests passing!

## Step 6: Check Coverage

```bash
make coverage

File                    | Coverage
------------------------|----------
internal/liquidity/...  | 100%

Coverage: 100% ✅ (Target: 80%)
```

✅ TDD session complete!

## Commands

| Command | Purpose |
|---------|---------|
| `make test` | Run unit tests |
| `make test-race` | Run tests with race detector |
| `make watch` | Auto-run tests on file changes |
| `make coverage` | Generate HTML coverage report |

## TDD Best Practices

**DO:**
- ✅ Write the test FIRST, before any implementation
- ✅ Run tests and verify they FAIL before implementing
- ✅ Write minimal code to make tests pass
- ✅ Refactor only after tests are green
- ✅ Add edge cases and error scenarios
- ✅ Use table-driven tests for multiple scenarios
- ✅ Aim for 80%+ coverage (100% for critical code)

**DON'T:**
- ❌ Write implementation before tests
- ❌ Skip running tests after each change
- ❌ Write too much code at once
- ❌ Ignore failing tests
- ❌ Test implementation details (test behavior)
- ❌ Mock everything (prefer integration tests where appropriate)

## Coverage Requirements

- **80% minimum** for all code
- **100% required** for:
  - Financial calculations
  - Authentication logic
  - Security-critical code
  - Core business logic

## Important Notes

**MANDATORY**: Tests must be written BEFORE implementation. The TDD cycle is:

1. **RED** - Write failing test
2. **GREEN** - Implement to pass
3. **REFACTOR** - Improve code

Never skip the RED phase. Never write code before tests.

## Integration with Other Commands

- Use `/tdd` to implement with tests
- Use `/fix-tests` if tests are failing
- Use `/fix-linting` if linting errors occur
- Use `/test-coverage` to verify coverage
- Use `/review-staged-unstaged` to review implementation

## Related Skills

This command loads these skills:
- `tdd-workflow` - TDD methodology
- `testing-patterns` - Go test patterns
- `go-coding-standards` - Go idioms

## Related Commands

For comprehensive testing beyond the TDD cycle:
- `make integration` - Run integration tests (requires Docker, uses `-tags=integration`)
- `make test-race-full` - Run all tests with race detector (no -short)
