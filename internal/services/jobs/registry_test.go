package jobs

import "testing"

func TestGetAvailableJobTypesFiltersSources(t *testing.T) {
	definitions := GetAvailableJobTypes([]string{"simkl"}, nil)

	if got := definitions["trending"].Sources; len(got) != 1 || got[0] != "simkl" {
		t.Fatalf("trending sources = %v, want [simkl]", got)
	}
	if got := definitions["anticipated"].Sources; len(got) != 0 {
		t.Fatalf("anticipated sources = %v, want none", got)
	}
	if got := JobTypeRegistry["trending"].Sources; len(got) != 3 {
		t.Fatalf("registry was mutated: %v", got)
	}
	if list := definitions["list"]; len(list.Sources) != 0 || len(list.KnownSources) != 4 {
		t.Fatalf("list availability = sources %v, known %v", list.Sources, list.KnownSources)
	}
}
