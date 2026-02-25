package theme

import "testing"

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex     string
		r, g, b uint8
		wantErr bool
	}{
		{"#FFFFFF", 255, 255, 255, false},
		{"#000000", 0, 0, 0, false},
		{"#FF0000", 255, 0, 0, false},
		{"#00FF00", 0, 255, 0, false},
		{"#0000FF", 0, 0, 255, false},
		{"invalid", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			r, g, b, err := HexToRGB(tt.hex)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HexToRGB(%s) expected error, got nil", tt.hex)
				}
				return
			}
			if err != nil {
				t.Errorf("HexToRGB(%s) unexpected error: %v", tt.hex, err)
				return
			}
			if r != tt.r || g != tt.g || b != tt.b {
				t.Errorf("HexToRGB(%s) = (%d,%d,%d), want (%d,%d,%d)",
					tt.hex, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestRelativeLuminance(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		wantMin float64
		wantMax float64
	}{
		{"Black", "#000000", -0.01, 0.01},
		{"White", "#FFFFFF", 0.99, 1.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lum := RelativeLuminance(tt.hex)
			if lum < tt.wantMin || lum > tt.wantMax {
				t.Errorf("RelativeLuminance(%s) = %.4f, want [%.2f, %.2f]",
					tt.hex, lum, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestContrastRatio(t *testing.T) {
	// Black on white should be ~21:1
	ratio := ContrastRatio("#000000", "#FFFFFF")
	if ratio < 20.9 || ratio > 21.1 {
		t.Errorf("ContrastRatio(black, white) = %.2f, want ~21.0", ratio)
	}

	// Same color should be 1:1
	ratio = ContrastRatio("#808080", "#808080")
	if ratio < 0.9 || ratio > 1.1 {
		t.Errorf("ContrastRatio(same, same) = %.2f, want ~1.0", ratio)
	}

	// Order shouldn't matter
	r1 := ContrastRatio("#000000", "#FFFFFF")
	r2 := ContrastRatio("#FFFFFF", "#000000")
	if r1 != r2 {
		t.Errorf("ContrastRatio should be order-independent: %f != %f", r1, r2)
	}
}

func TestHighContrastScheme_MeetsWCAGAA(t *testing.T) {
	scheme := HighContrastColorScheme()
	const minRatio = 4.5

	fgColors := map[string]struct{ Dark, Light string }{
		"Primary":    {scheme.Primary.Dark, scheme.Primary.Light},
		"Success":    {scheme.Success.Dark, scheme.Success.Light},
		"Warning":    {scheme.Warning.Dark, scheme.Warning.Light},
		"Danger":     {scheme.Danger.Dark, scheme.Danger.Light},
		"Info":       {scheme.Info.Dark, scheme.Info.Light},
		"TextMuted":  {scheme.TextMuted.Dark, scheme.TextMuted.Light},
		"TextNormal": {scheme.TextNormal.Dark, scheme.TextNormal.Light},
		"TextBright": {scheme.TextBright.Dark, scheme.TextBright.Light},
	}

	for name, fg := range fgColors {
		t.Run(name+"/dark", func(t *testing.T) {
			if _, _, _, err := HexToRGB(fg.Dark); err != nil {
				t.Fatalf("%s dark color is invalid: %v", name, err)
			}
			ratio := ContrastRatio(fg.Dark, scheme.SurfaceBg.Dark)
			if ratio < minRatio {
				t.Errorf("%s dark: ratio %.2f < %.1f", name, ratio, minRatio)
			}
		})
		t.Run(name+"/light", func(t *testing.T) {
			if _, _, _, err := HexToRGB(fg.Light); err != nil {
				t.Fatalf("%s light color is invalid: %v", name, err)
			}
			ratio := ContrastRatio(fg.Light, scheme.SurfaceBg.Light)
			if ratio < minRatio {
				t.Errorf("%s light: ratio %.2f < %.1f", name, ratio, minRatio)
			}
		})
	}
}
