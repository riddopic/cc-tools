---
name: tor-networking-review
description: Reviews code for Tor routing compliance. Use when reviewing .go files that create HTTP clients, make network calls, or spawn subprocesses with network access. Flags direct HTTP client creation that bypasses the Tor-aware factory.
---

# Tor Networking Review

## Review Checklist

- [ ] No direct `&http.Client{}` creation (must use `cmd.CreateHTTPClient()` or accept injected client)
- [ ] No `http.DefaultClient` usage (not Tor-aware)
- [ ] No `http.Get()` / `http.Post()` shorthand (uses DefaultClient)
- [ ] Config structs accept `*http.Client` for dependency injection
- [ ] Subprocess commands that need network inherit environment
- [ ] Proxy URLs use `socks5h://` scheme (not `socks5://`)
- [ ] No hardcoded `NO_PROXY` that could leak external traffic

## Patterns to Flag

| Pattern | Issue | Fix |
|---------|-------|-----|
| `&http.Client{}` | Bypasses Tor | Use `cmd.CreateHTTPClient()` |
| `http.Get(url)` | Uses DefaultClient | Create proper client first |
| `http.DefaultTransport` | Not Tor-aware | Use proxy factory transport |
| `socks5://` (without h) | DNS leak | Use `socks5h://` |
| `NO_PROXY=*` | Bypasses all proxy | Remove or limit to localhost |

## Valid Patterns (Do NOT Flag)

- `*http.Client` as struct field (dependency injection)
- `cmd.CreateHTTPClient()` usage
- `proxy.NewHTTPClientFactory()` usage
- Health checks to `localhost`/`127.0.0.1` (local Anvil)
- Test code using `httptest.NewServer()`

## Before Submitting Findings

Load and follow [review-verification-protocol](../review-verification-protocol/SKILL.md) before reporting any issue.
