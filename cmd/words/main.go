// Command words is the copy-editing export and import for the game's text.
//
//	go run ./cmd/words export [dir]   write every line a player reads to CSV sheets
//	go run ./cmd/words import [dir]   apply the rows whose "yours" column is filled
//	go run ./cmd/words status         how much is rewritten, by sheet
//	go run ./cmd/words left           what the export leaves out, and why
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
		return fmt.Errorf("usage: go run ./cmd/words export|import|status|left [dir]")
	}
	root, err := findRoot(wd)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(out)
	dry := fs.Bool("dry-run", false, "import: print what would change and what would be refused, and write nothing")
	noSame := fs.Bool("no-same", false, "import: do not give a blank same_as row the rewrite of the row it points to")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	dir := filepath.Join(root, filepath.FromSlash(defaultDir))
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(wd, dir)
		}
	}
	switch args[0] {
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
	return fmt.Errorf("unknown command %q: use export, import, status or left", args[0])
}
