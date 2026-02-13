---
name: rag-knowledge-system
description: Develop and maintain RAG retrieval, vector indexing, and safety gates. Use when working with vector store implementation, pattern indexing, or knowledge safety gates.
context: fork
---

# RAG Knowledge System Skill

Develop and maintain RAG retrieval, vector indexing, and safety gates for pattern-augmented analysis.

## Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Vector Store | `internal/knowledge/vector/` | ChromemGo-backed semantic search |
| Embedder | `internal/knowledge/embedder/` | OpenAI text-embedding-3-small |
| Indexer | `internal/knowledge/indexer/` | DeFiHackLabs parser + sanitizer |
| Retriever | `internal/knowledge/retriever/` | Three-phase: Scan -> Deep Dive -> Backtrack |
| Safety Gate | `internal/knowledge/safety/` | Discovery-oriented compliance |

## Three-Phase Retrieval

### Phase 1: Scan
- Parallel similarity search across pattern categories
- Returns candidate patterns above threshold (default: 0.7)
- Uses ChromemGo vector similarity

### Phase 2: Deep Dive
- Re-ranks candidates by function signature match
- Applies confidence weighting based on code overlap
- Filters low-relevance results

### Phase 3: Backtrack
- If confidence < 0.7, expand search to related patterns
- Uses semantic similarity to find adjacent vulnerabilities
- Prevents false negatives from narrow initial search

## Safety Gate Rules

The safety gate enforces discovery-oriented compliance on all retrieved patterns.

### Must Block
- Protocol names: Uniswap, Aave, Curve, Balancer, Compound, MakerDAO, etc.
- Dollar amounts: $1M, 1000 ETH, 500k USDC, etc.
- Explicit steps: "Step 1:", "Attack flow:", "Execute:"
- Transaction hashes: 0x prefixed 64-char hex strings
- Block numbers with context: "at block 18500000"

### Allowed (Infrastructure)
- DEX router addresses (for profit extraction)
- WETH/WBNB addresses
- Flash loan provider addresses
- Factory contract addresses

## CLI Commands

```bash
# Index DeFiHackLabs vulnerability patterns
quanta index --repo-path ~/DeFiHackLabs

# Index with custom embedding provider
quanta index --repo-path ~/DeFiHackLabs --embedder openai

# Check vector store health
quanta doctor

# Use RAG in analysis
quanta analyze 0x1234... --rag
```

## Testing

### Unit Tests
```go
// Use fixed embeddings for deterministic tests
embedder := mocks.NewMockEmbedder(t)
embedder.EXPECT().Embed(mock.Anything, "test input").
    Return([]float32{0.1, 0.2, 0.3}, nil)

// Mock vector store
store := mocks.NewMockVectorStore(t)
store.EXPECT().Search(mock.Anything, mock.Anything, 10).
    Return([]vector.Result{{ID: "p1", Score: 0.95}}, nil)
```

### Safety Gate Tests
```go
// Verify gate catches 100% of violations
gate := safety.NewGate(defaultBlocklist)

// Should block
assert.Error(t, gate.Validate("Uniswap V2 exploit"))
assert.Error(t, gate.Validate("Lost $26.6M"))
assert.Error(t, gate.Validate("Step 1: Flash loan"))

// Should allow
assert.NoError(t, gate.Validate("Public burn function vulnerability"))
assert.NoError(t, gate.Validate("INVESTIGATE: Check access control"))
```

## Enhanced RAG: Body-Aware Scanning

The `SolidityParser` extends indexing beyond function signatures to include full function bodies, modifiers, and state variables.

### SolidityElement Struct

Each parsed element carries:
- `Name` — function/modifier/variable name
- `ElementType` — `function`, `modifier`, `stateVar`
- `Signature` — full function signature
- `Body` — complete implementation source
- `Modifiers` — applied modifiers list
- `StateReads` / `StateWrites` — state variable access patterns

### Extended Three-Phase Retrieval

The existing three-phase pipeline gains two new sub-phases:

- **Phase 1 extended**: `ScanBodies()` — searches function body content for vulnerability patterns (e.g., unchecked external calls, missing access checks)
- **Phase 2 extended**: `ScanStatePatterns()` — identifies state variable manipulation patterns across functions (e.g., balance updates before external calls)

The `EnhancedRetriever` wraps the base retriever and merges body-aware results with signature-based results, deduplicating by element identity.

## Smart Chunking

The `Chunker` splits large Solidity elements into embedding-friendly segments:

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Max tokens | 512 | Fits embedding model context |
| Overlap | 64 tokens | Preserves cross-boundary patterns |
| Token estimate | ~4 chars/token | Heuristic for Solidity code |

Key methods:
- `ChunkText(text string) []Chunk` — splits raw text with overlap
- `ChunkElements(elements []SolidityElement) []Chunk` — splits parsed elements, preserving element metadata per chunk

## Index Versioning

The vector index includes a `metadata.json` file tracking the index schema version:

| Version | Schema | Description |
|---------|--------|-------------|
| v1 | Legacy | Signature-only indexing |
| v2 | Body-aware | Includes function bodies, modifiers, state variables |

- Version is stored in `metadata.json` `Version` field
- v1 → v2 migration requires full re-index (`quanta index --rebuild`)
- The retriever auto-detects version and adjusts query strategy accordingly

## Files

| File | Purpose |
|------|---------|
| `internal/knowledge/vector/store.go` | Vector store interface |
| `internal/knowledge/vector/chromem.go` | ChromemGo implementation |
| `internal/knowledge/embedder/openai.go` | OpenAI embedder |
| `internal/knowledge/embedder/chromem.go` | Built-in ChromemGo embedder |
| `internal/knowledge/indexer/indexer.go` | DeFiHackLabs parser |
| `internal/knowledge/indexer/sanitizer.go` | Protocol name sanitizer |
| `internal/knowledge/indexer/solidity_parser.go` | Body-aware Solidity parser |
| `internal/knowledge/indexer/chunker.go` | Smart text/element chunker |
| `internal/knowledge/retriever/retriever.go` | Three-phase retrieval |
| `internal/knowledge/retriever/retriever.go` | Base + body-aware enhanced retrieval (EnhancedRetriever wraps base) |
| `internal/knowledge/safety/gate.go` | Safety gate implementation |
| `cmd/index.go` | CLI index command |

## Related Skills

- `/discovery-oriented-prompts` - Principles for prompt content
- `/pattern-management` - Hardcoded pattern management
- `/testing-patterns` - Testing patterns for RAG components
