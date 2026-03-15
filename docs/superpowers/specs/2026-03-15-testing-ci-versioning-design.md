# Unit Testing, CI/CD & CalVer Versioning — Design Spec

## Overview

Add unit tests for core logic, a CI pipeline that runs tests on PRs, and a release pipeline that auto-versions (CalVer) and publishes cross-platform binaries on push to main.

## 1. Unit Tests

### Scope

Test core logic only — no platform-specific code (robotgo, ydotool, screen capture). Platform files use CGo and require display servers, making them unsuitable for CI unit tests.

### Test Files

#### `color_test.go` (new file, package main)

Tests for color matching and utility functions extracted from `main.go`:

- **`TestAbs`** — positive, negative, zero inputs
- **`TestIsColorMatch_NoColors`** — returns false with "no colors loaded" when target list is empty
- **Color distance logic** — since `isColorMatch` calls platform functions (`cursorPos`, `getPixelColor`), we need to extract the pure matching logic into a testable function:
  - Extract `matchesAnyColor(r, g, b int, colors []TargetColor) (bool, string)` from `isColorMatch`
  - Test: exact match returns true
  - Test: within tolerance returns true (e.g., tolerance=30, diff=29 per channel)
  - Test: outside tolerance returns false with closest-color debug info
  - Test: per-channel tolerance (one channel over, others under = no match)
  - Test: multiple target colors, matches second one

#### `data_test.go` (new file, package main)

Tests for data loading and JSON parsing:

- **`TestLoadColorsFile_Valid`** — load a well-formed JSON file, verify colors are appended to `targetColors`
- **`TestLoadColorsFile_MissingFile`** — returns error for nonexistent path
- **`TestLoadColorsFile_MalformedJSON`** — returns error for invalid JSON
- **`TestLoadColorsFile_DefaultTolerance`** — colors with tolerance=0 get config's default tolerance
- **`TestLoadColorsFile_CustomTolerance`** — colors with tolerance>0 keep their own tolerance
- **`TestLoadColorsFile_AppendsToExisting`** — calling twice accumulates colors (not replaces)

Note: `loadColorsFile` uses the global `targetColors` and `config` variables. Tests will need to reset these globals before each test case.

#### `cmd/colorpicker/main_test.go` (new file, package main)

Tests for color analysis functions:

- **`TestAnalyzeColors`** — given a synthetic `image.NRGBA` with fully opaque pixels, verify dominant colors are extracted correctly
- **`TestAnalyzeColors_SkipsBackground`** — black (0,0,0) and white (255,255,255) pixels are filtered out
- **`TestAnalyzeColors_AllBackgroundEdgeCase`** — when all top buckets are background, function returns empty (not stuck)
- **`TestAnalyzeColors_MaxTenColors`** — output is capped at 10 entries

**Constraint:** Colorpicker tests must NOT import or call `captureRegion()` — that function depends on platform-specific display tools. All test images must use fully opaque pixels (`alpha=255`) since `analyzeColors` calls `RGBA()` which returns premultiplied values.
- **`TestDeduplicateColors`** — colors within 16-unit tolerance are merged, counts are summed
- **`TestDeduplicateColors_NoMerge`** — colors >16 apart remain separate
- **`TestAbsInt`** — positive, negative, zero
- **`TestIsBackground`** — near-black and near-white return true, other colors return false

### Refactoring Required

Extract `matchesAnyColor(r, g, b int, colors []TargetColor) (bool, string)` from `isColorMatch()` in `main.go`. The current `isColorMatch` calls platform-specific `cursorPos()` and `getPixelColor()`, but the matching logic itself is pure and testable. After extraction, `isColorMatch` becomes:

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

## 2. Version Injection

### Version Variable

Add to `main.go`:

```go
var version = "dev"
```

### `--version` Flag

Add a `--version` flag to `parseConsoleArguments()`. When set, print the version and exit:

```go
pflag.BoolVar(&showVersion, "version", false, "Print version and exit")
// after pflag.Parse():
if showVersion {
    fmt.Println(version)
    os.Exit(0)
}
```

### Build-time Injection

GoReleaser injects the version via ldflags:

```
-ldflags "-X main.version={{.Version}}"
```

For the colorpicker binary, add the same `version` variable and `--version` flag to `cmd/colorpicker/main.go`.

## 3. CI Pipeline — Test on PR

### File: `.github/workflows/test.yml`

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

System dependencies are needed because `go vet ./...` and `go test ./...` compile all packages including platform files that import robotgo (CGo). The tests themselves only exercise pure Go logic.

## 4. CI Pipeline — Release on Push to Main

### File: `.github/workflows/release.yml`

The release workflow uses a matrix strategy: Linux+Windows builds run on `ubuntu-latest`, macOS builds run on `macos-latest`. This is required because robotgo's CGo dependencies need native platform toolchains (Cocoa frameworks on macOS, X11 headers on Linux).

The workflow has three jobs:
1. **test** — runs the full test suite (gate before release)
2. **release** — creates the CalVer tag and runs GoReleaser with `--split` per OS
3. **merge** — merges the split artifacts into a single GitHub release

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

### CalVer Tag Generation Logic

1. Get current year.month (e.g., `2026.3`)
2. Find latest existing tag matching `2026.3.*`
3. If none exists, patch = 0 (first release: `2026.3.0`)
4. Otherwise, increment patch (e.g., `2026.3.1`, `2026.3.2`)

**Note on race conditions:** If two pushes to main happen in quick succession, both runs could compute the same tag. The second `git push origin $TAG` would fail, which is acceptable — the failed run can be manually re-triggered. This is unlikely in practice for a project of this size.

## 5. GoReleaser Configuration

### File: `.goreleaser.yml`

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

### Build Targets

| OS | Arch | Notes |
|---|---|---|
| linux | amd64 | Primary platform |
| darwin | amd64 | Intel Mac |
| darwin | arm64 | Apple Silicon |
| windows | amd64 | Windows |

Linux arm64 and Windows arm64 are excluded — robotgo CGo cross-compilation for these targets is unreliable.

### CGo Cross-Compilation Strategy

macOS builds require native Cocoa frameworks, so we use a **matrix build strategy**: each OS builds on its native runner. GoReleaser's `--split` flag builds only the targets matching the current `GOOS`, and the `continue --merge` step combines all split artifacts into a single GitHub release.

For Windows cross-compilation from Ubuntu, we use `gcc-mingw-w64-x86-64` with CC/CXX overrides in the GoReleaser config.

## 6. File Summary

| File | Action |
|---|---|
| `main.go` | Add `version` var, `--version` flag, extract `matchesAnyColor()` |
| `cmd/colorpicker/main.go` | Add `version` var, `--version` flag |
| `color_test.go` | New — tests for color matching logic |
| `data_test.go` | New — tests for JSON loading |
| `cmd/colorpicker/main_test.go` | New — tests for color analysis |
| `.github/workflows/test.yml` | New — PR test pipeline |
| `.github/workflows/release.yml` | New — release pipeline |
| `.goreleaser.yml` | New — GoReleaser config |

## 7. Open Risks

1. **robotgo system deps on CI** — `go vet` and `go test` compile all packages including platform files. CI needs X11 development headers even though tests don't exercise platform code.
2. **Global state in tests** — `loadColorsFile` mutates global `targetColors` and reads global `config`. Tests must reset these between cases.
3. **GoReleaser split/merge** — The `--split` + `continue --merge` workflow requires GoReleaser v2. We pin to `~> v2` to avoid breakage.
4. **CalVer tag race** — Concurrent pushes to main could compute the same tag. Unlikely in practice; failed run can be re-triggered manually.
