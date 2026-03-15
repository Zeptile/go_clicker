//go:build linux

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"time"

	"github.com/go-vgo/robotgo"
)

func isWayland() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	return os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

func captureRegion() (image.Image, error) {
	if isWayland() {
		return captureWayland()
	}
	return captureX11()
}

func captureWayland() (image.Image, error) {
	slurp := exec.Command("slurp")
	regionBytes, err := slurp.Output()
	if err != nil {
		return nil, fmt.Errorf("slurp: %w", err)
	}
	region := string(regionBytes)
	region = region[:len(region)-1]
	fmt.Printf("Selected region: %s\n", region)

	tmpFile := "/tmp/colorpicker_capture.png"
	grim := exec.Command("grim", "-g", region, tmpFile)
	if err := grim.Run(); err != nil {
		return nil, fmt.Errorf("grim: %w", err)
	}
	defer os.Remove(tmpFile)

	f, err := os.Open(tmpFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return png.Decode(f)
}

func captureX11() (image.Image, error) {
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
