# Example: Codebase-Wide Error Handling Analysis

This example demonstrates recursive decomposition for analyzing error handling patterns across a large codebase.

## Task
"Analyze all error handling patterns in this codebase and provide a comprehensive report on consistency, gaps, and recommendations."

## Decomposition Strategy

### Phase 1: Filter and Identify (Constant complexity)

```
Step 1: Identify relevant file types
- Glob("**/*.go") → 450 files
- Total: 450 files

Step 2: Filter for error-related code
- Grep("if err != nil|errors\\.New|fmt\\.Errorf", type="go") → 89 files
- Grep("errors\\.Is|errors\\.As|errors\\.Join", type="go") → 67 files
- Union: 102 unique files with error handling
```

### Phase 2: Partition for Parallel Processing

```
Partition by module:
- cmd/* → 23 files (Batch A)
- internal/agent/* → 31 files (Batch B)
- internal/exploit/* → 28 files (Batch C)
- internal/utils/* → 12 files (Batch D)
- Other → 8 files (Batch E)
```

### Phase 3: Launch Parallel Sub-Agents

```
Task(subagent_type="Explore", prompt="""
Analyze error handling in src/api/*.
For each file with error handling:
1. Identify error handling patterns used
2. Note any error types defined or caught
3. Check for consistent error propagation
4. Flag any unhandled promise rejections
Return structured findings.
""")

# Launch 5 agents in parallel for batches A-E
```

### Phase 4: Aggregate Results

```
Collect findings from all sub-agents:
- Batch A: CLI error handling, cobra command errors
- Batch B: Agent-level error wrapping, logging patterns
- Batch C: Exploit execution errors, forge errors
- Batch D: Utility error wrappers, validation errors
- Batch E: Mixed patterns, some inconsistencies
```

### Phase 5: Synthesize Report

```
Categories identified:
1. CLI Layer: Cobra command errors, user-facing messages
2. Agent Layer: Wrapped errors with context, zap logging
3. Exploit Layer: Forge execution errors, revert handling
4. Utilities: Generic error wrappers, sentinel errors

Patterns:
- Consistent: Errors wrapped with fmt.Errorf context
- Gap: Some errors not using errors.Is/As for matching
- Recommendation: Standardize sentinel errors across packages
```

### Phase 6: Verify with Spot Checks

```
Verification queries:
1. "Confirm ApiError is used consistently in src/api/"
2. "Check if DatabaseError preserves stack traces"
3. "Verify error boundaries cover all route components"
```

## Expected Output Structure

```markdown
# Error Handling Analysis Report

## Executive Summary
- 102 files contain error handling logic
- 4 main error categories identified
- 3 consistency issues found
- 5 recommendations provided

## Error Type Taxonomy
### CLI Errors (cmd/)
- Cobra command errors with user-facing messages
- Configuration validation failures
- Flag parsing errors

### Agent Errors (internal/agent/)
...

## Pattern Analysis
### Consistent Patterns
1. All public functions return errors with wrapped context
2. Errors include structured fields via zap logging
...

### Inconsistencies Found
1. Some internal functions swallow errors without logging
2. Forge errors lose original revert reason
...

## Recommendations
1. Implement sentinel error constants per package
2. Use errors.Is/As consistently for error matching
...
```

## Metrics

- **Files analyzed:** 102
- **Sub-agents used:** 5
- **Total tokens processed:** ~150k (across all agents)
- **Equivalent direct context:** Would require 150k token window
- **Quality:** High (no context rot)
