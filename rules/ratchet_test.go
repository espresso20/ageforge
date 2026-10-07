package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The ratchet. An engine's definitions belong to its ruleset, so code that
// has an engine, a manager or a snapshot to hand reads that set. Two kinds
// of lookup still reach past it, to tables the whole process shares:
//
//   - config: a direct reference to one of package config's definition
//     lookups (config.Ages, config.BuildingByKey, config.AgeTargets, ...);
//   - core: a reference to rules.Core, the one set built from config, by
//     code with no engine or snapshot in reach.
//
// lookupBudget is how many of each the game, the UI and the map model hold
// today, test files left out. The numbers may only go down. A count above
// its budget fails: read the set the engine, the manager or the snapshot
// holds (ge.rules, m.rules, state.Ruleset()), or hand the set to the helper
// as an argument. When a count drops, lower its budget here so it stays
// down; the test says which.
var lookupBudget = map[string]lookupCount{
	"game":     {config: 6, core: 25},
	"ui":       {config: 25, core: 0},
	"mapmodel": {config: 4, core: 2},
}

type lookupCount struct{ config, core int }

// pureConfig is the part of package config that is not a definition lookup:
// formulas and formatters that depend on their arguments alone, and the log
// quips. Every other exported function or variable of config counts.
var pureConfig = map[string]bool{
	"AgePositions": true, "AnyResource": true, "StretchTicksBy": true,
	"ClampMastery": true, "MasteryK": true, "AgeSpeed": true,
	"DefenseMitigation": true, "AgeThreat": true,
	"DurationText": true, "FormatAmount": true, "FormatPercent": true, "FormatRateValue": true,
	"PickLogFlavor": true, "LogFlavorMoments": true,
	"PriceLevelsByAge": true, "Incomes": true, "FlowDealLevels": true,
	"ExchangeRateAt": true, "MarketRateAt": true, "MarketPairsAt": true, "MarketOffersAt": true,
	"DealPriceLevelAt": true, "PricedResourcesAt": true,
	"TechTerms": true, "RouteFeature": true,
}

const (
	modulePath = "github.com/espresso20/ageforge"
	configPath = modulePath + "/config"
	rulesPath  = modulePath + "/rules"
)

func TestConfigLookupRatchet(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	lookups := configLookups(t, filepath.Join(root, "config"))
	for name := range pureConfig {
		if !lookups[name] {
			t.Errorf("pureConfig lists %s, which package config does not export as a function or variable", name)
		}
		delete(lookups, name)
	}
	dirs := make([]string, 0, len(lookupBudget))
	for dir := range lookupBudget {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		budget := lookupBudget[dir]
		got, where := countLookups(t, filepath.Join(root, dir), lookups)
		t.Logf("%s: %d direct config lookups (budget %d), %d rules.Core references (budget %d)",
			dir, got.config, budget.config, got.core, budget.core)
		if got.config > budget.config {
			t.Errorf("%s has %d direct config lookups, over its budget of %d. Read the ruleset the engine, manager or snapshot holds instead. They are:\n%s",
				dir, got.config, budget.config, strings.Join(where.config, "\n"))
		}
		if got.core > budget.core {
			t.Errorf("%s has %d rules.Core references, over its budget of %d. Read the ruleset the engine, manager or snapshot holds instead. They are:\n%s",
				dir, got.core, budget.core, strings.Join(where.core, "\n"))
		}
		if got.config < budget.config || got.core < budget.core {
			t.Logf("%s is under budget: lower lookupBudget[%q] to {config: %d, core: %d} so it stays there", dir, dir, got.config, got.core)
		}
	}
}

// configLookups is every exported top-level function and variable of the
// package in dir, test files left out.
func configLookups(t *testing.T, dir string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, f := range parseDir(t, dir, false) {
		for _, d := range f.file.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name.IsExported() {
					out[d.Name.Name] = true
				}
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					for _, name := range spec.(*ast.ValueSpec).Names {
						if name.IsExported() {
							out[name.Name] = true
						}
					}
				}
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("no exported functions found in %s", dir)
	}
	return out
}

type lookupSites struct{ config, core []string }

// countLookups counts, in every non-test Go file under dir, the references
// to a config lookup and to rules.Core, and lists where they are.
func countLookups(t *testing.T, dir string, lookups map[string]bool) (lookupCount, lookupSites) {
	t.Helper()
	var n lookupCount
	var where lookupSites
	for _, f := range parseDir(t, dir, true) {
		configName, rulesName := "", ""
		for _, imp := range f.file.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			name := filepath.Base(path)
			if imp.Name != nil {
				name = imp.Name.Name
			}
			switch path {
			case configPath:
				configName = name
			case rulesPath:
				rulesName = name
			}
		}
		if configName == "" && rulesName == "" {
			continue
		}
		ast.Inspect(f.file, func(node ast.Node) bool {
			sel, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Obj != nil { // a local of the same name, not the package
				return true
			}
			site := f.rel + ":" + strconv.Itoa(f.fset.Position(sel.Pos()).Line) + ": " + pkg.Name + "." + sel.Sel.Name
			switch {
			case pkg.Name == configName && lookups[sel.Sel.Name]:
				n.config++
				where.config = append(where.config, "  "+site)
			case pkg.Name == rulesName && sel.Sel.Name == "Core":
				n.core++
				where.core = append(where.core, "  "+site)
			}
			return true
		})
	}
	return n, where
}

type parsedFile struct {
	rel  string
	fset *token.FileSet
	file *ast.File
}

// parseDir parses the non-test Go files in dir, and in the directories
// below it when deep is set, in path order.
func parseDir(t *testing.T, dir string, deep bool) []parsedFile {
	t.Helper()
	var out []parsedFile
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && !deep {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(filepath.Dir(dir), path)
		out = append(out, parsedFile{rel: filepath.ToSlash(rel), fset: fset, file: file})
		return nil
	})
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	if len(out) == 0 {
		if _, statErr := os.Stat(dir); statErr != nil {
			t.Fatalf("%s: %v", dir, statErr)
		}
		t.Fatalf("no Go files in %s", dir)
	}
	return out
}
