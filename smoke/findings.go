package smoke

import "fmt"

// Finding is one problem a scenario found (a failure) or one thing worth
// knowing that does not fail the run (a warning).
type Finding struct {
	Check   string
	Message string
	// Detail is long text for the full report only (a state dump, a stack).
	Detail string
}

// Section is one titled block of Markdown in the full report.
type Section struct {
	Title string
	Body  string
}

// Findings is what a check reports into: a one-line summary, the failures
// and warnings it found and the sections of its report. The static checks,
// the docs check and the accounts check each fill one in.
type Findings struct {
	Summary  string
	Failures []Finding
	Warnings []Finding
	Sections []Section
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
