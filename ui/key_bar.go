package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// key_bar.go makes the keys along a panel's foot into buttons: where they are
// in the grid and what a click on them does. A click on one presses its key
// as the keyboard would. The close button in a panel's top edge is one more of
// them: it presses Esc. How they look is button.go's.

// xkey is a key a click can press: the key, the cells that press it, and the
// cells the key and its word take.
type xkey struct {
	key kit.Key
	r   image.Rectangle
	// cap is the cells of the key's own cap; the keys of one cap share it.
	// Empty for a key that is only a hit (a card's action in a scrolled view).
	cap image.Rectangle
	// word is the cells of the word beside the cap, empty when there is none.
	word image.Rectangle
	kind keyKind
	// off marks a key that is drawn but unavailable; still, one that names no
	// key to press ("1-9"): drawn as a key, never pressed.
	off, still bool
	// why says in a line why the key is off, shown as a tooltip while the
	// pointer rests on it; off alone decides whether the key is off.
	why string
	// cmd is the command line a click runs, through the path a typed line
	// takes, in place of pressing key. Empty for a key that only presses.
	cmd string
	// ask is what a click asks first, when the command is one that needs a
	// yes; the command runs only on the answer. Empty to run at once.
	ask string
	// yes is the word on the danger button of that question.
	yes string
	// sign is the picture at the left of an action button.
	sign signKind
}

// tipKey is the target of the tooltip of one key of one panel.
type tipKey struct {
	k *keyPresser
	r image.Rectangle
}

// keySet is the buttons a panel drew and what it needs to read them back:
// every grid that holds buttons has one, and so does a panel drawn on a grid
// of another kind.
type keySet struct {
	keys []xkey
	// tabs are the panel's tabs (tabs.go).
	tabs []tabPlate
	// plain is set once the grid is folded to ASCII: its keys stay chips.
	plain bool
	// glyph reads the rune a cell of the grid holds.
	glyph func(x, y int) rune
}

// keyHolder is what holds a keySet: a grid, or the set itself.
type keyHolder interface{ keyset() *keySet }

func (s *keySet) keyset() *keySet { return s }

// keyset is the grid's buttons, reading its cells.
func (g *xgrid) keyset() *keySet {
	if g == nil {
		return nil
	}
	g.keySet.glyph = func(x, y int) rune { return g.c[y*g.w+x].r }
	return &g.keySet
}

// heldKeys is the keySet h holds, nil for none.
func heldKeys(h keyHolder) *keySet {
	if h == nil {
		return nil
	}
	return h.keyset()
}

// keyAt is the key the cell pt presses.
func (s *keySet) keyAt(pt image.Point) (xkey, bool) {
	for _, k := range s.keys {
		if pt.In(k.r) && !k.still {
			return k, true
		}
	}
	return xkey{}, false
}

// keysOff marks the keys drawn since the first len(s.keys) was from as
// unavailable.
func (s *keySet) keysOff(from int) {
	for i := from; i < len(s.keys); i++ {
		s.keys[i].off = true
	}
}

// capPart is one key named on a keycap, and the runes of the cap's label
// that are its own.
type capPart struct {
	key    kit.Key
	off, w int
}

// capParts splits a keycap's label into the keys it names: "Esc" is one key,
// "↑↓" two, "PgUp PgDn" two. A part that names no key ("1-9") is left out,
// and so is a note in brackets after the keys.
func capParts(label string) []capPart {
	var parts []capPart
	rs := []rune(label)
	start := 0
	word := func(end int) {
		if k, ok := capKey(string(rs[start:end])); ok {
			parts = append(parts, capPart{k, start, end - start})
		}
		start = end + 1
	}
	for i, r := range rs {
		if r == '(' {
			word(i)
			return parts
		}
		if r == ' ' || r == '/' {
			word(i)
		} else if k, ok := arrowKey(r); ok {
			word(i)
			parts = append(parts, capPart{k, i, 1})
		}
	}
	word(len(rs))
	return parts
}

// arrowKey is the key an arrow glyph stands for.
func arrowKey(r rune) (kit.Key, bool) {
	switch r {
	case '↑':
		return kit.Key{Code: kit.KeyUp}, true
	case '↓':
		return kit.Key{Code: kit.KeyDown}, true
	case '←':
		return kit.Key{Code: kit.KeyLeft}, true
	case '→':
		return kit.Key{Code: kit.KeyRight}, true
	}
	return kit.Key{}, false
}

// keyPut writes text at (x, y) into a grid's own cells; cap says the text is
// a key's cap and not its word.
type keyPut func(x, y int, text string, cap bool)

// keyButton draws a key and its word at (x, y) as a button, records what a
// click on it presses, and returns the next x. A cap that names one key is
// one button with its word. A cap that names several is a button a key, each
// with its own cells of the cap; the word is then nobody's. The cells hold the
// cap and the word as they read without pictures; paint restyles them.
func (g *xgrid) keyButton(x, y int, key, name string) int {
	return g.keyset().keyButton(x, y, key, name, func(x, y int, text string, cap bool) {
		k := xText
		if cap {
			k = xKeycap
		}
		g.put(x, y, text, k)
	})
}

// keyButton is xgrid.keyButton for any grid: put writes its cells.
func (s *keySet) keyButton(x, y int, key, name string, put keyPut) int {
	put(x, y, " "+key+" ", true)
	capEnd := x + len([]rune(key)) + 2
	put(capEnd, y, " "+name+" ", false)
	end := capEnd + len([]rune(name)) + 2
	cap, word := image.Rect(x, y, capEnd, y+1), image.Rect(capEnd, y, end, y+1)
	parts := capParts(key)
	kind := keySteel
	for _, p := range parts {
		if !moveKey(p.key) {
			kind = keyAccent
		}
	}
	if len(parts) == 0 {
		s.keys = append(s.keys, xkey{r: cap, cap: cap, word: word, still: true})
		return end
	}
	if len(parts) == 1 {
		s.keys = append(s.keys, xkey{key: parts[0].key, r: image.Rect(x, y, end, y+1), cap: cap, word: word, kind: kind})
		return end
	}
	for i, p := range parts {
		x0, x1 := x+1+p.off, x+1+p.off+p.w
		if i == 0 {
			x0 = x
		}
		if i == len(parts)-1 {
			x1 = capEnd
		}
		s.keys = append(s.keys, xkey{key: p.key, r: image.Rect(x0, y, x1, y+1), cap: cap, word: word, kind: kind})
	}
	return end
}

// moveKey reports whether k moves a selection or a view: the arrows, the
// paging keys, Home and End.
func moveKey(k kit.Key) bool {
	switch k.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight, kit.KeyPgUp, kit.KeyPgDn, kit.KeyHome, kit.KeyEnd:
		return true
	}
	return false
}

// escClose is the key that closes a panel, for a foot that ends with it.
func escClose() [2]string { return [2]string{"Esc", "close"} }

// keyButtonW is the cells keyButton takes.
func keyButtonW(key, name string) int { return len([]rune(key)) + len([]rune(name)) + 4 }

const (
	// closeButtonMinW is the narrowest box that gets a close button.
	closeButtonMinW = 24
	// closeButtonCols is the columns at the right of a box's top edge that
	// the close button takes, the corner and the edge beside it included:
	// anything else written in that edge ends before them.
	closeButtonCols = 5
)

// closeButton draws the close button in the top edge of a box w wide, at its
// right, and records that a click on it presses Esc.
func (g *xgrid) closeButton(w int) {
	g.keyset().closeButton(w, func(x, y int, text string, _ bool) { g.put(x, y, text, xKeycap) })
}

// closeButton is xgrid.closeButton for any grid: put writes its cells.
func (s *keySet) closeButton(w int, put keyPut) {
	if w < closeButtonMinW {
		return
	}
	x := w - closeButtonCols
	put(x, 0, " × ", true)
	// The whole slot takes the click; the key itself rests on its first cells.
	s.keys = append(s.keys, xkey{key: kit.Key{Code: kit.KeyEsc}, r: image.Rect(x, 0, x+closeButtonCols, 1), cap: image.Rect(x, 0, x+3, 1), kind: keyClose})
}

// keyPresser is what a panel keeps to make the keys it draws pressable: what
// presses a key, and where the pointer is.
type keyPresser struct {
	// press presses a key as the keyboard would: the key goes to whatever
	// has the keyboard. Nil in a panel drawn alone (a picture of the screen),
	// whose keys are then not buttons.
	press func(kit.Key)
	// run runs a command line as the command bar runs a typed one: a button
	// that carries a line (xkey.cmd) is clicked through it.
	run func(cmd string)
	// confirm asks the player a question and runs a command line if the
	// answer is yes: a button that carries a question (xkey.ask) is clicked
	// through it, and names the word on the danger button (xkey.yes).
	confirm func(question, yes, cmd string)
	// over is the cell the pointer is over, in the panel's own cells, while
	// in says it is over the panel.
	over image.Point
	in   bool
	// feel is what the keys typed and the buttons held look like: shared
	// with the app (wire), or the presser's own.
	feel *keyFeel
	// plates are the pictures of the keys drawn so far.
	plates plateSet
	// shows is the command of the button the pointer last came to rest on,
	// put in feel.preview once: a key typed takes it away until the pointer
	// has been to another button.
	shows string
}

// wire connects the presser to the app: a press is a key sent to it, a
// command is a line the command bar runs, and the keys typed on it show on
// the buttons.
func (k *keyPresser) wire(a *appLoop) {
	k.press, k.feel = a.Key, &a.feel
	k.run, k.confirm = a.runLine, a.askLine
}

// sense is the presser's feel, made when it has none.
func (k *keyPresser) sense() *keyFeel {
	if k.feel == nil {
		k.feel = &keyFeel{}
	}
	return k.feel
}

// lit is the key under the pointer.
func (k *keyPresser) lit(s *keySet) (xkey, bool) {
	if s == nil || k.press == nil || !k.in {
		return xkey{}, false
	}
	return s.keyAt(k.over)
}

// fire does what a click or the key of b does: runs its command line (after
// its question, if it has one), or presses its key. A command the command
// table marks Dangerous is never run without a question: a button that gives
// none is asked about with the line itself.
func (k *keyPresser) fire(b xkey) {
	if b.cmd != "" && b.ask == "" && lineDangerous(b.cmd) {
		b.ask, b.yes = dangerAsk(b.cmd), txtDangerYes.String()
	}
	switch {
	case b.cmd == "":
		k.press(b.key)
	case b.ask != "":
		if k.confirm != nil {
			k.confirm(b.ask, b.yes, b.cmd)
		}
	case k.run != nil:
		k.run(b.cmd)
	}
}

// txtDangerYes is the word on the danger button of the question a click on a
// Dangerous command asks without one of its own.
var txtDangerYes = rowid.E("ui.key_bar.dangerYes.01", "Run")

// dangerAsk is the question for a Dangerous command line a button gave none for.
func dangerAsk(line string) string {
	return rowid.Sprintf("ui.key_bar.dangerAsk.01", "Run \"%s\"?", line)
}

// lineDangerous reports whether line runs a command the command table marks
// Dangerous: the check the typed path makes before a second Enter.
func lineDangerous(line string) bool {
	p, ok := parse(registry(), strings.Fields(line))
	return ok && p.dangerous()
}

// preview keeps feel.preview as the command of the button under the pointer.
func (k *keyPresser) preview(b xkey, lit bool) {
	cmd := ""
	if lit && !b.off {
		cmd = b.cmd
	}
	if cmd == k.shows {
		return
	}
	k.shows = cmd
	k.sense().preview = cmd
}

// mouse takes m for the keys h drew, with the panel's top left at org. A left
// press on a key presses it. It reports whether the event was used: a key was
// pressed, or the key that is lit changed.
func (k *keyPresser) mouse(h keyHolder, org image.Point, m kit.Mouse) bool {
	s := heldKeys(h)
	was, wasLit := k.lit(s)
	wasTab := k.tabOver(s)
	if k.in = m.Kind != kit.MouseLeave; k.in {
		k.over = m.Pt().Sub(org)
	}
	now, isLit := k.lit(s)
	k.preview(now, isLit)
	if m.Kind == kit.MouseMove && isLit && now.off && now.why != "" {
		k.sense().tip.Aim(tipKey{k, now.r}, now.r.Add(org), kit.TipText{Reason: now.why}, false)
	}
	switch m.Kind {
	case kit.MousePress:
		// A press is a key's or it is not: one that only moved the light is
		// left for the panel to use. The key stays sunk until the button is up.
		// An unavailable one takes the press and does nothing.
		if isLit && m.Button == kit.ButtonLeft {
			if !now.off {
				k.fire(now)
				k.sense().hold(now.key, now.r)
			}
			return true
		}
		return false
	case kit.MouseRelease, kit.MouseLeave:
		if k.feel != nil && k.feel.release() {
			return true
		}
	case kit.MouseWheel:
		return false // the panel scrolls; the pointer's light follows on its own
	}
	return isLit != wasLit || now.r != was.r || k.tabOver(s) != wasTab
}

// pointer is the hand over a key a click would press.
func (k *keyPresser) pointer(h keyHolder, org, at image.Point) kit.Pointer {
	s := heldKeys(h)
	if s == nil || k.press == nil {
		return kit.PointerArrow
	}
	b, ok := s.keyAt(at.Sub(org))
	return handIf(ok && !b.off)
}
