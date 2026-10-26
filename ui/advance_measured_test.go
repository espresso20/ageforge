package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// measuredEngine is a game under the measured gate (AGEFORGE_GATES=v2) in age.
func measuredEngine(t *testing.T, age string) *game.GameEngine {
	t.Helper()
	restoreForge(t)
	src := rules.FromConfig()
	src.GatesV2 = true
	eng := game.NewGameEngineWith(rules.Compile(src))
	eng.SeedRNG(12)
	for _, a := range eng.Rules().AgeKeys() {
		if eng.GetState().Age == age {
			break
		}
		if a == "primitive_age" {
			continue
		}
		if err := eng.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	if got := eng.GetState().Age; got != age {
		t.Fatalf("staged the %s, want the %s", got, age)
	}
	return eng
}

// The Advance panel writes the measured gate's rows as the gate asks for
// them: a population against the population, the main materials' income a
// tick, the keystone and the key tech.
func TestAdvancePanelShowsTheMeasures(t *testing.T) {
	for _, age := range []string{"stone_age", "bronze_age", "industrial_age", "quantum_age"} {
		eng := measuredEngine(t, age)
		set := eng.Rules()
		m, ok := set.GateMeasure(age)
		if !ok {
			t.Fatalf("%s: no measure", age)
		}
		st := eng.GetState()
		st.Workers.TotalPop = m.People / 2
		var names []string
		for _, res := range m.Materials {
			rs := st.Resources[res]
			rs.Rate = m.Income / 8 / float64(len(m.Materials))
			st.Resources[res] = rs
			names = append(names, set.Name(rules.KindResource, res))
		}
		v := st.AdvanceView()
		people := fmt.Sprintf("%s / %s", textfmt.Number(float64(m.People/2)), textfmt.Number(float64(m.People)))
		income := fmt.Sprintf("%s / %s a tick", textfmt.Number(m.Income/8), textfmt.Number(m.Income))
		for _, size := range [][2]int{{100, 30}, {160, 45}} {
			w, h := size[0], size[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			for sel, want := range map[int][]string{
				0: {"Key tech: " + set.Name(rules.KindTech, m.KeyTech)},
				1: {"Population", people},
				2: {strings.Join(names, " + ") + " income", income},
			} {
				text := renderAdvance(v, advanceUI{sel: sel}, w, h).String()
				for _, s := range want {
					if !strings.Contains(text, s) {
						t.Errorf("%s, pillar %d: no %q:\n%s", where, sel, s, text)
					}
				}
				if strings.ContainsAny(text, "—!") {
					t.Errorf("%s, pillar %d: a dash or an exclamation mark:\n%s", where, sel, text)
				}
			}
		}

		// The Ledger waits on the same measures.
		l := ledgerOf(st, eng)
		waits := strings.Join(l.data.Waits, "\n")
		for _, s := range []string{"Population", "income"} {
			if !strings.Contains(waits, s) {
				t.Errorf("%s: the Ledger waits on %q, not on %s", age, waits, s)
			}
		}

		// With the measures met the two pillars stand and show it.
		st.Workers.TotalPop = m.People
		for _, res := range m.Materials {
			rs := st.Resources[res]
			rs.Rate = m.Income / float64(len(m.Materials))
			st.Resources[res] = rs
		}
		v = st.AdvanceView()
		if !v.Pillars[1].Met || !v.Pillars[2].Met {
			t.Errorf("%s: with the measures met the People and Industry pillars read %v and %v", age, v.Pillars[1].Met, v.Pillars[2].Met)
		}
		if text := renderAdvance(v, advanceUI{sel: 1}, 100, 30).String(); !strings.Contains(text, "✓ Population") {
			t.Errorf("%s: a met population is not ticked:\n%s", age, text)
		}
	}
}

// Under today's gate the panel and the Ledger carry no measured rows.
func TestAdvancePanelWithoutTheSwitchHasNoMeasures(t *testing.T) {
	if config.GatesV2() {
		t.Skip("the test run itself started with AGEFORGE_GATES=v2")
	}
	restoreForge(t)
	eng := game.NewGameEngine()
	eng.SeedRNG(12)
	text := ""
	for sel := 0; sel < 4; sel++ {
		text += renderAdvance(eng.GetState().AdvanceView(), advanceUI{sel: sel}, 100, 30).String()
	}
	for _, s := range []string{"Population", "income", "Key tech"} {
		if strings.Contains(text, s) {
			t.Errorf("today's gate shows %q:\n%s", s, text)
		}
	}
}
