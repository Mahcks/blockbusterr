package jobs

import "testing"

func TestGetAvailableJobTypesFiltersSources(t *testing.T) {
	definitions := GetAvailableJobTypes([]string{"simkl"})

	if got := definitions["trending"].Sources; len(got) != 1 || got[0] != "simkl" {
		t.Fatalf("trending sources = %v, want [simkl]", got)
	}
	if got := definitions["anticipated"].Sources; len(got) != 0 {
		t.Fatalf("anticipated sources = %v, want none", got)
	}
	if got := JobTypeRegistry["trending"].Sources; len(got) != 3 {
		t.Fatalf("registry was mutated: %v", got)
	}
}
