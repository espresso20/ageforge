package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// bigFigure finds a figure of four digits or more written against a unit:
// "7730Q", "11500Q", "1000Dc". The game prints three digits before a unit at
// most.
var bigFigure = regexp.MustCompile(`[0-9]{4,}[A-Za-z]`)

// TestLateGameStoresPrintInThreeDigits: the smallest store the game builds
// and the largest, at any count of copies up to well past the late game's
// top (about 25 quintillion), print in at most three digits before the unit,
// and the formatter's ceiling is far past them.
func TestLateGameStoresPrintInThreeDigits(t *testing.T) {
	smallest, largest := 0.0, 0.0
	for _, d := range config.BuildingByKey() {
		size := 0.0
		for _, e := range d.Effects {
			if e.Type == "storage" && e.Target == "all" {
				size += e.Value
			}
		}
		if size <= 0 || d.MaxCount > 0 {
			continue // not a store of its own (the monolith has one copy)
		}
		if smallest == 0 || size < smallest {
			smallest = size
		}
		largest = max(largest, size)
	}
	if got := FormatNumber(smallest); got != "100" {
		t.Errorf("the smallest store prints %q, want 100", got)
	}
	if got := FormatNumber(largest); got != "1.5Qi" {
		t.Errorf("the largest store prints %q, want 1.5Qi", got)
	}
	for copies := 1; copies <= 40; copies++ {
		for _, size := range []float64{smallest, largest} {
			if got := FormatNumber(size * float64(copies)); bigFigure.MatchString(got) {
				t.Errorf("%d copies of a %v store print %q", copies, size, got)
			}
		}
	}
	if got := FormatNumber(2.5e19); got != "25Qi" {
		t.Errorf("25 quintillion prints %q, want 25Qi", got)
	}
	if max := textfmt.MaxNumber(); max < largest*40*1e9 {
		t.Errorf("the formatter's ceiling %v is not three units past the late game's stores", max)
	}
}

// TestQuantumTownPrintsNoFourDigitFigure: the Resources box, the Buildings
// panel and the Next Age row of a Quantum Age town at full stores carry no
// figure of more than three digits against a unit, at either size.
func TestQuantumTownPrintsNoFourDigitFigure(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "quantum_age")
	for _, sz := range layoutSizes {
		w, h := sz[0], sz[1]
		drawnDashboard(t, d, pages, w, h)
		for name, text := range map[string]string{
			"Resources": d.economyTab.resourceTV.GetText(true),
			"Buildings": d.economyTab.buildingTV.GetText(true),
			"Next Age":  d.ageTV.GetText(true),
		} {
			if m := bigFigure.FindString(text); m != "" {
				t.Errorf("%s at %dx%d prints %q:\n%s", name, w, h, m, strings.TrimSpace(text))
			}
		}
	}
}
