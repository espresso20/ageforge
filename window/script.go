package window

import (
	"bytes"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
)

// script is a dev-only hook (AGEFORGE_WINDOW_SCRIPT=<file>) for checking the
// window without a person: one step per line, "wait N" frames, "key enter|esc|
// up|down|tab", "text abc", "shot <file.png>", "quit", and on the map's canvas (art pixels) "click x y", "focus x y", "wheel x y n", "drag x1 y1 x2 y2".
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

var scriptKeys = map[string]tcell.Key{"enter": tcell.KeyEnter, "esc": tcell.KeyEscape, "up": tcell.KeyUp, "down": tcell.KeyDown, "tab": tcell.KeyTab, "left": tcell.KeyLeft, "right": tcell.KeyRight, "home": tcell.KeyHome, "pgup": tcell.KeyPgUp, "pgdn": tcell.KeyPgDn}

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
		g.o.Host.InjectKey(scriptKeys[arg], 0, tcell.ModNone)
	case "text":
		for _, r := range arg {
			g.o.Host.InjectKey(tcell.KeyRune, r, tcell.ModNone)
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
			g.sky.wheel(v[0], v[1], float64(v[2]))
		case "drag":
			_, _, n := g.sky.size()
			g.sky.pan(float64((v[2]-v[0])*n), float64((v[3]-v[1])*n))
		}
	case "shot":
		s.shot = arg
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
