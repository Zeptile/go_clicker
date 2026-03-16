package color

import "fmt"

func MatchesAnyColor(r, g, b int, tc []TargetColor) (bool, string) {
	closest := ""
	minDist := 999999
	for _, t := range tc {
		dr, dg, db := Abs(r-t.R), Abs(g-t.G), Abs(b-t.B)
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

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
