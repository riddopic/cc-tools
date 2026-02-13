---
name: pattern-management
description: Add, update, or debug vulnerability patterns in the LLM prompts. Use when adding new exploit patterns, updating pattern detection logic, or debugging pattern selection issues.
---

# Pattern Management Skill

Add, update, or debug vulnerability patterns in the LLM prompts.

## Usage

```
/pattern-management <action> [pattern_num]
```

Actions: `list`, `add`, `update`, `debug`

## Quick Reference

### Current Patterns

| Pattern | Name | Detection Trigger | Key Guidance |
|---------|------|-------------------|--------------|
| 1 | Reentrancy | `call{value:` + external call | Check-effects-interactions violation |
| 3 | Flash Loan | `flashLoan`, `flashBorrow` | Borrow → manipulate → repay |
| 6 | LP Manipulation | `family()`, `burn()` on LP | Buy → burn LP → sync → sell |
| 8 | Access Control | Public `mint()`, `burn()` | Call privileged function directly |
| 9 | Oracle Manipulation | `getPrice`, `latestAnswer` | Manipulate price source |
| 10 | Proxy/Mint Logic | Proxy + `mint()` + allowlist | Swap to allowed token → mint → extract |
| 11 | Multi-Token Reward | `releasePressure`, `distribute` | Trigger with `transfer(self, 0)` → swap reward token |

### Pattern Structure

```go
func detectPatternN(source, sourceLower string, isProxy bool) *PatternHint {
    // 1. Check detection triggers
    if !hasRequiredSignature(source) {
        return nil
    }

    // 2. Return pattern hint with guidance
    return &PatternHint{
        PatternNum:  PatternN,
        PatternName: "Pattern Name",
        Confidence:  "high", // or "medium"
        Guidance:    "Step-by-step attack instructions with addresses",
    }
}
```

### Pattern Guidance Principles

Pattern guidance should enable discovery, NOT prescribe specific exploits.

**DO**:
- Use numbered "INVESTIGATE:" questions that guide analysis
- End with "KEY INSIGHT:" about the vulnerability mechanism
- Focus on what to look for, not what to do
- Describe code patterns generically (e.g., "public burn function")

**DON'T**:
- Include hardcoded addresses (router, WETH, etc.)
- Provide step-by-step attack recipes
- Reference specific contracts or dollar amounts
- Name protocols in pattern names (e.g., "Balancer-style")

See `/discovery-oriented-prompts` skill for detailed prompt writing guidance.

## Workflow

### Adding a New Pattern

1. Identify the vulnerability type from failed exploit or research
2. Find detection triggers (function names, patterns)
3. Create `detectPatternN()` helper function
4. Add call in `detectApplicablePatterns()`
5. Test with contract that has this vulnerability
6. Document in `docs/CLAUDE-CODE-SDK-E2E-FINDINGS.md`

### Updating Existing Pattern

1. Read current implementation:
   ```bash
   rg -A 20 "detectPattern10" internal/llm/prompts.go
   ```

2. Identify gap (missing step, wrong address, etc.)
3. Update guidance string with fix
4. Verify with `task lint` (check funlen < 100)
5. Re-test with `/table-ix-testing`

### Debugging Pattern Selection

1. Check if pattern is detected:
   ```bash
   # Add debug logging to detect function
   ```

2. Verify source code contains triggers
3. Check `isProxy` flag if proxy-dependent

## Indexed Patterns vs. Hardcoded Patterns

### Hardcoded Patterns (in internal/llm/prompts.go)
- 25+ pattern detection functions
- Triggered by code signatures
- Provides initial guidance

### Indexed Patterns (in vector store)
- 400+ DeFiHackLabs patterns
- Triggered by similarity search
- Provides contextual examples

### Adding Patterns to Vector Store
1. Run `quanta index --repo-path <DeFiHackLabs>`
2. Sanitizer removes protocol names/dollar amounts
3. Embeddings computed for vector search
4. Pattern card stored with metadata

### Enhanced RAG Body-Aware Indexing

The v2 index extends pattern indexing with function body content:

- **SolidityParser** extracts function bodies, modifiers, and state variables from DeFiHackLabs POC contracts
- **Chunker** splits large function bodies into 512-token chunks with 64-token overlap for embedding
- **Index version** tracked in `metadata.json` — v1 (signature-only) vs v2 (body-aware)
- Re-index with `quanta index --rebuild` to upgrade from v1 → v2
- PatternCard extensions: `FunctionBodies`, `StateVariables`, `ModifierList` fields added to indexed metadata

### RAG Pattern Files
- Vector Store: `internal/knowledge/vector/`
- Indexer: `internal/knowledge/indexer/`
- Solidity Parser: `internal/knowledge/indexer/solidity_parser.go`
- Chunker: `internal/knowledge/indexer/chunker.go`
- Safety Gate: `internal/knowledge/safety/gate.go`

## Files

- Pattern Implementation: `internal/llm/prompts.go`
- Pattern Constants: `internal/llm/prompts.go` (PatternN constants)
- Test Helpers: `internal/llm/prompts_test.go`
- Vector Store Interface: `internal/knowledge/vector/store.go`
- Pattern Sanitizer: `internal/knowledge/indexer/sanitizer.go`
- Solidity Parser: `internal/knowledge/indexer/solidity_parser.go`
- Chunker: `internal/knowledge/indexer/chunker.go`
