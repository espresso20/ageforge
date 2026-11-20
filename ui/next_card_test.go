package ui

import (
	"fmt"
	"strings"
	"testing"
)

// panelNames is every panel the Panels list names, in table order.
func panelNames() []string {
	var out []string
	for _, g := range panelGroups {
		for _, p := range g.panels {
			out = append(out, p.name)
		}
	}
	return out
}

// Each form of the list is whole, fits its width and clicks through.
func TestPanelRowsForms(t *testing.T) {
	for _, c := range []struct{ form, T, rows int }{{formRich, 60, 12}, {formGrouped, 30, 10}, {formWords, 18, 0}} {
		rows := panelRows(c.form, c.T, panelView{})
		plain := untag(strings.Join(rows, "\n"))
		for _, r := range rows {
			if n := runeLen(untag(r)); n > c.T+1 {
				t.Errorf("form %d: a row of %d cells in %d: %q", c.form, n, c.T, untag(r))
			}
		}
		wantAll(t, fmt.Sprintf("form %d", c.form), plain, panelNames()...)
		if c.rows > 0 && len(rows) > c.rows {
			t.Errorf("form %d: %d rows, want at most %d", c.form, len(rows), c.rows)
		}
		if c.form == formRich {
			for _, g := range panelGroups {
				for _, p := range g.panels {
					if !strings.Contains(plain, p.job.String()) {
						t.Errorf("the rich form lost the job line %q", p.job.String())
					}
				}
			}
		}
		if !strings.Contains(strings.Join(rows, "\n"), `["research"]`) {
			t.Errorf("form %d: a panel's name is not a click target", c.form)
		}
	}
	// Too narrow for a form: the next one down.
	if got := panelRows(formRich, 40, panelView{}); strings.Contains(untag(strings.Join(got, "\n")), "the tech tree") {
		t.Error("a box 40 wide has job lines")
	}
}

// The card at 279x74 fits above the list, in the order the rows give way.
func TestNextBoxText(t *testing.T) {
	d, _, _ := medievalTown(t, 279, 74)
	d.refresh()
	s := d.next
	if s.title == "" || len(s.then) == 0 {
		t.Skipf("no card to lay out: %+v", s)
	}
	full := strings.Split(nextBoxText(s, 62, 40, true), "\n")
	if len(full) != 40 {
		t.Fatalf("the box text is %d rows, want 40", len(full))
	}
	plain := untag(strings.Join(full, "\n"))
	wantAll(t, "the card", plain, "NOW", s.title, "THEN", "Show me", "Tab", s.cmd, "Panels")
	// Shorter and shorter: the THEN items go, then the label, the form, the card.
	prev := 1 << 10
	for h := 40; h >= 8; h-- {
		text := untag(nextBoxText(s, 62, h, true))
		rows := strings.Split(text, "\n")
		if len(rows) > h {
			t.Errorf("height %d: %d rows", h, len(rows))
		}
		if n := strings.Count(text, "THEN"); n > 1 || (h < 20 && prev < 20 && strings.Contains(text, s.then[0].title) && !strings.Contains(text, "THEN")) {
			t.Errorf("height %d: THEN label and items out of step", h)
		}
		wantAll(t, fmt.Sprintf("height %d", h), text, panelNames()...)
		prev = h
	}
	// The classic layout: the list alone, in words, no card.
	classic := untag(nextBoxText(s, 22, 30, false))
	if strings.Contains(classic, "NOW") || strings.Contains(classic, "TOWN") {
		t.Errorf("the classic box has a card or groups:\n%s", classic)
	}
	wantAll(t, "classic", classic, panelNames()...)
	// No command: no button row.
	s.cmd = ""
	if strings.Contains(untag(nextBoxText(s, 62, 40, true)), "Show me") {
		t.Error("a step with no command has a Show me button")
	}
}

// A THEN item shows its panel word only when the panel is on the Panels list.
func TestThenPanelWord(t *testing.T) {
	s := nextStep{title: "Build a mill", then: []nextItem{
		{title: "Study the stars", detail: "one lesson", panel: "research"},
		{title: "Call the feast", detail: "one lesson", panel: "festival"},
	}}
	out := strings.Join(cardRows(s, 40, cardCut{then: 2}), "\n")
	if !strings.Contains(out, "research") {
		t.Errorf("a THEN item on a real panel lost its panel word:\n%s", out)
	}
	if strings.Contains(out, "festival") {
		t.Errorf("a main-view THEN item showed a panel word:\n%s", out)
	}
}
