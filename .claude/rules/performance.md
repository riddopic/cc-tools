# Performance Guidelines

## Quick Commands

```bash
task bench              # Run benchmarks with memory stats
task test-race          # Check for race conditions
```

## Model Selection for Agents

| Model | Use Case | Cost/Performance |
| ------- | ---------- | ------------------ |
| **Haiku 4.5** | Lightweight agents, frequent invocation, pair programming | 3x cost savings vs Sonnet |
| **Sonnet 4.5** | Main development, orchestration, complex coding | Best coding model |
| **Opus 4.5** | Complex architecture, deep reasoning, research | Maximum reasoning |

## Context Window Management

Avoid last 20% of context window for:

- Large-scale refactoring
- Multi-file feature implementation
- Complex debugging sessions

## Performance Checklist

Before optimizing:

- [ ] Profiled to identify bottleneck
- [ ] Benchmark exists for the hot path
- [ ] Optimization addresses measured problem

After optimizing:

- [ ] Benchmark shows improvement
- [ ] No regressions in other benchmarks
- [ ] Code is still readable and maintainable
- [ ] Tests still pass
