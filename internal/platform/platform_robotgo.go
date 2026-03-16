//go:build darwin || windows

package platform

import (
	"strconv"

	"github.com/go-vgo/robotgo"
)

func Click() {
	robotgo.Click()
}

func CursorPos() (int, int) {
	return robotgo.Location()
}

func GetPixelColor(x, y int) (int, int, int, error) {
	hex := robotgo.GetPixelColor(x, y)
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int(r), int(g), int(b), nil
}
