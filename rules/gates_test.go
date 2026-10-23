package rules

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

const gatesChildEnv = "AGEFORGE_TEST_GATES_CHILD"

// The measured gate is picked by AGEFORGE_GATES=v2 when the program starts
// and never again: the value the variable has at start-up is what the config
// and the core ruleset report, and changing it later changes nothing.
func TestGateRuleIsReadAtStartUp(t *testing.T) {
	if os.Getenv(gatesChildEnv) == "1" {
		// The child process: say what it saw at start-up, then what it
		// sees after the variable is flipped.
		start := config.GatesV2()
		core := Core().GatesV2()
		flip := "v2"
		if start {
			flip = "v1"
		}
		os.Setenv(config.GatesEnv, flip)
		fmt.Printf("GATES-CHILD start=%v core=%v after=%v fresh=%v\n", start, core, config.GatesV2(), FromConfig().GatesV2)
		return
	}
	for _, c := range []struct {
		env  string
		want bool
	}{{"", false}, {"v1", false}, {"nonsense", false}, {"v2", true}, {"V2", true}, {" v2 ", true}} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestGateRuleIsReadAtStartUp$", "-test.v")
		cmd.Env = append(os.Environ(), gatesChildEnv+"=1", config.GatesEnv+"="+c.env)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%q: %v\n%s", c.env, err, out)
		}
		var line string
		for _, l := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(l, "GATES-CHILD") {
				line = l
			}
		}
		want := fmt.Sprintf("GATES-CHILD start=%v core=%v after=%v fresh=%v", c.want, c.want, c.want, c.want)
		if line != want {
			t.Errorf("AGEFORGE_GATES=%q: the process says %q, want %q", c.env, line, want)
		}
	}

	// In this process too: flipping the variable now moves nothing.
	before := config.GatesV2()
	flip := "v2"
	if before {
		flip = "v1"
	}
	t.Setenv(config.GatesEnv, flip)
	if config.GatesV2() != before || FromConfig().GatesV2 != before || Core().GatesV2() != before {
		t.Errorf("the gate rule changed after start-up: config %v, source %v, core %v, started as %v",
			config.GatesV2(), FromConfig().GatesV2, Core().GatesV2(), before)
	}
}

func TestParseGateRule(t *testing.T) {
	for in, want := range map[string]bool{"": false, "v1": false, "v2": true, "V2": true, "  v2": true, "v3": false, "2": false} {
		if got := config.ParseGateRule(in); got != want {
			t.Errorf("ParseGateRule(%q) = %v, want %v", in, got, want)
		}
	}
}

// A set carries the switch and the measures it was compiled with, and hands
// out copies of them.
func TestSetCarriesTheGateRule(t *testing.T) {
	src := FromConfig()
	src.GatesV2 = false
	if Compile(src).GatesV2() {
		t.Error("a source with GatesV2 off compiled to a measured set")
	}
	src.GatesV2 = true
	set := Compile(src)
	if !set.GatesV2() {
		t.Fatal("a source with GatesV2 on compiled to a plain set")
	}
	ages := set.Ages()
	for i, a := range ages {
		m, ok := set.GateMeasure(a.Key)
		if i == len(ages)-1 {
			if ok {
				t.Errorf("the last age %s has a gate measure: %+v", a.Key, m)
			}
			continue
		}
		if !ok || m.Age != a.Key || m.People <= 0 || len(m.Materials) == 0 || m.Income <= 0 || m.KeyTech == "" {
			t.Errorf("%s: measure %+v (%v)", a.Key, m, ok)
		}
		for _, res := range m.Materials {
			if _, ok := set.Resource(res); !ok {
				t.Errorf("%s: material %s is not a resource", a.Key, res)
			}
		}
		if td, ok := set.Tech(m.KeyTech); !ok || td.Age != a.Key {
			t.Errorf("%s: key tech %s is %+v (%v)", a.Key, m.KeyTech, td.Age, ok)
		}
		if m.KeyTech == set.Keystone(a.Key) {
			t.Errorf("%s: the key tech is the keystone", a.Key)
		}
	}
	m, _ := set.GateMeasure("stone_age")
	m.Materials[0] = "changed"
	if again, _ := set.GateMeasure("stone_age"); again.Materials[0] == "changed" {
		t.Error("GateMeasure hands out the set's own slice")
	}
}
