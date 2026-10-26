package mapgen

// reservePlots keeps ground for the wonders along the trunk arms, one plot
// per age the ground can show, the nearest plots for the earliest ages. Plot
// n sits on arm n%4, on the arm's own side (the other when that is blocked),
// at the first place from its turn that fits.
func (b *planner) reservePlots() {
	p := b.plan
	for n := 0; n < groundAges; n++ {
		p.Plots = append(p.Plots, b.findPlot(b.arms[n%4], n/4, n))
	}
}

// findPlot looks along a trunk arm for room for the plot that is the
// step-th on that arm; none (a zero Lot) if the arm has no room left.
func (b *planner) findPlot(road, step, n int) Lot {
	if road < 0 {
		return Lot{}
	}
	r := b.plan.Roads[road]
	horizontal := b.heading(r)[1] == 0
	span := plotH + 2
	if horizontal {
		span = plotW + 2
	}
	for i := span/2 + 3 + step*span; i < len(r.Cells); i++ {
		for s := 0; s < 2; s++ {
			side := (n%4/2 + s) % 2
			lot := Lot{Rect: plotBeside(r.Cells[i], horizontal, side), Door: r.Cells[i], Road: road, At: i}
			if b.free(lot.Rect) {
				b.take(lot.Rect)
				return lot
			}
		}
	}
	return Lot{}
}

// plotBeside is the wonder plot on one side of road cell c.
func plotBeside(c Pt, horizontal bool, side int) Rect {
	switch {
	case horizontal && side == 0:
		return Rect{c.X - plotW/2, c.Y - plotH, plotW, plotH}
	case horizontal:
		return Rect{c.X - plotW/2, c.Y + 1, plotW, plotH}
	case side == 0:
		return Rect{c.X - plotW, c.Y - plotH/2, plotW, plotH}
	}
	return Rect{c.X + 1, c.Y - plotH/2, plotW, plotH}
}

// free says whether every cell of r is dry, buildable and unclaimed.
func (b *planner) free(r Rect) bool {
	ok := true
	r.cells(func(x, y int) {
		if !b.buildable(x, y) || b.claim[at(x, y)] != freeCell {
			ok = false
		}
	})
	return ok
}

func (b *planner) take(r Rect) {
	r.cells(func(x, y int) { b.claim[at(x, y)] = lotCell })
}

// cutLots lines every road with parcels on both sides, packing each as close
// as the last one allows.
func (b *planner) cutLots() {
	p := b.plan
	for ri, r := range p.Roads {
		if len(r.Cells) < 2 {
			continue
		}
		horizontal := b.heading(r)[1] == 0
		step, offset := lotH, lotH/2
		if horizontal {
			step, offset = lotW, lotW/2
		}
		next := [2]int{offset, offset}
		for i, c := range r.Cells {
			if p.Bridge[c] {
				continue
			}
			for side := 0; side < 2; side++ {
				if i < next[side] {
					continue
				}
				lot := lotBeside(c, horizontal, side)
				if !b.free(lot) {
					next[side] = i + 1
					continue
				}
				b.take(lot)
				p.Lots = append(p.Lots, Lot{Rect: lot, Door: c, Road: ri, At: i})
				next[side] = i + step
			}
		}
	}
}

// lotBeside is the parcel on one side of road cell c.
func lotBeside(c Pt, horizontal bool, side int) Rect {
	if horizontal {
		if side == 0 {
			return Rect{c.X - lotW/2, c.Y - lotH, lotW, lotH}
		}
		return Rect{c.X - lotW/2, c.Y + 1, lotW, lotH}
	}
	if side == 0 {
		return Rect{c.X - lotW, c.Y - lotH/2, lotW, lotH}
	}
	return Rect{c.X + 1, c.Y - lotH/2, lotW, lotH}
}
