---
description: "Evidence-based skill quality audit derived from SkillsBench research. Evaluates skills against empirical criteria for procedural quality, length discipline, and pretraining gap coverage. Supports Quick Scan and Full Audit modes."
---

# skill-audit

Audit all Claude skills against evidence-based quality criteria from SkillsBench (arXiv 2602.12670). Evaluates each skill for procedural quality, appropriate length, uniqueness, and pretraining gap targeting. Supports Quick Scan (changed skills only) and Full Audit (complete review).

## Evidence Base

Audit criteria derive from SkillsBench (84 tasks, 7,308 trajectories, 7 model-harness configurations):

| Finding                                 | Number                         | Audit Impact                |
| --------------------------------------- | ------------------------------ | --------------------------- |
| Curated skills improve pass rate        | +16.2pp average                | Validates audit value       |
| 2-3 skills optimal per task class       | +18.6pp; 4+ drops to +5.9pp    | Flag co-loading >3          |
| Self-generated skills hurt              | -1.3pp average                 | Flag auto-generated content |
| Detailed/Compact format best            | +18.8pp / +17.1pp              | Flag appropriate length     |
| Comprehensive format hurts              | -2.9pp                         | Flag exhaustive skills      |
| Procedural >> encyclopedic              | "focused procedural guidance"  | Flag walls of facts         |
| Skills help most where pretraining weak | Healthcare +51.9pp, SWE +4.5pp | Flag model-known content    |
| Working examples required               | Paper criterion #1             | Verify examples exist       |

## Scope

The command targets these paths **relative to the directory where it is invoked**:

| Path                    | Description                                    |
| ----------------------- | ---------------------------------------------- |
| `~/.claude/skills/`     | Global skills (all projects)                   |
| `{cwd}/.claude/skills/` | Project-level skills (if the directory exists) |

**At the start of Phase 1, the command explicitly lists which paths were found and scanned.**

### Targeting a specific project

To include project-level skills, run from that project's root directory:

```bash
cd ~/path/to/my-project
/skill-audit
```

If the project has no `.claude/skills/` directory, only global skills are evaluated.

## Modes

| Mode       | Trigger                                       |
| ---------- | --------------------------------------------- |
| Quick Scan | `results.json` exists (default)               |
| Full Audit | `results.json` absent, or `/skill-audit full` |

**Results cache:** `~/.claude/skills/skill-audit/results.json`

## Quick Scan Flow

Re-evaluate only skills that have changed since the last run.

1. Read `~/.claude/skills/skill-audit/results.json`
2. Run: `bash ~/.claude/skills/skill-audit/scripts/quick-diff.sh \
      ~/.claude/skills/skill-audit/results.json`
   (Project dir is auto-detected from `$PWD/.claude/skills`; pass it explicitly only if needed)
3. If output is `[]`: report "No changes since last run." and stop
4. Re-evaluate only those changed files using the same Phase 2 criteria
5. Carry forward unchanged skills from previous results
6. Output only the diff
7. Run: `bash ~/.claude/skills/skill-audit/scripts/save-results.sh \
     ~/.claude/skills/skill-audit/results.json <<< "$EVAL_RESULTS"`

## Full Audit Flow

### Phase 1 — Inventory

Run: `bash ~/.claude/skills/skill-audit/scripts/scan.sh`

The script enumerates skill files, extracts frontmatter, and collects UTC mtimes.
Project dir is auto-detected from `$PWD/.claude/skills`; pass it explicitly only if needed.
Present the scan summary and inventory table from the script output:

```
Scanning:
  ✓ ~/.claude/skills/         (17 files)
  ✗ {cwd}/.claude/skills/    (not found — global skills only)
```

| Skill | 7d use | 30d use | Description |

### Phase 2 — Quality Evaluation

Launch a Task tool subagent (**Explore agent, model: opus**) with the full inventory and checklist.
The subagent reads each skill, applies the checklist, and returns per-skill JSON:

`{ "verdict": "Keep"|"Improve"|"Update"|"Retire"|"Merge into [X]", "reason": "..." }`

**Chunk guidance:** Process ~20 skills per subagent invocation to keep context manageable. Save intermediate results to `results.json` (`status: "in_progress"`) after each chunk.

After all skills are evaluated: set `status: "completed"`, proceed to Phase 3.

**Resume detection:** If `status: "in_progress"` is found on startup, resume from the first unevaluated skill.

Each skill is evaluated against this checklist:

```
Evidence-derived criteria (from SkillsBench):
- [ ] Procedural quality: provides stepwise guidance, not walls of facts
- [ ] Length discipline: Detailed or Compact format, NOT Comprehensive/exhaustive
- [ ] Working examples: at least one working example (code, command, or template)
- [ ] Pretraining gap: targets knowledge underrepresented in model pretraining
- [ ] Skill budget: would not push any task class above 3 co-loading skills

Operational criteria:
- [ ] Content overlap with other skills checked
- [ ] Overlap with CLAUDE.md / .claude/rules/ files checked
- [ ] Freshness of technical references verified (use WebSearch if tool names / CLI flags / APIs are present)
- [ ] Usage frequency considered
```

Verdict criteria:

| Verdict        | Meaning                                                                                                 |
| -------------- | ------------------------------------------------------------------------------------------------------- |
| Keep           | Useful, current, and meets evidence-based criteria                                                      |
| Improve        | Worth keeping, but specific improvements needed (see reason)                                            |
| Update         | Referenced technology is outdated (verify with WebSearch)                                               |
| Retire         | Low quality, stale, covers model-known content, or Comprehensive-length with no unique procedural value |
| Merge into [X] | Substantial overlap with another skill; name the merge target                                           |

Evaluation is **holistic AI judgment** informed by SkillsBench evidence. Guiding dimensions:

- **Procedural quality**: stepwise guidance with commands/steps you can act on immediately (not encyclopedic facts)
- **Length discipline**: Detailed (~50-150 lines) or Compact (~20-50 lines) — NOT Comprehensive (>200 lines of exhaustive documentation, which the paper shows hurts performance by -2.9pp)
- **Scope fit**: name, trigger, and content are aligned; not too broad or narrow
- **Uniqueness**: value not replaceable by CLAUDE.md / `.claude/rules/` / another skill
- **Pretraining gap**: covers knowledge models lack (domain-specific conventions, project-specific patterns) rather than standard patterns models already know well
- **Currency**: technical references work in the current environment

**Reason quality requirements** — the `reason` field must be self-contained and decision-enabling:

- Do NOT write "unchanged" alone — always restate the core evidence
- For **Retire**: state (1) what specific defect was found, (2) what covers the same need instead
- For **Merge**: name the target and describe what content to integrate
- For **Improve**: describe the specific change needed (what section, what action, target size if relevant)
- For **Keep** (mtime-only change in Quick Scan): restate the original verdict rationale, do not write "unchanged"

### Phase 3 — Skill Budget Analysis

After individual evaluations, analyze co-loading patterns:

1. Read `rules/agents.md` and `using-superpowers` skill to identify which skills trigger together
2. For each task class (e.g., "writing Go code", "reviewing code", "debugging"), list which skills would co-load
3. Flag any task class where >3 skills would co-load (68% effectiveness drop per SkillsBench)
4. Recommend which skills to merge or retire to stay within the 2-3 skill budget

Present as:

| Task Class      | Skills That Co-Load                                                              | Count | Status      |
| --------------- | -------------------------------------------------------------------------------- | ----- | ----------- |
| Writing Go code | go-coding-standards, tdd-workflow, testing-patterns              | 3     | OK          |
| Code review     | code-review, go-coding-standards, testing-patterns              | 3     | OK          |

### Phase 4 — Summary Table

| Skill | 7d use | Verdict | Reason |
| ----- | ------ | ------- | ------ |

### Phase 5 — Consolidation

1. **Retire / Merge**: present detailed justification per file before confirming with user:
   - What specific problem was found (overlap, staleness, broken references, Comprehensive length, covers model-known territory)
   - What alternative covers the same functionality (for Retire: which existing skill/rule; for Merge: the target file and what content to integrate)
   - Impact of removal (any dependent skills, workflow references affected)
2. **Improve**: present specific improvement suggestions with rationale:
   - What to change and why (e.g., "trim 430->200 lines because sections X/Y duplicate python-patterns")
   - User decides whether to act
3. **Update**: present updated content with sources checked
4. **Skill budget fixes**: if any task class is over 3 skills, propose specific merges/retires to bring it under budget

## Results File Schema

`~/.claude/skills/skill-audit/results.json`:

Key fields:
- **`evaluated_at`**: UTC timestamp of evaluation completion. Obtain via `date -u +%Y-%m-%dT%H:%M:%SZ`. Never use date-only approximations.
- **`mode`**: `"full"` or `"quick"`
- **`batch_progress`**: `{ total, evaluated, status }` — tracks chunked evaluation progress. `status` is `"in_progress"` or `"completed"`.
- **`skill_budget.task_classes`**: Map of task class name to `{ skills[], count, status }` for co-loading analysis.
- **`skills`**: Map of skill name to `{ path, verdict, reason, mtime }` — the per-skill evaluation results.

## Notes

- Evaluation is blind: the same checklist applies to all skills regardless of origin
- Archive/delete operations always require explicit user confirmation
- The paper's key insight: "models cannot reliably author the procedural knowledge they benefit from consuming" — prioritize human-curated skills over auto-generated ones
