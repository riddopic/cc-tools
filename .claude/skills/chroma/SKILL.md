---
name: chroma
description: Apply chromem-go vector database patterns. Use when working with vector store operations, embedding persistence, collection management, or semantic search in Go.
---

# chromem-go Vector Database Patterns

Patterns for using chromem-go (`github.com/philippgille/chromem-go`), the embedded pure-Go vector database backing Quanta's RAG system.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Store interface | `internal/knowledge/vector/store.go` | `vector.Store` contract |
| ChromemStore | `internal/knowledge/vector/chromem.go` | chromem-go persistent implementation |
| InMemoryStore | `internal/knowledge/vector/memory_store.go` | In-memory test double with cosine similarity |
| Embedder interface | `internal/knowledge/embedder/embedder.go` | `embedder.Embedder` contract |
| OpenAI Embedder | `internal/knowledge/embedder/openai.go` | Production embeddings (text-embedding-3-small) |
| Mock Embedder | `internal/knowledge/embedder/embedder.go` | Deterministic hash-based test embeddings |
| Config | `internal/knowledge/vector/store.go` | `vector.Config` with store path, dimensions |

## Key Types

```go
// Store — the vector store interface (internal/knowledge/vector/store.go)
type Store interface {
    Add(ctx context.Context, docs []Document) error
    Query(ctx context.Context, query string, k int) ([]SearchResult, error)
    QueryWithEmbedding(ctx context.Context, embedding []float32, k int) ([]SearchResult, error)
    Delete(ctx context.Context, ids []string) error
    Get(ctx context.Context, id string) (*Document, error)
    Count(ctx context.Context) (int, error)
    Close() error
}

// Document — indexable unit with optional pre-computed embedding
type Document struct {
    ID        string
    Content   string
    Metadata  map[string]any
    Embedding []float32      // Optional: pre-computed embedding
}

// SearchResult — query match with similarity score
type SearchResult struct {
    Document   Document
    Score      float64
    Similarity float64
}
```

## Creating a ChromemStore

```go
import (
    "github.com/riddopic/quanta/internal/knowledge/embedder"
    "github.com/riddopic/quanta/internal/knowledge/vector"
)

// Production: persistent store with OpenAI embedder
emb, err := embedder.NewOpenAI(embedder.Config{
    APIKey:     os.Getenv("OPENAI_API_KEY"),
    Model:      embedder.DefaultModel,        // "text-embedding-3-small"
    Dimensions: embedder.DefaultDimensions,    // 1536
})
store, err := vector.NewChromemStore(storePath, "patterns", emb)
defer store.Close()
```

Persistence uses GOB format at `<store_path>/chromem.gob`. The `rag.store_path` config key controls the base directory.

## Adding Documents

```go
docs := []vector.Document{
    {
        ID:      "vuln-reentrancy-001",
        Content: "Reentrancy via external call before state update",
        Metadata: map[string]any{
            "category": "reentrancy",
            "severity": "critical",
        },
    },
}
err := store.Add(ctx, docs)
```

When `Embedding` is nil, chromem-go calls the embedder automatically. Pre-compute embeddings for batch efficiency:

```go
embedding, err := emb.Embed(ctx, doc.Content)
doc.Embedding = embedding
```

## Querying

ChromemStore requires pre-computed query embeddings — `Query()` returns `ErrEmbeddingRequired`:

```go
// Embed the query first
queryEmb, err := emb.Embed(ctx, "unchecked external call")

// Search with pre-computed embedding
results, err := store.QueryWithEmbedding(ctx, queryEmb, 10)
for _, r := range results {
    fmt.Printf("ID: %s, Score: %.3f\n", r.Document.ID, r.Score)
}
```

## Thread Safety

`ChromemStore` uses `sync.RWMutex` — reads (`QueryWithEmbedding`, `Get`, `Count`) take read locks; writes (`Add`, `Delete`) take write locks. Safe for concurrent goroutine access.

## Embedder Integration

The embedder bridges the `Store` and external embedding APIs:

```go
// chromem.EmbeddingFunc adapter (internal to ChromemStore)
func createEmbeddingFunc(emb embedder.Embedder) chromem.EmbeddingFunc {
    return func(ctx context.Context, text string) ([]float32, error) {
        return emb.Embed(ctx, text)
    }
}
```

Validate embedder compatibility before querying an existing index:

```go
err := embedder.ValidateMetadata(indexMeta, queryMeta)
// Returns ErrProviderMismatch, ErrModelMismatch, or ErrDimensionsMismatch
```

## Testing Patterns

### Unit tests — mock the Store interface

```go
store := mocks.NewMockStore(t)
store.EXPECT().QueryWithEmbedding(mock.Anything, mock.Anything, 10).
    Return([]vector.SearchResult{
        {Document: vector.Document{ID: "p1", Content: "test"}, Score: 0.95},
    }, nil)
store.EXPECT().Close().Return(nil)
```

### Integration-style tests — use InMemoryStore

```go
cfg := vector.DefaultConfig()
store := vector.NewInMemoryStore(cfg)
defer store.Close()

// Add with pre-computed embeddings (required for InMemoryStore)
emb := embedder.NewMockEmbedder(1536)
vec, _ := emb.Embed(ctx, "reentrancy vulnerability")

store.Add(ctx, []vector.Document{{
    ID: "p1", Content: "reentrancy vulnerability", Embedding: vec,
}})

queryVec, _ := emb.Embed(ctx, "reentrancy vulnerability")
results, _ := store.QueryWithEmbedding(ctx, queryVec, 5)
assert.Equal(t, "p1", results[0].Document.ID)
```

### Deterministic embeddings

`MockEmbedder` uses FNV hashing — same text always produces the same vector. Use for reproducible similarity tests.

## Anti-Patterns

```go
// BAD: Never use Python chromadb — Quanta uses chromem-go (pure Go, embedded)
// pip install chromadb          // Wrong ecosystem
// import chromadb               // Wrong language

// BAD: Never create bare HTTP clients for embedding APIs
client := &http.Client{Timeout: 30 * time.Second}

// GOOD: Inject HTTPClient via embedder.Config
cfg := embedder.Config{HTTPClient: myClient}

// BAD: Never hardcode embedding dimensions
embedding := make([]float32, 1536)

// GOOD: Use the constant
embedding := make([]float32, vector.DefaultEmbeddingDimensions)

// BAD: Never skip metadata validation on existing indices
results, _ := store.QueryWithEmbedding(ctx, queryVec, 10) // May mix incompatible embeddings

// GOOD: Validate embedder compatibility first
err := embedder.ValidateMetadata(indexMeta, queryMeta)
```

## CLI Commands

```bash
# Index DeFiHackLabs vulnerability patterns
quanta index --repo-path ~/DeFiHackLabs

# Check vector store health
quanta doctor

# Analyze with RAG-augmented retrieval
quanta analyze 0x1234... --rag
```

## Files

| File | Purpose |
|------|---------|
| `internal/knowledge/vector/store.go` | `Store` interface, `Document`, `SearchResult`, `Config` types |
| `internal/knowledge/vector/chromem.go` | `ChromemStore` — persistent chromem-go implementation |
| `internal/knowledge/vector/memory_store.go` | `InMemoryStore` — in-memory test double |
| `internal/knowledge/embedder/embedder.go` | `Embedder` interface, `MockEmbedder`, `ValidateMetadata` |
| `internal/knowledge/embedder/openai.go` | `OpenAI` embedder (text-embedding-3-small) |
| `cmd/index.go` | CLI index command — builds and persists vector index |
| `cmd/analyze.go` | CLI analyze command — initializes RAG pipeline |

## Related Skills

- `rag-knowledge-system` — Three-phase retrieval pipeline, safety gates, indexing workflow
- `testing-patterns` — Mockery v3.5 patterns, table-driven tests
- `go-coding-standards` — Error wrapping, interface design, concurrency
