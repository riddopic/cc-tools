# Consolidate Ad-Hoc Logger/Interface Adapters

**Extracted:** 2026-02-10
**Status:** ALL DEBT RESOLVED (2026-02-10)
**Context:** When cmd/ packages define one-off struct adapters to satisfy internal interfaces

## Problem

`cmd/` packages contained 11 ad-hoc adapter structs that wrapped internal types to satisfy `internal/` interfaces. This pattern:

- Duplicates logic that belongs in `internal/` packages
- Cannot be reused by other commands or tests
- Requires implementation-detail imports in cmd/
- Creates hidden coupling: if the interface changes, cmd/ must change too
- In 4 cases, identical adapters were duplicated between cmd/ and sourcecache/

## Solution

Move each adapter to the appropriate `internal/` package as a reusable exported type. Key patterns:

```go
// internal/agent/multi_actor_adapter.go
var _ MultiActorCoordinator = (*MultiActorAdapter)(nil)  // compile-time check

type MultiActorAdapter struct {
    Coord         *coordinator.StateCoordinator  // exported for test construction
    LastCoordPlan *coordinator.AttackPlan
}
```

## Execution Learnings

### Import Cycle Prevention

Before moving ANY adapter, check bidirectional imports:

```bash
# Does destination already import source?
rg '"github.com/riddopic/quanta/internal/agent"' internal/sourcecache/ --glob '*.go'
# Does source already import destination?
rg '"github.com/riddopic/quanta/internal/sourcecache"' internal/agent/ --glob '*.go'
```

This caught a real issue: `stateCollectorAdapter` was planned for `internal/agent/` but `sourcecache` already imports `agent`, creating a cycle. Redirected to `internal/sourcecache/` instead.

### Duplicate Reconciliation

When two copies exist (cmd/ and sourcecache/), compare them and pick the richer one:

- `stateCollectorAdapter`: sourcecache version had nil guard + complete GetStateSize (ByteCode, CodeHash, Views)
- `multiActorAdapter`: functionally identical, but sourcecache version used `any` return type matching the existing type assertion
- `registrySourceFetcher`: cmd/ version was richer (disk cache, proxy resolution, flattening)

### ireturn Linter

Methods like `CloneWithContract() RAGRetriever` return interfaces by design (cloner pattern). Add exclusion in `.golangci.yml`:

```yaml
- path: 'internal/agent/rag_retriever_adapter\.go'
  linters: [ireturn]
```

### Type Assertion Compatibility

Go interface type assertions are exact on return types. `GetLastCoordinatorPlan() *coordinator.AttackPlan` does NOT satisfy `interface{ GetLastCoordinatorPlan() any }`. Must use `any` return type to match existing assertion in orchestrator.go.

### Exported Fields for Test Construction

Use exported struct fields (not constructors) for adapters that tests construct with different field combinations:

```go
// Tests can mix-and-match fields
adapter := &agent.MultiActorAdapter{Coord: coord, LastCoordPlan: nil}
```

## When to Use

- When reviewing cmd/ packages and finding struct types that wrap library types
- When adding a second caller that needs the same adapter
- When the adapter requires importing implementation details (`zapcore`, etc.) into cmd/

## Prevention Rules

See `.claude/rules/patterns.md` "Adapter Placement" section and `.claude/rules/coding-style.md` "No Ad-Hoc Adapters in cmd/" rule.

## Completed Inventory (all 11 adapters + 3 interfaces)

| # | Adapter | Destination | Commit |
|---|---------|-------------|--------|
| 1 | `PriceService` | `internal/pricing/price_service.go` | prior session |
| 2 | `mockContextCache` | `internal/agent/memory_context_cache.go` | prior session |
| 3 | `simpleChainClient` | `internal/interfaces/noop.go` | prior session |
| 4 | `ethClientAdapter` | replaced with `blockchain.EthClientAdapter` | `86eccc02` |
| 5 | `sourceFetcher` iface | `internal/sourcecache/source_fetcher.go` | `51cb031f` |
| 6 | `registrySourceFetcher` | `internal/sourcecache/registry_fetcher.go` | `53e168ab` |
| 7 | `sourceAwareExecutor` | `internal/sourcecache/source_aware_executor.go` | `b947a873` |
| 8 | `stateCollectorAdapter` | `internal/sourcecache/state_collector_adapter.go` | `b11edfb6` |
| 9 | `analyzerStateCollector` | `internal/analyzer/rpc_state_collector.go` | `137c7f63` |
| 10 | `ragRetrieverAdapter` | `internal/agent/rag_retriever_adapter.go` | `2b3a809c` |
| 11 | `multiActorAdapter` | `internal/agent/multi_actor_adapter.go` | `cf1aeec5` |

Also completed earlier:

- **Logger**: `output.ExtendedLogger` deleted. `regression.Logger` aliased to `interfaces.SugaredLogger`.
- **ForgeExecutor**: NOT a duplicate (different abstraction layers). No action needed.
- **HTTPClient**: Canonical `interfaces.HTTPClient` created. Aliases in embedder/providers.
- **HTTP factory violations**: All 3 fixed (PriceService, openAISummarizer, contractNameResolver).
