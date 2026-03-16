package color

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":100,"g":150,"b":200,"tolerance":25}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	colors, err := LoadFile(path, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	tc := colors[0]
	if tc.R != 100 || tc.G != 150 || tc.B != 200 {
		t.Errorf("unexpected color: R=%d G=%d B=%d", tc.R, tc.G, tc.B)
	}
	if tc.Tolerance != 25 {
		t.Errorf("expected tolerance 25, got %d", tc.Tolerance)
	}
}

func TestLoadFile_MissingFile(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/colors.json", 30)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFile_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path, 30)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestLoadFile_DefaultTolerance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":10,"g":20,"b":30,"tolerance":0}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	colors, err := LoadFile(path, 42)
	if err != nil {
		t.Fatal(err)
	}
	if colors[0].Tolerance != 42 {
		t.Errorf("expected default tolerance 42, got %d", colors[0].Tolerance)
	}
}

func TestLoadFile_CustomTolerance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":10,"g":20,"b":30,"tolerance":50}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	colors, err := LoadFile(path, 30)
	if err != nil {
		t.Fatal(err)
	}
	if colors[0].Tolerance != 50 {
		t.Errorf("expected custom tolerance 50, got %d", colors[0].Tolerance)
	}
}

func TestLoadFile_MultipleColors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "colors.json")
	data := []byte(`{"colors":[{"r":100,"g":200,"b":255,"tolerance":20},{"r":10,"g":20,"b":30,"tolerance":15}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	colors, err := LoadFile(path, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(colors) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(colors))
	}
	if colors[0].R != 100 {
		t.Error("first color should have R=100")
	}
	if colors[1].R != 10 {
		t.Error("second color should have R=10")
	}
}
