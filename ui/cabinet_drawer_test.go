package ui

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/cabinet"
	"github.com/espresso20/ageforge/window/kit"
)

func trayRig(t *testing.T) (*Dashboard, *cabinetPanel, func() *kit.Buffer) {
	t.Helper()
	d, p, paint := cabinetRig(t)
	age := config.Ages()[0].Key
	d.engine.Account().StrikeCoin(config.CoinTechs, age, 10, 1, "run", time.Unix(1700000000, 0), false)
	p.update(d.engine.GetState())
	return d, p, paint
}

func press(p *cabinetPanel, code kit.KeyCode) bool { return p.routeKey(&kit.Key{Code: code}, "") }

func TestCabinetDrawerOpensAndCloses(t *testing.T) {
	d, p, paint := trayRig(t)
	if press(p, kit.KeyEnter) != true || p.tray.slot != 0 {
		t.Fatalf("Enter on a reached age left drawer %d", p.tray.slot)
	}
	if p.routeKey(&kit.Key{Code: kit.KeyEnter}, "text") && p.tray.slot != 0 {
		t.Error("Enter with text typed was taken")
	}
	b := paint()
	l, _ := cabinet.PlanTray(160, 44, 12, 24)
	if got := plateText(b, l.Labels[2]); got != "Techs" {
		t.Errorf("label under the techs well is %q", got)
	}
	for _, k := range []kit.KeyCode{kit.KeyRight, kit.KeyRight, kit.KeyUp, kit.KeyLeft, kit.KeyDown, kit.KeyEnd} {
		press(p, k)
	}
	if p.tray.well != cabinet.Wells-1 {
		t.Errorf("End left well %d", p.tray.well)
	}
	if !p.closeDrawer() || p.tray.slot != -1 || p.closeDrawer() {
		t.Error("closing the drawer")
	}
	press(p, kit.KeyEnter)
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonRight})
	if p.tray.slot != -1 {
		t.Error("right-click left the drawer open")
	}
	d.overlayMgr.Hide()
}

func TestCabinetDrawerCommandAndClicks(t *testing.T) {
	d, p, paint := trayRig(t)
	d.runForTest("badges stone age techs")
	if p.tray.slot != 1 || p.tray.well != 2 {
		t.Fatalf("badges stone age techs: drawer %d well %d", p.tray.slot, p.tray.well)
	}
	if res := HandleCommand("badges stone age zzz", d.engine); res.Type != "error" {
		t.Errorf("an unknown well: %+v", res)
	}
	d.runForTest("badges stone age")
	paint()
	l, _ := cabinet.PlanTray(160, 44, 12, 24)
	c := l.Wells[3].Min.Add(image.Pt(l.Wells[3].Dx()/2, l.Wells[3].Dy()/2))
	at := image.Pt(c.X/l.Cell.X, c.Y/l.Cell.Y)
	if p.PointerAt(at) != handIf(true) {
		t.Error("no hand over a well")
	}
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X, Y: at.Y})
	if p.tray.well != 3 {
		t.Errorf("a click selected well %d", p.tray.well)
	}
	n := p.tray.renders
	press(p, kit.KeyLeft)
	paint()
	paint()
	if p.tray.renders != n || n != 1 {
		t.Errorf("the tray was drawn %d times, %d after moves", n, p.tray.renders)
	}
	d.engine.Account().StrikeCoin(config.CoinAge, config.Ages()[1].Key, 10, 1, "run", time.Unix(1700000000, 0), false)
	p.update(d.engine.GetState())
	paint()
	if p.tray.renders != n+1 {
		t.Errorf("a new coin drew the tray %d times", p.tray.renders-n)
	}
}

func TestCabinetFanOpensStepsAndCloses(t *testing.T) {
	d, p, paint := trayRig(t)
	acc, age := d.engine.Account(), config.Ages()[0].Key
	for i, run := range []string{"run-a", "run-b", "run-c"} {
		acc.StrikeCoin(config.CoinTechs, age, 10+i*400, 1, run, time.Unix(1700000000+int64(i)*86400, 0), false)
	}
	p.update(d.engine.GetState())
	press(p, kit.KeyEnter)
	press(p, kit.KeyEnter) // well 0 holds nothing: no fan
	if p.tray.fan {
		t.Fatal("an empty stack opened a fan")
	}
	p.tray.well = 2
	press(p, kit.KeyEnter)
	if !p.tray.fan || p.pickCopy() != 0 {
		t.Fatalf("Enter on a stack: fan %v copy %d", p.tray.fan, p.pickCopy())
	}
	l, _ := cabinet.PlanTray(160, 44, 12, 24)
	b := paint()
	first, img := plateText(b, l.Ticket), p.tray.fanImg
	if img == nil || !strings.Contains(first, "Level") || !strings.Contains(first, "4 struck") {
		t.Fatalf("ticket %q, fan picture %v", first, img != nil)
	}
	press(p, kit.KeyLeft)
	if p.pickCopy() != 2 || paint() == nil || p.tray.pickOf != 2 {
		t.Errorf("Left from the middle chose copy %d", p.pickCopy())
	}
	press(p, kit.KeyRight)
	press(p, kit.KeyRight)
	press(p, kit.KeyRight)
	b = paint()
	if p.tray.pick != 3 || p.tray.well != 2 || p.tray.fanImg != img {
		t.Errorf("stepping: pick %d well %d, fan rebuilt %v", p.tray.pick, p.tray.well, p.tray.fanImg != img)
	}
	row2 := image.Rect(l.Ticket.Min.X, l.Ticket.Min.Y+1, l.Ticket.Max.X, l.Ticket.Max.Y)
	if got := plateText(b, row2); strings.Contains(got, "Run") || len(got) > l.Ticket.Dx() {
		t.Errorf("a run with no number, or a line past the face: %q", got)
	}
	p.tray.recs[2][p.pickCopy()].Run = "12"
	if _, w := p.ticketWords(); !strings.Contains(w, "Run 12") {
		t.Errorf("numbered run: %q", w)
	}
	if !strings.Contains(plateText(b, row2), "Figure") {
		t.Errorf("second ticket row %q", plateText(b, row2))
	}
	if !p.closeDrawer() || p.tray.fan || p.tray.slot != 0 {
		t.Error("Esc did not close the fan first")
	}
	if !p.closeDrawer() || p.tray.slot != -1 {
		t.Error("the second Esc did not close the drawer")
	}
	p.update(d.engine.GetState())
	b = paint()
	lay, _ := p.plan()
	if got := plateText(b, lay.Hint); !strings.Contains(got, "Enter") {
		t.Errorf("front hint %q", got)
	}
}

// An ageless drawer opens like an age drawer and lists one well per coin;
// the page keys turn its pages and nothing goes past either end.
func TestCabinetAgelessDrawerOpensAndPages(t *testing.T) {
	d, p, paint := trayRig(t)
	for j, key := range []string{"ledgers", "curios", "themes"} {
		slot := len(config.Ages()) + j
		p.sel = slot
		if !press(p, kit.KeyEnter) || p.tray.slot != slot {
			t.Fatalf("Enter on %s left drawer %d", key, p.tray.slot)
		}
		paint()
		ws := p.agelessWells(slot)
		if want := []int{74, 87, 16}[j]; len(ws) != want || p.drawers[slot].Total != want {
			t.Errorf("%s lists %d wells (drawer total %d), want %d", key, len(ws), p.drawers[slot].Total, want)
		}
		if !p.closeDrawer() || p.tray.slot != -1 {
			t.Errorf("%s did not close", key)
		}
	}
	p.sel = len(config.Ages())
	press(p, kit.KeyEnter)
	per, _ := p.gridDims()
	press(p, kit.KeyPgDn)
	if p.tray.well != per {
		t.Errorf("PgDn went to well %d, want %d", p.tray.well, per)
	}
	press(p, kit.KeyEnd)
	press(p, kit.KeyPgDn)
	if want := len(p.agelessWells(p.tray.slot)) - 1; p.tray.well != want {
		t.Errorf("past the last page the selection is %d, want %d", p.tray.well, want)
	}
	press(p, kit.KeyHome)
	press(p, kit.KeyPgUp)
	if p.tray.well != 0 {
		t.Errorf("PgUp on the first page left well %d", p.tray.well)
	}
	d.overlayMgr.Hide()
}

// Every plate of an ageless drawer fits the plate, and no two plates of Ledgers read alike.
func TestAgelessPlatesAreUniqueAndFit(t *testing.T) {
	_, p, _ := cabinetRig(t)
	for i := range p.keys {
		seen := map[string]bool{}
		for _, w := range p.agelessWells(i) {
			if n := len([]rune(w.name)); n > 30 {
				t.Errorf("%s plate %q is %d cells, the plate holds 30", p.keys[i], w.name, n)
			}
			if seen[w.name] && p.keys[i] == "ledgers" {
				t.Errorf("two Ledgers plates read %q", w.name)
			}
			seen[w.name] = true
		}
	}
}
