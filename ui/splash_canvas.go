package ui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ── Star layers ─────────────────────────────────────────────────────────────

type splashStar struct {
	nx, ny float64 // normalised position [0,1]
	phase  float64
	speed  float64
	tier   int // 0 = dim background  1 = mid  2 = bright foreground
}

var (
	scTierChar  = []rune{'.', '·', '✦'}
	scTierBase  = []float64{0.06, 0.25, 0.50}
	scTierRange = []float64{0.20, 0.46, 0.50}
)

// ── Canvas primitive ─────────────────────────────────────────────────────────

// splashCanvas is an animated starfield that also renders the AGEFORGE title
// and tagline.  It implements tview.Primitive by embedding *tview.Box.
type splashCanvas struct {
	*tview.Box
	stars         []splashStar
	tick          float64
	prestigeLevel int
	mu            sync.RWMutex
	stop          chan struct{}
}

func newSplashCanvas(prestigeLevel int) *splashCanvas {
	sc := &splashCanvas{
		Box:           tview.NewBox(),
		prestigeLevel: prestigeLevel,
		stop:          make(chan struct{}),
		stars:         make([]splashStar, 280),
	}
	for i := range sc.stars {
		tier := 0
		r := rand.Float64()
		switch {
		case r > 0.91:
			tier = 2
		case r > 0.54:
			tier = 1
		}
		sc.stars[i] = splashStar{
			nx:    rand.Float64(),
			ny:    rand.Float64(),
			phase: rand.Float64() * math.Pi * 2,
			speed: 0.22 + rand.Float64()*1.6,
			tier:  tier,
		}
	}
	return sc
}

// animate starts the background goroutine that ticks the animation.
func (sc *splashCanvas) animate(app *tview.Application) {
	go func() {
		ticker := time.NewTicker(75 * time.Millisecond) // ~13 fps
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sc.mu.Lock()
				sc.tick += 0.075
				sc.mu.Unlock()
				app.QueueUpdateDraw(func() {})
			case <-sc.stop:
				return
			}
		}
	}()
}

// halt stops the animation goroutine; safe to call multiple times.
func (sc *splashCanvas) halt() {
	select {
	case <-sc.stop:
	default:
		close(sc.stop)
	}
}

// Draw implements tview.Primitive.
func (sc *splashCanvas) Draw(screen tcell.Screen) {
	sc.DrawForSubclass(screen, sc) // draws border/title if set
	bx, by, bw, bh := sc.GetInnerRect()

	sc.mu.RLock()
	tick := sc.tick
	sc.mu.RUnlock()

	// ── Stars ────────────────────────────────────────────────────────────────
	for _, s := range sc.stars {
		px := bx + int(s.nx*float64(bw))
		py := by + int(s.ny*float64(bh))
		if px < bx || px >= bx+bw || py < by || py >= by+bh {
			continue
		}
		raw := (math.Sin(tick*s.speed+s.phase) + 1.0) / 2.0
		alpha := scTierBase[s.tier] + raw*scTierRange[s.tier]
		if alpha < 0.04 {
			continue
		}
		v := int32(alpha * 220)
		if v > 220 {
			v = 220
		}
		fg := tcell.NewRGBColor(v+35, v+25, v+15)
		style := tcell.StyleDefault.Foreground(fg)
		if s.tier == 2 && alpha > 0.82 {
			style = style.Bold(true)
		}
		screen.SetContent(px, py, scTierChar[s.tier], nil, style)
	}

	// ── AGEFORGE title ───────────────────────────────────────────────────────
	lines := strings.Split(strings.TrimSpace(SplashArt), "\n")

	titleH := len(lines)
	titleW := 0
	for _, l := range lines {
		cols := 0
		for range l {
			cols++
		}
		if cols > titleW {
			titleW = cols
		}
	}

	// Place title in the upper-centre of the canvas (45% down)
	originY := by + bh*9/20 - titleH/2
	originX := bx + (bw-titleW)/2

	// Slow pulse: gold hue breathes between dim and bright
	pulse := (math.Sin(tick*0.55) + 1.0) / 2.0 // 0–1
	goldR := int32(182 + int(pulse*73))
	goldG := int32(112 + int(pulse*68))
	titleFG := tcell.NewRGBColor(goldR, goldG, 0)
	titleStyle := tcell.StyleDefault.Foreground(titleFG).Bold(true)

	for i, line := range lines {
		ty := originY + i
		if ty < by || ty >= by+bh {
			continue
		}
		col := 0
		for _, ch := range line {
			tx := originX + col
			col++
			if ch == ' ' || tx < bx || tx >= bx+bw {
				continue
			}
			screen.SetContent(tx, ty, ch, nil, titleStyle)
		}
	}

	// ── Tagline ──────────────────────────────────────────────────────────────
	tagline := SplashTagline
	tly := originY + titleH + 2
	tlx := bx + (bw-len(tagline))/2
	if tly >= by && tly < by+bh {
		mutedStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(139, 148, 158))
		col := 0
		for _, ch := range tagline {
			tx := tlx + col
			col++
			if tx < bx || tx >= bx+bw {
				break
			}
			screen.SetContent(tx, tly, ch, nil, mutedStyle)
		}
	}

	// ── Prestige badge ────────────────────────────────────────────────────────
	if sc.prestigeLevel > 0 {
		badge := fmt.Sprintf("  ★  Prestige Level %d  ★  ", sc.prestigeLevel)
		bly := tly + 2
		blx := bx + (bw-len(badge))/2
		if bly < by+bh {
			cyanStyle := tcell.StyleDefault.
				Foreground(tcell.NewRGBColor(56, 189, 248)).
				Bold(true)
			for k, ch := range badge {
				tx := blx + k
				if tx < bx || tx >= bx+bw {
					break
				}
				screen.SetContent(tx, bly, ch, nil, cyanStyle)
			}
		}
	}
}
