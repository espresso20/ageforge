// Package termimg draws pictures in terminals that can show them: it works out
// which graphics protocol the terminal speaks (sixel or the kitty graphics
// protocol), encodes an image for it, and keeps one picture on top of the
// text screen a tview application draws.
package termimg

import (
	"strconv"
	"strings"
)

// Mode is how pictures reach the terminal.
type Mode uint8

const (
	None  Mode = iota // text only: the caller draws with blocks
	Sixel             // iTerm2, WezTerm, Windows Terminal, xterm, foot
	Kitty             // Kitty, Ghostty
)

func (m Mode) String() string { return [...]string{"none", "sixel", "kitty"}[m] }

// Probe is what the terminal said when asked at startup. Answered is false
// when it did not answer (or could not be asked).
type Probe struct {
	Answered     bool
	Sixel, Kitty bool
	CellW, CellH int // pixels per cell from XTWINOPS 16, 0 if unknown
}

// Decide picks the mode from the environment and the probe. The override
// AGEFORGE_GRAPHICS (off, sixel, kitty, auto) wins; a multiplexer means
// text; a terminal that answered the probe is believed over the hints; with no
// answer, the hints stand in; anything else is text.
func Decide(getenv func(string) string, p Probe) Mode {
	switch strings.ToLower(strings.TrimSpace(getenv("AGEFORGE_GRAPHICS"))) {
	case "off", "0", "false", "none", "no":
		return None
	case "sixel":
		return Sixel
	case "kitty":
		return Kitty
	}
	term := getenv("TERM")
	if term == "dumb" || getenv("TMUX") != "" || getenv("STY") != "" || strings.HasPrefix(term, "screen") || strings.HasPrefix(term, "tmux") {
		return None
	}
	if p.Answered {
		switch {
		case p.Kitty:
			return Kitty
		case p.Sixel:
			return Sixel
		}
		return None
	}
	prog := getenv("TERM_PROGRAM")
	switch {
	case getenv("KITTY_WINDOW_ID") != "", term == "xterm-kitty", term == "xterm-ghostty", prog == "ghostty", getenv("GHOSTTY_RESOURCES_DIR") != "":
		return Kitty
	case getenv("WT_SESSION") != "", prog == "iTerm.app", prog == "WezTerm", strings.Contains(term, "foot"), strings.Contains(term, "mlterm"):
		return Sixel
	}
	return None
}

// ParseReplies reads what a terminal sent back to the probe's queries: the
// kitty graphics reply, the cell size and the primary device attributes (the
// last, which every terminal answers, so Answered).
func ParseReplies(b []byte) Probe {
	s := string(b)
	var p Probe
	if i := strings.Index(s, "\x1b_Gi=31;"); i >= 0 && strings.HasPrefix(s[i+len("\x1b_Gi=31;"):], "OK") {
		p.Kitty = true
	}
	if i := strings.Index(s, "\x1b[6;"); i >= 0 {
		if j := strings.IndexByte(s[i:], 't'); j > 0 {
			f := strings.Split(s[i+4:i+j], ";")
			if len(f) == 2 {
				h, _ := strconv.Atoi(f[0])
				w, _ := strconv.Atoi(f[1])
				if h > 0 && w > 0 {
					p.CellH, p.CellW = h, w
				}
			}
		}
	}
	if i := strings.Index(s, "\x1b[?"); i >= 0 {
		if j := strings.IndexByte(s[i:], 'c'); j > 0 {
			p.Answered = true
			for _, f := range strings.Split(s[i+3:i+j], ";") {
				if f == "4" {
					p.Sixel = true
				}
			}
		}
	}
	return p
}

// probeQueries are sent in this order: the kitty graphics query, the cell size
// and last the primary device attributes, so that its reply says all the
// others have been answered or never will be.
const probeQueries = "\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[16t\x1b[c"
