package flavor

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// validFolder is the smallest folder the loader accepts: one pool in each
// Moment's file and one bank, with room for a test to change one file.
func validFolder() fstest.MapFS {
	folder := fstest.MapFS{}
	for _, name := range textFiles {
		text := "== " + name + "_any\nplain kit | A sentence about nothing\n"
		if name == "harb_banks" {
			text = "== words harb_shared\nthe bank entry\n"
		}
		folder[name+".txt"] = &fstest.MapFile{Data: []byte(text)}
	}
	return folder
}

func TestTheShippedTextFilesAreWellFormed(t *testing.T) {
	entries, err := ReadEntries(os.DirFS("text"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := map[string]string{}
	for _, row := range VibeRows() {
		catalog[row.ID] = row.Plain
	}
	if len(entries) != len(catalog) {
		t.Fatalf("the files hold %d entries and the game's catalog has %d rows", len(entries), len(catalog))
	}
	for _, entry := range entries {
		if text, ok := catalog[entry.ID]; !ok || text != entry.Text {
			t.Errorf("%s:%d reads %q as %s, the game has %q", entry.File, entry.Line, entry.Text, entry.ID, text)
		}
		source, err := os.ReadFile("text/" + entry.File)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(source[entry.From:entry.To]); got != entry.Text {
			t.Errorf("%s:%d: the offsets give %q, the entry is %q", entry.File, entry.Line, got, entry.Text)
		}
	}
}

func TestMalformedTextFilesAreRefused(t *testing.T) {
	cases := map[string]struct{ file, want string }{
		"no separator":      {"== p\nplain kit text without bars\n", "tags | text"},
		"unknown tag":       {"== p\nplain kit loud | Words\n", `"loud" is not a register`},
		"two registers":     {"== p\nplain wry kit | Words\n", "two register tags"},
		"no register":       {"== p\nkit | Words\n", "no register"},
		"no topic":          {"== p\nplain | Words\n", "no topic"},
		"bad tone":          {"== p\nplain kit tone=sad | Words\n", "sad is not a tone"},
		"bad need":          {"== p\nplain kit need=gold | Words\n", "gold is not a need"},
		"two tones":         {"== p\nplain kit tone=grim tone=wry | Words\n", "two tone tags"},
		"slot without mark": {"== p\nplain kit slot=bank | Words\n", "exactly when"},
		"mark without slot": {"== p\nplain kit | Words about ~\n", "exactly when"},
		"empty pool":        {"== p\n", "has no sentences"},
		"bad era":           {"== p eras=medieval\nplain kit | Words\n", "medieval is not an era"},
		"bad age":           {"== p ages=stone_age,lost_age\nplain kit | Words\n", "lost_age is not an age"},
		"eras and ages":     {"== p eras=ancient ages=stone_age\nplain kit | Words\n", "not both"},
		"bad setting":       {"== p mood=sad\nplain kit | Words\n", "mood=sad is not a setting"},
		"empty bank":        {"== words b\n", "is empty"},
		"nameless bank":     {"== words\nentry\n", "== words name"},
		"entry first":       {"plain kit | Words\n", "before the first heading"},
	}
	for name, tc := range cases {
		_, err := parseTextFile("exp_fail", tc.file)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want an error with %q", name, err, tc.want)
		}
	}
}

func TestTextFilesAreHeldToEachOther(t *testing.T) {
	if _, err := readCatalogText(validFolder()); err != nil {
		t.Fatalf("the minimal folder should load: %v", err)
	}
	cases := map[string]struct {
		change func(fstest.MapFS)
		want   string
	}{
		"a file missing": {func(f fstest.MapFS) { delete(f, "war_raid.txt") }, "war_raid.txt"},
		"a stray file": {func(f fstest.MapFS) {
			f["extra.txt"] = &fstest.MapFile{Data: []byte("== extra\nx\n")}
		}, "extra.txt is not one of the catalog files"},
		"a pool of another catalog": {func(f fstest.MapFS) {
			f["exp_fail.txt"] = &fstest.MapFile{Data: []byte("== war_raid_any\nplain kit | A sentence\n")}
		}, "does not start with exp_fail_"},
		"two pools with one set of row ids": {func(f fstest.MapFS) {
			f["exp_fail.txt"] = &fstest.MapFile{Data: []byte("== exp_fail_iron_age\nplain kit | A sentence\n== exp_fail_age_iron_age ages=iron_age\nplain kit | Another\n")}
		}, "has the same row ids as exp_fail_iron_age"},
		"a file with no pool": {func(f fstest.MapFS) {
			f["exp_fail.txt"] = &fstest.MapFile{Data: []byte("# nothing here\n")}
		}, "exp_fail.txt has no pools"},
		"a repeated bank": {func(f fstest.MapFS) {
			f["exp_fail.txt"] = &fstest.MapFile{Data: []byte("== words harb_shared\nx\n== exp_fail_any\nplain kit | A sentence\n")}
		}, "harb_shared is already in another file"},
		"a slot with no bank": {func(f fstest.MapFS) {
			f["exp_fail.txt"] = &fstest.MapFile{Data: []byte("== exp_fail_any\nplain kit slot=nowhere | A sentence about ~\n")}
		}, "slot=nowhere names no word list"},
	}
	for name, tc := range cases {
		folder := validFolder()
		tc.change(folder)
		_, err := readCatalogText(folder)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want an error with %q", name, err, tc.want)
		}
	}
}

func TestSentencesKeepTheirTagsAndPlaces(t *testing.T) {
	text := "== exp_fail_any eras=ancient,feudal\n" +
		"wry ledger kit tone=grim,wry kind=military need=subject,res slot=bank | A ~ for {subject}\n"
	file, err := parseTextFile("exp_fail", text)
	if err != nil {
		t.Fatal(err)
	}
	pool := file.pools[0]
	sentence := pool.sentences[0]
	if sentence.Reg != rWry || sentence.Form != fLedger || sentence.Topic != "kit" || sentence.Slot != "bank" {
		t.Errorf("tags read as %+v", sentence.skel)
	}
	if want := (needSubject | needRes); sentence.Needs != want {
		t.Errorf("needs %v, want %v", sentence.Needs, want)
	}
	if len(sentence.Tones) != 2 || len(sentence.Kinds) != 1 || sentence.Kinds[0] != "military" {
		t.Errorf("tones %v, kinds %v", sentence.Tones, sentence.Kinds)
	}
	if got := text[sentence.from:sentence.to]; got != "A ~ for {subject}" {
		t.Errorf("the offsets give %q", got)
	}
	if len(pool.eras) != 2 || pool.bucket != "any" {
		t.Errorf("eras %v, bucket %q", pool.eras, pool.bucket)
	}
}

func TestAPoolForOneAgeIsNamedForTheAge(t *testing.T) {
	file, err := parseTextFile("harb_warn", "== harb_warn_age_iron_age ages=iron_age\nplain kit | A line\n")
	if err != nil {
		t.Fatal(err)
	}
	pool := file.pools[0]
	if pool.bucket != "iron_age" {
		t.Errorf("bucket %q, want iron_age", pool.bucket)
	}
	if len(pool.eras) != 1 || pool.eras[0] != eraAncient {
		t.Errorf("eras %v, want the ancient era the iron age falls in", pool.eras)
	}
}
