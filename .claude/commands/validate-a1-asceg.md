---
description: Validate implementation against A1-ASCEG specification
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
model: opus
skills:
  - a1-asceg-research
  - table-ix-testing
---

# Validate A1-ASCEG Specification Compliance

## Purpose

Comprehensively validate that the Quanta implementation correctly implements the A1 Smart Contract Exploit Generation system as described in the A1-ASCEG research paper. This command performs deep analysis to ensure all six domain-specific tools, the iterative agent loop, and core functionality requirements are properly implemented.

**IMPORTANT**: This is a critical quality gate that ensures Quanta correctly implements the A1 system architecture. Analyze thoroughly.

## Validation Scope

This command validates:

1. **Six Domain-Specific Tools** (as per A1-ASCEG Section IV)
2. **Iterative Agent Loop** with execution feedback
3. **Multi-Chain Support** (Ethereum and BSC)
4. **Economic Validation** and revenue normalization
5. **Output Formatting** and exploit validation

## Agent Orchestration Strategy

Use the **product-manager-orchestrator** to coordinate multiple specialized agents for comprehensive A1-ASCEG validation:

### Primary Validation Agents

1. **code-analyzer-debugger** - Implementation verification

   - Verify all 6 tools are implemented (not stubs)
   - Check agent loop implementation
   - Validate multi-chain architecture
   - Identify missing A1 components

2. **systems-architect** (specification compliance)

   - Cross-reference with A1-ASCEG paper
   - Compare against research requirements
   - Verify technical accuracy
   - Check algorithm implementations

3. **systems-architect** - Architecture validation

   - Validate modular tool design
   - Check component boundaries
   - Verify data flow patterns
   - Assess integration architecture
   - Verify agentic loop architecture (`internal/agent/agentic/`)

4. **qa-test-engineer** - Testing verification

   - Check test coverage for each tool
   - Validate integration tests
   - Verify exploit validation tests
   - Check multi-chain test scenarios

5. **technical-docs-writer** - Documentation
   - Create validation report
   - Document compliance gaps
   - Generate implementation matrix

## A1-ASCEG Six Tools Validation

### Tool 1: Source Code Fetcher (Section IV-A)

**Requirements from A1 Paper:**

- Resolves proxy contract relationships through bytecode pattern analysis
- Implementation slot examination (EIP-1967 and custom patterns)
- Maintains temporal consistency at specific historical blocks
- Accesses actual executable logic, not proxy interfaces

**Validation Checks:**

```bash
# Check proxy resolution implementation
rg "EIP1967|proxy|implementation" internal/blockchain --type go

# Verify bytecode analysis
rg "GetCode|bytecode" internal/blockchain --type go

# Check temporal consistency (block-specific fetching)
rg "BlockNumber|AtBlock" internal/blockchain --type go

# Verify Etherscan/BSCScan integration
rg "Etherscan|BSCScan|SourceCode" internal/blockchain --type go
```

**Expected Implementation:**

- [ ] Proxy detection logic in `internal/blockchain/proxy.go`
- [ ] Implementation address resolution
- [ ] Support for multiple proxy patterns
- [ ] Block-specific source fetching
- [ ] Verified source code retrieval

### Tool 2: Constructor Parameter Extractor (Section IV-A)

**Requirements from A1 Paper:**

- Analyzes deployment transaction calldata
- Reconstructs initialization parameters
- Provides configuration context (token addresses, fees, access control)

**Validation Checks:**

```bash
# Check constructor parameter extraction
rg "ConstructorArgs|deployment|initialization" internal/blockchain --type go

# Verify transaction analysis
rg "GetTransactionByHash|creation" internal/blockchain --type go

# Check ABI decoding
rg "UnpackValues|DecodeArgs" internal/blockchain --type go
```

**Expected Implementation:**

- [ ] Deployment transaction fetching
- [ ] Constructor argument decoding
- [ ] Initialization parameter extraction
- [ ] Support for proxy initialization

### Tool 3: State Reader (Section IV-A)

**Requirements from A1 Paper:**

- Performs ABI analysis for public/external view functions
- Captures contract state snapshots at target blocks
- Batch calls for efficiency

**Validation Checks:**

```bash
# Check state reading implementation
rg "CallContract|eth_call" internal/blockchain --type go

# Verify batch calling
rg "BatchCall|Multicall" internal/blockchain --type go

# Check view function identification
rg "view|constant|pure" internal/blockchain --type go
```

**Expected Implementation:**

- [ ] ABI parsing for view functions
- [ ] State snapshot at specific blocks
- [ ] Batch call optimization
- [ ] Comprehensive state collection

### Tool 4: Code Sanitizer (Section IV-A)

**Requirements from A1 Paper:**

- Eliminates non-essential elements (comments, unused code)
- Removes extraneous library dependencies
- Focuses analysis on executable logic

**Validation Checks:**

```bash
# Check code sanitization
rg "Sanitize|Clean|Strip" internal/analyzer --type go

# Verify comment removal
rg "RemoveComments|StripComments" internal/analyzer --type go

# Check unused code elimination
rg "UnusedCode|DeadCode" internal/analyzer --type go
```

**Expected Implementation:**

- [ ] Comment stripping logic
- [ ] Unused code detection
- [ ] Library dependency analysis
- [ ] Code flattening/simplification

### Tool 5: Concrete Execution Tester (Section IV-C)

**Requirements from A1 Paper:**

- Built on Forge for deterministic blockchain simulation
- Instantiates blockchain forks at targeted blocks
- Comprehensive execution analytics (traces, gas, state transitions)

**Validation Checks:**

```bash
# Check Forge integration
rg "forge test|ForgeExecutor" internal/forge --type go

# Verify fork creation
rg "fork-url|fork-block-number" internal/forge --type go

# Check execution tracing
rg "trace|vvvvv|debug" internal/forge --type go

# Verify exploit compilation and testing
rg "CompileSolidity|DeployContract" internal/forge --type go
```

**Expected Implementation:**

- [ ] Forge executor implementation
- [ ] Fork management at specific blocks
- [ ] Exploit compilation pipeline
- [ ] Execution result parsing
- [ ] Gas and trace analysis

### Tool 6: Revenue Normalizer (Section IV-D)

**Requirements from A1 Paper:**

- Token balance normalization
- Optimal DEX routing for conversion
- Surplus/deficit resolution
- Economic performance quantification

**Validation Checks:**

```bash
# Check revenue normalization
rg "Normalize|Revenue|Profit" internal/analyzer --type go

# Verify DEX integration
rg "Uniswap|PancakeSwap|DEX|swap" internal/dex --type go

# Check token conversion logic
rg "ConvertTo|SwapTo|BaseToken" internal/dex --type go

# Verify profit calculation
rg "CalculateProfit|NetRevenue" internal/analyzer --type go
```

**Expected Implementation:**

- [ ] Initial state normalization
- [ ] Post-execution reconciliation
- [ ] DEX router abstraction (DexUtils)
- [ ] Multi-hop routing support
- [ ] Profit calculation in ETH/BNB

## Iterative Agent Loop Validation

### Requirements from A1 Paper (Section IV-B):

- Agent autonomously decides approach based on tools and context
- Maintains history of previous attempts
- Integrates execution feedback (profitability, traces, revert reasons)
- Evolves understanding through iterations
- Constrained output format (Solidity code blocks)

**Validation Checks:**

```bash
# Check iterative loop implementation
rg "Iteration|Attempt|Loop" internal/agent --type go

# Verify feedback integration
rg "Feedback|ExecutionResult|Revert" internal/agent --type go

# Check LLM prompt management
rg "Prompt|SystemPrompt|UserPrompt" internal/llm --type go

# Verify attempt history tracking
rg "History|Previous|Attempts" internal/agent --type go
```

**Expected Implementation:**

- [ ] Main agent loop (up to 5 iterations)
- [ ] Feedback processing from execution
- [ ] Prompt construction with context
- [ ] History management
- [ ] Strategy adaptation logic

## Multi-Chain Support Validation

### Requirements from A1 Paper:

- Ethereum and BSC mainnet support
- Chain-specific configurations (WETH vs WBNB)
- Appropriate DEX routing (Uniswap vs PancakeSwap)

**Validation Checks:**

```bash
# Check multi-chain architecture
rg "ChainID|Ethereum|BSC|Binance" internal/blockchain --type go

# Verify chain-specific handling
rg "WETH|WBNB|BaseToken" internal/blockchain --type go

# Check RPC provider management
rg "RPCClient|Provider|Endpoint" internal/blockchain --type go
```

**Expected Implementation:**

- [ ] Chain abstraction interface
- [ ] Ethereum support (chain ID: 1)
- [ ] BSC support (chain ID: 56)
- [ ] Chain-specific token handling
- [ ] Provider fallback logic

## Output Format Validation

### Requirements from A1 Paper:

- Exploit code in Solidity code blocks
- Structured execution reports
- Economic validation metrics

**Validation Checks:**

````bash
# Check output formatting
rg "solidity|```|CodeBlock" internal/report --type go

# Verify report generation
rg "Report|Output|Result" internal/report --type go

# Check JSON/structured output
rg "json.Marshal|Export" internal/report --type go
````

## Validation Report Template

````markdown
# A1-ASCEG Specification Compliance Report

## Executive Summary

- **Overall Compliance**: X/100%
- **Critical Gaps**: [List any missing core components]
- **Recommendation**: [PASS/FAIL for production readiness]

## Tool Implementation Status

### 1. Source Code Fetcher

✅/❌ Proxy resolution
✅/❌ Temporal consistency
✅/❌ Source verification
**Compliance**: X%

### 2. Constructor Parameter Extractor

✅/❌ Deployment analysis
✅/❌ Parameter decoding
✅/❌ Initialization support
**Compliance**: X%

### 3. State Reader

✅/❌ ABI analysis
✅/❌ State snapshots
✅/❌ Batch optimization
**Compliance**: X%

### 4. Code Sanitizer

✅/❌ Comment removal
✅/❌ Dead code elimination
✅/❌ Focus on executable logic
**Compliance**: X%

### 5. Concrete Execution Tester

✅/❌ Forge integration
✅/❌ Fork management
✅/❌ Execution analytics
**Compliance**: X%

### 6. Revenue Normalizer

✅/❌ Balance normalization
✅/❌ DEX routing
✅/❌ Profit calculation
**Compliance**: X%

## Core Features Status

### Iterative Agent Loop

✅/❌ Iteration management
✅/❌ Feedback integration
✅/❌ Strategy evolution
**Compliance**: X%

### Multi-Chain Support

✅/❌ Ethereum support
✅/❌ BSC support
✅/❌ Chain abstraction
**Compliance**: X%

### Output Formatting

✅/❌ Solidity code blocks
✅/❌ Structured reports
✅/❌ Economic metrics
**Compliance**: X%

## Detailed Findings

### Critical Issues (Blocks Production)

1. [Issue description and location]
2. [Issue description and location]

### Major Issues (Should Fix)

1. [Issue description and location]
2. [Issue description and location]

### Minor Issues (Nice to Have)

1. [Issue description and location]
2. [Issue description and location]

## Implementation Verification Commands

```bash
# Run comprehensive tests
make test-race

# Check test coverage
make coverage

# Verify Forge integration
forge --version

# Test exploit generation
quanta run --chain ethereum --address 0x... --block 17000000

# Validate multi-chain
quanta blockchain detect 0x...
```
````

## Recommendations

### Immediate Actions

1. [Critical fixes needed]
2. [Missing implementations]

### Before Production

1. [Required improvements]
2. [Testing requirements]

### Future Enhancements

1. [Additional features]
2. [Optimization opportunities]

## Advanced Components Validation

### Feedback-Driven Exploitation

```bash
# Verify TraceAnalyzer implementation
rg "FailureRevert|FailureOutOfGas|FailureBalance|FailureAccess|FailureNoProfit|FailureRepayment" internal/agent --type go

# Verify FeedbackOrchestrator loop
rg "FeedbackOrchestrator|TraceFeedback|SuggestedFix" internal/agent --type go

# Verify BestAtNExecutor
rg "BestAtNExecutor|bestOf|parallelAttempts" internal/agent --type go
```

- [ ] 6 FailureType constants defined
- [ ] TraceFeedback struct with severity ranking
- [ ] FeedbackOrchestrator with configurable iteration limit
- [ ] BestAtNExecutor with parallel execution and model rotation

### Structured Planning

```bash
# Verify structured plan types
rg "AttackPlan|PlanActor|AttackStep|Extraction" internal/coordinator --type go

# Verify plan parsing pipeline
rg "PlanParser|PlanValidator|FoundryGenerator" internal/coordinator --type go

# Verify structured planner backend
rg "StructuredPlannerBackend|GeneratePlan" internal/coordinator --type go
```

- [ ] AttackPlan, PlanActor, AttackStep, Extraction structs
- [ ] PlanParser with JSON extraction and repair
- [ ] PlanValidator with sentinel errors (ErrInvalidPlan, ErrMissingActor, etc.)
- [ ] FoundryGenerator producing vm.deal/vm.startPrank code
- [ ] StructuredPlannerBackend orchestrating the pipeline

### Enhanced RAG

```bash
# Verify body-aware scanning
rg "SolidityParser|SolidityElement|ScanBodies|ScanStatePatterns" internal/knowledge --type go

# Verify chunking
rg "Chunker|ChunkText|ChunkElements" internal/knowledge --type go

# Verify index versioning
rg "metadata.json|Version|v2" internal/knowledge --type go
```

- [ ] SolidityParser extracting function bodies and modifiers
- [ ] Chunker with 512-token max and 64-token overlap
- [ ] Index versioning (v1 legacy, v2 body-aware)
- [ ] EnhancedRetriever merging body-aware and signature-based results

### Category-Specific Ensemble

```bash
# Verify category types
rg "VulnCategory|CategoryAccessControl|CategoryReentrancy" internal/agent --type go

# Verify two-stage voting
rg "CategoryCalibration|SetCategoryWeights|TopCategory|CategoryConfidence" internal/agent --type go
```

- [ ] 8 VulnCategory constants
- [ ] CategoryCalibration struct with per-category weights
- [ ] Two-stage voting (global → category refinement)
- [ ] VoteResult with TopCategory and CategoryConfidence fields

### Flash Loan Synthesis

```bash
# Verify flash loan selection
rg "FlashLoanSelector|FlashLoanConfig" internal/coordinator --type go

# Verify chain-aware defaults
rg "chainDefaults|dydx|balancer|aave" internal/coordinator --type go
```

- [ ] FlashLoanSelector interface with auto-selection
- [ ] Chain-aware provider defaults
- [ ] Fee-aware profit calculation
- [ ] CLI flags: --flash-loan-provider, --uniswap-v3-pool

## Compliance Score Calculation

| Component             | Weight | Score | Weighted  |
| --------------------- | ------ | ----- | --------- |
| Source Fetcher        | 12%    | X/100 | X         |
| Constructor Extractor | 8%     | X/100 | X         |
| State Reader          | 12%    | X/100 | X         |
| Code Sanitizer        | 8%     | X/100 | X         |
| Execution Tester      | 15%    | X/100 | X         |
| Revenue Normalizer    | 8%     | X/100 | X         |
| Agent Loop            | 8%     | X/100 | X         |
| Multi-Chain           | 4%     | X/100 | X         |
| Output Format         | 5%     | X/100 | X         |
| Feedback Exploitation | 5%     | X/100 | X         |
| Structured Planning   | 5%     | X/100 | X         |
| Enhanced RAG          | 4%     | X/100 | X         |
| Category Ensemble     | 3%     | X/100 | X         |
| Flash Loan Synthesis  | 3%     | X/100 | X         |
| **TOTAL**             | 100%   | -     | **X/100** |

## Certification

**A1-ASCEG Compliance Level**: [FULL/PARTIAL/MINIMAL]

**Production Ready**: [YES/NO]

**Auditor Notes**: [Additional observations and recommendations]

````

## Execution Instructions

**IMPORTANT**: The product-manager-orchestrator should:

1. **Think carefully** about the validation approach
2. Coordinate all specialist agents systematically
3. Ensure thorough coverage of all A1 requirements
4. Generate comprehensive documentation
5. Provide actionable recommendations

The validation must be **evidence-based** - every claim should be backed by:
- Code inspection results
- Test execution outcomes
- Documentation review
- Actual command execution

## Success Criteria

The validation is successful when:

1. All 6 A1-ASCEG tools are properly implemented
2. Iterative agent loop with feedback is functional
3. Multi-chain support (ETH/BSC) is operational
4. Output formatting matches specification
5. Comprehensive tests exist and pass
6. Documentation is complete

## Usage

```bash
/validate-a1-asceg
````

This will trigger a comprehensive validation of Quanta against the A1-ASCEG specification, producing a detailed compliance report and actionable recommendations.
