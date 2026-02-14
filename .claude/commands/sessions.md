# Sessions Command

Manage Claude Code session history - list, search, alias, and inspect sessions.

## Usage

`/sessions [list|info|alias|search|help] [options]`

## Actions

### List Sessions

Display all sessions with metadata and pagination.

```bash
/sessions                              # List all sessions (default)
/sessions list                         # Same as above
/sessions list --limit 10              # Show 10 sessions
```

**Script:**

```bash
cc-tools session list --limit 20
```

### Session Info

Show detailed information about a session (outputs JSON).

```bash
/sessions info <id|alias>              # Show session details
```

**Script:**

```bash
cc-tools session info "$ARGUMENTS"
```

### Create Alias

Create a memorable alias for a session.

```bash
/sessions alias set <name> <session-id>       # Create alias
/sessions alias set today-work a1b2c3d4       # Example
```

**Script:**

```bash
cc-tools session alias set "$ARGUMENTS"
```

### Remove Alias

Delete an existing alias.

```bash
/sessions alias remove <name>          # Remove alias
```

**Script:**

```bash
cc-tools session alias remove "$ARGUMENTS"
```

### List Aliases

Show all session aliases.

```bash
/sessions aliases                      # List all aliases
```

**Script:**

```bash
cc-tools session alias list
```

### Search Sessions

Search sessions by title or summary.

```bash
/sessions search <query>               # Search sessions
```

**Script:**

```bash
cc-tools session search "$ARGUMENTS"
```

## Arguments

$ARGUMENTS:

- `list [options]` - List sessions
  - `--limit <n>` - Max sessions to show (default: 50)
- `info <id|alias>` - Show session details (JSON output)
- `alias set <name> <session-id>` - Create alias for session
- `alias remove <name>` - Remove alias
- `alias list` - List all aliases
- `aliases` - Same as `alias list`
- `search <query>` - Search sessions by title or summary
- `help` - Show this help

## Examples

```bash
# List all sessions
/sessions list

# List with limit
/sessions list --limit 10

# Show session info
/sessions info a1b2c3d4

# Create an alias for a session
/sessions alias set today a1b2c3d4

# Show session info by alias
/sessions info today

# Remove alias
/sessions alias remove today

# List all aliases
/sessions aliases

# Search sessions
/sessions search "refactoring hooks"
```

## Notes

- Sessions are stored as JSON files, managed by cc-tools
- Aliases map human-readable names to session IDs
- Session IDs can be shortened (first 4-8 characters usually unique enough)
- Use aliases for frequently referenced sessions
- The `info` command outputs JSON for easy parsing and integration
