package game

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// GetState is a snapshot: the UI and the smoke suite hold on to them while
// the game runs on, and nothing a caller does to one may reach the engine.
// It used to hand out History, EpochEventHistory (edited in place when a
// catastrophe resolves) and CatastropheHistory by reference. These tests
// walk every map, slice and pointer in a snapshot generically, so a field
// added later that aliases engine state fails here too.

// richSnapshotEngine is a mid-game engine with every history populated: the
// determinism scenario's Medieval Age economy, a few hundred ticks of
// samples, and a resolved catastrophe on the record.
func richSnapshotEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge, sc := determinismSetup(t, 11)
	sc.play(ge, 0, 400)
	ge.mu.Lock()
	ge.pendingCatastrophe = ""
	ge.mu.Unlock()
	if err := ge.ForceCatastropheForTest(); err != nil {
		t.Fatalf("force catastrophe: %v", err)
	}
	if err := ge.Endure(); err != nil {
		t.Fatalf("endure: %v", err)
	}
	if err := ge.ForceCatastropheForTest(); err != nil {
		t.Fatalf("force second catastrophe: %v", err)
	}
	// A resolved harbinger thread on the record (its Chain is a slice).
	ge.mu.Lock()
	ge.harbingerHistory = append(ge.harbingerHistory, HarbingerRecord{
		Age: "medieval_age", Name: "test", Chain: []string{"iron_age", "medieval_age"},
		EpochKey: "iron_era", Outcome: HarbingerOutcomeSpared, Tick: ge.tick,
	})
	ge.mu.Unlock()
	st := ge.GetState()
	for name, n := range map[string]int{
		"History.Samples":    len(st.History.Samples),
		"EpochEventHistory":  len(st.EpochEventHistory),
		"CatastropheHistory": len(st.CatastropheHistory),
		"HarbingerHistory":   len(st.HarbingerHistory),
		"Log":                len(st.Log),
	} {
		if n == 0 {
			t.Fatalf("setup: %s is empty, so the test would not cover it", name)
		}
	}
	return ge
}

func TestGetState_MutatingSnapshotLeavesEngineAlone(t *testing.T) {
	isolateAccountDir(t)
	ge := richSnapshotEngine(t)
	snap := ge.GetState()
	want := snapCopy(snap)
	poison(reflect.ValueOf(&snap).Elem(), map[uintptr]bool{})
	if d := snapDiff(want, ge.GetState()); d != "" {
		t.Errorf("mutating a GetState snapshot changed the engine: %s", d)
	}
}

func TestGetState_HeldSnapshotSurvivesTicks(t *testing.T) {
	isolateAccountDir(t)
	ge := richSnapshotEngine(t)
	snap := ge.GetState()
	want := snapCopy(snap)
	// Resolve the pending catastrophe (setCatastropheOutcome edits the
	// history record in place), then play on so every history grows.
	if err := ge.Succumb(); err != nil {
		t.Fatalf("succumb: %v", err)
	}
	ge.StepTicks(300)
	if d := snapDiff(want, snap); d != "" {
		t.Errorf("a held snapshot changed as the game ran on: %s", d)
	}
}

// snapCopy deep-copies v (exported fields only; the diff skips the rest).
func snapCopy[T any](v T) T {
	src := reflect.ValueOf(&v).Elem()
	dst := reflect.New(src.Type()).Elem()
	copyVal(dst, src)
	return dst.Interface().(T)
}

var timeType = reflect.TypeOf(time.Time{})

func copyVal(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Ptr:
		if !src.IsNil() {
			p := reflect.New(src.Type().Elem())
			copyVal(p.Elem(), src.Elem())
			dst.Set(p)
		}
	case reflect.Interface:
		if !src.IsNil() {
			c := reflect.New(src.Elem().Type()).Elem()
			copyVal(c, src.Elem())
			dst.Set(c)
		}
	case reflect.Struct:
		if src.Type() == timeType {
			dst.Set(src)
			return
		}
		for i := 0; i < src.NumField(); i++ {
			if src.Type().Field(i).IsExported() {
				copyVal(dst.Field(i), src.Field(i))
			}
		}
	case reflect.Slice:
		if !src.IsNil() {
			s := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
			for i := 0; i < src.Len(); i++ {
				copyVal(s.Index(i), src.Index(i))
			}
			dst.Set(s)
		}
	case reflect.Array:
		for i := 0; i < src.Len(); i++ {
			copyVal(dst.Index(i), src.Index(i))
		}
	case reflect.Map:
		if !src.IsNil() {
			m := reflect.MakeMapWithSize(src.Type(), src.Len())
			for _, k := range src.MapKeys() {
				v := reflect.New(src.Type().Elem()).Elem()
				copyVal(v, src.MapIndex(k))
				m.SetMapIndex(k, v)
			}
			dst.Set(m)
		}
	case reflect.Func, reflect.Chan:
	default:
		dst.Set(src)
	}
}

// poison changes every value reachable from v in place: every leaf is
// altered, every map gets its values rewritten and a new key.
func poison(v reflect.Value, seen map[uintptr]bool) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || seen[v.Pointer()] {
			return
		}
		seen[v.Pointer()] = true
		poison(v.Elem(), seen)
	case reflect.Interface:
		if v.IsNil() || !v.CanSet() {
			return
		}
		c := reflect.New(v.Elem().Type()).Elem()
		c.Set(v.Elem())
		poison(c, seen)
		v.Set(c)
	case reflect.Struct:
		if v.Type() == timeType {
			if v.CanSet() {
				v.Set(reflect.ValueOf(time.Unix(1, 0)))
			}
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				poison(v.Field(i), seen)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			poison(v.Index(i), seen)
		}
	case reflect.Map:
		if v.IsNil() {
			return
		}
		for _, k := range v.MapKeys() {
			val := reflect.New(v.Type().Elem()).Elem()
			val.Set(v.MapIndex(k))
			poison(val, seen)
			v.SetMapIndex(k, val)
		}
		if v.Type().Key().Kind() == reflect.String {
			v.SetMapIndex(reflect.ValueOf("~poison~").Convert(v.Type().Key()), reflect.Zero(v.Type().Elem()))
		}
	case reflect.String:
		if v.CanSet() {
			v.SetString(v.String() + "~")
		}
	case reflect.Bool:
		if v.CanSet() {
			v.SetBool(!v.Bool())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.CanSet() {
			v.SetInt(v.Int() + 1)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.CanSet() {
			v.SetUint(v.Uint() + 1)
		}
	case reflect.Float32, reflect.Float64:
		if v.CanSet() {
			v.SetFloat(v.Float() + 1)
		}
	}
}

// snapDiff returns the path of the first exported difference between two
// snapshots, or "". Wall-clock fields are skipped.
func snapDiff(a, b GameState) string {
	return diffVal(reflect.ValueOf(a), reflect.ValueOf(b), "GameState")
}

func diffVal(a, b reflect.Value, path string) string {
	switch path {
	case "GameState.Stats.PlayTime", "GameState.SaveExists":
		return ""
	}
	if a.Kind() != b.Kind() {
		return path + ": kind differs"
	}
	switch a.Kind() {
	case reflect.Ptr, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			if a.IsNil() != b.IsNil() {
				return path + ": nil vs non-nil"
			}
			return ""
		}
		return diffVal(a.Elem(), b.Elem(), path)
	case reflect.Struct:
		if a.Type() == timeType {
			if !a.Interface().(time.Time).Equal(b.Interface().(time.Time)) {
				return path
			}
			return ""
		}
		for i := 0; i < a.NumField(); i++ {
			if f := a.Type().Field(i); f.IsExported() {
				if d := diffVal(a.Field(i), b.Field(i), path+"."+f.Name); d != "" {
					return d
				}
			}
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: length %d vs %d", path, a.Len(), b.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if d := diffVal(a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i)); d != "" {
				return d
			}
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: %d vs %d entries", path, a.Len(), b.Len())
		}
		for _, k := range a.MapKeys() {
			bv := b.MapIndex(k)
			if !bv.IsValid() {
				return fmt.Sprintf("%s[%v]: missing", path, k)
			}
			if d := diffVal(a.MapIndex(k), bv, fmt.Sprintf("%s[%v]", path, k)); d != "" {
				return d
			}
		}
	case reflect.Func, reflect.Chan:
	case reflect.Float32, reflect.Float64:
		if x, y := a.Float(), b.Float(); x != y && !(x != x && y != y) {
			return fmt.Sprintf("%s: %v vs %v", path, x, y)
		}
	default:
		if a.Interface() != b.Interface() {
			return fmt.Sprintf("%s: %v vs %v", path, a.Interface(), b.Interface())
		}
	}
	return ""
}
