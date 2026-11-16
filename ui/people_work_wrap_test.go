package ui

import (
	"image"
	"strings"
	"testing"
)

// workPanel is the People panel on the Work tab in the Medieval Age at w by h,
// with area row picked.
func workPanel(t *testing.T, w, h, row int) *peoplePanel {
	t.Helper()
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "medieval_age")
	p := newPeoplePanel()
	p.open(eng.GetState())
	p.view.tab, p.view.row = tabWork, row
	p.Layout(image.Rect(0, 0, w, h))
	return p
}

func gridRows(p *peoplePanel) []string {
	rows := make([]string, p.h)
	for y := range rows {
		var sb strings.Builder
		for x := 0; x < p.w; x++ {
			sb.WriteRune(p.grid.at(x, y).r)
		}
		rows[y] = sb.String()
	}
	return rows
}

func TestWorkAreasWrapTheirBuildingsInsteadOfCuttingThem(t *testing.T) {
	p := workPanel(t, 157, 46, 0)
	text := strings.Join(gridRows(p), "\n")
	if strings.Contains(text, "…") {
		t.Errorf("the Work tab cuts a line with an ellipsis:\n%s", text)
	}
	for _, r := range p.state.Buildings {
		if r.WorkerCapacity > 0 && r.Count > 0 && !strings.Contains(text, strings.Split(r.Name, " ")[0]) {
			t.Errorf("%s is not on the Work tab", r.Name)
		}
	}
	// A click on a wrapped row of an area picks that area.
	for row := 0; row < 6; row++ {
		if _, ok := p.grid.hitAt(hitRow, image.Pt(5, 4+row)); !ok {
			t.Fatalf("no area under row %d", row)
		}
	}
	// The area below the first one still picks the second.
	n, _ := p.grid.hitAt(hitRow, image.Pt(5, 4))
	m, _ := p.grid.hitAt(hitRow, image.Pt(5, 5))
	if n != m {
		t.Errorf("rows 4 and 5 belong to areas %d and %d, want one area", n, m)
	}
}

func TestWorkAreasKeepTwoRowsEachWhenTheyFit(t *testing.T) {
	p := workPanel(t, 279, 74, 0)
	rows := gridRows(p)
	for _, i := range []int{4, 6, 8} {
		if !strings.Contains(rows[i], "auto") || strings.Contains(rows[i+1], "auto") {
			t.Errorf("row %d is not an area row followed by its buildings: %q / %q", i, rows[i], rows[i+1])
		}
	}
}

func TestWorkTabShowsThePickedAreaWhole(t *testing.T) {
	for row := 0; row < 6; row++ {
		p := workPanel(t, 157, 20, row)
		if _, ok := p.grid.hitAt(hitRow, image.Pt(5, 4)); !ok {
			t.Fatal("no area at the top")
		}
		found := false
		for y := 0; y < p.h; y++ {
			if n, ok := p.grid.hitAt(hitRow, image.Pt(5, y)); ok && n == row {
				found = true
			}
		}
		if !found {
			t.Errorf("area %d is not on a 20 row Work tab", row)
		}
	}
}
