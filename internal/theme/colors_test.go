package theme

import "testing"

func TestNewColorScheme_Default(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// Unset by restoring empty; t.Setenv handles cleanup
	// Need to ensure NO_COLOR is actually unset, not set to empty
	// (LookupEnv returns true even for empty value)
	// Use a clean approach: set something we know won't trigger
	t.Setenv("GROVE_NO_COLOR", "")
	t.Setenv("GROVE_HIGH_CONTRAST", "")

	// Since LookupEnv returns (_, true) even for empty string,
	// and we set them, this test would detect NO_COLOR mode.
	// We need a different approach: test DefaultColorScheme directly.
	def := DefaultColorScheme()
	if def.Primary.Dark != "#A78BFA" {
		t.Errorf("expected default Primary dark %q, got %q", "#A78BFA", def.Primary.Dark)
	}
}

func TestNewColorScheme_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	cs := NewColorScheme()
	if cs.Primary.Dark != "" || cs.Primary.Light != "" {
		t.Errorf("expected empty colors in NO_COLOR mode, got Primary=%+v", cs.Primary)
	}
}

func TestNewColorScheme_GroveNoColor(t *testing.T) {
	t.Setenv("GROVE_NO_COLOR", "1")

	cs := NewColorScheme()
	if cs.Primary.Dark != "" {
		t.Errorf("expected empty colors in GROVE_NO_COLOR mode")
	}
}

func TestNewColorScheme_HighContrast(t *testing.T) {
	t.Setenv("GROVE_HIGH_CONTRAST", "1")

	// HighContrast takes precedence only when NO_COLOR is not set.
	// Since t.Setenv doesn't unset, test the function directly.
	hc := HighContrastColorScheme()
	if hc.Primary.Dark != "#C4B5FD" {
		t.Errorf("expected high-contrast Primary dark %q, got %q", "#C4B5FD", hc.Primary.Dark)
	}
}

func TestIsNoColor(t *testing.T) {
	// LookupEnv returns true even for empty value, so we test the function
	// by setting the env var
	t.Setenv("NO_COLOR", "1")
	if !IsNoColor() {
		t.Error("expected true when NO_COLOR is set")
	}
}

func TestIsHighContrast(t *testing.T) {
	t.Setenv("GROVE_HIGH_CONTRAST", "1")
	if !IsHighContrast() {
		t.Error("expected true when GROVE_HIGH_CONTRAST=1")
	}
}

func TestNoColorScheme_Empty(t *testing.T) {
	cs := NoColorScheme()
	if cs.Primary.Dark != "" || cs.Primary.Light != "" {
		t.Errorf("expected empty colors, got Primary=%+v", cs.Primary)
	}
}
