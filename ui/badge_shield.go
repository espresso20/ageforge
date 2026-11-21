package ui

import (
	"image"
	"image/color"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/wirebadge"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// badgeShield draws a badge's wire shield for a block of cols by rows cells,
// pixels square at a 7 by 15 cell (cols a multiple of 3). Earned shows full
// colour with a halo; the rest a dim outline. An integrity badge has no tier
// and gets nothing. It writes nothing outside the picture it returns.
func badgeShield(v game.BadgeView, cols, rows int) *image.RGBA {
	w, h := cols*7/3, rows*5
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if v.Level < config.BadgeBronze {
		return img
	}
	t, bg := wirebadge.Tier(v.Level-config.BadgeBronze), pixelInk(theme.Of(theme.RoleBackground))
	col := wirebadge.Color(t, int(time.Now().UnixMilli()/150))
	halo := wirebadge.Mix(col, bg, .72)
	if !v.Earned {
		col, halo = wirebadge.Mix(col, bg, .55), color.RGBA{}
	}
	s := min(1, float64(w)/104, float64(h)/92)
	wirebadge.Draw(func(x, y int, c color.RGBA) {
		if image.Pt(x, y).In(img.Rect) {
			img.SetRGBA(x, y, c)
		}
	}, w/2, int((float64(h)-88*s)/2+34*s), t, v.Domain, col, halo, s)
	return img
}

const (
	badgeCardW, badgeCardRows = 42, 20
	badgeNotePage             = "badge_note"
	badgeNoteFor              = 3 * time.Second
)

var txtBadgeEarned = rowid.E("ui.badge_note.earned", "BADGE EARNED")

// paintCard is the picked badge on the right of a wide list: its shield, name,
// tier and what it asks for.
func (p *ranksPanel) paintCard(s rankPaint, x, h int) {
	i := p.rowIndex(p.view.sel)
	if i < 0 || h < 36 {
		return
	}
	v := p.rows[i]
	s.p.Picture(image.Rect(s.o.X+x, s.o.Y+6, s.o.X+x+badgeCardW, s.o.Y+6+badgeCardRows), badgeShield(v, badgeCardW, badgeCardRows))
	y := 6 + badgeCardRows + 1
	if v.Hidden {
		return
	}
	s.text(x, y, v.Name, kit.Role(theme.RoleBright), true)
	s.text(x, y+1, rowid.Sprintf("ui.badge_note.tier", "%s  +%d", v.Tier, v.Points), tierCol(v.Level), false)
	for j, ln := range wrapCells(v.Desc, badgeCardW) {
		if j < 4 {
			s.text(x, y+3+j, ln, kit.Role(theme.RoleText), false)
		}
	}
}

// badgeNote is the earn notice: the shield, the name and the tier.
type badgeNote struct {
	*nativeBox
	v     game.BadgeView
	close func()
}

func (n *badgeNote) Paint(p *kit.Painter) {
	r := n.Bounds()
	p.Fill(r, theme.Of(theme.RoleSurface))
	p.Picture(image.Rect(r.Min.X+3, r.Min.Y+1, r.Min.X+45, r.Min.Y+16), badgeShield(n.v, 42, 15))
	st := kit.Style{}.Foreground(kit.Role(theme.RoleBright)).Bold(true)
	p.Print(r.Min.X+1, r.Min.Y+17, r.Dx()-2, txtBadgeEarned.String()+": "+n.v.Name, st)
	p.Print(r.Min.X+1, r.Min.Y+18, r.Dx()-2, rowid.Sprintf("ui.badge_note.tier", "%s  +%d", n.v.Tier, n.v.Points), kit.Style{}.Foreground(tierCol(n.v.Level)))
}

// Mouse: a click closes the notice.
func (n *badgeNote) Mouse(m kit.Mouse) bool {
	if m.Kind != kit.MousePress {
		return false
	}
	n.close()
	return true
}

// pumpBadgeNote shows the next earned badge about 3 s each, one after another.
func (d *Dashboard) pumpBadgeNote() {
	if d.badgeNote && time.Now().After(d.badgeUntil) {
		d.closeBadgeNote()
	}
	if d.badgeNote || len(d.badgeQ) == 0 {
		return
	}
	n := &badgeNote{nativeBox: newNativeBox(), v: d.badgeQ[0], close: d.closeBadgeNote}
	d.badgeQ, d.badgeNote, d.badgeUntil = d.badgeQ[1:], true, time.Now().Add(badgeNoteFor)
	page := floatingModal(n, 48, 20)
	page.hold = false
	d.pages.AddPage(badgeNotePage, page, true, true)
}

func (d *Dashboard) closeBadgeNote() {
	if d.badgeNote {
		d.pages.RemovePage(badgeNotePage)
		d.badgeNote = false
	}
}
