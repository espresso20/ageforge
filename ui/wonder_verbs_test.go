package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// wonderStage is the Wonders panel's buttons over the snapshot after mut,
// with what they run collected.
type wonderStage struct {
	wv   *wondersView
	ran  []string
	rect image.Rectangle
	key  string
}

func stageWonder(t *testing.T, mut func(st *game.GameState, key string)) *wonderStage {
	t.Helper()
	_, _, eng := stagedDashboard(t, "iron_age")
	st := eng.GetState()
	bs := make(map[string]game.BuildingState, len(st.Buildings))
	for k, v := range st.Buildings {
		bs[k] = v
	}
	st.Buildings = bs
	res := make(map[string]game.ResourceState, len(st.Resources))
	for k, v := range st.Resources {
		res[k] = v
	}
	st.Resources = res
	key := currentWonder(st).key
	if mut != nil {
		mut(&st, key)
	}
	s := &wonderStage{wv: &wondersView{}, rect: image.Rect(2, 1, 98, 29), key: key}
	s.wv.pr.press = func(kit.Key) {}
	s.wv.pr.run = func(c string) { s.ran = append(s.ran, c) }
	s.wv.provider(st, 96)
	s.lay()
	return s
}

func (s *wonderStage) lay() {
	s.wv.grid, s.wv.org = s.wv.layer(s.rect, 0), s.rect.Min
}

// button is the button that carries cmd.
func (s *wonderStage) button(t *testing.T, cmd string) xkey {
	t.Helper()
	for _, k := range s.wv.grid.keyset().keys {
		if k.cmd == cmd {
			return k
		}
	}
	t.Fatalf("no button carries %q", cmd)
	return xkey{}
}

func (s *wonderStage) click(k xkey) {
	pt := s.rect.Min.Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
	for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
		s.wv.mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
}

func typeKey(wv *wondersView, r rune) { wv.handleKey(&kit.Key{Code: kit.KeyRune, Rune: r}) }

func fullBank(st *game.GameState, key string, canBuild bool) {
	b := st.Buildings[key]
	b.WonderBankFull, b.CanBuild, b.NeedsTech = true, canBuild, ""
	st.Buildings[key] = b
}

func TestWonderBuildRunsOnceByClickAndKeyAndIsOffWithoutAFullBank(t *testing.T) {
	s := stageWonder(t, func(st *game.GameState, key string) { fullBank(st, key, true) })
	build := s.button(t, buildLine(s.wv.blk.buildCmd[len("build "):]))
	if build.off {
		t.Fatal("Build is off with a full bank")
	}
	s.click(build)
	typeKey(s.wv, 'b')
	typeKey(s.wv, 'B')
	if len(s.ran) != 3 || s.ran[0] != build.cmd || s.ran[1] != build.cmd || s.ran[2] != build.cmd {
		t.Errorf("Build ran %q, want %q three times (click, b, B)", s.ran, build.cmd)
	}
	off := stageWonder(t, nil)
	b := off.button(t, off.wv.blk.buildCmd)
	if !b.off {
		t.Fatal("Build is on with an empty bank")
	}
	off.click(b)
	typeKey(off.wv, 'b')
	if len(off.ran) != 0 {
		t.Errorf("an off Build ran %q", off.ran)
	}
}

func TestWonderCollectIsOffWithNothingToBank(t *testing.T) {
	s := stageWonder(t, func(st *game.GameState, key string) {
		for r := range st.Resources {
			v := st.Resources[r]
			v.Amount = 100
			st.Resources[r] = v
		}
	})
	c := s.button(t, wonderCollectCmd)
	if c.off {
		t.Fatal("Collect is off with stock on hand")
	}
	s.click(c)
	typeKey(s.wv, 'c')
	if len(s.ran) != 2 || s.ran[0] != wonderCollectCmd || s.ran[1] != wonderCollectCmd {
		t.Errorf("Collect ran %q, want it twice (click, c)", s.ran)
	}
	for name, mut := range map[string]func(st *game.GameState, key string){
		"no stock": func(st *game.GameState, _ string) {
			for r := range st.Resources {
				v := st.Resources[r]
				v.Amount = 0
				st.Resources[r] = v
			}
		},
		"a full bank": func(st *game.GameState, key string) { fullBank(st, key, true) },
	} {
		o := stageWonder(t, mut)
		k := o.button(t, wonderCollectCmd)
		if !k.off {
			t.Errorf("%s: Collect is on", name)
		}
		o.click(k)
		typeKey(o.wv, 'c')
		if len(o.ran) != 0 {
			t.Errorf("%s: an off Collect ran %q", name, o.ran)
		}
	}
}

func TestWonderFullBankNoLongerTypesTheBuildCommand(t *testing.T) {
	s := stageWonder(t, func(st *game.GameState, key string) { fullBank(st, key, true) })
	if got := s.wv.blk.text; contains(got, "Build it with") {
		t.Errorf("the summary still types the build command:\n%s", got)
	}
}

func contains(s, sub string) bool { return len(sub) > 0 && indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// With the mouse alone, through the dashboard: bank a wonder's price, then
// start it.
func TestWonderBankAndBuildByMouseAlone(t *testing.T) {
	d, pages, eng := stagedDashboard(t, "iron_age")
	st := eng.GetState()
	key := currentWonder(st).key
	def, _ := st.Ruleset().Building(key)
	for res, need := range def.BaseCost {
		eng.SetStockForTest(res, need*2)
	}
	pressAt := func(k xkey) {
		tv := d.overlayMgr.entries["wonders"].tv
		text := tv.view.Bounds()
		pt := text.Min.Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
		for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
			tv.Mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
		}
	}
	show := func() {
		d.overlayMgr.screenW = 100
		d.overlayMgr.Show("wonders", eng.GetState())
		drawPrimitive(t, pages, 100, 40)
	}
	show()
	wv := d.wonders
	pressAt(wv.grid.keyset().keys[0])
	if !eng.GetState().Buildings[key].WonderBankFull {
		t.Fatal("Collect did not bank the wonder's price")
	}
	show()
	for _, k := range wv.grid.keyset().keys {
		if k.cmd == wv.blk.buildCmd {
			if k.off {
				t.Fatal("Build is off with a full bank")
			}
			pressAt(k)
		}
	}
	eng.StepTicks(2)
	if b := eng.GetState().Buildings[key]; b.Count == 0 && len(eng.GetState().BuildQueue) == 0 {
		t.Errorf("Build did not start the wonder: %+v", b)
	}
}
