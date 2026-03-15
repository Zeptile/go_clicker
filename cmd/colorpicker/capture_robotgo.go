//go:build darwin || windows

package main

import (
	"fmt"
	"image"
	"time"

	"github.com/go-vgo/robotgo"
)

func captureRegion() (image.Image, error) {
	fmt.Println("Hover over the TOP-LEFT corner...")
	for i := 3; i > 0; i-- {
		fmt.Printf("  Capturing in %d...\n", i)
		time.Sleep(time.Second)
	}
	x1, y1 := robotgo.Location()
	fmt.Printf("Top-left: (%d, %d)\n", x1, y1)

	fmt.Println("Hover over the BOTTOM-RIGHT corner...")
	for i := 3; i > 0; i-- {
		fmt.Printf("  Capturing in %d...\n", i)
		time.Sleep(time.Second)
	}
	x2, y2 := robotgo.Location()
	fmt.Printf("Bottom-right: (%d, %d)\n", x2, y2)

	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	w := x2 - x1
	h := y2 - y1
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid region: width and height must be > 0")
	}

	img, err := robotgo.CaptureImg(x1, y1, w, h)
	if err != nil {
		return nil, fmt.Errorf("capture: %w", err)
	}

	return img, nil
}
