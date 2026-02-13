# Defense-in-Depth Validation

## Overview

When you fix a bug caused by invalid data, adding validation at one place feels sufficient. But that single check can be bypassed by different code paths, refactoring, or mocks.

**Core principle:** Validate at EVERY layer data passes through. Make the bug structurally impossible.

## Why Multiple Layers

Single validation: "We fixed the bug"
Multiple layers: "We made the bug impossible"

Different layers catch different cases:
- Entry validation catches most bugs
- Business logic catches edge cases
- Environment guards prevent context-specific dangers
- Debug logging helps when other layers fail

## The Four Layers

### Layer 1: Entry Point Validation
**Purpose:** Reject obviously invalid input at API boundary

```go
func CreateProject(name, workingDir string) (*Project, error) {
    if workingDir == "" {
        return nil, fmt.Errorf("workingDir cannot be empty")
    }

    info, err := os.Stat(workingDir)
    if err != nil {
        return nil, fmt.Errorf("workingDir does not exist: %w", err)
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("workingDir is not a directory: %s", workingDir)
    }

    // ... proceed
}
```

### Layer 2: Business Logic Validation
**Purpose:** Ensure data makes sense for this operation

```go
func (w *WorkspaceManager) Initialize(projectDir, sessionID string) error {
    if projectDir == "" {
        return fmt.Errorf("projectDir required for workspace initialization")
    }
    if sessionID == "" {
        return fmt.Errorf("sessionID required for workspace initialization")
    }

    // ... proceed
}
```

### Layer 3: Environment Guards
**Purpose:** Prevent dangerous operations in specific contexts

```go
func (s *Store) Write(path string, data []byte) error {
    resolved, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("resolving path: %w", err)
    }

    // In tests, refuse writes outside temp directories
    if testing.Short() {
        tmpDir := os.TempDir()
        if !strings.HasPrefix(resolved, tmpDir) {
            return fmt.Errorf("refusing write outside temp dir during tests: %s", resolved)
        }
    }

    // ... proceed
}
```

### Layer 4: Debug Instrumentation
**Purpose:** Capture context for forensics

```go
func (s *Store) Write(path string, data []byte) error {
    s.logger.Debug("writing to store",
        zap.String("path", path),
        zap.String("resolved", resolved),
        zap.Int("data_len", len(data)),
    )

    // ... proceed
}
```

## Applying the Pattern

When you find a bug:

1. **Trace the data flow** - Where does bad value originate? Where used?
2. **Map all checkpoints** - List every point data passes through
3. **Add validation at each layer** - Entry, business, environment, debug
4. **Test each layer** - Try to bypass layer 1, verify layer 2 catches it

## Example from Session

Bug: Empty `storePath` caused files written to working directory

**Data flow:**
1. Config loaded → empty string for `StorePath`
2. `NewStore("")` called
3. `store.Write(key, data)` resolves to `./key`
4. Files appear in project root

**Four layers added:**
- Layer 1: `NewStore()` returns error if path empty or doesn't exist
- Layer 2: `LoadConfig()` validates all required fields after unmarshal
- Layer 3: `Store.Write()` refuses paths outside configured root in short tests
- Layer 4: Debug logging of resolved paths before every write

**Result:** All tests passed, bug impossible to reproduce

## Key Insight

All four layers were necessary. During testing, each layer caught bugs the others missed:
- Different code paths bypassed entry validation
- Mocks bypassed business logic checks
- Edge cases on different platforms needed environment guards
- Debug logging identified structural misuse

**Don't stop at one validation point.** Add checks at every layer.
