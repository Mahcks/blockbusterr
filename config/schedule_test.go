package config

import "testing"

func TestParseSyncInterval(t *testing.T) {
	for _, interval := range []string{"1h", "1s", "0 8 * * *", "0 0 29 2 *"} {
		if _, _, err := ParseSyncInterval(interval); err != nil {
			t.Errorf("%q: %v", interval, err)
		}
	}
	for _, interval := range []string{"", "0s", "-1s", "nonsense", "0 0 31 2 *", "@every 0s"} {
		if _, _, err := ParseSyncInterval(interval); err == nil {
			t.Errorf("accepted invalid interval %q", interval)
		}
	}
}
