package color

import (
	"encoding/json"
	"os"
)

func LoadFile(path string, defaultTolerance int) ([]TargetColor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf ColorsFile
	err = json.Unmarshal(data, &cf)
	if err != nil {
		return nil, err
	}

	var colors []TargetColor
	for _, c := range cf.Colors {
		tol := c.Tolerance
		if tol <= 0 {
			tol = defaultTolerance
		}
		colors = append(colors, TargetColor{
			R: c.R, G: c.G, B: c.B,
			Tolerance: tol,
			Sampled:   true,
		})
	}

	return colors, nil
}
