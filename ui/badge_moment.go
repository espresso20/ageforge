package ui

import (
	"github.com/espresso20/ageforge/rowid"
	"image"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/coins"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

const (
	momentW     = 48 // the card's columns
	momentH     = 8  // and rows
	momentLife  = 80 // coinStep steps a card stays
	momentOpen  = 4  // steps it takes to open, and to close
	momentCoinW = 12
	momentCoinH = 6
	momentTextX = 16 // the text column's offset in the card
	momentTextW = momentW - momentTextX - 2
)

// badgeMoment is the card for the moment a badge is earned: a small card in
// the top right with the badge's coin, held still, and the badge's name, tier
// and description. Badges earned in a row wait in
// a queue and take the card one after another.
type badgeMoment struct {
	queue []game.BadgeView
	start time.Time // when the card on show began; zero when none is
	mint  *coins.Mint
}

// add queues a badge for the card.
func (m *badgeMoment) add(v game.BadgeView) { m.queue = append(m.queue, v) }

// paint draws the card on show into the top right of area, drops it after its
// life and starts the next one at the following paint. It reports whether a
// card is on show and moving, so the caller keeps redrawing.
func (m *badgeMoment) paint(p *kit.Painter, area image.Rectangle, now time.Time, motion bool, tier mapmodel.GlyphTier) bool {
	if len(m.queue) == 0 {
		return false
	}
	if m.start.IsZero() {
		m.start = now
	}
	steps := int(now.Sub(m.start) / coinStep)
	if steps >= momentLife {
		m.queue, m.start = m.queue[1:], time.Time{}
		return false
	}
	if !motion {
		steps = momentOpen // whole and still for the whole life
	}
	rows := momentH
	switch {
	case steps < momentOpen && motion:
		rows = 2 * (steps + 1)
	case steps >= momentLife-momentOpen && motion:
		rows = 2 * (momentLife - steps)
	}
	v := m.queue[0]
	if m.mint == nil {
		m.mint = coins.NewMint(16)
	}
	card := image.Rect(area.Max.X-momentW-1, area.Min.Y, area.Max.X-1, area.Min.Y+rows)
	m.draw(p, card, v, tier, steps >= momentOpen, steps >= momentOpen && steps < momentLife-momentOpen)
	return motion
}

// draw paints the card's top rows (the rectangle card), with its text when
// withText is set and the coin when withCoin is.
func (m *badgeMoment) draw(p *kit.Painter, card image.Rectangle, v game.BadgeView, tier mapmodel.GlyphTier, withText, withCoin bool) {
	plain := tier == mapmodel.TierASCII
	paleC := kit.Role(theme.RoleBright)
	ground := kit.Role(theme.RoleSurface)

	p.Fill(card, ground)
	border := kit.BorderRound
	if plain {
		border = kit.BorderASCII
	}
	p.Box(card, border, kit.Style{Fg: paleC, Bg: ground})

	org := image.Pt(card.Min.X, card.Min.Y)
	p = p.Clip(card.Inset(1))
	if withCoin {
		coin := image.Rect(org.X+2, org.Y+1, org.X+2+momentCoinW, org.Y+1+momentCoinH)
		cw, ch := cellPixels()
		paintCoin(p, m.mint, v.Coin, coin, cw, ch)
	}
	if !withText {
		return
	}
	x := org.X + momentTextX
	head, tierText := badgeToastWords(v)
	text := kit.Style{Bg: ground}
	p.Print(x, org.Y+2, momentTextW, lit(strings.TrimSpace(head)), text.Foreground(kit.Role(theme.RoleDim)))
	name := v.Name
	if runeLen(name) > momentTextW {
		name = clipRunes(name, momentTextW-1) + "…"
	}
	if plain {
		name = foldRunes(name)
	}
	p.Print(x, org.Y+3, momentTextW, lit(name), text.Foreground(paleC).Bold(true))
	if tierText != "" {
		p.Print(x, org.Y+4, momentTextW, lit(strings.TrimSpace(tierText)+"."), text.Foreground(kit.Role(theme.RoleDim)))
	}
	lines := wrapText(v.Desc, momentTextW)
	if len(lines) > 2 {
		lines = lines[:2]
		if runeLen(lines[1]) >= momentTextW {
			lines[1] = clipRunes(lines[1], momentTextW-1)
		}
		lines[1] += "…"
	}
	for i, l := range lines {
		if plain {
			l = foldRunes(l)
		}
		p.Print(x, org.Y+5+i, momentTextW, lit(l), text.Foreground(kit.Role(theme.RoleText)))
	}
}

// badgeToastWords are the card's head and the badge's tier text (" (gold)",
// or nothing).
func badgeToastWords(v game.BadgeView) (head, tierText string) {
	head = rowid.Line("ui.badge_list.badgeToast.01", "Badge earned:")
	if v.Tier != "" {
		tierText = " (" + v.Tier + ")"
	}
	return head, tierText
}

// foldRunes is s in ASCII, as the badges' plain tier folds it.
func foldRunes(s string) string {
	return strings.Map(func(r rune) rune { return foldRune(r, badgeFold) }, s)
}
