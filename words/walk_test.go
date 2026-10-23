package words

import (
	"reflect"
	"testing"
)

// The types below stand for a package of definitions. Their ids are in the
// table for the length of a test.
type thing struct {
	Key   string
	Name  string
	Lines []string
	Tags  [2]string
	Names map[string]string
	Num   int
	Kids  []kid
	ByKey map[string]kid
	Count int
}

type kid struct {
	Key   string
	Title string
	Skip  string
}

func withIDs(t *testing.T, ids map[string]string) {
	t.Helper()
	saved := map[string]string{}
	for k, v := range ids {
		saved[k], FieldIDs[k] = FieldIDs[k], v
	}
	t.Cleanup(func() {
		for k := range ids {
			if saved[k] == "" {
				delete(FieldIDs, k)
			} else {
				FieldIDs[k] = saved[k]
			}
		}
	})
}

func sample() []thing {
	return []thing{
		{Key: "oak", Name: "Oak", Lines: []string{"a", "", "c"}, Tags: [2]string{"x", "y"},
			Names: map[string]string{"p": "P", "q": "Q"}, Num: 7,
			Kids:  []kid{{Key: "k1", Title: "One", Skip: "s"}, {Key: "k2", Title: "Two"}},
			ByKey: map[string]kid{"m": {Key: "km", Title: "Em"}}},
		{Key: "elm", Name: "Elm", Count: 3},
	}
}

func TestCellsNameEveryText(t *testing.T) {
	withIDs(t, map[string]string{
		"words.thing.Name":  "thing.{Key}.name",
		"words.thing.Lines": "thing.{Key}.line.{n}",
		"words.thing.Tags":  "thing.{Key}.tag.{n}",
		"words.thing.Names": "thing.{Key}.alias.{k}",
		"words.kid.Title":   "kid.{^Key}.{Key}.{sn}.{k}",
		"words.thing.Num":   "thing.{Key}.num.{Num}",
		"words.thing.Skip":  "thing.{decl}.nothing",
	})
	var got []string
	for _, c := range Cells(sample()) {
		got = append(got, c.ID+"="+c.Text)
	}
	want := []string{
		"thing.oak.name=Oak",
		"thing.oak.line.01=a",
		"thing.oak.line.03=c",
		"thing.oak.tag.01=x",
		"thing.oak.tag.02=y",
		"thing.oak.alias.p=P",
		"thing.oak.alias.q=Q",
		"kid.oak.k1.01.=One",
		"kid.oak.k2.02.=Two",
		"kid.oak.km.01.m=Em",
		"thing.elm.name=Elm",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("cells:\n got %q\nwant %q", got, want)
	}
}

func TestRewriteCopiesWhatItChanges(t *testing.T) {
	withIDs(t, map[string]string{
		"words.thing.Name":  "thing.{Key}.name",
		"words.thing.Lines": "thing.{Key}.line.{n}",
		"words.thing.Tags":  "thing.{Key}.tag.{n}",
		"words.thing.Names": "thing.{Key}.alias.{k}",
		"words.kid.Title":   "kid.{Key}.title",
	})
	in := sample()
	snapshot := sample()
	out, n := Rewrite(in, func(c Cell) (string, bool) {
		switch c.ID {
		case "thing.oak.name":
			return "Oak!", true
		case "thing.oak.line.03":
			return "C", true
		case "thing.oak.tag.02":
			return "Y", true
		case "thing.oak.alias.q":
			return "QQ", true
		case "kid.k2.title":
			return "Deux", true
		case "kid.km.title":
			return "Meme", true
		case "thing.elm.name":
			return "Elm", true // no change: not counted
		}
		return "", false
	})
	if n != 6 {
		t.Errorf("%d texts changed, want 6", n)
	}
	if !reflect.DeepEqual(in, snapshot) {
		t.Error("Rewrite wrote into the value it was given")
	}
	o := out[0]
	if o.Name != "Oak!" || o.Lines[2] != "C" || o.Lines[0] != "a" || o.Tags != [2]string{"x", "Y"} ||
		o.Names["q"] != "QQ" || o.Names["p"] != "P" || o.Kids[1].Title != "Deux" || o.Kids[0].Title != "One" ||
		o.ByKey["m"].Title != "Meme" || o.Num != 7 || o.Kids[0].Skip != "s" {
		t.Errorf("the rewritten thing is %+v", o)
	}
	if !reflect.DeepEqual(out[1], in[1]) {
		t.Error("a thing nothing changed in was altered")
	}
	// The slices and maps it shares with the input were copied, not written.
	if &o.Lines[0] == &in[0].Lines[0] || reflect.ValueOf(o.Names).Pointer() == reflect.ValueOf(in[0].Names).Pointer() {
		t.Error("the changed thing shares its list or map with the input")
	}
}

func TestRewriteLeavesAnUntouchedValueAlone(t *testing.T) {
	withIDs(t, map[string]string{"words.thing.Name": "thing.{Key}.name"})
	in := sample()
	out, n := Rewrite(in, func(Cell) (string, bool) { return "", false })
	if n != 0 || !reflect.DeepEqual(in, out) {
		t.Errorf("a rewrite that changes nothing changed %d texts", n)
	}
	if &out[0] != &in[0] {
		t.Error("a rewrite that changes nothing copied the list")
	}
}

func TestResolvable(t *testing.T) {
	for tmpl, want := range map[string]bool{
		"building.{Key}.name":       true,
		"mission.{Age}.place.{n}":   true,
		"badge_reward.{k}{^Key}.x":  true,
		"badge_rung.{decl}.{sn}":    false,
		"boon.{decl}.{sn}.line.{n}": false,
		"ui.{a0}.label":             false,
		"plain.id":                  true,
	} {
		if got := Resolvable(tmpl); got != want {
			t.Errorf("Resolvable(%q) = %v", tmpl, got)
		}
	}
}

func TestIDSafe(t *testing.T) {
	for in, want := range map[string]string{
		"lumber_mill": "lumber_mill",
		"a b":         "a_b",
		" a  b ":      "a_b",
		"swift.stone": "swift_stone",
		"x-y":         "x-y",
	} {
		if got := IDSafe(in); got != want {
			t.Errorf("IDSafe(%q) = %q, want %q", in, got, want)
		}
	}
}
