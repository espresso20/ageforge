package smoke

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Performance budgets. They are generous on purpose (shared CI runners are
// slow and noisy); the report prints actuals so drift shows before a budget
// breaks.
const (
	TickBudget     = 250 * time.Microsecond
	GetStateBudget = time.Millisecond
	// A long run may grow the heap to at most HeapGrowthFactor x its early
	// size plus HeapGrowthSlack before it counts as unbounded growth.
	HeapGrowthFactor = 2.0
	HeapGrowthSlack  = 64 << 20
	// A single slice or map in GameState over CollectionFail elements is a
	// leak; one that grows at every sample past CollectionWarn is flagged.
	CollectionFail = 50000
	CollectionWarn = 2000
)

var benchLine = regexp.MustCompile(`^(Benchmark\w+?)(?:-\d+)?\s+\d+\s+([\d.]+) ns/op(?:\s+(\d+) B/op\s+(\d+) allocs/op)?`)

type benchResult struct {
	name   string
	nsOp   []float64
	bytes  int
	allocs int
}

func runPerf(e *Env, res *Result) {
	var rows []string
	budget := map[string]time.Duration{"BenchmarkTick": TickBudget, "BenchmarkGetState": GetStateBudget}
	if benches, out, err := runBenchmarks(e); err != nil {
		res.warn("bench_unavailable", "late-game benchmarks not run: %v", err).Detail = out
	} else {
		for _, name := range []string{"BenchmarkTick", "BenchmarkGetState"} {
			b, ok := benches[name]
			if !ok {
				res.fail("bench_missing", "%s did not report (see detail)", name).Detail = out
				continue
			}
			sort.Float64s(b.nsOp)
			med := time.Duration(b.nsOp[len(b.nsOp)/2])
			verdict := "✓"
			if med > budget[name] {
				verdict = "over budget"
				res.fail("perf_budget", "%s median %s per op, budget %s", name, med, budget[name]).Repro = "go test -run '^$' -bench '^" + name + "$' -benchmem ./game"
			}
			rows = append(rows, fmt.Sprintf("| %s (late-game engine) | %s | %s | %d B, %d allocs | %s |", name, med, budget[name], b.bytes, b.allocs, verdict))
		}
	}
	memRows, growth := perfLongRun(e, res)
	rows = append(rows, memRows...)
	res.Summary = growth
	res.section("Latency and memory", "| measure | actual | budget | per op | verdict |\n|---|---|---|---|---|\n%s", strings.Join(rows, "\n"))
}

// runBenchmarks runs the game package's late-game tick and GetState
// benchmarks (game/tick_perf_test.go) in a child `go test`.
func runBenchmarks(e *Env) (map[string]*benchResult, string, error) {
	if e.RepoRoot == "" {
		return nil, "", fmt.Errorf("repository root unknown")
	}
	count, n := "3", "2000x"
	if e.full() {
		count, n = "5", "5000x"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-run", "^$", "-bench", "^(BenchmarkTick|BenchmarkGetState)$",
		"-benchtime", n, "-count", count, "-benchmem", "./game")
	cmd.Dir = e.RepoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, string(out), fmt.Errorf("go test -bench: %v", err)
	}
	benches := map[string]*benchResult{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		m := benchLine.FindStringSubmatch(strings.TrimSpace(sc.Text()))
		if m == nil {
			continue
		}
		b := benches[m[1]]
		if b == nil {
			b = &benchResult{name: m[1]}
			benches[m[1]] = b
		}
		ns, _ := strconv.ParseFloat(m[2], 64)
		b.nsOp = append(b.nsOp, ns)
		b.bytes, _ = strconv.Atoi(m[3])
		b.allocs, _ = strconv.Atoi(m[4])
	}
	return benches, string(out), nil
}

// perfLongRun plays one long bot run, sampling the heap after a GC and the
// size of every collection in GameState at intervals, and timing GetState at
// every decision.
func perfLongRun(e *Env, res *Result) ([]string, string) {
	ticks := 30000
	if e.full() {
		ticks = 400000
	}
	const samples = 8
	type sample struct {
		tick  int
		heap  uint64
		sizes map[string]int
	}
	var got []sample
	var gs []time.Duration
	cfg := e.Base
	cfg.Cycles, cfg.MaxSim = 1, 100000*time.Hour
	start := time.Now()
	_, run := playUntil(cfg, e.SeedBase, func(r *runner) bool {
		t0 := time.Now()
		st := r.ge.GetState()
		gs = append(gs, time.Since(t0))
		if r.ticks > 0 && r.ticks%(ticks/samples) == 0 {
			runtime.GC()
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			sizes := map[string]int{}
			collectionSizes(reflect.ValueOf(st), "", sizes, 0)
			got = append(got, sample{tick: r.ticks, heap: ms.HeapAlloc, sizes: sizes})
		}
		return r.ticks >= ticks
	})
	wall := time.Since(start)
	for _, a := range run.Anomalies {
		res.fail(a.Kind+"/"+a.Check, "long run: %s", a.Message).Detail = a.Dump
	}
	var rows []string
	sort.Slice(gs, func(i, j int) bool { return gs[i] < gs[j] })
	if len(gs) > 0 {
		p50, p99 := gs[len(gs)/2], gs[len(gs)*99/100]
		verdict := "✓"
		if p99 > 2*GetStateBudget { // in-process timings include GC pauses and scheduler noise
			verdict = "p99 over 2x budget (in-process, not failing)"
			res.warn("getstate_p99", "in-process GetState p99 %s over twice the %s budget during a real run (age %s)", p99, GetStateBudget, run.FinalAge)
		}
		rows = append(rows, fmt.Sprintf("| GetState during play (%d calls, up to %s) | p50 %s, p99 %s | %s | - | %s |", len(gs), run.FinalAge, p50, p99, GetStateBudget, verdict))
	}
	if run.Ticks > 0 {
		rows = append(rows, fmt.Sprintf("| tick + bot decision, whole run | %s per tick | - | - | info |", (wall/time.Duration(run.Ticks)).Round(time.Microsecond)))
	}
	if len(got) < 2 {
		return rows, fmt.Sprintf("long run ended early (%s at %s after %d ticks)", run.Outcome, run.FinalAge, run.Ticks)
	}
	first, last := got[0], got[len(got)-1]
	verdict := "✓"
	if float64(last.heap) > HeapGrowthFactor*float64(first.heap)+HeapGrowthSlack {
		verdict = "unbounded?"
		res.fail("heap_growth", "heap after GC grew from %s at tick %d to %s at tick %d (limit %gx + %dMB)",
			bytesStr(first.heap), first.tick, bytesStr(last.heap), last.tick, HeapGrowthFactor, HeapGrowthSlack>>20)
	}
	rows = append(rows, fmt.Sprintf("| heap after GC, tick %d -> %d | %s -> %s | %gx + %dMB | - | %s |", first.tick, last.tick,
		bytesStr(first.heap), bytesStr(last.heap), HeapGrowthFactor, HeapGrowthSlack>>20, verdict))
	var growing []string
	for _, path := range sortedKeys(last.sizes) {
		n := last.sizes[path]
		if n > CollectionFail {
			res.fail("collection_growth", "GameState.%s holds %d elements at tick %d", path, n, last.tick)
		}
		mono := true
		for i := 1; i < len(got); i++ {
			mono = mono && got[i].sizes[path] > got[i-1].sizes[path]
		}
		if mono && n > CollectionWarn {
			res.warn("collection_growth", "GameState.%s grew at every sample, to %d elements by tick %d", path, n, last.tick)
		}
		if mono {
			growing = append(growing, fmt.Sprintf("%s %d->%d", path, first.sizes[path], n))
		}
	}
	summary := fmt.Sprintf("long run to %s over %d ticks; heap %s -> %s", run.FinalAge, run.Ticks, bytesStr(first.heap), bytesStr(last.heap))
	if len(growing) > 0 {
		rows = append(rows, fmt.Sprintf("| collections growing at every sample | %s | < %d | - | info |", cell(clip(strings.Join(growing, ", "), 600)), CollectionWarn))
	}
	return rows, summary
}

// collectionSizes records len() of every slice and map reachable in v,
// keyed by field path with indices and map keys folded away, summed.
func collectionSizes(v reflect.Value, path string, out map[string]int, depth int) {
	if depth > 6 || !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			collectionSizes(v.Elem(), path, out, depth+1)
		}
	case reflect.Struct:
		if v.Type() == timeType {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if f := v.Type().Field(i); f.PkgPath == "" {
				collectionSizes(v.Field(i), join(path, f.Name), out, depth+1)
			}
		}
	case reflect.Slice, reflect.Map:
		out[path] += v.Len()
		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				collectionSizes(v.Index(i), path+"[]", out, depth+1)
			}
		} else {
			it := v.MapRange()
			for it.Next() {
				collectionSizes(it.Value(), path+"[]", out, depth+1)
			}
		}
	}
}

func bytesStr(b uint64) string {
	return fmt.Sprintf("%.1fMB", float64(b)/(1<<20))
}
