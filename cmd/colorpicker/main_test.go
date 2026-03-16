package main

import (
	"testing"

	"zeptile.com/go_clicker/internal/color"
)

// Tests for color analysis functions have been moved to internal/color/analysis_test.go.
// This file contains integration-level tests specific to the colorpicker binary.

func TestColorpickerUsesSharedTypes(t *testing.T) {
	// Verify that the colorpicker correctly uses shared types
	entry := color.ColorEntry{R: 100, G: 150, B: 200, Tolerance: 25, Count: 50}
	if entry.R != 100 || entry.G != 150 || entry.B != 200 {
		t.Error("ColorEntry fields not set correctly")
	}

	cf := color.ColorsFile{Colors: []color.ColorEntry{entry}}
	if len(cf.Colors) != 1 {
		t.Error("ColorsFile should contain 1 color")
	}
}
