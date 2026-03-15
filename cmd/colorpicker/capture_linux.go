//go:build linux

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
)

func captureRegion() (image.Image, error) {
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
