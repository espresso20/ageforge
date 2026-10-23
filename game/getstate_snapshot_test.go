package game

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
)

// dumpSnapshot is dumpState of a GameState with the one thing the engine
// leaves in map order put in key order: the active trade routes. The overlay
// sorts its own copy, so two snapshots of one town list them in any order.
func dumpSnapshot(st GameState) string {
	st.Trade.ActiveRoutes = append([]ActiveRouteInfo(nil), st.Trade.ActiveRoutes...)
	sort.Slice(st.Trade.ActiveRoutes, func(i, j int) bool { return st.Trade.ActiveRoutes[i].Key < st.Trade.ActiveRoutes[j].Key })
	return dumpState(st)
}

// dumpState renders v as text, deeply and deterministically: maps in key
// order, pointers followed, unexported fields included, funcs left out. Two
// snapshots are the same snapshot exactly when their dumps are equal, so a
// failure can name the first line that differs.
func dumpState(v any) string {
	var sb strings.Builder
	dumpValue(&sb, reflect.ValueOf(v), map[uintptr]bool{}, 0)
	return sb.String()
}

func dumpValue(sb *strings.Builder, v reflect.Value, seen map[uintptr]bool, depth int) {
	if depth > 40 {
		sb.WriteString("<deep>")
		return
	}
	switch v.Kind() {
	case reflect.Invalid:
		sb.WriteString("<invalid>")
	case reflect.Bool:
		fmt.Fprintf(sb, "%t", v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprintf(sb, "%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprintf(sb, "%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		fmt.Fprintf(sb, "%v", v.Float())
	case reflect.String:
		fmt.Fprintf(sb, "%q", v.String())
	case reflect.Ptr:
		if v.IsNil() {
			sb.WriteString("nil")
			return
		}
		if seen[v.Pointer()] {
			sb.WriteString("<cycle>")
			return
		}
		seen[v.Pointer()] = true
		sb.WriteString("&")
		dumpValue(sb, v.Elem(), seen, depth+1)
		delete(seen, v.Pointer())
	case reflect.Interface:
		if v.IsNil() {
			sb.WriteString("nil")
			return
		}
		dumpValue(sb, v.Elem(), seen, depth+1)
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) { // wall clock: only whether it is set can repeat
			fmt.Fprintf(sb, "time(set=%t)", v.FieldByName("wall").Uint() != 0 || v.FieldByName("ext").Int() != 0)
			return
		}
		sb.WriteString(v.Type().String() + "{")
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Name
			if name == "Rules" { // the shared ruleset, not part of the snapshot
				continue
			}
			if name == "PlayTime" { // wall clock: differs between any two calls
				continue
			}
			sb.WriteString("\n" + strings.Repeat(" ", depth) + name + ": ")
			dumpValue(sb, v.Field(i), seen, depth+1)
		}
		sb.WriteString("}")
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			sb.WriteString("nil")
			return
		}
		sb.WriteString("[")
		for i := 0; i < v.Len(); i++ {
			sb.WriteString("\n" + strings.Repeat(" ", depth) + fmt.Sprint(i) + ": ")
			dumpValue(sb, v.Index(i), seen, depth+1)
		}
		sb.WriteString("]")
	case reflect.Map:
		if v.IsNil() {
			sb.WriteString("nil")
			return
		}
		type row struct{ key, val string }
		var rows []row
		for _, k := range v.MapKeys() {
			var kb, vb strings.Builder
			dumpValue(&kb, k, seen, depth+1)
			dumpValue(&vb, v.MapIndex(k), seen, depth+1)
			rows = append(rows, row{kb.String(), vb.String()})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].key < rows[j].key })
		sb.WriteString("map[")
		for _, r := range rows {
			sb.WriteString("\n" + strings.Repeat(" ", depth) + r.key + ": " + r.val)
		}
		sb.WriteString("]")
	default: // funcs and channels carry no state a caller reads
		sb.WriteString("<" + v.Kind().String() + ">")
	}
}

// firstDifference names the first line two dumps part at, with the field it
// sits under, so a failing comparison points at what changed.
func firstDifference(a, b string) string {
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	field := ""
	for i := 0; i < len(la) && i < len(lb); i++ {
		if len(la[i]) > 0 && la[i][0] != ' ' {
			field = la[i]
		}
		if la[i] != lb[i] {
			return fmt.Sprintf("line %d under %q:\n  first:  %.200s\n  second: %.200s", i, field, la[i], lb[i])
		}
	}
	return fmt.Sprintf("dumps differ in length (%d vs %d lines)", len(la), len(lb))
}

// busyUp gives a town something in every queue the snapshot reports: a tech
// in progress, buildings under construction, a mission board posted. Anything
// the town cannot do is skipped; a snapshot of a quiet town is still a case.
func busyUp(ge *GameEngine) {
	ge.EnsureMissionBoard()
	ge.mu.RLock()
	var techs, builds []string
	for key := range ge.Research.defs {
		techs = append(techs, key)
	}
	builds = append(builds, ge.Buildings.order...)
	ge.mu.RUnlock()
	sort.Strings(techs)
	for _, key := range techs {
		if ge.StartResearch(key) == nil {
			break
		}
	}
	queued := 0
	for _, key := range builds {
		if queued < 4 && ge.BuildBuilding(key) == nil {
			queued++
		}
	}
}

// snapshotCases are real towns at three stages: a typical early, middle and
// late town (every building and tech the age holds, wonders behind them),
// the late-game bench town with its trade routes running, and that town held
// by an account. A few ticks are played first, and the towns are busied up,
// so queues, events, research and logs hold something.
func snapshotCases(tb testing.TB) map[string]*GameEngine {
	tb.Helper()
	cases := map[string]*GameEngine{}
	for _, age := range []string{"bronze_age", "medieval_age", "space_age"} {
		ge := newTruthEngine(age, truthTypical)
		busyUp(ge)
		for i := 0; i < 40; i++ {
			ge.doTick()
		}
		cases[age] = ge
	}
	late := newLateGameEngine(tb)
	busyUp(late)
	for i := 0; i < 40; i++ {
		late.doTick()
	}
	cases["late_bench"] = late
	held := newStaffedLateGameEngine(tb, true)
	busyUp(held)
	for i := 0; i < 40; i++ {
		held.doTick()
	}
	cases["late_account"] = held
	return cases
}

// TestGetStateRepeats: two snapshots taken with nothing between them are the
// same snapshot. A cache that leaked a tick's work into the next call, or
// left a field half-built, shows here.
func TestGetStateRepeats(t *testing.T) {
	for name, ge := range snapshotCases(t) {
		first := dumpSnapshot(ge.GetState())
		second := dumpSnapshot(ge.GetState())
		if first != second {
			t.Errorf("%s: two snapshots in a row differ: %s", name, firstDifference(first, second))
		}
	}
}

// TestGetStateOwnsItsMemory: a caller may edit the snapshot it was handed
// (the dashboard sorts and trims what it gets) without the engine or the
// next snapshot seeing it. Every map, slice and pointer in the snapshot is
// wiped, then the engine is asked again.
func TestGetStateOwnsItsMemory(t *testing.T) {
	for name, ge := range snapshotCases(t) {
		want := dumpSnapshot(ge.GetState())
		st := ge.GetState()
		scribble(reflect.ValueOf(&st).Elem(), 0)
		got := dumpSnapshot(ge.GetState())
		if want != got {
			t.Errorf("%s: editing a snapshot changed the next one: %s", name, firstDifference(want, got))
		}
	}
}

// scribble wipes everything under v that a caller could edit in place:
// map entries are deleted, slice elements and fields zeroed.
func scribble(v reflect.Value, depth int) {
	if depth > 40 {
		return
	}
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			scribble(v.Elem(), depth+1)
		}
	case reflect.Interface:
		// an interface holds a copy; nothing behind it is settable
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == "Rules" {
				continue
			}
			// The badge list and its summary are shared between callers by design
			// (Account.badgeViews: "read it, never write to it").
			if name := v.Type().Field(i).Name; v.Type() == reflect.TypeOf(AccountStatsView{}) && (name == "Badges" || name == "BadgeSummary") {
				continue
			}
			if f := v.Field(i); f.CanSet() {
				scribble(f, depth+1)
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			scribble(v.Index(i), depth+1)
			v.Index(i).Set(reflect.Zero(v.Type().Elem()))
		}
	case reflect.Map:
		if v.IsNil() {
			return
		}
		for _, k := range v.MapKeys() {
			elem := v.MapIndex(k)
			switch elem.Kind() {
			case reflect.Ptr, reflect.Map, reflect.Slice:
				scribble(elem, depth+1)
			case reflect.Struct:
				cp := reflect.New(elem.Type()).Elem()
				cp.Set(elem)
				scribble(cp, depth+1)
			}
			v.SetMapIndex(k, reflect.Value{})
		}
	}
}

// TestSliceArena: copies out of the arena match slices.Clone (nil stays nil,
// empty stays empty), are disjoint, and are capped so an append by the caller
// cannot reach the copy beside it.
func TestSliceArena(t *testing.T) {
	srcs := [][]string{{"a", "b"}, nil, {}, {"c"}, {"d", "e", "f"}}
	var arena sliceArena[string]
	for _, s := range srcs {
		arena.reserve(s)
	}
	var got [][]string
	for _, s := range srcs {
		got = append(got, arena.clone(s))
	}
	for i, s := range srcs {
		if (s == nil) != (got[i] == nil) || !slices.Equal(s, got[i]) {
			t.Errorf("clone %d = %#v, want %#v", i, got[i], s)
		}
		if len(got[i]) > 0 && cap(got[i]) != len(got[i]) {
			t.Errorf("clone %d has cap %d for len %d", i, cap(got[i]), len(got[i]))
		}
	}
	got[0] = append(got[0], "x") // must not overwrite the next copy
	got[0][0] = "changed"
	if got[3][0] != "c" || srcs[0][0] != "a" {
		t.Errorf("an edit reached a neighbour or the source: %v %v", got[3], srcs[0])
	}
	// More than counted: still a correct copy.
	var short sliceArena[string]
	short.reserve([]string{"a"})
	if a, b := short.clone([]string{"a"}), short.clone([]string{"b", "c"}); a[0] != "a" || len(b) != 2 || b[1] != "c" {
		t.Errorf("short count: %v %v", a, b)
	}
	// take: room to append into, disjoint from its neighbours.
	var rows sliceArena[int]
	rows.want = 5
	x, y := rows.take(2), rows.take(3)
	x = append(x, 1, 2)
	y = append(y, 3, 4, 5)
	x = append(x, 9) // past its room: moves, leaves y alone
	if !slices.Equal(x, []int{1, 2, 9}) || !slices.Equal(y, []int{3, 4, 5}) {
		t.Errorf("take: %v %v", x, y)
	}
	if rows.take(0) != nil {
		t.Error("take(0) should be nil")
	}
}
