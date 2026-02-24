---
description: Fix issues from a specific task's evaluation
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Edit
  - Task
  - Bash
  - TaskCreate
  - TaskUpdate
  - TaskList
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "<prp-name> <task-identifier>"
model: opus
---

# Fix PRP Task Implementation Issues

## Arguments: $ARGUMENTS

Parse the input arguments as follows:

| Component | Description | Example |
|-----------|-------------|---------|
| **PRP Name** | First token - the PRP filename without extension | `feedback-driven-exploitation` |
| **Task Identifier** | Remaining tokens - the task to fix | `Task 1`, `Task 4` |

**PRP File Path**: `docs/PRPs/{prp-name}.md`

### Parsing Examples

| Input | PRP Path | Target Task |
|-------|----------|-------------|
| `feedback-driven-exploitation Task 1` | `docs/PRPs/feedback-driven-exploitation.md` | Task 1 |
| `feedback-driven-exploitation Task 4` | `docs/PRPs/feedback-driven-exploitation.md` | Task 4 |
| `multi-model-ensemble Task 2` | `docs/PRPs/multi-model-ensemble.md` | Task 2 |

### Task-Scoped Fixes

When fixing, focus ONLY on:
1. Issues identified in the task-specific evaluation report
2. Code changes made for the specified task
3. Tests written for the task's functionality

## Required Skills

This command uses the following skills (auto-loaded based on context):
- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `testing-patterns` - For table-driven tests and mocking
- `code-review` - For reviewing code quality

Systematically fix critical and major issues identified in a PRP evaluation report, following strict TDD practices and Quanta's Go coding standards.

## Input Files Required

- Original PRP: `docs/PRPs/{prp-name}.md`
- Evaluation Report: `docs/PRPs/{prp-name}-task-{N}-evaluation.md`
- Fixes Document: `docs/PRPs/{prp-name}-task-{N}-fixes.md`

Example: If fixing `feedback-driven-exploitation Task 1`:
- PRP: `docs/PRPs/feedback-driven-exploitation.md`
- Evaluation: `docs/PRPs/feedback-driven-exploitation-task-1-evaluation.md`
- Fixes: `docs/PRPs/feedback-driven-exploitation-task-1-fixes.md`

## Agent Orchestration Strategy

Use the **product-manager-orchestrator** to coordinate specialized agents based on the types of fixes needed:

1. **backend-systems-engineer** - Implement missing Go features and fix functionality
2. **code-review-specialist** - Guide Go code quality improvements
3. **senior-software-engineer** - Address architectural and design issues
4. **security-threat-analyzer** - Address security vulnerabilities
5. **qa-test-engineer** - Add missing tests and improve coverage
6. **performance-optimizer** - Optimize Go performance issues
7. **technical-docs-writer** - Fix documentation gaps

## Execution Process

1. **Load Context & Analyze Issues**

   - Read the original PRP file to understand requirements
   - Read the evaluation report to identify all violations
   - Read the fixes document for specific remediation tasks
   - Review discipline files in `docs/examples/` directory:
     - `philosophy/tdd-principles.md` - TDD is NON-NEGOTIABLE
     - `philosophy/tdd-workflow.md` - Red-Green-Refactor cycle
     - `philosophy/lever-framework.md` - LEVER principles (see also `coding-philosophy` skill)
     - `standards/go-specific.md` - Go idioms and patterns
     - `standards/interfaces.md` - Interface design patterns
     - `patterns/testing.md` - Go testing patterns
     - `patterns/concurrency.md` - Goroutine patterns
     - `patterns/cli.md` - Cobra CLI patterns
   - Review `docs/CODING_GUIDELINES.md` for Go standards
   - Check `CLAUDE.md` for project-specific requirements
   - Identify interconnected issues that must be fixed together
   - Determine which agent to spawn to resolve each task

2. **Deep Analysis of Fix Strategy**

   - **CRITICAL**: Categorize issues by severity (Critical/Major/Minor)
   - Map each issue to specific discipline violations and specialized agent
   - Plan fix order considering dependencies:
     - Interface definitions before implementations
     - Test files must exist alongside source files
     - Core packages before dependent packages
   - Use TaskCreate/TaskUpdate/TaskList to track TDD-focused approach:
     - Write failing test for expected behavior (RED)
     - Fix implementation with minimum code (GREEN)
     - Verify all tests pass
     - Refactor if needed while maintaining green tests
   - Plan validation strategy for each fix type
   - Consider ripple effects of fixes on other Go packages

   **IMPORTANT** The **product-manager-orchestrator** DOES NOT MAKE ANY CODE CHANGES, she only coordinates with the specialized agents to make the changes.

3. **Execute Fixes with TDD Discipline**

   - **For Missing Files (Critical)**:

     ```go
     // Step 1: Create test file first (e.g., service_test.go)
     // Step 2: Write failing test for expected behavior
     func TestService_Method(t *testing.T) {
         // Arrange-Act-Assert pattern
     }
     // Step 3: Implement minimum code to pass
     // Step 4: Verify all related tests pass
     ```

   - **For Interface Violations**:

     ```go
     // Step 1: Define interface FIRST
     type Service interface {
         Process(ctx context.Context, data []byte) error
     }

     // Step 2: Implement concrete type
     type serviceImpl struct {
         // fields
     }

     // Step 3: Ensure implementation satisfies interface
     var _ Service = (*serviceImpl)(nil) // Compile-time check
     ```

   - **For Missing Tests**:

     ```go
     // Write table-driven tests for uncovered code
     func TestFeature(t *testing.T) {
         tests := []struct {
             name    string
             input   string
             want    string
             wantErr bool
         }{
             {"happy path", "valid", "expected", false},
             {"error case", "invalid", "", true},
             {"edge case", "", "default", false},
         }

         for _, tt := range tests {
             t.Run(tt.name, func(t *testing.T) {
                 // Test implementation
             })
         }
     }
     ```

   - **For Code Quality Issues**:

     - Remove all fmt.Println statements (use log package)
     - Replace `interface{}` with specific types where possible
     - Add missing godoc comments with examples
     - Fix golangci-lint violations
     - Ensure proper error wrapping with context
     - Follow import organization (stdlib → third-party → internal)
     - Keep functions under 50 lines
     - Use early returns to reduce nesting

   - **For Missing Documentation**:

     ```go
     // ProcessData handles data processing with validation and transformation.
     // It validates the input data, applies configured transformations,
     // and returns the processed result.
     //
     // The data parameter must be non-nil and properly formatted.
     // Returns an error if validation fails or processing encounters issues.
     //
     // Example:
     //
     //  data := []byte(`{"key": "value"}`)
     //  result, err := ProcessData(data)
     //  if err != nil {
     //      log.Fatal(err)
     //  }
     func ProcessData(data []byte) ([]byte, error) {
         // Implementation
     }
     ```

4. **Validate Each Fix Category**

   - **After Critical Fixes**:

     ```bash
     # Ensure code compiles
     go build ./...

     # Run new unit tests
     go test -v ./path/to/package

     # Run with race detector
     go test -race ./...

     # Verify interface implementations
     go vet ./...
     ```

   - **After Major Fixes**:

     ```bash
     # Full linting check
     task lint
     # or
     golangci-lint run

     # Verify no fmt.Println remains
     ! rg "fmt\.Println" internal/ cmd/ --type go

     # Verify no empty interfaces
     ! rg "interface\{\}" internal/ cmd/ --type go | grep -v "// OK:"

     # Run all unit tests
     task test
     # or
     go test ./...

     # Run tests with coverage
     task coverage
     ```

   - **After Minor Fixes**:

     ```bash
     # Quick validation
     task fmt      # Format code
     task lint     # Run golangci-lint (includes vet)
     task test     # Quick test

     # Or run specific tests
     go test -v -run TestSpecificFunction ./internal/package
     ```

5. **Comprehensive Final Validation**

   ```bash
   # Level 1: Syntax & Style
   task fmt           # Format with gofmt and goimports
   task lint         # Linting (includes vet)
   go build ./...    # Ensure compilation

   # Level 2: Pattern Compliance
   task lint         # Run golangci-lint
   ! rg "fmt\.Println" internal/ cmd/ --type go
   ! rg "interface\{\}" internal/ cmd/ --type go | grep -v "// OK:"
   rg "fmt\.Errorf.*%w" internal/ --type go || echo "Check error wrapping"

   # Level 3: Test Coverage
   task coverage     # Generate coverage report
   # Verify coverage meets requirements (≥80% unit tests)

   # Run integration tests if they exist
   go test -tags=integration ./...

   # Level 4: Build Validation
   task build        # Build binary with version info

   # Level 5: Benchmarks (if applicable)
   task bench        # Run benchmarks

   # Level 6: Race Detection
   task test-race    # Test with race detector

   # Level 7: Pre-commit Check
   task check        # Run all pre-commit checks
   ```

6. **Fix Verification & Reporting**

- Re-read the evaluation report
- Verify each identified issue has been addressed:
  - ✅ Critical issues fixed and tested
  - ✅ Major issues resolved
  - ✅ Minor issues addressed where possible
- Run the complete validation suite
- Calculate improvement score (should be ≥8/10)
- Document any issues that couldn't be fixed with justification
- Create a summary of changes made

## Fix Prioritization

### Critical (Must Fix - Blocks Deployment)

- Missing schema-first implementations (Zod schemas)
- Missing test setup files
- Type safety violations (any types)
- Console.log in production code
- Missing core functionality
- Security vulnerabilities

### Major (Should Fix - Quality Gates)

- Missing JSDoc on exported functions/components
- Missing tests for new code
- Accessibility violations
- React hook dependency issues
- Improper error handling
- Database schema mismatches

### Minor (Consider Fixing - Technical Debt)

- Performance optimizations (React.memo)
- Enhanced error boundaries
- Additional test coverage
- Code organization improvements
- Enhanced loading states

## Common Fix Patterns

### Converting to Interface-First

```go
// ❌ Before: Concrete type without interface
type UserService struct {
    db *sql.DB
}

func (s *UserService) GetUser(id string) (*User, error) {
    // Implementation
}

// ✅ After: Interface-first design
type UserRepository interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
}

type userService struct {
    repo UserRepository  // Depend on interface
}

// Compile-time interface check
var _ UserRepository = (*userService)(nil)

func NewUserService(repo UserRepository) *userService {
    return &userService{repo: repo}
}
```

### Error Handling Pattern

```go
// ❌ Before: Poor error handling
func GetUser(id string) *User {
    user, _ := db.Query("SELECT * FROM users WHERE id = ?", id)
    return user
}

// ✅ After: Proper error handling with context
func GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, errors.New("user ID cannot be empty")
    }

    user, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("user %s not found", id)
        }
        return nil, fmt.Errorf("failed to get user %s: %w", id, err)
    }

    return user, nil
}

// Sentinel errors for comparison
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
)
```

### Fixing Concurrency Issues

```go
// ❌ Before: Race condition
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++ // Race condition!
}

// ✅ After: Thread-safe with mutex
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// Or use atomic operations
type AtomicCounter struct {
    value int64
}

func (c *AtomicCounter) Increment() {
    atomic.AddInt64(&c.value, 1)
}
```

### Adding Comprehensive Godoc

```go
// RegisterUser processes user registration with validation and optional notifications.
// It validates the registration data, creates the user account, and optionally
// sends a welcome email based on the provided options.
//
// The registration data must include a valid email address and a password
// of at least 8 characters. The email address must be unique in the system.
//
// Options can be used to control post-registration behavior such as
// sending welcome emails or requiring email verification.
//
// Returns the created user on success, or an error if registration fails.
// Common errors include duplicate email addresses, validation failures,
// or database connection issues.
//
// Example:
//
//  data := &RegistrationData{
//      Email:    "user@example.com",
//      Password: "securePassword123",
//      Name:     "John Doe",
//  }
//
//  opts := &RegistrationOptions{
//      SendWelcomeEmail:         true,
//      RequireEmailVerification: true,
//  }
//
//  user, err := RegisterUser(ctx, data, opts)
//  if err != nil {
//      log.Printf("Registration failed: %v", err)
//      return err
//  }
//  log.Printf("User created: %s", user.ID)
func RegisterUser(ctx context.Context, data *RegistrationData, opts *RegistrationOptions) (*User, error) {
    // Implementation
}
```

## Critical Reminders

1. **Never Skip Tests**: Even when fixing issues, write tests first
2. **Preserve Functionality**: Ensure existing tests continue to pass
3. **Fix Root Causes**: Don't just suppress symptoms
4. **Document Changes**: Add comments explaining why fixes were needed
5. **Validate Continuously**: Run validation after each category of fixes

## Expected Outcomes

After successful execution:

- All critical issues resolved
- All major issues fixed
- Score improved to ≥8/10
- All tests passing with ≥80% coverage
- No linting errors
- No TypeScript errors
- Build successful
- Documentation complete

## Usage Example

```bash
# After evaluation identifies issues
/evaluate-prp-task feedback-driven-exploitation Task 1
# Creates: docs/PRPs/feedback-driven-exploitation-task-1-evaluation.md (score: 7/10)
# Creates: docs/PRPs/feedback-driven-exploitation-task-1-fixes.md

# Fix the issues for this specific task
/fix-prp-task feedback-driven-exploitation Task 1
# Reads task-specific files and systematically fixes issues

# Re-evaluate to confirm fixes
/evaluate-prp-task feedback-driven-exploitation Task 1
# Should now score ≥8/10

# When task passes, proceed to next task
/execute-prp-task feedback-driven-exploitation Task 2
```

**IMPORTANT Final Step**: Run comprehensive validation

```bash
# Final validation
task check         # Run all pre-commit checks
```

Remember: The goal is not just to suppress errors, but to genuinely improve Go code quality while maintaining all functionality. Every fix should improve the codebase, not just change it. Follow TDD principles - write the test first, then fix the code.

## Post-Fix Re-Evaluation

**CRITICAL**: After all fixes are applied, automatically execute the evaluate-prp-task command to confirm the score has improved. Follow `verification-before-completion` before claiming fixes are complete.

1. Run `task pre-commit` - Ensure all quality checks pass
2. Execute `/evaluate-prp-task {prp-name} Task N` - Re-evaluate to confirm score ≥ 8

If the re-evaluation score is still below 8, continue fixing until quality gates are met.

## Follow-up

After fixing:
1. Re-run `/evaluate-prp-task {prp-name} Task N` to verify fixes
2. When task passes (score ≥ 8), proceed to next task with `/execute-prp-task {prp-name} Task N+1`

**Workflow**: Execute → Evaluate → Fix (if needed) → Re-evaluate → Next Task
