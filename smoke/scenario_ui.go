package smoke

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// The ui scenario runs the UI sweeps in ui/smoke_sweep_test.go (build tag
// smoke) in a child `go test`, since they drive the ui package's internals
// on a simulated terminal:
//   - TestSmokeUISweep: splash pages, a new game and every overlay and
//     read-only command under every theme, at a roomy 180x56;
//   - TestSmokeUISmallTerminals: the dashboard and every overlay at
//     80x24 and 100x30 (fast tier: the default and one light theme at 80x24).
// Both fail on a panic, a frozen event loop or a blank screen.

var uiResult = regexp.MustCompile(`^\s*--- (PASS|FAIL|SKIP): (\S+)`)

func runUI(e *Env, res *Result) {
	if e.RepoRoot == "" {
		res.fail("ui_unavailable", "repository root unknown; run from the repo")
		return
	}
	themes, sizes := "default,light", "80x24"
	timeout := 8 * time.Minute
	if e.full() {
		themes, sizes = "all", "80x24,100x30"
		timeout = 20 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := []string{"test", "-tags", "smoke", "-count=1", "-v", "-timeout", timeout.String(),
		"-run", "^(TestSmokeUISweep|TestSmokeUISmallTerminals)$", "./ui"}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = e.RepoRoot
	cmd.Env = append(os.Environ(), "SMOKE_UI_THEMES="+themes, "SMOKE_UI_SIZES="+sizes)
	start := time.Now()
	out, err := cmd.CombinedOutput()
	var rows []string
	passed := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if m := uiResult.FindStringSubmatch(line); m != nil {
			rows = append(rows, fmt.Sprintf("| %s | %s |", m[2], m[1]))
			passed[m[2]] = m[1] == "PASS"
			if m[1] == "FAIL" && !strings.Contains(m[2], "/") {
				f := res.fail("ui_"+m[2], "%s failed", m[2])
				f.Repro = fmt.Sprintf("SMOKE_UI_THEMES=%s SMOKE_UI_SIZES=%s go test -tags smoke -count=1 -v -run '^%s$' ./ui", themes, sizes, m[2])
				f.Detail = tail(string(out), 150)
			}
		}
	}
	for _, t := range []string{"TestSmokeUISweep", "TestSmokeUISmallTerminals"} {
		if _, ok := passed[t]; !ok && err == nil {
			res.fail("ui_missing", "%s did not run", t).Detail = tail(string(out), 60)
		}
	}
	if err != nil && len(res.Failures) == 0 {
		f := res.fail("ui_go_test", "go test ./ui (smoke) failed: %v", err)
		f.Detail = tail(string(out), 150)
		f.Repro = "go test -tags smoke -count=1 -v -run '^(TestSmokeUISweep|TestSmokeUISmallTerminals)$' ./ui"
	}
	var logs []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "UI sweep:") || strings.Contains(line, "small terminals:") {
			logs = append(logs, strings.TrimSpace(line))
		}
	}
	res.Summary = fmt.Sprintf("themes %s at 180x56, plus %s at %s; %s", "all", themes, sizes, time.Since(start).Round(time.Second))
	res.section("Tests", "| test | result |\n|---|---|\n%s\n\n%s", strings.Join(rows, "\n"), strings.Join(logs, "\n\n"))
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
