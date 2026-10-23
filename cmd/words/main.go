// Command words is the copy-editing export and import for the game's text.
//
//	go run ./cmd/words export [dir]   write every line a player reads to CSV sheets
//	go run ./cmd/words import [dir]   apply the rows whose "yours" column is filled
//	go run ./cmd/words status         how much is rewritten, by sheet
//	go run ./cmd/words left           what the export leaves out, and why
//	go run ./cmd/words vibe-export [dir]  write the rows VibeMode's script has no line for yet
//	go run ./cmd/words vibe-lint [dir]    check the script's rows against the export and the source
//	go run ./cmd/words vibe-wire [sheet]  write the vibe calls into the source for a sheet's texts (--dry-run to look, --only=ui.input. for a part)
//
// The export finds the text by parsing the Go source, so nothing in the game
// has to register its strings anywhere. The import rewrites the same string
// literals in place, by position, and runs gofmt over what it touched.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// defaultDir is where the sheets go, from the repository root. It is in
// .gitignore: the sheets are working files, not source.
const defaultDir = "words/export"

func main() {
	if err := run(os.Args[1:], ".", os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "words:", err)
		os.Exit(1)
	}
}

// run carries out one command line from the directory wd.
func run(args []string, wd string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: go run ./cmd/words export|import|status|left|vibe-export|vibe-lint|vibe-wire [dir]")
	}
	root, err := findRoot(wd)
	if err != nil {
		return err
	}
	// Flags may come before or after the folder: "import sheets --dry-run"
	// must not be read as a real import of two folders.
	var flags, rest []string
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
		} else {
			rest = append(rest, a)
		}
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(out)
	dry := fs.Bool("dry-run", false, "import: print what would change and what would be refused, and write nothing")
	only := fs.String("only", "", "vibe-wire: wire only the rows whose id starts with this")
	noSame := fs.Bool("no-same", false, "import: do not give a blank same_as row the rewrite of the row it points to")
	if err := fs.Parse(flags); err != nil {
		return err
	}
	if len(rest) > 1 {
		return fmt.Errorf("%s takes one folder, got %d: %s", args[0], len(rest), strings.Join(rest, " "))
	}
	if *noSame && args[0] != "import" {
		return fmt.Errorf("--no-same belongs to import")
	}
	if *only != "" && args[0] != "vibe-wire" {
		return fmt.Errorf("--only belongs to vibe-wire")
	}
	if *dry && args[0] != "import" && args[0] != "vibe-wire" {
		return fmt.Errorf("--dry-run belongs to import and vibe-wire")
	}
	dir := filepath.Join(root, filepath.FromSlash(defaultDir))
	if args[0] == "vibe-export" {
		dir = filepath.Join(root, filepath.FromSlash(vibeTodoDir))
	}
	if len(rest) == 1 {
		dir = rest[0]
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(wd, dir)
		}
	}
	switch args[0] {
	case "vibe-wire":
		sheet := areaLog.file
		if len(rest) == 1 {
			sheet = rest[0]
		}
		c, err := read(root)
		if err != nil {
			return err
		}
		return vibeWire(c, root, sheet, *only, *dry, out)
	case "export":
		c, err := read(root)
		if err != nil {
			return err
		}
		return export(c, dir, out)
	case "import":
		return importSheets(root, dir, importOptions{dryRun: *dry, noSame: *noSame, build: true, regenerate: true}, out)
	case "status":
		return status(root, out)
	case "vibe-export":
		c, err := read(root)
		if err != nil {
			return err
		}
		return vibeExport(c, root, "", dir, out)
	case "vibe-lint":
		c, err := read(root)
		if err != nil {
			return err
		}
		sheets := ""
		if len(rest) == 1 {
			sheets = dir
		}
		n, err := vibeLint(c, root, sheets, out)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("the script has %d errors", n)
		}
		return nil
	case "left":
		c, err := read(root)
		if err != nil {
			return err
		}
		for _, l := range c.left {
			fmt.Fprintf(out, "%s:%d\t%s\t%q\n", l.u.f.rel, l.u.line, l.why, l.u.text)
		}
		return nil
	}
	return fmt.Errorf("unknown command %q: use export, import, status, left, vibe-export, vibe-lint or vibe-wire", args[0])
}
