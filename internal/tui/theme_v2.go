package tui

import (
	"image/color"
	"os"
	"strconv"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

// adaptiveColor picks between a dark and light hex color.
// In v2 lipgloss, AdaptiveColor no longer exists.
// We default to dark mode and allow override via GROVE_LIGHT_MODE env var.
// Falls back to COLORFGBG heuristic if GROVE_LIGHT_MODE is not set.
// Colors are resolved once at package init time (not re-evaluated on theme changes).
func adaptiveColor(dark, light string) color.Color {
	if isLightMode() {
		return lipgloss.Color(light)
	}
	return lipgloss.Color(dark)
}

// isLightMode checks if the terminal is in light mode.
// Priority: GROVE_LIGHT_MODE env var (explicit) > COLORFGBG heuristic > dark default.
func isLightMode() bool {
	if val, ok := os.LookupEnv("GROVE_LIGHT_MODE"); ok {
		return val != "0" && val != "false" && val != "no"
	}
	// COLORFGBG is "fg;bg" — bg >= 8 typically means light background.
	if colorfgbg, ok := os.LookupEnv("COLORFGBG"); ok {
		parts := strings.Split(colorfgbg, ";")
		if len(parts) >= 2 {
			if bg, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				return bg >= 8
			}
		}
	}
	return false
}

// ColorScheme defines semantic colors for the TUI.
type ColorScheme struct {
	// Brand
	Primary   color.Color
	Secondary color.Color

	// Status
	Success color.Color
	Warning color.Color
	Danger  color.Color
	Info    color.Color

	// Surface
	SurfaceBg     color.Color
	SurfaceFg     color.Color
	SurfaceDim    color.Color
	SurfaceBorder color.Color

	// Selection / Header
	SelectionBg color.Color
	HeaderBg    color.Color

	// Text
	TextNormal color.Color
	TextBright color.Color
	TextMuted  color.Color
}

// Colors is the global color scheme. Initialized respecting NO_COLOR.
var Colors = NewColorScheme()

// defaultColorScheme returns the full color palette.
func defaultColorScheme() ColorScheme {
	return ColorScheme{
		// Brand — purple/blue inspired by lazygit/charm aesthetics
		Primary:   adaptiveColor("#A78BFA", "#7C3AED"),
		Secondary: adaptiveColor("#38BDF8", "#0369A1"),

		// Status — Tailwind-inspired semantic colors (light adjusted for WCAG AA)
		Success: adaptiveColor("#34D399", "#047857"),
		Warning: adaptiveColor("#FBBF24", "#92400E"),
		Danger:  adaptiveColor("#F87171", "#DC2626"),
		Info:    adaptiveColor("#60A5FA", "#2563EB"),

		// Surface — Catppuccin Mocha (dark) / Slate (light)
		SurfaceBg:     adaptiveColor("#1E1E2E", "#FFFFFF"),
		SurfaceFg:     adaptiveColor("#CDD6F4", "#1E293B"),
		SurfaceDim:    adaptiveColor("#585B70", "#94A3B8"),
		SurfaceBorder: adaptiveColor("#45475A", "#CBD5E1"),

		// Selection / Header
		SelectionBg: adaptiveColor("#313244", "#E2E8F0"),
		HeaderBg:    adaptiveColor("#181825", "#F1F5F9"),

		// Text
		TextNormal: adaptiveColor("#CDD6F4", "#1E293B"),
		TextBright: adaptiveColor("#FFFFFF", "#0F172A"),
		TextMuted:  adaptiveColor("#A6ADC8", "#475569"),
	}
}

// noColorScheme returns a ColorScheme with all nil colors for NO_COLOR mode.
func noColorScheme() ColorScheme {
	return ColorScheme{}
}

// NewColorScheme creates a ColorScheme, respecting NO_COLOR, GROVE_NO_COLOR,
// and GROVE_HIGH_CONTRAST environment variables.
func NewColorScheme() ColorScheme {
	if isNoColor() {
		return noColorScheme()
	}
	if isHighContrast() {
		return highContrastColorScheme()
	}
	return defaultColorScheme()
}

// isNoColor checks if color output should be suppressed.
func isNoColor() bool {
	_, nc := os.LookupEnv("NO_COLOR")
	_, gnc := os.LookupEnv("GROVE_NO_COLOR")
	return nc || gnc
}

// StyleSet holds pre-composed lipgloss styles built from a ColorScheme.
type StyleSet struct {
	// Borders
	RoundedBorder lipgloss.Style
	DetailBorder  lipgloss.Style

	// Text
	Title      lipgloss.Style
	TextNormal lipgloss.Style
	TextBright lipgloss.Style
	TextMuted  lipgloss.Style

	// Status text
	StatusSuccess lipgloss.Style
	StatusWarning lipgloss.Style
	StatusDanger  lipgloss.Style
	StatusInfo    lipgloss.Style

	// Layout
	Header lipgloss.Style
	Footer lipgloss.Style

	// List items
	SelectedItem  lipgloss.Style
	NormalItem    lipgloss.Style
	CurrentItem   lipgloss.Style
	DimmedItem    lipgloss.Style
	ListCursor    lipgloss.Style
	ListCursorDim lipgloss.Style

	// List selection
	SelectionRow lipgloss.Style

	// Status badges
	StatusClean          lipgloss.Style
	StatusDirty          lipgloss.Style
	StatusStale          lipgloss.Style
	TmuxBadge            lipgloss.Style
	TmuxBadgeActive      lipgloss.Style
	EnvBadge             lipgloss.Style
	ContainerBadge       lipgloss.Style
	ContainerBadgeActive lipgloss.Style
	ContainerBadgeWarn   lipgloss.Style

	// Detail panel
	DetailTitle   lipgloss.Style
	DetailLabel   lipgloss.Style
	DetailValue   lipgloss.Style
	DetailDim     lipgloss.Style
	DetailFile    lipgloss.Style
	DetailFileAdd lipgloss.Style
	DetailFileMod lipgloss.Style
	DetailFileDel lipgloss.Style

	// Layout
	HeaderBar lipgloss.Style

	// Overlay / dialogs
	OverlayBorder        lipgloss.Style
	OverlayBorderDanger  lipgloss.Style
	OverlayBorderSuccess lipgloss.Style
	OverlayBorderInfo    lipgloss.Style
	OverlayTitle         lipgloss.Style
	OverlayPrompt        lipgloss.Style
	WarningText          lipgloss.Style
	ErrorText            lipgloss.Style
	SuccessText          lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
	HelpSep  lipgloss.Style

	// Input
	InputBorder lipgloss.Style
	InputText   lipgloss.Style
}

// Styles is the global StyleSet initialized from Colors.
var Styles = NewStyleSet(Colors)

// NewStyleSet creates a StyleSet from a ColorScheme.
func NewStyleSet(cs ColorScheme) StyleSet {
	return StyleSet{
		// Borders
		RoundedBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.SurfaceBorder),
		DetailBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.SurfaceBorder),

		// Text
		Title:      lipgloss.NewStyle().Bold(true).Foreground(cs.TextBright),
		TextNormal: lipgloss.NewStyle().Foreground(cs.TextNormal),
		TextBright: lipgloss.NewStyle().Bold(true).Foreground(cs.TextBright),
		TextMuted:  lipgloss.NewStyle().Foreground(cs.TextMuted),

		// Status text
		StatusSuccess: lipgloss.NewStyle().Foreground(cs.Success),
		StatusWarning: lipgloss.NewStyle().Foreground(cs.Warning),
		StatusDanger:  lipgloss.NewStyle().Foreground(cs.Danger),
		StatusInfo:    lipgloss.NewStyle().Foreground(cs.Info),

		// List selection
		SelectionRow: lipgloss.NewStyle().Background(cs.SelectionBg),

		// Layout
		Header: lipgloss.NewStyle().Bold(true).Foreground(cs.Primary),
		Footer: lipgloss.NewStyle().Foreground(cs.TextMuted),
		HeaderBar: lipgloss.NewStyle().
			Background(cs.HeaderBg).
			Bold(true).
			Padding(0, 1),

		// List items
		SelectedItem:  lipgloss.NewStyle().Bold(true).Foreground(cs.TextBright),
		NormalItem:    lipgloss.NewStyle().Foreground(cs.TextNormal),
		CurrentItem:   lipgloss.NewStyle().Foreground(cs.Secondary),
		DimmedItem:    lipgloss.NewStyle().Foreground(adaptiveColor("#7F849C", "#64748B")),
		ListCursor:    lipgloss.NewStyle().Foreground(cs.Primary),
		ListCursorDim: lipgloss.NewStyle(),

		// Status badges
		StatusClean:          lipgloss.NewStyle().Foreground(cs.Success),
		StatusDirty:          lipgloss.NewStyle().Foreground(cs.Warning),
		StatusStale:          lipgloss.NewStyle().Foreground(cs.Danger),
		TmuxBadge:            lipgloss.NewStyle().Foreground(cs.Primary),
		TmuxBadgeActive:      lipgloss.NewStyle().Foreground(cs.Primary),
		EnvBadge:             lipgloss.NewStyle().Foreground(cs.Info),
		ContainerBadge:       lipgloss.NewStyle().Foreground(cs.Info),
		ContainerBadgeActive: lipgloss.NewStyle().Foreground(cs.Secondary),
		ContainerBadgeWarn:   lipgloss.NewStyle().Foreground(cs.Warning),

		// Detail panel
		DetailTitle:   lipgloss.NewStyle().Bold(true).Foreground(cs.TextBright),
		DetailLabel:   lipgloss.NewStyle().Foreground(cs.TextMuted),
		DetailValue:   lipgloss.NewStyle().Foreground(cs.TextNormal),
		DetailDim:     lipgloss.NewStyle().Foreground(cs.TextMuted),
		DetailFile:    lipgloss.NewStyle().Foreground(cs.TextNormal),
		DetailFileAdd: lipgloss.NewStyle().Foreground(cs.Success),
		DetailFileMod: lipgloss.NewStyle().Foreground(cs.Warning),
		DetailFileDel: lipgloss.NewStyle().Foreground(cs.Danger),

		// Overlay
		OverlayBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Primary).
			Padding(1, 2),
		OverlayBorderDanger: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Danger).
			Padding(1, 2),
		OverlayBorderSuccess: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Success).
			Padding(1, 2),
		OverlayBorderInfo: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Info).
			Padding(1, 2),
		OverlayTitle:  lipgloss.NewStyle().Bold(true).Foreground(cs.Primary),
		OverlayPrompt: lipgloss.NewStyle().Foreground(cs.TextNormal),
		WarningText:   lipgloss.NewStyle().Foreground(cs.Warning),
		ErrorText:     lipgloss.NewStyle().Foreground(cs.Danger),
		SuccessText:   lipgloss.NewStyle().Foreground(cs.Success),

		// Help
		HelpKey:  lipgloss.NewStyle().Foreground(cs.Primary).Bold(true),
		HelpDesc: lipgloss.NewStyle().Foreground(cs.TextMuted),
		HelpSep:  lipgloss.NewStyle().Foreground(cs.TextMuted).SetString(" · "),

		// Input
		InputBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cs.Primary),
		InputText: lipgloss.NewStyle().Foreground(cs.TextNormal),
	}
}
