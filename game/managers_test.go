package game

import "github.com/espresso20/ageforge/rules"

// The managers on the core ruleset, for tests that build one on its own.

// NewBuildingManager creates a building manager on the core ruleset.
func NewBuildingManager() *BuildingManager { return NewBuildingManagerWith(rules.Core()) }

// NewDiplomacyManager creates a new diplomacy manager on the core ruleset.
func NewDiplomacyManager() *DiplomacyManager { return NewDiplomacyManagerWith(rules.Core()) }

// NewEventManager creates a new event manager on the core ruleset.
func NewEventManager() *EventManager { return NewEventManagerWith(rules.Core()) }

// NewMilitaryManager creates a military manager on the core ruleset.
func NewMilitaryManager() *MilitaryManager { return NewMilitaryManagerWith(rules.Core()) }

// NewMilestoneManager creates a new milestone manager on the core ruleset.
func NewMilestoneManager() *MilestoneManager { return NewMilestoneManagerWith(rules.Core()) }

// NewProgressManager creates a new progress manager on the core ruleset.
func NewProgressManager() *ProgressManager { return NewProgressManagerWith(rules.Core()) }

// NewPrestigeManager creates a new prestige manager on the core ruleset.
func NewPrestigeManager() *PrestigeManager { return NewPrestigeManagerWith(rules.Core()) }

// NewResearchManager creates a new research manager on the core ruleset.
func NewResearchManager() *ResearchManager { return NewResearchManagerWith(rules.Core()) }

// NewResourceManager creates a resource manager on the core ruleset.
func NewResourceManager() *ResourceManager { return NewResourceManagerWith(rules.Core()) }

// NewWorkerManager creates a new single-pool worker manager on the core
// ruleset.
func NewWorkerManager() *WorkerManager { return NewWorkerManagerWith(rules.Core()) }
