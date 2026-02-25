package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable_Basic(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)

	tbl := NewTable(w,
		Column{Title: "NAME", MinWidth: 10},
		Column{Title: "STATUS", MinWidth: 8},
	)
	tbl.AddRow("main", "clean")
	tbl.AddRow("feature", "dirty")
	tbl.Render()

	got := buf.String()
	if !strings.Contains(got, "NAME") {
		t.Error("missing NAME header")
	}
	if !strings.Contains(got, "STATUS") {
		t.Error("missing STATUS header")
	}
	if !strings.Contains(got, "main") {
		t.Error("missing 'main' row")
	}
	if !strings.Contains(got, "feature") {
		t.Error("missing 'feature' row")
	}
	if !strings.Contains(got, "─") {
		t.Error("missing separator")
	}
}

func TestTable_Empty(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)

	tbl := NewTable(w, Column{Title: "NAME"})
	tbl.Render()

	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty table, got %q", buf.String())
	}
}

func TestTable_MaxWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)

	tbl := NewTable(w,
		Column{Title: "NAME", MaxWidth: 5},
	)
	tbl.AddRow("very-long-name")
	tbl.Render()

	got := buf.String()
	lines := strings.Split(got, "\n")
	// Check the data line (skip header and separator)
	if len(lines) >= 3 {
		dataLine := lines[2]
		if len(strings.TrimSpace(dataLine)) > 10 {
			// Should be truncated
			if !strings.Contains(dataLine, "…") {
				t.Errorf("expected truncation with ellipsis, got %q", dataLine)
			}
		}
	}
}
