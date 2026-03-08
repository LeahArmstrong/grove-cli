package commands

import (
	"strings"
	"testing"
)

func TestRenameCmd(t *testing.T) {
	if renameCmd == nil {
		t.Fatal("renameCmd is nil")
	}

	if renameCmd.Use != "rename <old> <new>" {
		t.Errorf("renameCmd.Use = %v, want 'rename <old> <new>'", renameCmd.Use)
	}

	if renameCmd.RunE == nil {
		t.Error("renameCmd.RunE is nil")
	}
}

func TestRenameCmdArgs(t *testing.T) {
	// Should require exactly 2 args
	if renameCmd.Args == nil {
		t.Fatal("renameCmd.Args is nil")
	}

	// Verify it errors with wrong number of args
	err := renameCmd.Args(renameCmd, []string{"only-one"})
	if err == nil {
		t.Error("should error with only 1 arg")
	}

	err = renameCmd.Args(renameCmd, []string{"one", "two"})
	if err != nil {
		t.Errorf("should accept 2 args, got error: %v", err)
	}

	err = renameCmd.Args(renameCmd, []string{"one", "two", "three"})
	if err == nil {
		t.Error("should error with 3 args")
	}
}

func TestRenameCmdHelp(t *testing.T) {
	long := renameCmd.Long
	if long == "" {
		t.Fatal("renameCmd.Long is empty")
	}

	required := []struct {
		label string
		text  string
	}{
		{"directory", "directory"},
		{"tmux session", "tmux session"},
		{"protected", "Protected"},
	}

	for _, tt := range required {
		if !strings.Contains(long, tt.text) {
			t.Errorf("renameCmd.Long should mention %s (missing %q)", tt.label, tt.text)
		}
	}
}

func TestRenameCmdRegistered(t *testing.T) {
	// Verify the rename command is registered on the root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "rename" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rename command not registered on root")
	}
}
