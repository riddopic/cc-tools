---
paths:
  - "**/*.go"
---

# Go Security Guidelines

Security practices for the Quanta codebase.

## Input Validation

```go
// ✅ DO: Validate all user input explicitly
func SetTheme(theme string) error {
    validThemes := []string{"default", "powerline", "minimal", "classic"}

    for _, valid := range validThemes {
        if theme == valid {
            return nil
        }
    }

    return fmt.Errorf("invalid theme: %s", theme)
}

// ✅ DO: Sanitize file paths to prevent directory traversal
func LoadConfigFile(path string) error {
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return errors.New("invalid path: directory traversal detected")
    }

    info, err := os.Stat(cleanPath)
    if err != nil {
        return fmt.Errorf("cannot access file: %w", err)
    }

    if info.IsDir() {
        return errors.New("path is a directory, not a file")
    }

    return nil
}

// ❌ DON'T: Trust user input without validation
func LoadFile(path string) ([]byte, error) {
    return os.ReadFile(path)  // Allows ../../../etc/passwd
}
```

## Secret Management

```go
// ✅ DO: Use environment variables for secrets
func getAPIKey() (string, error) {
    key := os.Getenv("QUANTA_API_KEY")
    if key == "" {
        return "", errors.New("QUANTA_API_KEY not configured")
    }
    return key, nil
}

// ✅ DO: Use SecurityFixtures for test credentials
func TestWithCredentials(t *testing.T) {
    creds := testutil.SecurityFixtures.GetTestCredentials()
    // Use creds in test...
}

// ❌ DON'T: Hardcode secrets in code
const apiKey = "sk-proj-xxxxx"  // NEVER DO THIS

// ❌ DON'T: Log sensitive information
log.Printf("Connecting with API key: %s", apiKey)
```

## Sensitive Data Handling

```go
// ✅ DO: Clear sensitive data after use
func processSecret(secret string) error {
    // Use the secret
    result, err := authenticate(secret)

    // Clear from memory (best effort)
    secretBytes := []byte(secret)
    for i := range secretBytes {
        secretBytes[i] = 0
    }

    return err
}

// ✅ DO: Redact secrets in error messages
func connectWithKey(key string) error {
    if err := connect(key); err != nil {
        return fmt.Errorf("connection failed (key: %s...)", key[:4])
    }
    return nil
}
```

## Command Injection Prevention

```go
// ✅ DO: Use exec.Command with separate arguments
func runCommand(binary string, args ...string) error {
    cmd := exec.Command(binary, args...)
    return cmd.Run()
}

// ❌ DON'T: Pass user input to shell
func runShellCommand(userInput string) error {
    cmd := exec.Command("sh", "-c", userInput)  // DANGEROUS!
    return cmd.Run()
}

// ✅ DO: Validate command arguments
func runForge(testFile string) error {
    if !strings.HasSuffix(testFile, ".t.sol") {
        return errors.New("invalid test file extension")
    }
    if strings.ContainsAny(testFile, ";&|`$") {
        return errors.New("invalid characters in filename")
    }

    cmd := exec.Command("forge", "test", "--match-path", testFile)
    return cmd.Run()
}
```

## Error Message Safety

```go
// ✅ DO: Return user-friendly errors without internal details
func authenticate(token string) error {
    user, err := db.FindByToken(token)
    if err != nil {
        // Log internal error for debugging
        log.Error("auth failed", zap.Error(err))
        // Return generic error to user
        return errors.New("authentication failed")
    }
    return nil
}

// ❌ DON'T: Expose internal details in errors
func authenticate(token string) error {
    user, err := db.FindByToken(token)
    if err != nil {
        return fmt.Errorf("SQL error: %v", err)  // Leaks DB info!
    }
    return nil
}
```

## Race Condition Prevention

```go
// ✅ DO: Use sync primitives for shared state
type SafeCounter struct {
    mu    sync.RWMutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.count
}

// ✅ DO: Run tests with race detector
// make test-race
```

## Pre-Commit Security Checklist

Before ANY commit, verify:

- [ ] No hardcoded secrets (API keys, passwords, tokens)
- [ ] All user inputs validated
- [ ] File paths sanitized with `filepath.Clean()`
- [ ] No command injection vulnerabilities
- [ ] Error messages don't leak sensitive data
- [ ] Secrets loaded from environment variables
- [ ] Race detector passes (`make test-race`)
- [ ] No sensitive data in logs

## Quick Commands

```bash
# Check for potential secrets
rg -i "(password|secret|api.?key|token)" --type go

# Check for command injection risks
rg "exec\.Command.*\+|sh.*-c" --type go

# Run security checks
make security
make vulncheck
```
