package words

import (
	"bufio"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
)

// This file reads the VibeMode script: the sheets in words/vibe, embedded in the
// binary. A sheet has a header row naming (at least) the columns id and text,
// then a row per line of the game: the row id the export gives that line, and
// the line in the VibeMode script. A row with no text is no row, and the game
// keeps the plain line.

//go:embed vibe/*.csv
var vibeFS embed.FS

// VibeDir is where the script's sheets live, from the repository root.
const VibeDir = "words/vibe"

// Row is one line of the script.
type Row struct {
	ID   string
	Text string
	File string // the sheet it came from
	Line int    // its line in the sheet, from 1
}

// VibeRows reads every sheet embedded in the binary, in file-name order. It
// returns the rows with text and the problems it found (a sheet it cannot
// read, a row id twice). A problem never stops the rest from loading. For an id
// that appears twice, the first row is kept.
func VibeRows() ([]Row, []string) {
	entries, err := vibeFS.ReadDir("vibe")
	if err != nil {
		return nil, []string{err.Error()}
	}
	var rows []Row
	var problems []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".csv") {
			continue
		}
		f, err := vibeFS.Open("vibe/" + e.Name())
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		got, probs := ParseRows(e.Name(), f)
		f.Close()
		rows = append(rows, got...)
		problems = append(problems, probs...)
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].File < rows[j].File })
	seen := map[string]Row{}
	out := rows[:0]
	for _, r := range rows {
		if first, dup := seen[r.ID]; dup {
			problems = append(problems, fmt.Sprintf("%s:%d: the id %s is also in %s:%d", r.File, r.Line, r.ID, first.File, first.Line))
			continue
		}
		seen[r.ID] = r
		out = append(out, r)
	}
	return out, problems
}

// ParseRows reads one sheet: a header row with the columns id and text (in any
// order, with other columns beside them, which are ignored), then a row per
// line. A row with no id or no text is left out. A row that cannot be read, is
// too short for the header or has an id with white space in it is left out and
// reported.
func ParseRows(file string, r io.Reader) ([]Row, []string) {
	br := bufio.NewReader(r)
	if b, err := br.Peek(3); err == nil && string(b) == "\xEF\xBB\xBF" {
		_, _ = br.Discard(3) // the byte-order mark Excel writes
	}
	cr := csv.NewReader(br)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	head, err := cr.Read()
	if err == io.EOF {
		return nil, []string{file + ": the sheet is empty: it needs a header row with id and text"}
	}
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", file, err)}
	}
	idCol, textCol := -1, -1
	for i, h := range head {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "id":
			idCol = i
		case "text":
			textCol = i
		}
	}
	if idCol < 0 || textCol < 0 {
		return nil, []string{file + ": the header row has no id and text columns"}
	}
	var rows []Row
	var problems []string
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		line, _ := cr.FieldPos(0)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s:%d: %v", file, line, err))
			continue
		}
		if idCol >= len(rec) || textCol >= len(rec) {
			problems = append(problems, fmt.Sprintf("%s:%d: the row is too short for its header", file, line))
			continue
		}
		id, text := strings.TrimSpace(rec[idCol]), rec[textCol]
		if id == "" || strings.TrimSpace(text) == "" {
			continue
		}
		if strings.ContainsAny(id, " \t\r\n") {
			problems = append(problems, fmt.Sprintf("%s:%d: %q is not a row id", file, line, id))
			continue
		}
		rows = append(rows, Row{ID: id, Text: text, File: file, Line: line})
	}
	return rows, problems
}
