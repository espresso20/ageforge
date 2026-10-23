package words

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRows(t *testing.T) {
	rows, probs := ParseRows("a.csv", strings.NewReader(
		"\xEF\xBB\xBF\"text\",notes,ID\r\n"+
			"Crash Pad,x,building.hut.name\r\n"+
			"\"A mid, quiet shelter\",,building.hut.description\r\n"+
			",,building.hut.flavor\r\n"+ // no text: no row
			"   ,,building.stash.name\r\n"+ // blank text: no row
			"Orphan,,\r\n"+ // no id
			"\r\n"+
			"Short\r\n"+
			"Bad,,not an id\r\n"))
	if len(rows) != 2 || rows[0] != (Row{"building.hut.name", "Crash Pad", "a.csv", 2}) ||
		rows[1] != (Row{"building.hut.description", "A mid, quiet shelter", "a.csv", 3}) {
		t.Errorf("rows: %+v", rows)
	}
	if len(probs) != 2 || !strings.Contains(probs[0], "a.csv:8: the row is too short") || !strings.Contains(probs[1], "a.csv:9") {
		t.Errorf("problems: %q", probs)
	}

	for name, in := range map[string]string{"empty": "", "no text column": "id,notes\r\nx,y\r\n", "no id column": "text\r\nx\r\n"} {
		if rows, probs := ParseRows(name, strings.NewReader(in)); len(rows) != 0 || len(probs) != 1 {
			t.Errorf("%s: rows %v, problems %q", name, rows, probs)
		}
	}
}

// TestTheEmbeddedScriptIsSound: every sheet in words/vibe is in the binary, reads
// without a problem, gives each id once and has no em dash or exclamation mark.
// This reads the folder the source is in, not the binary, so a sheet the embed
// pattern would miss (a file of another extension) shows.
func TestTheEmbeddedScriptIsSound(t *testing.T) {
	rows, probs := VibeRows()
	if len(probs) > 0 {
		t.Errorf("the script has problems: %q", probs)
	}
	if len(rows) == 0 {
		t.Fatal("the script has no rows")
	}
	disk, err := filepath.Glob(filepath.Join("vibe", "*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range disk {
		if !strings.HasSuffix(f, ".csv") {
			t.Errorf("%s is in the script's folder and is not a .csv sheet: the binary does not carry it", f)
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "—") || strings.Contains(string(data), "!") {
			t.Errorf("%s has an em dash or an exclamation mark: the house style has neither", f)
		}
	}
}
