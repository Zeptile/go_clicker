package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"zeptile.com/go_clicker/internal/color"
)

var version = "dev"

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Println(version)
			os.Exit(0)
		}
	}

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

	var allColors []color.ColorEntry
	if captures > 1 {
		if data, err := os.ReadFile(outFile); err == nil {
			var existing color.ColorsFile
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

		colors := color.AnalyzeColors(img)

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

	allColors = color.DeduplicateColors(allColors)

	// Filter out background/achromatic colors (including any from previous runs)
	var filtered []color.ColorEntry
	for _, c := range allColors {
		if !color.IsBackground(c.R, c.G, c.B) {
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

	cf := color.ColorsFile{Colors: allColors}
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
	fmt.Println("Usage: clicker --colorsFile", outFile)
}
