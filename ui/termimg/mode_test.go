package termimg

import "testing"

func env(kv ...string) func(string) string {
	m := map[string]string{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return func(k string) string { return m[k] }
}

func TestDecide(t *testing.T) {
	cases := []struct {
		name string
		env  []string
		p    Probe
		want Mode
	}{
		{"nothing", nil, Probe{}, None},
		{"iterm hint", []string{"TERM_PROGRAM", "iTerm.app"}, Probe{}, Sixel},
		{"wezterm hint", []string{"TERM_PROGRAM", "WezTerm"}, Probe{}, Sixel},
		{"windows terminal", []string{"WT_SESSION", "x"}, Probe{}, Sixel},
		{"foot", []string{"TERM", "foot"}, Probe{}, Sixel},
		{"kitty window", []string{"KITTY_WINDOW_ID", "1"}, Probe{}, Kitty},
		{"kitty term", []string{"TERM", "xterm-kitty"}, Probe{}, Kitty},
		{"ghostty", []string{"TERM_PROGRAM", "ghostty"}, Probe{}, Kitty},
		{"terminal.app", []string{"TERM_PROGRAM", "Apple_Terminal"}, Probe{}, None},
		{"tmux beats hints", []string{"TERM_PROGRAM", "iTerm.app", "TMUX", "/x"}, Probe{}, None},
		{"screen term", []string{"TERM", "screen-256color", "WT_SESSION", "x"}, Probe{}, None},
		{"probe sixel", nil, Probe{Answered: true, Sixel: true}, Sixel},
		{"probe kitty wins", nil, Probe{Answered: true, Sixel: true, Kitty: true}, Kitty},
		{"probe says no beats hint", []string{"TERM_PROGRAM", "iTerm.app"}, Probe{Answered: true}, None},
		{"off", []string{"AGEFORGE_GRAPHICS", "off", "TERM_PROGRAM", "iTerm.app"}, Probe{}, None},
		{"forced sixel", []string{"AGEFORGE_GRAPHICS", "sixel"}, Probe{}, Sixel},
		{"forced kitty in tmux", []string{"AGEFORGE_GRAPHICS", "KITTY", "TMUX", "/x"}, Probe{}, Kitty},
		{"auto", []string{"AGEFORGE_GRAPHICS", "auto", "WT_SESSION", "x"}, Probe{}, Sixel},
		{"garbage is auto", []string{"AGEFORGE_GRAPHICS", "banana"}, Probe{}, None},
	}
	for _, c := range cases {
		if got := Decide(env(c.env...), c.p); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

func TestParseReplies(t *testing.T) {
	p := ParseReplies([]byte("\x1b_Gi=31;OK\x1b\\\x1b[6;17;9t\x1b[?62;4;22c"))
	if !p.Answered || !p.Kitty || !p.Sixel || p.CellW != 9 || p.CellH != 17 {
		t.Errorf("%+v", p)
	}
	p = ParseReplies([]byte("\x1b[?1;2c"))
	if !p.Answered || p.Kitty || p.Sixel || p.CellW != 0 {
		t.Errorf("%+v", p)
	}
	if p = ParseReplies(nil); p.Answered {
		t.Errorf("%+v", p)
	}
}
