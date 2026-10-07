package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// TestMiniMapAppearsWhereTheDocsSay: the mini map shows from one terminal
// size up, the game works that size out from its layout, and the wiki and
// the minimap command quote it. A change to the layout that moves the size
// fails here until the docs say the new one.
func TestMiniMapAppearsWhereTheDocsSay(t *testing.T) {
	restoreForge(t)
	shown := func(w, h int) bool {
		d := NewDashboard(tview.NewApplication(), game.NewGameEngine(), tview.NewPages())
		pages := tview.NewPages()
		pages.AddPage("dashboard", d.Root(), true, true)
		drawnDashboard(t, d, pages, w, h)
		return d.mapDock.shown
	}
	cols, rows := miniMapMinCols, miniMapMinRows
	if !shown(cols, rows) {
		t.Errorf("no mini map at %dx%d, the size the game says it shows from", cols, rows)
	}
	if shown(cols-1, rows) || shown(cols, rows-1) {
		t.Errorf("the mini map shows on a terminal smaller than %dx%d", cols, rows)
	}
	phrase := fmt.Sprintf("%d columns by %d rows", cols, rows)
	for _, page := range []string{"map.md", "commands.md", "first-ten-minutes.md", "how-to-play.md"} {
		if !strings.Contains(readDoc(t, "../site/docs/"+page), phrase) {
			t.Errorf("site/docs/%s does not say the mini map shows from %s", page, phrase)
		}
	}
	if reply := HandleCommand("minimap on", game.NewGameEngine()).Message; !strings.Contains(reply, fmt.Sprintf("%dx%d", cols, rows)) {
		t.Errorf("minimap on does not name %dx%d: %q", cols, rows, reply)
	}
}

// TestWikiSidebarKeepsSearchOnTheWiki: Docsify's search reads every link in
// the sidebar as a wiki page to index. A link out of the wiki (back to the
// site) made it ask for /README.md on every load and get a 404, so such a
// link is written as an anchor marked data-nosearch, which search skips.
func TestWikiSidebarKeepsSearchOnTheWiki(t *testing.T) {
	sidebar := readDoc(t, "../site/docs/_sidebar.md")
	if strings.Contains(sidebar, "':ignore") {
		t.Error("site/docs/_sidebar.md has a link marked :ignore; search would fetch it as a page. Write it as <a href=\"...\" data-nosearch>")
	}
	link := regexp.MustCompile(`\]\(([^) ]+)`)
	for _, m := range link.FindAllStringSubmatch(sidebar, -1) {
		if target := m[1]; target != "/" && !strings.HasSuffix(target, ".md") {
			t.Errorf("site/docs/_sidebar.md links to %q, which is not a wiki page", target)
		}
	}
	if !strings.Contains(sidebar, `<a href="/" target="_self" data-nosearch>`) {
		t.Error("site/docs/_sidebar.md has no link back to the site that search skips")
	}
}

// TestWikiPhoneColumnHasOneGutter: on a phone the wiki's text column keeps
// one 16px gutter a side. The theme pads the column 45px at every width,
// and with the phone rule's 16px on top the text was 253px of a 375px
// screen.
func TestWikiPhoneColumnHasOneGutter(t *testing.T) {
	css := readDoc(t, "../site/docs/custom.css")
	at := strings.LastIndex(css, "@media (max-width: 768px)")
	if at < 0 {
		t.Fatal("site/docs/custom.css has no phone rules")
	}
	phone := css[at:]
	for _, want := range []string{
		".content { padding: 16px !important; }",
		".markdown-section { padding-left: 0; padding-right: 0; }",
		".markdown-section .screen-pair { margin-left: 0; margin-right: 0; }",
	} {
		if !strings.Contains(phone, want) {
			t.Errorf("the phone rules of site/docs/custom.css lack %q", want)
		}
	}
}
