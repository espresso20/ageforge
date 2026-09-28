package main

import (
	"container/heap"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The atlas model: everything the map shows, derived from one game state and
// the seeded world. It is a pure function of (seed, state), so the same save
// always draws the same map, and adding a building or meeting a civ only
// ever adds to it: claim orders, town sites and civ seats are fixed lists
// computed from the seed, and the state only decides how far along each list
// the map has got.

// Civ is a civilization on the map.
type Civ struct {
	Key       string
	Def       config.FactionDef
	Info      game.FactionInfo
	Known     bool
	Celestial bool // lives off-world: drawn only on the star charts
	Seat      Pt
	SeatName  string
	Regions   []int
	Fresh     bool // met since the last check-in
	Front     []Pt // war front, when at war
}

// Town is a settlement.
type Town struct {
	At      Pt
	Name    string
	Capital bool
	Owner   string // "" is the player; otherwise a civ key
	Fresh   bool
}

// Route is an active trade route drawn as a line.
type Route struct {
	Key, Name string
	Partner   string // civ key or "" for a foreign market
	Dest      string // name of the far end
	Path      []Pt
	Disrupted bool
	Export    map[string]float64
	Import    map[string]float64
	Fresh     bool
}

// Trail is an expedition's track.
type Trail struct {
	Key, Name, Category string
	Path                []Pt
	Runs                int
	LastTick            int
	Active              bool
	Progress            float64 // 0..1 along Path while active
}

// Scar is a catastrophe's mark on the land.
type Scar struct {
	Epoch, Name, Outcome string
	At                   Pt
	R                    float64
}

// Omen is a harbinger's origin.
type Omen struct {
	Name, Desc string
	At         Pt
	Active     bool
	Prob       float64
	Numeric    bool
	Target     string
}

// Landmark is a built wonder.
type Landmark struct {
	Key, Name string
	At        Pt
}

// Atlas is the whole derived model for one frame's state.
type Atlas struct {
	W         *World
	St        game.GameState
	Snap      *Snapshot
	AgeIdx    int
	Epoch     string
	Capital   Pt
	HomeName  string
	Owner     []int // per region: -1 free, 0 player, 1+ index into Civs +1
	Claim     []int // the player's claim order (regions)
	Owned     int   // how many of Claim are held
	Camp      float64
	Civs      []*Civ
	Towns     []Town
	Routes    []Route
	Trails    []Trail
	Scars     []Scar
	Omens     []Omen
	Wonders   []Landmark
	Camps     []Landmark // stone-era: each building type near the camp
	Buildings int
	knownR    float64
	knowAt    []disc
	FreshReg  map[int]bool
	Changes   []string // "since you last looked"
}

type disc struct {
	P Pt
	R float64
}

var terrestrialCivs = []struct {
	key    string
	target float64
	pref   string
}{
	{"riverlands_tribes", 42, "river"},
	{"ironhold_clans", 62, "high"},
	{"merchant_guild", 150, "coast"},
	{"artisan_league", 92, "low"},
	{"atomic_directorate", 128, "north"},
	{"tech_consortium", 175, "coast"},
	{"shadow_syndicate", 110, "low"},
	{"plasma_nomads", 145, "desert"},
}

func ageIndex(age string) int {
	for i, a := range config.AgeOrder() {
		if a == age {
			return i
		}
	}
	return 0
}

// buildAtlas derives the model. prev, if non-nil, is the state at the
// player's last check-in: whatever is new since then is marked fresh.
func buildAtlas(w *World, snap *Snapshot, prev *Snapshot) *Atlas {
	st := snap.State
	a := &Atlas{W: w, St: st, Snap: snap, AgeIdx: ageIndex(st.Age), Epoch: config.EpochForAge(st.Age)}
	for _, b := range st.Buildings {
		a.Buildings += b.Count
	}
	a.placeCapital()
	a.placeCivs()
	a.claimOrder()
	a.grow()
	a.placeTowns()
	a.placeTrails()
	a.placeRoutes()
	a.placeScars()
	a.placeOmens()
	a.placeWonders()
	a.placeFronts()
	a.knowledge()
	if prev != nil {
		a.diff(buildAtlas(w, prev, nil))
	}
	return a
}

func (a *Atlas) placeCapital() {
	w := a.W
	best, bs := Pt{150, 118}, math.Inf(-1)
	for i := 0; i < 600; i++ {
		p := Pt{110 + rnd01(w.Seed, i, 61)*80, 88 + rnd01(w.Seed, i, 62)*60}
		h := w.H(p.X, p.Y)
		if h < 0.04 || h > 0.3 {
			continue
		}
		score := -math.Abs(p.X-150)/40 - math.Abs(p.Y-118)/40
		// near a river mouth or a coast is where towns grow
		rd := 99.0
		for _, r := range w.Rivers {
			for _, q := range r {
				rd = math.Min(rd, q.Dist(p))
			}
		}
		score += 2 * math.Exp(-rd/3)
		// and within sight of the sea
		sea := 99.0
		for k := 0; k < 16; k++ {
			ang := float64(k) * math.Pi / 8
			for d := 2.0; d < 20; d += 2 {
				if w.H(p.X+math.Cos(ang)*d, p.Y+math.Sin(ang)*d) < 0 {
					sea = math.Min(sea, d)
					break
				}
			}
		}
		score += 1.6 * math.Exp(-sea/6)
		if sea > 16 {
			score -= 5 // a first camp within sight of the sea draws a better map
		}
		score += rnd01(w.Seed, i, 63) * 0.3
		if score > bs {
			best, bs = p, score
		}
	}
	a.Capital = best
	if r := w.RegionAt(best.X, best.Y); r >= 0 {
		a.HomeName = w.Regions[r].Name
	}
}

func (a *Atlas) placeCivs() {
	w := a.W
	a.Owner = make([]int, len(w.Regions))
	for i := range a.Owner {
		a.Owner[i] = -1
	}
	home := w.RegionAt(a.Capital.X, a.Capital.Y)
	if home >= 0 {
		a.Owner[home] = 0
	}
	defs := config.FactionByKey()
	taken := map[int]bool{home: true}
	names := newNamer(w.Seed + 505)
	for _, tc := range terrestrialCivs {
		def := defs[tc.key]
		best, bs := -1, math.Inf(1)
		for _, r := range w.Regions {
			if r.Area < 120 || taken[r.ID] {
				continue
			}
			near := false
			for t := range taken {
				if w.Regions[t].Adj[r.ID] && t != home {
					near = true
				}
			}
			d := r.Seed.Dist(a.Capital)
			s := math.Abs(d-tc.target) + rnd01(w.Seed, r.ID, int(hashStr(tc.key)%1000))*14
			if near {
				s += 25
			}
			switch tc.pref {
			case "river":
				s -= 12 * float64(countRiversNear(w, r.Seed, 10))
			case "high":
				s -= r.MeanH * 60
			case "low":
				s += r.MeanH * 30
			case "coast":
				if !r.Coastal {
					s += 40
				}
			case "north":
				s += r.Seed.Y / 4
			case "desert":
				if r.Biome != "desert" && r.Biome != "steppe" {
					s += 35
				}
			}
			if s < bs {
				best, bs = r.ID, s
			}
		}
		c := &Civ{Key: tc.key, Def: def, Info: a.St.Diplomacy.Factions[tc.key]}
		c.Known = c.Info.Discovered
		if best >= 0 {
			taken[best] = true
			c.Seat = w.Regions[best].Seed
			c.SeatName = names.town()
			c.Regions = []int{best}
			a.Owner[best] = len(a.Civs) + 1
			// neighbours by distance, as many as the civ is strong
			n := def.Strength
			if n > 4 {
				n = 4
			}
			for _, nb := range sortedAdj(w, best) {
				if len(c.Regions) >= n {
					break
				}
				if a.Owner[nb] == -1 && nb != home {
					a.Owner[nb] = len(a.Civs) + 1
					taken[nb] = true
					c.Regions = append(c.Regions, nb)
				}
			}
		}
		a.Civs = append(a.Civs, c)
	}
	for _, key := range []string{"stellar_federation", "void_reavers", "quantum_collective"} {
		c := &Civ{Key: key, Def: defs[key], Info: a.St.Diplomacy.Factions[key], Celestial: true}
		c.Known = c.Info.Discovered
		c.SeatName = names.word()
		a.Civs = append(a.Civs, c)
	}
}

func countRiversNear(w *World, p Pt, r float64) int {
	n := 0
	for _, rv := range w.Rivers {
		for _, q := range rv {
			if q.Dist(p) < r {
				n++
				break
			}
		}
	}
	return n
}

func sortedAdj(w *World, id int) []int {
	var out []int
	for k := range w.Regions[id].Adj {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		di := w.Regions[out[i]].Seed.Dist(w.Regions[id].Seed)
		dj := w.Regions[out[j]].Seed.Dist(w.Regions[id].Seed)
		if di != dj {
			return di < dj
		}
		return out[i] < out[j]
	})
	return out
}

// claimOrder is the fixed order the player's realm grows in: a frontier
// search from the home region, nearest first, skipping civ lands.
func (a *Atlas) claimOrder() {
	w := a.W
	home := w.RegionAt(a.Capital.X, a.Capital.Y)
	if home < 0 {
		return
	}
	seen := map[int]bool{home: true}
	a.Claim = []int{home}
	frontier := []int{home}
	for len(frontier) > 0 {
		sort.Slice(frontier, func(i, j int) bool {
			di := w.Regions[frontier[i]].Seed.Dist(a.Capital)
			dj := w.Regions[frontier[j]].Seed.Dist(a.Capital)
			if di != dj {
				return di < dj
			}
			return frontier[i] < frontier[j]
		})
		cur := frontier[0]
		frontier = frontier[1:]
		if cur != home {
			a.Claim = append(a.Claim, cur)
		}
		for _, nb := range sortedAdj(w, cur) {
			if seen[nb] || (a.Owner[nb] > 0) {
				continue
			}
			seen[nb] = true
			frontier = append(frontier, nb)
		}
	}
}

// grow decides how much of the claim order the realm holds. The Stone Era
// has no borders yet, only a camp and its hunting grounds.
func (a *Atlas) grow() {
	b := float64(a.Buildings)
	if a.AgeIdx < 3 {
		a.Camp = 4 + 1.6*math.Sqrt(b)
		a.Owned = 0
		return
	}
	n := 1 + int(float64(a.AgeIdx-3)*0.9+math.Log2(1+b/30))
	if n > len(a.Claim) {
		n = len(a.Claim)
	}
	a.Owned = n
	for _, r := range a.Claim[:n] {
		a.Owner[r] = 0
	}
}

func (a *Atlas) placeTowns() {
	w := a.W
	names := newNamer(w.Seed + 303)
	capName := names.town()
	a.Towns = append(a.Towns, Town{At: a.Capital, Name: capName, Capital: true})
	for i, r := range a.Claim {
		name := names.town() // drawn for every claim so names never shift
		if i == 0 || i >= a.Owned {
			continue
		}
		a.Towns = append(a.Towns, Town{At: w.Regions[r].Seed, Name: name})
	}
	for _, c := range a.Civs {
		if c.Celestial || len(c.Regions) == 0 {
			continue
		}
		a.Towns = append(a.Towns, Town{At: c.Seat, Name: c.SeatName, Capital: true, Owner: c.Key})
	}
	// Stone-era camp marks: one per building type, on a fixed spiral so a
	// new type never moves an old one.
	if a.AgeIdx < 3 {
		keys := make([]string, 0)
		for k, b := range a.St.Buildings {
			if b.Count > 0 {
				keys = append(keys, k)
			}
		}
		sort.Slice(keys, func(i, j int) bool { return hashStr(keys[i]) < hashStr(keys[j]) })
		for _, k := range keys {
			i := int(hashStr(k) % 12)
			ang := float64(i)*2.39996 + 0.4
			r := 1.8 + float64(i%4)*0.9
			p := Pt{a.Capital.X + math.Cos(ang)*r*1.6, a.Capital.Y + math.Sin(ang)*r}
			a.Camps = append(a.Camps, Landmark{Key: k, Name: a.St.Buildings[k].Name, At: p})
		}
	}
}

func (a *Atlas) placeTrails() {
	type agg struct {
		e        JournalEntry
		runs     int
		lastTick int
	}
	by := map[string]*agg{}
	var order []string
	for _, e := range a.Snap.Journal {
		g, ok := by[e.Key]
		if !ok {
			g = &agg{e: e}
			by[e.Key] = g
			order = append(order, e.Key)
		}
		g.runs++
		g.lastTick = e.Tick
	}
	active := ""
	if s := a.St.Military.ActiveScout; s != nil {
		active = s.Name
	}
	for _, k := range order {
		g := by[k]
		t := Trail{Key: k, Name: g.e.Name, Category: g.e.Category, Runs: g.runs, LastTick: g.lastTick}
		t.Path = a.trailPath(k, ageIndex(g.e.Age))
		if active != "" && g.e.Name == active {
			t.Active = true
			for _, ex := range a.St.Military.Expeditions {
				if ex.Name == active {
					dur := float64(ex.DurationMin+ex.DurationMax) / 2
					t.Progress = clamp(1-float64(a.St.Military.ActiveScout.TicksLeft)/dur, 0.02, 1)
				}
			}
		}
		a.Trails = append(a.Trails, t)
	}
}

func (a *Atlas) trailPath(key string, firstAge int) []Pt {
	h := hashStr(key)
	ang := rnd01(h, 1, 1) * 2 * math.Pi
	d := 18 + float64(firstAge)*7 + rnd01(h, 2, 2)*16
	dest := Pt{a.Capital.X + math.Cos(ang)*d, a.Capital.Y + math.Sin(ang)*d*0.7}
	dest.X = clamp(dest.X, 10, worldW-10)
	dest.Y = clamp(dest.Y, 10, worldH-10)
	var path []Pt
	n := 40
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		wob := math.Sin(t*math.Pi) * (rnd01(h, 3, 3) - 0.5) * d * 0.5
		wob += math.Sin(t*math.Pi*3+rnd01(h, 4, 4)*6) * d * 0.05
		px := lerp(a.Capital.X, dest.X, t) - math.Sin(ang)*wob
		py := lerp(a.Capital.Y, dest.Y, t) + math.Cos(ang)*wob*0.7
		path = append(path, Pt{px, py})
	}
	return path
}

func (a *Atlas) placeRoutes() {
	defs := config.FactionByKey()
	for i, r := range a.St.Trade.ActiveRoutes {
		rt := Route{Key: r.Key, Name: r.Name, Disrupted: r.Disrupted, Export: r.Export, Import: r.Import}
		var dest Pt
		found := false
		for _, c := range a.Civs {
			if c.Celestial || !c.Known {
				continue
			}
			if _, ok := r.Import[c.Def.Specialty]; ok {
				dest, rt.Partner, rt.Dest, found = c.Seat, c.Key, defs[c.Key].Name, true
				break
			}
		}
		if !found {
			// a foreign market: a coastal region outside the realm, fixed by the key
			var cands []*Region
			for _, reg := range a.W.Regions {
				if reg.Coastal && a.Owner[reg.ID] != 0 && reg.Area > 80 {
					d := reg.Seed.Dist(a.Capital)
					if d > 30 && d < 40+float64(a.AgeIdx)*9 {
						cands = append(cands, reg)
					}
				}
			}
			if len(cands) == 0 {
				continue
			}
			reg := cands[int(uint64(hashStr(r.Key)+int64(i))%uint64(len(cands)))]
			dest, rt.Dest = reg.Seed, reg.Name
		}
		rt.Path = a.W.route(a.Capital, dest)
		a.Routes = append(a.Routes, rt)
	}
}

func (a *Atlas) placeScars() {
	owned := a.Claim
	if a.Owned > 0 {
		owned = a.Claim[:a.Owned]
	}
	epochs := map[string]config.EpochDef{}
	for _, e := range config.Epochs() {
		epochs[e.Key] = e
	}
	for i, ev := range a.St.EpochEventHistory {
		if ev.EventType != "catastrophe" {
			continue
		}
		h := hashStr(ev.EpochKey + ev.EventKey)
		var at Pt
		if len(owned) > 0 {
			r := a.W.Regions[owned[int(uint64(h)%uint64(len(owned)))]]
			at = Pt{r.Seed.X + (rnd01(h, i, 1)-0.5)*12, r.Seed.Y + (rnd01(h, i, 2)-0.5)*8}
		} else {
			at = Pt{a.Capital.X + 14, a.Capital.Y - 6}
		}
		a.Scars = append(a.Scars, Scar{Epoch: ev.EpochKey, Name: ev.EventName, Outcome: ev.Outcome, At: at, R: 5})
	}
}

func (a *Atlas) placeOmens() {
	put := func(key string) Pt {
		h := hashStr(key)
		ang := rnd01(h, 9, 9) * 2 * math.Pi
		r := 26 + float64(a.AgeIdx)*6
		return Pt{clamp(a.Capital.X+math.Cos(ang)*r, 8, worldW-8), clamp(a.Capital.Y+math.Sin(ang)*r*0.6, 8, worldH-8)}
	}
	for _, h := range a.St.HarbingerHistory {
		a.Omens = append(a.Omens, Omen{Name: h.Name, At: put(h.EpochKey + h.Name)})
	}
	if v := a.St.Harbinger; v != nil {
		a.Omens = append(a.Omens, Omen{Name: v.Name, Desc: v.Description, At: put(v.Key + v.Name), Active: true,
			Prob: v.Probability, Numeric: v.Numeric, Target: v.TargetEpochName})
	}
}

func (a *Atlas) placeWonders() {
	var keys []string
	for k, b := range a.St.Buildings {
		if b.Count > 0 && b.Category == "wonder" {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return a.St.Buildings[keys[i]].AgeKey < a.St.Buildings[keys[j]].AgeKey || (a.St.Buildings[keys[i]].AgeKey == a.St.Buildings[keys[j]].AgeKey && keys[i] < keys[j])
	})
	for _, k := range keys {
		i := int(hashStr(k) % 16)
		ang := float64(i) * 2.39996
		r := 3.5 + float64(i%3)*1.4
		a.Wonders = append(a.Wonders, Landmark{Key: k, Name: a.St.Buildings[k].Name,
			At: Pt{a.Capital.X + math.Cos(ang)*r*1.8, a.Capital.Y + math.Sin(ang)*r}})
	}
}

func (a *Atlas) placeFronts() {
	w := a.W
	for ci, c := range a.Civs {
		if c.Celestial || !c.Known || !c.Info.AtWar {
			continue
		}
		id := ci + 1
		for y := 2; y < worldH-2; y += 2 {
			for x := 2; x < worldW-2; x += 2 {
				r := w.RegionAt(float64(x), float64(y))
				if r < 0 || a.Owner[r] != 0 {
					continue
				}
				for _, d := range [][2]int{{2, 0}, {0, 2}, {-2, 0}, {0, -2}} {
					o := w.RegionAt(float64(x+d[0]), float64(y+d[1]))
					if o >= 0 && a.Owner[o] == id {
						c.Front = append(c.Front, Pt{float64(x) + float64(d[0])/2, float64(y) + float64(d[1])/2})
						break
					}
				}
			}
		}
		if len(c.Front) == 0 {
			// no shared border: the war is fought halfway between the capitals
			mid := Pt{(a.Capital.X + c.Seat.X) / 2, (a.Capital.Y + c.Seat.Y) / 2}
			dx, dy := c.Seat.X-a.Capital.X, c.Seat.Y-a.Capital.Y
			l := math.Hypot(dx, dy)
			for t := -6.0; t <= 6; t += 1.5 {
				c.Front = append(c.Front, Pt{mid.X - dy/l*t, mid.Y + dx/l*t*0.6})
			}
		}
	}
}

// knowledge builds the discs of the known world. The Digital Era on sees
// the whole planet from orbit.
func (a *Atlas) knowledge() {
	n := len(a.Snap.Journal)
	a.knownR = 14 + float64(a.AgeIdx)*7.5 + math.Min(math.Sqrt(float64(n))*0.8, 30)
	a.knowAt = append(a.knowAt, disc{a.Capital, a.knownR})
	for i := 0; i < a.Owned; i++ {
		a.knowAt = append(a.knowAt, disc{a.W.Regions[a.Claim[i]].Seed, 20})
	}
	for _, c := range a.Civs {
		if c.Known && !c.Celestial {
			a.knowAt = append(a.knowAt, disc{c.Seat, 18})
		}
	}
	for _, t := range a.Trails {
		for i := 0; i < len(t.Path); i += 3 {
			r := 5.0
			if i == len(t.Path)-1 || i+3 >= len(t.Path) {
				r = 9
			}
			a.knowAt = append(a.knowAt, disc{t.Path[i], r})
		}
	}
	for _, r := range a.Routes {
		for i := 0; i < len(r.Path); i += 3 {
			a.knowAt = append(a.knowAt, disc{r.Path[i], 6})
		}
	}
}

// Known is how well the atlas knows (x, y): 1 charted, 0 terra incognita,
// with a ragged, hand-drawn edge.
func (a *Atlas) Known(x, y float64) float64 {
	if a.AgeIdx >= 12 {
		return 1
	}
	best := 0.0
	jit := (fbm(a.W.Seed+90, x/9, y/9, 3) - 0.5) * 7
	for _, d := range a.knowAt {
		dx, dy := x-d.P.X, y-d.P.Y
		dd := math.Sqrt(dx*dx+dy*dy) + jit
		v := clamp((d.R-dd)/3+0.5, 0, 1)
		if v > best {
			best = v
			if best >= 1 {
				return 1
			}
		}
	}
	return best
}

// OwnerAt is the owner id at (x, y): -1 none/sea, 0 player, 1+ civ.
func (a *Atlas) OwnerAt(x, y float64) int {
	if a.AgeIdx < 3 {
		if math.Hypot(x-a.Capital.X, (y-a.Capital.Y)*1.3) < a.Camp && a.W.H(x, y) > 0 {
			return 0
		}
		return -1
	}
	r := a.W.RegionAt(x, y)
	if r < 0 {
		return -1
	}
	o := a.Owner[r]
	if o > 0 && !a.Civs[o-1].Known {
		return -1
	}
	return o
}

func (a *Atlas) diff(p *Atlas) {
	a.FreshReg = map[int]bool{}
	gained := 0
	for i := 0; i < a.Owned; i++ {
		r := a.Claim[i]
		if i >= p.Owned {
			a.FreshReg[r] = true
			gained++
		}
	}
	if gained > 0 {
		a.Changes = append(a.Changes, plural(gained, "province", "provinces")+" annexed")
	}
	for i, c := range a.Civs {
		if c.Known && !p.Civs[i].Known {
			c.Fresh = true
			a.Changes = append(a.Changes, "met the "+c.Def.Name)
		} else if c.Known && c.Info.AtWar && !p.Civs[i].Info.AtWar {
			a.Changes = append(a.Changes, "war with the "+c.Def.Name)
		}
	}
	had := map[string]bool{}
	for _, r := range p.Routes {
		had[r.Key] = true
	}
	for i := range a.Routes {
		if !had[a.Routes[i].Key] {
			a.Routes[i].Fresh = true
			a.Changes = append(a.Changes, "route opened: "+a.Routes[i].Name)
		}
	}
	if d := len(a.Snap.Journal) - len(p.Snap.Journal); d > 0 {
		a.Changes = append(a.Changes, plural(d, "expedition", "expeditions")+" returned")
	}
	if d := len(a.Wonders) - len(p.Wonders); d > 0 {
		a.Changes = append(a.Changes, plural(d, "wonder", "wonders")+" raised")
	}
	if d := len(a.Scars) - len(p.Scars); d > 0 {
		a.Changes = append(a.Changes, "a catastrophe scarred the land")
	}
	for i := range a.Towns {
		fresh := true
		for _, t := range p.Towns {
			if t.Name == a.Towns[i].Name {
				fresh = false
			}
		}
		a.Towns[i].Fresh = fresh
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return itoa(n) + " " + many
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

func titleKey(k string) string {
	parts := strings.Split(k, "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// ---------------------------------------------------------------------------
// route finding on a 2-unit grid: sea is cheap (ships), mountains dear.

type pqItem struct {
	i int
	f float64
}
type pq []pqItem

func (q pq) Len() int            { return len(q) }
func (q pq) Less(i, j int) bool  { return q[i].f < q[j].f }
func (q pq) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *pq) Push(x interface{}) { *q = append(*q, x.(pqItem)) }
func (q *pq) Pop() interface{} {
	o := *q
	it := o[len(o)-1]
	*q = o[:len(o)-1]
	return it
}

func (w *World) route(from, to Pt) []Pt {
	const s = 2
	gw, gh := worldW/s, worldH/s
	idx := func(x, y int) int { return y*gw + x }
	sx, sy := int(from.X)/s, int(from.Y)/s
	tx, ty := int(to.X)/s, int(to.Y)/s
	cost := func(x, y int) float64 {
		h := w.gh(x*s, y*s)
		switch {
		case h < 0:
			return 1.0
		case h > 0.6:
			return 7
		case h > 0.35:
			return 3
		}
		return 1.7
	}
	g := make([]float64, gw*gh)
	came := make([]int, gw*gh)
	for i := range g {
		g[i] = math.Inf(1)
		came[i] = -1
	}
	start, goal := idx(sx, sy), idx(tx, ty)
	g[start] = 0
	q := &pq{{start, 0}}
	for q.Len() > 0 {
		cur := heap.Pop(q).(pqItem)
		if cur.i == goal {
			break
		}
		cx, cy := cur.i%gw, cur.i/gw
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := cx+dx, cy+dy
				if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
					continue
				}
				step := cost(nx, ny)
				if dx != 0 && dy != 0 {
					step *= 1.414
				}
				ng := g[cur.i] + step
				ni := idx(nx, ny)
				if ng < g[ni] {
					g[ni] = ng
					came[ni] = cur.i
					hx, hy := float64(tx-nx), float64(ty-ny)
					heap.Push(q, pqItem{ni, ng + math.Hypot(hx, hy)})
				}
			}
		}
	}
	var path []Pt
	for i := goal; i >= 0; i = came[i] {
		path = append([]Pt{{float64(i%gw*s) + 1, float64(i/gw*s) + 1}}, path...)
		if i == start {
			break
		}
	}
	// smooth the staircase
	for pass := 0; pass < 3; pass++ {
		for i := 1; i+1 < len(path); i++ {
			path[i] = Pt{(path[i-1].X + 2*path[i].X + path[i+1].X) / 4, (path[i-1].Y + 2*path[i].Y + path[i+1].Y) / 4}
		}
	}
	return path
}
