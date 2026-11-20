package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// A click on a button whose command the table marks Dangerous asks before it
// runs, whether or not the button brings a question of its own.
func TestFireAsksBeforeADangerousCommand(t *testing.T) {
	var ran, asked []string
	k := &keyPresser{press: func(kit.Key) {}, run: func(c string) { ran = append(ran, c) },
		confirm: func(q, yes, c string) { asked = append(asked, q+"|"+yes+"|"+c) }}
	k.fire(xkey{cmd: "sell hut"})
	if len(ran) != 0 || len(asked) != 1 || !strings.Contains(asked[0], "sell hut") || !strings.HasSuffix(asked[0], "|sell hut") {
		t.Errorf("a Dangerous line with no question: ran %q, asked %q", ran, asked)
	}
	asked = nil
	k.fire(xkey{cmd: "build hut"})
	if len(ran) != 1 || len(asked) != 0 {
		t.Errorf("a plain line: ran %q, asked %q", ran, asked)
	}
	k.fire(xkey{cmd: "sell hut", ask: "Sell one?", yes: "Sell"})
	if len(asked) != 1 || asked[0] != "Sell one?|Sell|sell hut" {
		t.Errorf("a button's own question was replaced: %q", asked)
	}
}

// The dashboard's Sell: the click opens the question and sells nothing; Cancel
// leaves the town as it was; the danger button sells once.
func TestSellByClickAsksFirst(t *testing.T) {
	keepTheme(t)
	d, st, eng := medievalTown(t, 160, 58)
	box := d.economyTab.buildBox
	setStock(eng, 1e6)
	d.refresh()
	st.paint()
	key := box.order[0]
	eng.SetTownForTest(map[string]int{key: 12}, 0)
	d.refresh()
	st.paint()
	pickCard(t, d, st, key)
	var sell xkey
	for _, k := range box.det.keys {
		if strings.HasPrefix(k.cmd, "sell ") {
			sell = k
		}
	}
	if sell.cmd == "" || sell.off {
		t.Fatalf("Sell is %+v, want a key that is on", sell)
	}
	if !strings.Contains(sell.ask, box.detail.bs.Name) {
		t.Errorf("the question is %q, want it to name the building", sell.ask)
	}
	owned := func() int { return eng.GetState().Buildings[key].Count }
	front := func() *nativeDialog {
		name, item := d.pages.GetFrontPage()
		dlg, _ := item.(*nativeDialog)
		if name != "ask_line" {
			return nil
		}
		return dlg
	}
	for round, danger := range []bool{false, true} {
		p := centre(d, sell.r)
		st.click(p.X, p.Y)
		dlg := front()
		if dlg == nil {
			t.Fatalf("round %d: the click opened no question", round)
		}
		if owned() != 12 {
			t.Fatalf("round %d: the click sold before the answer", round)
		}
		d.app.SetFocus(dlg)
		if danger {
			keyTo(dlg, tk(kit.KeyTab, 0, 0))
		}
		keyTo(dlg, tk(kit.KeyEnter, 0, 0))
		want := 12
		if danger {
			want = 11
		}
		if got := owned(); got != want {
			t.Errorf("round %d (danger %v): %d huts left, want %d", round, danger, got, want)
		}
	}
}
