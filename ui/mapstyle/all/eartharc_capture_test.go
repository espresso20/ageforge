//go:build mapcapture

package all

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// earthAges are the Earth arc's ages: the Modern Age through the Fusion Age.
var earthAges = []string{"modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"}

// earthShot is one Earth-arc review view.
type earthShot struct {
	name, style string
	tod         float64 // < 0: the state's own tick
	compact     bool
	keys        []*tcell.EventKey
}

var earthShots = []earthShot{
	{name: "roguelike", style: "roguelike", tod: -1},
	{name: "roguelike_region", style: "roguelike", tod: -1, keys: keys(key(tcell.KeyPgUp))},
	{name: "roguelike_mini", style: "roguelike", tod: -1, compact: true},
	{name: "skyline_day", style: "skyline", tod: 0.45},
	{name: "skyline_night", style: "skyline", tod: 0.95},
	{name: "skyline_mini", style: "skyline", tod: 0.95, compact: true},
}

// TestWriteEarthArcCaptures draws the Earth arc's five ages in both styles
// from the smoke-bot states (run TestGenerateStates in ui/mapstyle/capture
// first) into map_captures/earth_arc/<phase>, where EARTH_ARC_PHASE names
// the phase ("before" or "after", default "after"). Each view is an
// animated .html page (8 frames), a .txt and a .png thumbnail. The run
// then writes contact.html, which puts every age's before and after side
// by side for both styles, from whatever phases are on disk.
//
//	EARTH_ARC_PHASE=before go test -tags mapcapture -run TestWriteEarthArcCaptures ./ui/mapstyle/all
func TestWriteEarthArcCaptures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "map_captures")
	states := envOr("MAP_STATES_DIR", filepath.Join(root, "states"))
	base := envOr("EARTH_ARC_DIR", filepath.Join(root, "earth_arc"))
	phase := envOr("EARTH_ARC_PHASE", "after")
	out := filepath.Join(base, phase)
	for _, d := range []string{"png", "crop"} {
		if err := os.MkdirAll(filepath.Join(out, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	if err := theme.SetActive("forge"); err != nil {
		t.Fatal(err)
	}
	b := mapmodel.NewBuilder(nil)
	for _, age := range earthAges {
		st, err := capture.LoadState(filepath.Join(states, age+".json.gz"))
		if err != nil {
			t.Logf("skip %s: %v", age, err)
			continue
		}
		var since *mapmodel.Visit
		if prev, err := capture.LoadState(filepath.Join(states, "prev_"+age+".json.gz")); err == nil {
			since = mapmodel.VisitOf(b.Build(&prev, nil))
		}
		for _, sh := range earthShots {
			s := st
			if sh.tod >= 0 {
				s.Tick = clockTick(mapmodel.ClockAt(st.Tick).Day, sh.tod)
			}
			m := mapmodel.NewBuilder(b.Catalog()).Build(&s, since)
			w, h := 160, 48
			if sh.compact {
				w, h = 40, 15
			}
			earthShoot(t, out, short(age)+"_"+sh.name, sh, m, w, h)
		}
	}
	earthContact(t, base)
}

func earthShoot(t *testing.T, out, name string, sh earthShot, m *mapmodel.Model, w, h int) {
	t.Helper()
	view, _ := Registry().New(sh.style)
	scr := capture.NewScreen(w, h)
	defer scr.Fini()
	r := mapstyle.Rect{W: w, H: h}
	f := mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode}
	draw := func() {
		if sh.compact {
			view.DrawCompact(scr, r, f)
		} else {
			view.Draw(scr, r, f)
		}
	}
	draw()
	for _, k := range sh.keys {
		view.HandleKey(k, f)
		draw()
	}
	var frames []string
	var txt strings.Builder
	for i := 0; i < 8; i++ {
		f.Anim = 600 + i*3
		draw()
		frames = append(frames, capture.Frame(scr))
		if i == 0 {
			if err := capture.PNG(filepath.Join(out, "png", name+".png"), scr); err != nil {
				t.Fatal(err)
			}
			txt.WriteString(capture.Text(scr))
			if w > 80 { // halves at full size, for a close look
				for k, x0 := range []int{0, w / 2} {
					half := capture.NewScreen(w/2, h)
					for y := 0; y < h; y++ {
						for x := 0; x < w/2; x++ {
							r, comb, st, _ := scr.GetContent(x0+x, y)
							half.SetContent(x, y, r, comb, st)
						}
					}
					bg := capture.Hex(theme.Color(theme.RoleBackground))
					page := capture.HTML(name+[2]string{" west", " east"}[k], bg, bg, capture.Fonts, []string{capture.Frame(half)}, 1)
					if err := os.WriteFile(filepath.Join(out, "crop", name+[2]string{"_w", "_e"}[k]+".html"), []byte(page), 0o644); err != nil {
						t.Fatal(err)
					}
					half.Fini()
				}
			}
		}
	}
	title := fmt.Sprintf("%s · %s · %dx%d", name, m.AgeName, w, h)
	bg := capture.Hex(theme.Color(theme.RoleBackground))
	fg := capture.Hex(theme.Color(theme.RoleText))
	if err := os.WriteFile(filepath.Join(out, name+".html"), []byte(capture.HTML(title, bg, fg, capture.Fonts, frames, 6)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, name+".txt"), []byte(title+"\n"+txt.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// earthContact writes contact.html: one row per age, each view's before and
// after thumbnails side by side, linked to their animated pages.
func earthContact(t *testing.T, base string) {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>Earth arc contact sheet</title><style>
body{background:#111;color:#ccc;font-family:ui-monospace,Menlo,monospace;margin:16px}
h1{font-size:16px;font-weight:normal}h2{font-size:14px;margin:18px 0 6px;color:#8cf}
table{border-collapse:collapse}td,th{padding:4px;vertical-align:top;text-align:left;font-weight:normal}
th{color:#999;font-size:12px}img{display:block;border:1px solid #333}
.wide img{width:300px}.mini img{width:200px}a{color:#8cf;font-size:11px}
</style></head><body><h1>Earth arc: Modern to Fusion, before and after</h1>
<p>Thumbnails show composition and color; click through for the real cell-by-cell render (8 animated frames).</p>
`)
	for _, sh := range earthShots {
		cls := "wide"
		if sh.compact {
			cls = "mini"
		}
		fmt.Fprintf(&b, "<h2>%s</h2><table class=%q><tr>", html.EscapeString(sh.name), cls)
		for _, age := range earthAges {
			fmt.Fprintf(&b, "<th>%s</th>", html.EscapeString(strings.TrimSuffix(age, "_age")))
		}
		b.WriteString("</tr>")
		for _, phase := range []string{"before", "after"} {
			b.WriteString("<tr>")
			for _, age := range earthAges {
				name := short(age) + "_" + sh.name
				png := filepath.Join(phase, "png", name+".png")
				if _, err := os.Stat(filepath.Join(base, png)); err != nil {
					b.WriteString("<td>(none)</td>")
					continue
				}
				fmt.Fprintf(&b, `<td><a href="%s/%s.html"><img src="%s" alt="%s %s"></a>%s</td>`,
					phase, name, png, phase, name, phase)
			}
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
	}
	b.WriteString("</body></html>\n")
	if err := os.WriteFile(filepath.Join(base, "contact.html"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
