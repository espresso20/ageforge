package smoke

import "fmt"

// Finding is one problem a scenario found (a failure) or one thing worth
// knowing that does not fail the run (a warning).
type Finding struct {
	Check   string `json:"check"`
	Message string `json:"message"`
	Seed    int64  `json:"seed,omitempty"`
	// Repro is how to reproduce it: a command line, a seed and command list.
	Repro string `json:"repro,omitempty"`
	// Detail is long text for the full report only (a state dump, a stack).
	Detail string `json:"detail,omitempty"`
}

// Section is one titled block of Markdown in the full report.
type Section struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Findings is what a check reports into: a one-line summary, the failures
// and warnings it found and the sections of its report. A scenario's Result
// embeds one; the static checks and the docs check fill it in without
// knowing anything of the scenario runner around them.
type Findings struct {
	Summary  string    `json:"summary"`
	Failures []Finding `json:"failures"`
	Warnings []Finding `json:"warnings,omitempty"`
	Sections []Section `json:"-"`
}

func (r *Findings) fail(check, format string, args ...interface{}) *Finding {
	r.Failures = append(r.Failures, Finding{Check: check, Message: fmt.Sprintf(format, args...)})
	return &r.Failures[len(r.Failures)-1]
}

func (r *Findings) warn(check, format string, args ...interface{}) *Finding {
	r.Warnings = append(r.Warnings, Finding{Check: check, Message: fmt.Sprintf(format, args...)})
	return &r.Warnings[len(r.Warnings)-1]
}

func (r *Findings) section(title, format string, args ...interface{}) {
	r.Sections = append(r.Sections, Section{Title: title, Body: fmt.Sprintf(format, args...)})
}
