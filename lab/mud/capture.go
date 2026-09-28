package main

// capture.go renders the walk into a tcell SimulationScreen and writes the
// cells out as plain text and as HTML with inline colours: real terminal
// renders, one cell per character, nothing rasterised.

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// Frame is one captured screen plus what the player typed to get there.
type Frame struct {
	Caption string
	W, H    int
	Cells   []tcell.SimCell
}

func snap(s tcell.SimulationScreen, caption string) Frame {
	s.Show()
	cells, w, h := s.GetContents()
	cp := make([]tcell.SimCell, len(cells))
	copy(cp, cells)
	return Frame{Caption: caption, W: w, H: h, Cells: cp}
}

func (f Frame) Text() string {
	var b strings.Builder
	for y := 0; y < f.H; y++ {
		var line strings.Builder
		for x := 0; x < f.W; x++ {
			r := f.Cells[y*f.W+x].Runes
			if len(r) == 0 {
				line.WriteByte(' ')
			} else {
				line.WriteRune(r[0])
			}
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

func hex(c tcell.Color) string {
	if c == tcell.ColorDefault || c.Hex() < 0 {
		return ""
	}
	return fmt.Sprintf("#%06x", c.Hex())
}

// HTML renders the frame's cells as a <pre> of styled runs.
func (f Frame) HTML() string {
	var b strings.Builder
	bgDefault := hex(theme.Color(theme.RoleBackground))
	b.WriteString(`<pre class="scr">`)
	for y := 0; y < f.H; y++ {
		type run struct {
			fg, bg string
			bold   bool
			text   strings.Builder
		}
		var cur *run
		flush := func() {
			if cur == nil {
				return
			}
			st := "color:" + cur.fg
			if cur.bg != "" && cur.bg != bgDefault {
				st += ";background:" + cur.bg
			}
			if cur.bold {
				st += ";font-weight:bold"
			}
			fmt.Fprintf(&b, `<span style="%s">%s</span>`, st, html.EscapeString(cur.text.String()))
			cur = nil
		}
		for x := 0; x < f.W; x++ {
			cell := f.Cells[y*f.W+x]
			fg, bg, attr := cell.Style.Decompose()
			r := ' '
			if len(cell.Runes) > 0 {
				r = cell.Runes[0]
			}
			fs, bs, bold := hex(fg), hex(bg), attr&tcell.AttrBold != 0
			if cur == nil || cur.fg != fs || cur.bg != bs || cur.bold != bold {
				flush()
				cur = &run{fg: fs, bg: bs, bold: bold}
			}
			cur.text.WriteRune(r)
		}
		flush()
		b.WriteByte('\n')
	}
	b.WriteString(`</pre>`)
	return b.String()
}

func pageHead(title string) string {
	bg := hex(theme.Color(theme.RoleBackground))
	fg := hex(theme.Color(theme.RoleText))
	dim := hex(theme.Color(theme.RoleDim))
	return `<!doctype html><html><head><meta charset="utf-8"><title>` + html.EscapeString(title) + `</title><style>
body{background:` + bg + `;color:` + fg + `;margin:16px;font-family:ui-monospace,Menlo,Consolas,"DejaVu Sans Mono",monospace}
pre.scr{font-family:inherit;font-size:14px;line-height:1.2;margin:0;display:inline-block;background:` + bg + `}
.cap{color:` + dim + `;font-size:13px;margin:8px 0}
button{font-family:inherit;background:transparent;color:` + fg + `;border:1px solid ` + dim + `;padding:2px 10px;cursor:pointer}
</style></head><body>`
}

func writeFrames(dir, name, title string, frames []Frame, animate bool) error {
	var txt strings.Builder
	for i, f := range frames {
		if len(frames) > 1 {
			fmt.Fprintf(&txt, "=== %d/%d  %s ===\n", i+1, len(frames), f.Caption)
		}
		txt.WriteString(f.Text())
	}
	if err := os.WriteFile(filepath.Join(dir, name+".txt"), []byte(txt.String()), 0o644); err != nil {
		return err
	}
	var h strings.Builder
	h.WriteString(pageHead(title))
	switch {
	case len(frames) == 1:
		h.WriteString(frames[0].HTML())
	case animate:
		h.WriteString(`<div id="f"></div><div class="cap" id="c"></div><script>const F=[`)
		for i, f := range frames {
			if i > 0 {
				h.WriteByte(',')
			}
			fmt.Fprintf(&h, "%q", f.HTML())
		}
		h.WriteString(`];let i=0;function s(){document.getElementById('f').innerHTML=F[i];document.getElementById('c').textContent='frame '+(i+1)+'/'+F.length+' (700 ms per frame, as the panel ticks)';i=(i+1)%F.length}s();setInterval(s,700);</script>`)
	default:
		h.WriteString(`<div class="cap"><button onclick="go(-1)">◀ prev</button> <button onclick="go(1)">next ▶</button> <label><input type="checkbox" id="auto" checked> auto</label> <span id="c"></span></div><div id="f"></div><script>const F=[`)
		for i, f := range frames {
			if i > 0 {
				h.WriteByte(',')
			}
			fmt.Fprintf(&h, "[%q,%q]", f.Caption, f.HTML())
		}
		h.WriteString(`];let i=0;function show(){document.getElementById('f').innerHTML=F[i][1];document.getElementById('c').textContent='step '+(i+1)+'/'+F.length+'   > '+F[i][0]}function go(d){i=(i+d+F.length)%F.length;show()}show();setInterval(()=>{if(document.getElementById('auto').checked)go(1)},3200);</script>`)
	}
	h.WriteString(`</body></html>`)
	return os.WriteFile(filepath.Join(dir, name+".html"), []byte(h.String()), 0o644)
}

// Shot is one capture in the set.
type Shot struct {
	Name, State, Since, Theme string
	W, H                      int
	Steps                     []string // typed commands; "" is the opening look
	Mini                      bool
	Animate                   int // frames of animation instead of steps
	Title                     string
	// Mutate, if set, edits the real state before drawing: a labelled what-if
	// for situations the seeded run never reached (war, ruins).
	Mutate func(*game.GameState)
}

func (sh Shot) run(states string) ([]Frame, error) {
	if err := theme.SetActive(sh.Theme); err != nil {
		return nil, err
	}
	st, err := loadState(states, sh.State)
	if err != nil {
		return nil, err
	}
	if sh.Mutate != nil {
		sh.Mutate(&st)
	}
	var prev *City
	if sh.Since != "" {
		ps, err := loadState(states, sh.Since)
		if err != nil {
			return nil, err
		}
		prev = BuildCity(ps)
	}
	v := NewView(BuildCity(st), prev)
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		return nil, err
	}
	defer s.Fini()
	s.SetSize(sh.W, sh.H)
	draw := func() {
		if sh.Mini {
			v.DrawMini(s)
		} else {
			v.Draw(s)
		}
	}
	var frames []Frame
	steps := sh.Steps
	if len(steps) == 0 {
		steps = []string{""}
	}
	for _, cmd := range steps {
		for _, c := range strings.Split(cmd, ";") {
			if c != "" {
				v.Exec(c)
			}
		}
		if sh.Animate > 0 {
			for i := 0; i < sh.Animate; i++ {
				v.Frame = i
				draw()
				frames = append(frames, snap(s, fmt.Sprintf("frame %d", i+1)))
			}
			continue
		}
		draw()
		cap := cmd
		if cap == "" {
			cap = "(opening look)"
		}
		frames = append(frames, snap(s, cap))
	}
	return frames, nil
}

func captureAll(states, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	shots := []Shot{
		// One per age: the opening look from the square, 120×36, Forge (dark).
		{Name: "primitive_age_first", State: "primitive_age_first", Steps: []string{""}},
		{Name: "primitive_age", State: "primitive_age", Steps: []string{"go huts"}},
		{Name: "bronze_age", State: "bronze_age", Steps: []string{""}},
		{Name: "medieval_age", State: "medieval_age", Steps: []string{""}},
		{Name: "industrial_age", State: "industrial_age", Steps: []string{"go forge"}},
		{Name: "digital_age", State: "digital_age", Steps: []string{"go market"}},
		{Name: "cyberpunk_age", State: "cyberpunk_age", Steps: []string{"go homes"}},
		{Name: "interstellar_age", State: "interstellar_age", Steps: []string{"go academy"}},
		{Name: "galactic_age", State: "galactic_age", Steps: []string{""}},
		// Walks: a sequence of typed commands.
		{Name: "walk_medieval", State: "medieval_age", Title: "Walk: Medieval Age", Steps: []string{"", "north", "east", "visit great library", "south", "go gate", "out", "visit ironworks", "survey"}},
		{Name: "walk_galactic", State: "galactic_age", Title: "Walk: Galactic Age", Steps: []string{"", "hubward", "spinward", "go launch", "rimward", "out"}},
		{Name: "walk_primitive", State: "primitive_age_first", Title: "Walk: the first minutes", Steps: []string{"", "west", "south", "go track", "survey"}},
		// Idle return: the medieval city an age-length after the early snapshot.
		{Name: "away_medieval", State: "medieval_age", Since: "medieval_age_early", Title: "Idle return", Steps: []string{"", "look", "go homes", "go forge"}},
		// Light theme.
		{Name: "medieval_age_light", State: "medieval_age", Theme: "daylight", Steps: []string{""}},
		{Name: "industrial_age_light", State: "industrial_age", Theme: "parchment", Steps: []string{"go fields"}},
		// Small terminals.
		{Name: "industrial_age_100x30", State: "industrial_age", W: 100, H: 30, Steps: []string{""}},
		{Name: "medieval_age_80x24", State: "medieval_age", W: 80, H: 24, Steps: []string{""}},
		// The compact, glanceable form.
		{Name: "mini_medieval_40x15", State: "medieval_age", Since: "medieval_age_early", W: 40, H: 15, Mini: true},
		{Name: "mini_galactic_40x15", State: "galactic_age", W: 40, H: 15, Mini: true},
		{Name: "mini_primitive_40x15", State: "primitive_age_first", W: 40, H: 15, Mini: true},
		// What-if: the real Industrial state with a war declared and a
		// catastrophe's ruins added, since the seeded run met neither.
		{Name: "whatif_war_ruins_industrial", State: "industrial_age", Title: "What-if: war and ruins (real state, edited)",
			Steps: []string{"go gate", "go forge", "go terraces", "survey"}, Mutate: func(st *game.GameState) {
				for k, f := range st.Diplomacy.Factions {
					if f.Discovered && f.Personality == "aggressive" || k == "ironhold_clans" {
						f.AtWar, f.Status = true, "war"
						st.Diplomacy.Factions[k] = f
						break
					}
				}
				for _, k := range []string{"iron_works", "ironmonger", "row_house", "townhouse", "villa"} {
					if b, ok := st.Buildings[k]; ok && b.Count > 2 {
						b.RuinCount = b.Count / 3
						b.Count -= b.RuinCount
						st.Buildings[k] = b
					}
				}
			}},
		// Animation: frames of the panel ticking.
		{Name: "anim_industrial", State: "industrial_age", Title: "Industrial, ticking", Steps: []string{"go forge"}, Animate: 4},
		{Name: "anim_medieval_forge", State: "medieval_age", Title: "Medieval forge, ticking", Steps: []string{"go forge"}, Animate: 4},
	}
	var index []string
	for _, sh := range shots {
		if sh.Theme == "" {
			sh.Theme = "forge"
		}
		if sh.W == 0 {
			sh.W, sh.H = 120, 36
		}
		if sh.Title == "" {
			sh.Title = sh.Name
		}
		if _, err := os.Stat(filepath.Join(states, sh.State+".json.gz")); err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: no state %s\n", sh.Name, sh.State)
			continue
		}
		frames, err := sh.run(states)
		if err != nil {
			return fmt.Errorf("%s: %w", sh.Name, err)
		}
		if err := writeFrames(out, sh.Name, sh.Title, frames, sh.Animate > 0); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%d frame(s), %dx%d, %s)\n", sh.Name, len(frames), sh.W, sh.H, sh.Theme)
		index = append(index, fmt.Sprintf(`<li><a href="%s.html">%s</a> <span>%d frame(s) · %d×%d · %s</span> <a href="%s.txt">txt</a></li>`,
			sh.Name, html.EscapeString(sh.Title), len(frames), sh.W, sh.H, sh.Theme, sh.Name))
	}
	_ = theme.SetActive("forge")
	page := pageHead("Map Lab: mud captures") + `<h3>Map Lab · mud: the city you walk through</h3><ul>` + strings.Join(index, "\n") +
		`</ul><p class="cap">Real terminal renders: tcell SimulationScreen cells written out with inline colours. States come from the smoke bot (seed 7).</p></body></html>`
	return os.WriteFile(filepath.Join(out, "index.html"), []byte(page), 0o644)
}
