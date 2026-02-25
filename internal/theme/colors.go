package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// ColorScheme defines semantic colors using AdaptiveColor for automatic
// dark/light terminal adaptation.
type ColorScheme struct {
	// Brand
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor

	// Status
	Success lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Danger  lipgloss.AdaptiveColor
	Info    lipgloss.AdaptiveColor

	// Surface
	SurfaceBg     lipgloss.AdaptiveColor
	SurfaceFg     lipgloss.AdaptiveColor
	SurfaceDim    lipgloss.AdaptiveColor
	SurfaceBorder lipgloss.AdaptiveColor

	// Selection / Header
	SelectionBg lipgloss.AdaptiveColor
	HeaderBg    lipgloss.AdaptiveColor

	// Text
	TextNormal lipgloss.AdaptiveColor
	TextBright lipgloss.AdaptiveColor
	TextMuted  lipgloss.AdaptiveColor
}

// Colors is the global color scheme. Initialized respecting NO_COLOR.
var Colors = NewColorScheme()

// DefaultColorScheme returns the full color palette.
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		// Brand — purple/blue inspired by lazygit/charm aesthetics
		Primary:   lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#7C3AED"},
		Secondary: lipgloss.AdaptiveColor{Dark: "#38BDF8", Light: "#0369A1"},

		// Status — Tailwind-inspired semantic colors (light adjusted for WCAG AA)
		Success: lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#047857"},
		Warning: lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#92400E"},
		Danger:  lipgloss.AdaptiveColor{Dark: "#F87171", Light: "#DC2626"},
		Info:    lipgloss.AdaptiveColor{Dark: "#60A5FA", Light: "#2563EB"},

		// Surface — Catppuccin Mocha (dark) / Slate (light)
		SurfaceBg:     lipgloss.AdaptiveColor{Dark: "#1E1E2E", Light: "#FFFFFF"},
		SurfaceFg:     lipgloss.AdaptiveColor{Dark: "#CDD6F4", Light: "#1E293B"},
		SurfaceDim:    lipgloss.AdaptiveColor{Dark: "#585B70", Light: "#94A3B8"},
		SurfaceBorder: lipgloss.AdaptiveColor{Dark: "#45475A", Light: "#CBD5E1"},

		// Selection / Header
		SelectionBg: lipgloss.AdaptiveColor{Dark: "#313244", Light: "#E2E8F0"},
		HeaderBg:    lipgloss.AdaptiveColor{Dark: "#181825", Light: "#F1F5F9"},

		// Text
		TextNormal: lipgloss.AdaptiveColor{Dark: "#CDD6F4", Light: "#1E293B"},
		TextBright: lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#0F172A"},
		TextMuted:  lipgloss.AdaptiveColor{Dark: "#9399B2", Light: "#475569"},
	}
}

// NoColorScheme returns a ColorScheme with all empty colors for NO_COLOR mode.
func NoColorScheme() ColorScheme {
	return ColorScheme{}
}

// NewColorScheme creates a ColorScheme, respecting NO_COLOR, GROVE_NO_COLOR,
// and GROVE_HIGH_CONTRAST environment variables.
func NewColorScheme() ColorScheme {
	if IsNoColor() {
		return NoColorScheme()
	}
	if IsHighContrast() {
		return HighContrastColorScheme()
	}
	return DefaultColorScheme()
}

// IsNoColor checks if color output should be suppressed.
func IsNoColor() bool {
	_, nc := os.LookupEnv("NO_COLOR")
	_, gnc := os.LookupEnv("GROVE_NO_COLOR")
	return nc || gnc
}

// IsHighContrast checks if high-contrast mode is requested.
func IsHighContrast() bool {
	_, hc := os.LookupEnv("GROVE_HIGH_CONTRAST")
	return hc
}
