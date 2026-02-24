# Instinct Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Migrate the continuous-learning-v2 instinct system from Python/Bash into native Go as `cc-tools instinct` subcommands, eliminating duplicate observation hooks and consolidating configuration.

**Architecture:** New `internal/instinct/` package with Store interface for file-based YAML frontmatter storage at `~/.config/cc-tools/instincts/`. New `cc-tools instinct` Cobra command group with status, export, import, and evolve subcommands. Extend existing `observe.Event` to capture tool output. Add `InstinctValues` to existing config system.

**Tech Stack:** Go 1.26, Cobra, testify, gotestsum, golangci-lint, YAML frontmatter parsing (custom, no external dependency)

**Design doc:** `docs/plans/2026-02-24-instinct-migration-design.md`

---

## Task 1: Extend observe.Event with ToolOutput and Error

**Files:**
- Modify: `internal/observe/observe.go:18-25`
- Test: `internal/observe/observe_test.go`

**Step 1: Write the failing test**

Add a new test case to `internal/observe/observe_test.go` that verifies `ToolOutput` and `Error` fields are written to JSONL:

```go
{
    name: "writes tool_output and error fields when present",
    setupDir: func(t *testing.T) string {
        t.Helper()
        return t.TempDir()
    },
    events: []observe.Event{
        {
            Timestamp:  fixedTime,
            Phase:      "post",
            ToolName:   "Bash",
            ToolInput:  json.RawMessage(`{"command":"ls"}`),
            ToolOutput: json.RawMessage(`{"stdout":"file.txt"}`),
            SessionID:  "sess-005",
        },
        {
            Timestamp: fixedTime.Add(time.Second),
            Phase:     "failure",
            ToolName:  "Bash",
            ToolInput: json.RawMessage(`{"command":"false"}`),
            Error:     "exit code 1",
            SessionID: "sess-005",
        },
    },
    wantLines: 2,
    wantErr:   false,
},
```

Also add assertions in the verification loop to check `ToolOutput` and `Error`:

```go
// After existing assertions in the verification loop:
if len(tt.events[i].ToolOutput) > 0 {
    assert.JSONEq(t, string(tt.events[i].ToolOutput), string(parsed.ToolOutput))
}
if tt.events[i].Error != "" {
    assert.Equal(t, tt.events[i].Error, parsed.Error)
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestRecord ./internal/observe/...`
Expected: FAIL — `observe.Event` has no `ToolOutput` or `Error` fields

**Step 3: Write minimal implementation**

In `internal/observe/observe.go`, extend the Event struct at line 19-25:

```go
type Event struct {
	Timestamp  time.Time       `json:"timestamp"`
	Phase      string          `json:"phase"` // "pre", "post", or "failure".
	ToolName   string          `json:"tool_name"`
	ToolInput  json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput json.RawMessage `json:"tool_output,omitempty"`
	Error      string          `json:"error,omitempty"`
	SessionID  string          `json:"session_id"`
}
```

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestRecord ./internal/observe/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/observe/observe.go internal/observe/observe_test.go
git commit -m "feat: extend observe.Event with ToolOutput and Error fields"
```

---

## Task 2: Update ObserveHandler to pass ToolOutput and Error

**Files:**
- Modify: `internal/handler/tooluse.go:133-161`
- Test: `internal/handler/tooluse_test.go` (find existing ObserveHandler tests)

**Step 1: Write the failing test**

Add test cases for ObserveHandler that verify ToolOutput is passed for "post" phase and Error for "failure" phase. Create a test that reads back the JSONL and confirms the new fields:

```go
func TestObserveHandler_PostPhaseIncludesToolOutput(t *testing.T) {
    dir := t.TempDir()
    cfg := &config.Values{Observe: config.ObserveValues{Enabled: true, MaxFileSizeMB: 10}}
    h := NewObserveHandler(cfg, "post", WithObserveDir(dir))

    input := &hookcmd.HookInput{
        ToolName:  "Bash",
        ToolInput: json.RawMessage(`{"command":"ls"}`),
        ToolOutput: json.RawMessage(`{"stdout":"output"}`),
        SessionID: "sess-test",
    }

    resp, err := h.Handle(context.Background(), input)
    require.NoError(t, err)
    assert.Equal(t, 0, resp.ExitCode)

    data, err := os.ReadFile(filepath.Join(dir, "observations.jsonl"))
    require.NoError(t, err)

    var event observe.Event
    require.NoError(t, json.Unmarshal(data[:len(data)-1], &event))
    assert.JSONEq(t, `{"stdout":"output"}`, string(event.ToolOutput))
}

func TestObserveHandler_FailurePhaseIncludesError(t *testing.T) {
    dir := t.TempDir()
    cfg := &config.Values{Observe: config.ObserveValues{Enabled: true, MaxFileSizeMB: 10}}
    h := NewObserveHandler(cfg, "failure", WithObserveDir(dir))

    input := &hookcmd.HookInput{
        ToolName:  "Bash",
        ToolInput: json.RawMessage(`{"command":"false"}`),
        Error:     "command failed with exit code 1",
        SessionID: "sess-test",
    }

    resp, err := h.Handle(context.Background(), input)
    require.NoError(t, err)
    assert.Equal(t, 0, resp.ExitCode)

    data, err := os.ReadFile(filepath.Join(dir, "observations.jsonl"))
    require.NoError(t, err)

    var event observe.Event
    require.NoError(t, json.Unmarshal(data[:len(data)-1], &event))
    assert.Equal(t, "command failed with exit code 1", event.Error)
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestObserveHandler_PostPhaseIncludesToolOutput ./internal/handler/...`
Expected: FAIL — ToolOutput/Error not populated in the event

**Step 3: Write minimal implementation**

Update `ObserveHandler.Handle` in `internal/handler/tooluse.go:150-156` to include the new fields:

```go
if err := obs.Record(observe.Event{
    Timestamp:  time.Now(),
    Phase:      h.phase,
    ToolName:   input.ToolName,
    ToolInput:  input.ToolInput,
    ToolOutput: input.ToolOutput,
    Error:      input.Error,
    SessionID:  input.SessionID,
}); err != nil {
```

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestObserveHandler ./internal/handler/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/handler/tooluse.go internal/handler/tooluse_test.go
git commit -m "feat: pass ToolOutput and Error through ObserveHandler"
```

---

## Task 3: Add InstinctValues to config system

**Files:**
- Modify: `internal/config/values.go:6-17`
- Modify: `internal/config/keys.go`
- Modify: `internal/config/manager.go`
- Test: `internal/config/manager_test.go`

**Step 1: Write the failing test**

Add test that the new instinct config keys have correct defaults and can be get/set:

```go
func TestInstinctConfigDefaults(t *testing.T) {
    cfg := GetDefaultConfig()
    assert.Equal(t, 0.3, cfg.Instinct.MinConfidence)
    assert.Equal(t, 0.7, cfg.Instinct.AutoApprove)
    assert.Equal(t, 0.02, cfg.Instinct.DecayRate)
    assert.Equal(t, 100, cfg.Instinct.MaxInstincts)
    assert.Equal(t, 3, cfg.Instinct.ClusterThreshold)
    assert.Contains(t, cfg.Instinct.PersonalPath, "instincts/personal")
    assert.Contains(t, cfg.Instinct.InheritedPath, "instincts/inherited")
}

func TestInstinctConfigSetGet(t *testing.T) {
    tmpFile := filepath.Join(t.TempDir(), "config.json")
    m := NewManagerWithPath(tmpFile)
    ctx := context.Background()

    require.NoError(t, m.Set(ctx, "instinct.min_confidence", "0.5"))
    val, found, err := m.GetValue(ctx, "instinct.min_confidence")
    require.NoError(t, err)
    assert.True(t, found)
    assert.Equal(t, "0.5", val)

    require.NoError(t, m.Set(ctx, "instinct.max_instincts", "50"))
    intVal, found, err := m.GetInt(ctx, "instinct.max_instincts")
    require.NoError(t, err)
    assert.True(t, found)
    assert.Equal(t, 50, intVal)
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestInstinctConfig ./internal/config/...`
Expected: FAIL — no `Instinct` field on `Values`

**Step 3: Write minimal implementation**

Add to `internal/config/values.go` after line 16 (inside Values struct):

```go
Instinct InstinctValues `json:"instinct"`
```

Add the `InstinctValues` type after `StopReminderValues`:

```go
// InstinctValues represents instinct management settings.
type InstinctValues struct {
	PersonalPath     string  `json:"personal_path"`
	InheritedPath    string  `json:"inherited_path"`
	MinConfidence    float64 `json:"min_confidence"`
	AutoApprove      float64 `json:"auto_approve"`
	DecayRate        float64 `json:"decay_rate"`
	MaxInstincts     int     `json:"max_instincts"`
	ClusterThreshold int     `json:"cluster_threshold"`
}
```

Add keys to `internal/config/keys.go`:

```go
keyInstinctPersonalPath     = "instinct.personal_path"
keyInstinctInheritedPath    = "instinct.inherited_path"
keyInstinctMinConfidence    = "instinct.min_confidence"
keyInstinctAutoApprove      = "instinct.auto_approve"
keyInstinctDecayRate        = "instinct.decay_rate"
keyInstinctMaxInstincts     = "instinct.max_instincts"
keyInstinctClusterThreshold = "instinct.cluster_threshold"
```

Add defaults (use `~/.config/cc-tools/instincts/personal` etc.) and wire into:
- `GetDefaultConfig()` — set defaults
- `getDefaultValue()` — return defaults as strings
- `allKeys()` — add all 7 keys
- `GetValue()` / `GetInt()` / `GetString()` — switch cases
- `setField()` — switch cases
- `Reset()` — switch cases
- `ensureDefaults()` — zero-value checks
- `getExtendedValue()` / `setExtendedField()` / `resetExtended()` — add instinct cases
- `convertInstinctFromMap()` — backward compat loader

See the pattern in existing config for `DriftValues` and `StopReminderValues` — follow the same pattern exactly.

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestInstinctConfig ./internal/config/...`
Expected: PASS

**Step 5: Run full config tests**

Run: `gotestsum --format pkgname -- -tags=testmode ./internal/config/...`
Expected: PASS (no regressions)

**Step 6: Commit**

```bash
git add internal/config/values.go internal/config/keys.go internal/config/manager.go internal/config/manager_test.go
git commit -m "feat: add instinct configuration keys to config system"
```

---

## Task 4: Create internal/instinct package — types and parser

**Files:**
- Create: `internal/instinct/instinct.go`
- Create: `internal/instinct/parser.go`
- Create: `internal/instinct/parser_test.go`

**Step 1: Write the failing test**

Create `internal/instinct/parser_test.go`:

```go
package instinct_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riddopic/cc-tools/internal/instinct"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []instinct.Instinct
		wantErr bool
	}{
		{
			name: "single instinct with content",
			input: `---
id: prefer-functional-style
trigger: "when writing new functions"
confidence: 0.75
domain: "code-style"
source: "session-observation"
created_at: "2025-01-22T10:30:00Z"
updated_at: "2025-06-15T14:00:00Z"
---

## Action
Use functional patterns over classes when appropriate.

## Evidence
- Observed 5 instances`,
			want: []instinct.Instinct{
				{
					ID:         "prefer-functional-style",
					Trigger:    "when writing new functions",
					Confidence: 0.75,
					Domain:     "code-style",
					Source:     "session-observation",
					CreatedAt:  time.Date(2025, 1, 22, 10, 30, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2025, 6, 15, 14, 0, 0, 0, time.UTC),
					Content:    "## Action\nUse functional patterns over classes when appropriate.\n\n## Evidence\n- Observed 5 instances",
				},
			},
		},
		{
			name: "instinct without content",
			input: `---
id: test-first
trigger: "when adding features"
confidence: 0.9
domain: "testing"
source: "session-observation"
---`,
			want: []instinct.Instinct{
				{
					ID:         "test-first",
					Trigger:    "when adding features",
					Confidence: 0.9,
					Domain:     "testing",
					Source:     "session-observation",
				},
			},
		},
		{
			name:    "empty input",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "no frontmatter delimiters",
			input:   "just plain text",
			want:    nil,
			wantErr: false,
		},
		{
			name: "missing id is skipped",
			input: `---
trigger: "when writing"
confidence: 0.5
domain: "code-style"
---`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instinct.ParseFrontmatter(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, len(tt.want))
			for i := range got {
				assert.Equal(t, tt.want[i].ID, got[i].ID)
				assert.Equal(t, tt.want[i].Trigger, got[i].Trigger)
				assert.InDelta(t, tt.want[i].Confidence, got[i].Confidence, 0.001)
				assert.Equal(t, tt.want[i].Domain, got[i].Domain)
				assert.Equal(t, tt.want[i].Source, got[i].Source)
				if tt.want[i].Content != "" {
					assert.Equal(t, tt.want[i].Content, got[i].Content)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestParseFrontmatter ./internal/instinct/...`
Expected: FAIL — package doesn't exist

**Step 3: Write minimal implementation**

Create `internal/instinct/instinct.go`:

```go
// Package instinct manages atomic learned behaviors with confidence scoring.
package instinct

import "time"

// Instinct represents a single learned behavior.
type Instinct struct {
	ID         string    `json:"id" yaml:"id"`
	Trigger    string    `json:"trigger" yaml:"trigger"`
	Confidence float64   `json:"confidence" yaml:"confidence"`
	Domain     string    `json:"domain" yaml:"domain"`
	Source     string    `json:"source" yaml:"source"`
	SourceRepo string    `json:"source_repo,omitempty" yaml:"source_repo,omitempty"`
	Content    string    `json:"content,omitempty" yaml:"-"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" yaml:"updated_at"`
}

// ListOptions filters instinct listing.
type ListOptions struct {
	Domain        string
	MinConfidence float64
	Source        string
}
```

Create `internal/instinct/parser.go` with a simple state-machine parser for YAML frontmatter (no external YAML dependency — use key-value line parsing like the Python version does):

```go
package instinct

import (
	"strconv"
	"strings"
	"time"
)

// ParseFrontmatter parses one or more instincts from YAML frontmatter format.
// Each instinct is delimited by --- markers. Content after the closing ---
// is preserved as the Content field. Instincts without an id are skipped.
func ParseFrontmatter(input string) ([]Instinct, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	var result []Instinct
	// Split on "---" boundaries
	// Implementation: scan lines, track state (outside/frontmatter/content)
	// Parse key: value pairs from frontmatter section
	// Collect content lines after closing ---

	lines := strings.Split(input, "\n")
	var current *Instinct
	var contentLines []string
	inFrontmatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFrontmatter && current == nil {
				// Opening delimiter
				current = &Instinct{}
				inFrontmatter = true
				continue
			}
			if inFrontmatter {
				// Closing delimiter
				inFrontmatter = false
				contentLines = nil
				continue
			}
		}

		if inFrontmatter && current != nil {
			parseFrontmatterLine(current, trimmed)
			continue
		}

		if current != nil && !inFrontmatter {
			contentLines = append(contentLines, line)
		}
	}

	if current != nil && current.ID != "" {
		current.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
		result = append(result, *current)
	}

	return result, nil
}

func parseFrontmatterLine(inst *Instinct, line string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	value = strings.Trim(value, "\"")

	switch key {
	case "id":
		inst.ID = value
	case "trigger":
		inst.Trigger = value
	case "confidence":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			inst.Confidence = f
		}
	case "domain":
		inst.Domain = value
	case "source":
		inst.Source = value
	case "source_repo":
		inst.SourceRepo = value
	case "created_at":
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			inst.CreatedAt = t
		}
	case "updated_at":
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			inst.UpdatedAt = t
		}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestParseFrontmatter ./internal/instinct/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/instinct/instinct.go internal/instinct/parser.go internal/instinct/parser_test.go
git commit -m "feat: add instinct types and YAML frontmatter parser"
```

---

## Task 5: Create instinct Store (file-based)

**Files:**
- Create: `internal/instinct/store.go`
- Create: `internal/instinct/store_test.go`

**Step 1: Write the failing test**

Create `internal/instinct/store_test.go` with table-driven tests for Save, Get, List, Delete:

```go
func TestFileStore(t *testing.T) {
	t.Run("Save and Get", func(t *testing.T) {
		dir := t.TempDir()
		store := instinct.NewFileStore(dir, filepath.Join(dir, "inherited"))

		inst := instinct.Instinct{
			ID:         "test-first",
			Trigger:    "when adding features",
			Confidence: 0.8,
			Domain:     "testing",
			Source:     "session-observation",
			Content:    "## Action\nWrite tests first.",
		}
		require.NoError(t, store.Save(inst))

		got, err := store.Get("test-first")
		require.NoError(t, err)
		assert.Equal(t, "test-first", got.ID)
		assert.InDelta(t, 0.8, got.Confidence, 0.001)
		assert.Equal(t, "## Action\nWrite tests first.", got.Content)
	})

	t.Run("List with filters", func(t *testing.T) {
		dir := t.TempDir()
		store := instinct.NewFileStore(dir, filepath.Join(dir, "inherited"))

		for _, inst := range []instinct.Instinct{
			{ID: "a", Domain: "testing", Confidence: 0.8, Source: "session-observation"},
			{ID: "b", Domain: "code-style", Confidence: 0.5, Source: "session-observation"},
			{ID: "c", Domain: "testing", Confidence: 0.3, Source: "session-observation"},
		} {
			require.NoError(t, store.Save(inst))
		}

		all, err := store.List(instinct.ListOptions{})
		require.NoError(t, err)
		assert.Len(t, all, 3)

		testing_, err := store.List(instinct.ListOptions{Domain: "testing"})
		require.NoError(t, err)
		assert.Len(t, testing_, 2)

		highConf, err := store.List(instinct.ListOptions{MinConfidence: 0.6})
		require.NoError(t, err)
		assert.Len(t, highConf, 1)
	})

	t.Run("Delete", func(t *testing.T) {
		dir := t.TempDir()
		store := instinct.NewFileStore(dir, filepath.Join(dir, "inherited"))

		require.NoError(t, store.Save(instinct.Instinct{ID: "delete-me", Domain: "test", Confidence: 0.5}))
		require.NoError(t, store.Delete("delete-me"))

		_, err := store.Get("delete-me")
		assert.Error(t, err)
	})

	t.Run("Get nonexistent returns error", func(t *testing.T) {
		dir := t.TempDir()
		store := instinct.NewFileStore(dir, filepath.Join(dir, "inherited"))
		_, err := store.Get("nope")
		assert.Error(t, err)
	})
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestFileStore ./internal/instinct/...`
Expected: FAIL — `NewFileStore` doesn't exist

**Step 3: Write minimal implementation**

Create `internal/instinct/store.go` implementing `FileStore` that:
- `Save()` — serialize to YAML frontmatter format, write to `{personalDir}/{id}.yaml`
- `Get()` — read file, parse frontmatter, return
- `List()` — glob `*.yaml` in both personal and inherited dirs, parse all, apply filters
- `Delete()` — remove file by ID
- Uses `filepath.Clean` for path safety
- Creates directories on first save

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestFileStore ./internal/instinct/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/instinct/store.go internal/instinct/store_test.go
git commit -m "feat: add file-based instinct store with list/get/save/delete"
```

---

## Task 6: Create instinct confidence scoring

**Files:**
- Create: `internal/instinct/confidence.go`
- Create: `internal/instinct/confidence_test.go`

**Step 1: Write the failing test**

```go
func TestConfidenceFromObservations(t *testing.T) {
	tests := []struct {
		name         string
		observations int
		want         float64
	}{
		{"1 observation", 1, 0.3},
		{"2 observations", 2, 0.3},
		{"3 observations", 3, 0.5},
		{"5 observations", 5, 0.5},
		{"6 observations", 6, 0.7},
		{"10 observations", 10, 0.7},
		{"11 observations", 11, 0.85},
		{"20 observations", 20, 0.85},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instinct.ConfidenceFromObservations(tt.observations)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestClampConfidence(t *testing.T) {
	assert.InDelta(t, 0.3, instinct.ClampConfidence(0.1), 0.001)
	assert.InDelta(t, 0.5, instinct.ClampConfidence(0.5), 0.001)
	assert.InDelta(t, 0.9, instinct.ClampConfidence(1.0), 0.001)
}

func TestDecayConfidence(t *testing.T) {
	got := instinct.DecayConfidence(0.7, 2, 0.02) // 2 weeks, 0.02 rate
	assert.InDelta(t, 0.66, got, 0.001)

	// Decayed below min clamps to min
	got = instinct.DecayConfidence(0.32, 5, 0.02)
	assert.InDelta(t, 0.3, got, 0.001)
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestConfidence ./internal/instinct/...`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `internal/instinct/confidence.go`:

```go
package instinct

const (
	minConfidence = 0.3
	maxConfidence = 0.9
)

// ConfidenceFromObservations returns a base confidence for a given observation count.
func ConfidenceFromObservations(count int) float64 {
	switch {
	case count >= 11:
		return 0.85
	case count >= 6:
		return 0.7
	case count >= 3:
		return 0.5
	default:
		return 0.3
	}
}

// ClampConfidence restricts a confidence value to [0.3, 0.9].
func ClampConfidence(c float64) float64 {
	if c < minConfidence {
		return minConfidence
	}
	if c > maxConfidence {
		return maxConfidence
	}
	return c
}

// DecayConfidence reduces confidence by rate per week for the given number of weeks.
func DecayConfidence(confidence float64, weeks int, rate float64) float64 {
	result := confidence - (float64(weeks) * rate)
	return ClampConfidence(result)
}
```

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestConfidence ./internal/instinct/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/instinct/confidence.go internal/instinct/confidence_test.go
git commit -m "feat: add confidence scoring and decay for instincts"
```

---

## Task 7: Create instinct clustering for evolve

**Files:**
- Create: `internal/instinct/cluster.go`
- Create: `internal/instinct/cluster_test.go`

**Step 1: Write the failing test**

```go
func TestClusterInstincts(t *testing.T) {
	instincts := []instinct.Instinct{
		{ID: "a", Trigger: "when writing functions", Domain: "code-style", Confidence: 0.8},
		{ID: "b", Trigger: "when writing tests", Domain: "testing", Confidence: 0.7},
		{ID: "c", Trigger: "when writing code", Domain: "code-style", Confidence: 0.6},
		{ID: "d", Trigger: "when writing modules", Domain: "code-style", Confidence: 0.75},
		{ID: "e", Trigger: "when debugging errors", Domain: "debugging", Confidence: 0.5},
	}

	clusters := instinct.ClusterByTrigger(instincts, 3)
	// "writing" appears in a, b, c, d — at least 3 share the normalized trigger word
	assert.GreaterOrEqual(t, len(clusters), 1)

	// Verify cluster has 3+ members
	for _, c := range clusters {
		assert.GreaterOrEqual(t, len(c.Members), 3)
	}
}

func TestNormalizeTrigger(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"when writing new functions", []string{"functions"}},
		{"when creating a module", []string{"module"}},
		{"when debugging errors in tests", []string{"debugging", "errors", "tests"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := instinct.NormalizeTrigger(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestCluster ./internal/instinct/...`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `internal/instinct/cluster.go` with:
- `NormalizeTrigger(s string) []string` — lowercase, split on non-alpha, remove stop words (when, writing, creating, new, a, the, etc.)
- `ClusterByTrigger(instincts []Instinct, minSize int) []Cluster` — group by shared normalized words, filter by minSize
- `Cluster` struct with `Members []Instinct`, `Keywords []string`, `AvgConfidence float64`

**Step 4: Run test to verify it passes**

Run: `gotestsum --format pkgname -- -tags=testmode -run TestCluster ./internal/instinct/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/instinct/cluster.go internal/instinct/cluster_test.go
git commit -m "feat: add instinct clustering by trigger keywords"
```

---

## Task 8: Create instinct export and import

**Files:**
- Create: `internal/instinct/export.go`
- Create: `internal/instinct/export_test.go`

**Step 1: Write the failing test**

Test YAML export format and JSON export format:

```go
func TestExportYAML(t *testing.T) {
	instincts := []instinct.Instinct{
		{ID: "test-first", Trigger: "when adding features", Confidence: 0.8, Domain: "testing"},
	}
	var buf bytes.Buffer
	require.NoError(t, instinct.ExportYAML(&buf, instincts))

	output := buf.String()
	assert.Contains(t, output, "id: test-first")
	assert.Contains(t, output, "confidence: 0.8")
}

func TestExportJSON(t *testing.T) {
	instincts := []instinct.Instinct{
		{ID: "test-first", Trigger: "when adding features", Confidence: 0.8, Domain: "testing"},
	}
	var buf bytes.Buffer
	require.NoError(t, instinct.ExportJSON(&buf, instincts))

	var parsed []instinct.Instinct
	require.NoError(t, json.Unmarshal(buf.Bytes(), &parsed))
	assert.Len(t, parsed, 1)
	assert.Equal(t, "test-first", parsed[0].ID)
}
```

**Step 2: Run, verify fails, implement, verify passes**

Implement `ExportYAML(w io.Writer, instincts []Instinct) error` and `ExportJSON(w io.Writer, instincts []Instinct) error`.

**Step 3: Commit**

```bash
git add internal/instinct/export.go internal/instinct/export_test.go
git commit -m "feat: add YAML and JSON instinct export"
```

---

## Task 9: Create cc-tools instinct command group

**Files:**
- Create: `cmd/cc-tools/instinct.go`
- Modify: `cmd/cc-tools/main.go:41` (add `newInstinctCmd()` to AddCommand)

**Step 1: Write the command scaffolding**

Create `cmd/cc-tools/instinct.go` with the Cobra command group and four subcommands. Follow the pattern from `cmd/cc-tools/session.go`:

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/instinct"
)

func newInstinctCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instinct",
		Short: "Manage learned instincts",
	}
	cmd.AddCommand(
		newInstinctStatusCmd(),
		newInstinctExportCmd(),
		newInstinctImportCmd(),
		newInstinctEvolveCmd(),
	)
	return cmd
}
```

Then implement each subcommand:
- `newInstinctStatusCmd()` — loads store, lists instincts, groups by domain, prints with confidence bars
- `newInstinctExportCmd()` — loads store, filters, exports to stdout or file
- `newInstinctImportCmd()` — reads file arg, parses, deduplicates, saves to inherited dir
- `newInstinctEvolveCmd()` — loads store, clusters, prints candidates

Register in `cmd/cc-tools/main.go` by adding `newInstinctCmd()` to the AddCommand call at line 41.

**Step 2: Build and verify**

Run: `task build`
Expected: Binary builds without errors

Run: `bin/cc-tools instinct --help`
Expected: Shows subcommands: status, export, import, evolve

Run: `bin/cc-tools instinct status`
Expected: "No instincts found." (empty store)

**Step 3: Commit**

```bash
git add cmd/cc-tools/instinct.go cmd/cc-tools/main.go
git commit -m "feat: add cc-tools instinct command group with status/export/import/evolve"
```

---

## Task 10: Remove observe.sh hooks from settings.json

**Files:**
- Modify: `.claude/settings.json:31-68`

**Step 1: Edit settings.json**

Remove the observe.sh hook entries from PreToolUse and PostToolUse arrays.

**Before** (PreToolUse, lines 31-45):
```json
"PreToolUse": [
  {
    "matcher": "*",
    "hooks": [
      { "type": "command", "command": "cc-tools hook" },
      { "type": "command", "command": ".claude/skills/continuous-learning-v2/hooks/observe.sh pre" }
    ]
  }
]
```

**After**:
```json
"PreToolUse": [
  {
    "matcher": "*",
    "hooks": [
      { "type": "command", "command": "cc-tools hook" }
    ]
  }
]
```

Same for PostToolUse — remove the observe.sh post entry from the `"matcher": "*"` hooks array.

**Step 2: Verify cc-tools hook still works**

Run: `echo '{"hook_event_name":"PreToolUse","tool_name":"Bash","session_id":"test"}' | cc-tools hook`
Expected: exits 0, ObserveHandler writes to `~/.cache/cc-tools/observations/`

**Step 3: Commit**

```bash
git add .claude/settings.json
git commit -m "refactor: remove redundant observe.sh hooks from settings.json"
```

---

## Task 11: Update .claude/commands/ to use cc-tools instinct

**Files:**
- Modify: `.claude/commands/instinct-status.md`
- Modify: `.claude/commands/instinct-export.md`
- Modify: `.claude/commands/instinct-import.md`
- Modify: `.claude/commands/evolve.md`
- Modify: `.claude/commands/learn.md`
- Modify: `.claude/commands/learn-eval.md`

**Step 1: Update instinct-status.md**

Replace the Python implementation section with:
```markdown
## Implementation

Run the instinct CLI:

```bash
cc-tools instinct status
```
```

Update paths from `.claude/homunculus/instincts/` to `~/.config/cc-tools/instincts/`.

Remove the `skills:` frontmatter entry for `continuous-learning-v2` — no longer needed since cc-tools handles it natively.

**Step 2: Update instinct-export.md**

Same pattern — replace `python3 .claude/skills/continuous-learning-v2/scripts/instinct-cli.py export` with `cc-tools instinct export`. Update paths.

**Step 3: Update instinct-import.md**

Replace `python3 ... import` with `cc-tools instinct import`. Update paths from `.claude/homunculus/instincts/inherited/` to `~/.config/cc-tools/instincts/inherited/`.

**Step 4: Update evolve.md**

Replace `python3 ... evolve` with `cc-tools instinct evolve`. Update paths.

**Step 5: Update learn.md**

Update the output path from `.claude/skills/learned/` to `~/.config/cc-tools/instincts/personal/` where instinct files are saved. The `/learn` command itself is Claude-driven — no Python to replace.

**Step 6: Update learn-eval.md**

Update Global path from `~/.claude/skills/learned/` to `~/.config/cc-tools/instincts/personal/`. The Project path stays at `.claude/skills/learned/`.

**Step 7: Commit**

```bash
git add .claude/commands/instinct-status.md .claude/commands/instinct-export.md .claude/commands/instinct-import.md .claude/commands/evolve.md .claude/commands/learn.md .claude/commands/learn-eval.md
git commit -m "refactor: update instinct commands to use cc-tools instinct"
```

---

## Task 12: Remove obsolete Python/Bash files

**Files:**
- Delete: `.claude/skills/continuous-learning-v2/hooks/observe.sh`
- Delete: `.claude/skills/continuous-learning-v2/scripts/instinct-cli.py`
- Delete: `.claude/skills/continuous-learning-v2/scripts/test_parse_instinct.py`
- Delete: `.claude/skills/continuous-learning-v2/config.json`
- Delete: `.claude/skills/continuous-learning-v2/agents/start-observer.sh`
- Delete: `.claude/skills/continuous-learning-v2/agents/observer.md`

**Step 1: Remove the files**

```bash
git rm .claude/skills/continuous-learning-v2/hooks/observe.sh
git rm .claude/skills/continuous-learning-v2/scripts/instinct-cli.py
git rm .claude/skills/continuous-learning-v2/scripts/test_parse_instinct.py
git rm .claude/skills/continuous-learning-v2/config.json
git rm .claude/skills/continuous-learning-v2/agents/start-observer.sh
git rm .claude/skills/continuous-learning-v2/agents/observer.md
```

**Step 2: Remove empty directories if any remain**

```bash
rmdir .claude/skills/continuous-learning-v2/hooks/ 2>/dev/null || true
rmdir .claude/skills/continuous-learning-v2/scripts/ 2>/dev/null || true
rmdir .claude/skills/continuous-learning-v2/agents/ 2>/dev/null || true
```

**Step 3: Commit**

```bash
git commit -m "chore: remove obsolete Python/Bash instinct scripts"
```

---

## Task 13: Rewrite continuous-learning-v2 SKILL.md

**Files:**
- Modify: `.claude/skills/continuous-learning-v2/SKILL.md`

**Step 1: Rewrite SKILL.md**

Rewrite to document the cc-tools-native instinct system. Key changes:
- Reference `cc-tools instinct` commands instead of Python scripts
- Reference `~/.config/cc-tools/instincts/` instead of `.claude/homunculus/`
- Reference `~/.cache/cc-tools/observations/` instead of `.claude/homunculus/observations.jsonl`
- Remove observer agent references
- Update the hook explanation to note that observation is built into cc-tools

**Step 2: Commit**

```bash
git add .claude/skills/continuous-learning-v2/SKILL.md
git commit -m "docs: rewrite continuous-learning-v2 SKILL.md for cc-tools integration"
```

---

## Task 14: Update rules and agents

**Files:**
- Modify: `.claude/rules/hooks.md`
- Modify: `.claude/rules/agents.md`

**Step 1: Update hooks.md**

Remove the `observe.sh` hook configuration example. Update the description of UserPromptSubmit and Stop handlers to reflect current state. Remove any references to `.claude/homunculus/`.

**Step 2: Update agents.md**

Update the `continuous-learning-v2` entry in the Analysis & Patterns table to reflect that it now uses cc-tools natively.

**Step 3: Commit**

```bash
git add .claude/rules/hooks.md .claude/rules/agents.md
git commit -m "docs: update rules to reflect instinct migration to cc-tools"
```

---

## Task 15: Run full pre-commit checks

**Step 1: Format code**

Run: `task fmt`
Expected: No changes or minor formatting fixes

**Step 2: Run linter**

Run: `task lint`
Expected: PASS

**Step 3: Run tests**

Run: `task test`
Expected: PASS

**Step 4: Run race detector**

Run: `task test-race`
Expected: PASS

**Step 5: Build**

Run: `task build`
Expected: Binary builds, `bin/cc-tools instinct status` works

**Step 6: Final commit if fmt made changes**

```bash
git add -A
git commit -m "chore: fmt and lint fixes for instinct migration"
```

---

Plan complete and saved to `docs/plans/2026-02-24-instinct-migration-plan.md`. Two execution options:

**1. Subagent-Driven (this session)** — I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** — Open new session with executing-plans, batch execution with checkpoints

Which approach?
