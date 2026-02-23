---
name: reviewing-with-codex
description: Use when reviewing an implementation plan with OpenAI Codex CLI, or when the user requests a second opinion on a plan from a different model
---

# Reviewing Plans with Codex

## Overview

Iterative review loop: Codex reviews the plan, Claude revises based on feedback, re-submits until Codex approves. Max 5 rounds with session persistence.

Unlike `/consult codex` (single-shot query), this skill implements a multi-round review loop with active plan revision between rounds and structured verdict tracking.

**Announce at start:** "I'm using the reviewing-with-codex skill to review this plan."

## When to Use

- User runs `/reviewing-with-codex [model-name]`
- User wants iterative external review of an implementation plan
- User asks for a second opinion from a different model on a plan

## Quick Reference

| Setting        | Default                                 |
| -------------- | --------------------------------------- |
| Default model  | `gpt-5.3-codex` (override via argument) |
| Max rounds     | 5                                       |
| Sandbox mode   | `read-only`                             |
| Temp directory | Created via `mktemp -d`                 |

## Process

### Step 0: Verify Prerequisites

Before doing anything else:

1. Verify `codex` is on PATH: `command -v codex`
2. If missing, inform the user and suggest `npm install -g @openai/codex`
3. Verify API key is configured (check `~/.codex/config.toml` exists)
4. If either check fails, stop and report the issue

### Step 1: Generate Session ID

Generate a unique ID and a secure temp directory:

```bash
REVIEW_ID=$(uuidgen | tr '[:upper:]' '[:lower:]' | head -c 8 || date +%s | head -c 8)
REVIEW_DIR=$(mktemp -d)
```

Use `$REVIEW_DIR/plan.md` and `$REVIEW_DIR/review.md` for all temp files.

**Early exit rule:** If the review exits early for any reason (error, interrupt, context limit), clean up `$REVIEW_DIR` before reporting.

### Step 2: Capture the Plan

Write the current plan to `$REVIEW_DIR/plan.md`. The plan comes from the conversation context (plan mode output or a plan discussed in chat).

1. Write the full plan content to `$REVIEW_DIR/plan.md`
2. If no plan exists in context, ask the user what they want reviewed
3. If the plan exceeds ~8000 words, warn the user it may be truncated by Codex and suggest splitting into sections

### Step 3: Submit to Codex (Round 1)

Run Codex CLI in non-interactive mode. Use the default model (`gpt-5.3-codex`) unless the user specified a different one via argument:

```bash
timeout 120 codex exec \
  -m MODEL \
  -s read-only \
  -o $REVIEW_DIR/review.md \
  "Review the implementation plan in $REVIEW_DIR/plan.md. Focus on:
1. Correctness - Will this plan achieve the stated goals?
2. Risks - What could go wrong? Edge cases? Data loss?
3. Missing steps - Is anything forgotten?
4. Alternatives - Is there a simpler or better approach?
5. Security - Any security concerns?

Be specific and actionable. If the plan is solid and ready to implement, end your review with exactly: VERDICT: APPROVED

If changes are needed, end with exactly: VERDICT: REVISE"
```

Capture the Codex session ID from the output line `session id: <uuid>`. Store as `CODEX_SESSION_ID`. Use this exact ID for resume (not `--last`, which grabs the wrong session when multiple reviews run concurrently).

If the command times out (exit code 124), report the timeout and ask the user whether to retry or abort.

### Step 4: Read Review and Check Verdict

1. Read `$REVIEW_DIR/review.md`
2. Present to the user:

```
## Codex Review - Round N (model: MODEL)

[Codex's feedback]
```

3. Parse the verdict from the **last 10 lines**, case-insensitively:
   - `VERDICT: APPROVED` (any casing) -> Step 7
   - `VERDICT: REVISE` (any casing) -> Step 5
   - No clear verdict but no actionable feedback -> treat as approved
   - No clear verdict with actionable feedback -> ask the user whether to revise or accept
   - Max rounds (5) reached -> Step 7 with a note

### Step 5: Revise the Plan

Based on Codex's feedback:

1. **Check each suggestion against user requirements** - if a revision contradicts the user's explicit requirements, skip it and note why
2. **Revise the plan** - address each valid issue. Update both the conversation context and `$REVIEW_DIR/plan.md`
3. **Summarize** what changed:

```
### Revisions (Round N)
- [What changed and why, one bullet per issue addressed]
- Skipped: [Any skipped suggestions with reason]
```

4. Inform the user: "Sending revised plan back to Codex for re-review..."

### Step 6: Re-submit to Codex (Rounds 2-5)

Resume the existing session for full prior context. Note: `codex exec resume` does not support the `-o` flag, so capture from stdout:

```bash
timeout 120 codex exec resume ${CODEX_SESSION_ID} \
  "I've revised the plan based on your feedback. The updated plan is in $REVIEW_DIR/plan.md.

Here's what I changed:
[List the specific changes made]

Please re-review. If the plan is now solid, end with: VERDICT: APPROVED
If more changes needed, end with: VERDICT: REVISE" 2>&1 | tee $REVIEW_DIR/review.md
```

If `resume` fails (session expired), fall back to a fresh `codex exec` with `-o` flag and include context about prior rounds in the prompt.

Return to **Step 4**.

### Step 7: Present Final Result

Once approved (or max rounds reached), present the outcome to the user. For max rounds without approval, list remaining concerns so the user can decide whether to proceed.

### Step 8: Cleanup

```bash
rm -rf $REVIEW_DIR
```

## Common Mistakes

- **Codex not installed** - always check in Step 0, not after writing temp files
- **Session expired on resume** - fall back to fresh `codex exec` with prior context
- **No plan in context** - ask the user instead of sending an empty file
- **Plan too large** - warn about potential truncation; suggest splitting
- **Applying revisions that contradict user requirements** - always check before applying
- **Verdict not found** - search last 10 lines case-insensitively; ask user if ambiguous

## Related Skills

- **writing-plans** - creates the plans this skill reviews
- **executing-plans** - executes plans after review approval
- **brainstorming** - explores ideas before plan creation

## Rules

- Claude **actively revises** the plan between rounds - not just passing messages
- Accept model override from user arguments (e.g., `/reviewing-with-codex o4-mini`)
- Always use read-only sandbox - Codex should never write files
- Max 5 rounds to prevent infinite loops
- Show the user each round's feedback and revisions
