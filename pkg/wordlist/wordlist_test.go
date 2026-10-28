package wordlist

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseReadsListsAndSkipsCommentsAndBlanks(t *testing.T) {
	text := "# header\n\n== first one\n  a b  \n# aside\nc\n== second\r\nd\r\n"
	lists, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 2 {
		t.Fatalf("got %d lists, want 2", len(lists))
	}
	if got := lists[0].Head; !reflect.DeepEqual(got, []string{"first", "one"}) || lists[0].Name() != "first" || lists[0].Line != 3 {
		t.Errorf("first list: head %v at line %d", got, lists[0].Line)
	}
	var got []string
	for _, l := range lists {
		for _, e := range l.Entries {
			got = append(got, e.Text)
		}
	}
	if want := []string{"a b", "c", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("entries %q, want %q", got, want)
	}
}

func TestEntryPositionsPointAtTheText(t *testing.T) {
	text := "== x\n  padded entry  \r\nnext\n"
	lists, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range lists[0].Entries {
		if text[e.Start:e.End] != e.Text {
			t.Errorf("line %d: offsets give %q, entry is %q", e.Line, text[e.Start:e.End], e.Text)
		}
	}
	if lists[0].Entries[1].Line != 3 {
		t.Errorf("second entry is on line %d, want 3", lists[0].Entries[1].Line)
	}
}

func TestParseRefusesWhatIsNotAList(t *testing.T) {
	for name, tc := range map[string]struct{ text, want string }{
		"entry first":    {"loose line\n== a\n", "line 1: an entry before the first heading"},
		"nameless head":  {"== a\nx\n==\n", "line 3: a heading with no name"},
		"repeated heads": {"== a b\nx\n== a b\ny\n", "line 3: the list \"a b\" is already in this file"},
	} {
		_, err := Parse(tc.text)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want an error with %q", name, err, tc.want)
		}
	}
}
