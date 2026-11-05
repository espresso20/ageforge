package config

import (
	"testing"
)

func TestIsResourceEnded_AlwaysAvailable(t *testing.T) {
	for _, res := range []string{"food", "faith", "culture", "gold"} {
		for _, age := range AgeOrder() {
			if IsResourceEnded(res, age) {
				t.Errorf("resource %q should never be ended, but IsResourceEnded reported true in age %q", res, age)
			}
		}
	}
}

func TestIsResourceEnded_Wood(t *testing.T) {
	pos := AgePositions(AgeOrder())
	renPos := pos["renaissance_age"]
	colPos := pos["colonial_age"]

	for _, age := range AgeOrder() {
		p := pos[age]
		ended := IsResourceEnded("wood", age)
		if p <= renPos && ended {
			t.Errorf("wood should be available up to renaissance_age, but ended in %q", age)
		}
		if p >= colPos && !ended {
			t.Errorf("wood should be ended in colonial_age through transcendent_age, but not ended in %q", age)
		}
	}
}

func TestIsResourceEnded_StoneIronUranium(t *testing.T) {
	pos := AgePositions(AgeOrder())
	infoPos := pos["information_age"]

	for _, res := range []string{"stone", "iron", "uranium"} {
		for _, age := range AgeOrder() {
			p := pos[age]
			ended := IsResourceEnded(res, age)
			if p < infoPos && ended {
				t.Errorf("%s should be available before information_age, but ended in %q", res, age)
			}
			if p >= infoPos && !ended {
				t.Errorf("%s should be ended from information_age onward, but not ended in %q", res, age)
			}
		}
	}
}

func TestIsResourceEnded_Crypto(t *testing.T) {
	pos := AgePositions(AgeOrder())
	spacePos := pos["space_age"]

	for _, age := range AgeOrder() {
		p := pos[age]
		ended := IsResourceEnded("crypto", age)
		if p < spacePos && ended {
			t.Errorf("crypto should be available before space_age, but ended in %q", age)
		}
		if p >= spacePos && !ended {
			t.Errorf("crypto should be ended from space_age onward, but not ended in %q", age)
		}
	}
}
