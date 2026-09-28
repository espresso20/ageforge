package detmath

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The float rules for simulation code (CONTRIBUTING.md, "Float rules for
// simulation code"), enforced. simPackages is the code a run's result depends
// on: the game, its config and the packages it imports, and the smoke bot
// (its decisions steer the runs the determinism check compares).
var simPackages = []string{"boon", "config", "detmath", "flavor", "game", "smoke"}

const modulePath = "github.com/espresso20/ageforge"

func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found above %s: %v", root, err)
	}
	return root
}

// fmaInsn matches the fused multiply-add instructions the compiler emits:
// arm64 FMADDD/FMSUBD/FNMADDD/FNMSUBD (and the S forms), amd64 VFMADD231SD
// and friends (GOAMD64=v3 and up).
var fmaInsn = regexp.MustCompile(`\s(FN?M(ADD|SUB)[DS]|VFN?M(ADD|SUB)[0-9]{3}S[DS])\s`)

// asmPos pulls the source position out of a -S listing line:
// "\t0x0034 00052 (/abs/path/game/engine.go:410)\tFMADDD\t...".
var asmPos = regexp.MustCompile(`\(([^()]+\.go):(\d+)\)`)

// TestNoFusedMultiplyAdd compiles the simulation packages for arm64 and for
// amd64 v3 (the targets whose compilers fuse a*b+c into one FMA instruction)
// and fails on any FMA instruction in them. A fused a*b+c skips the rounding
// of a*b, so it can differ in the last bit from the separate multiply and add
// that amd64 (v1, the default) runs, and a run drifts apart. Fix a site by
// rounding the product explicitly: float64(a*b) + c.
func TestNoFusedMultiplyAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the simulation packages for two more architectures")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go command on PATH")
	}
	root := moduleRoot(t)
	dirs := map[string]bool{}
	args := []string{"build"}
	for _, p := range simPackages {
		dirs[filepath.Join(root, p)] = true
		args = append(args, "-gcflags="+modulePath+"/"+p+"=-S")
	}
	for _, p := range simPackages {
		args = append(args, "./"+p)
	}
	targets := []struct{ name, goarch, level string }{
		{"arm64", "arm64", ""},
		{"amd64 v3", "amd64", "v3"},
	}
	for _, tg := range targets {
		cmd := exec.Command(goBin, args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+tg.goarch, "CGO_ENABLED=0")
		if tg.level != "" {
			cmd.Env = append(cmd.Env, "GOAMD64="+tg.level)
		}
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		// The listing comes back from the build cache too, so a warm run is fast.
		if err := cmd.Run(); err != nil {
			t.Fatalf("%s: go %s: %v\n%s", tg.name, strings.Join(args, " "), err, tail(out.String(), 40))
		}
		sites := map[string]string{}
		sc := bufio.NewScanner(&out)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			insn := fmaInsn.FindStringSubmatch(line)
			if insn == nil {
				continue
			}
			m := asmPos.FindStringSubmatch(line)
			if m == nil || !dirs[filepath.Dir(m[1])] {
				continue
			}
			rel, _ := filepath.Rel(root, m[1])
			sites[rel+":"+m[2]] = insn[1]
		}
		if len(sites) == 0 {
			continue
		}
		keys := make([]string, 0, len(sites))
		for k := range sites {
			keys = append(keys, k)
		}
		sortPositions(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "\n\t%s\t%s\t%s", k, sites[k], sourceLine(root, k))
		}
		t.Errorf("%s: %d fused multiply-add site(s) in simulation code; round the product with float64(a*b) (CONTRIBUTING.md, \"Float rules for simulation code\"):%s",
			tg.name, len(keys), b.String())
	}
}

// transcendental are the package math functions whose results are not the
// same bits on every architecture (assembly on some, Go compiled with FMA on
// others, CPU-feature-dependent paths). Simulation code calls detmath for
// Log, Log10, Exp and Pow and uses none of the others.
var transcendental = map[string]string{
	"Pow": "detmath.Pow", "Exp": "detmath.Exp", "Log": "detmath.Log", "Log10": "detmath.Log10",
	"Exp2": "", "Expm1": "", "Log2": "", "Log1p": "", "Logb": "", "Ilogb": "",
	"Sin": "", "Cos": "", "Tan": "", "Sincos": "", "Asin": "", "Acos": "", "Atan": "", "Atan2": "",
	"Sinh": "", "Cosh": "", "Tanh": "", "Asinh": "", "Acosh": "", "Atanh": "",
	"Cbrt": "", "Hypot": "", "Gamma": "", "Lgamma": "", "Erf": "", "Erfc": "", "Erfinv": "", "Erfcinv": "",
	"J0": "", "J1": "", "Jn": "", "Y0": "", "Y1": "", "Yn": "", "Pow10": "",
}

// TestNoStdlibTranscendentals fails on a call to one of package math's
// architecture-dependent functions from simulation code (tests excluded).
// Sqrt, Floor, Ceil, Round, Abs, Mod, Min, Max and the like are exact and
// allowed.
func TestNoStdlibTranscendentals(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()
	var bad []string
	for _, p := range simPackages {
		if p == "detmath" {
			continue // the implementation itself
		}
		files, err := filepath.Glob(filepath.Join(root, p, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
			if err != nil {
				t.Fatal(err)
			}
			name := mathImportName(f)
			if name == "" {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				id, ok := sel.X.(*ast.Ident)
				if !ok || id.Name != name {
					return true
				}
				if use, ok := transcendental[sel.Sel.Name]; ok {
					pos := fset.Position(sel.Pos())
					rel, _ := filepath.Rel(root, pos.Filename)
					msg := fmt.Sprintf("%s:%d: math.%s", rel, pos.Line, sel.Sel.Name)
					if use != "" {
						msg += " (use " + use + ")"
					}
					bad = append(bad, msg)
				}
				return true
			})
		}
	}
	if len(bad) > 0 {
		t.Errorf("simulation code calls architecture-dependent math functions (CONTRIBUTING.md, \"Float rules for simulation code\"):\n\t%s",
			strings.Join(bad, "\n\t"))
	}
}

// mathImportName is the name file f refers to package math by, or "".
func mathImportName(f *ast.File) string {
	for _, imp := range f.Imports {
		if path, _ := strconv.Unquote(imp.Path.Value); path == "math" {
			if imp.Name != nil {
				return imp.Name.Name
			}
			return "math"
		}
	}
	return ""
}

// sortPositions sorts "file:line" keys by file, then numerically by line.
func sortPositions(keys []string) {
	split := func(k string) (string, int) {
		i := strings.LastIndexByte(k, ':')
		n, _ := strconv.Atoi(k[i+1:])
		return k[:i], n
	}
	sort.Slice(keys, func(i, j int) bool {
		fi, li := split(keys[i])
		fj, lj := split(keys[j])
		if fi != fj {
			return fi < fj
		}
		return li < lj
	})
}

// sourceLine returns the trimmed source at "file:line" (relative to root).
func sourceLine(root, pos string) string {
	i := strings.LastIndexByte(pos, ':')
	n, _ := strconv.Atoi(pos[i+1:])
	data, err := os.ReadFile(filepath.Join(root, pos[:i]))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	if n < 1 || n > len(lines) {
		return ""
	}
	return strings.TrimSpace(lines[n-1])
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
