package smoke

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// WriteJSON writes the machine-readable session report.
func (s *Session) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (s *Session) verdict() string {
	if s.Failed {
		return "FAIL"
	}
	return "PASS"
}

func statusMark(st string) string {
	switch st {
	case StatusPass:
		return "✓ pass"
	case StatusFail:
		return "✗ FAIL"
	default:
		return "- skip"
	}
}

func (s *Session) scenarioTable(sb *strings.Builder) {
	sb.WriteString("| scenario | result | time | summary |\n|---|---|---|---|\n")
	for _, r := range s.Scenarios {
		fmt.Fprintf(sb, "| %s | %s | %s | %s |\n", r.Name, statusMark(r.Status),
			(time.Duration(r.WallMs) * time.Millisecond).Round(100*time.Millisecond), cell(r.Summary))
	}
}

// cell keeps a table cell on one line.
func cell(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "|", "\\|")
}

// pacing returns the first progression summary, the one the pacing table
// is drawn from.
func (s *Session) pacing() *Summary {
	for _, r := range s.Scenarios {
		if r.Name == "progression" && len(r.Progression) > 0 {
			return r.Progression[0]
		}
	}
	return nil
}

// WriteSummary writes the short Markdown summary for the CI job page and the
// email: the verdict, one row per scenario, the pacing table and the
// failures, capped so a bad night stays readable.
func (s *Session) WriteSummary(w io.Writer) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## AgeForge smoke (%s tier): %s\n\n", s.Tier, s.verdict())
	fmt.Fprintf(&sb, "%d scenario(s) in %s. Pacing is **%s**", len(s.Scenarios),
		(time.Duration(s.WallMs) * time.Millisecond).Round(time.Second), s.Pacing)
	if s.Pacing != PacingEnforce {
		sb.WriteString(" (graded, never fails the job)")
	} else {
		sb.WriteString(" for the progression scenario's first cycle (graded only elsewhere)")
	}
	sb.WriteString(".\n\n")
	s.scenarioTable(&sb)
	if p := s.pacing(); p != nil && len(p.Pacing) > 0 {
		sb.WriteString("\n### Pacing (progression)\n\n")
		p.writePacingTable(&sb)
	}
	n := 0
	for _, r := range s.Scenarios {
		n += len(r.Failures)
	}
	if n > 0 {
		fmt.Fprintf(&sb, "\n### Failures (%d)\n\n", n)
		shown := 0
		for _, r := range s.Scenarios {
			for _, f := range r.Failures {
				if shown == 25 {
					break
				}
				fmt.Fprintf(&sb, "- **%s / %s**: %s", r.Name, f.Check, clip(f.Message, 400))
				if f.Repro != "" {
					fmt.Fprintf(&sb, " Repro: `%s`", clip(strings.ReplaceAll(f.Repro, "\n", "; "), 300))
				}
				sb.WriteString("\n")
				shown++
			}
		}
		if shown < n {
			fmt.Fprintf(&sb, "- ...and %d more in report.md\n", n-shown)
		}
	}
	var warns []string
	for _, r := range s.Scenarios {
		for _, f := range r.Warnings {
			warns = append(warns, fmt.Sprintf("- %s / %s: %s", r.Name, f.Check, clip(f.Message, 300)))
		}
	}
	if len(warns) > 0 {
		fmt.Fprintf(&sb, "\n### Warnings (%d, not failing)\n\n", len(warns))
		if len(warns) > 15 {
			warns = append(warns[:15], fmt.Sprintf("- ...and %d more in report.md", len(warns)-15))
		}
		sb.WriteString(strings.Join(warns, "\n") + "\n")
	}
	sb.WriteString("\nFull report: report.md in the smoke-report artifact.\n")
	_, err := io.WriteString(w, sb.String())
	return err
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// WriteMarkdown writes the full session report: the summary, then every
// scenario's findings and sections.
func (s *Session) WriteMarkdown(w io.Writer) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# AgeForge smoke report: %s\n\n", s.verdict())
	fmt.Fprintf(&sb, "Tier `%s`, pacing `%s`, started %s, took %s.\n\n", s.Tier, s.Pacing,
		s.Started.UTC().Format(time.RFC3339), (time.Duration(s.WallMs) * time.Millisecond).Round(time.Second))
	s.scenarioTable(&sb)
	for _, r := range s.Scenarios {
		fmt.Fprintf(&sb, "\n## %s: %s\n\n", r.Name, strings.ToUpper(r.Status))
		for _, sc := range Scenarios() {
			if sc.Name == r.Name {
				sb.WriteString(sc.Desc + ".\n\n")
			}
		}
		if r.Summary != "" {
			sb.WriteString(r.Summary + "\n\n")
		}
		if len(r.Failures) > 0 {
			fmt.Fprintf(&sb, "### Failures (%d)\n\n", len(r.Failures))
			writeFindings(&sb, r.Failures)
		}
		if len(r.Warnings) > 0 {
			fmt.Fprintf(&sb, "### Warnings (%d, not failing)\n\n", len(r.Warnings))
			writeFindings(&sb, r.Warnings)
		}
		for _, sec := range r.Sections {
			fmt.Fprintf(&sb, "### %s\n\n%s\n\n", sec.Title, strings.TrimRight(sec.Body, "\n"))
		}
	}
	_, err := io.WriteString(w, sb.String())
	return err
}

func writeFindings(sb *strings.Builder, fs []Finding) {
	for _, f := range fs {
		fmt.Fprintf(sb, "- **%s**", f.Check)
		if f.Seed != 0 {
			fmt.Fprintf(sb, " (seed %d)", f.Seed)
		}
		fmt.Fprintf(sb, ": %s\n", f.Message)
		if f.Repro != "" {
			fmt.Fprintf(sb, "\n  Repro:\n\n  ```\n  %s\n  ```\n\n", strings.ReplaceAll(strings.TrimRight(f.Repro, "\n"), "\n", "\n  "))
		}
		if f.Detail != "" {
			fmt.Fprintf(sb, "\n  <details><summary>detail</summary>\n\n  ```\n%s\n  ```\n\n  </details>\n\n", strings.TrimRight(f.Detail, "\n"))
		}
	}
	sb.WriteString("\n")
}
