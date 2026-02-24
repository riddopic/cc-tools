---
name: writing-skills
description: Use when creating new skills, editing existing skills, or verifying skills work before deployment
---

# Writing Skills

## Overview

**Writing skills IS Test-Driven Development applied to process documentation.**

You write test cases (pressure scenarios with subagents), watch them fail (baseline behavior), write the skill (documentation), watch tests pass (agents comply), and refactor (close loopholes).

**Core principle:** If you didn't watch an agent fail without the skill, you don't know if the skill teaches the right thing.

**REQUIRED BACKGROUND:** You MUST understand tdd-workflow before using this skill.

**Personal skills** live in agent-specific directories (`~/.claude/skills` for Claude Code, `~/.codex/skills` for Codex).

## When to Create a Skill

**Create when:** technique wasn't obvious, you'd reference it again, pattern applies broadly, others would benefit.

**Don't create for:** one-off solutions, well-documented standard practices, project-specific conventions (use CLAUDE.md), mechanically enforceable constraints (automate instead).

## Skill Types

- **Technique**: Concrete method with steps (condition-based-waiting, root-cause-tracing)
- **Pattern**: Way of thinking about problems (flatten-with-flags, test-invariants)
- **Reference**: API docs, syntax guides, tool documentation

## Directory Structure

```
skills/
  skill-name/
    SKILL.md              # Main reference (required)
    supporting-file.*     # Only if needed (heavy reference 100+ lines, reusable tools)
```

Keep inline: principles, concepts, code patterns (<50 lines), everything else.

## SKILL.md Structure

**Frontmatter (YAML) -- only two fields supported:**
- `name`: Letters, numbers, hyphens only. No parentheses or special chars.
- `description`: Third-person, max 1024 chars total. Start with "Use when..." describing ONLY triggering conditions.

**CRITICAL:** Description must describe when to use, NOT what the skill does. Testing revealed that workflow summaries in descriptions cause Claude to follow the description shortcut instead of reading the full skill body.

```markdown
---
name: Skill-Name-With-Hyphens
description: Use when [specific triggering conditions and symptoms]
---

# Skill Name

## Overview
Core principle in 1-2 sentences.

## When to Use
Bullet list with SYMPTOMS and use cases. When NOT to use.
[Small inline flowchart IF decision non-obvious]

## Core Pattern (for techniques/patterns)
Before/after code comparison

## Quick Reference
Table or bullets for scanning common operations

## Implementation
Inline code for simple patterns. Link to file for heavy reference.

## Common Mistakes
What goes wrong + fixes
```

## Search Optimization (CSO)

**Description field:** Use concrete triggers, symptoms, situations. Describe the problem, not language-specific symptoms. Keep technology-agnostic unless the skill itself is technology-specific.

**Keyword coverage:** Include error messages, symptoms, synonyms, tool names that Claude would search for.

**Naming:** Use active voice, verb-first (e.g., `condition-based-waiting` not `async-test-helpers`). Gerunds work well for processes.

**Token efficiency targets:**
- getting-started workflows: <150 words
- Frequently-loaded skills: <200 words
- Other skills: <500 words

**Techniques:** Move details to `--help`, use cross-references to other skills, compress examples, eliminate redundancy.

**Cross-referencing:** Use `**REQUIRED SUB-SKILL:** Use skill-name` or `**REQUIRED BACKGROUND:** You MUST understand skill-name`. Never use `@` links (force-loads files, burns context).

## The Iron Law

```
NO SKILL WITHOUT A FAILING TEST FIRST
```

Applies to new skills AND edits. Write skill before testing? Delete it. Start over. No exceptions -- not for "simple additions", not for "just adding a section", not for "documentation updates."

## RED-GREEN-REFACTOR for Skills

1. **RED:** Run pressure scenario WITHOUT skill. Document exact rationalizations agents use.
2. **GREEN:** Write minimal skill addressing those specific rationalizations. Re-run -- agent should comply.
3. **REFACTOR:** Agent found new rationalization? Add explicit counter. Re-test until bulletproof.

**Testing methodology:** See testing-skills-with-subagents.md for pressure scenarios, pressure types, and plugging holes.

## Testing by Skill Type

- **Discipline skills** (TDD, verification): Test with pressure scenarios (time + sunk cost + exhaustion). Success = agent follows rule under maximum pressure.
- **Technique skills** (how-to): Test with application and variation scenarios. Success = agent applies technique to new scenario.
- **Pattern skills** (mental models): Test with recognition and counter-examples. Success = agent correctly identifies when/how to apply.
- **Reference skills** (docs/APIs): Test with retrieval and gap scenarios. Success = agent finds and applies reference info.

## Skill Creation Checklist

**RED Phase:**
- [ ] Create pressure scenarios (3+ combined pressures for discipline skills)
- [ ] Run scenarios WITHOUT skill -- document baseline behavior verbatim
- [ ] Identify patterns in rationalizations/failures

**GREEN Phase:**
- [ ] YAML frontmatter with only name and description (max 1024 chars)
- [ ] Description starts with "Use when..." with specific triggers/symptoms
- [ ] Keywords throughout for search (errors, symptoms, tools)
- [ ] Address specific baseline failures identified in RED
- [ ] One excellent example (not multi-language)
- [ ] Run scenarios WITH skill -- verify compliance

**REFACTOR Phase:**
- [ ] Add explicit counters for new rationalizations
- [ ] Build rationalization table from all test iterations
- [ ] Create red flags list
- [ ] Re-test until bulletproof

**Quality Checks:**
- [ ] Flowcharts only if decision non-obvious
- [ ] Quick reference table present
- [ ] Common mistakes section present
- [ ] No narrative storytelling
- [ ] Supporting files only for tools or heavy reference

**Deployment:**
- [ ] Commit and push
- [ ] Consider contributing back via PR if broadly useful
