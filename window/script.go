package window

import (
	"bytes"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/espresso20/ageforge/window/kit"
)

// script is a dev-only hook (AGEFORGE_WINDOW_SCRIPT=<file>) for checking the
// window without a person: one step per line, "wait N" frames, "key enter|esc|
// up|down|tab", "text abc", "shot <file.png>", "quit", and on the map's canvas (art pixels) "click x y", "focus x y", "wheel x y n", "drag x1 y1 x2 y2".
// For the game's own screens, at a cell of the grid: "mouse move|press|release x y" and "mouse wheel x y rows";
// "pointer <file>" writes the shape the mouse pointer has there.
type script struct {
	steps []string
	i     int
	wait  int
	shot  string
}

func loadScript(path string) *script {
	b, _ := os.ReadFile(path)
	return &script{steps: strings.Split(strings.TrimSpace(string(b)), "\n")}
}

var scriptKeys = map[string]kit.KeyCode{"enter": kit.KeyEnter, "esc": kit.KeyEsc, "up": kit.KeyUp, "down": kit.KeyDown, "tab": kit.KeyTab, "left": kit.KeyLeft, "right": kit.KeyRight, "home": kit.KeyHome, "pgup": kit.KeyPgUp, "pgdn": kit.KeyPgDn}

var scriptMouse = map[string]kit.MouseKind{"move": kit.MouseMove, "press": kit.MousePress, "release": kit.MouseRelease, "wheel": kit.MouseWheel}

// step runs the next line; it reports true on "quit".
func (s *script) step(g *Game) bool {
	if s.wait > 0 || s.shot != "" {
		s.wait--
		return false
	}
	if s.i >= len(s.steps) {
		return true
	}
	cmd, arg, _ := strings.Cut(strings.TrimSpace(s.steps[s.i]), " ")
	s.i++
	switch cmd {
	case "wait":
		s.wait, _ = strconv.Atoi(arg)
	case "key":
		g.o.Screen.Key(kit.Key{Code: scriptKeys[arg]})
	case "text":
		for _, r := range arg {
			g.o.Screen.Key(kit.Key{Code: kit.KeyRune, Rune: r})
		}
	case "click", "focus", "wheel", "drag": // in art pixels of the map's canvas
		var v [4]int
		for i, f := range strings.Fields(arg) {
			if i < len(v) {
				v[i], _ = strconv.Atoi(f)
			}
		}
		switch cmd {
		case "click":
			g.sky.click(v[0], v[1])
		case "focus":
			g.sky.focus(v[0], v[1])
		case "wheel":
			f, _ := strconv.ParseFloat(strings.Fields(arg + " 0 0 0")[2], 64) // steps may be fractions
			g.sky.wheel(v[0], v[1], f)
		case "drag":
			_, _, n := g.sky.size()
			g.sky.pan(float64((v[2]-v[0])*n), float64((v[3]-v[1])*n))
		}
	case "mouse":
		f := strings.Fields(arg + " 0 0 0")
		x, _ := strconv.Atoi(f[1])
		y, _ := strconv.Atoi(f[2])
		rows, _ := strconv.Atoi(f[3])
		if kind, ok := scriptMouse[f[0]]; ok {
			g.o.Screen.Mouse(kit.Mouse{Kind: kind, X: x, Y: y, Wheel: rows})
		}
	case "shot":
		s.shot = arg
	case "pointer": // the pointer's shape, as the digit of its kit.Pointer
		_ = os.WriteFile(arg, []byte{'0' + byte(g.pointing)}, 0o644)
	case "quit":
		return true
	}
	return false
}

// capture writes the frame to the pending "shot" file.
func (s *script) capture(screen *ebiten.Image) {
	if s.shot == "" {
		return
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, screen)
	_ = os.WriteFile(s.shot, buf.Bytes(), 0o644)
	s.shot = ""
}
