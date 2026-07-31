package enums

import "testing"

func TestDiscoveryProviderValid(t *testing.T) {
	for _, provider := range []DiscoveryProvider{DiscoveryProviderTrakt, DiscoveryProviderTMDB, DiscoveryProviderSimkl} {
		if !provider.Valid() {
			t.Fatalf("expected %q to be valid", provider)
		}
	}
	if DiscoveryProvider("other").Valid() {
		t.Fatal("unexpected provider accepted")
	}
}
