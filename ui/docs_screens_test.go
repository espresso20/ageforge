package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The wiki shows real views of the game's screens: a page names one with
//
//	<figure class="screen" data-screen="dashboard"><figcaption>...</figcaption></figure>
//
// and site/docs/screens.js draws site/docs/screens/dashboard.json there. The
// files are written by TestWriteSiteScreens (site_screens_test.go, build tag
// mapcapture). This test holds the two together: a figure with no file, a
// file no figure shows, a figure the plugin could not find, or a file it
// could not draw fails the build.

// docScreenFigure is a whole figure, as it must be written: on one line,
// from the line's start, with its caption inside.
var docScreenFigure = regexp.MustCompile(`^<figure class="screen" data-screen="([a-z0-9][a-z0-9-]*)"><figcaption>([^<]+)</figcaption></figure>$`)

// docScreensBudget is the most the screens folder may weigh.
const docScreensBudget = 400 << 10

func TestDocScreens(t *testing.T) {
	pages, err := filepath.Glob("../site/docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	shown := map[string][]string{} // screen name -> where it is shown
	for _, page := range pages {
		for i, line := range strings.Split(readDoc(t, page), "\n") {
			if !strings.Contains(line, "data-screen") && !strings.Contains(line, `class="screen"`) {
				continue
			}
			at := relDoc(page) + ":" + strconv.Itoa(i+1)
			m := docScreenFigure.FindStringSubmatch(line)
			if m == nil {
				t.Errorf("%s: not a screen figure the wiki can draw. Write it on one line, from the line's start: "+
					`<figure class="screen" data-screen="name"><figcaption>One sentence.</figcaption></figure>`, at)
				continue
			}
			if caption := strings.TrimSpace(m[2]); !strings.HasSuffix(caption, ".") {
				t.Errorf("%s: the caption of %q is not a sentence: %q", at, m[1], caption)
			}
			shown[m[1]] = append(shown[m[1]], at)
		}
	}
	if len(shown) == 0 {
		t.Fatal("no wiki page shows a screen")
	}

	files, err := filepath.Glob("../site/docs/screens/*.json")
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	total := 0
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".json")
		have[name] = true
		if len(shown[name]) == 0 {
			t.Errorf("%s: no wiki page shows this screen. Put a figure for it on the page that explains it, or drop it from TestWriteSiteScreens and run it again", relDoc(f))
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		total += len(raw)
		if msg := docScreenFault(raw); msg != "" {
			t.Errorf("%s: %s", relDoc(f), msg)
		}
	}
	names := make([]string, 0, len(shown))
	for name := range shown {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !have[name] {
			t.Errorf("%s: no site/docs/screens/%s.json. Add the screen to TestWriteSiteScreens (go test -tags mapcapture -run TestWriteSiteScreens ./ui)", strings.Join(shown[name], ", "), name)
		}
	}
	if total > docScreensBudget {
		t.Errorf("site/docs/screens weighs %d bytes, over its %d budget", total, docScreensBudget)
	}

	// The plugin adds itself to window.$docsify.plugins as it loads: after
	// the page has set window.$docsify (or the setting would wipe it) and
	// before Docsify reads it.
	index := readDoc(t, "../site/docs/index.html")
	config, plugin, docsify := strings.Index(index, "window.$docsify = {"), strings.Index(index, `<script src="screens.js">`), strings.Index(index, "docsify.min.js")
	if config < 0 || plugin < config || docsify < plugin {
		t.Error("site/docs/index.html must load screens.js after it sets window.$docsify and before docsify.min.js")
	}
	if _, err := os.Stat("../site/docs/screens.js"); err != nil {
		t.Errorf("site/docs/screens.js: %v", err)
	}
}

var docScreenColor = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// docScreenFault says what is wrong with a screen file, or "" when
// screens.js can draw it.
func docScreenFault(raw []byte) string {
	var s struct {
		W       int      `json:"w"`
		H       int      `json:"h"`
		Bg      string   `json:"bg"`
		Palette []string `json:"palette"`
		Styles  [][3]int `json:"styles"`
		Rows    []struct {
			T string `json:"t"`
			R []int  `json:"r"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return "not a screen file: " + err.Error()
	}
	if s.W <= 0 || s.H <= 0 || len(s.Rows) != s.H {
		return "its size and its rows disagree"
	}
	if s.Bg != "" && !docScreenColor.MatchString(s.Bg) {
		return "bg is not a color"
	}
	for _, c := range s.Palette {
		if !docScreenColor.MatchString(c) {
			return "the palette holds " + c + ", which is not a color"
		}
	}
	if len(s.Styles) == 0 {
		return "it has no styles"
	}
	for _, st := range s.Styles {
		if st[0] < 0 || st[0] >= len(s.Palette) || st[1] < 0 || st[1] >= len(s.Palette) || st[2] < 0 || st[2] > 7 {
			return "a style points outside the palette"
		}
	}
	for y, row := range s.Rows {
		cells := len([]rune(row.T))
		if cells > s.W || len(row.R)%2 != 0 {
			return "row " + strconv.Itoa(y) + " is wider than the screen or has half a run"
		}
		sum := 0
		for i := 0; i < len(row.R); i += 2 {
			if row.R[i] < 0 || row.R[i] >= len(s.Styles) || row.R[i+1] <= 0 {
				return "row " + strconv.Itoa(y) + " has a run with no style or no length"
			}
			sum += row.R[i+1]
		}
		if sum != cells {
			return "row " + strconv.Itoa(y) + ": its runs do not cover its text"
		}
	}
	return ""
}
