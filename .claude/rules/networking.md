# Networking Guidelines

Rules for creating HTTP clients and making network calls in Quanta.

## HTTP Client Creation

NEVER create `http.Client{}` directly. Use the Tor-aware factory:

```go
// ✅ DO: Use centralized factory
client, err := cmd.CreateHTTPClient()

// ✅ DO: Accept injected client
type Config struct {
    HTTPClient *http.Client
}

// ❌ DON'T: Create bare clients
client := &http.Client{Timeout: 30 * time.Second}

// ❌ DON'T: Use default client
resp, err := http.Get(url)
```

## DNS Leak Prevention

Always use `socks5h://` (not `socks5://`) for proxy URLs.

## Subprocess Network Calls

ProcessRunner automatically injects Tor env vars. No manual setup needed.

## Key Files

- `cmd/root.go` - `CreateHTTPClient()`
- `internal/tor/environment.go` - `GetProxyEnvVars()`
- `internal/proxy/factory.go` - Transport creation
