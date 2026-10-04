package smoke

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui"
)

// The fuzz scenario types commands into the real command handler
// (ui.HandleCommand, what the dashboard's input field calls) against a live
// engine: valid commands from the command registry with valid arguments from
// the prompt's completions (ui.NewAutoCompleter), valid
// commands with hostile arguments, mangled commands and garbage. After each
// command the engine ticks a little, and it must not panic, hang, break an
// invariant or stop ticking. Every few commands the bot plays for a while
// so the fuzzer reaches states where more commands do something. The whole
// script is seeded; a failure is replayed and shrunk to a short command
// list.

// fuzzStep is one step of a script: a command, or a stretch of bot play.
type fuzzStep struct {
	Cmd   string `json:"cmd,omitempty"`
	Ticks int    `json:"ticks,omitempty"` // ticks after the command
	Bot   int    `json:"bot,omitempty"`   // bot decisions (5 ticks each)
}

func (s fuzzStep) String() string {
	if s.Bot > 0 {
		return fmt.Sprintf("[bot plays %d decisions]", s.Bot)
	}
	return fmt.Sprintf("%q +%d ticks", s.Cmd, s.Ticks)
}

// fuzzFailure is the first thing that broke in a script.
type fuzzFailure struct {
	check, msg, detail string
	step               int
}

// cmdTimeout is how long one command may run before it counts as a hang.
const cmdTimeout = 10 * time.Second

// hostileArgs are arguments no command expects.
var hostileArgs = []string{
	"-1", "0", "-0", "1e308", "-1e308", "1e309", "NaN", "nan", "Inf", "+Inf", "-Inf",
	"9223372036854775807", "-9223372036854775808", "18446744073709551616", "0x1F", "1_000",
	"3.14159", "1e-320", "0.0000001", "all", "max", "ALL", "Max", "💥", "ñandú", "日本語", "\u200b",
	"\x00", "'", "\"", "%s%n", "${HOME}", "*", "?", "--", "-h", "../x", "../../x",
	strings.Repeat("9", 400), strings.Repeat("z", 300), "food", "wood", "gold", "yes", "confirm",
}

// pathCommands take file paths; the fuzzer keeps them inside its temp dir.
func pathCommand(cmd string) bool {
	f := strings.Fields(strings.ToLower(cmd))
	return len(f) >= 2 && (f[0] == "account" || f[0] == "acct") && (f[1] == "export" || f[1] == "import")
}

type fuzzer struct {
	r        *rand.Rand
	commands []string
}

// newFuzzer builds the command corpus from the command registry: every
// command name and alias HandleCommand takes (quit is the dashboard's, and
// would only report an unknown command here).
func newFuzzer(seed int64, ge *game.GameEngine) *fuzzer {
	var cmds []string
	for _, c := range ui.Commands() {
		if c.Names[0] != "quit" {
			cmds = append(cmds, c.Names...)
		}
	}
	sort.Strings(cmds)
	return &fuzzer{r: rand.New(rand.NewSource(seed)), commands: cmds}
}

func (f *fuzzer) pick(xs []string) string { return xs[f.r.Intn(len(xs))] }

// next generates one command line against ge's current state.
func (f *fuzzer) next(ge *game.GameEngine) string {
	comp := ui.NewAutoCompleter(ge)
	var s string
	switch n := f.r.Intn(100); {
	case n < 45: // a valid command walked through its suggestions
		s = f.pick(f.commands)
		for depth := 0; depth < 3; depth++ {
			sugg := comp(s + " ")
			if len(sugg) == 0 || f.r.Intn(4) == 0 {
				break
			}
			s = strings.TrimSpace(f.pick(sugg))
		}
		if f.r.Intn(3) == 0 {
			s += " " + f.number()
		}
	case n < 70: // a valid command with hostile arguments
		s = f.pick(f.commands)
		for i := f.r.Intn(3) + 1; i > 0; i-- {
			s += " " + f.pick(hostileArgs)
		}
	case n < 80: // a valid line, mangled
		s = f.pick(f.commands)
		if sugg := comp(s + " "); len(sugg) > 0 {
			s = f.pick(sugg)
		}
		s = f.mangle(s)
	case n < 92: // garbage
		s = f.garbage(f.r.Intn(40))
	default: // edge shapes
		switch f.r.Intn(5) {
		case 0:
			s = ""
		case 1:
			s = " \t  "
		case 2:
			s = strings.Repeat("a", 5000)
		case 3:
			s = f.pick(f.commands) + strings.Repeat(" 1", 1000)
		default:
			s = strings.Repeat(f.pick(f.commands)+" ", 50)
		}
	}
	if pathCommand(s) {
		// Never let a fuzzed path escape the temp dir.
		fs := strings.Fields(s)
		for i := 2; i < len(fs); i++ {
			if strings.ContainsAny(fs[i], `/\`) || strings.Contains(fs[i], "..") {
				fs[i] = "fuzz-export.json"
			}
		}
		s = strings.Join(fs, " ")
	}
	return s
}

func (f *fuzzer) number() string {
	switch f.r.Intn(6) {
	case 0:
		return fmt.Sprint(f.r.Intn(10))
	case 1:
		return fmt.Sprint(f.r.Intn(1000))
	case 2:
		return fmt.Sprint(-f.r.Intn(1000))
	case 3:
		return fmt.Sprintf("%g", f.r.Float64()*1e12)
	case 4:
		return f.pick([]string{"all", "max", "1e308", "-1e308", "NaN", "Inf"})
	default:
		return fmt.Sprintf("%.3f", f.r.Float64()*100)
	}
}

func (f *fuzzer) mangle(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	switch f.r.Intn(5) {
	case 0:
		return strings.ToUpper(s)
	case 1:
		return string(rs[:f.r.Intn(len(rs))])
	case 2:
		return strings.ReplaceAll(s, " ", "   ")
	case 3:
		i := f.r.Intn(len(rs))
		return string(rs[:i]) + f.garbage(3) + string(rs[i:])
	default:
		return s + " " + s
	}
}

func (f *fuzzer) garbage(n int) string {
	pools := []string{
		"abcdefghijklmnopqrstuvwxyz0123456789 ",
		"!@#$%^&*()_+-=[]{};':\",./<>?\\|`~",
		"éüñçøßåæœ漢字かなカナ한국어Ωπ∑∞≠🙂🔥💀\u200b\u202e",
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		p := []rune(pools[f.r.Intn(len(pools))])
		sb.WriteRune(p[f.r.Intn(len(p))])
	}
	return sb.String()
}

// fuzzExec replays script on a fresh engine seeded with seed. gen, if set,
// fills each command step's Cmd from the fuzzer as it goes (the first
// pass); replays leave it nil. It stops at the first failure.
func fuzzExec(seed int64, script []fuzzStep, gen *fuzzer) (*fuzzFailure, []fuzzStep) {
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := NewBot(ge)
	defs := ge.Rules().BuildingMap()
	var done []fuzzStep
	for i, s := range script {
		if gen != nil && s.Bot == 0 {
			s.Cmd = gen.next(ge)
		}
		done = append(done, s)
		if f := fuzzStepRun(ge, bot, s, defs); f != nil {
			f.step = i
			return f, done
		}
	}
	return nil, done
}

func fuzzStepRun(ge *game.GameEngine, bot *Bot, s fuzzStep, defs map[string]config.BuildingDef) (fail *fuzzFailure) {
	if s.Bot > 0 {
		defer func() {
			if rec := recover(); rec != nil {
				fail = &fuzzFailure{check: "panic", msg: fmt.Sprintf("panic during bot play: %v", rec), detail: string(debug.Stack())}
			}
		}()
		for i := 0; i < s.Bot; i++ {
			st := ge.GetState()
			if st.PendingCatastrophe != "" {
				_ = ge.Endure()
			} else if st.LastPassage.Pending {
				_ = ge.EndureLastPassage()
			} else if st.AgeReady {
				_ = ge.AdvanceAge()
			}
			bot.Play(ge.GetState())
			ge.StepTicks(5)
		}
		return nil
	}

	type outcome struct {
		res   ui.CommandResult
		panic interface{}
		stack []byte
	}
	ch := make(chan outcome, 1)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				ch <- outcome{panic: rec, stack: debug.Stack()}
			}
		}()
		ch <- outcome{res: ui.HandleCommand(s.Cmd, ge)}
	}()
	select {
	case o := <-ch:
		if o.panic != nil {
			return &fuzzFailure{check: "panic", msg: fmt.Sprintf("command %q panicked: %v", s.Cmd, o.panic), detail: string(o.stack)}
		}
	case <-time.After(cmdTimeout):
		return &fuzzFailure{check: "hang", msg: fmt.Sprintf("command %q did not return within %s", s.Cmd, cmdTimeout), detail: dumpStacks()}
	}
	before := ge.GetState().Tick
	if f := func() (f *fuzzFailure) {
		defer func() {
			if rec := recover(); rec != nil {
				f = &fuzzFailure{check: "panic", msg: fmt.Sprintf("tick after %q panicked: %v", s.Cmd, rec), detail: string(debug.Stack())}
			}
		}()
		ge.StepTicks(s.Ticks)
		return nil
	}(); f != nil {
		return f
	}
	st := ge.GetState()
	if st.Tick != before+s.Ticks {
		return &fuzzFailure{check: "not_ticking", msg: fmt.Sprintf("after %q, %d ticks moved the counter from %d to %d", s.Cmd, s.Ticks, before, st.Tick)}
	}
	for _, p := range invariantProblems(st, defs) {
		// Storage feasibility is a property of config, not of a command.
		if strings.HasPrefix(p.check, "required_") || p.check == "requirement_over_storage" {
			continue
		}
		return &fuzzFailure{check: p.check, msg: fmt.Sprintf("after %q: %s", s.Cmd, p.msg), detail: Dump(st, 0, 0)}
	}
	return nil
}

func dumpStacks() string {
	buf := make([]byte, 1<<20)
	return string(buf[:runtime.Stack(buf, true)])
}

// fuzzScript is the shape of a run: commands with a stretch of bot play
// every botEvery commands.
func fuzzScript(r *rand.Rand, n, botEvery, botLen int) []fuzzStep {
	var s []fuzzStep
	for i := 0; i < n; i++ {
		if i%botEvery == 0 {
			s = append(s, fuzzStep{Bot: botLen})
		}
		s = append(s, fuzzStep{Ticks: 1 + r.Intn(3)})
	}
	return s
}

// shrink cuts a failing script down while the failure (same check) still
// happens, within budget: first the failing command alone on a fresh game,
// then chunks of other commands (ddmin), then chunks of bot play. Adjacent
// bot stretches are merged, which replays the same.
func shrink(seed int64, script []fuzzStep, check string, budget time.Duration) []fuzzStep {
	deadline := time.Now().Add(budget)
	fails := func(s []fuzzStep) bool {
		f, _ := fuzzExec(seed, s, nil)
		return f != nil && f.check == check
	}
	last := len(script) - 1
	if alone := script[last:]; fails(alone) {
		return alone
	}
	cur := script
	pass := func(isTarget func(fuzzStep) bool) {
		for chunk := len(cur) / 2; chunk >= 1 && time.Now().Before(deadline); chunk /= 2 {
			for i := 0; i < len(cur)-1 && time.Now().Before(deadline); {
				j := i
				var cand []fuzzStep
				removed := 0
				for k, st := range cur {
					if k >= i && removed < chunk && isTarget(st) && k != len(cur)-1 {
						removed++
						j = k
						continue
					}
					cand = append(cand, st)
				}
				if removed > 0 && fails(cand) {
					cur = cand
					continue
				}
				i = j + 1
			}
		}
	}
	pass(func(s fuzzStep) bool { return s.Bot == 0 })
	pass(func(s fuzzStep) bool { return s.Bot > 0 })
	return mergeBot(cur)
}

// mergeBot folds adjacent bot stretches into one.
func mergeBot(s []fuzzStep) []fuzzStep {
	var out []fuzzStep
	for _, st := range s {
		if n := len(out); n > 0 && st.Bot > 0 && out[n-1].Bot > 0 {
			out[n-1].Bot += st.Bot
			continue
		}
		out = append(out, st)
	}
	return out
}

func runFuzz(e *Env, res *Result) {
	cmds, seeds := 300, e.seeds(1)
	botEvery, botLen := 15, 200
	shrinkBudget := 60 * time.Second
	if e.full() {
		cmds, seeds = 25000, e.seeds(2)
		botEvery, botLen = 25, 200
		shrinkBudget = 5 * time.Minute
	}
	if e.FuzzCommands > 0 {
		cmds = e.FuzzCommands
	}
	total, byFirst := 0, map[string]int{}
	var rows []string
	for _, seed := range seeds {
		ge := game.NewGameEngine()
		gen := newFuzzer(seed, ge)
		script := fuzzScript(rand.New(rand.NewSource(seed^0x5eed)), cmds, botEvery, botLen)
		start := time.Now()
		f, done := fuzzExec(seed, script, gen)
		ran := 0
		for _, s := range done {
			if s.Bot == 0 {
				ran++
				fs := strings.Fields(s.Cmd)
				switch {
				case len(fs) == 0:
					byFirst["(empty)"]++
				case contains(gen.commands, strings.ToLower(fs[0])):
					byFirst[strings.ToLower(fs[0])]++
				default:
					byFirst["(not a command)"]++
				}
			}
		}
		total += ran
		verdict := "ok"
		if f != nil {
			verdict = f.check
			min := done
			if f.check != "hang" {
				min = shrink(seed, done, f.check, shrinkBudget)
			}
			var lines []string
			for _, s := range min {
				lines = append(lines, s.String())
			}
			fd := res.fail("fuzz_"+f.check, "seed %d, command %d: %s", seed, ran, f.msg)
			fd.Seed = seed
			fd.Detail = f.detail
			fd.Repro = fmt.Sprintf("go run ./cmd/smoke -scenario fuzz -seed-base %d -seeds 1 -v   # replays the whole script\n# shrunk to %d step(s) from a fresh engine seeded %d:\n%s",
				seed, len(min), seed, strings.Join(lines, "\n"))
		}
		rows = append(rows, fmt.Sprintf("| %d | %d | %s | %s |", seed, ran, time.Since(start).Round(100*time.Millisecond), verdict))
		e.logf("  fuzz seed %d: %d commands, %s", seed, ran, verdict)
	}
	var mix []string
	for _, k := range sortedKeys(byFirst) {
		mix = append(mix, fmt.Sprintf("%s %d", k, byFirst[k]))
	}
	res.Summary = fmt.Sprintf("%d command(s) over %d seed(s)", total, len(seeds))
	res.section("Runs", "| seed | commands | wall | result |\n|---|---|---|---|\n%s\n\nCommands by first word: %s", strings.Join(rows, "\n"), clip(strings.Join(mix, ", "), 3000))
}
