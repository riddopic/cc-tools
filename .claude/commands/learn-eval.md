---
description: Extract reusable patterns from the session, self-evaluate quality before saving, and determine the right save location (Global vs Project).
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
---

# /learn-eval - Extract, Evaluate, then Save

Adds a quality gate and save-location decision on top of `/learn`. Run `/learn` extraction first, then apply this evaluation before writing.

## Process

1. Run the `/learn` extraction process to identify patterns and draft a skill file

2. **Determine save location:**
   - Ask: "Would this pattern be useful in a different project?"
   - **Project** (`.claude/skills/learned/` in current project): Default choice. Project-specific knowledge, architecture decisions, integration patterns. Checked into version control and shared with collaborators.
   - **Global** (`~/.claude/skills/learned/`): Generic patterns usable across 2+ projects (bash compatibility, LLM API behavior, debugging techniques, etc.). Machine-local, not shared via git.
   - When in doubt, choose Project (version-controlled and shareable; moving Project → Global is easy if it proves broadly useful)

3. Verify the draft uses this format (matching existing learned skills):

   ```markdown
   # [Descriptive Pattern Name]

   **Extracted:** [Date]
   **Context:** [Brief description of when this applies]

   ## Problem
   [What problem this solves - be specific]

   ## Solution
   [The pattern/technique/workaround - with code examples]

   ## When to Use
   [Trigger conditions]
   ```

4. **Self-evaluate before saving** using this rubric:

   | Dimension | 1 | 3 | 5 |
   |-----------|---|---|---|
   | Specificity | Abstract principles only, no code examples | Representative code example present | Rich examples covering all usage patterns |
   | Actionability | Unclear what to do | Main steps are understandable | Immediately actionable, edge cases covered |
   | Scope Fit | Too broad or too narrow | Mostly appropriate, some boundary ambiguity | Name, trigger, and content perfectly aligned |
   | Non-redundancy | Nearly identical to another skill | Some overlap but unique perspective exists | Completely unique value |
   | Coverage | Covers only a fraction of the target task | Main cases covered, common variants missing | Main cases, edge cases, and pitfalls covered |

   - Score each dimension 1–5
   - **Gate:** All dimensions must be ≥ 3, AND total must be ≥ 18/25
   - If any dimension scores 1–2, or total is below 18, improve the draft and re-score
   - Show the user the scores table and the final draft

5. Ask user to confirm:
   - Show: proposed save path + scores table + final draft
   - Wait for explicit confirmation before writing

6. Save to the determined location

## Output Format for Step 4 (scores table)

| Dimension | Score | Rationale |
|-----------|-------|-----------|
| Specificity | N/5 | ... |
| Actionability | N/5 | ... |
| Scope Fit | N/5 | ... |
| Non-redundancy | N/5 | ... |
| Coverage | N/5 | ... |
| **Total** | **N/25** | |

## Notes

- Don't extract trivial fixes (typos, simple syntax errors)
- Don't extract one-time issues (specific API outages, etc.)
- Focus on patterns that will save time in future sessions
- Keep skills focused — one pattern per skill
- If Coverage score is low, add related variants before saving
