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
	sc.DrawForSubclass(screen, sc)
	bx, by, bw, bh := sc.GetInnerRect()

	sc.mu.RLock()
	tick := sc.tick
	sc.mu.RUnlock()

	// ── Pre-compute all text positions ───────────────────────────────────────
	// Must happen before the star loop so we can exclude the text zone.

	lines := strings.Split(strings.TrimSpace(SplashArt), "\n")
	titleH := len(lines)
	titleW := 0
	for _, l := range lines {
		n := 0
		for range l {
			n++
		}
		if n > titleW {
			titleW = n
		}
	}
	// Vertically centre title in the upper 45% of the canvas.
	originY := by + bh*9/20 - titleH/2
	originX := bx + (bw-titleW)/2

	tagline := SplashTagline
	tly := originY + titleH + 2
	tlx := bx + (bw-len(tagline))/2

	// Bounding box that excludes stars — title block + tagline + 1-cell margin.
	exY2 := tly + 2
	if sc.prestigeLevel > 0 {
		exY2 += 2
	}
	exX1, exY1 := originX-1, originY-1
	exX2 := originX + titleW + 1

	// ── Stars (skip the text exclusion zone) ─────────────────────────────────
	for _, s := range sc.stars {
		px := bx + int(s.nx*float64(bw))
		py := by + int(s.ny*float64(bh))
		if px < bx || px >= bx+bw || py < by || py >= by+bh {
			continue
		}
		// Leave the text area completely clear of stars.
		if px >= exX1 && px < exX2 && py >= exY1 && py < exY2 {
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
	pulse := (math.Sin(tick*0.55) + 1.0) / 2.0
	goldR := int32(182 + int(pulse*73))
	goldG := int32(112 + int(pulse*68))
	titleFG := tcell.NewRGBColor(goldR, goldG, 0)
	titleStyle := tcell.StyleDefault.Foreground(titleFG).Bold(true)
	blank := tcell.StyleDefault

	for i, line := range lines {
		ty := originY + i
		if ty < by || ty >= by+bh {
			continue
		}
		col := 0
		for _, ch := range line {
			tx := originX + col
			col++
			if tx < bx || tx >= bx+bw {
				continue
			}
			// Draw spaces explicitly — this erases any content underneath.
			// Without this, stars drawn in a previous tick's space cells linger.
			if ch == ' ' {
				screen.SetContent(tx, ty, ' ', nil, blank)
			} else {
				screen.SetContent(tx, ty, ch, nil, titleStyle)
			}
		}
		// Clear any remaining columns in this row for lines shorter than titleW.
		for col < titleW {
			tx := originX + col
			col++
			if tx >= bx && tx < bx+bw {
				screen.SetContent(tx, ty, ' ', nil, blank)
			}
		}
	}

	// ── Tagline ───────────────────────────────────────────────────────────────
	if tly >= by && tly < by+bh {
		mutedStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(139, 148, 158))
		// Wipe the row across the title width first.
		for col := 0; col < titleW; col++ {
			tx := originX + col
			if tx >= bx && tx < bx+bw {
				screen.SetContent(tx, tly, ' ', nil, blank)
			}
		}
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
