package main

type Configuration struct {
	debugMode         bool
	randomMode        bool
	colorMode         bool
	intervalMs        int64
	randomIntervalStart int64
	randomIntervalEnd   int64
	colorTolerance    int
	colorsFile        string
}

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
}

type ColorsFile struct {
	Colors []ColorEntry `json:"colors"`
}