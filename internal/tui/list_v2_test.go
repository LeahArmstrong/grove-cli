package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
)

func TestWorktreeDelegateV2_Indicators(t *testing.T) {
	tests := []struct {
		name          string
		item          WorktreeItem
		selected      bool
		wantIndicator string
	}{
		{
			name:          "Current worktree shows green dot",
			item:          WorktreeItem{IsCurrent: true, ShortName: "main", Branch: "main"},
			selected:      false,
			wantIndicator: "●",
		},
		{
			name:          "Dirty worktree shows yellow dot",
			item:          WorktreeItem{IsDirty: true, ShortName: "feature", Branch: "feat"},
			selected:      false,
			wantIndicator: "●",
		},
		{
			name:          "Selected shows cursor",
			item:          WorktreeItem{ShortName: "test", Branch: "test"},
			selected:      true,
			wantIndicator: "❯",
		},
		{
			name:          "Normal shows circle",
			item:          WorktreeItem{ShortName: "other", Branch: "other"},
			selected:      false,
			wantIndicator: "○",
		},
		{
			name:          "Current and selected shows cursor",
			item:          WorktreeItem{IsCurrent: true, ShortName: "main", Branch: "main"},
			selected:      true,
			wantIndicator: "❯",
		},
		{
			name:          "Dirty and selected shows cursor",
			item:          WorktreeItem{IsDirty: true, ShortName: "feature", Branch: "feat"},
			selected:      true,
			wantIndicator: "❯",
		},
		{
			name:          "Stale worktree shows stale indicator",
			item:          WorktreeItem{IsPrunable: true, ShortName: "old", Branch: "old"},
			selected:      false,
			wantIndicator: "✗",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicator := worktreeIndicator(tt.item, tt.selected)
			// The indicator string contains styling, but must include the expected rune
			if !strings.Contains(indicator, tt.wantIndicator) {
				t.Errorf("worktreeIndicator() = %q, want to contain %q", indicator, tt.wantIndicator)
			}
		})
	}
}

func TestWorktreeDelegateV2_StatusText(t *testing.T) {
	tests := []struct {
		name string
		item WorktreeItem
		want string
	}{
		{"Clean", WorktreeItem{}, "clean"},
		{"Dirty with files", WorktreeItem{IsDirty: true, DirtyFiles: []string{"a.go", "b.go"}}, "dirty"},
		{"Stale", WorktreeItem{IsPrunable: true}, "stale"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := worktreeStatusTextV2(tt.item)
			if !strings.Contains(text, tt.want) {
				t.Errorf("worktreeStatusTextV2() = %q, want to contain %q", text, tt.want)
			}
		})
	}
}

func TestWorktreeDelegateV2_TmuxBadge(t *testing.T) {
	tests := []struct {
		name string
		item WorktreeItem
		want string
	}{
		{"No tmux", WorktreeItem{TmuxStatus: "none"}, ""},
		{"Attached", WorktreeItem{TmuxStatus: "attached"}, "tmux"},
		{"Detached", WorktreeItem{TmuxStatus: "detached"}, "tmux"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := worktreeTmuxBadgeV2(tt.item)
			if tt.want == "" && badge != "" {
				t.Errorf("worktreeTmuxBadgeV2() = %q, want empty", badge)
			}
			if tt.want != "" && !strings.Contains(badge, tt.want) {
				t.Errorf("worktreeTmuxBadgeV2() = %q, want to contain %q", badge, tt.want)
			}
		})
	}
}

func TestComputeDelegateWidthsV2_ContentAdaptive(t *testing.T) {
	items := []list.Item{
		WorktreeItem{ShortName: "pr-13093-fix-disabled", Branch: "fix/disable-feature"},
		WorktreeItem{ShortName: "main", Branch: "main"},
	}

	tests := []struct {
		name        string
		width       int
		wantNameMin int
		wantNameMax int
	}{
		{"Narrow (80)", 80, 10, 40},
		{"Medium (100)", 100, 10, 40},
		{"Wide (140)", 140, 10, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ComputeDelegateWidthsV2(items, tt.width)
			if d.NameWidth < tt.wantNameMin || d.NameWidth > tt.wantNameMax {
				t.Errorf("ComputeDelegateWidthsV2(%d).NameWidth = %d, want in [%d, %d]",
					tt.width, d.NameWidth, tt.wantNameMin, tt.wantNameMax)
			}
		})
	}
}

func TestWorktreeSyncBadgeV2(t *testing.T) {
	tests := []struct {
		name string
		item WorktreeItem
		want string
	}{
		{"No remote", WorktreeItem{HasRemote: false}, ""},
		{"Synced", WorktreeItem{HasRemote: true}, ""},
		{"Ahead", WorktreeItem{HasRemote: true, AheadCount: 3}, "↑3"},
		{"Behind", WorktreeItem{HasRemote: true, BehindCount: 2}, "↓2"},
		{"Diverged", WorktreeItem{HasRemote: true, AheadCount: 1, BehindCount: 4}, "↑1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := worktreeSyncBadgeV2(tt.item)
			if tt.want == "" && badge != "" {
				t.Errorf("worktreeSyncBadgeV2() = %q, want empty", badge)
			}
			if tt.want != "" && !strings.Contains(badge, tt.want) {
				t.Errorf("worktreeSyncBadgeV2() = %q, want to contain %q", badge, tt.want)
			}
		})
	}
}

func TestWorktreeDelegateV2_Height(t *testing.T) {
	d := NewWorktreeDelegateV2()
	if d.Height() != 2 {
		t.Errorf("Height() = %d, want 2", d.Height())
	}
}

func TestWorktreeDelegateV2_Spacing(t *testing.T) {
	d := NewWorktreeDelegateV2()
	if d.Spacing() != 0 {
		t.Errorf("Spacing() = %d, want 0", d.Spacing())
	}
}

func TestWorktreeDelegateV2_Update(t *testing.T) {
	d := NewWorktreeDelegateV2()
	l := list.New(nil, d, 80, 20)
	cmd := d.Update(nil, &l)
	if cmd != nil {
		t.Errorf("Update() returned non-nil cmd")
	}
}
