#!/bin/bash
# PostToolUse hook: run golangci-lint on edited .go files for immediate feedback.
# Parses CLAUDE_TOOL_INPUT to extract the file path from Edit/Write/MultiEdit.

FILE=$(echo "$CLAUDE_TOOL_INPUT" | jq -r '.file_path // empty')
if [[ -z "$FILE" ]] || [[ "$FILE" != *.go ]] || [[ ! -f "$FILE" ]]; then
  exit 0
fi

DIR=$(dirname "$FILE")
OUTPUT=$(golangci-lint run --timeout=15s "$DIR/..." 2>&1)
EXIT_CODE=$?

if [[ $EXIT_CODE -ne 0 ]] && [[ -n "$OUTPUT" ]]; then
  echo "$OUTPUT" | head -20
  exit 1
fi
