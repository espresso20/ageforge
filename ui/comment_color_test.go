package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// The comment color (card KkocMbpB) is the flavor sentences' alone. The
// per-theme rules (set, used by no other role, contrast) are in the theme
// package's tests; these hold the other side: where the color is drawn.

// commentSites is the inventory: every source file that draws the comment
// color, and how many places in it do (a "[comment" tag, a RoleComment, a
// commentInk call). Each place is a flavor sentence: a building's line, an
// age's quip, a milestone's or a quest chain's, the log's quips (building,
// starvation, research, expedition, age and catastrophe), a harbinger's lines
// and verdict flavor, the Last Passage's, a random or epoch event's sentence,
// a raid's or an encounter's, the catastrophe and Last Passage windows, the
// epoch overlay and the full-screen age arrival and ending. A new place, or a
// removed one, is a change to this table, so the color cannot drift onto
// anything else unseen.
var commentSites = map[string]int{
	"game/catastrophe.go":       2,
	"game/encounters.go":        2,
	"game/engine.go":            11,
	"game/event_size.go":        1,
	"game/expedition_flavor.go": 1,
	"game/harbinger.go":         3,
	"game/last_passage.go":      1,
	"ui/age_splash.go":          2,
	"ui/arrival_page.go":        3,
	"ui/catastrophe_modal.go":   2,
	"ui/ending_page.go":         1,
	"ui/mapstyle/cursor.go":     1, // skips the role when it asks "are all the roles grey"
	"ui/overlay_buildings.go":   1,
	"ui/overlay_epoch.go":       1,
	"ui/overlay_harbinger.go":   1,
	"ui/tab_economy.go":         1,
}

func TestCommentColorIsDrawnOnlyOnFlavor(t *testing.T) {
	got := map[string]int{}
	for _, dir := range []string{"game", "ui", "ui/mapstyle", "config", "flavor", "cmd", "smoke"} {
		files, _ := filepath.Glob(filepath.Join("..", dir, "*.go"))
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			s := string(b)
			n := strings.Count(s, "[comment") + strings.Count(s, "RoleComment") + strings.Count(s, "commentInk(")
			if n > 0 {
				got[filepath.ToSlash(strings.TrimPrefix(f, "../"))] = n
			}
			// The old way of drawing a quip, a dim tag around the sentence.
			if dir == "game" && strings.Contains(s, `"  [gray]%s[-]"`) {
				t.Errorf("%s still draws a flavor line in gray", f)
			}
		}
	}
	for f, want := range commentSites {
		if got[f] != want {
			t.Errorf("%s draws the comment color in %d places, the inventory says %d", f, got[f], want)
		}
	}
	for f, n := range got {
		if _, ok := commentSites[f]; !ok {
			t.Errorf("%s draws the comment color (%d places) and is not in the inventory", f, n)
		}
	}
}

// TestAgeQuipIsTheOnlyCommentOnTheSplash: the age splash sets the quip in the
// comment color and nothing else on it.
func TestAgeQuipIsTheOnlyCommentOnTheSplash(t *testing.T) {
	seen := 0
	for age := range config.AgeByKey() {
		def := config.AgeByKey()[age]
		if def.Quip == "" {
			continue
		}
		for _, l := range ageSplashLines(age, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{}) {
			has := strings.Contains(l.text, "[comment")
			if (l.kind == skQuip) != has {
				t.Errorf("%s: line kind %v has comment tag = %v: %q", age, l.kind, has, l.text)
			}
			if l.kind == skQuip {
				seen++
			}
		}
	}
	if seen == 0 {
		t.Fatal("no age has a quip on its splash")
	}
}

// TestCommentInkReadsOnThePage: the full-screen scenes set a flavor line in
// the theme's comment color, kept legible on the page.
func TestCommentInkReadsOnThePage(t *testing.T) {
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		pal := &menuPalette{bg: theme.Color(theme.RoleBackground)}
		if r := theme.ContrastRatio(commentInk(pal), pal.bg); r < 4.5 {
			t.Errorf("%s: comment ink is %.2f:1 on the page", th.Key, r)
		}
	}
	_ = theme.SetActive(theme.DefaultKey)
}
