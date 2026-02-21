package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// wonderInfo holds config + state for a wonder
type wonderInfo struct {
	key     string
	name    string
	ageKey  string
	ageName string
	def     config.BuildingDef
}

// getWonderList returns all wonders in age order
func getWonderList() []wonderInfo {
	ages := config.Ages()
	buildings := config.BuildingByKey()

	var wonders []wonderInfo
	for _, age := range ages {
		for _, bKey := range age.UnlockBuildings {
			if def, ok := buildings[bKey]; ok && def.Category == "wonder" {
				wonders = append(wonders, wonderInfo{
					key:     bKey,
					name:    def.Name,
					ageKey:  age.Key,
					ageName: age.Name,
					def:     def,
				})
				break // one wonder per age
			}
		}
	}
	return wonders
}

// ─── Current Age Wonder Panel (dashboard strip) ───

// WonderPanel shows the current age's wonder with pixel art + perks
type WonderPanel struct {
	root     *tview.Flex
	lastHash uint64
}

// NewWonderPanel creates the single-wonder panel for the dashboard
func NewWonderPanel() *WonderPanel {
	wp := &WonderPanel{}
	wp.root = tview.NewFlex().SetDirection(tview.FlexColumn)
	wp.root.SetBorder(true).SetTitle(" Wonder ").SetTitleColor(ColorTitle)
	return wp
}

// Primitive returns the underlying tview primitive
func (wp *WonderPanel) Primitive() tview.Primitive {
	return wp.root
}

// UpdateState refreshes the current-age wonder display
func (wp *WonderPanel) UpdateState(state game.GameState) {
	// Find the wonder for the current age
	var current *wonderInfo
	for _, w := range getWonderList() {
		if w.ageKey == state.Age {
			wCopy := w
			current = &wCopy
			break
		}
	}

	// Hash to detect changes
	var h uint64
	h = hashKey(state.Age)
	if current != nil {
		if bs, ok := state.Buildings[current.key]; ok {
			h ^= uint64(bs.Count)*7 + hashKey(current.key)
			if bs.Unlocked {
				h ^= 13
			}
		}
	}
	// Also hash wonder count for speed display
	for _, w := range getWonderList() {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			h ^= hashKey(w.key) * 5
		}
	}
	if h == wp.lastHash {
		return
	}
	wp.lastHash = h

	wp.root.Clear()

	_, _, totalW, ht := wp.root.GetInnerRect()
	if totalW < 10 || ht < 3 {
		return
	}

	if current == nil {
		// No wonder for this age (shouldn't happen, but handle gracefully)
		tv := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
		tv.SetText("[gray]No wonder available this age[-]")
		wp.root.AddItem(tv, 0, 1, false)
		return
	}

	built := false
	unlocked := false
	if bs, ok := state.Buildings[current.key]; ok {
		built = bs.Count > 0
		unlocked = bs.Unlocked
	}
	_ = unlocked

	// Left side: pixel art
	pixH := ht * 4
	pixW := ht * 4 // roughly square
	if pixW < 16 {
		pixW = 16
	}

	imgWidget := tview.NewImage()
	imgWidget.SetColors(tview.TrueColor)
	img := renderWonderIcon(*current, pixW, pixH, built)
	imgWidget.SetImage(img)

	// Right side: text info
	infoTV := tview.NewTextView().SetDynamicColors(true)

	var sb strings.Builder
	if built {
		fmt.Fprintf(&sb, "[gold::b]★ %s[-]\n", current.name)
		fmt.Fprintf(&sb, "[green]BUILT[-] — [gray]%s[-]\n\n", current.ageName)
	} else {
		fmt.Fprintf(&sb, "[yellow::b]%s[-]\n", current.name)
		fmt.Fprintf(&sb, "[gray]%s — Not yet built[-]\n\n", current.ageName)
	}

	// Show effects/perks
	fmt.Fprintf(&sb, "[cyan]Perks:[-]\n")
	for _, eff := range current.def.Effects {
		fmt.Fprintf(&sb, "  %s\n", formatEffect(eff))
	}
	fmt.Fprintf(&sb, "  [gold]+0.5x game speed[-]\n")

	// Build cost if not built
	if !built {
		fmt.Fprintf(&sb, "\n[cyan]Cost:[-]\n")
		costKeys := make([]string, 0, len(current.def.BaseCost))
		for k := range current.def.BaseCost {
			costKeys = append(costKeys, k)
		}
		sort.Strings(costKeys)
		for _, k := range costKeys {
			v := current.def.BaseCost[k]
			have := 0.0
			if rs, ok := state.Resources[k]; ok {
				have = rs.Amount
			}
			clr := "red"
			if have >= v {
				clr = "green"
			}
			fmt.Fprintf(&sb, "  [%s]%s: %s / %s[-]\n", clr, k, FormatNumber(have), FormatNumber(v))
		}
		fmt.Fprintf(&sb, "\n[gray]Build ticks: %d[-]\n", current.def.BuildTicks)
	}

	// Wonder count / speed
	wonderCount := 0
	for _, w := range getWonderList() {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			wonderCount++
		}
	}
	maxSpeed := 1.0 + float64(wonderCount)*0.5
	fmt.Fprintf(&sb, "\n[gold]Wonders built: %d[-] — [cyan]Max speed: %.1fx[-]", wonderCount, maxSpeed)

	infoTV.SetText(sb.String())

	// Layout: pixel art on left, info on right
	artWidth := totalW / 3
	if artWidth < 10 {
		artWidth = 10
	}
	if artWidth > 25 {
		artWidth = 25
	}

	wp.root.AddItem(imgWidget, artWidth, 0, false)
	wp.root.AddItem(infoTV, 0, 1, false)
}

// formatEffect formats a building effect for display
func formatEffect(eff config.Effect) string {
	switch eff.Type {
	case "production":
		return fmt.Sprintf("[green]+%.1f %s/tick[-]", eff.Value, eff.Target)
	case "capacity":
		return fmt.Sprintf("[yellow]+%.0f %s cap[-]", eff.Value, eff.Target)
	case "storage":
		if eff.Target == "all" {
			return fmt.Sprintf("[yellow]+%.0f all storage[-]", eff.Value)
		}
		return fmt.Sprintf("[yellow]+%.0f %s storage[-]", eff.Value, eff.Target)
	case "bonus":
		return fmt.Sprintf("[cyan]+%.0f%% %s[-]", eff.Value*100, eff.Target)
	default:
		return fmt.Sprintf("%s %s: %.1f", eff.Type, eff.Target, eff.Value)
	}
}

// renderWonderIcon generates pixel art for a single wonder
func renderWonderIcon(w wonderInfo, pixW, pixH int, built bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, pixW, pixH))

	era := eraFromAge(w.ageKey)
	pal := getTerrainPalette(era)
	vis := getBuildingVisual(w.key, "wonder")

	// Background
	for y := 0; y < pixH; y++ {
		for x := 0; x < pixW; x++ {
			n := noise2D(x, y, hashKey(w.key))
			bg := lerp(pal.Ground, pal.GroundAlt, n*0.3)
			img.SetRGBA(x, y, bg)
		}
	}

	cx, cy := pixW/2, pixH/2
	r := min(pixW, pixH) / 3
	if r < 4 {
		r = 4
	}

	b := bldInfo{
		key:      w.key,
		category: "wonder",
		x:        cx,
		y:        cy,
		size:     r * 2,
	}

	if built {
		// Glow halo
		glowR := r + 4
		for dy := -glowR; dy <= glowR; dy++ {
			for dx := -glowR; dx <= glowR; dx++ {
				d := math.Sqrt(float64(dx*dx + dy*dy))
				if d <= float64(r) || d >= float64(glowR) {
					continue
				}
				px, py := cx+dx, cy+dy
				if px < 0 || px >= pixW || py < 0 || py >= pixH {
					continue
				}
				fade := 1.0 - (d-float64(r))/float64(glowR-r)
				gc := pal.Accent1
				if era >= 6 {
					gc = pal.Accent2
				}
				existing := img.RGBAAt(px, py)
				img.SetRGBA(px, py, lerp(existing, gc, fade*0.45))
			}
		}
		drawBuildingShape(img, pixW, pixH, b, vis, era, 0)
	} else {
		// Ghost: dim outline
		drawGhostBuilding(img, pixW, pixH, cx, cy, r, vis)
	}

	return img
}

// drawGhostBuilding draws a translucent outline of an available-but-unbuilt wonder
func drawGhostBuilding(img *image.RGBA, w, h, cx, cy, r int, vis BuildingVisual) {
	ghostPrimary := color.RGBA{
		R: vis.Primary.R / 3,
		G: vis.Primary.G / 3,
		B: vis.Primary.B / 3,
		A: 255,
	}
	ghostAccent := color.RGBA{
		R: vis.Accent.R / 3,
		G: vis.Accent.G / 3,
		B: vis.Accent.B / 3,
		A: 255,
	}

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			inside, isAccent := testShape(vis.Shape, dx, dy, r)
			if !inside {
				continue
			}
			px, py := cx+dx, cy+dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}

			bc := ghostPrimary
			if isAccent {
				bc = ghostAccent
			}

			// Stipple for "buildable" hint
			if (dx+dy)%3 == 0 {
				bc = lerp(bc, c(80, 80, 60), 0.3)
			}

			// Edge highlight
			isEdge := false
			for _, dd := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				ni, _ := testShape(vis.Shape, dx+dd[0], dy+dd[1], r)
				if !ni {
					isEdge = true
					break
				}
			}
			if isEdge {
				bc = lerp(bc, c(120, 110, 60), 0.5)
			}

			img.SetRGBA(px, py, bc)
		}
	}
}
