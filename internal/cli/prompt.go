package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"github.com/LeahArmstrong/grove-cli/internal/theme"
)

// IsInteractive returns true if stdin is connected to a terminal.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Confirm asks the user a yes/no question using a Huh form when interactive,
// falling back to simple stdin prompt otherwise.
func Confirm(question string, defaultYes bool) (bool, error) {
	if !IsInteractive() {
		return false, fmt.Errorf("not an interactive terminal")
	}

	result := defaultYes

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(question).
				Affirmative("Yes").
				Negative("No").
				Value(&result),
		),
	).WithTheme(huh.ThemeCharm()).
		WithAccessible(theme.IsHighContrast())

	if err := form.Run(); err != nil {
		return false, err
	}

	return result, nil
}

// ConfirmWithDetails shows information before asking for confirmation.
func ConfirmWithDetails(w *Writer, header string, details []string, question string, defaultYes bool) (bool, error) {
	if !IsInteractive() {
		return false, fmt.Errorf("not an interactive terminal")
	}

	// Print styled header and details
	Bold(w, "%s", header)
	for _, detail := range details {
		_, _ = fmt.Fprintf(w, "  %s\n", detail)
	}
	_, _ = fmt.Fprintln(w)

	return Confirm(question, defaultYes)
}

// Choose presents a selection menu and returns the chosen option.
func Choose(title string, options []string) (string, error) {
	if !IsInteractive() {
		return "", fmt.Errorf("not an interactive terminal")
	}

	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}

	var result string

	selectOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		selectOptions[i] = huh.NewOption(opt, opt)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(selectOptions...).
				Value(&result),
		),
	).WithTheme(huh.ThemeCharm()).
		WithAccessible(theme.IsHighContrast())

	if err := form.Run(); err != nil {
		return "", err
	}

	return result, nil
}
