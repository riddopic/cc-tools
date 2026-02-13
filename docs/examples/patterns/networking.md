# Networking Patterns

## Creating Tor-Aware HTTP Clients

```go
// Factory pattern - returns Tor-routed client when --use-tor enabled
client, err := cmd.CreateHTTPClient()
if err != nil {
    return fmt.Errorf("failed to create HTTP client: %w", err)
}
```

## Accepting Injected Clients

```go
type ExplorerConfig struct {
    APIKey     string
    HTTPClient *http.Client // Caller injects Tor-aware client
    Timeout    time.Duration
}

func NewExplorer(config ExplorerConfig) *Explorer {
    httpClient := config.HTTPClient
    if httpClient == nil {
        httpClient = &http.Client{Timeout: config.Timeout}
    }
    return &Explorer{httpClient: httpClient}
}
```

## Subprocess Proxy Injection

The `ProcessRunner` automatically injects Tor proxy environment variables
into all subprocess environments when Tor routing is enabled. Tor config
is injected via the constructor, not read from global state:

```go
// internal/forge/runner.go - automatic injection via TorConfig
if r.torConfig.Enabled {
    for k, v := range tor.GetProxyEnvVars(r.torConfig.Proxy) {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }
}
```

No manual configuration is needed for Forge/Anvil subprocesses.

## DNS Leak Prevention

Always use `socks5h://` (not `socks5://`) when constructing proxy URLs.
The `h` suffix ensures DNS resolution goes through the proxy:

```go
// ✅ DO: Use socks5h for DNS-safe proxying
proxyURL := "socks5h://" + proxyAddr

// ❌ DON'T: Use socks5 (leaks DNS queries)
proxyURL := "socks5://" + proxyAddr
```
