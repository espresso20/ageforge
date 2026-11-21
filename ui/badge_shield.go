package ui

import (
	"image"
	"image/color"
	"time"
	"unicode/utf8"

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

// paintBadge draws a badge's shield in a block of cols by rows cells at (x,y)
// and under it the name, tier and what it asks for, centred when mid is set.
func paintBadge(s rankPaint, x, y, cols, rows int, v game.BadgeView, mid bool) {
	s.p.Picture(image.Rect(s.o.X+x, s.o.Y+y, s.o.X+x+cols, s.o.Y+y+rows), badgeShield(v, cols, rows))
	if v.Hidden {
		return
	}
	put := func(dy int, str string, col kit.Color, bold bool) {
		dx := 0
		if mid {
			dx = max(0, (cols-utf8.RuneCountInString(str))/2)
		}
		s.text(x+dx, y+rows+1+dy, str, col, bold)
	}
	put(0, v.Name, kit.Role(theme.RoleBright), true)
	put(1, rowid.Sprintf("ui.badge_note.tier", "%s  +%d", v.Tier, v.Points), tierCol(v.Level), false)
	for j, ln := range wrapCells(v.Desc, cols) {
		if j < 3 {
			put(3+j, ln, kit.Role(theme.RoleText), false)
		}
	}
}

// paintCard is the picked badge on the right of a wide list.
func (p *ranksPanel) paintCard(s rankPaint, x, h int) {
	if i := p.rowIndex(p.view.sel); i >= 0 && h >= 36 {
		paintBadge(s, x, 6, badgeCardW, badgeCardRows, p.rows[i], false)
	}
}

// badgeNote is the earn notice: a framed box with the shield, name and tier.
type badgeNote struct {
	*nativeBox
	v     game.BadgeView
	close func()
}

func (n *badgeNote) Paint(p *kit.Painter) {
	r := n.Bounds()
	p.Fill(r, kit.Role(theme.RoleSelection))
	p.Fill(r.Inset(1), theme.Of(theme.RoleSurface))
	p.Box(r, kit.BorderLine, kit.Style{Fg: kit.Role(theme.RoleBorder)})
	p.Print(r.Min.X+2, r.Min.Y, r.Dx()-4, " "+txtBadgeEarned.String()+" ", kit.Style{}.Foreground(kit.Role(theme.RoleAccent)).Bold(true))
	paintBadge(rankPaint{p, r.Min, r.Dx()}, 3, 1, 42, 15, n.v, true)
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
	// Wait while any dialog, splash or card is up: the notice never covers one.
	if d.badgeNote || len(d.badgeQ) == 0 || len(d.pages.GetPageNames(false)) > 1 {
		return
	}
	n := &badgeNote{nativeBox: newNativeBox(), v: d.badgeQ[0], close: d.closeBadgeNote}
	d.badgeQ, d.badgeNote, d.badgeUntil = d.badgeQ[1:], true, time.Now().Add(badgeNoteFor)
	page := floatingModal(n, 48, 25)
	page.hold = false
	d.pages.AddPage(badgeNotePage, page, true, true)
}

func (d *Dashboard) closeBadgeNote() {
	if d.badgeNote {
		d.pages.RemovePage(badgeNotePage)
		d.badgeNote = false
	}
}
