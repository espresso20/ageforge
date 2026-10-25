package words

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
)

// This file reads the ids off a running ruleset. The export finds a row by
// parsing the source. This finds the same row on the value the game built from
// that source, with the same FieldIDs table, so a text the overlay replaces is
// the text the export gave that id. A test in cmd/words holds the two together
// (TestRuntimeIdsAreTheExportsOwn).

// Cell is one string of a definition that a player reads.
type Cell struct {
	ID    string // the row id
	Field string // "config.BuildingDef.Name"
	Text  string
	Owner any // the struct the text is a field of, as it stands
}

// Rewrite walks the definitions under v and offers every text that has a row id
// to fn, which answers the text to put there or false to leave it. It returns v
// with the answers in, and how many texts it changed (an answer equal to the
// text does not count).
//
// Nothing v shares with the caller is written to: a struct, slice or map with a
// text to change is copied first, so the tables of package config that a
// definition was built from stay as they were. v may be a struct, a slice, a map
// with string keys or a pointer to one.
//
// A text that is empty is never offered: an empty description or hint is the
// game saying there is none, and a script has none to swap in.
func Rewrite[T any](v T, fn func(c Cell) (string, bool)) (T, int) {
	w := &walker{fn: fn}
	out, _ := w.walk(reflect.ValueOf(&v).Elem(), pos{idx: -1}, 0)
	return out.Interface().(T), w.changed
}

// Cells lists every text under v that has a row id, in the order Rewrite walks
// them. A map of structs is walked in map order, which varies.
func Cells[T any](v T) []Cell {
	var cells []Cell
	Rewrite(v, func(c Cell) (string, bool) {
		cells = append(cells, c)
		return "", false
	})
	return cells
}

// Resolvable reports whether an id template can be filled from a running
// ruleset: it names no slot that only the source has ({decl}, {a0} to {a9}).
func Resolvable(tmpl string) bool {
	for _, m := range TemplateRe.FindAllStringSubmatch(tmpl, -1) {
		if n := m[3]; n == "decl" || (len(n) == 2 && n[0] == 'a' && n[1] >= '0' && n[1] <= '9') {
			return false
		}
	}
	return true
}

// pos is where a value sits in the collection around it.
type pos struct {
	idx   int    // place in a list, or -1
	key   string // key in a map
	isMap bool
}

// frame is a struct on the way down.
type frame struct {
	val reflect.Value
	typ string // "config.BuildingDef"
	at  pos
}

type walker struct {
	fn      func(c Cell) (string, bool)
	frames  []frame
	keys    []string // map keys on the way down, outermost first
	changed int
}

const maxDepth = 16

// walk returns v with the answers in and whether anything changed. A value that
// nothing changed in is returned as it came.
func (w *walker) walk(v reflect.Value, at pos, depth int) (reflect.Value, bool) {
	if depth > maxDepth {
		return v, false
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return v, false
		}
		inner, ok := w.walk(v.Elem(), at, depth+1)
		if !ok {
			return v, false
		}
		p := reflect.New(v.Type().Elem())
		p.Elem().Set(inner)
		return p, true
	case reflect.Struct:
		return w.walkStruct(v, at, depth)
	case reflect.Slice:
		if v.IsNil() || !holdsStructs(v.Type().Elem(), 0) {
			return v, false
		}
		var out reflect.Value
		for i := 0; i < v.Len(); i++ {
			e, ok := w.walk(v.Index(i), pos{idx: i}, depth+1)
			if !ok {
				continue
			}
			if !out.IsValid() {
				out = reflect.MakeSlice(v.Type(), v.Len(), v.Len())
				reflect.Copy(out, v)
			}
			out.Index(i).Set(e)
		}
		if !out.IsValid() {
			return v, false
		}
		return out, true
	case reflect.Map:
		if v.IsNil() || v.Type().Key().Kind() != reflect.String || !holdsStructs(v.Type().Elem(), 0) {
			return v, false
		}
		var out reflect.Value
		it := v.MapRange()
		for it.Next() {
			k := it.Key().String()
			w.keys = append(w.keys, k)
			e, ok := w.walk(it.Value(), pos{idx: -1, key: k, isMap: true}, depth+1)
			w.keys = w.keys[:len(w.keys)-1]
			if !ok {
				continue
			}
			if !out.IsValid() {
				out = reflect.MakeMapWithSize(v.Type(), v.Len())
				for it2 := v.MapRange(); it2.Next(); {
					out.SetMapIndex(it2.Key(), it2.Value())
				}
			}
			out.SetMapIndex(it.Key(), e)
		}
		if !out.IsValid() {
			return v, false
		}
		return out, true
	}
	return v, false
}

// holdsStructs reports whether a value of type t can have a struct in it,
// so a slice of numbers or a map of strings is not walked element by
// element.
func holdsStructs(t reflect.Type, depth int) bool {
	if depth > maxDepth {
		return false
	}
	switch t.Kind() {
	case reflect.Struct:
		return true
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return holdsStructs(t.Elem(), depth+1)
	case reflect.Map:
		return t.Key().Kind() == reflect.String && holdsStructs(t.Elem(), depth+1)
	}
	return false
}

// walkStruct offers the texts of a struct and walks into the fields that
// hold more structs.
func (w *walker) walkStruct(v reflect.Value, at pos, depth int) (reflect.Value, bool) {
	t := v.Type()
	if t.Name() == "" || t.PkgPath() == "" {
		return v, false
	}
	typ := path.Base(t.PkgPath()) + "." + t.Name()
	w.frames = append(w.frames, frame{val: v, typ: typ, at: at})
	defer func() { w.frames = w.frames[:len(w.frames)-1] }()

	var out reflect.Value
	set := func(i int, nf reflect.Value) {
		if !out.IsValid() {
			out = reflect.New(t).Elem()
			out.Set(v)
		}
		out.Field(i).Set(nf)
	}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		f := v.Field(i)
		tmpl := FieldIDs[typ+"."+sf.Name]
		switch {
		case tmpl != "" && Resolvable(tmpl) && f.Kind() == reflect.String:
			if s, ok := w.offer(typ+"."+sf.Name, tmpl, f.String(), pos{idx: -1}); ok {
				nf := reflect.New(sf.Type).Elem()
				nf.SetString(s)
				set(i, nf)
			}
		case tmpl != "" && Resolvable(tmpl) && f.Kind() == reflect.Slice && f.Type().Elem().Kind() == reflect.String:
			var nf reflect.Value
			for j := 0; j < f.Len(); j++ {
				s, ok := w.offer(typ+"."+sf.Name, tmpl, f.Index(j).String(), pos{idx: j})
				if !ok {
					continue
				}
				if !nf.IsValid() {
					nf = reflect.MakeSlice(f.Type(), f.Len(), f.Len())
					reflect.Copy(nf, f)
				}
				nf.Index(j).SetString(s)
			}
			if nf.IsValid() {
				set(i, nf)
			}
		case tmpl != "" && Resolvable(tmpl) && f.Kind() == reflect.Array && f.Type().Elem().Kind() == reflect.String:
			var nf reflect.Value
			for j := 0; j < f.Len(); j++ {
				s, ok := w.offer(typ+"."+sf.Name, tmpl, f.Index(j).String(), pos{idx: j})
				if !ok {
					continue
				}
				if !nf.IsValid() {
					nf = reflect.New(f.Type()).Elem()
					nf.Set(f)
				}
				nf.Index(j).SetString(s)
			}
			if nf.IsValid() {
				set(i, nf)
			}
		case tmpl != "" && Resolvable(tmpl) && f.Kind() == reflect.Map && f.Type().Key().Kind() == reflect.String && f.Type().Elem().Kind() == reflect.String:
			var nf reflect.Value
			for _, k := range sortedKeys(f) {
				w.keys = append(w.keys, k.String())
				s, ok := w.offer(typ+"."+sf.Name, tmpl, f.MapIndex(k).String(), pos{idx: -1, key: k.String(), isMap: true})
				w.keys = w.keys[:len(w.keys)-1]
				if !ok {
					continue
				}
				if !nf.IsValid() {
					nf = reflect.MakeMapWithSize(f.Type(), f.Len())
					for it := f.MapRange(); it.Next(); {
						nf.SetMapIndex(it.Key(), it.Value())
					}
				}
				nf.SetMapIndex(k, reflect.ValueOf(s).Convert(f.Type().Elem()))
			}
			if nf.IsValid() {
				set(i, nf)
			}
		default:
			if nf, ok := w.walk(f, pos{idx: -1}, depth+1); ok {
				set(i, nf)
			}
		}
	}
	if !out.IsValid() {
		return v, false
	}
	return out, true
}

// sortedKeys returns the keys of a string-keyed map in order, so the texts of a
// map field are offered in the same order every time.
func sortedKeys(m reflect.Value) []reflect.Value {
	keys := m.MapKeys()
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j].String() < keys[j-1].String(); j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// offer names one text, hands it to fn and says what to put there.
func (w *walker) offer(field, tmpl, text string, at pos) (string, bool) {
	if text == "" {
		return "", false
	}
	id := Expand(tmpl, func(name string, outer bool) (string, bool) { return w.slot(name, outer, at) }, true, nil)
	var owner any
	if top := len(w.frames) - 1; top >= 0 && w.frames[top].val.CanInterface() {
		owner = w.frames[top].val.Interface()
	}
	s, ok := w.fn(Cell{ID: id, Field: field, Text: text, Owner: owner})
	if !ok || s == text {
		return "", false
	}
	w.changed++
	return s, true
}

// slot fills one slot of an id template, as the export does from the source.
func (w *walker) slot(name string, outer bool, at pos) (string, bool) {
	top := len(w.frames) - 1
	switch {
	case outer:
		if top < 1 {
			return "", true
		}
		return fieldText(w.frames[top-1].val, name), true
	case name == "k" || name == "k2":
		// The keys the text sits under, innermost first.
		i := 0
		if name == "k2" {
			i = 1
		}
		if n := len(w.keys); i < n {
			return w.keys[n-1-i], true
		}
		return "", true
	case name == "n":
		if at.idx >= 0 {
			return fmt.Sprintf("%02d", at.idx+1), true
		}
		return "01", true
	case name == "sn":
		if top >= 0 && w.frames[top].at.idx >= 0 {
			return fmt.Sprintf("%02d", w.frames[top].at.idx+1), true
		}
		return "01", true
	}
	if top >= 0 {
		v := w.frames[top].val
		if _, ok := v.Type().FieldByName(name); ok {
			return fieldText(v, name), true
		}
		if name[0] >= 'A' && name[0] <= 'Z' || name == "key" || name == "from" || name == "title" || name == "name" {
			return "", true
		}
	}
	return "", false
}

// fieldText is a field of a struct value as the text of an id: a string as it
// is, a signed integer in digits, anything else "".
func fieldText(v reflect.Value, name string) string {
	f := v.FieldByName(name)
	if !f.IsValid() {
		return ""
	}
	switch f.Kind() {
	case reflect.String:
		return f.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(f.Int(), 10)
	}
	return ""
}
