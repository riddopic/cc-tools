# Table-Driven Tests

Table-driven tests allow efficient definition of multiple test cases with different inputs, expected outputs, and assertions. This pattern should be consistently applied across the codebase.

## Basic Structure

- **Test Table Definition**: Define struct with fields for inputs, expected outputs, and error conditions
- **Test Cases**: Create slice of test structs with descriptive names for each scenario
- **Test Execution**: Loop through cases using `t.Run()` with the case name
- **Assertion Pattern**: Check errors first, return early, then verify expected outputs

```go
// ✅ DO: Use this standard table-driven pattern
func TestFunction(t *testing.T) {
  tests := []struct {
    name        string
    input       SomeType
    wantOutput  ExpectedType
    wantErr     bool
    errContains string
  }{
    {
      name:        "happy path",
      input:       SomeType{...},
      wantOutput:  ExpectedType{...},
      wantErr:     false,
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

// ❌ DON'T: Write separate test functions for similar scenarios
func TestFunctionSuccess(t *testing.T) { ... }
func TestFunctionErrorCase1(t *testing.T) { ... }
func TestFunctionErrorCase2(t *testing.T) { ... }
```

## With Mock Setup

For tests requiring mocks, include a `setupMock` function field in the test struct:

```go
// ✅ DO: Include mock setup in the test table
tests := []struct {
  name       string
  input      SomeType
  setupMock  func(*mocks.MockType)
  wantOutput ExpectedType
  wantErr    bool
}{
  {
    name:  "successful case",
    input: SomeType{...},
    setupMock: func(m *mocks.MockType) {
      m.On("Method", mock.Anything).Return(response, nil)
    },
    wantOutput: ExpectedType{...},
    wantErr:    false,
  },
}

// In the test function:
mockObj := new(mocks.MockType)
tt.setupMock(mockObj)
// Don't forget verification
mockObj.AssertExpectations(t)
```

### Logger Setup for Tests

When testing components that require loggers, use the appropriate no-op logger implementation:

```go
// ✅ DO: Use NoOpLogger for tests
tests := []struct {
  name       string
  setupMocks func(*mocks.MockClient)
  // Other fields...
}{
  // Test cases...
}

for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
    // Setup mocks
    mockClient := new(mocks.MockClient)
    tt.setupMocks(mockClient)

    // Use the appropriate no-op logger:
    // For logger.Logger (internal implementation):
    log := logger.NewNoOpLogger()

    // For logr.Logger (when required by interfaces):
    logrLogger := logr.Discard()

    // Create service with mocks and logger
    service := NewService(log, mockClient)

    // Test execution...

    // Verify expectations
    mockClient.AssertExpectations(t)
  })
}
```

## Best Practices

- **Name Test Cases Clearly**: Use descriptive names indicating the scenario being tested
- **Group Related Fields**: Keep related fields together in the struct definition
- **Test One Behavior Per Case**: Each test case should focus on a single behavior or condition
- **Verify All Expectations**: Always use `AssertExpectations(t)` to verify mocks
- **Early Returns**: Return after error assertions to avoid nil/zero value assertions
- **Use Helper Functions**: Extract complex setup to helpers for readability
- **Be Consistent With Field Names**: Use consistent names (`wantErr`, `wantOutput`, etc.)

## Special Cases

### With Context

```go
tests := []struct {
  name         string
  setupContext func() context.Context
  input        SomeType
  // Other fields
}

// In the test function:
ctx := tt.setupContext()
got, err := FunctionUnderTest(ctx, tt.input)
```

### HTTP Handlers

```go
tests := []struct {
  name           string
  requestBody    string
  setupMock      func(*mocks.MockType)
  wantStatusCode int
  wantResponse   string
}

// Create and execute request in test function
req, _ := http.NewRequest(http.MethodPost, "/endpoint", strings.NewReader(tt.requestBody))
recorder := httptest.NewRecorder()
handler.ServeHTTP(recorder, req)

assert.Equal(t, tt.wantStatusCode, recorder.Code)
assert.JSONEq(t, tt.wantResponse, recorder.Body.String())
```
