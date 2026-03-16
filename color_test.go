package main

import (
	"testing"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-1, 1},
	}
	for _, tt := range tests {
		got := abs(tt.input)
		if got != tt.expected {
			t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestMatchesAnyColor_ExactMatch(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, _ := matchesAnyColor(100, 150, 200, colors)
	if !matched {
		t.Error("expected exact match to return true")
	}
}

func TestMatchesAnyColor_WithinTolerance(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, _ := matchesAnyColor(129, 179, 171, colors)
	if !matched {
		t.Error("expected match within tolerance (diff=29 per channel)")
	}
}

func TestMatchesAnyColor_OutsideTolerance(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, debugInfo := matchesAnyColor(200, 150, 200, colors)
	if matched {
		t.Error("expected no match when R diff=100 exceeds tolerance=30")
	}
	if debugInfo == "" {
		t.Error("expected debug info with closest color details")
	}
}

func TestMatchesAnyColor_PerChannelTolerance(t *testing.T) {
	colors := []TargetColor{{R: 100, G: 150, B: 200, Tolerance: 30}}
	matched, _ := matchesAnyColor(131, 150, 200, colors)
	if matched {
		t.Error("expected no match when one channel exceeds tolerance")
	}
}

func TestMatchesAnyColor_MultipleColors_MatchesSecond(t *testing.T) {
	colors := []TargetColor{
		{R: 10, G: 10, B: 10, Tolerance: 5},
		{R: 200, G: 200, B: 200, Tolerance: 10},
	}
	matched, _ := matchesAnyColor(205, 195, 200, colors)
	if !matched {
		t.Error("expected match against second target color")
	}
}

func TestMatchesAnyColor_EmptyColors(t *testing.T) {
	matched, debugInfo := matchesAnyColor(100, 100, 100, nil)
	if matched {
		t.Error("expected no match with nil color list")
	}
	if debugInfo != "" {
		t.Error("expected empty debug info with nil color list")
	}
}
