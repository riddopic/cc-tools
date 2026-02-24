---
name: testing-patterns
description: Apply Go testing patterns and best practices. Use when writing tests, setting up mocks, creating test fixtures, or reviewing test code. Includes table-driven tests and Mockery v3.5 patterns.
---

# Go Testing Patterns

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

## Build Tags

All test commands require the `-tags=testmode` build tag:

```bash
gotestsum --format pkgname -- -tags=testmode -run TestFunctionName ./internal/hooks/...
```

## Test Execution

```bash
task test              # Fast tests (-short, 30s timeout)
task test-race         # Tests with race detector
task coverage          # HTML coverage report
```
