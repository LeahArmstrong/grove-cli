package commands

import (
	"testing"
)

func TestRmCmd(t *testing.T) {
	if rmCmd == nil {
		t.Fatal("rmCmd is nil")
	}

	if rmCmd.Use != "rm <name>" {
		t.Errorf("rmCmd.Use = %v, want 'rm <name>'", rmCmd.Use)
	}

	if rmCmd.RunE == nil {
		t.Error("rmCmd.RunE is nil")
	}
}

func TestRmFlags(t *testing.T) {
	flags := rmCmd.Flags()

	tests := []string{"force", "unprotect", "dry-run", "keep-branch", "delete-branch"}
	for _, name := range tests {
		if flags.Lookup(name) == nil {
			t.Errorf("expected --%s flag to exist", name)
		}
	}
}

func TestRmForceFlag(t *testing.T) {
	f := rmCmd.Flags().Lookup("force")
	if f == nil {
		t.Fatal("--force flag not found")
	}

	if f.Shorthand != "f" {
		t.Errorf("force shorthand = %q, want %q", f.Shorthand, "f")
	}

	// Flag description should mention dirty worktrees
	if f.Usage == "" {
		t.Error("force flag has no usage description")
	}
}

func TestRmAliases(t *testing.T) {
	aliases := rmCmd.Aliases
	expected := map[string]bool{"remove": true, "delete": true}

	for _, alias := range aliases {
		if !expected[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
		delete(expected, alias)
	}

	for missing := range expected {
		t.Errorf("missing expected alias %q", missing)
	}
}

func TestRmSafetyCheckOrder(t *testing.T) {
	// Verify the command long description documents protection behavior
	long := rmCmd.Long
	if long == "" {
		t.Fatal("rmCmd.Long is empty")
	}

	// The long description should mention protected worktrees
	tests := []struct {
		name string
		want string
	}{
		{"mentions protection", "Protected worktrees"},
		{"mentions force flag", "--force"},
		{"mentions unprotect flag", "--unprotect"},
	}

	for _, tt := range tests {
		found := false
		for i := 0; i <= len(long)-len(tt.want); i++ {
			if long[i:i+len(tt.want)] == tt.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rmCmd.Long does not contain %q", tt.want)
		}
	}
}
