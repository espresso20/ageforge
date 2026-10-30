// Command bannerpng draws the README banner, docs/banner.png, from the main
// menu's own wordmark, sky and colours (ui.BannerImage).
//
//	go run ./cmd/bannerpng
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/ui"
)

func main() {
	icon := flag.Bool("icon", false, "write the square icon instead of the banner")
	out := flag.String("o", "", "where to write the PNG")
	cellW := flag.Int("cw", 10, "pixels per cell, across")
	cellH := flag.Int("ch", 20, "pixels per cell, down")
	flag.Parse()

	// Set default output path based on which image is being generated
	if *out == "" {
		if *icon {
			*out = "docs/icon.png"
		} else {
			*out = "docs/banner.png"
		}
	}

	if *icon {
		if err := runIcon(*out); err != nil {
			fmt.Fprintln(os.Stderr, "bannerpng:", err)
			os.Exit(1)
		}
	} else {
		if err := run(*out, *cellW, *cellH); err != nil {
			fmt.Fprintln(os.Stderr, "bannerpng:", err)
			os.Exit(1)
		}
	}
}

func runIcon(out string) error {
	const size = 1024
	img := ui.IconImage(size)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

func run(out string, cellW, cellH int) error {
	if cellW < 1 || cellH < 1 {
		return fmt.Errorf("cell size %dx%d: both must be at least 1", cellW, cellH)
	}
	img := ui.BannerImage(len(rules.Core().AgeKeys()), cellW, cellH)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}
