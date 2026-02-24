---
description: Execute a specific task from a PRP using sub-agent orchestration
allowed-tools:
  - Read
  - Grep
  - Glob
  - Task
  - Bash
  - TaskCreate
  - TaskUpdate
  - TaskList
  - AskUserQuestion
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "<prp-name> <task-identifier>"
model: opus
---

# Execute PRP Task

## Arguments: $ARGUMENTS

Parse the input arguments as follows:

| Component | Description | Example |
| ----------- | ------------- | --------- |
| **PRP Name** | First token - the PRP filename without extension | `feedback-driven-exploitation` |
| **Task Identifier** | Remaining tokens - the task to execute | `Task 1`, `Task 4`, `Task 10` |

**PRP File Path**: `docs/PRPs/{prp-name}.md`

### Parsing Examples

| Input | PRP Path | Target Task |
| ------- | ---------- | ------------- |
| `feedback-driven-exploitation Task 1` | `docs/PRPs/feedback-driven-exploitation.md` | Task 1 |
| `feedback-driven-exploitation Task 4` | `docs/PRPs/feedback-driven-exploitation.md` | Task 4 |
| `multi-model-ensemble Task 2` | `docs/PRPs/multi-model-ensemble.md` | Task 2 |

### Task Extraction

When loading the PRP, extract ONLY the specified task section:

1. Find the heading matching `### Task N:` (where N is the task number)
2. Extract content from that heading until the next `### Task` heading or `---` separator
3. Focus implementation on that single task's requirements

## Required Skills

This command uses the following skills (auto-loaded based on context):

- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `testing-patterns` - For table-driven tests and mocking

**Objective**: Implement a **single task** from a PRP file using intelligent sub-agent orchestration with strict adherence to quanta's Go coding standards and idiomatic patterns.

**Scope**: This command executes ONE specific task from the PRP, not the entire PRP. Use this for incremental task-by-task execution.

## CRITICAL RULE: Claude Code MUST NOT Write Code

**🚨 MANDATORY**: Claude Code (the primary assistant) acts ONLY as an orchestrator and MUST delegate ALL coding tasks to specialized sub-agents. If you find yourself about to write code, STOP and delegate to the appropriate sub-agent instead.

## Workflow Overview

This command implements a comprehensive PRP execution workflow using specialized sub-agents:

1. **Analysis Phase**: Load PRP and analyze requirements (research agents only)
2. **Planning Phase**: Create implementation strategy following Go idioms (orchestrator only)
3. **Implementation Phase**: Execute using specialized agents (coding agents only)
4. **Validation Phase**: Verify quality gates pass (using go tools and bash commands)
5. **Review Phase**: Final assessment and optimization (review agents only)

## Claude Code's Orchestrator Role

As Claude Code, your ONLY responsibilities are:

1. **Reading** - Read PRP files and gather context (but delegate research to agents)
2. **Orchestrating** - Delegate tasks to appropriate sub-agents with clear instructions
3. **Monitoring** - Run task test, task lint, and other tools to check results
4. **Coordinating** - Manage the workflow between different sub-agents
5. **Reporting** - Provide status updates on progress and results

You MUST NOT:

- Write any Go code or any other programming language code
- Create or modify any .go files
- Implement any functionality directly

## Sub-Agent Orchestration Workflow

**Execute the following workflow (Claude Code orchestrates but NEVER codes):**

### Phase 1: Load PRP and Extract Target Task

**Step 1**: Parse the arguments from `$ARGUMENTS`:

- Extract the PRP name (first token)
- Extract the task identifier (remaining tokens, e.g., "Task 1")
- Construct the PRP path: `docs/PRPs/{prp-name}.md`

**Step 2**: Load and extract the specific task:

- Read the PRP file at the constructed path
- Find the section matching `### {Task Identifier}:` (e.g., `### Task 1:`)
- Extract content from that heading until the next `### Task` or section separator
- This extracted content is the **scope** for this execution

**Step 3**: Use the **deep-research-specialist** sub-agent to:

- Analyze ONLY the extracted task requirements
- Review relevant Go context in `docs/CODING_GUIDELINES.md`
- Research any unfamiliar Go packages or libraries mentioned in the task
- Document task-specific requirements and technical considerations

**Step 4**: Use the **product-manager-orchestrator** sub-agent to:

- Break down the single task into implementation steps
- Categorize requirements by package (cmd/, internal/, pkg/)
- Create an interface-first implementation strategy for this task
- Define success criteria based on the task's acceptance criteria

### Phase 2: Interface Design & Test Creation

Based on the implementation strategy, use the **systems-architect** sub-agent to:

- Design interfaces following Interface Segregation Principle
- Define package boundaries and dependencies
- Plan error types and quanta errors
- Design for dependency injection and testability

Then use the **qa-test-engineer** sub-agent to:

- Create table-driven test plans following Go testing patterns
- Write failing tests FIRST for each interface/function
- Design benchmark tests for performance-critical paths
- Ensure tests follow patterns from `docs/examples/patterns/testing.md`

### Phase 3: Package Structure Design

If the PRP involves new packages or modules:

- Use the **systems-architect** sub-agent to design package structure (internal/, cmd/, pkg/)
- Use the **database-schema-engineer** sub-agent if data persistence is needed
- Use the **api-docs-writer** sub-agent to create godoc documentation templates

### Phase 4: Implementation

**🚨 CRITICAL: ALL implementation MUST be done by sub-agents. Claude Code MUST NOT write any code.**

**Execute implementation based on PRP requirements:**

For CLI features (Cobra commands):

- **MUST** use the **cli-tool-developer** sub-agent to implement ALL Cobra commands and Viper configuration
- The agent will ensure commands follow patterns from `docs/examples/patterns/cli.md`
- The agent will implement with proper flag validation and help text

For core business logic:

- **MUST** use the **backend-systems-engineer** sub-agent to implement ALL internal packages
- The agent will follow interface-first design with small, focused interfaces
- The agent will implement proper error wrapping and context propagation

For concurrent features:

- **MUST** use the **concurrency-specialist** sub-agent to implement ALL goroutines and channels
- The agent will follow patterns from `docs/examples/patterns/concurrency.md`
- The agent will ensure no goroutine leaks and proper cancellation

For API/service features:

- **MUST** use the **api-backend-engineer** sub-agent to implement ALL HTTP handlers and middleware
- The agent will implement with proper context usage and error handling
- The agent will follow RESTful principles or gRPC patterns as specified

For performance-critical features:

- **MUST** use the **performance-optimizer** sub-agent to implement optimized code paths
- The agent will preallocate slices, use sync.Pool, and minimize allocations
- The agent will implement comprehensive benchmarks

**ENFORCEMENT**: If Claude Code attempts to write code, immediately stop and delegate to the appropriate sub-agent listed above.

### Phase 5: Quality Validation

Follow `verification-before-completion` before reporting phase completion. When dispatching independent task agents, follow `dispatching-parallel-agents` patterns.

After implementation, run comprehensive quality checks:

```bash
# Format check
gofmt -l . && test -z "$(gofmt -l .)"

# Imports organization
goimports -l . && test -z "$(goimports -l .)"

# Comprehensive linting
golangci-lint run --timeout=5m

# Run tests with race detection
task test-race

# Run benchmarks
task bench

# Check test coverage
task coverage
# View coverage details
go tool cover -func=coverage/coverage.out

# Build validation
task build

# Module verification
task tidy && git diff --exit-code go.mod go.sum

# Security scan (if gosec is installed)
gosec -quiet ./...

# Check for shadow variables
go vet -shadow ./...
```

**Quality Gate Criteria:**

- ✅ All tests passing with -race flag
- ✅ Test coverage ≥80% for new code
- ✅ Zero formatting issues (gofmt, goimports)
- ✅ Zero linting errors from golangci-lint
- ✅ All exported items have godoc comments
- ✅ No race conditions detected
- ✅ Successful build
- ✅ go.mod is tidy

### Phase 6: Code Quality & Security Review

If all quality gates pass:

- Use the **code-review-specialist** sub-agent to review code quality and Go idioms
- Use the **security-threat-analyst** sub-agent to assess security vulnerabilities
- Use the **performance-optimizer** sub-agent to review benchmarks and profiling

If quality gates fail:

- Use the **code-analyzer-debugger** sub-agent to identify root causes
- Return to Phase 4 with specific fixes needed
- Maximum 3 iterations allowed

### Phase 7: Documentation & Completion

Once all quality gates pass:

- Use the **technical-docs-writer** sub-agent to create/update godoc documentation
- Use the **api-docs-writer** sub-agent to update API documentation if needed
- Verify all PRP requirements have been implemented

## Conditional Agent Selection Logic

The workflow dynamically selects agents based on PRP content:

**CLI-Heavy PRPs** (commands, flags, configuration):

- Primary: cli-tool-developer
- Support: systems-architect for command structure

**Service/API PRPs** (HTTP handlers, gRPC services):

- Primary: api-backend-engineer
- Support: database-schema-engineer for data models

**Concurrent PRPs** (worker pools, pipelines, channels):

- Primary: concurrency-specialist
- Support: performance-optimizer for throughput

**Library PRPs** (reusable packages):

- Primary: backend-systems-engineer
- Support: technical-docs-writer for godoc

**Performance PRPs** (optimization, profiling):

- Primary: performance-optimizer
- Support: concurrency-specialist for parallel processing

## Success Criteria

The **task** is successfully executed when:

1. **Task requirements implemented** - All acceptance criteria for the specific task are met
2. **Interface-first design** - Interfaces defined before implementations
3. **Table-driven tests** - Comprehensive test coverage with subtests for new code
4. **Quality gates passed** - All Go validation checks succeed
5. **Code review approved** - Follows Go idioms and project patterns
6. **Security assessment passed** - No vulnerabilities found
7. **Documentation complete** - All exported items have godoc comments

**Note**: This command completes ONE task. Run again with the next task identifier to continue PRP execution.

## Important Rules

**🚨 ABSOLUTE RULE**: Claude Code (primary assistant) MUST NEVER write implementation code. ALL code must be written by specialized sub-agents.

**CRITICAL**: The **product-manager-orchestrator** ONLY analyzes and coordinates - it NEVER writes code.

**INTERFACE FIRST**: Interfaces MUST be designed before implementations (by systems-architect sub-agent).

**NO GUESSING**: Always gather real context - never guess imports, function names, or package paths.

**IDIOMATIC GO** (enforced by sub-agents):

- No ignored errors (`_ = someFunc()`)
- No `panic` for normal error handling
- Proper context propagation as first parameter
- Follow patterns from `docs/examples/` directory

**ITERATION LIMIT**: Maximum 3 iterations for quality gate passage.

**ENFORCEMENT CHECKLIST**:

- [ ] Did Claude Code write any code? If yes, STOP and delegate
- [ ] Are all coding tasks assigned to specific sub-agents? If no, revise
- [ ] Is each sub-agent given clear, complete instructions? If no, improve

## Workflow Tracking

Throughout execution:

- Log progress after each phase
- Track which agents completed their tasks
- Monitor quality gate results (task test, task lint, etc.)
- Report any blockers or issues
- Update strategy based on findings

The workflow continues until all success criteria are met or maximum iterations are reached.

## How to Properly Delegate to Sub-Agents

When delegating to a sub-agent, provide:

1. **Clear Task Description**: Exactly what needs to be implemented
2. **Context from PRP**: Relevant requirements and specifications
3. **File Paths**: Where code should be created/modified (internal/, cmd/, etc.)
4. **Dependencies**: What packages, imports, or existing code to use
5. **Standards**: Reference specific files from `docs/examples/` directory
6. **Success Criteria**: How to know when the task is complete

Example delegation:

```text
Use the **cli-tool-developer** sub-agent to implement the status command:
- Create a new command at cmd/status.go
- The command should display current session status using Cobra
- Use Viper for configuration management
- Follow patterns from docs/examples/patterns/cli.md
- Include table output using tabwriter from standard library
- Add flags for --format (json|text|table) and --verbose
- The command is complete when it displays status in all three formats
```

## Common Pitfall: Accidental Coding

If you catch yourself starting to write code like:

- "Let me create a function..."
- "I'll implement this interface..."
- "Here's the Go code for..."

**STOP IMMEDIATELY** and instead say:

- "I'll use the [appropriate sub-agent] to implement..."
- "Let me delegate this coding task to..."
- "The [sub-agent] will create..."

## Special Go Sub-Agents

For Go-specific tasks, prefer these specialized agents:

- **cli-tool-developer**: For Cobra commands and CLI features
- **concurrency-specialist**: For goroutines, channels, and concurrent patterns
- **api-backend-engineer**: For HTTP/gRPC services
- **performance-optimizer**: For benchmarking and optimization
- **systems-architect**: For package structure and interfaces

## Final Step

After completing the task:

1. **Atomic commit** - Stage only this task's files and commit using the task's suggested commit message:
   ```bash
   git add <task-specific-files>
   git commit -m "<commit message from PRP task>"
   ```
2. **Mark task completed** - Use `TaskUpdate(taskId, status: "completed")` to mark this task done
3. **Show progress** - Run `TaskList` to display remaining work and newly unblocked tasks
4. **Report task completion** - Summarize what was implemented for this task
5. **Evaluate the task** - Run `/evaluate-prp-task {prp-name} Task N` to assess implementation quality
6. **Fix issues (if any)** - If evaluation score < 8, run `/fix-prp-task {prp-name} Task N`
7. **Suggest next task** - Recommend running `/execute-prp-task {prp-name} Task N+1` for the next unblocked task

**Workflow**: Execute → Commit → TaskUpdate → Evaluate → Fix (if needed) → Re-evaluate → Next Task
