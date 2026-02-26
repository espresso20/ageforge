package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// CreateSplashPage creates the main menu splash screen.
// onWiki is called when the player selects Wiki (opens dashboard wiki tab).
func CreateSplashPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, onWiki func()) tview.Primitive {
	saveExists := game.SaveExists("autosave")
	prestigeLevel := engine.Prestige.GetLevel()

	// ── Left panel: pixel art Greek temple ──────────────────────────────────
	imgView := tview.NewImage()
	imgView.SetColors(tview.TrueColor)
	imgView.SetBorder(true).
		SetTitle(" AgeForge ").
		SetTitleColor(tcell.ColorGold)
	imgView.SetImage(renderSplashTemple(600, 360))

	// ── Right panel ─────────────────────────────────────────────────────────

	// Title / tagline
	tagline := "[white]" + SplashTagline + "[-]"
	if prestigeLevel > 0 {
		tagline += fmt.Sprintf("\n[cyan]  Prestige Level %d[-]", prestigeLevel)
	}
	titleTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[gold]%s[-]\n[white]%s[-]", SplashArt, tagline))

	// Primary action list (Load → New → Wiki)
	mainList := tview.NewList()
	mainList.SetBorder(false)
	mainList.SetSelectedBackgroundColor(tcell.ColorGold)
	mainList.SetSelectedTextColor(tcell.ColorBlack)
	mainList.ShowSecondaryText(false)

	loadLabel := "  Load Game"
	if !saveExists {
		loadLabel = "  Load Game  (no save)"
	}
	mainList.AddItem(loadLabel, "", 'l', func() {
		if saveExists {
			if err := engine.LoadGame("autosave"); err != nil {
				engine.AddLog("error", fmt.Sprintf("Load failed: %v", err))
			} else {
				engine.AddLog("success", "Game loaded!")
			}
		}
		pages.SwitchToPage("dashboard")
		go engine.Start()
	})
	mainList.AddItem("  New Game", "", 'n', func() {
		pages.SwitchToPage("dashboard")
		go engine.Start()
	})
	mainList.AddItem("  Wiki", "", 'w', func() {
		if onWiki != nil {
			onWiki()
		}
	})

	// Visual separator
	sepTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]──────────────────[-]")

	// Danger action list (Wipe → Quit)
	dangerList := tview.NewList()
	dangerList.SetBorder(false)
	dangerList.SetSelectedBackgroundColor(tcell.ColorDarkRed)
	dangerList.SetSelectedTextColor(tcell.ColorWhite)
	dangerList.ShowSecondaryText(false)

	dangerList.AddItem("  Wipe Save", "", 'x', func() {
		showWipeConfirmation(app, pages, engine, onWiki)
	})
	dangerList.AddItem("  Quit", "", 'q', func() {
		app.Stop()
	})

	// Hint footer
	footerTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]Arrow keys / Enter   Tab to switch panel[-]")

	// Right panel flex (row)
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titleTV, 10, 0, false).
		AddItem(mainList, 5, 0, true).
		AddItem(sepTV, 1, 0, false).
		AddItem(dangerList, 4, 0, false).
		AddItem(nil, 0, 1, false).
		AddItem(footerTV, 1, 0, false)
	rightPanel.SetBorder(true).
		SetTitle(" Menu ").
		SetTitleColor(tcell.ColorGold)

	// Default selection: Load if save exists, else New Game
	if !saveExists {
		mainList.SetCurrentItem(1)
	}

	// Outer layout: image (3 parts) | menu (2 parts)
	outer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(imgView, 0, 3, false).
		AddItem(rightPanel, 0, 2, true)

	// Tab cycles between the two lists; arrow keys are handled by the lists
	outer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if mainList.HasFocus() {
				app.SetFocus(dangerList)
			} else {
				app.SetFocus(mainList)
			}
			return nil
		case tcell.KeyBacktab:
			if dangerList.HasFocus() {
				app.SetFocus(mainList)
			} else {
				app.SetFocus(dangerList)
			}
			return nil
		}
		return event
	})

	return outer
}

// showWipeConfirmation shows the "are you sure?" modal before wiping data.
func showWipeConfirmation(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, onWiki func()) {
	modal := tview.NewModal().
		SetText("⚠  WIPE ALL DATA  ⚠\n\nThis will permanently delete ALL save files\nand reset the game to zero.\n\nPrestige, upgrades, progress — everything gone.\n\nAre you sure?").
		AddButtons([]string{"Cancel", "WIPE EVERYTHING"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage("wipe_confirm")
			if buttonLabel == "WIPE EVERYTHING" {
				game.WipeAllSaves()
				engine.Reset()
				pages.RemovePage("splash")
				newSplash := CreateSplashPage(app, pages, engine, onWiki)
				pages.AddPage("splash", newSplash, true, true)
			}
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	pages.AddPage("wipe_confirm", modal, true, true)
}

// ────────────────────────────────────────────────────────────────────────────
// Pixel art: Civ-1 style Greek temple (600 × 360 pixels)
// ────────────────────────────────────────────────────────────────────────────

func renderSplashTemple(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// ── Palette ─────────────────────────────────────────────────────────────
	// Sky: deep navy → rich cornflower → bright horizon
	skyTop := color.RGBA{12, 38, 105, 255}
	skyMid := color.RGBA{55, 118, 190, 255}
	skyBot := color.RGBA{120, 178, 225, 255}
	// Sun
	sunDisk := color.RGBA{255, 235, 65, 255}
	sunGlow := color.RGBA{255, 195, 70, 200}
	// Clouds
	cloudW := color.RGBA{242, 246, 253, 255}
	cloudSh := color.RGBA{182, 205, 228, 255}
	// Ground
	grassH := color.RGBA{55, 102, 40, 255} // bright grass
	grassL := color.RGBA{43, 82, 30, 255}  // dark grass
	dirt := color.RGBA{128, 98, 55, 255}   // path/dirt
	dirtAlt := color.RGBA{112, 86, 46, 255}

	// Warm Pentelic marble (ivory/buff, slightly golden — like the Parthenon)
	mHigh := color.RGBA{252, 248, 232, 255} // cream highlight
	mMain := color.RGBA{230, 220, 194, 255} // warm buff
	mMid := color.RGBA{195, 184, 158, 255}  // warm mid-tone
	mShad := color.RGBA{155, 144, 120, 255} // warm shadow
	mDark := color.RGBA{108, 100, 80, 255}  // deep shadow

	// Frieze decoration — ancient Greek temples had painted friezes!
	trigBlue := color.RGBA{42, 58, 148, 255}  // triglyph blue
	trigLite := color.RGBA{90, 110, 195, 255} // triglyph groove highlight
	metOchre := color.RGBA{178, 88, 38, 255}  // metope terracotta

	gY := h * 76 / 100 // ground line

	// ── Sky: 3-stop gradient ─────────────────────────────────────────────────
	for y := 0; y < gY; y++ {
		t := float64(y) / float64(gY)
		var c color.RGBA
		if t < 0.45 {
			c = lerp(skyTop, skyMid, t/0.45)
		} else {
			c = lerp(skyMid, skyBot, (t-0.45)/0.55)
		}
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}

	// ── Ground ───────────────────────────────────────────────────────────────
	for y := gY; y < h; y++ {
		for x := 0; x < w; x++ {
			depth := float64(y-gY) / float64(h-gY)
			n := noise2D(x, y, 77)
			if depth < 0.28 {
				if n > 0.5 {
					img.SetRGBA(x, y, grassL)
				} else {
					img.SetRGBA(x, y, grassH)
				}
			} else {
				if n > 0.55 {
					img.SetRGBA(x, y, dirtAlt)
				} else {
					img.SetRGBA(x, y, dirt)
				}
			}
		}
	}

	// Stone path leading up to temple stairs
	pathW := w * 16 / 100
	pathX := w/2 - pathW/2
	for y := gY; y < h; y++ {
		for x := pathX; x < pathX+pathW; x++ {
			if x >= 0 && x < w {
				n := noise2D(x, y, 133)
				if n > 0.5 {
					img.SetRGBA(x, y, mMid)
				} else {
					img.SetRGBA(x, y, mShad)
				}
			}
		}
	}

	// ── Sun: large disk with warm halo ───────────────────────────────────────
	sunX := w * 82 / 100
	sunY := h * 11 / 100
	sunR := max(w/18, 12)
	for dy := -sunR * 5; dy <= sunR*5; dy++ {
		for dx := -sunR * 5; dx <= sunR*5; dx++ {
			px, py := sunX+dx, sunY+dy
			if px < 0 || px >= w || py < 0 || py >= gY {
				continue
			}
			d2 := dx*dx + dy*dy
			r2 := sunR * sunR
			if d2 <= r2 {
				img.SetRGBA(px, py, sunDisk)
			} else if d2 <= r2*25 {
				t := 1.0 - float64(d2-r2)/float64(r2*24)
				c := color.RGBA{sunGlow.R, sunGlow.G, sunGlow.B, uint8(float64(sunGlow.A) * t * 0.6)}
				bg := img.RGBAAt(px, py)
				fa := float64(c.A) / 255.0
				img.SetRGBA(px, py, color.RGBA{
					uint8(float64(c.R)*fa + float64(bg.R)*(1-fa)),
					uint8(float64(c.G)*fa + float64(bg.G)*(1-fa)),
					uint8(float64(c.B)*fa + float64(bg.B)*(1-fa)),
					255,
				})
			}
		}
	}

	// ── Clouds ───────────────────────────────────────────────────────────────
	splashCloud(img, w*14/100, h*20/100, w*12/100, h*6/100, cloudW, cloudSh, gY)
	splashCloud(img, w*42/100, h*13/100, w*8/100, h*4/100, cloudW, cloudSh, gY)
	splashCloud(img, w*63/100, h*24/100, w*7/100, h*3/100, cloudW, cloudSh, gY)

	// ── Temple ───────────────────────────────────────────────────────────────
	templeW := w * 66 / 100
	templeX := (w - templeW) / 2

	// Stairs — 3 steps (crepidoma), each wider toward the ground
	numSteps := 3
	stepH := max(h*3/100, 4)
	stepPad := w * 25 / 1000 // lateral expansion per step

	for i := 0; i < numSteps; i++ {
		sy := gY - (numSteps-i)*stepH
		sx := templeX - i*stepPad
		sw := templeW + i*stepPad*2
		// Each step a shade darker going down (they're in more shadow)
		sc := lerp(mMain, mMid, float64(numSteps-1-i)/float64(numSteps)*0.6)
		splashFill(img, sx, sy, sx+sw, sy+stepH, sc)
		splashFill(img, sx, sy, sx+sw, sy+3, mHigh)              // lit top face
		splashFill(img, sx+sw-4, sy+3, sx+sw, sy+stepH, mShad)   // right-side shadow
		splashFill(img, sx, sy+stepH-2, sx+sw-4, sy+stepH, mMid) // bottom edge
	}

	colFloorY := gY - numSteps*stepH

	// Columns — 6 Doric with fluting and capitals
	numCols := 6
	colHW := max(templeW/42, 6) // half-width
	colH := h * 40 / 100
	colTopY := colFloorY - colH
	colStep := (templeW - 2*colHW) / (numCols - 1)

	for i := 0; i < numCols; i++ {
		cx := templeX + colHW + i*colStep

		// Shaft body (warm buff)
		splashFill(img, cx-colHW, colTopY, cx+colHW+1, colFloorY, mMain)

		// Fluting — 3 highlight/shadow pairs to suggest curved grooves
		splashFill(img, cx-colHW, colTopY, cx-colHW+3, colFloorY, mHigh)         // far-left flute highlight
		splashFill(img, cx-colHW/2-1, colTopY, cx-colHW/2+1, colFloorY, mMid)    // left groove
		splashFill(img, cx-1, colTopY, cx+2, colFloorY, lerp(mMain, mHigh, 0.6)) // center ridge
		splashFill(img, cx+colHW/2-1, colTopY, cx+colHW/2+1, colFloorY, mMid)    // right groove
		splashFill(img, cx+colHW-2, colTopY, cx+colHW+1, colFloorY, mShad)       // far-right shadow

		// Capital: echinus (bulging ring) — expands outward from shaft top
		capH := max(colH/11, 5)
		for y := colTopY; y < colTopY+capH; y++ {
			tCap := float64(y-colTopY) / float64(capH)
			extra := int(float64(colHW) * tCap * 0.55)
			splashFill(img, cx-colHW-extra, y, cx+colHW+extra+1, y+1, lerp(mHigh, mMid, tCap*0.4))
		}
		// Abacus (flat rectangular slab atop echinus)
		abacY := colTopY + capH
		abacH := max(capH/2, 3)
		abacW := colHW + colHW/2 + 3
		splashFill(img, cx-abacW, abacY, cx+abacW+1, abacY+abacH, mMid)
		splashFill(img, cx-abacW, abacY, cx+abacW+1, abacY+2, mHigh)
		splashFill(img, cx-abacW, abacY+abacH-2, cx+abacW+1, abacY+abacH, mShad)

		// Column base (simple torus at floor)
		splashFill(img, cx-colHW-2, colFloorY-5, cx+colHW+3, colFloorY, mHigh)
		splashFill(img, cx-colHW-2, colFloorY-2, cx+colHW+3, colFloorY, mMid)
	}

	// ── Entablature (architrave + frieze + cornice) ──────────────────────────
	entH := max(h*11/100, 12)
	entY := colTopY - entH
	entX1 := templeX - 5
	entX2 := templeX + templeW + 5

	// Architrave: bottom third, plain warm stone
	archH := entH * 3 / 10
	splashFill(img, entX1, entY+entH-archH, entX2, colTopY, mMain)
	splashFill(img, entX1, entY+entH-archH, entX2, entY+entH-archH+3, mHigh) // lit top
	splashFill(img, entX1, colTopY-3, entX2, colTopY, mShad)                 // shadow under

	// Frieze: middle third — ochre metopes with blue triglyphs
	frY1 := entY + entH*2/10
	frY2 := entY + entH*7/10
	splashFill(img, entX1, frY1, entX2, frY2, metOchre) // terracotta background

	trigSpacing := templeW / (numCols * 2)
	trigW := max(trigSpacing*3/5, 6)
	for i := 0; i < numCols*2+1; i++ {
		if i%2 == 0 {
			tx := templeX + i*trigSpacing - trigW/2
			splashFill(img, tx, frY1, tx+trigW, frY2, trigBlue)
			// Two vertical lighter grooves on each triglyph
			gw := max(trigW/5, 1)
			splashFill(img, tx+gw, frY1, tx+gw+gw/2, frY2, trigLite)
			splashFill(img, tx+trigW-gw-gw/2, frY1, tx+trigW-gw, frY2, trigLite)
		}
	}

	// Cornice: top third, overhangs slightly
	cornH := entH * 3 / 10
	cornX1 := entX1 - 5
	cornX2 := entX2 + 5
	splashFill(img, cornX1, entY, cornX2, entY+cornH, mMain)
	splashFill(img, cornX1, entY, cornX2, entY+3, mHigh)
	splashFill(img, cornX1, entY+cornH-3, cornX2, entY+cornH, mShad)
	// Drop shadow from cornice onto columns
	splashFill(img, entX1, entY+cornH, entX2, entY+cornH+5, lerp(mDark, color.RGBA{0, 0, 0, 0}, 0.6))

	// ── Pediment: triangular gable, APEX AT TOP ──────────────────────────────
	pedH := max(h*16/100, 16)
	pedY := entY - pedH // pedY = apex (highest point)

	for y := pedY; y <= entY; y++ {
		// t=1 at apex, t=0 at base — so slant=max when t=1 (narrow apex)
		t := float64(entY-y) / float64(pedH)
		slant := int(float64(templeW/2+4) * t)
		x1 := templeX + slant
		x2 := templeX + templeW - slant
		if x2 <= x1 {
			continue
		}
		// Tympanum fill: slightly lighter toward apex (receives more light)
		tympC := lerp(mMid, mMain, t*0.45)
		splashFill(img, x1, y, x2, y+1, tympC)

		// Raking cornice: 4px angled edge on both sides
		for k := 0; k < 4; k++ {
			shade := lerp(mShad, mDark, float64(k)/3.0)
			if x1+k >= 0 && x1+k < w {
				img.SetRGBA(x1+k, y, shade)
			}
			if x2-k >= 0 && x2-k < w {
				img.SetRGBA(x2-k, y, mShad)
			}
		}
	}

	// Horizontal geison (thick moulding at pediment base)
	geisH := max(h*2/100, 5)
	splashFill(img, cornX1-2, entY-geisH, cornX2+2, entY+2, mMid)
	splashFill(img, cornX1-2, entY-geisH, cornX2+2, entY-geisH+3, mHigh)
	splashFill(img, cornX1-2, entY-2, cornX2+2, entY+2, mShad)

	// Acroterion at apex (small decorative finial)
	acW := max(h*2/100, 5)
	acX := w / 2
	acY := pedY - acW - 1
	splashFill(img, acX-acW, acY, acX+acW, pedY, mMain)
	splashFill(img, acX-acW, acY, acX+acW, acY+2, mHigh)
	splashFill(img, acX-acW-2, acY, acX-acW, pedY, mShad)
	splashFill(img, acX+acW, acY, acX+acW+2, pedY, mShad)

	return img
}

// splashFill fills a rectangle clipped to image bounds.
func splashFill(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	b := img.Bounds()
	if x1 < b.Min.X {
		x1 = b.Min.X
	}
	if y1 < b.Min.Y {
		y1 = b.Min.Y
	}
	if x2 > b.Max.X {
		x2 = b.Max.X
	}
	if y2 > b.Max.Y {
		y2 = b.Max.Y
	}
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// splashCloud draws a fluffy elliptical cloud with bumps on top.
func splashCloud(img *image.RGBA, cx, cy, rx, ry int, main, shad color.RGBA, maxY int) {
	b := img.Bounds()
	// Top bumps
	for _, bx := range []int{-rx / 2, 0, rx / 2} {
		brx := rx * 2 / 3
		bry := ry * 4 / 3
		bcy := cy - ry/3
		for dy := -bry; dy <= bry; dy++ {
			for dx := -brx; dx <= brx; dx++ {
				px, py := cx+bx+dx, bcy+dy
				if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= maxY {
					continue
				}
				ex := float64(dx) / float64(brx)
				ey := float64(dy) / float64(bry)
				if ex*ex+ey*ey <= 1.0 {
					if dy > bry/3 {
						img.SetRGBA(px, py, shad)
					} else {
						img.SetRGBA(px, py, main)
					}
				}
			}
		}
	}
	// Main body ellipse
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			px, py := cx+dx, cy+dy
			if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= maxY {
				continue
			}
			ex := float64(dx) / float64(rx)
			ey := float64(dy) / float64(ry)
			if ex*ex+ey*ey <= 1.0 {
				if img.RGBAAt(px, py).R < 150 { // only draw over sky
					if dy > ry/3 {
						img.SetRGBA(px, py, shad)
					} else {
						img.SetRGBA(px, py, main)
					}
				}
			}
		}
	}
}
