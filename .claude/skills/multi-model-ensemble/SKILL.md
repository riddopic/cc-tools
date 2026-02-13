---
name: multi-model-ensemble
description: Implement multi-model analysis with calibrated weighted voting. Use when working with multi-model LLM calls, ensemble voting, or model calibration.
context: fork
---

# Multi-Model Ensemble Skill

Implement multi-model analysis with calibrated weighted voting for improved vulnerability detection accuracy.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Ensemble Voter | `internal/agent/ensemble_voter.go` | Weighted consensus voting |
| Calibration Harness | `internal/agent/ensemble_calibration.go` | F1-score based weight calibration |
| Multi-Model Client | `internal/llm/multi_model.go` | Parallel model invocation |
| Orchestrator Integration | `internal/agent/orchestrator.go` | Multi-model coordination |

## Weighted Voting Formula

```
confidence = Sum(weight_i * confidence_i) / Sum(weight_i)
```

### Default Weights (from LLMBugScanner research)
| Model | Weight | Rationale |
|-------|--------|-----------|
| Claude Sonnet | 0.35 | Highest accuracy on exploit detection |
| GPT-4o | 0.30 | Strong reasoning, good at edge cases |
| Gemini Pro | 0.25 | Fast, good at pattern matching |
| Gemini Flash | 0.10 | Speed tier, verification only |

## Calibration Workflow

### 1. Run Calibration
```bash
quanta calibrate --benchmark-path ~/VERITE --sample-size 10
```

### 2. Process
1. Run TABLE IX test suite with each model independently
2. Calculate per-model success rate (F1 score)
3. Derive weights: `w_i = F1_i / Sum(F1_scores)`
4. Store in `~/.quanta/calibration.json`

### 3. Calibration Output
```json
{
  "version": "1.0",
  "calibrated_at": "2024-01-15T10:30:00Z",
  "weights": {
    "anthropic/claude-sonnet": 0.38,
    "openai/gpt-4o": 0.32,
    "google/gemini-pro": 0.30
  },
  "metrics": {
    "anthropic/claude-sonnet": {"f1": 0.85, "precision": 0.82, "recall": 0.88},
    "openai/gpt-4o": {"f1": 0.78, "precision": 0.80, "recall": 0.76},
    "google/gemini-pro": {"f1": 0.72, "precision": 0.75, "recall": 0.69}
  }
}
```

## CLI Usage

```bash
# Use multi-model with default models
quanta analyze 0x1234... --multi-model

# Specify custom models
quanta analyze 0x1234... --multi-model \
  --models anthropic/claude-sonnet,openai/gpt-4o,google/gemini-pro

# Run calibration before first use
quanta calibrate --benchmark-path ~/VERITE --sample-size 10

# Check calibration status
quanta doctor
```

## Voting Algorithm

### Consensus Modes

1. **Weighted Average** (default): Uses calibrated weights
2. **Majority Vote**: 2-of-3 agreement required
3. **Unanimous**: All models must agree (high precision, low recall)

### Confidence Aggregation
```go
func (v *EnsembleVoter) Vote(results []ModelResult) *AggregatedResult {
    totalWeight := 0.0
    weightedConfidence := 0.0

    for _, r := range results {
        weight := v.weights[r.Model]
        totalWeight += weight
        weightedConfidence += weight * r.Confidence
    }

    return &AggregatedResult{
        Confidence: weightedConfidence / totalWeight,
        Agreement:  v.calculateAgreement(results),
        Individual: results,
    }
}
```

## Fallback Handling

When a model fails:
1. Log the failure with model name and error
2. Remove model from current voting round
3. Re-normalize weights for remaining models
4. Continue with reduced ensemble
5. Warn if < 2 models remain (degraded mode)

```go
// Fallback example
if err := claude.Analyze(ctx, contract); err != nil {
    log.Warn("claude failed, continuing with remaining models", zap.Error(err))
    voter.ExcludeModel("anthropic/claude-sonnet")
}
```

## Testing

### Unit Tests
```go
// Mock each provider's LLM client
claude := mocks.NewMockLLMClient(t)
gpt := mocks.NewMockLLMClient(t)

// Test voting with different distributions
claude.EXPECT().Analyze(mock.Anything).Return(&Result{Confidence: 0.9}, nil)
gpt.EXPECT().Analyze(mock.Anything).Return(&Result{Confidence: 0.7}, nil)

voter := ensemble.NewVoter(weights)
result := voter.Vote([]ModelResult{claudeResult, gptResult})
assert.InDelta(t, 0.82, result.Confidence, 0.01)

// Test fallback when model fails
gpt.EXPECT().Analyze(mock.Anything).Return(nil, errors.New("rate limited"))
// Voter should continue with remaining models
```

### Integration Tests
```go
// Test full ensemble pipeline
func TestEnsembleIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    config := agent.Config{
        MultiModel: true,
        Models: []string{"anthropic/claude-sonnet", "openai/gpt-4o"},
    }

    result, err := agent.Analyze(ctx, contract, config)
    assert.NoError(t, err)
    assert.True(t, result.Ensemble)
    assert.Len(t, result.ModelResults, 2)
}
```

## Category-Specific Calibration

The ensemble now supports per-category weight calibration using 8 vulnerability categories:

### VulnCategory Types

| Category | Description |
|----------|-------------|
| `CategoryAccessControl` | Public/unprotected privileged functions |
| `CategoryReentrancy` | Cross-function or cross-contract reentrancy |
| `CategoryPriceManipulation` | Oracle or DEX price manipulation |
| `CategoryFlashLoan` | Flash loan-dependent attacks |
| `CategoryLogicError` | Business logic flaws |
| `CategoryFrontRunning` | MEV / sandwich / front-running |
| `CategoryIntegerOverflow` | Arithmetic overflow/underflow |
| `CategoryOther` | Unclassified vulnerabilities |

### CategoryCalibration Struct

```go
type CategoryCalibration struct {
    Category    VulnCategory
    Weights     map[string]float64 // model -> weight for this category
    F1Score     float64
    SampleSize  int
}
```

- Minimum sample threshold: categories with < 5 samples fall back to global weights
- Calibration output includes `category_weights` map and `calibrated_at` timestamp

### Calibration Output (Extended)

```json
{
  "version": "2.0",
  "calibrated_at": "2024-06-15T10:30:00Z",
  "weights": { ... },
  "category_weights": {
    "access_control": {
      "anthropic/claude-sonnet": 0.45,
      "openai/gpt-4o": 0.35,
      "google/gemini-pro": 0.20
    },
    "reentrancy": {
      "anthropic/claude-sonnet": 0.30,
      "openai/gpt-4o": 0.40,
      "google/gemini-pro": 0.30
    }
  },
  "metrics": { ... }
}
```

## Two-Stage Voting

Voting now proceeds in two stages:

1. **Stage 1 — Global Vote**: Standard weighted average using global calibrated weights
2. **Stage 2 — Category Refinement**: If global confidence > 0.7, re-vote using category-specific weights for the top-predicted category

### VoteResult Extensions

| Field | Type | Description |
|-------|------|-------------|
| `TopCategory` | `VulnCategory` | Highest-confidence vulnerability category |
| `CategoryConfidence` | `float64` | Confidence for the top category after Stage 2 |

### SetCategoryWeights Method

```go
// Apply category-specific calibration
voter.SetCategoryWeights(categoryCalibrations)

// Vote returns extended result
result := voter.Vote(modelResults)
fmt.Println(result.TopCategory)        // "access_control"
fmt.Println(result.CategoryConfidence) // 0.92
```

## CLI Category Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--categories` | all | Comma-separated vulnerability categories to check (e.g., `access_control,reentrancy`) |

```bash
# Focus on specific categories
quanta analyze 0x1234... --multi-model --categories access_control,flash_loan
```

## Files

| File | Purpose |
|------|---------|
| `internal/agent/ensemble_voter.go` | Voting algorithm |
| `internal/agent/ensemble_voter_test.go` | Voter unit tests |
| `internal/agent/ensemble_calibration.go` | Calibration harness |
| `internal/agent/ensemble_calibration_test.go` | Calibration tests |
| `internal/agent/calibration_categories.go` | Category-specific calibration |
| `internal/llm/multi_model.go` | Parallel model invocation |
| `cmd/calibrate.go` | CLI calibrate command |

## Related Skills

- `/testing-patterns` - Testing patterns for ensemble components
- `/rag-knowledge-system` - RAG integration with ensemble
- `/tdd-workflow` - TDD approach for ensemble development
