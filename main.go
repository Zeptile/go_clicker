package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

var running = false
var config Configuration
var targetColors []TargetColor
var colorMu sync.RWMutex

func main() {
	config = parseConsoleArguments()

	if config.colorsFile != "" {
		config.colorMode = true
		err := loadColorsFile(config.colorsFile)
		if err != nil {
			fmt.Printf("Error loading colors file: %v\n", err)
			os.Exit(1)
		}
	}

	sigToggle := make(chan os.Signal, 1)
	sigSample := make(chan os.Signal, 1)
	signal.Notify(sigToggle, syscall.SIGUSR1)
	signal.Notify(sigSample, syscall.SIGUSR2)

	wg := sync.WaitGroup{}
	wg.Add(3)

	// Click loop
	go func() {
		defer wg.Done()

		for {
			if !running {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			rand.Seed(time.Now().UnixNano())
			var waitForMs int

			if config.randomMode {
				waitForMs = rand.Intn(int(config.randomIntervalEnd))
			} else {
				waitForMs = int(config.intervalMs)
			}

			time.Sleep(time.Duration(waitForMs) * time.Millisecond)

			if config.colorMode {
				matched, debugInfo := isColorMatch()
				if !matched {
					if config.debugMode {
						fmt.Println("[Debug] Color mismatch, skipping click.", debugInfo)
					}
					continue
				}
			}

			click()
			if config.debugMode {
				fmt.Printf("[Debug] Clicked after waiting %dms\n", waitForMs)
			}
		}
	}()

	// Toggle on SIGUSR1
	go func() {
		defer wg.Done()

		for range sigToggle {
			if !running {
				fmt.Println("Resumed")
			} else {
				fmt.Println("Paused")
			}
			running = !running
		}
	}()

	// Sample color on SIGUSR2
	go func() {
		defer wg.Done()

		for range sigSample {
			x, y := cursorPos()
			r, g, b, err := getPixelColor(x, y)
			if err != nil {
				fmt.Printf("Error sampling color: %v\n", err)
				continue
			}

			tol := config.colorTolerance
			tc := TargetColor{R: r, G: g, B: b, Tolerance: tol, Sampled: true}

			colorMu.Lock()
			targetColors = append(targetColors, tc)
			colorMu.Unlock()

			fmt.Printf("Sampled color at (%d, %d): R:%d G:%d B:%d tol:%d [%d colors loaded]\n",
				x, y, r, g, b, tol, len(targetColors))
		}
	}()

	fmt.Printf("PID: %d\n", os.Getpid())
	fmt.Println("Toggle:       kill -USR1", os.Getpid())
	fmt.Println("Sample color: kill -USR2", os.Getpid())

	if config.colorMode {
		fmt.Println("Color mode enabled (default tolerance:", config.colorTolerance, ")")
		if config.colorsFile != "" {
			fmt.Printf("Loaded %d colors from %s\n", len(targetColors), config.colorsFile)
		}
	}

	wg.Wait()
}

// cursorPos gets the cursor position via hyprctl
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

// getPixelColor captures a 1x1 pixel via grim and returns RGB
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

// click sends a left click via ydotool
func click() {
	exec.Command("ydotool", "click", "0xC0").Run()
}

func loadColorsFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cf ColorsFile
	err = json.Unmarshal(data, &cf)
	if err != nil {
		return err
	}

	colorMu.Lock()
	defer colorMu.Unlock()

	for _, c := range cf.Colors {
		tol := c.Tolerance
		if tol <= 0 {
			tol = config.colorTolerance
		}
		targetColors = append(targetColors, TargetColor{
			R: c.R, G: c.G, B: c.B,
			Tolerance: tol,
			Sampled:   true,
		})
	}

	return nil
}

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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func parseConsoleArguments() Configuration {
	var config Configuration

	pflag.BoolVar(&config.randomMode, "random", false, "Flag to enable random mode, click interval is random between 0 and randomIntervalEnd")
	pflag.BoolVar(&config.debugMode, "debug", false, "Flag to enable debug mode, prints debug logs to console")
	pflag.BoolVar(&config.colorMode, "color", false, "Flag to enable color mode, only clicks when pixel under cursor matches target color")

	pflag.Int64Var(&config.intervalMs, "intervalMs", 50, "Interval in milliseconds to click in normal mode")
	pflag.Int64Var(&config.randomIntervalEnd, "randomIntervalEnd", 100, "Interval treshold in milliseconds to click in random mode")
	pflag.IntVar(&config.colorTolerance, "colorTolerance", 30, "RGB tolerance per channel for color matching (0-255)")
	pflag.StringVar(&config.colorsFile, "colorsFile", "", "Path to colors.json file generated by colorpicker tool")

	pflag.Parse()

	if config.randomMode {
		if config.randomIntervalEnd < 0 {
			fmt.Println("Random interval end must be greater than 0")
			pflag.PrintDefaults()
			panic("Random interval end must be greater than 0")
		}
	} else {
		if config.intervalMs < 0 {
			fmt.Println("Interval must be greater than 0")
			pflag.PrintDefaults()
			panic("Interval must be greater than 0")
		}
	}

	return config
}
