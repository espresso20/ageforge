package kit

import (
	"image"
	"image/color"
	"sync"
	"time"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/theme"
)

const (
	// TipDelay is how long a target is rested on before its tip shows.
	TipDelay = 400 * time.Millisecond
	// TipMaxCols and TipMaxRows bound the plate, border included.
	TipMaxCols, TipMaxRows = 40, 6
)

// TipText is what a tip says: plain lines, then an optional reason that is
// drawn in the warning colour behind a sign.
type TipText struct {
	Lines  []string
	Reason string
}

// A Tip is the one tooltip of a screen: a small plate that shows after the
// pointer, or the pick, has rested on a target, and goes on the next move off
// the target, any key, a click or a change of screen. Aim is called by the
// widget under the pointer; Move and Clear by the loop that owns the tip.
// It is safe to ask Waiting from another goroutine.
type Tip struct {
	// Now is the clock; nil means time.Now.
	Now func() time.Time
	// Cell reports the pixel size of one cell. Nil, or a zero size, draws
	// the plate in cells with an ASCII border.
	Cell func() (w, h int)

	mu     sync.Mutex
	id     any
	target image.Rectangle
	text   TipText
	due    time.Time
	img    *image.RGBA
	imgKey [3]int
}

func (t *Tip) now() time.Time {
	if t.Now != nil {
		return t.Now()
	}
	return time.Now()
}

// Aim says the pointer (or, with byKey, the pick) is on the target id, which
// covers target in cells. A new id starts the delay; the same id keeps it
// when it came by key or the tip already shows, and restarts it when the
// pointer moved within the target.
func (t *Tip) Aim(id any, target image.Rectangle, text TipText, byKey bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	same := t.id != nil && t.id == id
	shown := same && !t.now().Before(t.due)
	if !same || (!byKey && !shown) {
		t.due = t.now().Add(TipDelay)
	}
	t.id, t.target, t.text = id, target, text
}

// Move clears the tip unless the pointer only moved within the target, and
// reports whether it took a shown tip away (the screen is drawn again).
func (t *Tip) Move(m Mouse) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.id == nil || (m.Kind == MouseMove && m.Pt().In(t.target)) {
		return false
	}
	shown := !t.now().Before(t.due)
	t.id = nil
	return shown
}

// Clear takes the tip away.
func (t *Tip) Clear() {
	t.mu.Lock()
	t.id = nil
	t.mu.Unlock()
}

// Visible reports whether the tip is showing.
func (t *Tip) Visible() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.id != nil && !t.now().Before(t.due)
}

// Waiting reports whether a tip is due but not yet showing, so that the
// screen is drawn again once it is.
func (t *Tip) Waiting() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.id != nil && t.now().Before(t.due)
}

// cut shortens s to w columns, ending in ell when it was cut.
func cut(s string, w int, ell string) string {
	if uniseg.StringWidth(s) <= w {
		return s
	}
	w -= uniseg.StringWidth(ell)
	out, n := "", 0
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		if n += g.Width(); n > w {
			break
		}
		out += g.Str()
	}
	return out + ell
}

// body fits the text to the plate: the lines (and the reason, behind its
// sign) cut to cols columns and to rows rows, the last kept row ending in an
// ellipsis when lines were dropped.
func (t *Tip) body(cols, rows int, ell string) (lines []string, reason string) {
	lines = append(lines, t.text.Lines...)
	keep := rows
	if t.text.Reason != "" {
		keep--
		reason = cut("! "+t.text.Reason, cols, ell)
	}
	if keep < 0 {
		keep = 0
	}
	if len(lines) > keep {
		lines = lines[:keep]
		if keep > 0 {
			lines[keep-1] = cut(lines[keep-1], cols-uniseg.StringWidth(ell), "") + ell
		}
	}
	for i := range lines {
		lines[i] = cut(lines[i], cols, ell)
	}
	return lines, reason
}

// Rect is where the plate sits, in cells, within area (the screen without
// the command bar): below the target with left edges aligned, above it when
// there is no room below, beside it when there is room in neither, and
// moved left to stay inside. It is empty when no tip shows.
func (t *Tip) Rect(area image.Rectangle) image.Rectangle {
	if !t.Visible() {
		return image.Rectangle{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rect(area)
}

func (t *Tip) rect(area image.Rectangle) image.Rectangle {
	pic := t.pictured()
	ell, ink, padRows := "…", 2, 0
	if !pic {
		ell, ink, padRows = "...", 4, 2
	}
	lines, reason := t.body(TipMaxCols-ink, TipMaxRows-padRows, ell)
	w, h := 0, len(lines)
	for _, l := range lines {
		w = max(w, uniseg.StringWidth(l))
	}
	if reason != "" {
		w = max(w, uniseg.StringWidth(reason))
		h++
	}
	w, h = min(w+ink, TipMaxCols), min(h+padRows, TipMaxRows)
	x, y := t.target.Min.X, t.target.Max.Y
	if y+h > area.Max.Y {
		y = t.target.Min.Y - h
		if y < area.Min.Y {
			x, y = t.target.Max.X, t.target.Min.Y
		}
	}
	x = max(min(x, area.Max.X-w), area.Min.X)
	y = max(min(y, area.Max.Y-h), area.Min.Y)
	return image.Rect(x, y, x+w, y+h)
}

func (t *Tip) pictured() bool {
	if t.Cell == nil {
		return false
	}
	w, h := t.Cell()
	return w > 0 && h > 0
}

// Paint draws the plate over whatever is painted, within area.
func (t *Tip) Paint(p *Painter, area image.Rectangle) {
	if !t.Visible() {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	r := t.rect(area)
	if r.Empty() {
		return
	}
	text := Style{}.Foreground(Role(theme.RoleText))
	warn := Style{}.Foreground(Role(theme.RoleWarning))
	if !t.pictured() {
		p.Fill(r, Role(theme.RoleSurface))
		p.Box(r, BorderASCII, text)
		lines, reason := t.body(r.Dx()-4, r.Dy()-2, "...")
		for i, l := range lines {
			p.Print(r.Min.X+2, r.Min.Y+1+i, r.Dx()-4, Escape(l), text)
		}
		if reason != "" {
			p.Print(r.Min.X+2, r.Min.Y+1+len(lines), r.Dx()-4, Escape(reason), warn)
		}
		return
	}
	cw, ch := t.Cell()
	fill := tipFill()
	p.Fill(r, RGB(fill.R, fill.G, fill.B))
	p.Picture(r, t.plate(r.Dx()*cw, r.Dy()*ch, fill))
	lines, reason := t.body(r.Dx()-2, r.Dy(), "…")
	for i, l := range lines {
		p.Label(r.Min.X+1, r.Min.Y+i, Escape(l), text)
	}
	if reason != "" {
		p.Label(r.Min.X+1, r.Min.Y+len(lines), Escape(reason), warn)
	}
}

// tipFill is the plate's fill: the surface one step toward the text.
func tipFill() color.RGBA {
	return mixRGBA(theme.RGBA(theme.RoleSurface), theme.RGBA(theme.RoleText), 0.12)
}

// plate draws the pixels: a one-pixel dark outline, the flat fill, a lighter
// line under the top edge, and corners cut in steps.
func (t *Tip) plate(w, h int, fill color.RGBA) *image.RGBA {
	key := [3]int{w, h, int(fill.R)<<16 | int(fill.G)<<8 | int(fill.B)}
	if t.img != nil && t.imgKey == key {
		return t.img
	}
	edge := mixRGBA(fill, color.RGBA{0, 0, 0, 255}, 0.6)
	top := mixRGBA(fill, color.RGBA{255, 255, 255, 255}, 0.22)
	gap := theme.RGBA(theme.RoleBackground)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := fill
			switch {
			case x == 0 || y == 0 || x == w-1 || y == h-1:
				c = edge
			case y == 1 && x > 1 && x < w-2:
				c = top
			}
			if (x == 0 || x == w-1) && (y == 0 || y == h-1) {
				c = gap
			}
			img.SetRGBA(x, y, c)
		}
	}
	t.img, t.imgKey = img, key
	return img
}
