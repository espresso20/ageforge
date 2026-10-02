//go:build mapcapture

package all

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// skyShotAges are the sky arc's ages, and the Fusion Age before them.
var skyShotAges = []string{"fusion_age", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}

// TestWriteSkyCaptures draws the sky arc in both styles from the real
// smoke-bot states (run TestGenerateStates in ui/mapstyle/capture first):
// each age full size (animated HTML, its first frame as a PNG, the text)
// and as the mini map, into SKY_CAPTURE_DIR (default map_captures/sky/after
// at the repo root, which git ignores). It uses only the public styles, so
// the same file run on an older checkout writes the "before" set; when
// SKY_BEFORE_DIR holds one, it also writes a contact sheet setting the two
// side by side (contact.html and contact.png in the parent folder).
//
//	go test -tags mapcapture -run TestWriteSkyCaptures ./ui/mapstyle/all
func TestWriteSkyCaptures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "map_captures")
	states := envOr("MAP_STATES_DIR", filepath.Join(root, "states"))
	out := envOr("SKY_CAPTURE_DIR", filepath.Join(root, "sky", "after"))
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	th := envOr("SKY_THEME", "forge")
	if err := theme.SetActive(th); err != nil {
		t.Fatal(err)
	}
	for _, age := range skyShotAges {
		st, err := capture.LoadState(filepath.Join(states, age+".json.gz"))
		if err != nil {
			t.Logf("skip %s: %v", age, err)
			continue
		}
		m := mapmodel.NewBuilder(nil).Build(&st, nil)
		for _, style := range Registry().Names() {
			for _, mini := range []bool{false, true} {
				w, h := 160, 48
				name := style + "_" + short(age)
				if mini {
					w, h, name = 40, 15, name+"_mini"
				}
				v, _ := Registry().New(style)
				scr := capture.NewScreen(w, h)
				var frames []string
				var txt strings.Builder
				for i := 0; i < 12; i++ {
					f := mapstyle.Frame{Model: m, Anim: 2000 + i*3, Tier: mapmodel.TierUnicode}
					if mini {
						v.DrawCompact(scr, mapstyle.Rect{W: w, H: h}, f)
					} else {
						v.Draw(scr, mapstyle.Rect{W: w, H: h}, f)
					}
					frames = append(frames, capture.Frame(scr))
					if i == 0 {
						txt.WriteString(capture.Text(scr))
						if err := capture.PNG(filepath.Join(out, name+".png"), scr); err != nil {
							t.Fatal(err)
						}
					}
				}
				title := fmt.Sprintf("%s · %s · %dx%d · theme %s", style, age, w, h, th)
				bg := capture.Hex(theme.Color(theme.RoleBackground))
				fg := capture.Hex(theme.Color(theme.RoleText))
				_ = os.WriteFile(filepath.Join(out, name+".html"), []byte(capture.HTML(title, bg, fg, capture.Fonts, frames, 6)), 0o644)
				_ = os.WriteFile(filepath.Join(out, name+".txt"), []byte(title+"\n"+txt.String()), 0o644)
				scr.Fini()
			}
		}
	}
	if before := os.Getenv("SKY_BEFORE_DIR"); before != "" {
		contactSheet(t, before, out)
	}
}

// contactSheet writes contact.html and contact.png beside the two capture
// folders: one row per age, before and after for each style.
func contactSheet(t *testing.T, before, after string) {
	t.Helper()
	dir := filepath.Dir(after)
	rel := func(p string) string {
		r, err := filepath.Rel(dir, p)
		if err != nil {
			return p
		}
		return filepath.ToSlash(r)
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>The sky arc: before and after</title>
<style>body{background:#111;color:#ccc;font-family:ui-monospace,Menlo,monospace;margin:16px}
table{border-collapse:collapse}td,th{padding:6px;vertical-align:top;text-align:center}
img{width:480px;border:1px solid #333}a{color:#8cf}h1{font-size:16px}</style></head><body>
<h1>The sky arc: before (master) and after, 160x48, forge theme. Click a picture for the animated capture.</h1><table>
<tr><th>age</th><th>roguelike before</th><th>roguelike after</th><th>skyline before</th><th>skyline after</th></tr>
`)
	type cell struct{ img, html string }
	var grid [][]cell
	for _, age := range skyShotAges {
		row := []cell{}
		fmt.Fprintf(&b, "<tr><th>%s</th>", short(age))
		for _, style := range Registry().Names() {
			for _, d := range []string{before, after} {
				base := filepath.Join(d, style+"_"+short(age))
				c := cell{img: base + ".png", html: base + ".html"}
				row = append(row, c)
				fmt.Fprintf(&b, `<td><a href="%s"><img src="%s"></a><br><a href="%s">mini</a></td>`, rel(c.html), rel(c.img),
					rel(base+"_mini.html"))
			}
		}
		b.WriteString("</tr>\n")
		grid = append(grid, row)
	}
	b.WriteString("</table></body></html>\n")
	if err := os.WriteFile(filepath.Join(dir, "contact.html"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	// the same grid as one picture, half size, for a quick look
	const cw, ch, gap = 480, 288, 8
	sheet := image.NewRGBA(image.Rect(0, 0, 4*(cw+gap)+gap, len(grid)*(ch+gap)+gap))
	draw.Draw(sheet, sheet.Bounds(), image.Black, image.Point{}, draw.Src)
	for r, row := range grid {
		for c, cl := range row {
			f, err := os.Open(cl.img)
			if err != nil {
				continue
			}
			img, err := png.Decode(f)
			f.Close()
			if err != nil {
				continue
			}
			x0, y0 := gap+c*(cw+gap), gap+r*(ch+gap)
			sb := img.Bounds()
			for y := 0; y < ch; y++ {
				for x := 0; x < cw; x++ {
					sheet.Set(x0+x, y0+y, img.At(sb.Min.X+x*sb.Dx()/cw, sb.Min.Y+y*sb.Dy()/ch))
				}
			}
		}
	}
	f, err := os.Create(filepath.Join(dir, "contact.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, sheet); err != nil {
		t.Fatal(err)
	}
}
