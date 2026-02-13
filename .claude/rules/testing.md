---
paths:
  - "**/*_test.go"
---

# Go Testing Standards

TDD is mandatory. Every line of production code must be written in response to a failing test.

## TDD Workflow

```text
┌─────────────┐
│     RED     │ Write a failing test
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    GREEN    │ Write MINIMUM code to pass
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  REFACTOR   │ Improve if needed (keep tests green)
└─────────────┘
```

**Critical Rules:**

1. No production code without a failing test
2. Write minimum code to pass the test
3. Refactor only when it adds value

## Table-Driven Tests

```go
// ✅ DO: Use table-driven tests for comprehensive coverage
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name        string
        price       float64
        tier        string
        want        float64
        wantErr     bool
        errContains string
    }{
        {
            name:  "premium tier gets 20% discount",
            price: 100.0,
            tier:  "premium",
            want:  80.0,
        },
        {
            name:    "negative price returns error",
            price:   -10.0,
            tier:    "standard",
            wantErr: true,
            errContains: "invalid price",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateDiscount(tt.price, tt.tier)

            if tt.wantErr {
                require.Error(t, err)
                if tt.errContains != "" {
                    assert.Contains(t, err.Error(), tt.errContains)
                }
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}

// ❌ DON'T: Write separate functions for each scenario
func TestCalculateDiscountSuccess(t *testing.T) { /* ... */ }
func TestCalculateDiscountError(t *testing.T) { /* ... */ }
```

## Mock Setup with Mockery v3.5

```go
// ✅ DO: Include mock setup in test tables
func TestService_Process(t *testing.T) {
    tests := []struct {
        name       string
        setupMocks func(*mocks.MockClient)
        input      string
        want       *Result
        wantErr    bool
    }{
        {
            name:  "successful processing",
            input: "test-data",
            setupMocks: func(m *mocks.MockClient) {
                m.EXPECT().Fetch(mock.Anything, "test-data").
                    Return(&Response{Data: "processed"}, nil).Once()
            },
            want: &Result{Value: "processed"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := mocks.NewMockClient(t)
            tt.setupMocks(mockClient)

            svc := NewService(mockClient)
            got, err := svc.Process(context.Background(), tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Test Isolation

```go
// ✅ DO: Use in-memory databases for unit tests
func setupTestStorage(t *testing.T) interfaces.Storage {
    t.Helper()
    storage, err := NewSQLiteStorage(":memory:")
    require.NoError(t, err)

    t.Cleanup(func() {
        _ = storage.Close()
    })

    return storage
}

// ✅ DO: Use t.TempDir() for file-based tests
func TestConfigLoader(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.yaml")
    // Test with temporary file...
}

// ✅ DO: Reset Viper state when testing configuration
t.Run("test config", func(t *testing.T) {
    viper.Reset()
    viper.Set("setting", "value")
    // Test...
})
```

## Test Helpers

```go
// ✅ DO: Use t.Helper() for test utilities
func requireValidUser(t *testing.T, user *User) {
    t.Helper()
    require.NotNil(t, user)
    require.NotEmpty(t, user.ID)
    require.NotEmpty(t, user.Email)
}

// ✅ DO: Use require for fatal assertions, assert for non-fatal
func TestUser(t *testing.T) {
    user, err := GetUser(ctx, "123")
    require.NoError(t, err)        // Fatal if fails
    assert.Equal(t, "123", user.ID) // Continue if fails
}
```

## Running Tests

```bash
task test          # Fast unit tests (-short, 30s timeout)
task test-race     # Tests with race detector
task watch         # Auto-run on file changes (TDD essential!)
task coverage      # Generate HTML coverage report

# Run specific tests
go test -v -run TestName ./internal/agent/
```

## Test Checklist

- [ ] Tests written BEFORE implementation (TDD)
- [ ] Table-driven tests for multiple scenarios
- [ ] Both success and error paths tested
- [ ] Mocks use `EXPECT()` pattern
- [ ] Test isolation (no shared state)
- [ ] `t.Helper()` on utility functions
- [ ] `require` for fatal, `assert` for non-fatal
- [ ] Coverage meets 80% threshold
