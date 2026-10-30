package ui

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// epoch_card.go is the Discord card: a picture of the epoch screen's front
// page, drawn offscreen at a fixed size so that every player's picture looks
// alike, in the player's theme, the Unicode glyph tier and no motion. `epoch
// card` and the c key save it next to the saves and put it on the clipboard.

// The screen the picture is drawn on: epochPicCols wide, the width at which
// the front page keeps its large banner and its tiles sit well, and as tall
// as the page is (epochPicMinRows at least). The page is laid out as for a
// screen of epochLayoutRows rows, which gives it the roomy banner and the
// plain gaps at any height, so two players with the same facts get the same
// picture.
const (
	epochPicCols    = 110
	epochLayoutRows = 48
	epochPicMinRows = 34
)

// epochCardFile is the card's file name, in the account's data directory.
const epochCardFile = "epoch-card.png"

// epochPicture draws the front page of the epoch screen for a state.
func epochPicture(state game.GameState) *image.RGBA {
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		return nil
	}
	defer sim.Fini()
	// The page says how tall it is; the frame adds its edges (3 rows) and a
	// blank row under the text.
	page := hallFrontPage(state, epochPicCols, epochLayoutRows, mapmodel.TierUnicode, hallMotion{}, false)
	picRows := max(epochPicMinRows, strings.Count(page, "\n")+1+panelChromeRows+1)
	sim.SetSize(epochPicCols, picRows)
	screen := theme.WrapScreen(sim)

	tv := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	tv.SetBackgroundColor(theme.Color(theme.RoleBackground))
	// The page has no key hint here: the picture is of the screen, and the
	// hint is for the player who is looking at it.
	tv.SetText(safeTags(page))
	frame := newPanelFrame(vibe.E("ui.epoch_card.epochPicture.01", "Epoch"), tv)
	frame.SetRect(0, 0, epochPicCols, picRows)
	screen.Clear()
	frame.Draw(screen)
	screen.Show()

	cells, cols, rows := sim.GetContents()
	return ScreenImage(cells, cols, rows, shotCellW, shotCellH)
}

// epochCardCopier is what the `epoch card` command puts the picture on the
// clipboard with. The screen's c key is given its own (dashboard.go); this is
// a variable only so that tests of the command never touch the real clipboard.
var epochCardCopier = copyImageToClipboard

// epochCardResult is what making the card did: the file it was saved to
// and whether the clipboard took it too.
type epochCardResult struct {
	Path   string
	Copied bool
}

// makeEpochCard saves the picture for a state to the data folder and then
// puts it on the clipboard with copy, which is given the file's path (nil
// skips the clipboard). A clipboard that fails is not an error: the file is
// the fallback, and Copied says false.
func makeEpochCard(state game.GameState, copy func(path string) error) (epochCardResult, error) {
	dir := game.DataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return epochCardResult{}, errors.New(vibe.Line("ui.epoch_card.cmdEpoch.02", "Could not create the data folder: ") + shortIOError(err) + ".")
	}
	img := epochPicture(state)
	var buf bytes.Buffer
	if img == nil || png.Encode(&buf, img) != nil {
		return epochCardResult{}, errors.New(vibe.Line("ui.epoch_card.makeEpochCard.01", "Could not draw the picture."))
	}
	path := filepath.Join(dir, epochCardFile)
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return epochCardResult{}, errors.New(vibe.Line("ui.epoch_card.cmdEpoch.03", "Could not save the picture: ") + shortIOError(err) + ".")
	}
	res := epochCardResult{Path: path}
	if copy != nil && copy(path) == nil {
		res.Copied = true
	}
	return res, nil
}
