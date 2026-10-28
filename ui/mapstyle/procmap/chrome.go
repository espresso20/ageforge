package procmap

import (
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// chrome.go is everything round the picture: the header, the news line, the
// inspector and status line, the flows panel and the compact view's two bars.

// sideWidth is the width of the sidebar.
const sideWidth = 26

// The first epoch whose map is drawn at night.
const firstNightEpoch = 5

// inks are the colours the chrome is drawn in for one theme and epoch.
type inks struct {
	*mapstyle.Palette
	theme  string
	border tcell.Style
}

// inksFor returns in when it still fits the theme and the epoch, else a new
// set. In the night epochs the land recedes so lit structures glow, which the
// cursor's fill is chosen against.
func inksFor(in *inks, epoch int) *inks {
	key := theme.Active().Key
	if in != nil && in.theme == key && in.Epoch == epoch {
		return in
	}
	pal := mapstyle.NewPalette(epoch)
	if epoch >= firstNightEpoch {
		for _, c := range []mapmodel.Class{mapmodel.CGround, mapmodel.CFlora, mapmodel.CRock, mapmodel.CHill, mapmodel.CWater} {
			pal.Fg[c] = theme.Legible(theme.Mix(pal.Fg[c], pal.Bg, 0.35), pal.Bg, mapstyle.TerrainFloor)
		}
	}
	pal.FitCursor()
	border := tcell.StyleDefault.Foreground(theme.Legible(theme.Color(theme.RoleBorder), pal.Bg, 1.8)).Background(pal.Bg)
	return &inks{Palette: pal, theme: key, border: border}
}

// begin readies the chrome for a frame and returns its inks, or nil when there
// is no model.
func (v *view) begin(f mapstyle.Frame) *inks {
	if f.Model == nil {
		return nil
	}
	v.tier = f.Tier
	v.ink = inksFor(v.ink, f.Model.Epoch)
	return v.ink
}

// drawNoMap says there is no map yet, for a frame with no model.
func drawNoMap(cv *mapstyle.Canvas) {
	dim := mapstyle.NewPalette(0).Style(mapmodel.CDim)
	cv.Text(1, 0, 0, vibe.Line("ui.mapstyle.roguelike.chrome.Draw.01", "No map yet."), dim)
}

// drawHeader is the top row: settlement, age, epoch, weather, people, idle,
// civs met and the zoom's name.
func (v *view) drawHeader(cv *mapstyle.Canvas, width int, m *mapmodel.Model, zoom string) {
	ink := v.ink.Chrome
	_, background, _ := ink.Decompose()
	cv.Fill(0, 0, width, 1, ' ', ink)
	x := cv.Text(1, 0, width-2, "◆ "+m.TownName, ink.Foreground(theme.Legible(theme.Color(theme.RoleAccent), background, 3)).Bold(true))
	x = cv.Text(x, 0, width-x-1, " · "+m.AgeName+" · "+m.EpochName, ink)
	idle := ink
	if m.Workers.Idle > 0 {
		idle = ink.Foreground(theme.Legible(v.ink.Fg[mapmodel.CIdle], background, 3)).Bold(true)
	}
	segments := []struct {
		text  string
		style tcell.Style
	}{
		{string(weatherRune(m, v.tier)) + " " + m.Clock.String() + "  ", ink},
		{vibe.Line("ui.mapstyle.roguelike.chrome.header.01", "pop ") + strconv.Itoa(m.Workers.Pop) + "  ", ink},
		{vibe.Line("ui.mapstyle.roguelike.chrome.header.02", "idle ") + strconv.Itoa(m.Workers.Idle), idle},
		{vibe.Line("ui.mapstyle.roguelike.chrome.header.03", "  civs ") + strconv.Itoa(m.DiscoveredCount()) + "/" + strconv.Itoa(len(m.Factions)), ink},
		{"  " + zoom + " ", ink.Bold(true)},
	}
	for ; len(segments) > 0; segments = segments[1:] { // drop from the left until it fits
		total := 0
		for _, segment := range segments {
			total += mapstyle.TextLen(segment.text)
		}
		if x+total+2 > width {
			continue
		}
		right := width - total
		for _, segment := range segments {
			right = cv.Text(right, 0, 0, segment.text, segment.style)
		}
		return
	}
}

var weatherSyms = map[mapmodel.WeatherKind]mapmodel.Sym{
	mapmodel.Cloudy: mapmodel.SymCloud,
	mapmodel.Smog:   mapmodel.SymCloud,
	mapmodel.Rain:   mapmodel.SymRain,
	mapmodel.Storm:  mapmodel.SymBolt,
	mapmodel.Snow:   mapmodel.SymSnow,
}

func weatherRune(m *mapmodel.Model, tier mapmodel.GlyphTier) rune {
	if sym, ok := weatherSyms[m.Weather.Kind]; ok {
		return mapmodel.R(sym, tier)
	}
	if m.Clock.Night > 0.5 {
		return mapmodel.R(mapmodel.SymMoon, tier)
	}
	return mapmodel.R(mapmodel.SymSun, tier)
}

// news is the one line worth knowing: the recap headline, the pending
// catastrophe, or what is stuck.
func (v *view) news(m *mapmodel.Model, width int) (string, tcell.Style) {
	danger := v.ink.Role(theme.RoleNegative).Bold(true)
	headline := ""
	if v.changes {
		headline = m.Recap.Headline(width - 2)
	}
	if width < 60 && headline != "" { // narrow: the items without the preamble
		var parts []string
		for _, item := range m.Recap.Items {
			parts = append(parts, item.Text)
		}
		headline = mapmodel.Clip(strings.Join(parts, " · "), width)
	}
	switch {
	case headline != "" && m.Recap.Items[0].Bad:
		return headline, danger
	case headline != "":
		return headline, v.ink.Role(theme.RolePositive)
	case m.Catastrophe.Pending != "":
		return mapmodel.Clip(vibe.Line("ui.mapstyle.roguelike.chrome.news.01", "DANGER: ")+m.Catastrophe.PendingName+vibe.Line("ui.mapstyle.roguelike.chrome.news.02", " is upon you"), width), danger
	case m.Flows.Worst != "" && v.changes:
		return mapmodel.Clip(vibe.Line("ui.mapstyle.roguelike.chrome.news.03", "stuck: ")+m.Flows.Worst, width), v.ink.Role(theme.RoleWarning)
	}
	return vibe.Line("ui.mapstyle.roguelike.chrome.news.04", "all quiet"), v.ink.Style(mapmodel.CDim)
}

// drawBottom is the news, the two inspector lines and the status line, whose
// left end names the zoom.
func (v *view) drawBottom(cv *mapstyle.Canvas, width, height int, m *mapmodel.Model, in mapstyle.Inspection, ok bool, zoom string) {
	text, style := v.news(m, width-2)
	cv.Text(1, height-4, width-2, text, style)
	dim := v.ink.Style(mapmodel.CDim)
	short := vibe.Line("ui.mapstyle.roguelike.chrome.bottom.02", "arrows move  -/= or PgUp/PgDn zoom  Tab next")
	hints := short
	if width >= 100 {
		hints = vibe.Line("ui.mapstyle.roguelike.chrome.bottom.01", "arrows move  Shift fast  -/= or PgUp/PgDn zoom  Tab next  Home center  map flows")
	}
	switch {
	case !ok:
		cv.Text(1, height-3, width-2, vibe.Line("ui.mapstyle.roguelike.chrome.bottom.03", "cursor hidden; move it with the arrow keys"), v.ink.Style(mapmodel.CText))
		cv.Text(3, height-2, width-4, hints, dim)
	default:
		cv.Text(1, height-3, width-2, "▸ "+strings.Join(append([]string{in.Title}, in.Lines...), " · "), v.ink.Role(theme.RoleBright).Bold(true))
		if in.Command == "" {
			cv.Text(3, height-2, width-4, vibe.Line("ui.mapstyle.roguelike.chrome.bottom.04", "nothing to type here; Tab jumps to the next building"), dim)
		} else {
			x := cv.Text(3, height-2, width-4, vibe.Line("ui.mapstyle.roguelike.chrome.bottom.05", "type: "), dim)
			cv.Text(x, height-2, width-x-1, in.Command, v.ink.Style(mapmodel.CAccent).Bold(true))
		}
	}
	if v.flows {
		zoom += vibe.Line("ui.mapstyle.roguelike.chrome.bottom.06", " · flows")
	}
	x := cv.Text(1, height-1, width-2, zoom, v.ink.Style(mapmodel.CLabel))
	if x+2+mapstyle.TextLen(hints) > width-1 {
		hints = short // the long one does not fit beside the zoom name
	}
	if hintWidth := mapstyle.TextLen(hints); x+2+hintWidth <= width-1 {
		cv.Text(width-1-hintWidth, height-1, hintWidth, hints, dim)
	}
}

// drawSidebar draws the sidebar's rule and, when the flows are on, their
// panel; it returns the first row below, where the legend goes.
func (v *view) drawSidebar(cv *mapstyle.Canvas, m *mapmodel.Model, x0, y0, width, height int) int {
	for y := y0; y < y0+height; y++ {
		cv.Put(x0, y, '│', v.ink.border)
	}
	y := y0
	if !v.flows {
		return y
	}
	cv.Text(x0+2, y, width-3, vibe.Line("ui.mapstyle.roguelike.chrome.sidebar.01", "FLOWS"), v.ink.Style(mapmodel.CAccent).Bold(true))
	for _, line := range v.flowLines(m) {
		if y++; y < y0+height {
			v.drawFlowLine(cv, x0+2, y, width-3, line)
		}
	}
	return y + 2
}

// flowLine is one flows row: a tag and text, or a share bar.
type flowLine struct {
	tag, text string
	class     mapmodel.Class
	bar       float64 // > 0: a share bar
	glyph     rune
}

func (v *view) flowLines(m *mapmodel.Model) []flowLine {
	flows := m.Flows
	var lines []flowLine
	add := func(ok bool, tag, text string, class mapmodel.Class) {
		if ok {
			lines = append(lines, flowLine{tag: tag, text: text, class: class})
		}
	}
	var draining []string
	for _, store := range flows.Draining {
		draining = append(draining, strings.ToLower(store.Name))
	}
	add(len(flows.Draining) > 0, vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.01", "DRAIN"), strings.Join(draining, ", "), mapmodel.CDanger)
	add(flows.IdleWorkers > 0, vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.02", "IDLE "), plural(flows.IdleWorkers, vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.03", "worker"), vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.04", "workers")), mapmodel.CIdle)
	short := len(flows.Understaffed) + len(flows.Idle)
	add(short > 0, vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.05", "SHORT"), plural(short, vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.06", "building"), vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.07", "buildings")), mapmodel.CIdle)
	add(len(lines) == 0, vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.08", "OK   "), vibe.Line("ui.mapstyle.roguelike.chrome.flowLines.09", "nothing stuck"), mapmodel.CFresh)
	for i, share := range flows.Shares {
		if i == 5 || share.Share <= 0 {
			break
		}
		lines = append(lines, flowLine{
			text:  lineageLabel(share.Lineage),
			bar:   share.Share,
			class: mapmodel.LineageClass(share.Lineage),
			glyph: mapmodel.R(mapmodel.LineageSym(share.Lineage, m.Epoch), v.tier),
		})
	}
	return lines
}

func (v *view) drawFlowLine(cv *mapstyle.Canvas, x, y, width int, line flowLine) {
	if line.bar <= 0 {
		after := cv.Text(x, y, width, line.tag, v.ink.Style(line.class).Bold(true))
		cv.Text(after+1, y, x+width-after-1, line.text, v.ink.Style(mapmodel.CText))
		return
	}
	cv.Put(x, y, line.glyph, v.ink.Style(line.class))
	filled := max(1, int(math.Round(float64(line.bar*6))))
	for i := 0; i < 6 && i+2 < width; i++ {
		if i < filled {
			cv.Put(x+2+i, y, '█', v.ink.Style(line.class))
		} else {
			cv.Put(x+2+i, y, '░', v.ink.Style(mapmodel.CMemory))
		}
	}
	after := cv.Text(x+9, y, width-9, strconv.Itoa(int(float64(line.bar*100)+0.5))+"% ", v.ink.Style(mapmodel.CText))
	cv.Text(after, y, x+width-after, line.text, v.ink.Style(mapmodel.CDim))
}

// drawFlowsBox draws the flows panel in a frame over the picture, for a screen
// with no room for a sidebar.
func (v *view) drawFlowsBox(cv *mapstyle.Canvas, m *mapmodel.Model, x, y, width, maxHeight int) {
	lines := v.flowLines(m)
	height := min(len(lines)+2, maxHeight)
	if width < 8 || height < 3 {
		return
	}
	border := v.ink.border
	cv.Fill(x, y, width, height, ' ', v.ink.Style(mapmodel.CText))
	for i := 0; i < width; i++ {
		cv.Put(x+i, y, '─', border)
		cv.Put(x+i, y+height-1, '─', border)
	}
	for j := 0; j < height; j++ {
		cv.Put(x, y+j, '│', border)
		cv.Put(x+width-1, y+j, '│', border)
	}
	corners := [4][2]int{{x, y}, {x + width - 1, y}, {x, y + height - 1}, {x + width - 1, y + height - 1}}
	for i, corner := range corners {
		cv.Put(corner[0], corner[1], []rune("┌┐└┘")[i], border)
	}
	cv.Text(x+2, y, width-4, vibe.Line("ui.mapstyle.roguelike.chrome.box.01", "FLOWS"), v.ink.Style(mapmodel.CAccent).Bold(true))
	for i := 0; i < len(lines) && i < height-2; i++ {
		v.drawFlowLine(cv, x+1, y+1+i, width-2, lines[i])
	}
}

// drawCompactBars draws the first and last lines of the compact view: the age
// and the people, and the news.
func (v *view) drawCompactBars(cv *mapstyle.Canvas, r mapstyle.Rect, m *mapmodel.Model) {
	title := strings.TrimSuffix(m.AgeName, " Age") + vibe.Line("ui.mapstyle.roguelike.chrome.DrawCompact.02", " · pop ") + strconv.Itoa(m.Workers.Pop)
	x := cv.Text(0, 0, r.W, title, v.ink.Style(mapmodel.CAccent).Bold(true))
	if m.Workers.Idle > 0 {
		cv.Text(x, 0, r.W-x, vibe.Line("ui.mapstyle.roguelike.chrome.DrawCompact.03", " · idle ")+strconv.Itoa(m.Workers.Idle), v.ink.Style(mapmodel.CIdle).Bold(true))
	}
	if r.H >= 2 {
		text, style := v.news(m, r.W)
		cv.Text(0, r.H-1, r.W, text, style)
	}
}

// drawLegend lists the signs of the building lines the town has.
func (v *view) drawLegend(cv *mapstyle.Canvas, mp *mapstyle.Palette, x, y, w, h int, m *mapmodel.Model) {
	cv.Text(x, y, w, vibe.Line("ui.mapstyle.roguelike.chrome.drawLegend.01", "LEGEND"), mp.Style(mapmodel.CAccent).Bold(true))
	for i, l := range m.Lineages {
		if i+1 >= h {
			return
		}
		st := mp.Style(mapmodel.CText)
		_, bg, _ := st.Decompose()
		cv.Put(x, y+1+i, mapgen.Sign(l.Key), tcell.StyleDefault.Foreground(v.pal.colour(mapgen.LineHue(l.Key))).Background(bg))
		cv.Text(x+2, y+1+i, w-2, l.Name, st)
	}
}

// lineageLabel names a building line in the inspector and the flows.
func lineageLabel(lineage string) string {
	if lineage == mapmodel.LinHarbor {
		return "harbor"
	}
	return mapmodel.LineageName(lineage)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}
