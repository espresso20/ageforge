package game

import "time"

// A run ends in a prestige (plain, or from the final era past the Last
// Passage) or in a Succumb to an era's catastrophe. Either way the town is
// given up and a new run starts in the first age. RunEnding is what the run
// that ended leaves to be told: how it ended, where, what it earned, and
// the lines the game writes about it, each marked by what it is so that a
// screen can place them without reading their wording.
//
// It rides the EventRunEnded bus event, published at the end of the reset
// under the engine's write lock (a handler takes the payload and nothing
// else). It is not part of the game's state and is not saved.

// The ways a run ends (RunEnding.Kind).
const (
	RunEndPrestige  = "prestige"  // a prestige with nothing in its way
	RunEndSpared    = "spared"    // a prestige from the final era that the Last Passage let by
	RunEndEndured   = "endured"   // a prestige through the Last Passage, endured: part of the points
	RunEndSuccumbed = "succumbed" // a prestige the Last Passage took the points of, for the Cosmic Legacy
	RunEndFallen    = "fallen"    // a Succumb to an era's catastrophe: no prestige
)

// The kinds of line an ending reports (RunEndingLine.Kind).
const (
	EndLineVerdict   = "verdict"   // what the catastrophe or the Last Passage did
	EndLineVoice     = "voice"     // the run's closing line, in the voice of its age
	EndLineComplete  = "complete"  // the prestige itself: its level and its points
	EndLineEarly     = "early"     // an early prestige pays little, and what a deeper one pays
	EndLineLegacy    = "legacy"    // a legacy earned or carried: the Cosmic Legacy, an era's bonus
	EndLineMastery   = "mastery"   // the ages that gained a mastery level
	EndLineGround    = "ground"    // how the first age runs for it
	EndLineKnowledge = "knowledge" // Ancient Knowledge
	EndLineRuins     = "ruins"     // ruins carried forward
	EndLineKit       = "kit"       // what the legacy kit set up for the new run
)

// RunEndingLine is one line the game wrote about an ending: its kind, the
// log category it was written in and its text, tags and all.
type RunEndingLine struct {
	Kind  string
	Level string
	Text  string
}

// RunEnding is the record of a run that just ended.
type RunEnding struct {
	Kind string
	// Age is the age the run ended in.
	Age string
	// Level is the prestige level after it.
	Level int
	// Points is the prestige points the run paid; Full is what it would
	// have paid with nothing taken (they differ after the Last Passage).
	Points, Full int
	// Badges is how many badges the account earned during the run, -1 when
	// that cannot be told (no account, or a run with no start on record).
	Badges int
	// Catastrophe names what struck: the era's catastrophe for a fall, the
	// Last Passage for a prestige it came to. "" otherwise.
	Catastrophe string
	Lines       []RunEndingLine
}

// Prestige reports whether the run ended in a prestige (of any kind).
func (e RunEnding) Prestige() bool { return e.Kind != RunEndFallen }

// Came reports whether a catastrophe is part of the ending: a fall, or a
// prestige the Last Passage came to.
func (e RunEnding) Came() bool {
	return e.Kind == RunEndFallen || e.Kind == RunEndEndured || e.Kind == RunEndSuccumbed
}

// endingKind is the RunEnding.Kind of a prestige.
func (how prestigeEnding) endingKind() string {
	switch how {
	case lastPassageSpared:
		return RunEndSpared
	case lastPassageEndured:
		return RunEndEndured
	case lastPassageSuccumbed:
		return RunEndSuccumbed
	}
	return RunEndPrestige
}

// say writes a line of the ending to the log and to the record.
func (ge *GameEngine) sayEnding(end *RunEnding, kind, level, text string) {
	ge.addLog(level, text)
	end.Lines = append(end.Lines, RunEndingLine{Kind: kind, Level: level, Text: text})
}

// takeEndingLines adds to the record the log lines written since the log
// held from entries, all of one kind.
func (ge *GameEngine) takeEndingLines(end *RunEnding, from int, kind string) {
	for _, l := range ge.log[min(from, len(ge.log)):] {
		end.Lines = append(end.Lines, RunEndingLine{Kind: kind, Level: l.Type, Text: l.Message})
	}
}

// publishRunEnded counts the run's badges and publishes the record. Under
// the write lock, as the last step of the reset: the new run is in place.
func (ge *GameEngine) publishRunEnded(end RunEnding, runStarted time.Time) {
	end.Badges = -1
	if acct := ge.accountForRecordsLocked(); acct != nil && !runStarted.IsZero() {
		end.Badges = acct.BadgesEarnedSince(runStarted)
	}
	ge.Bus.Publish(EventData{Type: EventRunEnded, Payload: map[string]interface{}{"ending": end}})
}
