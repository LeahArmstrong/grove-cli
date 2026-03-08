package commands

import (
	"testing"
)

func TestResolveDirtyAction(t *testing.T) {
	tests := []struct {
		name          string
		dirtyHandling string
		isDirty       bool
		isPeek        bool
		isInteractive bool
		want          DirtyAction
	}{
		// Peek always allows, regardless of dirty state or config
		{
			name:          "peek skips dirty handling even when dirty",
			dirtyHandling: "refuse",
			isDirty:       true,
			isPeek:        true,
			isInteractive: true,
			want:          actionAllow,
		},
		{
			name:          "peek skips dirty handling when clean",
			dirtyHandling: "refuse",
			isDirty:       false,
			isPeek:        true,
			isInteractive: false,
			want:          actionAllow,
		},

		// Clean worktree always allows, regardless of config
		{
			name:          "refuse allows clean worktree",
			dirtyHandling: "refuse",
			isDirty:       false,
			isPeek:        false,
			isInteractive: true,
			want:          actionAllow,
		},
		{
			name:          "auto-stash allows clean worktree",
			dirtyHandling: "auto-stash",
			isDirty:       false,
			isPeek:        false,
			isInteractive: true,
			want:          actionAllow,
		},
		{
			name:          "prompt allows clean worktree",
			dirtyHandling: "prompt",
			isDirty:       false,
			isPeek:        false,
			isInteractive: true,
			want:          actionAllow,
		},

		// Refuse mode with dirty worktree
		{
			name:          "refuse blocks dirty worktree",
			dirtyHandling: "refuse",
			isDirty:       true,
			isPeek:        false,
			isInteractive: true,
			want:          actionRefuse,
		},
		{
			name:          "refuse blocks dirty worktree non-interactive",
			dirtyHandling: "refuse",
			isDirty:       true,
			isPeek:        false,
			isInteractive: false,
			want:          actionRefuse,
		},

		// Auto-stash mode with dirty worktree
		{
			name:          "auto-stash stashes dirty worktree",
			dirtyHandling: "auto-stash",
			isDirty:       true,
			isPeek:        false,
			isInteractive: true,
			want:          actionStash,
		},
		{
			name:          "auto-stash stashes dirty worktree non-interactive",
			dirtyHandling: "auto-stash",
			isDirty:       true,
			isPeek:        false,
			isInteractive: false,
			want:          actionStash,
		},

		// Prompt mode
		{
			name:          "prompt prompts on dirty interactive",
			dirtyHandling: "prompt",
			isDirty:       true,
			isPeek:        false,
			isInteractive: true,
			want:          actionPrompt,
		},
		{
			name:          "prompt refuses on dirty non-interactive",
			dirtyHandling: "prompt",
			isDirty:       true,
			isPeek:        false,
			isInteractive: false,
			want:          actionRefuse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveDirtyAction(tt.dirtyHandling, tt.isDirty, tt.isPeek, tt.isInteractive)
			if got != tt.want {
				t.Errorf("resolveDirtyAction(%q, dirty=%v, peek=%v, interactive=%v) = %v, want %v",
					tt.dirtyHandling, tt.isDirty, tt.isPeek, tt.isInteractive, got, tt.want)
			}
		})
	}
}
