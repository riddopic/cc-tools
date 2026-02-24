---
name: reference-docs
description: Reference documentation patterns for API and symbol documentation. Use when writing reference docs, API docs, parameter tables, or technical specifications. Triggers on reference docs, API reference, function reference, parameters table, symbol documentation.
---

# Reference Documentation Patterns

Reference documentation is information-oriented -- helping experienced users find precise technical details quickly. Brevity, consistency, and completeness are the expectations.

## Document Structure Template

```markdown
---
title: "[Symbol/API Name]"
description: "One-line description of what it does"
---

# [Name]

Brief description (1-2 sentences). What it is and its primary purpose.

## Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `param1` | `string` | Yes | - | What this parameter controls |
| `param2` | `number` | No | `10` | Optional behavior modification |

## Returns

| Type | Description |
|------|-------------|
| `ReturnType` | What the function returns and when |

## Errors

| Error | Condition |
|-------|-----------|
| `NotFoundError` | Resource does not exist |
| `UnauthorizedError` | Invalid or expired credentials |

## Example

Complete, runnable example with realistic values and expected output.

## Related

- [RelatedSymbol](/reference/related) - Brief description
```

## Writing Principles

1. **Brevity over explanation.** State facts, not rationale. Avoid "why" -- save that for Explanation docs.
2. **Scannable tables, not prose.** Use tables for parameters, returns, errors. Never describe parameter lists in paragraph form.
3. **Consistent format across entries.** All reference pages for similar items must follow identical structure: same heading order, table columns, example format.
4. **Every example must be runnable.** Include all imports, use realistic values (not "foo"/"bar"), show expected output.

## Parameter Documentation

### Required vs Optional
Always indicate required/optional and provide defaults for optional parameters.

### Complex Types
Document object parameters with a sub-table:

```markdown
### UserOptions

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `includeDeleted` | `boolean` | No | Include soft-deleted users |
| `fields` | `string[]` | No | Fields to return |
```

### Enum Values
Document allowed values inline: `status` | `string` | `active`, `pending`, `suspended` | Account status

## Godoc Conventions

For Go packages, follow godoc patterns:
- Package comments describe purpose and functionality
- Function docs start with the function name
- Use complete sentences with proper grammar
- Document parameters, return values, and errors

```go
// Package analyzer provides vulnerability detection for smart contracts.
package analyzer

// Analyze examines the contract and returns detected vulnerabilities.
// It returns an error if the bytecode is invalid or analysis fails.
func (a *Analyzer) Analyze(ctx context.Context, bytecode []byte) ([]Vulnerability, error)
```

## Checklist

- [ ] Title matches the symbol/API name exactly
- [ ] Description is one clear sentence
- [ ] All parameters documented with types
- [ ] Required vs optional clearly marked
- [ ] Default values specified for optional parameters
- [ ] Return type and structure documented
- [ ] At least one complete, runnable example
- [ ] Example uses realistic values
- [ ] Related pages linked
- [ ] Format matches other reference pages in the docs

## Related Skills

- **docs-style**: Core writing conventions and components
- **howto-docs**: How-To guide patterns for task-oriented content
- **tutorial-docs**: Tutorial patterns for learning-oriented content
- **explanation-docs**: Conceptual documentation patterns
