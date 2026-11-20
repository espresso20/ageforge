package planboard

import (
	"image"
	"image/color"
	"time"
)

// Mode is where a row sits on the ruler.
type Mode uint8

const (
	Timed  Mode = iota // a bar from Start to End
	Parked             // the estimate puts no time on it: the hatched gutter
	After              // behind the gate: a block in the next age's tint
)

// A Span is a start and an end.
type Span struct{ Start, End time.Duration }

// A Row is one plan item as the schedule draws it.
type Row struct {
	Seed, Sign string        // the item's kind and key; the kind's sign
	Mode       Mode          //
	Start, End time.Duration // Timed only
	Copies     []time.Duration
	Cause      string // "" ready; funds, research, age, income, hand
	Gate       bool   // the advance
	Old        *Span  // the bar's place before an optimize
}

// A Sched is everything the schedule picture depends on.
type Sched struct {
	Rows              []Row
	First, Sel        int  // the first item shown; the selected one (-1 none)
	Drop              int  // a drag's drop line goes above this item (-1 none)
	Paused            bool // bars are outlines, lamps grey
	FinishW, UntimedW int  // the two title texts' widths, in cells
}

// A Tick is a mark of the ruler: its cell (panel column) and name.
type Tick struct {
	Cell int
	Name string
}

// Drawn is a schedule picture and where it lies.
type Drawn struct {
	Rect     image.Rectangle // cells of the panel
	Img      *image.RGBA
	Slots    []Slot
	More     int
	Ticks    []Tick
	GateX    int // the gate's column, -1 when there is none
	SelX     int // the selected bar's start column, -1 when it has none
	Finish   time.Duration
	GateRect image.Rectangle // not used when GateX is -1
}

// Lamp is a lamp's colour for a cause.
func Lamp(cause string, paused bool) color.RGBA {
	switch {
	case paused:
		return Grey
	case cause == "":
		return Ok
	case cause == "funds" || cause == "hand":
		return Wait
	case cause == "income":
		return Bad
	}
	return Steel
}

// Schedule draws the list's marks, the title block's rules, the ruler and the
// lanes. Rect is from the title row to the last schedule row, from the list's
// left edge to the gap before the right column.
func (l Layout) Schedule(s Sched) Drawn {
	d := Drawn{Rect: image.Rect(l.ListX, 2, l.Lanes.Max.X+1, l.SchedEnd), GateX: -1, SelX: -1}
	c := part(d.Rect, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	var advs []int
	var marks []time.Duration
	var gate time.Duration
	adv, parked := -1, false
	for i, r := range s.Rows {
		if r.Gate {
			advs = append(advs, i)
			if adv < 0 {
				adv = i
			}
		}
		if r.Mode == Timed && !r.Gate {
			marks = append(marks, r.Start, r.End)
			if adv < 0 {
				gate = max(gate, r.End)
			}
			d.Finish = max(d.Finish, r.End)
		} else if r.Mode == Parked && adv < 0 {
			parked = true
		}
	}
	d.Finish = max(d.Finish, gate)
	if adv >= 0 {
		marks = append(marks, gate)
	}
	d.Slots, d.More = l.Slots(s.First, len(s.Rows), advs)
	ax := newAxis(marks, d.Finish, l.Lanes.Min.X+1, l.RulerCells, cw)
	gateX := -1
	if adv >= 0 {
		gateX = ax.x(gate) + 10
		d.GateX = gateX / cw
	}
	top, bot := l.ListY*ch, l.SchedEnd*ch
	laneX := l.Lanes.Min.X * cw
	end := ax.end()
	// Title block rules: the plate and the age box, the finish and the untimed count.
	line := Mix(Sheet, Ink, .4)
	c.frame(image.Rect(3, 2, 32, 4), line)
	c.vline(13*cw-1, 2*ch, 2*ch, line)
	if s.FinishW > 0 {
		x := l.Lanes.Min.X + 1
		c.frame(image.Rect(x, 2, x+s.FinishW, 3), line)
		x += s.FinishW + 1
		if s.UntimedW > 0 {
			c.frame(image.Rect(x, 2, x+s.UntimedW, 3), line)
		}
	}
	// Lanes: the ground, the gate's band, the grid.
	c.cells(l.Lanes, Sheet)
	if gateX >= 0 {
		c.rect(gateX, top, end-gateX, bot-top, Sheet2)
	}
	for gx := laneX; gx < end; gx += 4 * cw {
		c.vline(gx, top, bot-top, Grid)
	}
	for _, sl := range d.Slots {
		if sl.Item >= 0 {
			c.hline(laneX, sl.Y*ch, end-laneX, Grid)
		}
	}
	c.hline(laneX, bot-1, end-laneX, Grid)
	sel := -1
	for _, sl := range d.Slots {
		if sl.Item == s.Sel {
			sel = sl.Y
		}
	}
	if sel >= 0 {
		c.rect(laneX, sel*ch, l.Lanes.Dx()*cw, l.RowH*ch, Band)
	}
	for _, x := range ax.xs {
		c.dotted(x, top, bot-top, false, Mix(Sheet, Dim, .3), 1, 3)
	}
	px := l.Park.Min.X * cw
	c.rect(px, top, l.Park.Dx()*cw, bot-top, Mix(Sheet, black, .25))
	c.hatch(image.Rect(px, top, px+l.Park.Dx()*cw, bot), Mix(Sheet, Dim, .3), 7, 2)
	// Ruler.
	c.cells(image.Rect(l.Ruler.Min.X, l.Ruler.Min.Y, l.Ruler.Max.X, l.Ruler.Max.Y), Dark)
	for i := 0; i < len(ax.xs)-1; i++ {
		c.rect(ax.xs[i], 4*ch, ax.xs[i+1]-ax.xs[i], 2*ch, Ruler[i%2])
	}
	for i, x := range ax.xs {
		c.vline(x, 5*ch+3, 11, Brass)
		d.Ticks = append(d.Ticks, Tick{x / cw, LadderNames[i]})
	}
	if gateX >= 0 {
		dash := func(y, h int) {
			if parked {
				c.dotted(gateX-1, y, h, false, Ink, 4, 3)
				c.dotted(gateX, y, h, false, Ink, 4, 3)
				c.dotted(gateX+1, y, h, false, Ink, 4, 3)
			} else {
				c.rect(gateX-1, y, 3, h, Ink)
			}
		}
		dash(6*ch, bot-6*ch)
		c.arch(gateX-10, 4*ch, 21, 2*ch, Brass)
		c.arch(gateX-6, 4*ch+4, 13, 2*ch-4, Dark)
		c.rect(gateX-1, 4*ch+10, 3, 2*ch-10, Ink)
	}
	// Rows.
	late, half := 0, l.RowH*ch/2
	for _, sl := range d.Slots {
		if sl.Item < 0 {
			c.hline((l.ListX+1)*cw, sl.Y*ch+7, (l.ListW-2)*cw, Mix(Sheet, Brass, .5))
			continue
		}
		r := s.Rows[sl.Item]
		rn := rng(seed(r.Seed))
		yc := sl.Y*ch + half
		ground := Sheet
		if sl.Item == s.Sel {
			ground = Band
			c.cells(image.Rect(l.ListX, sl.Y, l.ListX+l.ListW, sl.Y+l.RowH), Band)
		}
		for _, dy := range []int{-6, -2, 2, 6} {
			if l.RowH == 1 && dy%4 != 2 {
				continue
			}
			c.rect(l.ListX*cw+1+rn.jit(), yc+dy/l.RowH+rn.jit(), 2, 2, Dim)
			c.rect(l.ListX*cw+4+rn.jit(), yc+dy/l.RowH+rn.jit(), 2, 2, Dim)
		}
		if l.RowH == 2 {
			c.sign(r.Sign, (l.ListX+3)*cw, sl.Y*ch, rn.n(3), Sheet)
		} else {
			c.signSmall(r.Sign, (l.ListX+3)*cw, sl.Y*ch, rn.n(3), Sheet)
		}
		lc := Lamp(r.Cause, s.Paused)
		lx, ly := (l.ListX+l.ListW-3)*cw+3+rn.n(3), yc+rn.jit()
		c.disc(lx, ly, 4, Mix(lc, black, .4))
		c.disc(lx, ly, 3, lc)
		c.pt(lx-1+rn.jit()/2, ly-2, Mix(lc, white, .6))
		if r.Cause == "hand" && !s.Paused {
			c.rect(lx-3, ly-1, 7, 4, black)
			for q := 0; q < 3; q++ {
				c.rect(lx-3+q*3, ly-4, 1, 4, black)
			}
		}
		switch {
		case r.Gate:
		case r.Mode == After:
			c.block(min(gateX+10+late*10, end-cw-2), yc, cw, Late, s.Paused)
			late++
		case r.Mode == Parked:
			x := px + 6 + rn.n(3)*(l.Park.Dx()*cw-30)/3
			switch r.Cause {
			case "income":
				c.disc(x+6, yc, 6, Bad)
				c.rect(x+2, yc-1, 9, 3, white)
			case "hand":
				c.rect(x, yc-1, 14, 8, Wait)
				for q := 0; q < 4; q++ {
					c.rect(x+q*4, yc-6-q%2, 3, 6, Wait)
				}
			default:
				c.block(x, yc, 14, Dim, true)
			}
		default:
			bx := ax.x(r.Start)
			ex := max(ax.x(r.End), bx+2*cw)
			if r.End <= r.Start {
				ex = bx + cw
			}
			if o := r.Old; o != nil {
				ox := ax.x(o.Start)
				oe := max(ax.x(o.End), ox+2*cw)
				c.outline(image.Rect(ox, yc-8+l.RowH/2, oe, yc+8-l.RowH/2), Dim)
			}
			if bx > ax.xs[0] {
				c.dotted(ax.xs[0], yc, bx-ax.xs[0], true, Mix(Sheet, lc, .85), 2, 3)
			}
			c.block(bx, yc, ex-bx, lc, s.Paused)
			for _, cp := range r.Copies[min(1, len(r.Copies)):] {
				if x := ax.x(cp); x > bx && x < ex {
					c.vline(x, yc-6, 12, ground)
				}
			}
			if sl.Item == s.Sel {
				d.SelX = bx / cw
				c.dotted(bx, 5*ch+3, yc-(5*ch+3), false, Brass, 2, 2)
			}
		}
	}
	// The scroll rail, the selection's brackets and the drop line.
	if d.More > 0 || s.First > 0 {
		t, b := (l.ListY+1)*ch, (l.SchedEnd-1)*ch
		n := max(len(s.Rows), 1)
		c.rect((l.ListX+l.ListW-1)*cw+2, t, 3, b-t, Mix(Sheet, Dim, .3))
		c.rect((l.ListX+l.ListW-1)*cw+2, t+(b-t)*s.First/n, 3, max((b-t)*(len(s.Rows)-s.First-d.More)/n, 6), Brass)
	}
	if sel >= 0 {
		for _, bxp := range []int{(l.ListX+l.ListW-1)*cw + 2, l.Lanes.Max.X*cw + 2} {
			dx := 0
			if bxp > l.Lanes.Min.X*cw {
				dx = -4
			}
			c.rect(bxp, sel*ch+1, 3, l.RowH*ch-2, Brass)
			c.rect(bxp+dx, sel*ch+1, 7, 2, Brass)
			c.rect(bxp+dx, sel*ch+l.RowH*ch-3, 7, 2, Brass)
		}
	}
	if s.Drop >= 0 {
		y := l.SchedEnd * ch
		for _, sl := range d.Slots {
			if sl.Item == s.Drop {
				y = sl.Y * ch
			}
		}
		c.rect(l.ListX*cw, y-1, (l.Lanes.Max.X-l.ListX)*cw, 3, Brass)
	}
	d.Img = c.img
	return d
}
