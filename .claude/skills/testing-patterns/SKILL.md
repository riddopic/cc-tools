---
name: testing-patterns
description: Apply Go testing patterns and best practices. Use when writing tests, setting up mocks, creating test fixtures, or reviewing test code. Includes table-driven tests and Mockery v3.5 patterns.
---

# Go Testing Patterns for Quanta

## Table-Driven Tests (Standard Pattern)

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       SomeType
        wantOutput  ExpectedType
        wantErr     bool
        errContains string
    }{
        {
            name:       "happy path",
            input:      SomeType{...},
            wantOutput: ExpectedType{...},
            wantErr:    false,
        },
        {
            name:        "error case",
            input:       SomeType{...},
            wantErr:     true,
            errContains: "expected error message",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errContains != "" {
                    assert.Contains(t, err.Error(), tt.errContains)
                }
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.wantOutput, got)
        })
    }
}
```

## Mockery v3.5 Usage

```go
// Create mock with automatic cleanup
executor := mocks.NewMockForgeExecutor(t)

// Type-safe expectation setting
executor.EXPECT().Execute(
    mock.Anything,      // context
    mock.MatchedBy(func(cfg Config) bool {
        return cfg.TestFile != ""
    }),
).Return(&Result{Success: true}, nil).Once()

// AssertExpectations called automatically via t.Cleanup
```

### Table-Driven Tests with Mocks

```go
tests := []struct {
    name       string
    setupMocks func(*mocks.MockForgeExecutor)
    wantResult *Result
    wantErr    error
}{
    {
        name: "successful execution",
        setupMocks: func(m *mocks.MockForgeExecutor) {
            m.EXPECT().Execute(mock.Anything, mock.Anything).
                Return(&Result{Success: true}, nil).Once()
        },
        wantResult: &Result{Success: true},
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        executor := mocks.NewMockForgeExecutor(t)
        tt.setupMocks(executor)
        // ... test logic
    })
}
```

## Mock Commands

```bash
task mocks          # Generate all mocks (regenerates from scratch)
```

## Test Organization

- **Test behaviors, not implementation** - Focus on public APIs
- **Use `t.Helper()`** for test utilities
- **Coverage targets**: ≥80% unit tests, ≥70% integration tests
- **Run with race detector**: `task test-race`

## Security in Tests

**NEVER hardcode secrets!**

```go
// WRONG
const testAPIKey = "sk-test-12345"

// RIGHT
apiKey := "test-" + generateRandomString(32)
```

## Mock-Mode Integration Tests (Orchestrator Pipeline)

When writing integration tests that exercise the `run` or `analyze` CLI commands
through the orchestrator pipeline, use `TEST_MODE=mock` with fast iteration config.
This avoids real LLM/forge calls while testing CLI wiring and orchestrator lifecycle.

### Required Setup

```go
// 1. Enable mock executor/evaluator (bypasses real LLM + forge)
t.Setenv("TEST_MODE", "mock")

// 2. Keep API keys set (required for command validation)
t.Setenv("OPENROUTER_API_KEY", "test-key")

// 3. Use fast iteration flags to avoid Best@N/feedback paths
runCmd.SetArgs([]string{
    "--target", target,
    "--best-of", "1",
    "--feedback-iterations", "1",  // Takes fast executeAnalysis path
    "--format", "text",
    "--model", "openai/o3-mini",
})
```

### Why each setting matters

| Setting | Purpose |
|---------|---------|
| `TEST_MODE=mock` | Uses `mockAgentExecutor` (instant) + `mockSuccessEvaluator` |
| `--best-of 1` | Single attempt, no parallel Best@N |
| `--feedback-iterations 1` | Avoids feedback-driven path (orchestrator.go:836) |
| `IterationDelay = 0` | Set automatically in mock mode, eliminates 2s inter-iteration sleep |

### What it tests
- CLI flag parsing and command wiring
- Orchestrator lifecycle (Start → Analyze → Stop)
- Config building and defaults
- Output formatting (text/JSON)

### What it does NOT test
- Real LLM API calls (use httptest mock servers for that)
- Forge compilation/execution
- Profit evaluation

## Testing RAG and Ensemble Components

### Testing RAG Components
```go
// Use fixed embeddings for deterministic tests
embedder := mocks.NewMockEmbedder(t)
embedder.EXPECT().Embed(mock.Anything, mock.Anything).
    Return([]float32{0.1, 0.2, 0.3}, nil).Once()

// Mock ChromemGo for isolation
store := mocks.NewMockVectorStore(t)
store.EXPECT().Search(mock.Anything, mock.Anything, mock.Anything).
    Return([]vector.Result{{ID: "pattern-1", Score: 0.95}}, nil)

// Verify safety gate catches 100% of violations
gate := safety.NewGate(blocklist)
assert.Error(t, gate.Validate("Uniswap V2 exploit"))
```

### Testing Multi-Model Ensemble
```go
// Mock each provider's LLM client
claude := mocks.NewMockLLMClient(t)
gpt := mocks.NewMockLLMClient(t)
gemini := mocks.NewMockLLMClient(t)

// Test voting with different confidence distributions
claude.EXPECT().Analyze(mock.Anything).Return(&Result{Confidence: 0.9}, nil)
gpt.EXPECT().Analyze(mock.Anything).Return(&Result{Confidence: 0.7}, nil)
gemini.EXPECT().Analyze(mock.Anything).Return(&Result{Confidence: 0.5}, nil)

// Verify fallback when a model fails
gpt.EXPECT().Analyze(mock.Anything).Return(nil, errors.New("rate limited"))
```

### Testing Multi-Actor Coordination
```go
// Use forked blockchain state
fork := testutil.NewForkFixture(t, "mainnet", blockNum)

// Mock executor for isolated testing
executor := mocks.NewMockCoordinatorExecutor(t)
executor.EXPECT().Execute(mock.Anything, mock.Anything).
    Return(&coordinator.Result{State: coordinator.StateComplete}, nil)

// Test state machine transitions
coord := coordinator.New(executor)
assert.Equal(t, coordinator.StateReconnaissance, coord.State())
coord.Advance()
assert.Equal(t, coordinator.StatePlanning, coord.State())

// Verify atomic execution semantics
result, err := coord.ExecuteAtomic(ctx, plan)
assert.NoError(t, err)
assert.True(t, result.AllOrNothing)
```

### Testing Agent Memory
```go
// Mock SQLite store
memStore := mocks.NewMockMemoryStore(t)
memStore.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
memStore.EXPECT().Load(mock.Anything, "session-123").
    Return(&memory.Session{ID: "session-123"}, nil)

// Verify persistence across test boundaries
store := memory.NewSQLiteStore(":memory:")
store.Save(ctx, session)
loaded, _ := store.Load(ctx, session.ID)
assert.Equal(t, session.Data, loaded.Data)
```

## Testing Advanced Pipeline Components

### Testing Structured JSON Plans

```go
// Test PlanParser extracts valid JSON from markdown-wrapped LLM output
func TestPlanParser_Parse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:  "valid JSON in code fence",
            input: "```json\n{\"actors\":[{\"name\":\"attacker\",\"role\":\"attacker\"}],\"steps\":[]}\n```",
        },
        {
            name:    "invalid JSON triggers repair",
            input:   "{actors: [{name: 'attacker'}]}",  // smart quotes, unquoted keys
            wantErr: false, // repair should fix it
        },
    }
    // ... table-driven test body
}

// Test PlanValidator sentinel errors
validator := NewSchemaValidator()
err := validator.Validate(planWithMissingActor)
assert.ErrorIs(t, err, ErrMissingActor)

err = validator.Validate(planWithFlashLoanNoRepay)
assert.ErrorIs(t, err, ErrFlashLoanNoRepay)

// Test FoundryGenerator output assertions
gen := NewFoundryGenerator(tokenRegistry)
result, err := gen.Generate(validPlan, "0x1234...")
assert.NoError(t, err)
assert.Contains(t, result.SolidityCode, "vm.deal")
assert.Contains(t, result.SolidityCode, "vm.startPrank")
assert.NotEmpty(t, result.Imports)
```

### Testing Flash Loan Synthesis

```go
// Test FlashLoanSelector auto-selection
selector := NewFlashLoanSelector(chainConfigs)

// Prefers zero-fee providers
config, err := selector.Select(ctx, "ethereum", "WETH", big.NewInt(1e18))
assert.NoError(t, err)
assert.Equal(t, "dydx", config.Provider) // 0% fee preferred

// Fee-aware profit calculation
assert.Equal(t, expectedNet, config.NetProfit(grossProfit))

// Provider override via CLI flag
config, err = selector.Select(ctx, "ethereum", "WETH", big.NewInt(1e18),
    WithProvider("aave"))
assert.Equal(t, "aave", config.Provider)
```

### Testing Enhanced RAG

```go
// Test SolidityParser element extraction
parser := NewSolidityParser()
elements, err := parser.Parse(soliditySource)
assert.NoError(t, err)
assert.True(t, len(elements) > 0)

// Verify element types
funcElements := filterByType(elements, "function")
assert.NotEmpty(t, funcElements)
assert.NotEmpty(t, funcElements[0].Body)
assert.NotEmpty(t, funcElements[0].Signature)

// Test Chunker overlap verification
chunker := NewChunker(512, 64)
chunks := chunker.ChunkText(longText)
for i := 1; i < len(chunks); i++ {
    // Verify overlap: end of chunk[i-1] overlaps start of chunk[i]
    prevEnd := chunks[i-1].Text[len(chunks[i-1].Text)-64*4:]
    nextStart := chunks[i].Text[:64*4]
    assert.Contains(t, nextStart, prevEnd[:20]) // partial overlap check
}
```

### Testing Category-Specific Ensemble

```go
// Test two-stage voting
voter := NewEnsembleVoter(globalWeights)
voter.SetCategoryWeights(categoryCalibrations)

result := voter.Vote(modelResults)
assert.NotEmpty(t, result.TopCategory)
assert.Greater(t, result.CategoryConfidence, 0.0)

// Test category weight application
result := voter.Vote([]ModelResult{
    {Model: "claude", Confidence: 0.9, Category: CategoryAccessControl},
    {Model: "gpt4o", Confidence: 0.8, Category: CategoryAccessControl},
})
assert.Equal(t, CategoryAccessControl, result.TopCategory)
assert.Greater(t, result.CategoryConfidence, 0.7) // passes confidence gate

// Test VoteResult extensions
assert.NotZero(t, result.TopCategory)
assert.NotZero(t, result.CategoryConfidence)
```

### Testing Feedback-Driven Exploitation

```go
// Test TraceAnalyzer failure categorization
analyzer := NewTraceAnalyzer()
feedback := analyzer.Analyze(traceOutput)
assert.Equal(t, FailureRevert, feedback.FailureType)
assert.NotEmpty(t, feedback.SuggestedFix)

// Test FeedbackOrchestrator iteration
orch := NewFeedbackOrchestrator(llmClient, forgeExecutor, analyzer,
    WithMaxIterations(3))
result, err := orch.Run(ctx, contract)
assert.NoError(t, err)
assert.LessOrEqual(t, result.Iterations, 3)

// Test BestAtNExecutor parallel attempts
executor := NewBestAtNExecutor(llmClients, forgeExecutor,
    WithAttempts(4), WithConcurrency(2))
result, err := executor.Run(ctx, contract)
assert.NoError(t, err)
assert.LessOrEqual(t, result.AttemptsUsed, 4)
```

## Detailed Patterns

For testing patterns, see [testing.md](../../../docs/examples/patterns/testing.md)
For mocking guide, see [mocking.md](../../../docs/examples/patterns/mocking.md)
