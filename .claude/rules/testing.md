---
paths:
  - "**/*_test.go"
---

# Go Testing Standards

TDD is mandatory. Every line of production code must be written in response to a failing test.

## Running Tests

```bash
task test          # Fast unit tests (-short, 30s timeout)
task test-race     # Tests with race detector
task watch         # Auto-run on file changes (TDD essential!)
task coverage      # Generate HTML coverage report
```

All test commands require the `-tags=testmode` build tag. Tests use `gotestsum` (not raw `go test`).

Run a single test:

```bash
gotestsum --format pkgname -- -tags=testmode -run TestFunctionName ./internal/hooks/...
```

## Mock Setup with Mockery v3.5

```bash
task mocks              # Regenerate all mocks
```

Mocks live in `internal/{package}/mocks/`. Config: `.mockery.yml` at project root.

Mock pattern in tests:

```go
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

## Test Checklist

- [ ] Tests written BEFORE implementation (TDD)
- [ ] Table-driven tests for multiple scenarios
- [ ] Both success and error paths tested
- [ ] Mocks use `EXPECT()` pattern
- [ ] Test isolation (no shared state)
- [ ] `t.Helper()` on utility functions
- [ ] `require` for fatal, `assert` for non-fatal
- [ ] Coverage meets 80% threshold
