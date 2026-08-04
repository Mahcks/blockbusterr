package config

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestListLocatorJSONAndYAMLRoundTrip(t *testing.T) {
	want := DynamicJob{Type: "list", Source: "trakt", List: &ListLocator{Kind: "public_list", Owner: "mahcks", Slug: "favorites", Ordering: "source"}}
	for name, marshal := range map[string]func(any) ([]byte, error){"JSON": json.Marshal, "YAML": yaml.Marshal} {
		t.Run(name, func(t *testing.T) {
			data, err := marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var got DynamicJob
			if name == "JSON" {
				err = json.Unmarshal(data, &got)
			} else {
				err = yaml.Unmarshal(data, &got)
			}
			if err != nil || got.List == nil || *got.List != *want.List {
				t.Fatalf("round trip = %+v, %v", got.List, err)
			}
		})
	}
}
