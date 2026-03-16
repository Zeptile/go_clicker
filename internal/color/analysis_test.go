package color

import (
	"image"
	"image/color"
	"testing"
)

func TestIsBackground(t *testing.T) {
	tests := []struct {
		r, g, b  int
		expected bool
	}{
		{0, 0, 0, true},           // near-black
		{14, 14, 14, true},        // near-black
		{255, 255, 255, true},     // near-white
		{241, 241, 241, true},     // near-white
		{100, 100, 100, true},     // achromatic (gray, maxC-minC < 30)
		{15, 15, 15, true},        // achromatic
		{240, 240, 240, true},     // achromatic
		{200, 50, 50, false},      // chromatic (high saturation)
		{100, 150, 200, false},    // chromatic
		{50, 200, 50, false},      // chromatic
	}
	for _, tt := range tests {
		got := IsBackground(tt.r, tt.g, tt.b)
		if got != tt.expected {
			t.Errorf("IsBackground(%d,%d,%d) = %v, want %v", tt.r, tt.g, tt.b, got, tt.expected)
		}
	}
}

func makeUniformImage(w, h int, r, g, b uint8) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func makeTwoColorImage(w, h int, r1, g1, b1, r2, g2, b2 uint8) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < w/2 {
				img.SetNRGBA(x, y, color.NRGBA{R: r1, G: g1, B: b1, A: 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: r2, G: g2, B: b2, A: 255})
			}
		}
	}
	return img
}

func TestAnalyzeColors(t *testing.T) {
	img := makeTwoColorImage(100, 100, 200, 50, 50, 50, 50, 200)
	colors := AnalyzeColors(img)
	if len(colors) < 2 {
		t.Fatalf("expected at least 2 colors, got %d", len(colors))
	}
}

func TestAnalyzeColors_SkipsBackground(t *testing.T) {
	img := makeUniformImage(50, 50, 5, 5, 5)
	colors := AnalyzeColors(img)
	if len(colors) != 0 {
		t.Errorf("expected 0 colors (background filtered), got %d", len(colors))
	}
}

func TestAnalyzeColors_AllBackgroundEdgeCase(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if x < 50 {
				img.SetNRGBA(x, y, color.NRGBA{R: 5, G: 5, B: 5, A: 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 250, G: 250, B: 250, A: 255})
			}
		}
	}
	colors := AnalyzeColors(img)
	if len(colors) != 0 {
		t.Errorf("expected 0 colors (all background), got %d", len(colors))
	}
}

func TestAnalyzeColors_MaxTenColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 150, 100))
	for x := 0; x < 150; x++ {
		band := x / 10
		r := uint8(20 + band*16)
		g := uint8(50)
		b := uint8(100 + band*10)
		for y := 0; y < 100; y++ {
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	colors := AnalyzeColors(img)
	if len(colors) > 10 {
		t.Errorf("expected at most 10 colors, got %d", len(colors))
	}
}

func TestDeduplicateColors(t *testing.T) {
	colors := []ColorEntry{
		{R: 100, G: 100, B: 100, Count: 50},
		{R: 110, G: 110, B: 110, Count: 30},
	}
	result := DeduplicateColors(colors)
	if len(result) != 1 {
		t.Fatalf("expected 1 merged color, got %d", len(result))
	}
	if result[0].Count != 80 {
		t.Errorf("expected merged count 80, got %d", result[0].Count)
	}
}

func TestDeduplicateColors_NoMerge(t *testing.T) {
	colors := []ColorEntry{
		{R: 100, G: 100, B: 100, Count: 50},
		{R: 200, G: 200, B: 200, Count: 30},
	}
	result := DeduplicateColors(colors)
	if len(result) != 2 {
		t.Fatalf("expected 2 separate colors, got %d", len(result))
	}
}
