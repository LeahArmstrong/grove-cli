package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestSuccess_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Success(w, "operation completed")
	if got := buf.String(); got != "✓ operation completed\n" {
		t.Errorf("got %q, want %q", got, "✓ operation completed\n")
	}
}

func TestWarning_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Warning(w, "something odd: %s", "details")
	if got := buf.String(); !strings.Contains(got, "⚠ something odd: details") {
		t.Errorf("got %q, want warning message", got)
	}
}

func TestError_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Error(w, "failed: %v", "reason")
	if got := buf.String(); !strings.Contains(got, "✗ failed: reason") {
		t.Errorf("got %q, want error message", got)
	}
}

func TestInfo_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Info(w, "note: %s", "info")
	if got := buf.String(); !strings.Contains(got, "ℹ note: info") {
		t.Errorf("got %q, want info message", got)
	}
}

func TestHeader_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Header(w, "grove doctor")
	got := buf.String()
	if !strings.Contains(got, "grove doctor\n") {
		t.Errorf("missing title in %q", got)
	}
	if !strings.Contains(got, "━") {
		t.Errorf("missing separator in %q", got)
	}
}

func TestStep_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Step(w, "doing stuff")
	if got := buf.String(); got != "→ doing stuff\n" {
		t.Errorf("got %q", got)
	}
}

func TestStatusText_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	w := NewWriter(&bytes.Buffer{}, false)
	got := StatusText(w, "clean", "clean")
	if got != "clean" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestLabel_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Label(w, "Path:", "/some/path")
	if got := buf.String(); !strings.Contains(got, "Path: /some/path") {
		t.Errorf("got %q", got)
	}
}

func TestFaint_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Faint(w, "muted text")
	if got := buf.String(); got != "muted text\n" {
		t.Errorf("got %q, want %q", got, "muted text\n")
	}
}

func TestBold_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	Bold(w, "bold text")
	if got := buf.String(); got != "bold text\n" {
		t.Errorf("got %q, want %q", got, "bold text\n")
	}
}

func TestAccent_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	w := NewWriter(&bytes.Buffer{}, false)
	got := Accent(w, "accented")
	if got != "accented" {
		t.Errorf("expected plain text, got %q", got)
	}
}

func TestStatusText_AllCategories(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	w := NewWriter(&bytes.Buffer{}, false)

	tests := []struct {
		status string
		text   string
	}{
		// success
		{"clean", "clean"},
		{"ok", "ok"},
		{"active", "active"},
		{"attached", "attached"},
		// warning
		{"dirty", "dirty"},
		{"warning", "warning"},
		{"detached", "detached"},
		// danger
		{"stale", "stale"},
		{"error", "error"},
		{"fail", "fail"},
		// muted
		{"info", "info"},
		{"none", "none"},
		// unknown — returns plain text
		{"unknown-status", "some text"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusText(w, tt.status, tt.text)
			if got != tt.text {
				t.Errorf("StatusText(%q, %q) = %q, want plain %q", tt.status, tt.text, got, tt.text)
			}
		})
	}
}
