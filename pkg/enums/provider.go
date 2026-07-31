package enums

type DiscoveryProvider string

const (
	DiscoveryProviderTrakt DiscoveryProvider = "trakt"
	DiscoveryProviderTMDB  DiscoveryProvider = "tmdb"
	DiscoveryProviderSimkl DiscoveryProvider = "simkl"
)

func (provider DiscoveryProvider) Valid() bool {
	switch provider {
	case DiscoveryProviderTrakt, DiscoveryProviderTMDB, DiscoveryProviderSimkl:
		return true
	default:
		return false
	}
}
