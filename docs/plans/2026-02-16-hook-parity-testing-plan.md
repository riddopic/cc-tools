# Hook Parity Testing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Add 31 parity test cases across 5 test files to verify that all 10 cc-tools handlers behave equivalently to the original JS/shell hook scripts.

**Architecture:** Per-handler table-driven Go tests in `internal/handler/`. Each test constructs `HookInput` structs matching Claude Code's JSON protocol, calls `Handle()` directly, and asserts the full I/O contract (exit code, stdout JSON, stderr, file side effects).

**Tech Stack:** Go 1.26, testify (assert/require), existing `newTestConfig()` helper, `t.TempDir()` for file isolation, functional options for dependency injection.

**Important context:** These tests verify existing handler implementations. Tests should pass on first run. If any test fails, it reveals a parity gap in the handler that needs fixing.

---

### Task 1: LogCompaction parity tests

**Files:**
- Modify: `internal/handler/compact_test.go`

**Step 1: Add parity test cases**

Add these tests after the existing `TestLogCompactionHandler_Handle` test:

```go
func TestLogCompactionHandler_AppendsOnMultipleCalls(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	h := handler.NewLogCompactionHandler(handler.WithCompactLogDir(tmpDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreCompact,
	}

	// Call three times.
	for range 3 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 0, resp.ExitCode)
	}

	logFile := filepath.Join(tmpDir, "compaction-log.txt")
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 3, "each call should append one line")

	for i, line := range lines {
		assert.Contains(t, line, "compaction triggered",
			"line %d should contain 'compaction triggered'", i)
	}
}

func TestLogCompactionHandler_EntryFormat(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	h := handler.NewLogCompactionHandler(handler.WithCompactLogDir(tmpDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreCompact,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	logFile := filepath.Join(tmpDir, "compaction-log.txt")
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)

	line := strings.TrimSpace(string(data))
	// Format: [YYYY-MM-DD HH:MM:SS] compaction triggered
	assert.Regexp(t, `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] compaction triggered$`,
		line, "entry should match timestamp format")
}
```

Note: `compact_test.go` needs `"strings"` added to the import block.

**Step 2: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestLogCompactionHandler" ./internal/handler/...
```

Expected: All tests PASS (including existing ones).

**Step 3: Commit**

```bash
git add internal/handler/compact_test.go
git commit -m "test: add LogCompaction parity tests for append and format"
```

---

### Task 2: SuggestCompact parity tests

**Files:**
- Modify: `internal/handler/tooluse_test.go`

**Step 1: Add parity test cases**

Add these tests after the existing `TestSuggestCompactHandler_SuggestsAtThreshold` test:

```go
func TestSuggestCompactHandler_BelowThreshold(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 5
	cfg.Compact.ReminderInterval = 10

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "below-threshold",
	}

	// Make threshold-1 calls.
	for range 4 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 0, resp.ExitCode)
		assert.Empty(t, resp.Stderr, "no suggestion below threshold")
	}
}

func TestSuggestCompactHandler_ReminderInterval(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 3
	cfg.Compact.ReminderInterval = 2

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "interval-session",
	}

	// Call 5 times: threshold=3, interval=2, so suggestions at 3 and 5.
	var stderrResults []string
	for range 5 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		stderrResults = append(stderrResults, resp.Stderr)
	}

	assert.Empty(t, stderrResults[0], "call 1: no suggestion")
	assert.Empty(t, stderrResults[1], "call 2: no suggestion")
	assert.NotEmpty(t, stderrResults[2], "call 3: threshold suggestion")
	assert.Empty(t, stderrResults[3], "call 4: between intervals")
	assert.NotEmpty(t, stderrResults[4], "call 5: interval suggestion")
}

func TestSuggestCompactHandler_SeparateSessions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 3
	cfg.Compact.ReminderInterval = 10

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))

	inputA := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "session-a",
	}
	inputB := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "session-b",
	}

	// Make 2 calls on session A — below threshold.
	for range 2 {
		resp, err := h.Handle(context.Background(), inputA)
		require.NoError(t, err)
		assert.Empty(t, resp.Stderr)
	}

	// Make 3 calls on session B — should hit threshold independently.
	var lastResp *handler.Response
	for range 3 {
		resp, err := h.Handle(context.Background(), inputB)
		require.NoError(t, err)
		lastResp = resp
	}

	assert.NotEmpty(t, lastResp.Stderr, "session B should hit threshold at 3")

	// Session A counter should still be at 2.
	counterA := filepath.Join(stateDir, "cc-tools-compact-session-a.count")
	data, err := os.ReadFile(counterA)
	require.NoError(t, err)
	assert.Equal(t, "2", strings.TrimSpace(string(data)))
}

func TestSuggestCompactHandler_CounterFileIncrement(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 100 // High threshold so no suggestion noise.
	cfg.Compact.ReminderInterval = 100

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "increment-test",
	}

	for range 7 {
		_, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
	}

	counterFile := filepath.Join(stateDir, "cc-tools-compact-increment-test.count")
	data, err := os.ReadFile(counterFile)
	require.NoError(t, err)
	assert.Equal(t, "7", strings.TrimSpace(string(data)),
		"counter should equal number of Handle calls")
}

func TestSuggestCompactHandler_ZeroThreshold(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 0
	cfg.Compact.ReminderInterval = 0

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "zero-threshold",
	}

	for range 10 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		assert.Empty(t, resp.Stderr, "zero threshold should never suggest")
	}
}

func TestSuggestCompactHandler_SuggestionMessage(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "compact")

	cfg := newTestConfig()
	cfg.Compact.Threshold = 2
	cfg.Compact.ReminderInterval = 5

	h := handler.NewSuggestCompactHandler(cfg, handler.WithCompactStateDir(stateDir))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		SessionID:     "message-test",
	}

	// Reach threshold.
	var lastResp *handler.Response
	for range 2 {
		resp, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
		lastResp = resp
	}

	assert.Contains(t, lastResp.Stderr, "2 tool calls",
		"should include call count")
	assert.Contains(t, lastResp.Stderr, "/compact",
		"should mention /compact command")
}
```

Note: `tooluse_test.go` needs `"strings"` added to the import block.

**Step 2: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestSuggestCompactHandler" ./internal/handler/...
```

Expected: All tests PASS.

**Step 3: Commit**

```bash
git add internal/handler/tooluse_test.go
git commit -m "test: add SuggestCompact parity tests for intervals and counters"
```

---

### Task 3: PreCommitReminder parity tests

**Files:**
- Modify: `internal/handler/tooluse_test.go`

**Step 1: Add parity test cases**

Add these tests after the existing `TestPreCommitReminderHandler_NoGitCommit` test:

```go
func TestPreCommitReminderHandler_GitCommitAmFlag(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -am 'quick fix'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit",
		"should detect git commit with -am flag")
}

func TestPreCommitReminderHandler_ChainedGitCommit(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "task pre-commit"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{
		"command": "git add . && git commit -m 'fix: resolve race'",
	})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "task pre-commit",
		"should detect git commit in chained command")
}

func TestPreCommitReminderHandler_CustomCommand(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig()
	cfg.PreCommit.Enabled = true
	cfg.PreCommit.Command = "make check"

	h := handler.NewPreCommitReminderHandler(cfg)

	toolInput, _ := json.Marshal(map[string]string{"command": "git commit -m 'test'"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Stderr, "make check",
		"should use configured custom command")
	assert.NotContains(t, resp.Stderr, "task pre-commit",
		"should not use default command")
}
```

**Step 2: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestPreCommitReminderHandler" ./internal/handler/...
```

Expected: All tests PASS.

**Step 3: Commit**

```bash
git add internal/handler/tooluse_test.go
git commit -m "test: add PreCommitReminder parity tests for flag variants and custom command"
```

---

### Task 4: Observe parity tests

**Files:**
- Modify: `internal/handler/tooluse_test.go`

**Step 1: Add parity test cases**

Add these tests after the existing `TestObserveHandler_RecordsEvent` test:

```go
func TestObserveHandler_PostPhase(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "post", handler.WithObserveDir(obsDir))

	toolInput, _ := json.Marshal(map[string]string{"command": "ls"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPostToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
		SessionID:     "post-phase-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), `"phase":"post"`)
}

func TestObserveHandler_FailurePhase(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "failure", handler.WithObserveDir(obsDir))

	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPostToolUseFailure,
		ToolName:      "Bash",
		SessionID:     "failure-phase-session",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), `"phase":"failure"`)
}

func TestObserveHandler_MultipleEventsAppend(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	tools := []string{"Bash", "Read", "Edit"}
	for _, tool := range tools {
		toolInput, _ := json.Marshal(map[string]string{"file_path": "/tmp/" + tool})
		input := &hookcmd.HookInput{
			HookEventName: hookcmd.EventPreToolUse,
			ToolName:      tool,
			ToolInput:     toolInput,
			SessionID:     "multi-append",
		}

		_, err := h.Handle(context.Background(), input)
		require.NoError(t, err)
	}

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 3, "should have one JSONL line per event")

	// Each line should be valid JSON.
	for i, line := range lines {
		assert.True(t, json.Valid([]byte(line)),
			"line %d should be valid JSON", i)
	}
}

func TestObserveHandler_DisabledMarkerFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	// Create the observations dir and .disabled marker file.
	require.NoError(t, os.MkdirAll(obsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(obsDir, ".disabled"), []byte(""), 0o600))

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	toolInput, _ := json.Marshal(map[string]string{"command": "ls"})
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Bash",
		ToolInput:     toolInput,
		SessionID:     "disabled-marker",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// No observations file should be created.
	obsFile := filepath.Join(obsDir, "observations.jsonl")
	_, statErr := os.Stat(obsFile)
	assert.True(t, os.IsNotExist(statErr),
		"observations file should not exist when disabled via marker")
}

func TestObserveHandler_EmptyToolInput(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	obsDir := filepath.Join(tmpDir, "observations")

	cfg := newTestConfig()
	cfg.Observe.Enabled = true
	cfg.Observe.MaxFileSizeMB = 10

	h := handler.NewObserveHandler(cfg, "pre", handler.WithObserveDir(obsDir))

	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventPreToolUse,
		ToolName:      "Read",
		ToolInput:     nil, // No tool input.
		SessionID:     "empty-input",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	obsFile := filepath.Join(obsDir, "observations.jsonl")
	data, readErr := os.ReadFile(obsFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "Read",
		"should record event even with nil tool input")
}
```

Note: `tooluse_test.go` needs `"strings"` added to the import block (if not already added in Task 2).

**Step 2: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestObserveHandler" ./internal/handler/...
```

Expected: All tests PASS.

**Step 3: Commit**

```bash
git add internal/handler/tooluse_test.go
git commit -m "test: add Observe parity tests for phases, append, and disabled marker"
```

---

### Task 5: SessionStart parity tests (Superpowers + PkgManager + SessionContext)

**Files:**
- Modify: `internal/handler/session_start_test.go`

**Step 1: Add Superpowers parity tests**

Add after `TestSuperpowersHandler_Handle_WithSkillFile`:

```go
func TestSuperpowersHandler_Handle_MultipleSkills(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create two skill directories.
	for _, name := range []string{"skill-a", "skill-b"} {
		skillDir := filepath.Join(tmpDir, ".claude", "skills", name)
		require.NoError(t, os.MkdirAll(skillDir, 0o755))
		require.NoError(t, os.WriteFile(
			filepath.Join(skillDir, "SKILL.md"),
			[]byte("Skill "+name+" content."),
			0o600,
		))
	}

	h := handler.NewSuperpowersHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	require.NotNil(t, resp.Stdout)
	require.NotNil(t, resp.Stdout.HookSpecificOutput)
}
```

**Step 2: Add PkgManager parity tests**

Add after `TestPkgManagerHandler_Handle_DetectsYarn`:

```go
func TestPkgManagerHandler_Handle_DetectsNpm(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0o600,
	))

	h := handler.NewPkgManagerHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, ".claude", ".env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=npm")
}

func TestPkgManagerHandler_Handle_DetectsPnpm(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(""), 0o600,
	))

	h := handler.NewPkgManagerHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, ".claude", ".env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=pnpm")
}

func TestPkgManagerHandler_Handle_DetectsBun(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "bun.lock"), []byte(""), 0o600,
	))

	h := handler.NewPkgManagerHandler()
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
		Cwd:           tmpDir,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)

	envFile := filepath.Join(tmpDir, ".claude", ".env")
	data, readErr := os.ReadFile(envFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "PREFERRED_PACKAGE_MANAGER=bun")
}
```

**Step 3: Add SessionContext parity test**

Add after `TestSessionContextHandler_Handle_WithAliases`:

```go
func TestSessionContextHandler_Handle_MultipleSessionsUsesRecent(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	storeDir := filepath.Join(tmpHome, ".claude", "sessions")
	store := session.NewStore(storeDir)

	// Save two sessions — older first, then newer.
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "older-session",
		Date:          "2025-01-10",
		Started:       time.Date(2025, 1, 10, 9, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "Older session",
		Summary:       "Old work done here",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))
	require.NoError(t, store.Save(&session.Session{
		Version:       "1",
		ID:            "newer-session",
		Date:          "2025-01-15",
		Started:       time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC),
		Ended:         time.Time{},
		Title:         "Newer session",
		Summary:       "Recent work done here",
		ToolsUsed:     nil,
		FilesModified: nil,
		MessageCount:  0,
	}))

	h := handler.NewSessionContextHandler(handler.WithHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventSessionStart,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Stdout)
	require.NotEmpty(t, resp.Stdout.AdditionalContext)
	assert.Contains(t, resp.Stdout.AdditionalContext[0], "Recent work done here",
		"should use most recent session's summary")
}
```

**Step 4: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestSuperpowersHandler|TestPkgManagerHandler|TestSessionContextHandler" ./internal/handler/...
```

Expected: All tests PASS.

**Step 5: Commit**

```bash
git add internal/handler/session_start_test.go
git commit -m "test: add SessionStart parity tests for multi-skill, pkg detection, session recency"
```

---

### Task 6: SessionEnd parity tests

**Files:**
- Modify: `internal/handler/session_end_test.go`

**Step 1: Add parity test cases**

Add after `TestSessionEndHandler_DefaultMinSessionLength`:

```go
func TestSessionEndHandler_ShortSessionNoSignal(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	// Create a transcript with fewer messages than threshold.
	transcriptDir := t.TempDir()
	transcriptPath := filepath.Join(transcriptDir, "transcript.jsonl")

	var b strings.Builder
	for range 3 {
		b.WriteString("{\"type\":\"human\",\"content\":\"short msg\"}\n")
	}
	require.NoError(t, os.WriteFile(transcriptPath, []byte(b.String()), 0o600))

	cfg := &config.Values{
		Learning: config.LearningValues{
			MinSessionLength: 10,
		},
	}

	h := handler.NewSessionEndHandler(cfg, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName:  hookcmd.EventSessionEnd,
		SessionID:      "short-session",
		TranscriptPath: transcriptPath,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotContains(t, resp.Stderr, "evaluate for extractable patterns",
		"short session should not emit learning signal")
}

func TestSessionEndHandler_SessionMetadata(t *testing.T) {
	t.Parallel()
	tmpHome := t.TempDir()

	transcriptDir := t.TempDir()
	transcriptPath := filepath.Join(transcriptDir, "transcript.jsonl")
	content := strings.Join([]string{
		`{"type":"human","content":"hello"}`,
		`{"type":"tool_use","name":"Edit","input":{"file_path":"/tmp/foo.go"}}`,
		`{"type":"tool_use","name":"Bash","input":{"command":"go test"}}`,
		`{"type":"human","content":"thanks"}`,
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(content), 0o600))

	cfg := &config.Values{}

	h := handler.NewSessionEndHandler(cfg, handler.WithSessionEndHomeDir(tmpHome))
	input := &hookcmd.HookInput{
		HookEventName:  hookcmd.EventSessionEnd,
		SessionID:      "metadata-session",
		TranscriptPath: transcriptPath,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)

	// Verify session file was saved.
	sessDir := filepath.Join(tmpHome, ".claude", "sessions")
	matches, _ := filepath.Glob(filepath.Join(sessDir, "*metadata-session.json"))
	require.NotEmpty(t, matches, "session file should exist")

	data, readErr := os.ReadFile(matches[0])
	require.NoError(t, readErr)

	var saved map[string]any
	require.NoError(t, json.Unmarshal(data, &saved))

	assert.Equal(t, "metadata-session", saved["id"])
	assert.NotEmpty(t, saved["date"], "should include date")
	assert.NotEmpty(t, saved["title"], "should include title")
}
```

Note: `session_end_test.go` needs `"encoding/json"` added to the import block.

**Step 2: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestSessionEndHandler" ./internal/handler/...
```

Expected: All tests PASS.

**Step 3: Commit**

```bash
git add internal/handler/session_end_test.go
git commit -m "test: add SessionEnd parity tests for short sessions and metadata"
```

---

### Task 7: Notification parity tests (Audio + Desktop)

**Files:**
- Modify: `internal/handler/notification_test.go`

**Step 1: Add mock types and audio parity tests**

Add at the top of the file (after imports), then add test cases after existing tests:

```go
// mockAudioPlayer records Play calls for assertion.
type mockAudioPlayer struct {
	played []string
}

func (m *mockAudioPlayer) Play(filepath string) error {
	m.played = append(m.played, filepath)
	return nil
}

// mockCmdRunner records Run calls for assertion.
type mockCmdRunner struct {
	calls []cmdRunnerCall
}

type cmdRunnerCall struct {
	name string
	args []string
}

func (m *mockCmdRunner) Run(name string, args ...string) error {
	m.calls = append(m.calls, cmdRunnerCall{name: name, args: args})
	return nil
}
```

Add audio parity tests after `TestNotifyAudioHandler_Disabled`:

```go
func TestNotifyAudioHandler_NoPlayerInjected(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled:   true,
				Directory: "/tmp/sounds",
			},
		},
	}

	// No WithAudioPlayer option — player is nil.
	h := handler.NewNotifyAudioHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyAudioHandler_EnabledWithPlayer(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create an audio file so PlayRandom has something to pick.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "beep.wav"), []byte("fake-audio"), 0o600,
	))

	player := &mockAudioPlayer{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled:   true,
				Directory: tmpDir,
			},
		},
	}

	h := handler.NewNotifyAudioHandler(cfg, handler.WithAudioPlayer(player))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotEmpty(t, player.played, "should have played an audio file")
}

func TestNotifyAudioHandler_QuietHoursSkipsPlay(t *testing.T) {
	t.Parallel()
	player := &mockAudioPlayer{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Audio: config.AudioValues{
				Enabled:   true,
				Directory: "/tmp/sounds",
			},
			QuietHours: config.QuietHoursValues{
				Enabled: true,
				Start:   "00:00",
				End:     "23:59",
			},
		},
	}

	h := handler.NewNotifyAudioHandler(cfg, handler.WithAudioPlayer(player))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Empty(t, player.played, "should not play during quiet hours")
}
```

**Step 2: Add desktop parity tests**

Add after `TestNotifyDesktopHandler_QuietHoursActive`:

```go
func TestNotifyDesktopHandler_NoRunnerInjected(t *testing.T) {
	t.Parallel()
	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	// No WithCmdRunner option — runner is nil.
	h := handler.NewNotifyDesktopHandler(cfg)
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Test",
		Message:       "Hello",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
}

func TestNotifyDesktopHandler_EnabledWithRunner(t *testing.T) {
	t.Parallel()
	runner := &mockCmdRunner{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg, handler.WithCmdRunner(runner))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Build Done",
		Message:       "All tests passed",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotEmpty(t, runner.calls, "should have called runner")
}

func TestNotifyDesktopHandler_CustomTitleAndMessage(t *testing.T) {
	t.Parallel()
	runner := &mockCmdRunner{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg, handler.WithCmdRunner(runner))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		Title:         "Custom Title",
		Message:       "Custom body text",
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, runner.calls)
	// The runner should receive the custom title/message (exact args depend on
	// the notify.Desktop implementation, but it should be called).
}

func TestNotifyDesktopHandler_DefaultTitleAndMessage(t *testing.T) {
	t.Parallel()
	runner := &mockCmdRunner{}

	cfg := &config.Values{
		Notify: config.NotifyValues{
			Desktop: config.DesktopValues{
				Enabled: true,
			},
		},
	}

	h := handler.NewNotifyDesktopHandler(cfg, handler.WithCmdRunner(runner))
	input := &hookcmd.HookInput{
		HookEventName: hookcmd.EventNotification,
		// No Title or Message — should use defaults.
	}

	resp, err := h.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.ExitCode)
	assert.NotEmpty(t, runner.calls,
		"should still send notification with default title/message")
}
```

Note: `notification_test.go` needs `"os"` and `"path/filepath"` added to imports.

**Step 2: Run tests**

```bash
gotestsum --format pkgname -- -tags=testmode -run "TestNotifyAudioHandler|TestNotifyDesktopHandler" ./internal/handler/...
```

Expected: All tests PASS.

**Step 3: Commit**

```bash
git add internal/handler/notification_test.go
git commit -m "test: add Notification parity tests for audio, desktop, quiet hours, mocks"
```

---

### Final Verification

After all 7 tasks, run the full handler test suite:

```bash
gotestsum --format pkgname -- -tags=testmode ./internal/handler/...
```

Expected: All tests PASS with 0 failures.

Then run the full project checks:

```bash
task check
```
