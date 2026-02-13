# Auth Failures Masquerade as Compilation Failures

**Extracted:** 2026-02-09
**Context:** Long-running regression test suites with OAuth-based LLM providers

## Problem

During extended regression runs (28 parallel tests, 60-minute timeout each), OAuth tokens can expire mid-run. When the LLM provider returns an auth error, the system falls back to generating stub/placeholder code that fails to compile. This inflates the "compilation failure" count, misleading root cause analysis into thinking there's a prompt problem when it's actually an infrastructure issue.

## Solution

Add a dedicated `auth_failure` classification category that fires BEFORE the generic compilation catch-all in the failure classification chain. Pattern match on auth-specific strings:

```go
case containsAny(errLower, "oauth token has expired", "authentication_error",
    "failed to authenticate"):
    return FailureCategoryAuthFailure
```

Key: auth failure detection must be ordered BEFORE compilation detection, because auth failures surface as compilation errors (the stub code doesn't compile).

## Example

R7 reported 10 compilation failures. R8 triage revealed 3 were actually auth failures (pdz, sizecredit, d3xai) — reducing true compilation to 7.

## When to Use

- When classifying test failures in any multi-category taxonomy
- When long-running parallel jobs use time-limited auth tokens
- When a "compilation" or "syntax" failure seems inconsistent with the test's history
- General principle: always check if upstream failures masquerade as downstream symptoms
