package smoke

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"
)

// firstDiff walks a and b (same type) and returns the path of the first
// field that differs and a description of both values, or "" when they
// match. Floats compare bit for bit; maps walk in sorted key order; nil and
// empty slices and maps are equal; time.Time compares with Equal; unexported
// fields, funcs and channels are skipped. skip reports paths to ignore.
func firstDiff(a, b interface{}, skip func(path string) bool) string {
	return diffValue(reflect.ValueOf(a), reflect.ValueOf(b), "", skip)
}

var timeType = reflect.TypeOf(time.Time{})

func diffValue(a, b reflect.Value, path string, skip func(string) bool) string {
	if skip != nil && path != "" && skip(path) {
		return ""
	}
	if !a.IsValid() || !b.IsValid() {
		if a.IsValid() != b.IsValid() {
			return fmt.Sprintf("%s: one side missing", orRoot(path))
		}
		return ""
	}
	if a.Type() != b.Type() {
		return fmt.Sprintf("%s: type %s vs %s", orRoot(path), a.Type(), b.Type())
	}
	if a.Type() == timeType && a.CanInterface() {
		ta, tb := a.Interface().(time.Time), b.Interface().(time.Time)
		if !ta.Equal(tb) {
			return fmt.Sprintf("%s: %s vs %s", orRoot(path), ta, tb)
		}
		return ""
	}
	switch a.Kind() {
	case reflect.Ptr, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			if a.IsNil() != b.IsNil() {
				return fmt.Sprintf("%s: nil vs non-nil (%s vs %s)", orRoot(path), short(a), short(b))
			}
			return ""
		}
		return diffValue(a.Elem(), b.Elem(), path, skip)
	case reflect.Struct:
		t := a.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" { // unexported
				continue
			}
			if d := diffValue(a.Field(i), b.Field(i), join(path, f.Name), skip); d != "" {
				return d
			}
		}
		return ""
	case reflect.Map:
		ka, kb := mapKeys(a), mapKeys(b)
		sa, sb := keyStrings(ka), keyStrings(kb)
		if strings.Join(sa, "\x00") != strings.Join(sb, "\x00") {
			return fmt.Sprintf("%s: keys differ: only in first %v, only in second %v", orRoot(path), minus(sa, sb), minus(sb, sa))
		}
		for i, k := range ka {
			if d := diffValue(a.MapIndex(k), b.MapIndex(kb[i]), fmt.Sprintf("%s[%s]", path, sa[i]), skip); d != "" {
				return d
			}
		}
		return ""
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: length %d vs %d", orRoot(path), a.Len(), b.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if d := diffValue(a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i), skip); d != "" {
				return d
			}
		}
		return ""
	case reflect.Float32, reflect.Float64:
		if math.Float64bits(a.Float()) != math.Float64bits(b.Float()) {
			return fmt.Sprintf("%s: %v vs %v (differ by %g)", orRoot(path), a.Float(), b.Float(), a.Float()-b.Float())
		}
		return ""
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return ""
	default:
		if a.CanInterface() && b.CanInterface() && !reflect.DeepEqual(a.Interface(), b.Interface()) {
			return fmt.Sprintf("%s: %v vs %v", orRoot(path), a.Interface(), b.Interface())
		}
		return ""
	}
}

func mapKeys(m reflect.Value) []reflect.Value {
	keys := m.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j]) })
	return keys
}

func keyStrings(keys []reflect.Value) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = fmt.Sprint(k)
	}
	return out
}

func minus(a, b []string) []string {
	in := map[string]bool{}
	for _, s := range b {
		in[s] = true
	}
	var out []string
	for _, s := range a {
		if !in[s] {
			out = append(out, s)
		}
	}
	return out
}

func join(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

func orRoot(path string) string {
	if path == "" {
		return "(root)"
	}
	return path
}

func short(v reflect.Value) string {
	if v.IsNil() {
		return "nil"
	}
	s := fmt.Sprintf("%+v", v.Elem())
	if len(s) > 120 {
		s = s[:120] + "..."
	}
	return s
}

// jsonTree decodes a JSON document keeping numbers as their exact text, so
// two trees compare float for float.
func jsonTree(data []byte) (interface{}, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	var v interface{}
	err := dec.Decode(&v)
	return v, err
}
