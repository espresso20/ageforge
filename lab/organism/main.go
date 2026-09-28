// Command organism is the Map Lab's "organism" concept: the civilisation
// drawn as one living tree. Limbs are lineages, leaves are how well they are
// staffed, fruit are wonders, the trunk is the population, roots are storage
// reaching down through strata of history, and the horizon holds the other
// civs, with trade drifting between crowns as pollen.
//
//	go run ./lab/organism -gen                 # play the smoke bot, save states
//	go run ./lab/organism -age medieval_age    # interactive
//	go run ./lab/organism -capture             # write lab/organism/captures
//	go run ./lab/organism -dump bronze_age -w 100 -h 30
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	gen := flag.Bool("gen", false, "play the smoke bot and save one state per age into -states")
	seed := flag.Int64("seed", 7, "bot seed for -gen")
	stop := flag.String("stop", "transcendent_age", "last age to save with -gen")
	states := flag.String("states", "lab/organism/states", "directory holding the saves")
	age := flag.String("age", "medieval_age", "save to open interactively")
	themeKey := flag.String("theme", "forge", "theme key")
	capture := flag.Bool("capture", false, "write every capture into -out")
	out := flag.String("out", "lab/organism/captures", "capture directory")
	dump := flag.String("dump", "", "print one frame of this save as text")
	sel := flag.String("sel", "", "with -dump: lineage key to inspect")
	w := flag.Int("w", 160, "width for -dump")
	h := flag.Int("h", 48, "height for -dump")
	t := flag.Float64("t", 0, "with -dump: animation time")
	trade := flag.String("trade", "", "open every trade route in this age's save, save it as <age>_trade")
	flag.Parse()

	var err error
	switch {
	case *gen:
		logf := func(f string, a ...any) { fmt.Printf(f+"\n", a...) }
		err = generate(*states, *seed, *stop, logf)
		for _, a := range []string{"colonial_age", "information_age"} {
			if err == nil {
				err = withTrade(*states, a, logf)
			}
		}
	case *trade != "":
		err = withTrade(*states, *trade, func(f string, a ...any) { fmt.Printf(f+"\n", a...) })
	case *dump != "":
		err = dumpOne(*states, *dump, *themeKey, *sel, *w, *h, *t)
	case *capture:
		err = captureAll(*states, *out)
	default:
		err = interactive(*states, *age, *themeKey, nil)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dumpOne(states, name, themeKey, sel string, w, h int, t float64) error {
	cs := captureSpec{State: name, Theme: themeKey, W: w, H: h, Sel: sel}
	dir, err := os.MkdirTemp("", "organism")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	cs.Out = "dump"
	if t > 0 {
		cs.Frames = 1
	}
	if err := runCapture(states, dir, cs); err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(dir, "dump.txt"))
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}

// captures is the comparison page's set.
var captures = []captureSpec{
	{State: "primitive_seedling", Out: "00_primitive_seedling", Theme: "forge", W: 160, H: 48},
	{State: "primitive_age", Out: "01_primitive_age", Theme: "forge", W: 160, H: 48},
	{State: "bronze_age", Out: "02_bronze_age", Theme: "forge", W: 160, H: 48},
	{State: "medieval_age", Out: "03_medieval_age", Theme: "forge", W: 160, H: 48},
	{State: "industrial_age", Out: "04_industrial_age", Theme: "forge", W: 160, H: 48},
	{State: "digital_age", Out: "05_digital_age", Theme: "forge", W: 160, H: 48},
	{State: "cyberpunk_age", Out: "06_cyberpunk_age", Theme: "forge", W: 160, H: 48},
	{State: "galactic_age", Out: "07_galactic_age", Theme: "forge", W: 160, H: 48},
	{State: "transcendent_age", Out: "08_transcendent_age", Theme: "forge", W: 160, H: 48},
	{State: "medieval_age", Out: "10_medieval_age_light", Theme: "daylight", W: 160, H: 48},
	{State: "digital_age", Out: "11_digital_age_light", Theme: "daylight", W: 160, H: 48},
	{State: "industrial_age", Out: "12_industrial_age_parchment", Theme: "parchment", W: 160, H: 48},
	{State: "industrial_age", Out: "20_industrial_age_animated", Theme: "forge", W: 160, H: 48, Frames: 4},
	{State: "galactic_age", Out: "21_galactic_age_animated", Theme: "forge", W: 160, H: 48, Frames: 4},
	{State: "medieval_age", Out: "30_medieval_age_100x30", Theme: "forge", W: 100, H: 30},
	{State: "cyberpunk_age", Out: "31_cyberpunk_age_80x24", Theme: "forge", W: 80, H: 24},
	{State: "medieval_age", Out: "32_medieval_age_mini_40x15", Theme: "forge", W: 40, H: 15},
	{State: "digital_age", Out: "33_digital_age_mini_40x15_light", Theme: "daylight", W: 40, H: 15},
	{State: "industrial_age", Out: "40_industrial_age_inspect_metal", Theme: "forge", W: 160, H: 48, Sel: "metallurgy"},
	{State: "modern_age_catastrophe", Out: "41_modern_age_catastrophe", Theme: "forge", W: 160, H: 48, Frames: 4},
	{State: "information_age_trade", Out: "43_information_age_trade_animated", Theme: "forge", W: 160, H: 48, Frames: 4},
	{State: "colonial_age_trade", Out: "44_colonial_age_trade_cosmic_theme", Theme: "cosmic", W: 160, H: 48},
	{State: "industrial_age", Out: "42_industrial_age_16color", Theme: "forge", W: 160, H: 48, Q16: true},
}

func captureAll(states, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, cs := range captures {
		if err := runCapture(states, out, cs); err != nil {
			return err
		}
		fmt.Println("wrote", filepath.Join(out, cs.Out))
	}
	// An index for the comparison page.
	var b strings.Builder
	b.WriteString("<!doctype html><meta charset=utf-8><title>organism captures</title>" +
		"<body style=\"background:#111;color:#ddd;font:14px/1.6 ui-monospace,Menlo,monospace;padding:24px\">" +
		"<h1 style=\"font-size:16px\">organism: captures</h1><ul>\n")
	for _, cs := range captures {
		fmt.Fprintf(&b, "<li><a style=\"color:#9ad44a\" href=\"%s.html\">%s</a> · <a style=\"color:#888\" href=\"%s.txt\">txt</a> · %s, %s, %dx%d%s</li>\n",
			cs.Out, cs.Out, cs.Out, cs.State, cs.Theme, cs.W, cs.H, map[bool]string{true: ", animated", false: ""}[cs.Frames > 1])
	}
	b.WriteString("</ul></body>\n")
	return os.WriteFile(filepath.Join(out, "index.html"), []byte(b.String()), 0o644)
}
