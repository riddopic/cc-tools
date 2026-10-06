package compact

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// tailWindow is how many bytes from the end of the transcript are scanned.
// The most recent assistant turn is near the end, so a full read of a
// multi-megabyte transcript on every tool call is unnecessary.
const tailWindow = 512 * 1024

// ErrNoUsage is returned when the transcript tail holds no main-chain
// assistant entry with token usage.
var ErrNoUsage = errors.New("no assistant usage found in transcript")

// transcriptEntry is the subset of a transcript JSONL line needed to read
// context size.
type transcriptEntry struct {
	Type        string `json:"type"`
	IsSidechain bool   `json:"isSidechain"`
	Message     struct {
		Usage *usage `json:"usage"`
	} `json:"message"`
}

// usage holds the token counts Claude reports for one assistant turn.
type usage struct {
	InputTokens              int `json:"input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// ContextTokens returns the context size, in tokens, of the latest main-chain
// assistant turn in the transcript at transcriptPath. The size is the sum of
// the turn's input, cache-read and cache-creation tokens. Only the last
// tailWindow bytes of the file are read.
func ContextTokens(transcriptPath string) (int, error) {
	if transcriptPath == "" {
		return 0, errors.New("empty transcript path")
	}

	f, err := os.Open(transcriptPath) // #nosec G304 -- path comes from the hook payload
	if err != nil {
		return 0, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	r, err := tailReader(f)
	if err != nil {
		return 0, err
	}

	tokens, found, err := lastUsage(r)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, ErrNoUsage
	}

	return tokens, nil
}

// tailReader positions f at the first full line inside the tail window and
// returns a buffered reader over the rest of the file.
func tailReader(f *os.File) (*bufio.Reader, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat transcript: %w", err)
	}

	offset := info.Size() - tailWindow
	if offset <= 0 {
		return bufio.NewReader(f), nil
	}

	// Start one byte early so a window that begins exactly on a line
	// boundary does not drop that line when the partial line is skipped.
	if _, seekErr := f.Seek(offset-1, io.SeekStart); seekErr != nil {
		return nil, fmt.Errorf("seek transcript: %w", seekErr)
	}

	r := bufio.NewReader(f)
	if _, skipErr := r.ReadBytes('\n'); skipErr != nil {
		if errors.Is(skipErr, io.EOF) {
			return r, nil
		}

		return nil, fmt.Errorf("skip partial line: %w", skipErr)
	}

	return r, nil
}

// lastUsage scans every line from r and returns the context size of the last
// main-chain assistant entry that reports usage.
func lastUsage(r *bufio.Reader) (int, bool, error) {
	tokens, found := 0, false

	for {
		line, err := r.ReadBytes('\n')
		if n, ok := parseUsage(line); ok {
			tokens, found = n, true
		}

		if errors.Is(err, io.EOF) {
			return tokens, found, nil
		}
		if err != nil {
			return 0, false, fmt.Errorf("read transcript: %w", err)
		}
	}
}

// parseUsage returns the context size reported by one transcript line, if it
// is a main-chain assistant entry with usage.
func parseUsage(line []byte) (int, bool) {
	if !bytes.Contains(line, []byte(`"usage"`)) {
		return 0, false
	}

	var entry transcriptEntry
	if err := json.Unmarshal(line, &entry); err != nil {
		return 0, false
	}

	if entry.Type != "assistant" || entry.IsSidechain || entry.Message.Usage == nil {
		return 0, false
	}

	u := entry.Message.Usage

	return u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens, true
}
