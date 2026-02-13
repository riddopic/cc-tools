---
description: Create comprehensive PRP with deep research for Go feature implementation
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Task
  - WebFetch
  - WebSearch
  - AskUserQuestion
  - TaskCreate
  - TaskUpdate
  - TaskList
  - mcp__context7__resolve-library-id
  - mcp__context7__query-docs
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "[feature-file-path]"
model: opus
---

# Create PRP for Go Feature

## Required Skills

This command uses the following skills (auto-loaded based on context):
- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `testing-patterns` - For table-driven tests and mocking
- `interface-design` - For interface-first design
- `cli-development` - For Cobra/Viper CLI patterns
- `concurrency-patterns` - For goroutines and channels

**PRP File Path**: `docs/PRPs/{feature-name}.md`

## YOU MUST DO IN-DEPTH RESEARCH, FOLLOW THE <RESEARCH PROCESS>

<RESEARCH PROCESS>

- Don't only research one page, and don't only use your own web scraping tool - instead scrape many relevant pages from all documentation links mentioned, use Augments, Ref, Context 7 and Deep Wiki MCP servers to get additional documentation.
- Take my tech as sacred truth, for example if I say a model name then research that model name for LLM usage - don't assume from your own knowledge at any point
- When I say don't just research one page, I mean do incredibly in-depth research, like to the point where it's just absolutely ridiculous how much research you've actually done, then when you create the PRP document you need to put absolutely everything into that including references to the .md files you put inside the `docs/research/` directory so any AI can pick up your PRP and generate WORKING and COMPLETE production ready code.

</RESEARCH PROCESS>

## Feature file: $ARGUMENTS

Generate a complete PRP for Go feature implementation with deep and thorough research, following quanta's strict Go idioms and coding standards. Ensure rich context is passed to the AI through the PRP to enable one-pass implementation success through self-validation and iterative refinement.

The AI agent only gets the context you are appending to the PRP and training data. Assume the AI agent has access to the codebase and the same knowledge cutoff as you, so it's important that your research findings are included or referenced in the PRP. Include all Go-specific patterns, idiomatic Go approaches, and testing patterns. The Agent has Web search capabilities, so pass URLs to Go, standard library, and third-party library documentation.

## Research Process

> During the research process, create clear tasks and spawn as many agents and subagents as needed using the batch tools. The deeper research we do here the better the PRP will be. We optimize for chance of success and not for speed.

1. **Go Codebase Analysis**

   - Create clear todos and spawn subagents to search the codebase for similar features/patterns. Think hard and plan your approach
   - Identify all the necessary files to reference in the PRP
   - Note all existing conventions to follow
   - Review project guidelines:
     - `docs/CODING_GUIDELINES.md` - Main coding standards
     - `docs/examples/patterns/` - Implementation patterns:
       - `concurrency.md` - Goroutines, channels, worker pools
       - `cli.md` - Cobra/Viper patterns for CLI
       - `testing.md` - Table-driven tests with testify
       - `mocking.md` - Interface-based mocking
     - `docs/examples/standards/` - Coding standards:
       - `documentation.md` - Godoc comment standards
       - `interfaces.md` - Interface design principles
       - `go-specific.md` - Go idioms and best practices
   - Check existing test patterns using table-driven tests
   - Note interface design and dependency injection patterns
   - Review error handling patterns (wrapping, quanta errors)
   - Use the batch tools to spawn subagents to search the codebase for similar features/patterns

2. **External Research at Scale**

   - Create clear todos and spawn with instructions subagents to do deep research for similar features/patterns online and include URLs to documentation and examples
   - Go standard library documentation (pkg.go.dev)
   - Third-party library documentation (Cobra, Viper, Zap, etc.)
   - For critical pieces of documentation add a .md file to `docs/research` and reference it in the PRP with clear reasoning and instructions
   - Implementation examples (GitHub/StackOverflow/blogs)
   - Best practices and common pitfalls found during research
   - Use the batch tools to spawn subagents to search for similar features/patterns online

3. **Go External Research**

   - Go documentation (go.dev/doc)
   - Effective Go guide
   - Go Code Review Comments
   - Cobra CLI framework documentation
   - Viper configuration management
   - Standard library packages relevant to feature
   - Concurrency patterns if applicable
   - Performance profiling tools (pprof)

4. **Architecture Review**

   - Read `docs/CODING_GUIDELINES.md` for project-specific requirements
   - Review project structure (`cmd/`, `internal/`, `pkg/`)
   - Study module structure (`go.mod`, `go.sum`)
   - Check `golangci-lint` configuration
   - Review `Makefile` for available tasks
   - Understand internal package privacy

5. **Refinement Interview**

   Before writing the PRP, conduct a structured refinement interview using AskUserQuestion to surface ambiguities and design decisions:

   - **Scope boundaries**: What's in scope vs out of scope? Any edge cases to handle or explicitly exclude?
   - **Design decisions**: When multiple approaches exist (e.g., sync vs async, polling vs events), present options with trade-offs and get a decision
   - **Integration points**: How should this feature connect to existing commands, packages, or workflows?
   - **Priority trade-offs**: If implementation is large, which tasks are MVP vs nice-to-have?

   This interview ensures the PRP captures the user's intent accurately and reduces rework during execution.

6. **User Clarification**
   - Ask for any remaining clarification needed

## PRP Generation

Using `docs/PRPs/templates/PRP_base.md` as template:

### Critical Go Context to Include in the PRP

- **Documentation**:

  - Go language specification
  - Standard library packages (context, io, fmt, etc.)
  - Cobra command structure and patterns
  - Viper configuration management
  - Table-driven testing patterns
  - Benchmark writing guidelines
  - Race condition detection

- **Code Examples**:

  - Interface definitions from interfaces.md
  - Concurrency patterns from concurrency.md
  - CLI command structure from cli.md
  - Table-driven test examples
  - Mock setup patterns
  - Error wrapping examples
  - Context propagation patterns
  - Proper defer usage

- **Go Gotchas**:

  - Never ignore errors (no \_ = someFunc())
  - Always use context for cancellation
  - Close channels only from sender side
  - Check for nil before dereferencing
  - Avoid goroutine leaks
  - No panic for normal error handling
  - Use defer for cleanup immediately after resource acquisition
  - Avoid naked returns in long functions

- **Patterns**:

  - Accept interfaces, return concrete types
  - Small, focused interfaces (Interface Segregation)
  - Functional options for optional parameters
  - Worker pools for concurrent processing
  - Context for cancellation and timeouts
  - quanta errors for known error conditions
  - Table-driven tests for comprehensive testing

- **Best Practices**:
  - Early returns to reduce nesting
  - Preallocation of slices when size is known
  - strings.Builder for string concatenation
  - sync.Pool for frequently allocated objects
  - Consistent receiver types (all pointer or all value)

### Implementation Blueprint

- Start with interface definitions
- Define error types and quanta errors
- Show package structure with internal/
- Include table-driven test examples
- Reference pattern files for implementation
- List tasks following Go idioms in the order they should be completed
- Use dependency injection for testability

### Task Structure Requirements

Each task in the PRP MUST have:
- **files**: List of files touched by this task
- **depends_on**: Task dependencies (e.g., `[Task 1, Task 2]`) — maps to `blockedBy` during execution
- **acceptance**: Per-task acceptance criteria (testable conditions)
- **commit**: Suggested atomic commit message following conventional commits format
- **description**: What to create/modify and how

This enables the `execute-prp` command to create TaskCreate items with proper dependency mapping and atomic commits per task.

### Validation Gates (Must be Executable) for Go

```bash
# Format check and fix
make fmt

# Run comprehensive linting
make lint

# Run tests with race detector
make test-race

# Generate test coverage report
make coverage
# Check for test coverage threshold (e.g., 80%)
go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if ($1 < 80) exit 1}'

# Run benchmarks
make bench

# Check for go.mod tidiness
make tidy
git diff --exit-code go.mod go.sum

# Build validation
make build

# Check for TODO comments without issue references
! rg "TODO" --type go | grep -v "TODO([a-zA-Z0-9]*)"

# Verify no shadowed variables
go vet -shadow ./...

# Check for inefficient assignments
ineffassign ./...

# Verify all exported items have comments
golint ./... | grep -c "exported" | test $(cat) -eq 0

# Security check with gosec
gosec -quiet ./...

# Check for nil pointer dereferences
staticcheck -checks="SA5011" ./...

# Validate interfaces are small (no more than 5 methods)
for file in $(find . -name "*.go"); do
  awk '/^type .* interface {/,/^}/' "$file" | grep -c "^\s*[A-Z]" | while read count; do
    if [ "$count" -gt 5 ]; then
      echo "Interface too large in $file"
      exit 1
    fi
  done
done

# Check context is first parameter in functions
rg "func.*context.Context" --type go | grep -v "ctx context.Context" | grep -v "context.Context)" | test $(wc -l) -eq 0
```

The more validation gates the better, but make sure they are executable by the AI agent.
Include tests, build validation, linting, race detection, and any other relevant validation gates. Get creative with the validation gates.

**_ CRITICAL AFTER YOU ARE DONE RESEARCHING AND EXPLORING THE CODEBASE BEFORE YOU START WRITING THE PRP _**

**_ PLAN YOUR APPROACH IN DETAILED TODOS THEN START WRITING THE PRP _**

### CRITICAL: Research and Planning Phase

**After completing codebase research and exploration, BEFORE writing the PRP:**

- Thoroughly plan your approach
- Consider all architectural implications following Go idioms
- Map out the complete implementation strategy
- Ensure comprehensive context inclusion
- Plan interface boundaries and package structure

## Output

Save as: `docs/PRPs/{feature-name}.md`

## Quality Checklist

- [ ] All necessary Go context included with standard library references
- [ ] Interface-first design with small, focused interfaces
- [ ] Table-driven tests defined with comprehensive test cases
- [ ] Validation gates test Go-specific requirements (formatting, linting, race detection)
- [ ] References pattern files from docs/examples/patterns/
- [ ] Error handling with proper wrapping and context
- [ ] Concurrency patterns if applicable (no goroutine leaks)
- [ ] Context propagation for cancellation
- [ ] Benchmark tests for performance-critical code
- [ ] All exported items have godoc comments
- [ ] Internal packages used for private implementation
- [ ] Follows project structure (cmd/, internal/, pkg/)
- [ ] Dependency injection for testability
- [ ] No magic numbers (use named constants)
- [ ] Proper resource cleanup with defer

## Go-Specific PRP Requirements

1. **Interface Design**: Define interfaces before implementations
2. **Error Handling**: Use error wrapping with context, define quanta errors
3. **Testing**: Table-driven tests with subtests (t.Run)
4. **Documentation**: All exported items must have godoc comments
5. **Performance**: Include benchmarks for critical paths
6. **Concurrency**: Document goroutine lifecycle and cancellation
7. **Package Design**: Use internal/ for private packages
8. **CLI Structure**: Follow Cobra command patterns if applicable

Score the PRP on a scale of 1-10 (confidence level to succeed in one-pass implementation using Claude Code)

Remember: The goal is one-pass implementation success through comprehensive Go context, following all idiomatic Go patterns and project guidelines. Simplicity and clarity are paramount in Go.
