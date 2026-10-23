package rules

import (
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/words"
)

// lines is a script of three rows: a building's name, a tech's description
// and an event's name. The tests that read it say what they expect of those
// three alone.
var lines = map[string]string{
	"building.hut.name":                 "Crash Pad",
	"tech.language.description":         "Words go brrr.",
	"event.bountiful_harvest.name":      "Harvest Drop",
	"building.nonesuch.name":            "No such building",
	"tech.language.nonesuch":            "No such field",
	"lexicon.primitive_age.workers.one": "Grinder",
}

// TestVibedReplacesExactlyTheRowsPresent: the overlay changes the text of
// the rows the script has and nothing else. Every other text is the plain
// one, and with the texts the script covers blanked, the two sources are
// equal in every number, key and list.
func TestVibedReplacesExactlyTheRowsPresent(t *testing.T) {
	defer vibe.UseScriptForTest(lines)()
	plain := FromConfig()
	vibed := Vibed(plain)

	pc, vc := words.Cells(plain), words.Cells(vibed)
	if len(pc) != len(vc) {
		t.Fatalf("the overlay changed how many texts there are: %d to %d", len(pc), len(vc))
	}
	changed := map[string]string{}
	for i := range pc {
		if pc[i].ID != vc[i].ID {
			t.Fatalf("the overlay moved a row: %s became %s", pc[i].ID, vc[i].ID)
		}
		if pc[i].Text != vc[i].Text {
			changed[pc[i].ID] = vc[i].Text
		}
	}
	want := map[string]string{
		"building.hut.name":            "Crash Pad",
		"tech.language.description":    "Words go brrr.",
		"event.bountiful_harvest.name": "Harvest Drop",
	}
	if !reflect.DeepEqual(changed, want) {
		t.Errorf("the overlay changed %v, want %v", changed, want)
	}

	blank := func(src Source) Source {
		out, _ := words.Rewrite(src, func(words.Cell) (string, bool) { return "-", true })
		return out
	}
	if !reflect.DeepEqual(blank(plain), blank(vibed)) {
		t.Error("the overlay changed something that is not a text a player reads")
	}
	for i, b := range vibed.Buildings {
		if b.Key == "hut" && b.Name != "Crash Pad" {
			t.Errorf("the hut is %q, want Crash Pad", b.Name)
		}
		if b.Key != "hut" && b.Name != plain.Buildings[i].Name {
			t.Errorf("%s is %q, was %q", b.Key, b.Name, plain.Buildings[i].Name)
		}
	}
}

// TestVibedLeavesTheTablesAlone: building a vibed source writes into nothing
// the plain tables share, so a game that turns VibeMode off gets its own
// words back and a second build starts from the plain ones.
func TestVibedLeavesTheTablesAlone(t *testing.T) {
	before := FromConfig()
	restore := vibe.UseScriptForTest(map[string]string{
		"building.hut.name":                "Crash Pad",
		"building.hut.description":         "A mid shelter.",
		"event.bountiful_harvest.name":     "Harvest Drop",
		"mission.primitive_age.place.01":   "the Grind",
		"badge_family.age_key.description": "x {name}",
		"mechanic.gather_amount.text":      "Everything goes %s",
	})
	got := map[string]bool{}
	for _, c := range words.Cells(Vibed(FromConfig())) {
		if _, ok := vibe.Text(c.ID); ok {
			got[c.ID] = true
		}
	}
	Vibed(FromConfig())
	restore()
	if len(got) != 6 {
		t.Errorf("the overlay found %d of the 6 rows (list, map and plain fields among them): %v", len(got), got)
	}
	if after := FromConfig(); !reflect.DeepEqual(before, after) {
		t.Error("building a vibed source changed a table of package config")
	}
}

// TestVibedKeepsTheNumbersInTheText: a building's description is the line
// followed by what the building does, and an era event's has its duration
// written in, as they are in plain English.
func TestVibedKeepsTheNumbersInTheText(t *testing.T) {
	defer vibe.UseScriptForTest(map[string]string{
		"building.hut.description": "A mid shelter.",
	})()
	plain, vibed := FromConfig(), Vibed(FromConfig())
	var p, v config.BuildingDef
	for i := range plain.Buildings {
		if plain.Buildings[i].Key == "hut" {
			p, v = plain.Buildings[i], vibed.Buildings[i]
		}
	}
	fx := strings.TrimPrefix(p.Description, "A crude shelter of sticks and leaves.")
	if fx == p.Description || strings.TrimSpace(fx) == "" {
		t.Fatalf("the hut's plain description is %q: the test needs the sentence of numbers after it", p.Description)
	}
	if v.Description != "A mid shelter."+fx {
		t.Errorf("the vibed description is %q, want the line then %q", v.Description, strings.TrimSpace(fx))
	}
}

// TestCoreForTheSwitch: the core set is the plain one until VibeMode is on,
// each built once; turning it off gives back the very same plain set, and
// the two differ in words alone.
func TestCoreForTheSwitch(t *testing.T) {
	vibe.Set(false)
	plain := Core()
	if plain != CoreFor(false) || !IsCore(plain) {
		t.Fatal("the core set is not the plain one while VibeMode is off")
	}
	vibe.Set(true)
	defer vibe.Set(false)
	on := Core()
	if on == plain || on != CoreFor(true) || !IsCore(on) {
		t.Fatal("the core set is not the VibeMode one while it is on")
	}
	vibe.Set(false)
	if Core() != plain {
		t.Error("turning VibeMode off did not give back the plain set")
	}
	if !IsCore(plain) || IsCore(Compile(FromConfig())) {
		t.Error("IsCore is wrong")
	}
	if on.Digest() == plain.Digest() {
		t.Error("the VibeMode set reads the same as the plain one: the shipped script changes no text")
	}
}
