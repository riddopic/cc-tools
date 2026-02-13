---
name: multi-actor-coordination
description: Orchestrate complex multi-transaction exploits with role-based actors. Use when working with multi-actor exploit strategies, flash loan coordination, or transaction bundling.
context: fork
---

# Multi-Actor Coordination Skill

Orchestrate complex multi-transaction exploits requiring multiple addresses or actors working in coordination.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Coordinator | `internal/coordinator/coordinator.go` | State machine orchestration |
| Actors | `internal/coordinator/actors.go` | Planner, Executor, Arbitrageur, Validator |
| State Machine | `internal/coordinator/state.go` | 8-state attack lifecycle |
| Flash Loan Builder | `internal/coordinator/flashloan.go` | Atomic bundle construction |
| Memory Store | `internal/memory/` | Cross-session state persistence |

## State Machine

```
Reconnaissance -> Planning -> Setup -> Execution -> Extraction -> Cleanup -> Complete
                                          |
                                          v (on failure)
                                       Failed
```

### State Descriptions

| State | Description | Exit Condition |
|-------|-------------|----------------|
| Reconnaissance | Gather contract state, identify targets | Target identified |
| Planning | Generate attack strategy | Valid plan produced |
| Setup | Deploy helper contracts, fund actors | Setup complete |
| Execution | Execute attack transactions | Attack executed |
| Extraction | Convert profits to base currency | Profits secured |
| Cleanup | Remove traces, return flash loans | Cleanup complete |
| Complete | Success state | Terminal |
| Failed | Failure state | Terminal |

## Actor Roles

| Actor | Responsibility | Example Actions |
|-------|----------------|-----------------|
| Planner | Generates execution strategy | Analyze contract, identify vulnerability |
| Executor | Submits transactions | Send tx, manage gas, handle reverts |
| Arbitrageur | Optimizes profit extraction | DEX routing, slippage management |
| Validator | Verifies attack success | Check profit, validate state changes |

## Flash Loan Providers

| Provider | Contract | Fee | Max Amount |
|----------|----------|-----|------------|
| Aave V3 | LendingPool | 0.05% | Pool liquidity |
| dYdX | SoloMargin | 0% | Pool liquidity |
| Balancer | Vault | 0% | Pool liquidity |
| Uniswap V3 | Pool | 0.3% (swap) | Pool liquidity |

## CLI Usage

```bash
# Enable multi-actor coordination
quanta analyze 0x1234... --multi-actor

# Combine with RAG for pattern-guided coordination
quanta analyze 0x1234... --multi-actor --rag

# Full feature stack
quanta analyze 0x1234... --multi-actor --rag --multi-model
```

## Coordination Patterns

### Pattern 1: Simple Flash Loan

```
Actor1: Borrow -> Exploit -> Repay (single tx)
```

### Pattern 2: Ownership Transfer (SGETH-style)

```
Actor1: Become owner
Actor2: Call privileged function
Actor1: Extract profits
```

### Pattern 3: Multi-Step with Helper Contract

```
Actor1: Deploy helper
Actor2: Fund helper
Actor1: Execute via helper
Actor2: Withdraw profits
```

## Implementation Example

```go
// Create coordinator with actors
coord := coordinator.New(
    coordinator.WithPlanner(planner),
    coordinator.WithExecutor(executor),
    coordinator.WithArbitrageur(arb),
    coordinator.WithValidator(validator),
)

// Execute coordinated attack
plan, err := coord.Plan(ctx, contract)
if err != nil {
    return fmt.Errorf("planning failed: %w", err)
}

result, err := coord.Execute(ctx, plan)
if err != nil {
    return fmt.Errorf("execution failed: %w", err)
}

// Validate and extract
if result.Success {
    profit, _ := coord.Extract(ctx, result)
    log.Info("extracted profit", zap.String("amount", profit.String()))
}
```

## Testing

### Unit Tests

```go
// Mock executor for isolated testing
executor := mocks.NewMockCoordinatorExecutor(t)
executor.EXPECT().Execute(mock.Anything, mock.Anything).
    Return(&coordinator.Result{State: coordinator.StateComplete}, nil)

// Test state machine transitions
coord := coordinator.New(executor)
assert.Equal(t, coordinator.StateReconnaissance, coord.State())

coord.Advance()
assert.Equal(t, coordinator.StatePlanning, coord.State())
```

### Integration Tests

```go
// Use forked blockchain state
func TestCoordinatorIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    fork := testutil.NewFork(t, "mainnet", 18500000)
    defer fork.Close()

    coord := coordinator.New(
        coordinator.WithFork(fork),
        coordinator.WithTimeout(5 * time.Minute),
    )

    result, err := coord.ExecuteFull(ctx, contract)
    assert.NoError(t, err)
    assert.Equal(t, coordinator.StateComplete, result.FinalState)
}
```

### Atomic Execution Tests

```go
// Verify atomic execution semantics
result, err := coord.ExecuteAtomic(ctx, plan)
assert.NoError(t, err)
assert.True(t, result.AllOrNothing)

// If any step fails, all should revert
executor.EXPECT().Execute(mock.Anything, step2).
    Return(nil, errors.New("revert"))
result, _ := coord.ExecuteAtomic(ctx, plan)
assert.False(t, result.Success)
assert.Equal(t, coordinator.StateFailed, result.FinalState)
```

## Memory Persistence

Multi-actor coordination uses agent memory for cross-session state:

```go
// Save coordination state
memory.Save(ctx, &Session{
    ID:        sessionID,
    Contract:  contractAddr,
    State:     coord.State(),
    Plan:      plan,
    Progress:  coord.Progress(),
})

// Resume from previous session
session, _ := memory.Load(ctx, sessionID)
coord.Restore(session.State, session.Plan)
coord.Resume(ctx)
```

## Structured JSON Attack Plans

The LLM now produces structured JSON `AttackPlan` objects instead of free-form text:

### Core Structs

```go
type AttackPlan struct {
    Actors     []PlanActor  `json:"actors"`
    Steps      []AttackStep `json:"steps"`
    Extraction Extraction   `json:"extraction"`
}

type PlanActor struct {
    Name    string `json:"name"`
    Role    string `json:"role"`     // "attacker", "helper", "flashloan_receiver"
    Address string `json:"address"`  // "" = auto-assigned
}

type AttackStep struct {
    Actor      string `json:"actor"`
    Action     string `json:"action"`    // "call", "deploy", "transfer", "flash_loan"
    Target     string `json:"target"`
    Function   string `json:"function"`
    Args       []any  `json:"args"`
    Value      string `json:"value"`     // ETH value
    DependsOn  int    `json:"depends_on"` // step index, -1 = none
}

type Extraction struct {
    Token  string `json:"token"`
    Method string `json:"method"` // "swap_to_eth", "direct_transfer"
}
```

## Plan Parsing Pipeline

The `PlanParser` interface extracts structured plans from LLM output:

1. **JSON extraction** — Finds JSON block within markdown code fences or raw output
2. **Smart repair** — Fixes common LLM JSON errors: smart quotes → straight quotes, trailing commas, unquoted keys
3. **LLM-based repair fallback** — If JSON is still invalid after smart repair, sends it back to the LLM with error context for correction
4. **Retry** — Up to 3 attempts before failing

```go
type PlanParser interface {
    Parse(ctx context.Context, llmOutput string) (*AttackPlan, error)
}
```

## Plan Validation

The `PlanValidator` interface validates parsed plans before code generation:

```go
type PlanValidator interface {
    Validate(plan *AttackPlan) error
}
```

### SchemaValidator

Uses `go-playground/validator` for struct validation. Checks:

- Required fields present on all actors and steps
- Valid action types: `call`, `deploy`, `transfer`, `flash_loan`
- Flash loan steps must have matching repayment step
- Actor references in steps match defined actors
- Step dependency indices are valid (no circular deps)

### Sentinel Errors

| Error | Meaning |
|-------|---------|
| `ErrInvalidPlan` | Plan failed structural validation |
| `ErrMissingActor` | Step references undefined actor |
| `ErrInvalidAction` | Unknown action type |
| `ErrFlashLoanNoRepay` | Flash loan without repayment step |
| `ErrCircularDependency` | Step dependency cycle detected |

## Foundry Code Generation

The `FoundryGenerator` interface converts validated plans to executable Forge tests:

```go
type FoundryGenerator interface {
    Generate(plan *AttackPlan, contractAddr string) (*GeneratedFoundryTest, error)
}
```

### Generation Pipeline

1. **Actor setup** — `vm.deal` for ETH balances, `vm.startPrank` for caller identity
2. **Step translation** — Each `AttackStep` maps to Solidity call/deploy/transfer
3. **Balance capture** — `balanceBefore`/`balanceAfter` for profit calculation
4. **Token resolution** — `TokenRegistry` maps token symbols to addresses per chain

### GeneratedFoundryTest

```go
type GeneratedFoundryTest struct {
    TestName    string // e.g., "testExploit_AccessControl"
    SolidityCode string
    SetupCode   string
    Imports     []string
}
```

## Flash Loan Selection

The `FlashLoanSelector` interface auto-selects the optimal flash loan provider:

```go
type FlashLoanSelector interface {
    Select(ctx context.Context, chain string, token string, amount *big.Int) (*FlashLoanConfig, error)
}
```

### Auto-Selection Logic

1. Check chain support (e.g., dYdX only on Ethereum mainnet)
2. Verify pool liquidity ≥ requested amount
3. Rank by fee: dYdX (0%) > Balancer (0%) > Aave (0.05%) > Uniswap V3 (0.3%)
4. Fee-aware profit: `net_profit = gross_profit - flash_loan_fee`

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--flash-loan-provider` | auto | Force specific provider: `aave`, `dydx`, `balancer`, `uniswap_v3` |
| `--uniswap-v3-pool` | (none) | Specific Uniswap V3 pool address for flash loan |

```bash
# Auto-select provider
quanta analyze 0x1234... --multi-actor

# Force Aave V3
quanta analyze 0x1234... --multi-actor --flash-loan-provider aave

# Specific Uniswap pool
quanta analyze 0x1234... --multi-actor --flash-loan-provider uniswap_v3 \
  --uniswap-v3-pool 0x8ad5...
```

## Chain Threading Architecture

Explicit parameter passing replaces implicit global state:

- `SharedState.ContractRef` — Carries target contract address through the pipeline
- `SetContract()` propagation — Each pipeline stage receives and forwards the contract reference
- `SkipExtractionPhase` — Flag to skip extraction when testing plan generation only

## Structured Planner Backend

The `StructuredPlannerBackend` interface orchestrates the full planning pipeline:

```go
type StructuredPlannerBackend interface {
    GeneratePlan(ctx context.Context, analysis *VulnAnalysis) (*AttackPlan, error)
}
```

Pipeline: LLM generation → `PlanParser.Parse()` → `PlanValidator.Validate()` → set contract/vulnType on plan.

## Files

| File | Purpose |
|------|---------|
| `internal/coordinator/coordinator.go` | Main coordinator |
| `internal/coordinator/state.go` | State machine |
| `internal/coordinator/actors.go` | Actor definitions |
| `internal/coordinator/flashloan.go` | Flash loan builder |
| `internal/coordinator/actors.go` | Actor definitions + AttackPlan types |
| `internal/coordinator/interfaces.go` | Interface definitions |
| `internal/coordinator/plan_parser.go` | JSON plan parser |
| `internal/coordinator/validator.go` | Plan validation |
| `internal/coordinator/foundry_generator.go` | Forge test code generation |
| `internal/coordinator/structured_planner_backend.go` | Full planning pipeline |
| `internal/memory/store.go` | Memory store interface |
| `internal/memory/sqlite_store.go` | SQLite implementation |

## Related Skills

- `/rag-knowledge-system` - Pattern retrieval for coordination strategies
- `/multi-model-ensemble` - Multi-model for plan validation
- `/testing-patterns` - Testing coordinator components
- `/exploit-debugging` - Debugging failed coordination
