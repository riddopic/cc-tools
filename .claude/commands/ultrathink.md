---
description: Deep thinking mode for complex task exploration, planning, and implementation
allowed-tools:
  - Read
  - Grep
  - Glob
  - Task
  - Bash
  - Write
  - Edit
  - TodoWrite
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "<task>"
model: opus
---

# Deep thinking on a task

## Required Skills

This command uses the following skills (auto-loaded based on context):
- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `verification-before-completion` - For verifying work completeness

Think carefully and thoroughly through this. ALWAYS FOLLOW OUR PROJECT GUIDELINES AND BEST PRACTICES.
Start by exploring the codebase and gathering the necessary context.
You should use sub-agents where it makes sense.

I want you to explore, plan and implement the following:

$ARGUMENTS
