package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/LeahArmstrong/grove-cli/internal/theme"
)

// RelativeLuminance delegates to the shared theme package.
var RelativeLuminance = theme.RelativeLuminance

// ContrastRatio delegates to the shared theme package.
var ContrastRatio = theme.ContrastRatio

// isHighContrast delegates to the shared theme package.
func isHighContrast() bool {
	return theme.IsHighContrast()
}

// NewAccessibleCreateNameForm creates a Huh form with accessible mode enabled.
func NewAccessibleCreateNameForm(nameValue *string, projectName string, existingItems []WorktreeItem) *huh.Form {
	description := "Worktree name"
	if projectName != "" {
		description = fmt.Sprintf("Will create: %s-<name>", projectName)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Worktree Name").
				Description(description).
				Placeholder("feature-name").
				Validate(createNameValidator(existingItems, "")).
				Value(nameValue),
		),
	).WithTheme(huh.ThemeCharm()).WithAccessible(true)

	return form
}
