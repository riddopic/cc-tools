---
description: Validate component implementation quality against project standards
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
  - AskUserQuestion
model: opus
skills:
  - code-review
  - go-coding-standards
---

# Validate Component Implementation Quality

## Component: $ARGUMENTS

Validate the implementation quality of a specific component in the Quanta project against PRD requirements, coding standards, and TDD principles. This command ensures components are production-ready before deployment.

**Output Files**:

- Validation Report: `docs/USER-GUIDE/{COMPONENT}.md` (created/updated)
- Issues Report: `docs/validation/{component}-validation.md` (if issues found)

**Supported Components**:

- `blockchain` - Blockchain command with subcommands (detect, fetch, verify)
- `dex` - DEX integration command
- `analyze` - Smart contract analysis command
- `doctor` - System diagnostics command
- `export` - Export functionality command
- `report` - Report generation command
- `run` - Main exploit generation command
- `target` - Target configuration command
- `all` - Validate all components

Example usage:

- `/validate-command blockchain` - Validates blockchain command implementation
- `/validate-command all` - Validates all components

## Agent Orchestration Strategy

**IMPORTANT**: Think carefully about this, and instruct all sub-agents to do the same.

Use the **product-manager-orchestrator** to coordinate specialized agents for comprehensive validation:

### Primary Analysis Agents

1. **code-analyzer-debugger** - Implementation verification

   - Verify actual implementation exists (not mocks/stubs)
   - Check command structure matches PRD requirements
   - Validate subcommand implementations
   - Identify missing functionality

2. **qa-test-engineer** - Test coverage validation

   - Verify TDD compliance (tests written first)
   - Check test coverage (≥80% for units)
   - Validate table-driven test structure
   - Ensure integration tests exist
   - Check for race condition tests

3. **code-review-specialist** - Go code quality

   - Verify Go idioms and patterns
   - Check interface design principles
   - Validate error handling patterns
   - Ensure proper context usage
   - Review package structure

4. **security-threat-analyst** - Security validation

   - Check credential handling
   - Validate input sanitization
   - Review proxy/Tor integration security
   - Verify no hardcoded secrets
   - Check for injection vulnerabilities

5. **technical-docs-writer** - Documentation management
   - Create/update USER-GUIDE documentation
   - Document command usage and examples
   - Include configuration options
   - Add troubleshooting sections

### Supporting Agents

6. **deep-research-specialist** - Requirements verification

   - Cross-reference with `docs/A1-ASCEG-PRD.md`
   - Verify against sprint documentation
   - Check `docs/MVP-SPECIFICATION.md` compliance

7. **systems-architect** - Architecture compliance

   - Validate component boundaries
   - Check interface segregation
   - Verify dependency injection patterns

8. **performance-optimizer** - Performance validation
   - Check for benchmarks
   - Validate concurrent operations
   - Review resource management

**IMPORTANT**: The **product-manager-orchestrator** coordinates all validation work but DOES NOT make code changes. All fixes must be delegated to appropriate specialist agents.

## Validation Process

### Step 1: Component Discovery

Identify all files related to the component:

```bash
# Command implementation files
find cmd -name "${component}*.go" -type f | grep -v _test.go

# Internal implementation
find internal -path "*/${component}/*" -name "*.go" | grep -v _test.go

# Test files
find . -name "*${component}*_test.go" -type f

# Interface definitions
rg "type.*${component}.*interface" --type go
```

### Step 2: PRD Requirements Validation

Check against `docs/A1-ASCEG-PRD.md` requirements:

#### For 'run' command:

- ✓ LLM-powered security agent integration
- ✓ Multi-chain live data support (Ethereum/BSC)
- ✓ Six domain-specific tools integrated
- ✓ Iterative agent reasoning with feedback
- ✓ Structured output and reporting
- ✓ Performance and parallelization

#### For 'blockchain' command:

- ✓ Multi-chain support (detect, fetch, verify)
- ✓ Proxy pattern detection
- ✓ Source code fetching
- ✓ State verification

#### For 'analyze' command:

- ✓ Static analysis integration
- ✓ Vulnerability detection
- ✓ Report generation

### Step 3: Implementation Quality Checks

#### TDD Verification

```bash
# Check test files exist before implementation
git log --follow --diff-filter=A -- "*${component}*.go" | head -20
git log --follow --diff-filter=A -- "*${component}*_test.go" | head -20

# Verify test-first development
# Tests should have earlier timestamps than implementation
```

#### Code Quality Checks

```bash
# Format verification
gofmt -l cmd/${component}*.go internal/*/${component}/*.go
test -z "$(gofmt -l cmd/${component}*.go)"

# Import organization
goimports -l cmd/${component}*.go
test -z "$(goimports -l cmd/${component}*.go)"

# Comprehensive linting
golangci-lint run --timeout=5m cmd/${component}*.go internal/*/${component}/*.go

# Run component tests with race detector
go test -v -race ./cmd -run "Test.*${component}.*"
go test -v -race ./internal/... -run "Test.*${component}.*"

# Check test coverage
go test -coverprofile=coverage.out ./cmd ./internal/...
go tool cover -func=coverage.out | grep -i ${component}
```

#### Security Checks

```bash
# Check for hardcoded secrets
! rg -i "(api[_-]?key|secret|password|token)\s*[:=]\s*['\"]" cmd/${component}*.go

# Verify secure credential handling
rg "viper\.(Get|Set)" cmd/${component}*.go | grep -E "(key|secret|token|password)"

# Check input validation
rg "cobra\..*StringVar|IntVar" cmd/${component}*.go -A 5 | grep -i "validate"
```

### Step 4: Documentation Generation

Create/update `docs/USER-GUIDE/{COMPONENT}.md`:

**_ PLAN YOUR APPROACH TO REPORT AND DOCUMENTATION GENERATION IN DETAILED TODOS THEN START WRITING THE DOCUMENT _**

````markdown
# {COMPONENT} Command Guide

## Overview

[Description of what the command does, extracted from implementation]

## Installation

```bash
quanta {component} [subcommands] [flags]
```
````

## Usage

### Basic Usage

```bash
# Example commands
quanta {component} --flag value
```

### Advanced Usage

[Complex examples with multiple flags]

## Configuration

### Environment Variables

- `QUANTA_{COMPONENT}_*` - Component-specific variables

### Configuration File

```yaml
{ component }:
  option1: value1
  option2: value2
```

## Subcommands

[If applicable, list all subcommands with descriptions]

## Flags

[Table of all flags with descriptions and defaults]

## Examples

### Example 1: [Use Case]

```bash
quanta {component} [specific example]
```

### Example 2: [Another Use Case]

```bash
quanta {component} [another example]
```

## Troubleshooting

### Common Issues

#### Issue 1: [Description]

**Solution**: [How to fix]

#### Issue 2: [Description]

**Solution**: [How to fix]

## Security Considerations

[Any security-related notes]

## Performance Tips

[Performance optimization suggestions]

## See Also

- [Related commands]
- [Related documentation]

````

### Step 5: Scoring and Reporting

## Validation Scoring Rubric (1-10 scale)

### Implementation Completeness (3 points)
- [ ] All PRD features implemented (1.5 points)
- [ ] No mock implementations (1.5 points)

### Test Coverage (2 points)
- [ ] TDD compliance verified (1 point)
- [ ] Coverage ≥80% (1 point)

### Code Quality (2 points)
- [ ] Go idioms followed (1 point)
- [ ] Clean golangci-lint output (1 point)

### Documentation (1 point)
- [ ] USER-GUIDE created/updated (0.5 points)
- [ ] Godoc comments complete (0.5 points)

### Security (1 point)
- [ ] No hardcoded secrets (0.5 points)
- [ ] Input validation present (0.5 points)

### Architecture (1 point)
- [ ] Proper package structure (0.5 points)
- [ ] Interface-driven design (0.5 points)

## Output Format

### Validation Report (docs/USER-GUIDE/{COMPONENT}.md)
[Documentation as shown in Step 4]

### Issues Report (if score < 8)
```markdown
# {COMPONENT} Validation Report

## Score: X/10

### Component Files Analyzed
- cmd/{component}.go
- cmd/{component}_*.go
- internal/{component}/*.go
- Tests: X files, Y% coverage

### PRD Compliance
✅ Features Implemented:
- Feature 1
- Feature 2

❌ Missing Features:
- Feature 3 (required by PRD section X.Y)

### Code Quality Issues

**Critical** (blocks deployment):
- ❌ Mock implementation found in {file}:{line}
- ❌ Missing error handling in {function}

**Major** (should fix):
- ❌ Missing godoc on exported function
- ❌ Interface too large (X methods)

**Minor** (consider fixing):
- ⚠️ Could improve test table structure
- ⚠️ Consider adding benchmarks

### Security Issues
- ❌ Potential SQL injection in {file}:{line}
- ⚠️ Consider rate limiting for {endpoint}

### Recommended Actions

1. **Immediate**:
   - Implement missing {feature}
   - Fix security vulnerability in {location}

2. **Before Release**:
   - Add integration tests
   - Complete documentation

3. **Technical Debt**:
   - Refactor {component} for better testability
   - Add performance benchmarks

### Commands to Fix
```bash
# Fix formatting
make fmt

# Run tests
make test

# Check coverage
make coverage
````

````

## Component-Specific Validation

### blockchain Command
- Verify detect, fetch, verify subcommands
- Check proxy pattern handling
- Validate multi-chain support

### run Command
- Verify LLM integration
- Check Forge execution environment
- Validate iterative loop implementation

### analyze Command
- Check vulnerability detection logic
- Verify Solidity parsing
- Validate report generation

### target Command
- Check configuration persistence
- Verify validation logic
- Test multi-chain support

## Execution Flow

1. Parse component argument
2. Run code-analyzer-debugger for implementation discovery
3. Run qa-test-engineer for test validation
4. Run code-review-specialist for quality checks
5. Run security-threat-analyst for security validation
6. Run technical-docs-writer to create/update documentation
7. Generate validation report with scoring
8. If issues found, create detailed fix recommendations

## Usage Examples

```bash
# Validate blockchain command
/validate-command blockchain

# Validate all components
/validate-command all

# Output locations:
# - Documentation: docs/USER-GUIDE/BLOCKCHAIN.md
# - Issues: docs/validation/blockchain-validation-20250106.md
````

## AI Validation Instructions

**IMPORTANT**: After running validation checks, actively search for issues:

1. **Check for Mock Implementations**:

   - Search for "mock", "stub", "TODO", "FIXME" in implementation files
   - Verify actual functionality exists, not placeholders

2. **Verify PRD Alignment**:

   - Cross-reference each command with `docs/A1-ASCEG-PRD.md` requirements
   - Check sprint documentation for additional requirements

3. **Validate TDD Compliance**:

   - Use git history to verify tests were written before implementation
   - Check that tests drive the design, not vice versa

4. **Security Deep Dive**:

   - Look for any credential handling
   - Check all user inputs are validated
   - Verify no sensitive data in logs

5. **Documentation Quality**:
   - Ensure USER-GUIDE is comprehensive
   - Verify examples actually work
   - Check troubleshooting covers real issues

**Final Step**: After validation, if score < 8, generate a PRP for fixes:

- Save as: `docs/PRPs/{component}-validation-fixes.md`
- Include specific tasks for each issue
- Prioritize by severity (Critical > Major > Minor)

The validation should be thorough, constructive, and actionable. Help teams understand not just what's wrong, but why it matters and how to fix it properly.
