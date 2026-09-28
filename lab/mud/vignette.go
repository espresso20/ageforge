package main

// vignette.go draws a room's scene: a strip of sky, a row of building stamps,
// and the ground. One grammar for every age. The dials are the stamp set (by
// the era a building comes from, so an old hut keeps its thatch beside a new
// terrace), the ground texture, and the sky (time of day, weather, and the
// stars once the city leaves the planet). Colour comes from a handful of glyph
// classes mapped to theme roles, so the shapes carry the meaning on their own.

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/espresso20/ageforge/theme"
)

// band is the stamp era: 0 ancient (stone, iron), 1 industrial (steel,
// electric), 2 modern (digital, neon), 3 orbital (cosmic).
func band(epoch int) int {
	switch {
	case epoch <= 1:
		return 0
	case epoch <= 3:
		return 1
	case epoch <= 5:
		return 2
	}
	return 3
}

// bandForAge maps an age index to its stamp band.
func bandForAge(ageIdx int) int {
	switch {
	case ageIdx <= 5:
		return 0
	case ageIdx <= 11:
		return 1
	case ageIdx <= 17:
		return 2
	}
	return 3
}

// stamps[place][band] is one building drawn in that era; rows align at the
// bottom. A missing band falls back to the one before it.
var stamps = map[PlaceKey][4][]string{
	Square: { // culture: theatres, halls, studios
		{"▄▀▀▀▄", "▌≡≡≡▐"},
		{" ╭─╮ ", "╭┴─┴╮", "▐▫▫▫▌"},
		{"▐▀▀▌", "▐◘ ▌"},
		{"╭◇╮", "╰─╯"},
	},
	Homes: {
		{" /\\ ", "/▫ \\", "|__|"},
		{"┌┴┐", "│▫│", "│▫│"},
		{"▐▀▀▌", "▐▫▫▌", "▐▫▫▌"},
		{" ╭─╮ ", "─┤▫├─", " ╰─╯ "},
	},
	Fields: {
		{" ψ ψ ", "\"ψ\"ψ\""},
		{"ψψψψψ", "ψψψψψ"},
		{" ╭──╮ ", "╭╯ψψ╰╮"},
		{"╔ψψψ╗", "╚═══╝"},
	},
	Woods: {
		{" ♣ ", "♣♣♣", " | "},
		{" ╔╗ ", " ║╠╗", "▄╨╨▄"},
		{" ║ ║ ", "╔╩═╩╗", "║▫ ▫║"},
		{" ◇ ◇ ", "╔═╦═╗"},
	},
	Quarry: {
		{" ▄▀▄ ", "▓▒░▒▓"},
		{" ╔╗  ", " ║╚╗ ", "▄╨─╨▄"},
		{" ┬┬ ", " ││ ", "▓▓▓▓"},
		{" ▄▄ ", "▐▓▒▌", " ▀│ "},
	},
	Forge: {
		{" ° ", "▄*▄", "███"},
		{" °  ° ", " ▌  ▌ ", "▐█▄▄█▌"},
		{"▐█▌▐█▌", "▐▫▙▟▫▌"},
		{" \\│/ ", "─ ☼ ─", " /│\\ "},
	},
	Works: {
		{"  °  ", " /¯\\ ", "|▫ ▫|"},
		{"╱│╱│╱│", "│▫▫▫▫│"},
		{" ╲╱ ", " ╱╲ ", "╱  ╲"},
		{"  ▲  ", " ▐█▌ ", " ╱║╲ "},
	},
	Temple: {
		{"  o  ", " ▄█▄ ", "▀███▀"},
		{"  †  ", " ╱▲╲ ", "│▫▫▫│"},
		{" ─┬─ ", "╱▔▔▔╲", " │▫│ "},
		{" ( ) ", "╱ † ╲"},
	},
	Academy: {
		{"  ∩  ", " │§│ ", " └─┘ "},
		{" ╭─╮ ", "╭┴─┴╮", "│▫▫▫│"},
		{"┌──┐", "│▪▪│", "│▪▪│"},
		{"  ╱ ", " ▄█▄", "▐███▌"},
	},
	Market: {
		{" ¤ o ¤", "▀▀ ▀▀ "},
		{"╱▔▔▔╲", "║ ║ ║", "▀▀▀▀▀"},
		{" ▐▌  ", "▐▫▌▐▌", "▐▫▌▐▫"},
		{" ╭◊╮ ", "◊┤$├◊"},
	},
	Harbour: {
		{"  _/ ", "\\___/"},
		{" │╲ │╲ ", " │_╲│_╲", "╲_____╱"},
		{"═╦══ ", " ║ ▄ ", "▀▀▀▀▀"},
		{" ═╪═ ", "  ║  ", " ═╪═ "},
	},
	Stores: {
		{" ___ ", "(▒▒▒)"},
		{"╱▔▔▔╲", "│▫ ▫│"},
		{"╔═══╗", "║ ◘ ║"},
		{"[▒][▒]"},
	},
	Barracks: {
		{" /\\ ║", "/▫ \\║"},
		{"▙▟▙▟▙", "█▫ ▫█"},
		{" ╲▁ ", " ╱ │", "▐██▌"},
		{" ▄▄▄> ", "▐████>"},
	},
	Wonders: {
		{"▐▌  ▐▌", "▐▌▀▀▐▌", "▐▌  ▐▌"},
		{"  ║  ", " ╱║╲ ", "╱ ╨ ╲"},
		{"  ▲  ", " ▐▫▌ ", "▐▫▫▫▌"},
		{"  ┼  ", "  ║  ", " ╱║╲ "},
	},
	Gate: {
		{"║   ║", "╨   ╨"},
		{"╱▔▔▔▔╲", "│ o  │"},
		{"═╗ ╔═", "═╝ ╚═"},
		{"╭───╮", "│   │", "╰───╯"},
	},
}

// squareStamps are the square's landmark per epoch: the one place whose look
// changes with every epoch, since it is where every walk starts.
var squareStamps = [7][]string{
	{"  )  ", " (*) ", "o─^─o"},
	{"▁▄▄▄▄▄▁", " ║ ║ ║ ", "▀▀▀▀▀▀▀"},
	{"  ╷  ", " ╲│╱ ", "▐▄▄▄▌"},
	{" ▲ ", "▐o▌", "▐▫▌", "▐▫▌"},
	{" ╱──╲ ", "╱ ▫▫ ╲", "│ ▫▫ │"},
	{" ╱╲╱╲ ", "▐▫▫▫▫▌", "▐▫▫▫▫▌", "▐▫  ▫▌"},
	{"╭───╮", "│ ◦ │", "╰───╯"},
}

// wonderStamps give every wonder its own silhouette, so the wonder walk reads
// as your history rather than a row of the same monument.
var wonderStamps = map[string][]string{
	"sacred_grove":             {" ♣♣♣ ", "♣♣†♣♣"},
	"great_monolith":           {" ▄ ", " █ ", "▐█▌"},
	"stonehenge":               {"▄▄▄ ▄▄▄", "█ █ █ █"},
	"colosseum":                {"▄▄▄▄▄▄▄", "▌∩∩∩∩∩▐", "▌∩∩∩∩∩▐"},
	"parthenon":                {"▁▄▄▄▄▄▁", " ║║║║║ ", "▀▀▀▀▀▀▀"},
	"great_library":            {" ▄▄▄▄▄ ", "▐≡≡≡≡≡▌", "▐≡≡≡≡≡▌"},
	"sistine_chapel":           {"  †  ", " ╱▔╲ ", "▐▫▫▫▌"},
	"grand_lighthouse":         {" * ", "▐▀▌", "▐ ▌", "▟█▙"},
	"crystal_palace":           {" ╭─────╮ ", "╭┴▫▫▫▫▫┴╮"},
	"eiffel_tower":             {"  ╷  ", "  ║  ", " ╱╫╲ ", "╱ ╨ ╲"},
	"hoover_dam":               {"▀▄▄▄▄▄▀", " ▀███▀ ", "≈≈▐█▌≈≈"},
	"particle_accelerator":     {"╭──────╮", "│ ◦  ◦ │", "╰──────╯"},
	"space_program":            {"  ▲  ", " ▐█▌ ", " ▐█▌ ", "╱▀▀▀╲"},
	"global_network":           {"╭─┬─╮", "├─┼─┤", "╰─┴─╯"},
	"world_simulation":         {"▐▪▪▪▌", "▐▪▪▪▌", "▐▪▪▪▌"},
	"neon_citadel":             {" ▲ ▲ ", "▐█▀█▌", "▐▫▫▫▌"},
	"stellar_cradle":           {"\\ | /", " (*) ", "/ | \\"},
	"dyson_scaffold":           {"╭┄┄┄╮", "┆ ☼ ┆", "╰┄┄┄╯"},
	"warp_nexus":               {"╲ ╱", "─◊─", "╱ ╲"},
	"cosmic_beacon":            {"  ┼  ", "  ║  ", " ╱║╲ "},
	"reality_anchor":           {" ╥ ", "─╫─", "╰╨╯"},
	"singularity_core":         {"▗▄▖", "▐●▌", "▝▀▘"},
	"cultural_obelisk":         {" ▲ ", " █ ", "▐█▌"},
	"eternal_library_monument": {" ▄▄▄ ", "▐≡≡≡▌"},
}

var (
	ruinStamp      = []string{"  ▗  ", " ▟x▖ ", "▟▘ ▀▙"}
	harbingerStamp = []string{" o ", "/█\\", "/ \\"}
	caravanStamp   = []string{" ▄▄o", "(o)(o)"}
	barricadeStamp = []string{"┼─┼─┼", "│x│x│"}
	pathStamp      = []string{" ┬ ", " │ "}
)

// ground textures per band (by place where it matters).
func groundFor(p PlaceKey, b int) string {
	switch p {
	case Harbour:
		if b == 3 {
			return "═╪══╪══"
		}
		return "≈~≈≈~≈~"
	case Fields:
		if b <= 1 {
			return "\",',\"'."
		}
	case Quarry:
		if b <= 2 {
			return "░▒░░▒░▓"
		}
	case Woods:
		if b == 0 {
			return ",.♣,.,'"
		}
	}
	switch b {
	case 0:
		return "._,.'_."
	case 1:
		return "▁▁▁▁▁▁▁"
	case 2:
		return "▔▔▔▔▔▔▔"
	}
	return "═╪═══╪═"
}

// glyph classes → theme roles. Shape carries meaning; colour only helps.
func roleFor(r rune, night bool) theme.Role {
	switch {
	case strings.ContainsRune(`/\^▲◢◣╱╲∧∩†`, r):
		return theme.RoleAccent
	case strings.ContainsRune(`*o°☼✦¤$◊◘◦>`, r):
		return theme.RoleHighlight
	case r == '▫' || r == '▪':
		if night {
			return theme.RoleHighlight
		}
		return theme.RoleDim
	case strings.ContainsRune(`~≈`, r):
		return theme.RoleLabel
	case strings.ContainsRune(`♣♠"'ψ`, r):
		return theme.RolePositive
	case strings.ContainsRune(`.,:░▒▓_▁▔`+"`", r):
		return theme.RoleDim
	case strings.ContainsRune(`x▗▖▟▙▘▝`, r):
		return theme.RoleNegative
	}
	return theme.RoleText
}

// Cell is one character of a drawn scene.
type Cell struct {
	R    rune
	Role theme.Role
	Bold bool
}

// Grid is a w×h block of cells; zero cells are blank.
type Grid struct {
	W, H  int
	Cells []Cell
	Solid []bool // inside a drawn stamp: weather does not fall here
}

func NewGrid(w, h int) *Grid {
	w, h = max(w, 0), max(h, 0)
	return &Grid{W: w, H: h, Cells: make([]Cell, w*h), Solid: make([]bool, w*h)}
}

// solid marks a w×h box as occupied.
func (g *Grid) solid(x, y, w, h int) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if i >= 0 && j >= 0 && i < g.W && j < g.H {
				g.Solid[j*g.W+i] = true
			}
		}
	}
}

func (g *Grid) Set(x, y int, r rune, role theme.Role) {
	if x < 0 || y < 0 || x >= g.W || y >= g.H {
		return
	}
	g.Cells[y*g.W+x] = Cell{R: r, Role: role}
}

func (g *Grid) At(x, y int) Cell {
	if x < 0 || y < 0 || x >= g.W || y >= g.H {
		return Cell{}
	}
	return g.Cells[y*g.W+x]
}

func stampWidth(s []string) int {
	w := 0
	for _, l := range s {
		if n := utf8.RuneCountInString(l); n > w {
			w = n
		}
	}
	return w
}

func stampFor(p PlaceKey, b int) []string {
	set := stamps[p]
	for ; b >= 0; b-- {
		if len(set[b]) > 0 {
			return set[b]
		}
	}
	return pathStamp
}

// hash is a small deterministic mixer (FNV-1a over the parts).
func hash(parts ...string) uint32 {
	h := uint32(2166136261)
	for _, p := range parts {
		for i := 0; i < len(p); i++ {
			h ^= uint32(p[i])
			h *= 16777619
		}
		h ^= 0xff
		h *= 16777619
	}
	return h
}

// Scene is everything a vignette needs to know beyond the place itself.
type Scene struct {
	Band      int
	Hour      float64 // 0..24
	Weather   string
	Harbinger bool
	Frame     int
	Orbital   bool
}

func (s Scene) night() bool { return s.Hour < 5.5 || s.Hour >= 20 }

// drawVignette renders place p into a w×h grid.
func drawVignette(c *City, p *Place, sc Scene, w, h int) *Grid {
	g := NewGrid(w, h)
	if w < 8 || h < 3 {
		return g
	}
	night := sc.night() || sc.Orbital
	groundY := h - 1

	// Sky.
	skyRows := h - 4
	if skyRows < 1 {
		skyRows = 1
	}
	seed := string(p.Key) + c.St.Age
	if night {
		for y := 0; y < skyRows; y++ {
			for x := 0; x < w; x++ {
				v := hash(seed, "star", itoa(x), itoa(y)) % 23
				switch {
				case v == 0:
					twinkle := (int(hash(itoa(x), itoa(y))) + sc.Frame) % 5
					if twinkle == 0 {
						g.Set(x, y, '+', theme.RoleHighlight)
					} else {
						g.Set(x, y, '·', theme.RoleDim)
					}
				case v == 1 && sc.Orbital:
					g.Set(x, y, '.', theme.RoleDim)
				}
			}
		}
		if sc.Orbital {
			// the planet's limb along the top-right corner.
			for x := w - 12; x < w; x++ {
				if x >= 0 {
					g.Set(x, 0, '▄', theme.RoleLabel)
				}
			}
			if w-15 >= 0 {
				g.Set(w-14, 0, '▗', theme.RoleLabel)
				g.Set(w-13, 0, '▄', theme.RoleLabel)
			}
		} else {
			mx := int(float64(w-1) * math.Mod(sc.Hour+4, 24) / 10)
			if mx >= w {
				mx = w - 1
			}
			g.Set(clamp(mx, 0, w-1), 0, '☾', theme.RoleBright)
		}
	} else {
		// Sun: east (left) at dawn to west (right) at dusk.
		f := (sc.Hour - 6) / 14
		sx := clamp(int(f*float64(w-1)), 0, w-1)
		g.Set(sx, 0, '☼', theme.RoleWarning)
		if sc.Weather == "cloud" || sc.Weather == "rain" {
			// Clouds: a few banks of uneven length that drift a cell per frame.
			n := 2 + int(hash(seed, "banks")%3)
			for i := 0; i < n; i++ {
				start := int(hash(seed, "bank", itoa(i))%uint32(w)) + sc.Frame
				length := 5 + int(hash(seed, "len", itoa(i))%9)
				row := int(hash(seed, "row", itoa(i)) % 2)
				for j := 0; j < length; j++ {
					ch := '▀'
					if j == 0 || j == length-1 {
						ch = '▔'
					}
					g.Set((start+j)%w, row, ch, theme.RoleDim)
				}
			}
		}
	}
	// Ground.
	gpat := []rune(groundFor(p.Key, sc.Band))
	for x := 0; x < w; x++ {
		off := 0
		if p.Key == Harbour {
			off = sc.Frame
		}
		r := gpat[(x+off)%len(gpat)]
		g.Set(x, groundY, r, roleFor(r, night))
	}

	// Stamps: newest holdings first, sub-linear in count, old eras keep their look.
	type item struct {
		art  []string
		ruin bool
	}
	var items []item
	if sc.Harbinger && p.Key == Square {
		items = append(items, item{art: harbingerStamp})
	}
	if p.Key == Square {
		items = append(items, item{art: squareStamps[c.Epoch]})
	}
	if p.Path || len(p.Holdings) == 0 {
		if p.Key == Gate {
			items = append(items, item{art: stampFor(Gate, sc.Band)})
		} else if p.Key != Square {
			items = append(items, item{art: pathStamp})
		}
	} else if p.Key == Wonders {
		for _, hd := range p.Holdings {
			if hd.Count > 0 {
				art, ok := wonderStamps[hd.Key]
				if !ok {
					art = stampFor(Wonders, bandForAge(hd.Tier))
				}
				items = append(items, item{art: art})
			}
		}
	} else {
		total := p.Total()
		k := 1 + int(math.Log2(float64(max(total, 1))))
		left := k
		for i, hd := range p.Holdings {
			if hd.Count == 0 || left == 0 {
				continue
			}
			n := int(math.Round(float64(k) * float64(hd.Count) / float64(total)))
			if n < 1 {
				n = 1
			}
			if i == 0 && n < 2 && k >= 3 {
				n = 2
			}
			if n > left {
				n = left
			}
			left -= n
			art := stampFor(p.Key, bandForAge(hd.Tier))
			for j := 0; j < n; j++ {
				items = append(items, item{art: art})
			}
		}
	}
	if p.Key == Gate && len(c.atWar()) > 0 {
		items = append(items, item{art: barricadeStamp, ruin: true})
	}
	if p.Key == Gate && len(c.St.Trade.ActiveRoutes) > 0 {
		items = append(items, item{art: caravanStamp})
	}
	if p.Ruins() > 0 {
		items = append(items, item{art: ruinStamp, ruin: true})
	}
	// Fit: drop from the end until the row fits.
	gap := 2
	fits := func(n int) int {
		t := 0
		for i := 0; i < n; i++ {
			t += stampWidth(items[i].art)
		}
		return t + gap*(n-1)
	}
	n := len(items)
	for n > 1 && fits(n) > w {
		n--
	}
	if n > 0 && n < len(items) && items[len(items)-1].ruin {
		items[n-1] = items[len(items)-1] // the ruin always shows
	}
	total := fits(n)
	x := (w - total) / 2
	if x < 0 {
		x = 0
	}
	// Scenery: the neighbours you could walk to, faint at the edges of the
	// scene on the side their road leaves from (west left, east right). It is
	// what makes the first clearing show its huts, and it tells you, without
	// the map, what lies which way.
	drawSide := func(k PlaceKey, left bool, limit int) {
		q := c.Places[k]
		if q == nil || q.Path || len(q.Holdings) == 0 || limit < 4 {
			return
		}
		var art []string
		if k == Wonders {
			art = wonderStamps[q.Holdings[0].Key]
		}
		if art == nil {
			art = stampFor(k, bandForAge(q.Holdings[0].Tier))
		}
		sw := stampWidth(art)
		reps := min(2, max(1, q.Total()/4))
		for r := 0; r < reps; r++ {
			if (r+1)*(sw+1) > limit {
				break
			}
			sx := 1 + r*(sw+1)
			if !left {
				sx = w - 1 - (r+1)*(sw+1) + 1
			}
			top := groundY - len(art)
			g.solid(sx, top, sw, len(art))
			for row, line := range art {
				col := 0
				for _, ch := range line {
					if ch != ' ' && g.At(sx+col, top+row).R == 0 {
						g.Set(sx+col, top+row, ch, theme.RoleDim)
					}
					col++
				}
			}
		}
	}
	if w >= 40 {
		if k, ok := p.Exits[West]; ok {
			drawSide(k, true, x-2)
		}
		if k, ok := p.Exits[East]; ok {
			drawSide(k, false, w-(x+total)-2)
		}
	}
	for i := 0; i < n; i++ {
		art := items[i].art
		sw := stampWidth(art)
		top := groundY - len(art)
		g.solid(x, top, sw, len(art))
		for row, line := range art {
			col := 0
			for _, r := range line {
				if r != ' ' {
					role := roleFor(r, night)
					if items[i].ruin {
						role = theme.RoleNegative
					}
					if sc.Harbinger && p.Key == Square && i == 0 {
						role = theme.RoleWarning
					}
					g.Set(x+col, top+row, r, role)
				}
				col++
			}
		}
		// Smoke rises from anything that burns: one puff per chimney, climbing
		// a row per frame and drifting with the wind.
		for col, r := range []rune(art[0]) {
			if r == '°' {
				k := 1 + (sc.Frame+col+i)%3
				yy := top - k
				dx := k / 2
				if yy >= 0 && g.At(x+col+dx, yy).R == 0 {
					ch := '°'
					if k == 3 {
						ch = '·'
					}
					g.Set(x+col+dx, yy, ch, theme.RoleDim)
				}
			}
		}
		x += sw + gap
	}

	// Weather last, only into empty air.
	switch sc.Weather {
	case "rain":
		// Scattered drops that fall a row per frame and slant with the wind.
		for y := 1; y < groundY; y++ {
			for x := 0; x < w; x++ {
				if hash(seed, "rain", itoa(x+(y-sc.Frame)), itoa(y-sc.Frame))%23 == 0 && g.At(x, y).R == 0 && !g.Solid[y*g.W+x] {
					g.Set(x, y, '╱', theme.RoleDim)
				}
			}
		}
	case "fog":
		for x := 0; x < w; x++ {
			if hash(itoa((x+sc.Frame)/3), "fog")%3 != 0 && g.At(x, groundY-1).R == 0 {
				g.Set(x, groundY-1, '░', theme.RoleDim)
			}
		}
	case "solar", "meteor":
		x := (sc.Frame*5 + int(hash(seed)%uint32(w))) % w
		if g.At(x, 1).R == 0 {
			g.Set(x, 1, '─', theme.RoleHighlight)
		}
		g.Set(clamp(x+1, 0, w-1), 1, '•', theme.RoleHighlight)
	}
	return g
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		b[n] = '-'
	}
	return string(b[n:])
}
