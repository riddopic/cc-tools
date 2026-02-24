# Troubleshooting

This guide covers common issues you may encounter when using cc-tools, with symptoms, causes, and step-by-step solutions.

## Hook not firing

**Problem:** You configured cc-tools in `~/.claude/settings.json` but hooks do not run.

**Cause:** The hook configuration may have incorrect structure, or cc-tools is not in your PATH.

**Solution:**

1. Verify cc-tools is accessible:

   ```bash
   which cc-tools
   ```

   If this returns nothing, add the directory containing `cc-tools` to your PATH.

2. Check that your `~/.claude/settings.json` uses the correct hook structure:

   ```json
   {
     "hooks": {
       "PreToolUse": [
         {
           "matcher": "*",
           "command": "cc-tools hook"
         }
       ],
       "PostToolUse": [
         {
           "matcher": "*",
           "command": "cc-tools hook"
         }
       ]
     }
   }
   ```

3. Ensure the `matcher` value matches the event you want to intercept. Use `"*"` to catch all events during initial debugging.

4. Enable debug logging and check the log file for errors:

   ```bash
   cc-tools debug enable
   cc-tools debug filename
   ```

## Validation blocking legitimate edits

**Problem:** `cc-tools validate` blocks file edits even though your code is correct.

**Cause:** The linter or test suite may be failing for reasons unrelated to the current edit, such as pre-existing failures or flaky tests.

**Solution:**

1. Run each check manually to identify the actual failure:

   ```bash
   task lint
   task test
   ```

2. Skip validation temporarily for the current directory:

   ```bash
   cc-tools skip all
   ```

3. If commands are timing out, increase the timeout:

   ```bash
   cc-tools config set validate.timeout 120
   ```

4. Re-enable validation after fixing the underlying issues:

   ```bash
   cc-tools unskip all
   ```

## Skip registry not applying

**Problem:** You set `cc-tools skip lint` but validation still runs linting.

**Cause:** The skip registry is per-directory. You may have set the skip in a different directory than where validation runs.

**Solution:**

1. Check the skip status for your current directory:

   ```bash
   cc-tools skip status
   ```

2. List all skip configurations across directories:

   ```bash
   cc-tools skip list
   ```

3. Run `cc-tools skip` from the project root where validation executes. The skip applies only to the directory where you set it.

## Drift detection false positives

**Problem:** Drift detection warns about topic divergence when you intentionally switched topics.

**Cause:** The drift handler tracks keywords from your first prompt. Changing topics triggers warnings unless you use recognized pivot phrases.

**Solution:**

1. Use pivot phrases to reset intent tracking. Phrases like "now let's", "switch to", and "moving on to" signal an intentional topic change.

2. Increase the divergence threshold to reduce sensitivity:

   ```bash
   cc-tools config set drift.threshold 0.3
   ```

3. Increase the minimum number of edits before detection activates:

   ```bash
   cc-tools config set drift.min_edits 10
   ```

4. Disable drift detection entirely if it does not fit your workflow:

   ```bash
   cc-tools config set drift.enabled false
   ```

## Notifications not arriving

**Problem:** You expected a notification but did not receive one.

**Cause:** Quiet hours may be active, audio may be disabled, or the push notification topic is not configured.

**Solution:**

1. Check whether quiet hours are active:

   ```bash
   cc-tools config get notify.quiet_hours.enabled
   cc-tools config get notify.quiet_hours.start
   ```

2. Verify audio notifications are enabled:

   ```bash
   cc-tools config get notify.audio.enabled
   ```

3. Confirm that audio files exist in the expected directory:

   ```bash
   ls ~/.claude/audio/
   ```

4. For push notifications, set your ntfy topic:

   ```bash
   cc-tools config set notifications.ntfy_topic YOUR_TOPIC
   ```

5. Adjust quiet hours if the current window is too broad:

   ```bash
   cc-tools config set notify.quiet_hours.start 23:00
   ```

## Instinct import failures

**Problem:** `cc-tools instinct import` fails or imports nothing.

**Cause:** The source file may have an unsupported format, or all instincts in the file already exist locally.

**Solution:**

1. Preview the import without making changes:

   ```bash
   cc-tools instinct import source.yaml --dry-run
   ```

2. Ensure the source file is valid YAML or JSON. Use a linter or validator if you are unsure.

3. Overwrite existing instincts with `--force`:

   ```bash
   cc-tools instinct import source.yaml --force
   ```

4. Lower the confidence filter to include lower-confidence instincts:

   ```bash
   cc-tools instinct import source.yaml --min-confidence 0
   ```

## Instinct export empty

**Problem:** `cc-tools instinct export` outputs "No instincts to export."

**Cause:** No instincts have been created yet, or existing instincts fall below the default confidence filter.

**Solution:**

1. Check instinct status to see what exists:

   ```bash
   cc-tools instinct status
   ```

2. Lower the confidence filter to include all instincts:

   ```bash
   cc-tools instinct export --min-confidence 0
   ```

3. Ensure observation logging is enabled so instincts can be generated:

   ```bash
   cc-tools config get observe.enabled
   ```

## Debug logging setup

**Problem:** You need to diagnose cc-tools behavior but do not know where logs are stored.

**Solution:**

1. Enable debug logging:

   ```bash
   cc-tools debug enable
   ```

2. Find the log file path:

   ```bash
   cc-tools debug filename
   ```

3. Check whether debug logging is currently active:

   ```bash
   cc-tools debug status
   ```

4. Logs are written to `~/.cache/cc-tools/debug/`. You can also set the `CLAUDE_HOOKS_DEBUG=1` environment variable for verbose output to stderr.

5. Disable debug logging when you are done:

   ```bash
   cc-tools debug disable
   ```

## Pre-commit check failures

**Problem:** `task pre-commit` (or `task check`) fails before committing.

**Cause:** Code formatting, linting, or test failures need to be resolved before the commit can proceed.

**Solution:**

1. Run each check individually to isolate the failure:

   ```bash
   task fmt    # Fix formatting issues
   task lint   # Identify linter warnings
   task test   # Run the test suite
   ```

2. Run tests with the race detector to catch concurrency issues:

   ```bash
   task test-race
   ```

3. Check for common issues:
   - **Missing or incorrect imports:** Run `goimports -w .` to fix import grouping and unused imports.
   - **Unused variables or functions:** Remove them or prefix with `_` if intentionally unused.
   - **Unchecked errors:** Wrap or handle every returned error value.

## Compact suggestions too frequent

**Problem:** cc-tools suggests compacting context too often.

**Cause:** The compact threshold or reminder interval is set too low for your workflow.

**Solution:**

1. Increase the threshold before compact suggestions trigger:

   ```bash
   cc-tools config set compact.threshold 100
   ```

2. Increase the interval between reminders:

   ```bash
   cc-tools config set compact.reminder_interval 50
   ```

## MCP server issues

**Problem:** `cc-tools mcp list` shows no servers, or you cannot enable or disable servers.

**Cause:** MCP servers are defined in Claude Code's own configuration, not in the cc-tools config file.

**Solution:**

1. Check that MCP servers are defined in `~/.claude/.mcp.json` or your project-level `.mcp.json`.

2. List available servers and their current status:

   ```bash
   cc-tools mcp list
   ```

3. Enable a specific server:

   ```bash
   cc-tools mcp enable <name>
   ```

4. Enable all configured servers:

   ```bash
   cc-tools mcp enable-all
   ```

## Getting help

If you encounter an issue not covered in this guide:

1. Enable debug logging and reproduce the issue.
2. Check the log file for error details: `cc-tools debug filename`.
3. Confirm the installed version: `cc-tools version`.
4. File an issue with the debug log output attached.
