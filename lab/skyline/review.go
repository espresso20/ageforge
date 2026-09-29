package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/theme"
)

// review renders an 80-column crop of a state for close inspection:
//
//	-review medieval_age:0.45:1   (state : time of day : screens west of the present)
func writeReview(dir, out, spec string) error {
	parts := strings.Split(spec, ":")
	state := parts[0]
	tod := 0.45
	back := 0
	themeKey := "forge"
	if len(parts) > 1 {
		tod, _ = strconv.ParseFloat(parts[1], 64)
	}
	if len(parts) > 2 {
		back, _ = strconv.Atoi(parts[2])
	}
	if len(parts) > 3 {
		themeKey = parts[3]
	}
	W, H := 80, 45
	w, err := loadWorld(dir, state, sceneRows(H))
	if err != nil {
		return err
	}
	th, _ := theme.ByKey(themeKey)
	cam := camFor(w, 160, "end") + 80 - back*80
	v := View{W: W, H: H, Cam: max(0, cam), TOD: tod, Theme: th, Cursor: -1, Weather: -1, Frame: 40}
	_, h := shot(w, v, false)
	name := fmt.Sprintf("review_%s_%d.html", state, back)
	return os.WriteFile(filepath.Join(out, name), []byte(wrapHTML(name, "#000", "#ccc", []string{h}, 1)), 0o644)
}
