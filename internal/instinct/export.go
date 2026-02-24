package instinct

import (
	"encoding/json"
	"fmt"
	"io"
)

// ExportYAML writes instincts in YAML frontmatter format to w.
// Each instinct is separated by a blank line.
func ExportYAML(w io.Writer, instincts []Instinct) error {
	for i, inst := range instincts {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return fmt.Errorf("write separator: %w", err)
			}
		}

		data := marshalFrontmatter(inst)
		if _, err := io.WriteString(w, data); err != nil {
			return fmt.Errorf("write instinct %s: %w", inst.ID, err)
		}
	}

	return nil
}

// ExportJSON writes instincts in JSON array format to w.
func ExportJSON(w io.Writer, instincts []Instinct) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(instincts); err != nil {
		return fmt.Errorf("encode instincts: %w", err)
	}

	return nil
}
