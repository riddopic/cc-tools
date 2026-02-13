#!/bin/bash
# PreToolUse hook: remind to run pre-commit checks before git commit.
# Non-blocking — prints a reminder but does not prevent the commit.

CMD=$(echo "$CLAUDE_TOOL_INPUT" | jq -r '.command // empty')
if echo "$CMD" | grep -q "git commit"; then
  echo "Reminder: Run 'make pre-commit' (fmt + lint + test) before committing."
fi
