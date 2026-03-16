package main

import (
	"os"
	"path/filepath"
	"testing"
)

func resetGlobals() {
	targetColors = nil
	config = Configuration{}
}

func TestLoadColorsFile_Valid(t *testing.T) {
	resetGlobals()
	config.colorTolerance = 30

	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":100,"g":150,"b":200,"tolerance":25}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	err := loadColorsFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targetColors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(targetColors))
	}
	tc := targetColors[0]
	if tc.R != 100 || tc.G != 150 || tc.B != 200 {
		t.Errorf("unexpected color: R=%d G=%d B=%d", tc.R, tc.G, tc.B)
	}
	if tc.Tolerance != 25 {
		t.Errorf("expected tolerance 25, got %d", tc.Tolerance)
	}
}

func TestLoadColorsFile_MissingFile(t *testing.T) {
	resetGlobals()
	err := loadColorsFile("/nonexistent/path/colors.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadColorsFile_MalformedJSON(t *testing.T) {
	resetGlobals()
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0644); err != nil {
		t.Fatal(err)
	}

	err := loadColorsFile(path)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestLoadColorsFile_DefaultTolerance(t *testing.T) {
	resetGlobals()
	config.colorTolerance = 42

	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":10,"g":20,"b":30,"tolerance":0}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := loadColorsFile(path); err != nil {
		t.Fatal(err)
	}
	if targetColors[0].Tolerance != 42 {
		t.Errorf("expected default tolerance 42, got %d", targetColors[0].Tolerance)
	}
}

func TestLoadColorsFile_CustomTolerance(t *testing.T) {
	resetGlobals()
	config.colorTolerance = 30

	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":10,"g":20,"b":30,"tolerance":50}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := loadColorsFile(path); err != nil {
		t.Fatal(err)
	}
	if targetColors[0].Tolerance != 50 {
		t.Errorf("expected custom tolerance 50, got %d", targetColors[0].Tolerance)
	}
}

func TestLoadColorsFile_AppendsToExisting(t *testing.T) {
	resetGlobals()
	config.colorTolerance = 30
	targetColors = []TargetColor{{R: 1, G: 2, B: 3, Tolerance: 10, Sampled: true}}

	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":100,"g":200,"b":255,"tolerance":20}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := loadColorsFile(path); err != nil {
		t.Fatal(err)
	}
	if len(targetColors) != 2 {
		t.Fatalf("expected 2 colors after append, got %d", len(targetColors))
	}
	if targetColors[0].R != 1 {
		t.Error("first color should be the pre-existing one")
	}
	if targetColors[1].R != 100 {
		t.Error("second color should be from file")
	}
}
