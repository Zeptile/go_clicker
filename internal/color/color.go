package color

type TargetColor struct {
	R, G, B   int
	Tolerance int
	Sampled   bool
}

type ColorEntry struct {
	R         int `json:"r"`
	G         int `json:"g"`
	B         int `json:"b"`
	Tolerance int `json:"tolerance"`
	Count     int `json:"count,omitempty"`
}

type ColorsFile struct {
	Colors []ColorEntry `json:"colors"`
}
