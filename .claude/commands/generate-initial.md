---
description: Generate INITIAL.md from sprint documentation file with Go-specific patterns
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Task
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "[sprint-file-path]"
model: opus
---

# Generate INITIAL.md from Sprint File

## Required Skills

This command uses the following skills (auto-loaded based on context):
- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `testing-patterns` - For table-driven tests and mocking
- `interface-design` - For interface-first design

## Sprint file: $ARGUMENTS

Generate an `INITIAL.md` file from a sprint documentation file, following the quanta Go project structure and development guidelines. Read the sprint file to extract feature details, technical implementation, and considerations.

## Process

1. **Read Sprint File**

   - Read the sprint file from `$ARGUMENTS`
   - Extract feature name, overview, and technical details
   - Identify Go code examples and implementation patterns
   - Gather testing, benchmarking, and performance requirements
   - Note concurrency patterns if applicable

2. **Transform to `INITIAL.md` Format**

   - Create structured sections: FEATURE, EXAMPLES, DOCUMENTATION, OTHER CONSIDERATIONS
   - Format Go code examples with proper syntax highlighting
   - Include relevant Go-specific patterns and idioms
   - Reference project guidelines from `docs/CODING_GUIDELINES.md`
   - Include interface definitions and package structure

3. **Generate Output**
   - Write the transformed content to `INITIAL.md`
   - Follow the exact format structure shown in the template
   - Include all relevant technical details from the sprint

## Structure Template

```markdown
## FEATURE:

[Extract from sprint title and overview section]

## EXAMPLES:

[Extract and format Go code examples from Technical Implementation section]

### Package Structure
```

quanta/
├── cmd/
│ └── [command files]
├── internal/
│ ├── [feature packages]
│ └── [implementation]
├── pkg/
│ └── [public APIs if any]
└── testdata/
└── [test fixtures]

````

### Interface Definitions
```go
// Define interfaces first
type FeatureInterface interface {
    Method(ctx context.Context) error
}
````

### Implementation Examples

```go
// Show concrete implementations
type featureImpl struct {
    // fields
}

func (f *featureImpl) Method(ctx context.Context) error {
    // implementation
}
```

### Table-Driven Tests

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Benchmark Examples

```go
func BenchmarkFeature(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // benchmark code
    }
}
```

## DOCUMENTATION:

[List relevant documentation sources]

### Go Documentation

- Go language specification (go.dev/ref/spec)
- Effective Go (go.dev/doc/effective_go)
- Go Code Review Comments (go.dev/wiki/CodeReviewComments)
- Standard library docs (pkg.go.dev/std)

### Project Documentation

- `docs/CODING_GUIDELINES.md` - Main coding standards
- `docs/examples/patterns/` - Implementation patterns
  - `concurrency.md` - Goroutines and channels
  - `cli.md` - Cobra/Viper CLI patterns
  - `testing.md` - Table-driven tests
  - `mocking.md` - Interface mocking
- `docs/examples/standards/` - Coding standards
  - `documentation.md` - Godoc standards
  - `interfaces.md` - Interface design
  - `go-specific.md` - Go idioms

### External Libraries

- Cobra CLI framework (github.com/spf13/cobra)
- Viper configuration (github.com/spf13/viper)
- Testify testing (github.com/stretchr/testify)
- Zap logging (go.uber.org/zap)
- [Other relevant libraries from go.mod]

## OTHER CONSIDERATIONS:

[Extract from sprint file]

### Testing Requirements

- Table-driven tests for all functions
- Minimum 80% code coverage
- Race condition testing with -race flag
- Benchmark tests for performance-critical code
- Mock interfaces for external dependencies

### Performance Requirements

- Profile with pprof for CPU and memory
- Benchmark critical paths
- Preallocate slices when size is known
- Use sync.Pool for frequently allocated objects
- Avoid unnecessary allocations

### Concurrency Considerations

- Use context for cancellation
- Prevent goroutine leaks
- Proper channel closing (sender side only)
- Worker pool patterns for parallel processing
- Rate limiting where appropriate

### Security Considerations

- Input validation on all user inputs
- No SQL injection vulnerabilities
- Sanitize file paths
- Use context values with typed keys
- Never log sensitive information

### Development Process

- Interface-first design
- Dependency injection for testability
- Use internal/ for private packages
- Follow error wrapping patterns
- All exported items need godoc comments

### Implementation Checklist

- [ ] Define interfaces before implementation
- [ ] Create quanta errors for known conditions
- [ ] Write table-driven tests first
- [ ] Add benchmark tests for critical paths
- [ ] Document all exported items
- [ ] Run golangci-lint for code quality
- [ ] Test with -race flag
- [ ] Profile if performance-critical
- [ ] Update go.mod dependencies
- [ ] Add integration tests if applicable

```

### CRITICAL: Research and Planning Phase

**After completing codebase research and exploration, BEFORE writing the INITIAL.md:**

- Thoroughly plan your approach
- Consider Go idioms and patterns
- Map out package structure and interfaces
- Plan error handling strategy
- Design for testability with dependency injection
- Consider concurrency patterns if applicable

## Execution Steps

1. Validate that the sprint file exists at the provided path
2. Read and parse the sprint file content
3. Extract sections:
   - Feature title and overview → FEATURE section
   - Technical Implementation → EXAMPLES section (with Go code)
   - References to Go docs/libraries → DOCUMENTATION section
   - Testing/Performance/Security/Checklist → OTHER CONSIDERATIONS
4. Format the content according to Go patterns:
   - Interface definitions first
   - Table-driven test examples
   - Benchmark examples if performance-critical
   - Proper error handling patterns
5. Write the formatted content to INITIAL.md
6. Provide confirmation of successful generation

## Output

Save as: `INITIAL.md`

## Quality Checklist

- [ ] Feature description is clear and follows Go conventions
- [ ] All Go code examples use proper idioms
- [ ] Interface definitions precede implementations
- [ ] Table-driven test examples included
- [ ] Documentation sources include Go-specific references
- [ ] Error handling patterns are demonstrated
- [ ] Concurrency patterns included if applicable
- [ ] Package structure follows internal/ convention
- [ ] Benchmark examples for performance-critical code
- [ ] All exported items have godoc comments in examples
- [ ] Security and performance requirements are Go-specific
- [ ] Follows project structure (cmd/, internal/, pkg/)
- [ ] References correct pattern files from docs/examples/
```
