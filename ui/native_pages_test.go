package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/rivo/tview"
)

// A nativePages keeps its pages, and moves the focus, exactly as tview's
// Pages does, whatever is done to it: the same run of adds, removes, shows,
// hides and switches leaves both with the same pages in the same order, the
// same page in front, and the keyboard on the same one.
func TestNativePagesBehaveAsTviewPagesDo(t *testing.T) {
	names := []string{"a", "b", "c", "d"}
	for seed := int64(1); seed <= 40; seed++ {
		rng := rand.New(rand.NewSource(seed))
		old, now := tview.NewPages(), newNativePages()
		oldApp, nowApp := tview.NewApplication(), newAppLoop()
		// One item a name, the same on every add, so "which has the focus"
		// can be told by name on both sides.
		oldItems, nowItems := map[string]*tview.Box{}, map[string]*nativeBox{}
		for _, n := range names {
			oldItems[n], nowItems[n] = tview.NewBox(), newNativeBox()
		}
		oldApp.SetFocus(old)
		nowApp.SetFocus(now)
		var log []string
		state := func() string {
			var focus [2]string
			for _, n := range names {
				if oldItems[n].HasFocus() {
					focus[0] = n
				}
				if nowItems[n].HasFocus() {
					focus[1] = n
				}
			}
			of, _ := old.GetFrontPage()
			nf, _ := now.GetFrontPage()
			a := fmt.Sprint(old.GetPageNames(false), old.GetPageNames(true), of, focus[0], old.HasFocus(), old.GetPageCount())
			b := fmt.Sprint(now.GetPageNames(false), now.GetPageNames(true), nf, focus[1], now.HasFocus(), now.GetPageCount())
			if a != b {
				return fmt.Sprintf("tview %s, native %s", a, b)
			}
			return ""
		}
		for step := 0; step < 60; step++ {
			n := names[rng.Intn(len(names))]
			switch op := rng.Intn(7); op {
			case 0, 1:
				vis := rng.Intn(4) > 0
				log = append(log, fmt.Sprintf("add %s %v", n, vis))
				old.AddPage(n, oldItems[n], true, vis)
				now.AddPage(n, nowItems[n], true, vis)
			case 2:
				log = append(log, "remove "+n)
				old.RemovePage(n)
				now.RemovePage(n)
			case 3:
				log = append(log, "show "+n)
				old.ShowPage(n)
				now.ShowPage(n)
			case 4:
				log = append(log, "hide "+n)
				old.HidePage(n)
				now.HidePage(n)
			case 5:
				log = append(log, "switch "+n)
				old.SwitchToPage(n)
				now.SwitchToPage(n)
			case 6:
				log = append(log, "front "+n)
				old.SendToFront(n)
				now.SendToFront(n)
			}
			if old.HasPage(n) != now.HasPage(n) || (old.GetPage(n) == nil) != (now.GetPage(n) == nil) {
				t.Fatalf("seed %d: after %s the two disagree on having %q", seed, strings.Join(log, ", "), n)
			}
			if diff := state(); diff != "" {
				t.Fatalf("seed %d: after %s: %s", seed, strings.Join(log, ", "), diff)
			}
		}
	}
}
