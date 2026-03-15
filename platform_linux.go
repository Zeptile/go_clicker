//go:build linux

package main

import (
	"fmt"
	"image/png"
	"os/exec"
	"strconv"
	"strings"
)

func click() {
	exec.Command("ydotool", "click", "0xC0").Run()
}

func cursorPos() (int, int) {
	out, err := exec.Command("hyprctl", "cursorpos").Output()
	if err != nil {
		return 0, 0
	}
	parts := strings.Split(strings.TrimSpace(string(out)), ", ")
	if len(parts) != 2 {
		return 0, 0
	}
	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(parts[1])
	return x, y
}

func getPixelColor(x, y int) (int, int, int, error) {
	region := fmt.Sprintf("%d,%d 1x1", x, y)
	cmd := exec.Command("grim", "-g", region, "-")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("grim: %w", err)
	}

	img, err := png.Decode(strings.NewReader(string(out)))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("decode: %w", err)
	}

	c := img.At(0, 0)
	r, g, b, _ := c.RGBA()
	return int(r >> 8), int(g >> 8), int(b >> 8), nil
}
