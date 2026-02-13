# Mocking Standards and Mockery v3.5 Testing Guide

This document provides comprehensive guidance on mocking standards and using Mockery v3.5 for testing in the Quanta project, including TDD workflows and advanced patterns.

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
├── blockchain/
│   ├── cache/mocks/         # Generated mocks for cache interfaces
│   ├── explorer/mocks/      # Generated mocks for explorer interfaces
│   ├── proxy/mocks/         # Generated mocks for proxy interfaces
│   ├── ratelimit/mocks/     # Generated mocks for ratelimit interfaces
│   ├── rpc/mocks/          # Generated mocks for RPC interfaces
│   └── source/mocks/       # Generated mocks for source interfaces
├── foundry/mocks/          # Generated mocks for foundry interfaces
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
// internal/interfaces/exploit_analyzer.go
package interfaces

import (
    "context"
    "math/big"
)

type ExploitAnalyzer interface {
    Analyze(ctx context.Context, exploitPath string, blockNumber uint64) (*AnalysisResult, error)
    Validate(exploitPath string) error
}

type AnalysisResult struct {
    IsValid     bool
    Severity    string
    GasRequired uint64
    ProfitWei   *big.Int
    Confidence  float64
}
```

#### Step 2: Generate the Mock

Add the interface to `.mockery.yml`:

```yaml
packages:
  github.com/riddopic/quanta/internal/interfaces:
    interfaces:
      ExploitAnalyzer:
```

Generate the mock:

```bash
task mocks
```

#### Step 3: Write the Failing Test (RED Phase)

Write a test that uses the mock before implementing the service:

```go
// internal/analyzer/service_test.go
package analyzer_test

import (
    "context"
    "testing"
    "math/big"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "github.com/riddopic/quanta/internal/analyzer"
    "github.com/riddopic/quanta/internal/interfaces"
    "github.com/riddopic/quanta/internal/interfaces/mocks"
)

func TestAnalyzerService_RunAnalysis(t *testing.T) {
    // Arrange - Create mock with expectations
    mockAnalyzer := mocks.NewMockExploitAnalyzer(t)

    mockAnalyzer.EXPECT().Validate("exploit.sol").Return(nil).Once()

    mockAnalyzer.EXPECT().Analyze(
        mock.Anything,
        "exploit.sol",
        uint64(18500000),
    ).Return(&interfaces.AnalysisResult{
        IsValid:     true,
        Severity:    "HIGH",
        GasRequired: 150000,
        ProfitWei:   big.NewInt(1000000000000000000), // 1 ETH
        Confidence:  0.95,
    }, nil).Once()

    // Act - This will fail because AnalyzerService doesn't exist yet
    service := analyzer.NewAnalyzerService(mockAnalyzer)
    result, err := service.RunAnalysis(context.Background(), "exploit.sol", 18500000)

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
// internal/analyzer/service.go
package analyzer

import (
    "context"

    "github.com/riddopic/quanta/internal/interfaces"
)

type AnalyzerService struct {
    analyzer interfaces.ExploitAnalyzer
}

func NewAnalyzerService(analyzer interfaces.ExploitAnalyzer) *AnalyzerService {
    return &AnalyzerService{
        analyzer: analyzer,
    }
}

func (s *AnalyzerService) RunAnalysis(ctx context.Context, exploitPath string, blockNumber uint64) (*interfaces.AnalysisResult, error) {
    // Minimal implementation - just delegate to the analyzer
    if err := s.analyzer.Validate(exploitPath); err != nil {
        return nil, err
    }

    return s.analyzer.Analyze(ctx, exploitPath, blockNumber)
}
```

#### Step 5: Refactor (REFACTOR Phase)

Improve the implementation while keeping tests green:

```go
func (s *AnalyzerService) RunAnalysis(ctx context.Context, exploitPath string, blockNumber uint64) (*interfaces.AnalysisResult, error) {
    // Add validation
    if exploitPath == "" {
        return nil, fmt.Errorf("exploit path cannot be empty")
    }

    // Validate first
    if err := s.analyzer.Validate(exploitPath); err != nil {
        return nil, fmt.Errorf("validation failed for %s: %w", exploitPath, err)
    }

    // Run analysis
    result, err := s.analyzer.Analyze(ctx, exploitPath, blockNumber)
    if err != nil {
        return nil, fmt.Errorf("analysis failed for %s: %w", exploitPath, err)
    }

    return result, nil
}
```

## Basic Mock Usage

### Creating Mocks with Constructors

All generated mocks include a constructor function that automatically registers cleanup:

```go
func TestForgeExecutor_BasicUsage(t *testing.T) {
    // Create mock with automatic cleanup
    executor := mocks.NewMockForgeExecutor(t)

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
        setupMocks func(*mocks.MockForgeExecutor)
        config     foundry.ForgeConfig
        wantResult *foundry.ForgeResult
        wantErr    error
    }{
        {
            name: "successful execution",
            setupMocks: func(m *mocks.MockForgeExecutor) {
                m.EXPECT().Execute(mock.Anything, mock.MatchedBy(func(config foundry.ForgeConfig) bool {
                    return config.TestFile != ""
                })).Return(&foundry.ForgeResult{
                    Success: true,
                    GasUsed: 21000,
                }, nil).Once()
            },
            config: foundry.ForgeConfig{
                TestFile: "test.t.sol",
            },
            wantResult: &foundry.ForgeResult{
                Success: true,
                GasUsed: 21000,
            },
            wantErr: nil,
        },
        {
            name: "invalid test file",
            setupMocks: func(m *mocks.MockForgeExecutor) {
                m.EXPECT().Execute(mock.Anything, mock.MatchedBy(func(config foundry.ForgeConfig) bool {
                    return config.TestFile == ""
                })).Return(nil, foundry.ErrInvalidTestFile).Once()
            },
            config: foundry.ForgeConfig{
                TestFile: "", // Empty test file
            },
            wantResult: nil,
            wantErr:    foundry.ErrInvalidTestFile,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor := mocks.NewMockForgeExecutor(t)
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
    executor := mocks.NewMockForgeExecutor(t)

    // Return different results based on input
    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.ExecuteOptions) (*interfaces.TestResult, error) {
            switch opts.ExploitPath {
            case "valid.t.sol":
                return &interfaces.TestResult{
                    Success: true,
                    GasUsed: 21000,
                }, nil
            case "invalid.t.sol":
                return nil, errors.New("invalid exploit")
            case "timeout.t.sol":
                // Simulate timeout
                <-ctx.Done()
                return nil, ctx.Err()
            default:
                return &interfaces.TestResult{
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
    var blockNumber uint64 = 1000

    provider.EXPECT().BlockNumber(mock.Anything).RunAndReturn(
        func(ctx context.Context) (uint64, error) {
            blockNumber++
            return blockNumber, nil
        }).Times(3)

    // Each call returns incremented block number
    block1, _ := provider.BlockNumber(context.Background())
    block2, _ := provider.BlockNumber(context.Background())
    block3, _ := provider.BlockNumber(context.Background())

    assert.Equal(t, uint64(1001), block1)
    assert.Equal(t, uint64(1002), block2)
    assert.Equal(t, uint64(1003), block3)
}
```

### Complex Argument Matching with MatchedBy

```go
func TestComplexArgumentMatching(t *testing.T) {
    executor := mocks.NewMockForgeExecutor(t)

    // Match complex struct fields
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.ExecuteOptions) bool {
            return opts.BlockNumber > 18000000 &&
                   opts.BlockNumber < 19000000 &&
                   strings.HasPrefix(opts.ForkURL, "https://") &&
                   opts.Timeout >= 30*time.Second &&
                   opts.MemoryLimit > 0
        }),
    ).Return(&interfaces.TestResult{
        Success: true,
    }, nil).Once()

    // This will match
    result, err := executor.Execute(context.Background(), &interfaces.ExecuteOptions{
        BlockNumber: 18500000,
        ForkURL:     "https://eth-mainnet.alchemyapi.io",
        Timeout:     60 * time.Second,
        MemoryLimit: 4096,
    })

    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

### Side Effects with Run

Use `Run` to simulate side effects or capture arguments:

```go
func TestSideEffects(t *testing.T) {
    runner := mocks.NewMockProcessRunner(t)
    var capturedConfig foundry.ProcessConfig

    runner.EXPECT().Run(mock.Anything, mock.Anything).Run(
        func(ctx context.Context, config foundry.ProcessConfig) {
            // Capture the config for verification
            capturedConfig = config
            // Simulate logging or metrics
            t.Logf("Running: %s %v", config.Command, config.Args)
        }).Return(&foundry.ProcessResult{
        ExitCode: 0,
    }, nil).Once()

    // Test and verify capturedConfig...
}
```

### Sequential and Ordered Expectations

```go
func TestOrderedMockCalls(t *testing.T) {
    executor := mocks.NewMockForgeExecutor(t)

    // Create an expectation order
    inOrder := make(chan int, 3)

    // First call - compile
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.ExecuteOptions) bool {
            return strings.Contains(opts.ExploitPath, "compile")
        }),
    ).Run(func(args mock.Arguments) {
        inOrder <- 1
    }).Return(&interfaces.TestResult{Success: true}, nil).Once()

    // Second call - test
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.ExecuteOptions) bool {
            return strings.Contains(opts.ExploitPath, "test")
        }),
    ).Run(func(args mock.Arguments) {
        require.Equal(t, 1, <-inOrder)
        inOrder <- 2
    }).Return(&interfaces.TestResult{Success: true}, nil).Once()

    // Execute in order
    executor.Execute(context.Background(), &interfaces.ExecuteOptions{ExploitPath: "compile.sol"})
    executor.Execute(context.Background(), &interfaces.ExecuteOptions{ExploitPath: "test.sol"})

    require.Equal(t, 2, <-inOrder)
}
```

## Concurrent Testing with Mocks

### Thread-Safe Mock Usage

```go
func TestConcurrentMockUsage(t *testing.T) {
    executor := mocks.NewMockForgeExecutor(t)

    // Setup expectations for concurrent calls
    var callCount atomic.Int32

    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.ExecuteOptions) (*interfaces.TestResult, error) {
            count := callCount.Add(1)

            // Simulate some work
            time.Sleep(10 * time.Millisecond)

            return &interfaces.TestResult{
                Success: true,
                GasUsed: uint64(21000 * count),
            }, nil
        }).Times(10)

    // Run concurrent executions
    var wg sync.WaitGroup
    results := make(chan *interfaces.TestResult, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            result, err := executor.Execute(context.Background(), &interfaces.ExecuteOptions{
                ExploitPath: fmt.Sprintf("test_%d.sol", id),
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
        totalGas += result.GasUsed
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

    executor := mocks.NewMockForgeExecutor(t)

    // Shared state to detect races
    var sharedCounter int
    var mu sync.Mutex

    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.ExecuteOptions) (*interfaces.TestResult, error) {
            // Proper synchronization
            mu.Lock()
            sharedCounter++
            count := sharedCounter
            mu.Unlock()

            return &interfaces.TestResult{
                Success: true,
                GasUsed: uint64(count * 1000),
            }, nil
        }).Times(100)

    // Run with race detector: go test -race
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            executor.Execute(context.Background(), &interfaces.ExecuteOptions{})
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
    provider.EXPECT().BlockNumber(mock.Anything).RunAndReturn(
        func(ctx context.Context) (uint64, error) {
            attempts++
            if attempts <= 2 {
                return 0, errors.New("connection refused")
            }
            return uint64(18500000), nil
        }).Times(3)

    // Service with retry logic
    service := NewBlockchainServiceWithRetry(provider, 3)

    blockNum, err := service.GetLatestBlockWithRetry(context.Background())
    require.NoError(t, err)
    assert.Equal(t, uint64(18500000), blockNum)
    assert.Equal(t, 3, attempts)
}
```

### Timeout Simulation

```go
func TestTimeoutSimulation(t *testing.T) {
    executor := mocks.NewMockForgeExecutor(t)

    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, opts *interfaces.ExecuteOptions) (*interfaces.TestResult, error) {
            select {
            case <-time.After(2 * time.Second):
                return &interfaces.TestResult{Success: true}, nil
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }).Once()

    // Use short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    result, err := executor.Execute(ctx, &interfaces.ExecuteOptions{})

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
    provider.EXPECT().BlockNumber(mock.Anything).Return(uint64(18500000), nil).Maybe()
    provider.EXPECT().ChainID(mock.Anything).Return(big.NewInt(1), nil).Maybe()
    provider.EXPECT().SuggestGasPrice(mock.Anything).Return(big.NewInt(30000000000), nil).Maybe()

    // Service might call some or all methods
    service := NewBlockchainService(provider)

    // Only BlockNumber is called
    blockNum, err := service.GetLatestBlock(context.Background())
    require.NoError(t, err)
    assert.Equal(t, uint64(18500000), blockNum)

    // Maybe() allows unused expectations - test still passes
}
```

### Custom Assertion Functions

```go
func TestCustomAssertions(t *testing.T) {
    executor := mocks.NewMockForgeExecutor(t)

    var capturedOptions *interfaces.ExecuteOptions

    executor.EXPECT().Execute(mock.Anything, mock.Anything).Run(
        func(args mock.Arguments) {
            // Capture arguments for custom verification
            capturedOptions = args.Get(1).(*interfaces.ExecuteOptions)
        }).Return(&interfaces.TestResult{Success: true}, nil).Once()

    // Execute
    service := NewForgeService(executor)
    service.RunExploit(context.Background(), "test.sol", 18500000)

    // Custom assertions on captured arguments
    require.NotNil(t, capturedOptions)
    assert.Equal(t, "test.sol", capturedOptions.ExploitPath)
    assert.Equal(t, uint64(18500000), capturedOptions.BlockNumber)
    assert.Greater(t, capturedOptions.Timeout, 30*time.Second)
}
```

## Integration Testing with Mocks

### Testing Multiple Components

```go
func TestIntegration_ExploitAnalysisWorkflow(t *testing.T) {
    // Create all necessary mocks
    executor := mocks.NewMockForgeExecutor(t)
    provider := mocks.NewMockProvider(t)
    analyzer := mocks.NewMockExploitAnalyzer(t)
    logger := mocks.NewMockLogger(t)

    // Setup mock expectations for complete workflow

    // 1. Fork creation
    provider.EXPECT().BlockNumber(mock.Anything).Return(uint64(18500000), nil).Once()
    provider.EXPECT().GetBlock(mock.Anything, mock.Anything).Return(&types.Block{}, nil).Once()

    // 2. Exploit validation
    analyzer.EXPECT().Validate("exploit.sol").Return(nil).Once()

    // 3. Exploit execution
    executor.EXPECT().Execute(
        mock.Anything,
        mock.MatchedBy(func(opts *interfaces.ExecuteOptions) bool {
            return opts.ExploitPath == "exploit.sol" &&
                   opts.BlockNumber == 18500000
        }),
    ).Return(&interfaces.TestResult{
        Success: true,
        GasUsed: 150000,
    }, nil).Once()

    // 4. Analysis
    analyzer.EXPECT().Analyze(
        mock.Anything,
        "exploit.sol",
        uint64(18500000),
    ).Return(&interfaces.AnalysisResult{
        IsValid:    true,
        Severity:   "HIGH",
        Confidence: 0.95,
    }, nil).Once()

    // 5. Logging (use Maybe() for optional calls)
    logger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything).Maybe()

    // Create services with mocks
    forkManager := forge.NewForkManager(provider)
    analyzerService := analyzer.NewAnalyzerService(analyzer, logger)
    exploitService := NewExploitService(executor, forkManager, analyzerService, logger)

    // Run complete workflow
    result, err := exploitService.RunExploitWorkflow(
        context.Background(),
        "exploit.sol",
        18500000,
    )

    // Verify workflow completed successfully
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "HIGH", result.Severity)
    assert.Equal(t, uint64(150000), result.GasUsed)
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

func (f *MockFactory) ForgeExecutor(opts ...ForgeExecutorOption) *mocks.MockForgeExecutor {
    executor := mocks.NewMockForgeExecutor(f.t)

    // Apply default expectations
    executor.EXPECT().GetVersion(mock.Anything).Return("0.2.0", nil).Maybe()

    // Apply custom options
    for _, opt := range opts {
        opt(executor)
    }

    return executor
}

type ForgeExecutorOption func(*mocks.MockForgeExecutor)

func WithSuccessfulExecution(gasUsed uint64) ForgeExecutorOption {
    return func(m *mocks.MockForgeExecutor) {
        m.EXPECT().Execute(mock.Anything, mock.Anything).Return(&interfaces.TestResult{
            Success: true,
            GasUsed: gasUsed,
        }, nil).Maybe()
    }
}

// Usage in tests
func TestWithMockFactory(t *testing.T) {
    factory := NewMockFactory(t)

    executor := factory.ForgeExecutor(
        WithSuccessfulExecution(21000),
    )

    service := NewForgeService(executor)
    result, err := service.Run(context.Background())

    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Common Patterns in Quanta

### Mocking RPC Providers

```go
func TestBlockchainService_GetBlockNumber(t *testing.T) {
    provider := mocks.NewMockProvider(t)

    provider.EXPECT().BlockNumber(mock.Anything).Return(uint64(18500000), nil).Once()

    service := NewBlockchainService(provider)
    blockNum, err := service.GetLatestBlock(context.Background())

    require.NoError(t, err)
    assert.Equal(t, uint64(18500000), blockNum)
}
```

### Context Handling and Cancellation

```go
func TestForgeExecutor_ContextCancellation(t *testing.T) {
    executor := mocks.NewMockForgeExecutor(t)

    // Mock long-running execution that respects context
    executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(
        func(ctx context.Context, config foundry.ForgeConfig) (*foundry.ForgeResult, error) {
            select {
            case <-time.After(10 * time.Second):
                return &foundry.ForgeResult{Success: true}, nil
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }).Once()

    // Create context with short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    result, err := executor.Execute(ctx, foundry.ForgeConfig{})

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
    executor := mocks.NewMockForgeExecutor(b)

    executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(
        &interfaces.TestResult{Success: true},
        nil,
    ).Times(b.N)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        executor.Execute(context.Background(), &interfaces.ExecuteOptions{})
    }
}
```

### Efficient Mock Setup for Large Test Suites

```go
func TestSuiteWithSharedMocks(t *testing.T) {
    // Setup phase - create all mocks once
    type testEnv struct {
        executor *mocks.MockForgeExecutor
        provider *mocks.MockProvider
        service  *Service
    }

    setupEnv := func(t *testing.T) *testEnv {
        executor := mocks.NewMockForgeExecutor(t)
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
        env.provider.EXPECT().BlockNumber(mock.Anything).Return(uint64(18500000), nil).Once()
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
    executor := mocks.NewMockForgeExecutor(t)

    executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(&foundry.ForgeResult{
        Success: true,
    }, nil).Once()

    // Use real internal components
    service := NewExploitAnalysisService(executor)
    result, err := service.AnalyzeExploit(context.Background(), "exploit.t.sol")

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
- ❌ Don't forget to run `task mocks` after interface changes

## Debugging Mock Issues

### Verbose Mock Logging

```go
func TestWithMockDebugging(t *testing.T) {
    if testing.Verbose() {
        // Enable detailed mock logging
        mock.TestingT(t)
    }

    executor := mocks.NewMockForgeExecutor(t)

    // Add debug logging to expectations
    executor.EXPECT().Execute(mock.Anything, mock.Anything).Run(
        func(args mock.Arguments) {
            ctx := args.Get(0).(context.Context)
            opts := args.Get(1).(*interfaces.ExecuteOptions)

            t.Logf("Execute called with context deadline: %v", ctx.Deadline())
            t.Logf("Execute options: %+v", opts)
        }).Return(&interfaces.TestResult{Success: true}, nil).Once()

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
executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(&foundry.ForgeResult{}, nil)
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
go build ./internal/foundry/mocks/...
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
- [Quanta Mockery Implementation Guide](../../PRPs/research/mockery-v3-implementation-guide.md)
