# Comment Guidelines

## No Project Management References in Code

Never include sprint numbers, ticket IDs, issue numbers, or other project
management references in code comments, function names, or variable names.

- Git history tracks *when* changes were made
- Comments should explain *why*, not *when*

Bad:

- `// Sprint 26: Honor rag.store_path configuration`
- `// JIRA-1234: Fix race condition`
- `func buildSprint25Options()`

Good:

- `// Honor rag.store_path configuration`
- `// Fix race condition in concurrent map access`
- `func buildFeatureOptions()`
