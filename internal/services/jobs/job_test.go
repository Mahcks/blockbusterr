package jobs

import "testing"

func TestDryRunEnabled(t *testing.T) {
	t.Setenv("BLOCKBUSTERR_DRY_RUN", "")
	if !DryRunEnabled("dev") {
		t.Fatal("development builds must be dry-run")
	}
	if DryRunEnabled("v2.0.0-beta.4") {
		t.Fatal("release builds must deliver by default")
	}

	t.Setenv("BLOCKBUSTERR_DRY_RUN", "true")
	if !DryRunEnabled("v2.0.0-beta.4") {
		t.Fatal("BLOCKBUSTERR_DRY_RUN=true must enable dry-run")
	}

	t.Setenv("BLOCKBUSTERR_DRY_RUN", "invalid")
	if DryRunEnabled("v2.0.0-beta.4") {
		t.Fatal("invalid values must not enable dry-run")
	}
}
