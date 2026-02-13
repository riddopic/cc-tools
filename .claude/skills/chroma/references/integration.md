# chromem-go Integration Patterns

How chromem-go connects with Quanta subsystems.

## RAG Pipeline Initialization

The `cmd/analyze.go` command wires the full pipeline:

```
Embedder → ChromemStore → Retriever → Safety Gate → Analysis
```

```go
// 1. Create embedder
emb, err := embedder.NewOpenAI(embedder.Config{
    APIKey:     viper.GetString("openai.api_key"),
    Model:      embedder.DefaultModel,
    Dimensions: embedder.DefaultDimensions,
})

// 2. Open persistent store at rag.store_path
storePath := viper.GetString("rag.store_path")
store, err := vector.NewChromemStore(storePath, "patterns", emb)
defer store.Close()

// 3. Wire into retriever for three-phase search
retriever := retriever.New(store, emb, gate)
```

## Indexing Pipeline

The `cmd/index.go` command builds the vector index:

```go
// Parse Solidity files from DeFiHackLabs
elements := parser.ParseDirectory(repoPath)

// Chunk large elements for embedding-friendly segments
chunks := chunker.ChunkElements(elements)

// Convert to documents with metadata
docs := make([]vector.Document, len(chunks))
for i, c := range chunks {
    docs[i] = vector.Document{
        ID:       c.ID,
        Content:  c.Text,
        Metadata: c.Metadata,
    }
}

// Batch-embed and store
store.Add(ctx, docs)
```

## Embedder ↔ Store Connection

The `ChromemStore` adapts `embedder.Embedder` to chromem-go's `EmbeddingFunc`:

```go
// Internal adapter — called by chromem when Document.Embedding is nil
chromem.EmbeddingFunc(func(ctx context.Context, text string) ([]float32, error) {
    return emb.Embed(ctx, text)
})
```

For pre-computed embeddings, set `Document.Embedding` before calling `Add()` — chromem-go skips the embedding function when the vector is already present.

## Retriever ↔ Store Connection

The three-phase retriever uses `QueryWithEmbedding`:

```go
// Phase 1 (Scan): parallel similarity search across categories
queryVec, _ := emb.Embed(ctx, contractSource)
results, _ := store.QueryWithEmbedding(ctx, queryVec, topK)

// Phase 2 (Deep Dive): re-rank by function signature match
// Phase 3 (Backtrack): expand search if confidence < 0.7
```

## Common Integration Mistakes

### BAD: Mixing embedder providers between index and query time

```go
// Indexed with OpenAI text-embedding-3-small (1536 dims)
store, _ := vector.NewChromemStore(path, "patterns", openaiEmb)
store.Add(ctx, docs)

// Later, querying with mock embedder — dimensions match but vectors are incompatible
mockVec, _ := mockEmb.Embed(ctx, query)
results, _ := store.QueryWithEmbedding(ctx, mockVec, 10) // Garbage results
```

### GOOD: Validate metadata before querying

```go
err := embedder.ValidateMetadata(indexMeta, queryMeta)
if err != nil {
    return fmt.Errorf("incompatible embedder: %w", err)
}
```

### BAD: Forgetting to close the store

```go
store, _ := vector.NewChromemStore(path, "patterns", emb)
// No Close() — may lose persistence data
```

### GOOD: Always defer Close

```go
store, _ := vector.NewChromemStore(path, "patterns", emb)
defer store.Close()
```

### BAD: Using Query() on ChromemStore

```go
results, err := store.Query(ctx, "reentrancy", 10)
// Returns ErrEmbeddingRequired — ChromemStore requires pre-computed embeddings
```

### GOOD: Embed first, then QueryWithEmbedding

```go
vec, _ := emb.Embed(ctx, "reentrancy")
results, _ := store.QueryWithEmbedding(ctx, vec, 10)
```
