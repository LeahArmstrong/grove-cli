package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/LeahArmstrong/grove-cli/internal/plugins"
)

// ComputeDelegateWidthsV2 computes content-adaptive column widths for the V2
// delegate by scanning items for max name/branch rune lengths.
func ComputeDelegateWidthsV2(items []list.Item, width int) WorktreeDelegateV2 {
	maxName, maxBranch := 0, 0
	for _, li := range items {
		item, ok := li.(WorktreeItem)
		if !ok {
			continue
		}
		if n := len([]rune(item.ShortName)); n > maxName {
			maxName = n
		}
		if n := len([]rune(item.Branch)); n > maxBranch {
			maxBranch = n
		}
	}

	d := WorktreeDelegateV2{}

	// Name: use actual max, capped at 40, min 10
	d.NameWidth = maxName
	if d.NameWidth > 40 {
		d.NameWidth = 40
	}
	if d.NameWidth < 10 {
		d.NameWidth = 10
	}

	// Branch: use actual max, capped at 30, min 0; hidden at narrow widths
	if width < 60 {
		d.BranchWidth = 0
	} else {
		d.BranchWidth = maxBranch
		if d.BranchWidth > 30 {
			d.BranchWidth = 30
		}
	}

	return d
}

// worktreeIndicator returns the leading indicator for a worktree item.
// Selected always shows ❯. Non-selected shows status: ● (current/dirty), ✗ (stale), ○ (clean).
func worktreeIndicator(item WorktreeItem, selected bool) string {
	if selected {
		return Styles.ListCursor.Render("❯")
	}
	switch {
	case item.IsCurrent:
		return Styles.StatusSuccess.Render("●")
	case item.IsDirty:
		return Styles.StatusWarning.Render("●")
	case item.IsPrunable:
		return Styles.StatusDanger.Render("✗")
	default:
		return Styles.TextMuted.Render("○")
	}
}

// worktreeStatusTextV2 returns a status string for the item.
func worktreeStatusTextV2(item WorktreeItem) string {
	if item.IsPrunable {
		return Styles.StatusDanger.Render("stale")
	}
	if item.IsDirty {
		count := len(item.DirtyFiles)
		if count > 0 {
			return Styles.StatusWarning.Render(fmt.Sprintf("dirty (%d)", count))
		}
		return Styles.StatusWarning.Render("dirty")
	}
	return Styles.StatusSuccess.Render("clean")
}

// worktreeTmuxBadgeV2 returns a tmux badge if a session exists.
// Uses ⬢ (filled) for attached, ⬡ (unfilled) for detached — consistent
// with the symbol convention in the table-style list delegate.
func worktreeTmuxBadgeV2(item WorktreeItem) string {
	switch item.TmuxStatus {
	case "attached":
		return Styles.TmuxBadge.Render("⬢ tmux")
	case "detached":
		return Styles.TmuxBadge.Render("⬡ tmux")
	default:
		return ""
	}
}

// worktreeContainerBadgeV2 returns a container status badge from plugin statuses.
func worktreeContainerBadgeV2(item WorktreeItem) string {
	for _, s := range item.PluginStatuses {
		if s.Short == "" {
			continue
		}
		switch s.Level {
		case plugins.StatusActive:
			return Styles.ContainerBadgeActive.Render("◆ " + s.Short)
		case plugins.StatusWarning:
			return Styles.ContainerBadgeWarn.Render("◆ " + s.Short)
		case plugins.StatusInfo:
			return Styles.ContainerBadge.Render("◇ " + s.Short)
		default:
			return Styles.TextMuted.Render("◇ " + s.Short)
		}
	}
	return ""
}

// worktreeSyncBadgeV2 returns a compact sync status badge for the list.
func worktreeSyncBadgeV2(item WorktreeItem) string {
	if !item.HasRemote {
		return ""
	}
	if item.AheadCount == 0 && item.BehindCount == 0 {
		return ""
	}
	var parts []string
	if item.AheadCount > 0 {
		parts = append(parts, Styles.StatusSuccess.Render(fmt.Sprintf("↑%d", item.AheadCount)))
	}
	if item.BehindCount > 0 {
		parts = append(parts, Styles.StatusDanger.Render(fmt.Sprintf("↓%d", item.BehindCount)))
	}
	return strings.Join(parts, "")
}

// WorktreeDelegateV2 implements list.ItemDelegate with visual indicators.
type WorktreeDelegateV2 struct {
	NameWidth   int
	BranchWidth int
}

// NewWorktreeDelegateV2 creates a new V2 delegate with default widths.
func NewWorktreeDelegateV2() WorktreeDelegateV2 {
	return WorktreeDelegateV2{NameWidth: 20, BranchWidth: 16}
}

func (d WorktreeDelegateV2) Height() int                             { return 2 }
func (d WorktreeDelegateV2) Spacing() int                            { return 0 }
func (d WorktreeDelegateV2) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d WorktreeDelegateV2) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(WorktreeItem)
	if !ok {
		return
	}

	selected := index == m.Index()
	width := m.Width()

	// Number prefix for quick-switch (1-9)
	numPrefix := "  "
	if index < 9 {
		numPrefix = Styles.TextMuted.Render(fmt.Sprintf("%d ", index+1))
	}

	indicator := worktreeIndicator(item, selected) + " "

	// Name styling
	nameStyle := Styles.NormalItem
	if selected {
		nameStyle = Styles.SelectedItem
	}
	if item.IsCurrent {
		nameStyle = Styles.CurrentItem
		if selected {
			nameStyle = nameStyle.Bold(true)
		}
	}
	name := nameStyle.Render(truncate(item.ShortName, d.NameWidth))

	// Line 1: numPrefix + indicator + name
	line1 := numPrefix + indicator + name

	// Line 2: metadata row (indented to align with name)
	prefixPad := "     " // align under name (2 num + 1 indicator + 1 space + 1)
	var metaParts []string

	if d.BranchWidth > 0 {
		metaParts = append(metaParts, truncate(item.Branch, d.BranchWidth))
	}

	if item.CommitAge != "" {
		metaParts = append(metaParts, compactAge(item.CommitAge))
	}

	metaParts = append(metaParts, cleanAnsi(worktreeStatusTextV2(item)))

	syncBadge := worktreeSyncBadgeV2(item)
	if syncBadge != "" {
		metaParts = append(metaParts, cleanAnsi(syncBadge))
	}
	tmuxBadge := worktreeTmuxBadgeV2(item)
	if tmuxBadge != "" {
		metaParts = append(metaParts, cleanAnsi(tmuxBadge))
	}
	containerBadge := worktreeContainerBadgeV2(item)
	if containerBadge != "" {
		metaParts = append(metaParts, cleanAnsi(containerBadge))
	}

	line2 := prefixPad + Styles.TextMuted.Render(strings.Join(metaParts, " · "))

	// Apply selection background to both lines
	if selected {
		line1 = padToWidth(line1, width)
		line2 = padToWidth(line2, width)
		line1 = Styles.SelectionRow.MaxWidth(width).Render(line1)
		line2 = Styles.SelectionRow.MaxWidth(width).Render(line2)
	} else {
		line1 = lipgloss.NewStyle().MaxWidth(width).Render(line1)
		line2 = lipgloss.NewStyle().MaxWidth(width).Render(line2)
	}

	_, _ = fmt.Fprint(w, lipgloss.JoinVertical(lipgloss.Left, line1, line2))
}

// padToWidth pads a string with spaces to reach the target width.
func padToWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

// cleanAnsi removes ANSI escape sequences from a string for plain-text display.
func cleanAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}
