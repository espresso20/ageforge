package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// export.go writes the sheets.

// The columns of every sheet, in order.
var columns = []string{"id", "where", "kind", "age", "keep", "max", "current", "yours", "notes", "same_as"}

const (
	colID = iota
	colWhere
	colKind
	colAge
	colKeep
	colMax
	colCurrent
	colYours
	colNotes
	colSameAs
)

// The files the export writes beside the area sheets.
const (
	firstHoursFile = "first-hours.csv"
	readmeFile     = "README.md"
	// keptFile holds cells the owner had filled that the new sheets have no
	// place for: the line is gone, or its text changed under the rewrite.
	keptFile = "kept-from-last-export.csv"
)

// firstHoursAges is how far first-hours.csv goes: the main menu and the
// first six ages, Primitive to Medieval.
const firstHoursAges = 6

// bom is the byte-order mark Excel needs to read a CSV file as UTF-8.
const bom = "\xEF\xBB\xBF"

// record writes a row as a line of a sheet.
func (c *catalog) record(r *row, yours, notes string) []string {
	rec := make([]string, len(columns))
	rec[colID] = r.id
	rec[colWhere] = r.where
	rec[colKind] = r.kind
	rec[colAge] = c.ageText(r)
	rec[colKeep] = keepColumn(r.kept())
	if r.max > 0 {
		rec[colMax] = strconv.Itoa(r.max)
	}
	rec[colCurrent] = r.current
	rec[colYours] = yours
	rec[colNotes] = notes
	rec[colSameAs] = r.sameAs
	return rec
}

// kept lists the pieces of the row's text a rewrite must carry over.
func (r *row) kept() []kept {
	return tokensOf(r.current, r.rules&rulesSkel != 0)
}

// sheet renders records as one CSV file.
func sheet(records [][]string) []byte {
	var buf bytes.Buffer
	buf.WriteString(bom)
	w := csv.NewWriter(&buf)
	w.UseCRLF = true
	_ = w.Write(columns)
	_ = w.WriteAll(records)
	w.Flush()
	return buf.Bytes()
}

// filled is what the owner had typed against one id in an earlier export.
type filled struct {
	current, yours, notes string
	file                  string
	clash                 bool // filled differently in two files
}

// export writes every sheet and the README into dir, and reports what it
// wrote to out.
func export(c *catalog, dir string, out io.Writer) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Cells filled in an earlier export of this folder are carried into the
	// new sheets, so exporting again never loses typed work.
	old, oldRecs, err := readFilled(dir)
	if err != nil {
		return err
	}
	var keptRecs [][]string
	used := map[string]bool{}
	carry := func(r *row) (yours, notes string) {
		f, ok := old[r.id]
		if !ok {
			return "", ""
		}
		used[r.id] = true
		switch {
		case f.clash:
			return "", ""
		case f.current == r.current:
			return f.yours, f.notes
		case f.yours != "" && restoreEdges(r.current, f.yours) == r.current:
			return "", f.notes // the rewrite is in the source now
		}
		used[r.id] = false
		return "", ""
	}

	files := map[string][]byte{}
	var firstRows []*row
	for _, a := range areas {
		var recs [][]string
		for _, r := range c.rows {
			if r.area != a {
				continue
			}
			yours, notes := carry(r)
			recs = append(recs, c.record(r, yours, notes))
			if inFirstHours(r) {
				firstRows = append(firstRows, r)
			}
		}
		files[a.file] = sheet(recs)
	}
	sort.SliceStable(firstRows, func(i, j int) bool {
		a, b := firstRows[i], firstRows[j]
		if a.seen != b.seen {
			return a.seen < b.seen
		}
		if a.pool != b.pool {
			return a.pool < b.pool
		}
		return a.age < b.age
	})
	var firstRecs [][]string
	for _, r := range firstRows {
		yours, notes := carry(r)
		firstRecs = append(firstRecs, c.record(r, yours, notes))
	}
	files[firstHoursFile] = sheet(firstRecs)

	for _, rec := range oldRecs {
		id := rec[colID]
		if (rec[colYours] != "" || rec[colNotes] != "") && !used[id] {
			keptRecs = append(keptRecs, rec)
		}
	}
	if len(keptRecs) > 0 {
		sort.SliceStable(keptRecs, func(i, j int) bool { return keptRecs[i][colID] < keptRecs[j][colID] })
		files[keptFile] = sheet(dedupe(keptRecs))
	} else {
		_ = os.Remove(filepath.Join(dir, keptFile))
	}
	files[readmeFile] = []byte(readme(c))

	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), files[n], 0o644); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Wrote %d sheets to %s\n\n", len(names)-1, dir)
	writeCounts(c, out)
	if n := len(firstRows); n > 0 {
		fmt.Fprintf(out, "\n%s holds %d of those rows: the main menu through the %s.\n", firstHoursFile, n, c.w.ageNames[firstHoursAges-1])
	}
	if len(keptRecs) > 0 {
		fmt.Fprintf(out, "\n%d filled rows from the last export have no place in the new sheets (the line is gone or has changed). They are in %s.\n", len(dedupe(keptRecs)), keptFile)
	}
	return nil
}

// inFirstHours reports whether a row belongs in first-hours.csv.
func inFirstHours(r *row) bool {
	return r.age >= 0 && r.age <= firstHoursAges && !r.unsure && r.area != areaSaveNames
}

// writeCounts prints rows and words by sheet.
func writeCounts(c *catalog, out io.Writer) {
	fmt.Fprintf(out, "%-34s %7s %8s\n", "sheet", "rows", "words")
	var rows, words int
	for _, a := range areas {
		n, w := 0, 0
		for _, r := range c.rows {
			if r.area == a {
				n++
				w += countWords(r.current)
			}
		}
		fmt.Fprintf(out, "%-34s %7d %8d\n", a.file, n, w)
		rows += n
		words += w
	}
	fmt.Fprintf(out, "%-34s %7d %8d\n", "total", rows, words)
}

// dedupe drops repeated records.
func dedupe(recs [][]string) [][]string {
	seen := map[string]bool{}
	var out [][]string
	for _, r := range recs {
		k := strings.Join(r, "\x00")
		if !seen[k] {
			seen[k] = true
			out = append(out, r)
		}
	}
	return out
}

// readSheet reads one CSV file written by export (or saved over by a
// spreadsheet): the byte-order mark is optional, the header row is checked.
func readSheet(path string) ([][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data = bytes.TrimPrefix(data, []byte(bom))
	rd := csv.NewReader(bytes.NewReader(data))
	rd.FieldsPerRecord = -1
	rd.LazyQuotes = true
	recs, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	if len(recs) == 0 {
		return nil, nil
	}
	head := recs[0]
	if len(head) < len(columns) {
		return nil, fmt.Errorf("%s: the first row should be %s", filepath.Base(path), strings.Join(columns, ","))
	}
	for i, name := range columns {
		if strings.TrimSpace(strings.ToLower(head[i])) != name {
			return nil, fmt.Errorf("%s: column %d should be %q, found %q", filepath.Base(path), i+1, name, head[i])
		}
	}
	var out [][]string
	for _, rec := range recs[1:] {
		for len(rec) < len(columns) {
			rec = append(rec, "")
		}
		rec = rec[:len(columns)]
		// A spreadsheet may write a cell's line breaks as CR LF.
		for i := range rec {
			rec[i] = strings.ReplaceAll(rec[i], "\r\n", "\n")
		}
		if strings.TrimSpace(rec[colID]) == "" {
			continue
		}
		rec[colID] = strings.TrimSpace(rec[colID])
		out = append(out, rec)
	}
	return out, nil
}

// sheetFiles lists the CSV files of a folder, sorted.
func sheetFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".csv") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// readFilled reads what the sheets already in dir have filled in.
func readFilled(dir string) (map[string]*filled, [][]string, error) {
	names, err := sheetFiles(dir)
	if err != nil {
		return nil, nil, err
	}
	out := map[string]*filled{}
	var all [][]string
	for _, n := range names {
		recs, err := readSheet(filepath.Join(dir, n))
		if err != nil {
			return nil, nil, err
		}
		for _, rec := range recs {
			if rec[colYours] == "" && rec[colNotes] == "" {
				continue
			}
			all = append(all, rec)
			f, ok := out[rec[colID]]
			if !ok {
				out[rec[colID]] = &filled{current: rec[colCurrent], yours: rec[colYours], notes: rec[colNotes], file: n}
				continue
			}
			if rec[colYours] != "" && f.yours != "" && rec[colYours] != f.yours {
				f.clash = true
			}
			if f.yours == "" {
				f.yours = rec[colYours]
			}
			if f.notes == "" {
				f.notes = rec[colNotes]
			}
		}
	}
	return out, all, nil
}

// restoreEdges gives a rewrite the white space the original starts and ends
// with. Spreadsheets and people both lose it, and the code depends on it:
// " idle" is joined to a number.
func restoreEdges(current, yours string) string {
	lead, trail := edges(current)
	return lead + strings.TrimSpace(yours) + trail
}
