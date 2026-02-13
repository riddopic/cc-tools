---
description: Analyze test coverage and generate missing tests to achieve 80%+ coverage
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
  - testing-patterns
  - tdd-workflow
  - go-coding-standards
---

# Test Coverage

Analyze test coverage and generate missing tests to achieve 80%+ coverage.

**🚨 CRITICAL — TEST BEHAVIOR, NOT IMPLEMENTATION!** Tests should read like business requirements documentation and remain valid even if the implementation changes completely.

## Required Standards

Follow the coding guidelines in `docs/CODING_GUIDELINES.md`:
- Table-driven tests for all code
- 80%+ coverage target
- Tests co-located with source (*_test.go)

Reference patterns in:
- `docs/examples/patterns/testing.md` - Test structure patterns
- `docs/examples/patterns/mocking.md` - Mock patterns with Mockery

## Execution Steps

### 1. Run Tests with Coverage

```bash
# Generate coverage profile
make coverage

# Or manually:
go test -coverprofile=coverage.out ./...
```

### 2. Analyze Coverage Report

```bash
# View function-level coverage
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open HTML report
open coverage.html
```

### 3. Identify Under-Covered Files

Look for files below 80% coverage threshold:

```bash
# List files with coverage below 80%
go tool cover -func=coverage.out | grep -v "100.0%" | awk '$3+0 < 80'
```

### 4. Analyze Untested Code Paths

For each under-covered file:

1. Read the source file
2. Identify untested functions and branches
3. Determine test scenarios needed:
   - Happy path scenarios
   - Error handling paths
   - Edge cases (nil, empty, max values)
   - Boundary conditions

### 5. Generate Missing Tests

Use table-driven test patterns:

```go
func TestFunctionName(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "happy path with valid input",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "error on invalid input",
            input:   invalidInput,
            wantErr: true,
        },
        {
            name:  "edge case with empty input",
            input: emptyInput,
            want:  defaultOutput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            got, err := FunctionName(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 6. Verify New Tests Pass

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Regenerate coverage
make coverage

# For comprehensive verification (if applicable):
make integration        # Integration tests (requires Docker, uses -tags=integration)
```

### 7. Show Before/After Metrics

```bash
# Display summary
go tool cover -func=coverage.out | tail -1
```

## Commands

| Command | Purpose |
|---------|---------|
| `make test` | Run unit tests |
| `make test-race` | Run tests with race detector |
| `make coverage` | Generate HTML coverage report |
| `make integration` | Run integration tests (requires Docker) |
| `make test-race-full` | Run all tests with race detector |

> **Note:** `make coverage` measures unit test coverage. Integration tests run separately via `make integration` and use the `-tags=integration` build tag.

## Coverage Targets

| Code Type | Required Coverage |
|-----------|-------------------|
| Business logic | 80%+ |
| Financial calculations | 100% |
| Authentication | 100% |
| Security-critical | 100% |
| Utility functions | 80%+ |

## Test Types to Include

**Unit Tests** (Function-level):
- Happy path scenarios
- Error conditions
- Edge cases (nil, empty, max values)
- Boundary values

**Integration Tests** (Component-level):
- API endpoints
- Database operations
- External service calls

## Focus Areas

When generating tests, focus on:

1. **Uncovered branches** - if/else paths not executed
2. **Error handling** - ensure errors are tested
3. **Edge cases** - boundary conditions
4. **Concurrency** - race condition testing

## Integration with Other Commands

- Use `/tdd` for test-first development
- Use `/fix-tests` if tests are failing
- Use `/fix-linting` after adding tests
- Use `/review-staged-unstaged` to review test quality
