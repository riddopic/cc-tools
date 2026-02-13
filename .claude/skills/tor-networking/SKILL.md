---
name: tor-networking
description: Apply Tor-compatible networking patterns. Use when creating HTTP clients, making network calls, adding API integrations, or writing code that makes external requests. Ensures all network traffic routes through Tor when enabled.
---

# Tor-Compatible Networking Patterns

## Core Rule

**NEVER create `http.Client{}` directly.** Always use the Tor-aware factory.

## Creating HTTP Clients

```go
// ✅ DO: Use CreateHTTPClient() from cmd package
client, err := cmd.CreateHTTPClient()
if err != nil {
    return fmt.Errorf("failed to create HTTP client: %w", err)
}

// ✅ DO: Accept *http.Client in config structs
type MyConfig struct {
    HTTPClient *http.Client // Injected by caller
    Timeout    time.Duration
}

// ❌ DON'T: Create bare HTTP clients
client := &http.Client{Timeout: 30 * time.Second}
```

## Subprocess Network Calls

Forge/Anvil subprocesses inherit Tor env vars automatically via ProcessRunner.
No additional configuration needed for whitelisted commands.

For new subprocess commands that make network calls:
1. Ensure the command is in the ProcessRunner whitelist
2. The ProcessRunner automatically injects `ALL_PROXY`, `HTTP_PROXY`, `HTTPS_PROXY`

## Key Files

| File | Purpose |
|------|---------|
| `cmd/root.go` | `CreateHTTPClient()` - Tor-aware HTTP client factory |
| `internal/tor/environment.go` | `GetProxyEnvVars()` - Tor env var management |
| `internal/proxy/factory.go` | SOCKS5 transport creation |
| `internal/forge/runner.go` | Subprocess env var injection |

## DNS Leak Prevention

Always use `socks5h://` (not `socks5://`) when constructing proxy URLs.
The 'h' suffix ensures DNS resolution goes through the proxy.

## Local Connections

Local connections (localhost, 127.0.0.1, ::1) bypass Tor automatically.
This is correct behavior for Anvil health checks.
