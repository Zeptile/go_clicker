package color

import (
	"image"
	"sort"
)

func DeduplicateColors(colors []ColorEntry) []ColorEntry {
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
			if Abs(c.R-o.R) <= 16 && Abs(c.G-o.G) <= 16 && Abs(c.B-o.B) <= 16 {
				used[j] = true
				merged.Count += o.Count
			}
		}
		used[i] = true
		deduped = append(deduped, merged)
	}

	return deduped
}

func AnalyzeColors(img image.Image) []ColorEntry {
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

			if Abs(avgR-oR) <= 16 && Abs(avgG-oG) <= 16 && Abs(avgB-oB) <= 16 {
				used[j] = true
				mergedCount += other.count
			}
		}

		used[i] = true

		if IsBackground(avgR, avgG, avgB) {
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

func IsBackground(r, g, b int) bool {
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
