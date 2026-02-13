# Examples Directory Structure

This directory contains detailed code examples and standards that complement the main [docs/CODING_GUIDELINES.md](../docs/CODING_GUIDELINES.md).

## Directory Organization

### 📁 philosophy/

Fundamental development principles and practices:

- **[README.md](philosophy/README.md)** - Overview of development philosophy
- **[tdd-principles.md](philosophy/tdd-principles.md)** - Core TDD principles for Go

  - The sacred Red-Green-Refactor cycle
  - Why TDD is about design, not testing
  - Common TDD violations and Go-specific examples
  - StatusLine development with TDD mindset

- **[lever-framework.md](philosophy/lever-framework.md)** - LEVER framework for Go development

  - Leverage existing patterns and libraries
  - Extend before creating from scratch
  - Verify through reactive patterns and channels
  - Eliminate knowledge duplication
  - Reduce complexity at every opportunity

- **[tdd-workflow.md](philosophy/tdd-workflow.md)** - Concrete TDD workflow examples
  - Step-by-step TDD development process
  - StatusLine renderer implementation
  - Configuration loading with error handling
  - Async service patterns with context
  - Anti-patterns to avoid

### 📁 patterns/

Practical code patterns and implementation examples:

- **[concurrency.md](patterns/concurrency.md)** - Goroutines, channels, and concurrency patterns

  - Worker pools for metric collection
  - Fan-out/fan-in patterns
  - Pipeline patterns for data transformation
  - Rate limiting and graceful shutdown
  - Broadcast patterns for event distribution

- **[cli.md](patterns/cli.md)** - CLI development with Cobra and Viper

  - Command structure and subcommands
  - Configuration management
  - Output formatting and colors
  - Interactive elements (prompts, progress bars)
  - Shell completion scripts

- **[testing.md](patterns/testing.md)** - Testing patterns and practices

  - Table-driven test structure
  - Test organization and helpers
  - Golden file testing
  - Integration testing patterns

- **[mocking.md](patterns/mocking.md)** - Comprehensive Mockery v3.5 testing guide
  - Core principles and directory structure
  - Mock generation with `task mocks`
  - Complete TDD workflow with step-by-step examples
  - Basic usage with constructors and EXPECT()
  - Table-driven tests with mocks
  - Advanced patterns (RunAndReturn, stateful mocks, MatchedBy)
  - Concurrent testing and race condition detection
  - Error simulation and timeout handling
  - Integration testing with multiple components
  - Mock factory pattern for reusable setups
  - Performance considerations and debugging
  - Best practices and troubleshooting

### 📁 standards/

Coding standards and language-specific rules:

- **[documentation.md](standards/documentation.md)** - Documentation and comment standards

  - Godoc comment conventions
  - Package, function, and type documentation
  - Error documentation patterns
  - Example documentation

- **[interfaces.md](standards/interfaces.md)** - Interface design and best practices

  - Interface segregation principle
  - Naming conventions
  - Standard interfaces (error, Stringer)
  - Interface composition patterns

- **[go-specific.md](standards/go-specific.md)** - Go language-specific guidelines
  - Memory management and performance
  - Error handling patterns
  - Struct patterns (functional options, builder)
  - Context usage and propagation
  - Package design principles

## How to Use These Examples

1. **Start with the main guidelines**: Read [docs/CODING_GUIDELINES.md](../docs/CODING_GUIDELINES.md) first for the overall structure and philosophy.

2. **Reference patterns during implementation**: When implementing a specific feature, consult the relevant pattern file for concrete examples.

3. **Follow standards consistently**: Use the standards documents to ensure consistent code style across the project.

4. **Copy and adapt**: The code examples are designed to be copied and adapted for your specific use cases.

## Quick Reference

### Need to understand...?

- **Development philosophy** → [philosophy/README.md](philosophy/README.md)
- **TDD principles and practices** → [philosophy/tdd-principles.md](philosophy/tdd-principles.md)
- **LEVER framework** → [philosophy/lever-framework.md](philosophy/lever-framework.md)
- **TDD workflow examples** → [philosophy/tdd-workflow.md](philosophy/tdd-workflow.md)

### Need to implement...?

- **Concurrent metric collection** → [patterns/concurrency.md](patterns/concurrency.md)
- **CLI commands** → [patterns/cli.md](patterns/cli.md)
- **Unit tests** → [patterns/testing.md](patterns/testing.md)
- **Mock objects** → [patterns/mocking.md](patterns/mocking.md)
- **API documentation** → [standards/documentation.md](standards/documentation.md)
- **Clean interfaces** → [standards/interfaces.md](standards/interfaces.md)
- **Idiomatic Go code** → [standards/go-specific.md](standards/go-specific.md)

## Contributing

When adding new examples:

1. Place fundamental principles in `philosophy/`
2. Place pattern examples in `patterns/`
3. Place standards and rules in `standards/`
4. Update this README with the new file
5. Ensure examples compile and run
6. Include both correct and incorrect examples where helpful
7. Reference docs/CODING_GUIDELINES.md appropriately
8. Update philosophy documents when adding new principles

## Related Documentation

- [docs/CODING_GUIDELINES.md](../docs/CODING_GUIDELINES.md) - Main coding guidelines
- [docs/research/](../docs/research/) - In-depth research documents
- [Project README](../README.md) - Project overview
