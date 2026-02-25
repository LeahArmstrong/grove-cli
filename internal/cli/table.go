package cli

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/LeahArmstrong/grove-cli/internal/theme"
)

// Column defines a table column with optional styling.
type Column struct {
	Title    string
	MinWidth int
	MaxWidth int
	// ColorFn optionally colors cell values. Receives the raw cell value,
	// returns the styled string. Only called when color is enabled.
	ColorFn func(value string) string
}

// Table renders a styled columnar table.
type Table struct {
	columns []Column
	rows    [][]string
	w       *Writer
}

// NewTable creates a new table for the given writer and column definitions.
func NewTable(w *Writer, columns ...Column) *Table {
	return &Table{
		columns: columns,
		w:       w,
	}
}

// AddRow appends a data row. Values are positional by column.
func (t *Table) AddRow(values ...string) {
	// Pad to column count
	row := make([]string, len(t.columns))
	for i := range row {
		if i < len(values) {
			row[i] = values[i]
		}
	}
	t.rows = append(t.rows, row)
}

// Render prints the table to the writer.
func (t *Table) Render() {
	if len(t.rows) == 0 {
		return
	}

	cs := theme.Colors
	useColor := t.w.UseColor()

	// Compute column widths from headers and data
	widths := make([]int, len(t.columns))
	for i, col := range t.columns {
		widths[i] = len(col.Title)
		if col.MinWidth > widths[i] {
			widths[i] = col.MinWidth
		}
	}
	for _, row := range t.rows {
		for i, val := range row {
			if i < len(widths) && len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}
	// Apply max width constraints
	for i, col := range t.columns {
		if col.MaxWidth > 0 && widths[i] > col.MaxWidth {
			widths[i] = col.MaxWidth
		}
	}

	// Print header
	headerStyle := lipgloss.NewStyle()
	if useColor {
		headerStyle = headerStyle.Foreground(cs.TextMuted).Bold(true)
	}

	var headerParts []string
	for i, col := range t.columns {
		cell := fmt.Sprintf("%-*s", widths[i], col.Title)
		if useColor {
			cell = headerStyle.Render(cell)
		}
		headerParts = append(headerParts, cell)
	}
	_, _ = fmt.Fprintln(t.w, strings.Join(headerParts, "  "))

	// Print separator
	var sepParts []string
	for _, w := range widths {
		sepParts = append(sepParts, strings.Repeat("─", w))
	}
	sep := strings.Join(sepParts, "──")
	if useColor {
		sep = lipgloss.NewStyle().Foreground(cs.SurfaceDim).Render(sep)
	}
	_, _ = fmt.Fprintln(t.w, sep)

	// Print rows
	for _, row := range t.rows {
		var parts []string
		for i, val := range row {
			if i >= len(t.columns) {
				break
			}

			// Truncate if needed
			display := val
			if t.columns[i].MaxWidth > 0 && len(display) > widths[i] {
				display = display[:widths[i]-1] + "…"
			}

			// Apply color function if available.
			// ColorFn receives the original value for correct status matching;
			// padding is based on the display (possibly truncated) length.
			// If truncation occurs on a colored column, the rendered text may
			// slightly exceed the column width — an acceptable tradeoff since
			// MinWidth should be set to accommodate expected values.
			if useColor && t.columns[i].ColorFn != nil {
				colored := t.columns[i].ColorFn(val)
				padding := widths[i] - len(display)
				if padding > 0 {
					colored += strings.Repeat(" ", padding)
				}
				parts = append(parts, colored)
			} else {
				parts = append(parts, fmt.Sprintf("%-*s", widths[i], display))
			}
		}
		_, _ = fmt.Fprintln(t.w, strings.Join(parts, "  "))
	}
}
