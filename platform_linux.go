//go:build linux

package main

import (
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-vgo/robotgo"
)

func isWayland() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	return os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

func click() {
	if isWayland() {
		exec.Command("ydotool", "click", "0xC0").Run()
	} else {
		robotgo.Click()
	}
}

func cursorPos() (int, int) {
	if isWayland() {
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
	return robotgo.Location()
}

func getPixelColor(x, y int) (int, int, int, error) {
	if isWayland() {
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

	hex := robotgo.GetPixelColor(x, y)
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int(r), int(g), int(b), nil
}
