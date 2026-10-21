package smoke

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The One-Resource-Line check: a building line never switches resource. The
// six lineages whose resource changes partway are split into lines, one per
// resource, and every building makes its line's resource (a production
// effect, main or side output). The split is written out again here from the
// table, by age number, not read back from the line keys the config sets.

// lineSpan is a run of ages (1 = Primitive ... 22 = Transcendent), both ends in.
type lineSpan struct{ from, to int }

// lineTable lists, per split lineage, the resource each line makes and the
// ages whose buildings belong to it. A line may pause and resume.
var lineTable = map[string][]struct {
	res   string
	spans []lineSpan
}{
	"organic_extraction": {
		{"wood", []lineSpan{{1, 6}}}, {"coal", []lineSpan{{7, 9}}}, {"oil", []lineSpan{{10, 14}}},
		{"nanobots", []lineSpan{{15, 17}}}, {"quantum_flux", []lineSpan{{18, 21}}},
	},
	"geological_extraction": {
		{"stone", []lineSpan{{2, 5}}}, {"iron_ore", []lineSpan{{6, 9}}}, {"uranium", []lineSpan{{10, 12}}},
		{"titanium_ore", []lineSpan{{13, 15}}}, {"dark_matter_crystals", []lineSpan{{16, 18}}}, {"antimatter", []lineSpan{{19, 21}}},
	},
	"engineering": {
		{"iron", []lineSpan{{3, 6}}}, {"steel", []lineSpan{{7, 10}}}, {"electricity", []lineSpan{{11, 15}}},
		{"plasma", []lineSpan{{16, 18}}}, {"dark_matter", []lineSpan{{19, 20}}}, {"quantum_flux", []lineSpan{{21, 22}}},
	},
	"metallurgy": {
		{"iron", []lineSpan{{4, 6}}}, {"steel", []lineSpan{{7, 12}}}, {"titanium", []lineSpan{{13, 15}, {18, 18}}},
		{"dark_matter", []lineSpan{{16, 17}}}, {"antimatter", []lineSpan{{19, 20}}}, {"quantum_flux", []lineSpan{{21, 21}}},
	},
	"energy": {
		{"coal", []lineSpan{{9, 9}}}, {"electricity", []lineSpan{{10, 12}, {14, 16}}}, {"oil", []lineSpan{{13, 13}}},
		{"plasma", []lineSpan{{17, 19}}}, {"dark_matter", []lineSpan{{20, 20}}}, {"quantum_flux", []lineSpan{{21, 21}}},
	},
	"trade": {
		{"gold", []lineSpan{{3, 15}, {17, 22}}}, {"crypto", []lineSpan{{16, 16}}},
	},
}

// LineCount is how many lines the building lineages make in all: the eight
// that never change resource count as one each.
const LineCount = 39

// LineProblem is one way the lines break the rule.
type LineProblem struct{ Building, Age, Line, Why string }

func (p LineProblem) String() string {
	return fmt.Sprintf("%s (%s, line %s): %s", p.Building, p.Age, p.Line, p.Why)
}

// StaticLines lists every building whose line, by the table, is not the one
// the config gives it, or that does not make its line's resource.
func StaticLines() []LineProblem {
	ageNo := map[string]int{}
	for i, a := range config.Ages() {
		ageNo[a.Key] = i + 1
	}
	makes := func(d config.BuildingDef, res string) bool {
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == res && e.Value > 0 {
				return true
			}
		}
		return false
	}
	var out []LineProblem
	lines := map[string][]config.BuildingDef{}
	for _, d := range config.BaseBuildings() {
		switch d.LineageKey {
		case "", "storage", "wonder", "monument":
			continue
		}
		lines[d.Line()] = append(lines[d.Line()], d)
		rows, split := lineTable[d.LineageKey]
		if !split {
			if d.LineKey != "" {
				out = append(out, LineProblem{d.Key, d.RequiredAge, d.Line(), "its lineage never changes resource, so it has no line key of its own"})
			}
			continue
		}
		age := ageNo[d.RequiredAge]
		want := ""
		for _, row := range rows {
			for _, s := range row.spans {
				if age >= s.from && age <= s.to {
					if want != "" {
						out = append(out, LineProblem{d.Key, d.RequiredAge, d.Line(), "the table gives this age two resources: " + want + " and " + row.res})
					}
					want = row.res
				}
			}
		}
		if want == "" {
			out = append(out, LineProblem{d.Key, d.RequiredAge, d.Line(), "the table has no resource for this age"})
			continue
		}
		if got := d.LineageKey + "_" + want; d.Line() != got {
			out = append(out, LineProblem{d.Key, d.RequiredAge, d.Line(), "the table puts it in " + got})
		}
		if !makes(d, want) {
			out = append(out, LineProblem{d.Key, d.RequiredAge, d.Line(), "makes no " + want})
		}
	}
	// A line that never changes resource: the buildings that make anything
	// share a resource, and none of them makes nothing.
	for line, defs := range lines {
		if _, split := lineTable[defs[0].LineageKey]; split {
			continue
		}
		var common map[string]bool
		for _, d := range defs {
			got := map[string]bool{}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 {
					got[e.Target] = true
				}
			}
			if common == nil {
				common = got
				continue
			}
			for res := range common {
				if !got[res] {
					delete(common, res)
				}
			}
		}
		if len(common) == 0 && anyProduces(defs) {
			out = append(out, LineProblem{defs[0].Key, defs[0].RequiredAge, line, "its buildings make different resources and no one resource in common"})
		}
	}
	if len(lines) != LineCount {
		names := make([]string, 0, len(lines))
		for l := range lines {
			names = append(names, l)
		}
		sort.Strings(names)
		out = append(out, LineProblem{"-", "-", "-", fmt.Sprintf("%d lines, want %d: %s", len(lines), LineCount, strings.Join(names, ", "))})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Building < out[j].Building })
	return out
}

func anyProduces(defs []config.BuildingDef) bool {
	for _, d := range defs {
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 {
				return true
			}
		}
	}
	return false
}
