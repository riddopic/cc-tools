# Mocking Standards and Mockery v3.5 Testing Guide

This document provides comprehensive guidance on mocking standards and using Mockery v3.5 for testing in the cc-tools project, including TDD workflows and advanced patterns.

## Core Principles

- **Mock at Interface Boundaries**: Create mocks for interfaces, not concrete implementations
- **Minimize Mock Knowledge**: Mocks should know as little as possible about implementation details
- **Use Mockery v3.5**: All mocks are generated using Mockery v3.5 and checked into version control
- **Keep Mocks Simple**: Focus on verifying behavior, not implementation details
- **TDD Integration**: Generate mocks as part of the Red-Green-Refactor cycle

## Directory Structure

Generated mocks follow this structure:

```text
internal/
├── hooks/mocks/            # Generated mocks for hook validation interfaces
├── config/mocks/           # Generated mocks for configuration interfaces
├── linter/mocks/           # Generated mocks for linter interfaces
├── runner/mocks/           # Generated mocks for command runner interfaces
└── interfaces/mocks/       # Generated mocks for core interfaces
```

## Mock Generation with Mockery v3.5

### Key Commands

```bash
task mocks          # Generate all mocks (regenerates from scratch)
task tools-install  # Install mockery and other required tools
```

### Configuration

Mock generation is configured in `.mockery.yml`. Each interface is explicitly listed with its output directory:

```yaml
template: testify
structname: "Mock{{.InterfaceName}}"
filename: "{{.InterfaceName}}.go"
force: true
```

### Version Control Strategy

✅ **We check generated mocks into Git** for:

- Better developer experience (tests run immediately after clone)
- Simplified CI/CD (no need to generate mocks in pipeline)
- Consistency across all environments
- Following Go community best practices

## TDD Workflow with Mockery

### The Red-Green-Refactor Cycle

The TDD cycle with Mockery follows the standard pattern:

1. **RED**: Write failing test with interface
2. **Generate mocks**: `task mocks`
3. **GREEN**: Write minimal implementation
4. **REFACTOR**: Improve while keeping tests green

### Step-by-Step TDD Example

#### Step 1: Define the Interface (RED Phase)

Start by defining the interface you need:

```go
// internal/interfaces/hook_validator.go
package interfaces

import (
    "context"
    "time"
)

type HookValidator interface {
    Validate(ctx context.Context, hookPath string, timeout time.Duration) (*ValidationResult, error)
    Check(hookPath string) error
}

type ValidationResult struct {
    IsValid     bool
    Severity    string
    PassCount   int
    FailCount   int
    Confidence  float64
}
```

#### Step 2: Generate the Mock

Add the interface to `.mockery.yml`:

```yaml
packages:
  github.com/riddopic/cc-tools/internal/interfaces:
    interfaces:
      HookValidator:
```

Generate the mock:

```bash
task mocks
```

#### Step 3: Write the Failing Test (RED Phase)

Write a test that uses the mock before implementing the service:

```go
// internal/validator/service_test.go
package validator_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "github.com/riddopic/cc-tools/internal/validator"
    "github.com/riddopic/cc-tools/internal/interfaces"
    "github.com/riddopic/cc-tools/internal/interfaces/mocks"
)

func TestValidatorService_RunValidation(t *testing.T) {
    // Arrange - Create mock with expectations
    mockValidator := mocks.NewMockHookValidator(t)

    mockValidator.EXPECT().Check("pre-commit-hook.sh").Return(nil).Once()

    mockValidator.EXPECT().Validate(
        mock.Anything,
        "pre-commit-hook.sh",
        30*time.Second,
    ).Return(&interfaces.ValidationResult{
        IsValid:     true,
        Severity:    "HIGH",
        PassCount:   15,
        FailCount:   0,
        Confidence:  0.95,
    }, nil).Once()

    // Act - This will fail because ValidatorService doesn't exist yet
    service := validator.NewValidatorService(mockValidator)
    result, err := service.RunValidation(context.Background(), "pre-commit-hook.sh", 30*time.Second)

    // Assert
    require.NoError(t, err)
    assert.True(t, result.IsValid)
    assert.Equal(t, "HIGH", result.Severity)
}
```

Run the test to see it fail:

```bash
go test ./internal/analyzer/... -v
# FAIL: undefined: analyzer.NewAnalyzerService
```

#### Step 4: Write Minimal Implementation (GREEN Phase)

Write just enough code to make the test pass:

```go
// internal/validator/service.go
package validator

import (
    "context"
    "time"

    "github.com/riddopic/cc-tools/internal/interfaces"
)

type ValidatorService struct {
    validator interfaces.HookValidator
}

func NewValidatorService(validator interfaces.HookValidator) *ValidatorService {
    return &ValidatorService{
        validator: validator,
    }
}

func (s *ValidatorService) RunValidation(ctx context.Context, hookPath string, timeout time.Duration) (*interfaces.ValidationResult, error) {
    // Minimal implementation - just delegate to the validator
    if err := s.validator.Check(hookPath); err != nil {
        return nil, err
    }

    return s.validator.Validate(ctx, hookPath, timeout)
}
```

#### Step 5: Refactor (REFACTOR Phase)

Improve the implementation while keeping tests green:

```go
func (s *ValidatorService) RunValidation(ctx context.Context, hookPath string, timeout time.Duration) (*interfaces.ValidationResult, error) {
    // Add validation
    if hookPath == "" {
        return nil, fmt.Errorf("hook path cannot be empty")
    }

    // Check first
    if err := s.validator.Check(hookPath); err != nil {
        return nil, fmt.Errorf("check failed for %s: %w", hookPath, err)
    }

    // Run validation
    result, err := s.validator.Validate(ctx, hookPath, timeout)
    if err != nil {
        return nil, fmt.Errorf("validation failed for %s: %w", hookPath, err)
    }

    return result, nil
}
```

## Basic Mock Usage

### Creating Mocks with Constructors

All generated mocks include a constructor function that automatically registers cleanup:

```go
func TestCommandRunner_BasicUsage(t *testing.T) {
    // Create mock with automatic cleanup
    runner := mocks.NewMockCommandRunner(t)

    // Test implementation here...
    // AssertExpectations is called automatically in t.Cleanup
}
```

### Setting Expectations with EXPECT()

Mockery v3.5+ generates an `EXPECT()` method for type-safe expectation setting:

```go
func TestWithMocks(t *testing.T) {
    // Create mock using the generated constructor
    mockClient := mocks.NewMockClient(t)

    // Type-safe expectation setting
    mockClient.EXPECT().GetData(
        mock.Anything,          // context.Context
        "test-id",             // specific ID
    ).Return(&Response{
        Data: "test-data",
    }, nil).Once()

    // Use the mock
    sut := NewService(mockClient)
    result, err := sut.ProcessData(context.Background(), "test-id")

    // Verify results (AssertExpectations called automatically)
    assert.NoError(t, err)
    assert.Equal(t, "processed: test-data", result)
}
```

### Common Matchers

```go
// ✅ DO: Use the appropriate matchers
mock.Anything                        // Matches any argument
mock.AnythingOfType("string")        // Matches any string
mock.MatchedBy(func(arg T) bool {})  // Custom matcher function
```

## Table-Driven Tests with Mocks

Define setup functions in test cases for clean organization:

```go
func TestWithTableDrivenMocks(t *testing.T) {
    tests := []struct {
        name       string
        setupMocks func(*mocks.MockCommandRunner)
        config     hooks.HookConfig
        wantResult *hooks.ValidationResult
        wantErr    error
    }{
        {
            name: "successful execution",
            setupMocks: func(m *mocks.MockCommandRunner) {
                m.EXPECT().Execute(mock.Anything, mock.MatchedBy(func(config hooks.HookConfig) bool {
                    return config.HookFile != ""
                })).Return(&hooks.ValidationResult{
                    Success: true,
                    PassCount: 21000,
                }, nil).Once()
            },
            config: hooks.HookConfig{
                HookFile: "test-hook.sh",
            },
            wantResult: &hooks.ValidationResult{
                Success: true,
                PassCount: 21000,
            },
            wantErr: nil,
        },
        {
            name: "invalid test file",
            setupMocks: func(m *mocks.MockCommandRunner) {
                m.EXPECT().Execute(mock.Anything, mock.MatchedBy(func(config hooks.HookConfig) bool {
                    return config.HookFile == ""
                })).Return(nil, hooks.ErrInvalidHookFile).Once()
            },
            config: hooks.HookConfig{
                HookFile: "", // Empty test file
            },
            wantResult: nil,
            wantErr:    hooks.ErrInvalidHookFile,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor := mocks.NewMockCommandRunner(t)
            tt.setupMocks(executor)

            result, err := executor.Execute(context.Background(), tt.config)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, result)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.wantResult, result)
            }
        })
    }
}
```

## Advanced Mock Patterns

### Dynamic Return Values with RunAndReturn

Use `RunAndReturn` for complex logic that depends on input parameters:

```go
func TestDynamicMockBehavior(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    // Return different results based on input
    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.RunnerOptions) (*interfaces.ValidationResult, error) {
            switch opts.CommandPath {
            case "valid.t.sol":
                return &interfaces.ValidationResult{
                    Success: true,
                    PassCount: 21000,
                }, nil
            case "invalid.t.sol":
                return nil, errors.New("invalid command")
            case "timeout.t.sol":
                // Simulate timeout
                <-ctx.Done()
                return nil, ctx.Err()
            default:
                return &interfaces.ValidationResult{
                    Success: false,
                    ExitCode: 1,
                }, nil
            }
        }).Times(3)
}
```

### Stateful Mocks

Create mocks that maintain state across calls:

```go
func TestStatefulMock(t *testing.T) {
    provider := mocks.NewMockProvider(t)

    // Track state in closure
    var timeout uint64 = 1000

    provider.EXPECT().Timeout(mock.Anything).RunAndReturn(
        func(ctx context.Context) (uint64, error) {
            timeout++
            return timeout, nil
        }).Times(3)

    // Each call returns incremented block number
    block1, _ := provider.Timeout(context.Background())
    block2, _ := provider.Timeout(context.Background())
    block3, _ := provider.Timeout(context.Background())

    assert.Equal(t, uint64(1001), block1)
    assert.Equal(t, uint64(1002), block2)
    assert.Equal(t, uint64(1003), block3)
}
```

### Complex Argument Matching with MatchedBy

```go
func TestComplexArgumentMatching(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    // Match complex struct fields
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.RunnerOptions) bool {
            return opts.Timeout > 18000000 &&
                   opts.Timeout < 19000000 &&
                   strings.HasPrefix(opts.ConfigPath, "https://") &&
                   opts.Timeout >= 30*time.Second &&
                   opts.RetryLimit > 0
        }),
    ).Return(&interfaces.ValidationResult{
        Success: true,
    }, nil).Once()

    // This will match
    result, err := executor.Execute(context.Background(), &interfaces.RunnerOptions{
        Timeout: 30,
        ConfigPath:     "https://eth-mainnet.alchemyapi.io",
        Timeout:     60 * time.Second,
        RetryLimit: 4096,
    })

    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

### Side Effects with Run

Use `Run` to simulate side effects or capture arguments:

```go
func TestSideEffects(t *testing.T) {
    runner := mocks.NewMockCommandRunner(t)
    var capturedConfig hooks.RunnerConfig

    runner.EXPECT().Run(mock.Anything, mock.Anything).Run(
        func(ctx context.Context, config hooks.RunnerConfig) {
            // Capture the config for verification
            capturedConfig = config
            // Simulate logging or metrics
            t.Logf("Running: %s %v", config.Command, config.Args)
        }).Return(&hooks.RunnerResult{
        ExitCode: 0,
    }, nil).Once()

    // Test and verify capturedConfig...
}
```

### Sequential and Ordered Expectations

```go
func TestOrderedMockCalls(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    // Create an expectation order
    inOrder := make(chan int, 3)

    // First call - compile
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.RunnerOptions) bool {
            return strings.Contains(opts.CommandPath, "compile")
        }),
    ).Run(func(args mock.Arguments) {
        inOrder <- 1
    }).Return(&interfaces.ValidationResult{Success: true}, nil).Once()

    // Second call - test
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.RunnerOptions) bool {
            return strings.Contains(opts.CommandPath, "test")
        }),
    ).Run(func(args mock.Arguments) {
        require.Equal(t, 1, <-inOrder)
        inOrder <- 2
    }).Return(&interfaces.ValidationResult{Success: true}, nil).Once()

    // Execute in order
    executor.Execute(context.Background(), &interfaces.RunnerOptions{CommandPath: "compile.sol"})
    executor.Execute(context.Background(), &interfaces.RunnerOptions{CommandPath: "test.sol"})

    require.Equal(t, 2, <-inOrder)
}
```

## Concurrent Testing with Mocks

### Thread-Safe Mock Usage

```go
func TestConcurrentMockUsage(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    // Setup expectations for concurrent calls
    var callCount atomic.Int32

    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.RunnerOptions) (*interfaces.ValidationResult, error) {
            count := callCount.Add(1)

            // Simulate some work
            time.Sleep(10 * time.Millisecond)

            return &interfaces.ValidationResult{
                Success: true,
                PassCount: uint64(21000 * count),
            }, nil
        }).Times(10)

    // Run concurrent executions
    var wg sync.WaitGroup
    results := make(chan *interfaces.ValidationResult, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            result, err := executor.Execute(context.Background(), &interfaces.RunnerOptions{
                CommandPath: fmt.Sprintf("test_%d.sol", id),
            })

            require.NoError(t, err)
            results <- result
        }(i)
    }

    wg.Wait()
    close(results)

    // Verify all results
    var totalGas uint64
    for result := range results {
        assert.True(t, result.Success)
        totalGas += result.PassCount
    }

    // Gas should be sum of 21000 * (1+2+...+10)
    expectedGas := uint64(21000 * (10 * 11 / 2))
    assert.Equal(t, expectedGas, totalGas)
}
```

### Race Condition Testing

```go
func TestRaceConditionDetection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping race condition test in short mode")
    }

    executor := mocks.NewMockCommandRunner(t)

    // Shared state to detect races
    var sharedCounter int
    var mu sync.Mutex

    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.RunnerOptions) (*interfaces.ValidationResult, error) {
            // Proper synchronization
            mu.Lock()
            sharedCounter++
            count := sharedCounter
            mu.Unlock()

            return &interfaces.ValidationResult{
                Success: true,
                PassCount: uint64(count * 1000),
            }, nil
        }).Times(100)

    // Run with race detector: go test -race
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            executor.Execute(context.Background(), &interfaces.RunnerOptions{})
        }()
    }

    wg.Wait()
    assert.Equal(t, 100, sharedCounter)
}
```

## Error Simulation Patterns

### Simulating Network Errors

```go
func TestNetworkErrorSimulation(t *testing.T) {
    provider := mocks.NewMockProvider(t)

    attempts := 0

    // Fail first 2 attempts, succeed on third
    provider.EXPECT().Timeout(mock.Anything).RunAndReturn(
        func(ctx context.Context) (uint64, error) {
            attempts++
            if attempts <= 2 {
                return 0, errors.New("connection refused")
            }
            return 30*time.Second, nil
        }).Times(3)

    // Service with retry logic
    service := NewConfigServiceWithRetry(provider, 3)

    blockNum, err := service.GetLatestBlockWithRetry(context.Background())
    require.NoError(t, err)
    assert.Equal(t, 30*time.Second, blockNum)
    assert.Equal(t, 3, attempts)
}
```

### Timeout Simulation

```go
func TestTimeoutSimulation(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.RunnerOptions) (*interfaces.ValidationResult, error) {
            select {
            case <-time.After(2 * time.Second):
                return &interfaces.ValidationResult{Success: true}, nil
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }).Once()

    // Use short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    result, err := executor.Execute(ctx, &interfaces.RunnerOptions{})

    require.ErrorIs(t, err, context.DeadlineExceeded)
    assert.Nil(t, result)
}
```

## Mock Verification Patterns

### Partial Mock Verification

```go
func TestPartialMockVerification(t *testing.T) {
    provider := mocks.NewMockProvider(t)

    // Setup multiple expectations
    provider.EXPECT().Timeout(mock.Anything).Return(30*time.Second, nil).Maybe()
    provider.EXPECT().ChainID(mock.Anything).Return(big.NewInt(1), nil).Maybe()
    provider.EXPECT().SuggestGasPrice(mock.Anything).Return(big.NewInt(30000000000), nil).Maybe()

    // Service might call some or all methods
    service := NewConfigService(provider)

    // Only Timeout is called
    blockNum, err := service.GetLatestBlock(context.Background())
    require.NoError(t, err)
    assert.Equal(t, 30*time.Second, blockNum)

    // Maybe() allows unused expectations - test still passes
}
```

### Custom Assertion Functions

```go
func TestCustomAssertions(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    var capturedOptions *interfaces.RunnerOptions

    executor.EXPECT().Execute(mock.Anything, mock.Anything).Run(
        func(args mock.Arguments) {
            // Capture arguments for custom verification
            capturedOptions = args.Get(1).(*interfaces.RunnerOptions)
        }).Return(&interfaces.ValidationResult{Success: true}, nil).Once()

    // Execute
    service := NewForgeService(executor)
    service.RunValidation(context.Background(), "test.sol", 30)

    // Custom assertions on captured arguments
    require.NotNil(t, capturedOptions)
    assert.Equal(t, "test.sol", capturedOptions.CommandPath)
    assert.Equal(t, 30*time.Second, capturedOptions.Timeout)
    assert.Greater(t, capturedOptions.Timeout, 30*time.Second)
}
```

## Integration Testing with Mocks

### Testing Multiple Components

```go
func TestIntegration_HookValidationWorkflow(t *testing.T) {
    // Create all necessary mocks
    executor := mocks.NewMockCommandRunner(t)
    provider := mocks.NewMockProvider(t)
    validator := mocks.NewMockHookValidator(t)
    logger := mocks.NewMockLogger(t)

    // Setup mock expectations for complete workflow

    // 1. Config loading
    provider.EXPECT().Timeout(mock.Anything).Return(30*time.Second, nil).Once()
    provider.EXPECT().GetConfig(mock.Anything, mock.Anything).Return(&types.Config{}, nil).Once()

    // 2. Hook validation
    validator.EXPECT().Validate("lint-command.sh").Return(nil).Once()

    // 3. Command execution
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.RunnerOptions) bool {
            return opts.CommandPath == "lint-command.sh" &&
                   opts.Timeout == 30
        }),
    ).Return(&interfaces.ValidationResult{
        Success: true,
        PassCount: 150000,
    }, nil).Once()

    // 4. Analysis
    validator.EXPECT().Analyze(
        mock.Anything,
        "lint-command.sh",
        30*time.Second,
    ).Return(&interfaces.ValidationResult{
        IsValid:    true,
        Severity:   "HIGH",
        Confidence: 0.95,
    }, nil).Once()

    // 5. Logging (use Maybe() for optional calls)
    logger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything).Maybe()

    // Create services with mocks
    configManager := command.NewConfigManager(provider)
    validatorService := validator.NewValidatorService(validator, logger)
    hookService := NewHookService(executor, configManager, validatorService, logger)

    // Run complete workflow
    result, err := hookService.RunValidationWorkflow(
        context.Background(),
        "lint-command.sh",
        30,
    )

    // Verify workflow completed successfully
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "HIGH", result.Severity)
    assert.Equal(t, uint64(150000), result.PassCount)
}
```

## Mock Factory Pattern

Create reusable mock setups:

```go
// test/mocks/factory.go
type MockFactory struct {
    t *testing.T
}

func NewMockFactory(t *testing.T) *MockFactory {
    return &MockFactory{t: t}
}

func (f *MockFactory) CommandRunner(opts ...CommandRunnerOption) *mocks.MockCommandRunner {
    executor := mocks.NewMockCommandRunner(f.t)

    // Apply default expectations
    executor.EXPECT().GetVersion(mock.Anything).Return("0.2.0", nil).Maybe()

    // Apply custom options
    for _, opt := range opts {
        opt(executor)
    }

    return executor
}

type CommandRunnerOption func(*mocks.MockCommandRunner)

func WithSuccessfulExecution(gasUsed uint64) CommandRunnerOption {
    return func(m *mocks.MockCommandRunner) {
        m.EXPECT().Execute(mock.Anything, mock.Anything).Return(&interfaces.ValidationResult{
            Success: true,
            PassCount: gasUsed,
        }, nil).Maybe()
    }
}

// Usage in tests
func TestWithMockFactory(t *testing.T) {
    factory := NewMockFactory(t)

    executor := factory.CommandRunner(
        WithSuccessfulExecution(21000),
    )

    service := NewForgeService(executor)
    result, err := service.Run(context.Background())

    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Common Patterns in cc-tools

### Mocking RPC Providers

```go
func TestConfigService_GetTimeout(t *testing.T) {
    provider := mocks.NewMockProvider(t)

    provider.EXPECT().Timeout(mock.Anything).Return(30*time.Second, nil).Once()

    service := NewConfigService(provider)
    blockNum, err := service.GetLatestBlock(context.Background())

    require.NoError(t, err)
    assert.Equal(t, 30*time.Second, blockNum)
}
```

### Context Handling and Cancellation

```go
func TestCommandRunner_ContextCancellation(t *testing.T) {
    executor := mocks.NewMockCommandRunner(t)

    // Mock long-running execution that respects context
    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, config hooks.HookConfig) (*hooks.ValidationResult, error) {
            select {
            case <-time.After(10 * time.Second):
                return &hooks.ValidationResult{Success: true}, nil
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }).Once()

    // Create context with short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    result, err := executor.Execute(ctx, hooks.HookConfig{})

    // Verify context cancellation is respected
    require.ErrorIs(t, err, context.DeadlineExceeded)
    assert.Nil(t, result)
}
```

### Mocking Multiple Related Calls

```go
// ✅ DO: Set up related calls with appropriate returns
mockClient.On("BeginTransaction", mock.Anything).
  Return("tx-123", nil)
mockClient.On("Commit", mock.Anything, "tx-123").
  Return(nil)
```

### Mocking Channels and Go Routines

```go
// ✅ DO: Use channels for asynchronous code
func (m *MockProcessor) Process(ctx context.Context, data string) <-chan Result {
  args := m.Called(ctx, data)
  return args.Get(0).(<-chan Result)
}

// In the test:
resultCh := make(chan Result, 1)
resultCh <- Result{Data: "processed", Error: nil}
close(resultCh)

mockProcessor.On("Process", mock.Anything, "test").
  Return(resultCh)
```

## Performance Considerations

### Minimizing Mock Overhead

```go
func BenchmarkMockOverhead(b *testing.B) {
    // Setup mock once
    executor := mocks.NewMockCommandRunner(b)

    executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(
        &interfaces.ValidationResult{Success: true},
        nil,
    ).Times(b.N)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        executor.Execute(context.Background(), &interfaces.RunnerOptions{})
    }
}
```

### Efficient Mock Setup for Large Test Suites

```go
func TestSuiteWithSharedMocks(t *testing.T) {
    // Setup phase - create all mocks once
    type testEnv struct {
        executor *mocks.MockCommandRunner
        provider *mocks.MockProvider
        service  *Service
    }

    setupEnv := func(t *testing.T) *testEnv {
        executor := mocks.NewMockCommandRunner(t)
        provider := mocks.NewMockProvider(t)

        return &testEnv{
            executor: executor,
            provider: provider,
            service:  NewService(executor, provider),
        }
    }

    t.Run("test case 1", func(t *testing.T) {
        env := setupEnv(t)
        env.executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(nil, nil).Once()
        // Test logic
    })

    t.Run("test case 2", func(t *testing.T) {
        env := setupEnv(t)
        env.provider.EXPECT().Timeout(mock.Anything).Return(30*time.Second, nil).Once()
        // Test logic
    })
}
```

## Best Practices

### When to Use Mocks vs Real Implementations

```go
// ✅ DO: Mock external dependencies
func TestService_WithExternalDependencies(t *testing.T) {
    // Mock RPC client
    rpcClient := mocks.NewMockProvider(t)

    // Mock file system operations
    fileSystem := mocks.NewMockFileSystem(t)

    // Use real internal components
    calculator := NewProfitCalculator() // Real implementation

    service := NewService(rpcClient, fileSystem, calculator)
    // ... test implementation
}

// ✅ DO: Use real implementations for simple dependencies
func TestCalculator_Add(t *testing.T) {
    // Don't mock simple value objects or pure functions
    calculator := NewCalculator() // Real implementation
    result := calculator.Add(2, 3)
    assert.Equal(t, 5, result)
}
```

### Avoiding Over-mocking

```go
// ❌ DON'T: Mock everything
func TestBadExample(t *testing.T) {
    // Too many mocks makes test brittle
    logger := mocks.NewMockLogger(t)
    config := mocks.NewMockConfig(t)
    timer := mocks.NewMockTimer(t)
    // This test is testing implementation, not behavior
}

// ✅ DO: Mock at boundaries, use real objects internally
func TestGoodExample(t *testing.T) {
    // Mock only external dependencies
    executor := mocks.NewMockCommandRunner(t)

    executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(&hooks.ValidationResult{
        Success: true,
    }, nil).Once()

    // Use real internal components
    service := NewHookValidationService(executor)
    result, err := service.ValidateHooks(context.Background(), "lint-command.sh")

    require.NoError(t, err)
    assert.True(t, result.IsValid)
}
```

## Best Practices Summary

### DO

- ✅ Generate mocks before writing tests
- ✅ Commit generated mocks to version control
- ✅ Use constructor functions (NewMock\*)
- ✅ Write tests before implementation (TDD)
- ✅ Use EXPECT() pattern for setting expectations
- ✅ Test both success and failure paths
- ✅ Use Maybe() for optional mock calls
- ✅ Keep mocks simple and focused
- ✅ Mock at interface boundaries
- ✅ Use real implementations for simple dependencies
- ✅ Verify expectations with AssertExpectations()
- ✅ Use `.Once()` vs `.Times(n)` to be explicit about call counts

### DON'T

- ❌ Create manual mock implementations
- ❌ Skip the RED phase of TDD
- ❌ Mock internal components unnecessarily
- ❌ Forget to regenerate mocks after interface changes
- ❌ Use mock.Anything when specific matching is better
- ❌ Test implementation details instead of behavior
- ❌ Mock structs that you own - use interfaces
- ❌ Have mocks return other mocks (mock chains)
- ❌ Use `.Return(mock.Anything)` - be explicit about return values
- ❌ Create tests that require too many mocks (sign of tight coupling)

## Common Anti-Patterns to Avoid

- ❌ Don't mock structs that you own - use interfaces
- ❌ Don't have mocks return other mocks (mock chains)
- ❌ Don't verify implementation details
- ❌ Don't use `.Return(mock.Anything)` - be explicit about return values
- ❌ Don't commandt to run `task mocks` after interface changes

## Debugging Mock Issues

### Verbose Mock Logging

```go
func TestWithMockDebugging(t *testing.T) {
    if testing.Verbose() {
        // Enable detailed mock logging
        mock.TestingT(t)
    }

    executor := mocks.NewMockCommandRunner(t)

    // Add debug logging to expectations
    executor.EXPECT().Execute(mock.Anything, mock.Anything).Run(
        func(args mock.Arguments) {
            ctx := args.Get(0).(context.Context)
            opts := args.Get(1).(*interfaces.RunnerOptions)

            t.Logf("Execute called with context deadline: %v", ctx.Deadline())
            t.Logf("Execute options: %+v", opts)
        }).Return(&interfaces.ValidationResult{Success: true}, nil).Once()

    // Test execution
    service := NewForgeService(executor)
    service.Run(context.Background())
}
```

## Troubleshooting

### Common Errors and Solutions

#### "No return value specified" panic

```go
// ❌ Problem: Mock method called without return value
executor.EXPECT().Execute(mock.Anything, mock.Anything) // Missing .Return()

// ✅ Solution: Always specify return values
executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(&hooks.ValidationResult{}, nil)
```

#### "Mock call missing" error

```go
// ❌ Problem: Expected call was not made
executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(result, nil).Once()
// But Execute() was never called

// ✅ Solution: Verify your code path actually calls the mocked method
```

#### Mock Not Found

```bash
# Verify interface is in .mockery.yml
rg "YourInterface" .mockery.yml

# Regenerate mocks
task mocks

# Check mock was generated
rg --files -g "*YourInterface.go" mocks/
```

#### Test Failures After Interface Change

```bash
# Regenerate all mocks
task mocks

# Run tests to find issues
go test ./... -v

# Fix test expectations to match new interface
```

#### Regenerating Mocks After Interface Changes

```bash
# Regenerate all mocks
task mocks

# Verify mocks compile
go build ./internal/hooks/mocks/...
```

## Mock Generation Workflow

### Complete Development Cycle

1. **Define Interface**

   ```bash
   # Create new interface file
   touch internal/interfaces/new_service.go
   ```

2. **Add to Mockery Config**

   ```bash
   # Edit .mockery.yml
   vim .mockery.yml
   ```

3. **Generate Mock**

   ```bash
   task mocks
   ```

4. **Write Test First**

   ```bash
   # Create test file
   touch internal/service/new_service_test.go
   # Write failing test using generated mock
   ```

5. **Run Test (See RED)**

   ```bash
   go test ./internal/service/... -v
   ```

6. **Implement Code**

   ```bash
   # Create implementation
   touch internal/service/new_service.go
   # Write minimal code
   ```

7. **Run Test (See GREEN)**

   ```bash
   go test ./internal/service/... -v
   ```

8. **Refactor**

   ```bash
   # Improve code while keeping tests green
   ```

9. **Commit Everything**

   ```bash
   git add .
   git commit -m "feat: implement new service with TDD and mocks"
   ```

## CI/CD Integration

Ensure mocks are always up to date in CI:

```yaml
# .github/workflows/test.yml
- name: Verify mocks are up to date
  run: |
    task mocks
    if [ -n "$(git status --porcelain)" ]; then
      echo "Mocks are out of date. Please run 'task mocks' and commit."
      exit 1
    fi
```

## Logger Mocking

For mocking the Zap logger interface:

```go
// ✅ DO: Create a logger mock in internal/logger/mocks
// internal/logger/mocks/mock_logger.go
package mocks

import (
  "github.com/stretchr/testify/mock"
  "go.uber.org/zap"
)

type MockLogger struct {
  mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
  args := []interface{}{msg}
  for _, field := range fields {
    args = append(args, field)
  }
  m.Called(args...)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
  args := []interface{}{msg}
  for _, field := range fields {
    args = append(args, field)
  }
  m.Called(args...)
}

// Additional methods as required by the Logger interface
```

Or use NoOpLogger for Tests:

```go
// For internal/logger.Logger:
logger := logger.NewNoOpLogger()

// For logr.Logger (when required by interfaces):
logger := logr.Discard()
```

## References

- [TDD Principles](../philosophy/tdd-principles.md)
- [Mockery v3.5 Documentation](https://vektra.github.io/mockery/latest/)
- [Mockery v3.5 Release Notes](https://github.com/vektra/mockery/releases/tag/v3.5.0)
- [Testify Mock Package](https://pkg.go.dev/github.com/stretchr/testify/mock)
- [Testify Mock Advanced Usage](https://pkg.go.dev/github.com/stretchr/testify/mock#readme-usage)
- [Go Testing Tutorial](https://go.dev/doc/tutorial/add-a-test)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/fuzz)
- [cc-tools Mockery Implementation Guide](../../PRPs/research/mockery-v3-implementation-guide.md)
