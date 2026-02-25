package tui

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// hexToRGB parses a hex color string like "#FF00AA" into RGB components.
func hexToRGB(hex string) (r, g, b uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var ri, gi, bi int
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &ri, &gi, &bi)
	return uint8(ri), uint8(gi), uint8(bi)
}

// sRGBToLinear converts an sRGB channel value (0-255) to linear light.
func sRGBToLinear(c uint8) float64 {
	v := float64(c) / 255.0
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// RelativeLuminance computes the WCAG relative luminance of a hex color.
// https://www.w3.org/TR/WCAG20/#relativeluminancedef
func RelativeLuminance(hex string) float64 {
	r, g, b := hexToRGB(hex)
	return 0.2126*sRGBToLinear(r) + 0.7152*sRGBToLinear(g) + 0.0722*sRGBToLinear(b)
}

// ContrastRatio computes the WCAG contrast ratio between two hex colors.
// https://www.w3.org/TR/WCAG20/#contrast-ratiodef
func ContrastRatio(fg, bg string) float64 {
	l1 := RelativeLuminance(fg)
	l2 := RelativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// isHighContrast checks if high-contrast mode is requested.
func isHighContrast() bool {
	_, hc := os.LookupEnv("GROVE_HIGH_CONTRAST")
	return hc
}

// highContrastColorScheme returns a ColorScheme with higher contrast values.
// All foreground colors are adjusted to meet WCAG AA (4.5:1) against both
// dark and light backgrounds, including TextMuted which is normally exempt.
func highContrastColorScheme() ColorScheme {
	return ColorScheme{
		// Brand — brighter for dark, darker for light
		Primary:   adaptiveColor("#C4B5FD", "#6D28D9"),
		Secondary: adaptiveColor("#7DD3FC", "#0369A1"),

		// Status — pushed to higher contrast
		Success: adaptiveColor("#6EE7B7", "#047857"),
		Warning: adaptiveColor("#FDE68A", "#B45309"),
		Danger:  adaptiveColor("#FCA5A5", "#B91C1C"),
		Info:    adaptiveColor("#93C5FD", "#1D4ED8"),

		// Surface — same as default (backgrounds)
		SurfaceBg:     adaptiveColor("#1E1E2E", "#FFFFFF"),
		SurfaceFg:     adaptiveColor("#E4E8F7", "#0F172A"),
		SurfaceDim:    adaptiveColor("#7F849C", "#64748B"),
		SurfaceBorder: adaptiveColor("#585B70", "#94A3B8"),

		// Selection / Header
		SelectionBg: adaptiveColor("#313244", "#E2E8F0"),
		HeaderBg:    adaptiveColor("#181825", "#F1F5F9"),

		// Text — all meet 4.5:1 including muted
		TextNormal: adaptiveColor("#E4E8F7", "#0F172A"),
		TextBright: adaptiveColor("#FFFFFF", "#000000"),
		TextMuted:  adaptiveColor("#A6ADC8", "#475569"),
	}
}
