package output_test

import (
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/output"
)

func TestTableInterface(_ *testing.T) {
	// Verify that both TableRenderer and MockTableRenderer implement TableInterface.
	var _ output.TableInterface = (*output.TableRenderer)(nil)
	var _ output.TableInterface = (*MockTableRenderer)(nil)
}

func assertTableContains(t *testing.T, result string, values ...string) {
	t.Helper()

	for _, val := range values {
		if !strings.Contains(result, val) {
			t.Errorf("expected %q in rendered table, got: %s", val, result)
		}
	}
}

func assertTableExcludes(t *testing.T, result string, values ...string) {
	t.Helper()

	for _, val := range values {
		if strings.Contains(result, val) {
			t.Errorf("did not expect %q in rendered table, got: %s", val, result)
		}
	}
}

func TestNewTable(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		widths  []int
		rows    [][]string
		want    []string
		exclude []string
	}{
		{
			name:    "empty table",
			headers: []string{"Column1", "Column2"},
			widths:  []int{10, 10},
			rows:    [][]string{},
			want:    []string{"Column1", "Column2"},
			exclude: nil,
		},
		{
			name:    "table with data",
			headers: []string{"Name", "Value"},
			widths:  []int{20, 30},
			rows: [][]string{
				{"Item1", "Value1"},
				{"Item2", "Value2"},
			},
			want:    []string{"Name", "Value", "Item1", "Value2"},
			exclude: nil,
		},
		{
			name:    "mismatched columns and widths",
			headers: []string{"Col1", "Col2", "Col3"},
			widths:  []int{10, 20}, // Only 2 widths for 3 columns.
			rows:    [][]string{{"a", "b"}},
			want:    []string{"Col1", "Col2"},
			exclude: []string{"Col3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := output.NewTable(tt.headers, tt.widths)
			for _, row := range tt.rows {
				tbl.AddRow(row)
			}

			result := tbl.Render()
			assertTableContains(t, result, tt.want...)
			assertTableExcludes(t, result, tt.exclude...)
		})
	}
}

func TestMockTableRenderer(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		widths  []int
		rows    [][]string
		want    string
	}{
		{
			name:    "empty mock table",
			headers: []string{"Col1", "Col2"},
			widths:  []int{10, 10},
			rows:    [][]string{},
			want:    "Col1 | Col2\n",
		},
		{
			name:    "mock table with data",
			headers: []string{"Name", "Status"},
			widths:  []int{20, 10},
			rows: [][]string{
				{"Test1", "PASS"},
				{"Test2", "FAIL"},
			},
			want: "Name | Status\nTest1 | PASS\nTest2 | FAIL\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockTable(tt.headers, tt.widths)
			for _, row := range tt.rows {
				mock.AddRow(row)
			}

			got := mock.Render()
			if got != tt.want {
				t.Errorf("MockTableRenderer.Render() = %q, want %q", got, tt.want)
			}

			// Test GetHeaders.
			headers := mock.GetHeaders()
			if len(headers) != len(tt.headers) {
				t.Errorf("GetHeaders() returned %d headers, want %d", len(headers), len(tt.headers))
			}

			// Test GetRows.
			rows := mock.GetRows()
			if len(rows) != len(tt.rows) {
				t.Errorf("GetRows() returned %d rows, want %d", len(rows), len(tt.rows))
			}
		})
	}
}

func TestTableRendererAddRow(t *testing.T) {
	tbl := output.NewTable([]string{"Col1", "Col2"}, []int{10, 10})

	// Add multiple rows.
	tbl.AddRow([]string{"a", "b"})
	tbl.AddRow([]string{"c", "d"})
	tbl.AddRow([]string{"e", "f"})

	result := tbl.Render()

	assertTableContains(t, result, "a", "b", "c", "d", "e", "f")
}
