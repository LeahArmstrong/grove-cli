package commands

import (
	"testing"
)

func TestNewCmd(t *testing.T) {
	if newCmd == nil {
		t.Fatal("newCmd is nil")
	}

	if newCmd.Use != "new <name>" {
		t.Errorf("newCmd.Use = %v, want 'new <name>'", newCmd.Use)
	}

	if newCmd.Short == "" {
		t.Error("newCmd.Short is empty")
	}

	if newCmd.RunE == nil {
		t.Error("newCmd.RunE is nil")
	}
}

func TestNewCmd_Flags(t *testing.T) {
	flags := newCmd.Flags()

	tests := []string{"json", "branch", "from", "mirror", "no-docker"}
	for _, name := range tests {
		if flags.Lookup(name) == nil {
			t.Errorf("expected --%s flag to exist", name)
		}
	}
}

func TestNewCmd_BranchFlag(t *testing.T) {
	flag := newCmd.Flags().Lookup("branch")
	if flag == nil {
		t.Fatal("newCmd missing --branch flag")
	}
	if flag.Shorthand != "b" {
		t.Errorf("--branch shorthand = %q, want %q", flag.Shorthand, "b")
	}
	if flag.DefValue != "" {
		t.Errorf("--branch should default to empty, got %q", flag.DefValue)
	}
}

func TestNewCmd_FromFlag(t *testing.T) {
	flag := newCmd.Flags().Lookup("from")
	if flag == nil {
		t.Fatal("newCmd missing --from flag")
	}
	if flag.Shorthand != "f" {
		t.Errorf("--from shorthand = %q, want %q", flag.Shorthand, "f")
	}
	if flag.DefValue != "" {
		t.Errorf("--from should default to empty, got %q", flag.DefValue)
	}
}

func TestNewCmd_RequiresExactlyOneArg(t *testing.T) {
	if newCmd.Args == nil {
		t.Error("newCmd.Args should not be nil — should require exactly one argument")
	}
}

func TestNewCmd_HasSpawnAlias(t *testing.T) {
	found := false
	for _, alias := range newCmd.Aliases {
		if alias == "spawn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("newCmd should have 'spawn' alias")
	}
}
