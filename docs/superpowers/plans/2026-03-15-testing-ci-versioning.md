# Unit Testing, CI/CD & CalVer Versioning — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unit tests for core logic, CI pipelines for PR testing and automated releases, and CalVer versioning with GoReleaser.

**Architecture:** Extract pure logic from platform-dependent functions for testability. CI uses GitHub Actions with matrix builds (Ubuntu for Linux/Windows, macOS runner for darwin). GoReleaser v2 with `--split`/`--merge` handles CGo cross-compilation.

**Tech Stack:** Go 1.24, GitHub Actions, GoReleaser v2, pflag, CalVer (YYYY.MM.PATCH)

**Spec:** `docs/superpowers/specs/2026-03-15-testing-ci-versioning-design.md`

---

## Chunk 1: Refactoring & Core Tests

**TDD approach:** Write tests first (Task 1-3), then extract `matchesAnyColor` to make them pass (Task 4). Tasks 1-3 test existing functions that already compile. Task 1's `matchesAnyColor` tests won't compile until Task 4, which is the intended red-green cycle.

### Task 1: Write color matching tests (`color_test.go`)

**Files:**
- Create: `color_test.go`

- [ ] **Step 1: Create `color_test.go` with all color matching tests**

```go
package main

import (
	"testing"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-1, 1},
	}
	for _, tt := range tests {
		got := abs(tt.input)
		if got != tt.expected {
			t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestMatchesAnyColor_ExactMatch(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, _ := matchesAnyColor(100, 150, 200, colors)
	if !matched {
		t.Error("expected exact match to return true")
	}
}

func TestMatchesAnyColor_WithinTolerance(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, _ := matchesAnyColor(129, 179, 171, colors)
	if !matched {
		t.Error("expected match within tolerance (diff=29 per channel)")
	}
}

func TestMatchesAnyColor_OutsideTolerance(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, debugInfo := matchesAnyColor(200, 150, 200, colors)
	if matched {
		t.Error("expected no match when R diff=100 exceeds tolerance=30")
	}
	if debugInfo == "" {
		t.Error("expected debug info with closest color details")
	}
}

func TestMatchesAnyColor_PerChannelTolerance(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	// R diff=31 (over), G diff=0 (under), B diff=0 (under)
	matched, _ := matchesAnyColor(131, 150, 200, colors)
	if matched {
		t.Error("expected no match when one channel exceeds tolerance")
	}
}

func TestMatchesAnyColor_MultipleColors_MatchesSecond(t *testing.T) {
	colors := []TargetColor{
		{R: 10, G: 10, B: 10, Tolerance: 5},
		{R: 200, G: 200, B: 200, Tolerance: 10},
	}
	matched, _ := matchesAnyColor(205, 195, 200, colors)
	if !matched {
		t.Error("expected match against second target color")
	}
}

func TestMatchesAnyColor_EmptyColors(t *testing.T) {
	matched, debugInfo := matchesAnyColor(100, 100, 100, nil)
	if matched {
		t.Error("expected no match with nil color list")
	}
	if debugInfo != "" {
		t.Error("expected empty debug info with nil color list")
	}
}
```

- [ ] **Step 2: Run `TestAbs` to verify existing function tests pass**

Run: `go test -v -run "TestAbs" .`
Expected: PASS. (The `TestMatchesAnyColor_*` tests won't compile yet — `matchesAnyColor` doesn't exist. This is expected; we'll extract it in Task 4.)

- [ ] **Step 3: Commit**

```bash
git add color_test.go
git commit -m "test: add color matching unit tests (matchesAnyColor tests pending extraction)"
```

---

### Task 2: Write data loading tests (`data_test.go`)

**Files:**
- Create: `data_test.go`

- [ ] **Step 1: Create `data_test.go` with all JSON loading tests**

```go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

// resetGlobals clears global state before each test case.
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
```

- [ ] **Step 2: Run the tests**

Run: `go test -v -run "TestLoadColorsFile" .`
Expected: All PASS.

- [ ] **Step 3: Commit**

```bash
git add data_test.go
git commit -m "test: add data loading unit tests"
```

---

### Task 3: Write colorpicker tests (`cmd/colorpicker/main_test.go`)

**Files:**
- Create: `cmd/colorpicker/main_test.go`

**Constraint:** Do NOT import or call `captureRegion()`. Use fully opaque pixels only (`alpha=255`).

- [ ] **Step 1: Create `cmd/colorpicker/main_test.go`**

```go
package main

import (
	"image"
	"image/color"
	"testing"
)

func TestAbsInt(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
	}
	for _, tt := range tests {
		got := absInt(tt.input)
		if got != tt.expected {
			t.Errorf("absInt(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestIsBackground(t *testing.T) {
	tests := []struct {
		r, g, b  int
		expected bool
	}{
		{0, 0, 0, true},       // pure black
		{14, 14, 14, true},    // near-black
		{15, 15, 15, false},   // just outside threshold
		{255, 255, 255, true}, // pure white
		{241, 241, 241, true}, // near-white
		{240, 240, 240, false},// just outside threshold
		{100, 100, 100, false},// mid-gray
		{200, 50, 50, false},  // colored
	}
	for _, tt := range tests {
		got := isBackground(tt.r, tt.g, tt.b)
		if got != tt.expected {
			t.Errorf("isBackground(%d,%d,%d) = %v, want %v", tt.r, tt.g, tt.b, got, tt.expected)
		}
	}
}

// makeUniformImage creates a test image filled with a single opaque color.
func makeUniformImage(w, h int, r, g, b uint8) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// makeTwoColorImage creates an image with two distinct colors split vertically.
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
	colors := analyzeColors(img)
	if len(colors) < 2 {
		t.Fatalf("expected at least 2 colors, got %d", len(colors))
	}
}

func TestAnalyzeColors_SkipsBackground(t *testing.T) {
	// Image is all near-black (r,g,b < 15)
	img := makeUniformImage(50, 50, 5, 5, 5)
	colors := analyzeColors(img)
	if len(colors) != 0 {
		t.Errorf("expected 0 colors (background filtered), got %d", len(colors))
	}
}

func TestAnalyzeColors_AllBackgroundEdgeCase(t *testing.T) {
	// Mix of near-black and near-white — both should be filtered
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
	colors := analyzeColors(img)
	if len(colors) != 0 {
		t.Errorf("expected 0 colors (all background), got %d", len(colors))
	}
}

func TestAnalyzeColors_MaxTenColors(t *testing.T) {
	// Create image with 15 distinct color bands
	img := image.NewNRGBA(image.Rect(0, 0, 150, 100))
	for x := 0; x < 150; x++ {
		band := x / 10 // 0-14
		r := uint8(20 + band*16)
		g := uint8(50)
		b := uint8(100 + band*10)
		for y := 0; y < 100; y++ {
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	colors := analyzeColors(img)
	if len(colors) > 10 {
		t.Errorf("expected at most 10 colors, got %d", len(colors))
	}
}

func TestDeduplicateColors(t *testing.T) {
	colors := []ColorEntry{
		{R: 100, G: 100, B: 100, Count: 50},
		{R: 110, G: 110, B: 110, Count: 30}, // within 16 tolerance
	}
	result := deduplicateColors(colors)
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
		{R: 200, G: 200, B: 200, Count: 30}, // far apart
	}
	result := deduplicateColors(colors)
	if len(result) != 2 {
		t.Fatalf("expected 2 separate colors, got %d", len(result))
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test -v ./cmd/colorpicker/`
Expected: All PASS.

- [ ] **Step 3: Commit**

```bash
git add cmd/colorpicker/main_test.go
git commit -m "test: add colorpicker analysis unit tests"
```

---

### Task 4: Extract `matchesAnyColor` from `isColorMatch` (GREEN phase)

**Files:**
- Modify: `main.go:154-185` — extract pure matching logic into `matchesAnyColor`

- [ ] **Step 1: Extract `matchesAnyColor` function**

In `main.go`, add a new function `matchesAnyColor` containing the pure matching logic, and update `isColorMatch` to call it. Place `matchesAnyColor` directly above `isColorMatch`.

```go
func matchesAnyColor(r, g, b int, tc []TargetColor) (bool, string) {
	closest := ""
	minDist := 999999
	for _, t := range tc {
		dr, dg, db := abs(r-t.R), abs(g-t.G), abs(b-t.B)
		dist := dr + dg + db
		if dist < minDist {
			minDist = dist
			closest = fmt.Sprintf("pixel=(%d,%d,%d) closest_target=(%d,%d,%d) diff=(%d,%d,%d) tol=%d",
				r, g, b, t.R, t.G, t.B, dr, dg, db, t.Tolerance)
		}
		if dr <= t.Tolerance && dg <= t.Tolerance && db <= t.Tolerance {
			return true, ""
		}
	}
	return false, closest
}
```

Update `isColorMatch` to delegate to it:

```go
func isColorMatch() (bool, string) {
	colorMu.RLock()
	tc := targetColors
	colorMu.RUnlock()

	if len(tc) == 0 {
		return false, "no colors loaded"
	}

	x, y := cursorPos()
	r, g, b, err := getPixelColor(x, y)
	if err != nil {
		return false, fmt.Sprintf("pixel error: %v", err)
	}

	return matchesAnyColor(r, g, b, tc)
}
```

Note: The `"no colors loaded"` path remains in `isColorMatch` (not `matchesAnyColor`) since it's an I/O-level concern. This path cannot be unit-tested without mocking platform deps, which is an accepted gap.

- [ ] **Step 2: Run all tests to verify GREEN**

Run: `go test -v -race ./...`
Expected: All PASS — including `TestMatchesAnyColor_*` tests from Task 1.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "refactor: extract matchesAnyColor from isColorMatch for testability"
```

---

## Chunk 2: Version Injection

### Task 5: Add `--version` flag to main binary

**Files:**
- Modify: `main.go:14` — add `version` variable
- Modify: `main.go:194-223` — add `--version` flag to `parseConsoleArguments`

- [ ] **Step 1: Add version variable and flag**

Add after line 14 (`var running = false`):

```go
var version = "dev"
```

In `parseConsoleArguments()`, add a local `showVersion` variable, register the flag, and handle it after `pflag.Parse()`:

```go
func parseConsoleArguments() Configuration {
	var config Configuration
	var showVersion bool

	pflag.BoolVar(&config.randomMode, "random", false, "Flag to enable random mode, click interval is random between 0 and randomIntervalEnd")
	pflag.BoolVar(&config.debugMode, "debug", false, "Flag to enable debug mode, prints debug logs to console")
	pflag.BoolVar(&config.colorMode, "color", false, "Flag to enable color mode, only clicks when pixel under cursor matches target color")

	pflag.Int64Var(&config.intervalMs, "intervalMs", 50, "Interval in milliseconds to click in normal mode")
	pflag.Int64Var(&config.randomIntervalEnd, "randomIntervalEnd", 100, "Interval treshold in milliseconds to click in random mode")
	pflag.IntVar(&config.colorTolerance, "colorTolerance", 30, "RGB tolerance per channel for color matching (0-255)")
	pflag.StringVar(&config.colorsFile, "colorsFile", "", "Path to colors.json file generated by colorpicker tool")
	pflag.BoolVar(&showVersion, "version", false, "Print version and exit")

	pflag.Parse()

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// ... rest of validation unchanged
```

- [ ] **Step 2: Verify it compiles**

Run: `go build -o /dev/null .`
Expected: Success.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: add --version flag and version variable to main binary"
```

---

### Task 6: Add `--version` flag to colorpicker

**Files:**
- Modify: `cmd/colorpicker/main.go` — add version variable and `--version` check

The colorpicker uses raw `os.Args`, not pflag. Add a simple `os.Args` check before the existing argument parsing.

- [ ] **Step 1: Add version variable and check**

Add after the import block:

```go
var version = "dev"
```

At the start of `main()`, before the existing `os.Args` parsing:

```go
func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Println(version)
			os.Exit(0)
		}
	}

	captures := 1
	// ... rest unchanged
```

- [ ] **Step 2: Verify it compiles**

Run: `go build -o /dev/null ./cmd/colorpicker/`
Expected: Success.

- [ ] **Step 3: Commit**

```bash
git add cmd/colorpicker/main.go
git commit -m "feat: add --version flag to colorpicker"
```

---

## Chunk 3: CI Pipelines & GoReleaser

### Task 7: Create PR test pipeline

**Files:**
- Create: `.github/workflows/test.yml`

- [ ] **Step 1: Create `.github/workflows/` directory and `test.yml`**

```yaml
name: Test
on:
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: Install system dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libx11-dev libxcursor-dev libxrandr-dev \
            libxtst-dev libxi-dev libxinerama-dev libxkbcommon-dev
      - name: Verify modules
        run: go mod verify
      - name: Vet
        run: go vet ./...
      - name: Test
        run: go test -v -race ./...
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: add PR test pipeline with go vet and go test"
```

---

### Task 8: Create GoReleaser configuration

**Files:**
- Create: `.goreleaser.yml`

- [ ] **Step 1: Create `.goreleaser.yml`**

```yaml
version: 2

builds:
  - id: go_clicker
    main: .
    binary: go_clicker
    ldflags:
      - -s -w -X main.version={{.Version}}
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: linux
        goarch: arm64
      - goos: windows
        goarch: arm64
    overrides:
      - goos: windows
        goarch: amd64
        env:
          - CC=x86_64-w64-mingw32-gcc
          - CXX=x86_64-w64-mingw32-g++

  - id: colorpicker
    dir: cmd/colorpicker
    binary: colorpicker
    ldflags:
      - -s -w -X main.version={{.Version}}
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: linux
        goarch: arm64
      - goos: windows
        goarch: arm64
    overrides:
      - goos: windows
        goarch: amd64
        env:
          - CC=x86_64-w64-mingw32-gcc
          - CXX=x86_64-w64-mingw32-g++

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

- [ ] **Step 2: Commit**

```bash
git add .goreleaser.yml
git commit -m "ci: add GoReleaser v2 config for cross-platform builds"
```

---

### Task 9: Create release pipeline

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create `.github/workflows/release.yml`**

```yaml
name: Release
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: Install system dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libx11-dev libxcursor-dev libxrandr-dev \
            libxtst-dev libxi-dev libxinerama-dev libxkbcommon-dev
      - run: go test -v -race ./...

  tag:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: write
    outputs:
      tag: ${{ steps.calver.outputs.tag }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Generate CalVer tag
        id: calver
        run: |
          YEAR_MONTH=$(date +%Y.%-m)
          LATEST=$(git tag -l "${YEAR_MONTH}.*" --sort=-v:refname | head -n1)
          if [ -z "$LATEST" ]; then
            PATCH=0
          else
            PATCH=$(echo "$LATEST" | awk -F. '{print $3+1}')
          fi
          TAG="${YEAR_MONTH}.${PATCH}"
          echo "tag=$TAG" >> "$GITHUB_OUTPUT"
      - name: Create and push tag
        run: |
          git tag "${{ steps.calver.outputs.tag }}" "${{ github.sha }}"
          git push origin "${{ steps.calver.outputs.tag }}"

  release:
    needs: tag
    strategy:
      matrix:
        include:
          - runner: ubuntu-latest
            goos: linux
          - runner: ubuntu-latest
            goos: windows
          - runner: macos-latest
            goos: darwin
    runs-on: ${{ matrix.runner }}
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: Install Linux/Windows dependencies
        if: matrix.runner == 'ubuntu-latest'
        run: |
          sudo apt-get update
          sudo apt-get install -y libx11-dev libxcursor-dev libxrandr-dev \
            libxtst-dev libxi-dev libxinerama-dev libxkbcommon-dev \
            gcc-mingw-w64-x86-64
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean --split
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_CURRENT_TAG: ${{ needs.tag.outputs.tag }}
          GOOS: ${{ matrix.goos }}
      - uses: actions/upload-artifact@v4
        with:
          name: artifacts-${{ matrix.goos }}
          path: dist/

  merge:
    needs: [tag, release]
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/download-artifact@v4
        with:
          pattern: artifacts-*
          path: dist/
          merge-multiple: true
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: continue --merge
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_CURRENT_TAG: ${{ needs.tag.outputs.tag }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add release pipeline with CalVer tagging and matrix builds"
```

---

### Task 10: Run all tests locally to verify

- [ ] **Step 1: Run the full test suite**

Run: `go test -v -race ./...`
Expected: All tests PASS.

- [ ] **Step 2: Run go vet**

Run: `go vet ./...`
Expected: No issues.

- [ ] **Step 3: Verify GoReleaser config syntax (if goreleaser is available)**

Run: `goreleaser check` (skip if not installed)
Expected: config is valid, or skip.
