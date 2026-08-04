package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
)

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

func TestJobTemplatesAreValidRecipes(t *testing.T) {
	cfg := &config.Config{}
	cfg.MigrateRuleSets()
	seen := map[string]bool{}
	for _, recipe := range JobTemplates {
		if recipe.ID == "" || seen[recipe.ID] || recipe.Version < 1 {
			t.Fatalf("invalid recipe identity: %+v", recipe)
		}
		seen[recipe.ID] = true
		if !SupportsMediaType(recipe.Type, recipe.MediaType) || (recipe.Type != "list" && !SupportsSource(recipe.Type, recipe.Source)) {
			t.Fatalf("unsupported recipe job: %+v", recipe)
		}
		if recipe.SyncInterval == "" {
			t.Fatalf("recipe %s has no cadence", recipe.ID)
		}
		if _, err := time.ParseDuration(recipe.SyncInterval); err != nil {
			t.Fatalf("recipe %s cadence: %v", recipe.ID, err)
		}
		encoded, err := json.Marshal(AvailableJobTemplate{JobTemplate: recipe, Ready: true})
		if err != nil {
			t.Fatalf("marshal recipe %s: %v", recipe.ID, err)
		}
		var decoded AvailableJobTemplate
		if err := json.Unmarshal(encoded, &decoded); err != nil || decoded.ID != recipe.ID || decoded.Source != recipe.Source || decoded.SyncInterval != recipe.SyncInterval || decoded.DeliveryLimit != recipe.DeliveryLimit {
			t.Fatalf("recipe %s JSON round trip: %+v, %v", recipe.ID, decoded, err)
		}
		rules := config.RuleSet{ID: recipe.ID, Name: recipe.RuleSetName, Media: recipe.MediaType, Revision: 1, Movies: recipe.Movies, Shows: recipe.Shows}
		config.ApplyRuleSetDefaults(&rules)
		if err := cfg.ValidateRuleSet(rules, ""); err != nil {
			t.Fatalf("recipe %s rules: %v", recipe.ID, err)
		}
	}
}
