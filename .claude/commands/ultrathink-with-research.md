---
description: Deep thinking mode with extensive web and documentation research
allowed-tools:
  - Read
  - Grep
  - Glob
  - Task
  - Bash
  - Write
  - Edit
  - TodoWrite
  - WebFetch
  - WebSearch
  - mcp__sequential-thinking__sequentialthinking
  - mcp__context7__resolve-library-id
  - mcp__context7__query-docs
argument-hint: "<task>"
model: opus
---

# Deep thinking on a task with in-depth research

## Required Skills

This command uses the following skills (auto-loaded based on context):
- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `verification-before-completion` - For verifying work completeness
- `discovery-oriented-prompts` - For research-driven discovery

Think carefully and thoroughly through this. ALWAYS FOLLOW OUR PROJECT GUIDELINES AND BEST PRACTICES.
Start by exploring the codebase and gathering the necessary context.
You should use sub-agents where it makes sense.

Do in-depth research on this.

## YOU MUST DO IN-DEPTH RESEARCH, FOLLOW THE <RESEARCH PROCESS>

<RESEARCH PROCESS>

- Don't only research one page, and don't only use your own webs craping tool - instead scrape many relevant pages from all documentation links mentioned, use Ref, Context 7 and Deep Wiki MCP servers to get additional documentation.
- Take my tech as sacred truth, for example if I say a model name then research that model name for LLM usage - don't assume from your own knowledge at any point
- When I say don't just research one page, I mean do incredibly in-depth research, like to the point where it's just absolutely ridiculous how much research you've actually done, then when you create the plan, you need to put absolutely everything into that including references to the .md files you put inside the `docs/research/` directory so any AI and sub-agents can pick up the work and generate WORKING and COMPLETE production ready code.

</RESEARCH PROCESS>

Please use sub-agents to help you with these tasks. Consider using the @agent-deep-research-specialist and @agent-crypto-smart-contracts-researcher for this.

I want you to explore, plan and implement the following:

$ARGUMENTS
