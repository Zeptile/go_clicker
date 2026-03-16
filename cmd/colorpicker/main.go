package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"sort"
	"strconv"
)

type ColorEntry struct {
	R         int `json:"r"`
	G         int `json:"g"`
	B         int `json:"b"`
	Tolerance int `json:"tolerance"`
	Count     int `json:"count"`
}

type ColorsFile struct {
	Colors []ColorEntry `json:"colors"`
}

func main() {
	captures := 1
	outFile := "colors.json"

	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil {
			captures = n
		} else {
			outFile = os.Args[1]
		}
	}
	if len(os.Args) > 2 {
		outFile = os.Args[2]
	}

	var allColors []ColorEntry
	if captures > 1 {
		if data, err := os.ReadFile(outFile); err == nil {
			var existing ColorsFile
			if json.Unmarshal(data, &existing) == nil {
				allColors = existing.Colors
				fmt.Printf("Loaded %d existing colors from %s\n", len(allColors), outFile)
			}
		}
	}

	fmt.Printf("=== Color Picker Tool === (%d capture(s))\n", captures)

	reader := bufio.NewReader(os.Stdin)

	for i := 0; i < captures; i++ {
		if i > 0 {
			fmt.Printf("\nPress Enter when ready for capture %d/%d...", i+1, captures)
			reader.ReadString('\n')
		}
		fmt.Printf("[Capture %d/%d] Select a region...\n", i+1, captures)

		img, err := captureRegion()
		if err != nil {
			fmt.Printf("Error capturing: %v\n", err)
			os.Exit(1)
		}

		colors := analyzeColors(img)

		bounds := img.Bounds()
		totalPixels := bounds.Dx() * bounds.Dy()
		fmt.Printf("Found %d dominant colors from %d pixels:\n", len(colors), totalPixels)
		for j, c := range colors {
			pct := float64(c.Count) / float64(totalPixels) * 100
			fmt.Printf("  %d. \033[48;2;%d;%d;%dm    \033[0m #%02x%02x%02x (R:%d G:%d B:%d) - %.1f%% of pixels\n",
				j+1, c.R, c.G, c.B, c.R, c.G, c.B, c.R, c.G, c.B, pct)
		}

		allColors = append(allColors, colors...)
	}

	allColors = deduplicateColors(allColors)

	// Filter out background/achromatic colors (including any from previous runs)
	var filtered []ColorEntry
	for _, c := range allColors {
		if !isBackground(c.R, c.G, c.B) {
			filtered = append(filtered, c)
		}
	}
	allColors = filtered

	if len(allColors) == 0 {
		fmt.Println("No meaningful colors found")
		os.Exit(1)
	}

	fmt.Printf("\nTotal: %d unique colors\n", len(allColors))
	for i, c := range allColors {
		fmt.Printf("  %d. \033[48;2;%d;%d;%dm    \033[0m #%02x%02x%02x (R:%d G:%d B:%d) count:%d\n",
			i+1, c.R, c.G, c.B, c.R, c.G, c.B, c.R, c.G, c.B, c.Count)
	}

	cf := ColorsFile{Colors: allColors}
	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nColors written to %s\n", outFile)
	fmt.Println("Usage: go_clicker --colorsFile", outFile)
}

func deduplicateColors(colors []ColorEntry) []ColorEntry {
	var deduped []ColorEntry
	used := make([]bool, len(colors))

	for i, c := range colors {
		if used[i] {
			continue
		}
		merged := c
		for j := i + 1; j < len(colors); j++ {
			if used[j] {
				continue
			}
			o := colors[j]
			if absInt(c.R-o.R) <= 16 && absInt(c.G-o.G) <= 16 && absInt(c.B-o.B) <= 16 {
				used[j] = true
				merged.Count += o.Count
			}
		}
		used[i] = true
		deduped = append(deduped, merged)
	}

	return deduped
}

func analyzeColors(img image.Image) []ColorEntry {
	type bucket struct {
		r, g, b int
	}
	freq := make(map[bucket]int)
	totalR := make(map[bucket]int64)
	totalG := make(map[bucket]int64)
	totalB := make(map[bucket]int64)

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(b>>8)

			bk := bucket{r: (r8 / 8) * 8, g: (g8 / 8) * 8, b: (b8 / 8) * 8}
			freq[bk]++
			totalR[bk] += int64(r8)
			totalG[bk] += int64(g8)
			totalB[bk] += int64(b8)
		}
	}

	type bucketCount struct {
		bk    bucket
		count int
	}
	var sorted []bucketCount
	for bk, count := range freq {
		sorted = append(sorted, bucketCount{bk, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var colors []ColorEntry
	used := make(map[int]bool)

	for i, bc := range sorted {
		if len(colors) >= 10 {
			break
		}
		if used[i] {
			continue
		}

		avgR := int(totalR[bc.bk] / int64(bc.count))
		avgG := int(totalG[bc.bk] / int64(bc.count))
		avgB := int(totalB[bc.bk] / int64(bc.count))
		mergedCount := bc.count

		for j := i + 1; j < len(sorted); j++ {
			if used[j] {
				continue
			}
			other := sorted[j]
			oR := int(totalR[other.bk] / int64(other.count))
			oG := int(totalG[other.bk] / int64(other.count))
			oB := int(totalB[other.bk] / int64(other.count))

			if absInt(avgR-oR) <= 16 && absInt(avgG-oG) <= 16 && absInt(avgB-oB) <= 16 {
				used[j] = true
				mergedCount += other.count
			}
		}

		used[i] = true

		if isBackground(avgR, avgG, avgB) {
			continue
		}

		colors = append(colors, ColorEntry{
			R:         avgR,
			G:         avgG,
			B:         avgB,
			Tolerance: 25,
			Count:     mergedCount,
		})
	}

	return colors
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func isBackground(r, g, b int) bool {
	if r < 15 && g < 15 && b < 15 {
		return true
	}
	if r > 240 && g > 240 && b > 240 {
		return true
	}
	// Filter achromatic (gray/neutral) colors — these are common UI
	// backgrounds that shouldn't be picked as target colors.
	maxC := r
	if g > maxC {
		maxC = g
	}
	if b > maxC {
		maxC = b
	}
	minC := r
	if g < minC {
		minC = g
	}
	if b < minC {
		minC = b
	}
	if maxC-minC < 30 {
		return true
	}
	return false
}
