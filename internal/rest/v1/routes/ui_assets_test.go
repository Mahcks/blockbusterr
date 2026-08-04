package routes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIRuntimeAssetsAreLocal(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	base := readUIFile(t, filepath.Join(root, "web", "templates", "base.html"))
	activity := readUIFile(t, filepath.Join(root, "web", "templates", "activity.html"))

	for name, content := range map[string]string{"base.html": base, "activity.html": activity} {
		if strings.Contains(content, `<script src="http://`) || strings.Contains(content, `<script src="https://`) {
			t.Fatalf("%s loads a browser library from a public CDN", name)
		}
	}

	for _, reference := range []string{
		`href="/static/css/app.css"`,
		`src="/static/js/vendor/htmx.min.js"`,
		`src="/static/js/vendor/lucide.min.js"`,
		`src="/static/js/app.js"`,
	} {
		if !strings.Contains(base, reference) {
			t.Errorf("base.html is missing %s", reference)
		}
	}
	if !strings.Contains(activity, `src="/static/js/vendor/chart.umd.min.js"`) {
		t.Error("activity.html is missing the local Chart.js reference")
	}

	for _, path := range []string{
		"web/static/css/app.css",
		"web/static/js/app.js",
		"web/static/js/activity.js",
		"web/static/js/filters.js",
		"web/static/js/jobs.js",
		"web/static/js/vendor/htmx.min.js",
		"web/static/js/vendor/lucide.min.js",
		"web/static/js/vendor/chart.umd.min.js",
	} {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil || info.Size() == 0 {
			t.Errorf("required local asset %s is missing or empty", path)
		}
	}
}

func TestFiltersPageUsesExternalScriptAndDelegatedActions(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	template := readUIFile(t, filepath.Join(root, "web", "templates", "filters.html"))
	script := readUIFile(t, filepath.Join(root, "web", "static", "js", "filters.js"))

	if !strings.Contains(template, `defer src="/static/js/filters.js"`) {
		t.Error("filters.html is missing its external script")
	}
	for _, inline := range []string{"onclick=", "onchange=", "onkeyup=", "hx-on=", "<script>"} {
		if strings.Contains(template, inline) {
			t.Errorf("filters.html still contains inline behavior %s", inline)
		}
	}
	for _, expected := range []string{"[data-rules-view]", "function renderRuleList", "function saveRuleSet", "function saveExceptions"} {
		if !strings.Contains(script, expected) {
			t.Errorf("filters.js is missing %s", expected)
		}
	}
	for _, expected := range []string{`id="rule-certification-country"`, `id="rule-unknown-certification"`, `data-chip-field="allowedCertifications"`, `data-chip-field="blockedCertifications"`} {
		if !strings.Contains(template, expected) {
			t.Errorf("filters.html is missing %s", expected)
		}
	}
	for _, expected := range []string{"certification_country", "allowed_certifications", "blocked_certifications", "unknown_certification"} {
		if !strings.Contains(script, expected) {
			t.Errorf("filters.js is missing %s", expected)
		}
	}
}

func TestActivityUsesExternalScriptAndDelegatedActions(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	activity := readUIFile(t, filepath.Join(root, "web", "templates", "activity.html"))
	table := readUIFile(t, filepath.Join(root, "web", "templates", "activity_table.html"))
	script := readUIFile(t, filepath.Join(root, "web", "static", "js", "activity.js"))

	if !strings.Contains(activity, `defer src="/static/js/activity.js"`) {
		t.Error("activity.html is missing its external script")
	}
	for name, content := range map[string]string{"activity.html": activity, "activity_table.html": table} {
		if strings.Contains(content, "onclick=") || strings.Contains(content, "onchange=") || strings.Contains(content, "onkeyup=") || strings.Contains(content, "<script>") {
			t.Errorf("%s still contains inline behavior", name)
		}
	}
	for _, expected := range []string{"function escapeHTML", "function applyFilters", "function renderActivityTimeline", "[data-action]"} {
		if !strings.Contains(script, expected) {
			t.Errorf("activity.js is missing %s", expected)
		}
	}
}

func TestSharedUIFunctionsAndDynamicStylesAreCompiled(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	app := readUIFile(t, filepath.Join(root, "web", "static", "js", "app.js"))
	css := readUIFile(t, filepath.Join(root, "web", "static", "css", "app.css"))

	// testConnection and toggleAdvanced moved to settings.js (the former became
	// Settings-specific inline results; the latter had no remaining callers).
	for _, function := range []string{"togglePassword", "showNotification", "renderLucideIcons"} {
		if !strings.Contains(app, "window."+function) {
			t.Errorf("shared app.js is missing window.%s", function)
		}
	}
	for _, class := range []string{".bg-green-500", ".bg-red-500", ".bg-yellow-500"} {
		if !strings.Contains(css, class) {
			t.Errorf("compiled CSS is missing dynamic class %s", class)
		}
	}
}

func TestSettingsPageUsesExternalScriptAndServerDataAttributes(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	index := readUIFile(t, filepath.Join(root, "web", "templates", "index.html"))
	script := readUIFile(t, filepath.Join(root, "web", "static", "js", "settings.js"))

	if strings.Contains(index, "onclick=") || strings.Contains(index, "onchange=") || strings.Contains(index, "onkeyup=") || strings.Contains(index, "<script>") {
		t.Error("index.html still contains inline behavior")
	}
	for _, expected := range []string{
		`id="settings-page"`,
		`id="settings-form"`,
		`id="save-bar"`,
		`defer src="/static/js/settings.js"`,
	} {
		if !strings.Contains(index, expected) {
			t.Errorf("index.html is missing %s", expected)
		}
	}
	for _, expected := range []string{
		"function saveSettings",
		"function discardChanges",
		"function testConnection",
		"function renderServiceStatuses",
		"function updateWeightTotal",
		"function renderLimitsSummary",
		"[data-action]",
	} {
		if !strings.Contains(script, expected) {
			t.Errorf("settings.js is missing %s", expected)
		}
	}
	// Every form field name the /config/save handler parses must still be
	// present so the redesign never silently drops a setting.
	for _, field := range []string{
		"trakt.client_id", "trakt.client_secret", "tmdb.api_key", "simkl.client_id", "mdblist.api_key", "letterboxd.experimental_scraping",
		"radarr.url", "radarr.api_key", "radarr.quality_profile", "radarr.root_folder", "radarr.minimum_availability", "radarr.monitor",
		"sonarr.url", "sonarr.api_key", "sonarr.quality_profile", "sonarr.root_folder", "sonarr.monitor",
		"jellyseerr.url", "jellyseerr.api_key", "jellyseerr.user_id", "jellyseerr.request_credentials.email", "jellyseerr.request_credentials.password",
		"jobs.mode", "jobs.sync_interval", "jobs.global_limit_movies", "jobs.global_limit_shows", "jobs.global_period",
		"jobs.selection.enabled", "jobs.selection.sync_interval", "jobs.selection.movie_limit", "jobs.selection.show_limit",
		"scoring.enabled", "scoring.rating_weight", "scoring.popularity_weight", "scoring.recency_weight", "scoring.rating_scale", "scoring.popularity_metric", "scoring.recency_days",
	} {
		if !strings.Contains(index, `name="`+field+`"`) {
			t.Errorf("index.html is missing form field %s", field)
		}
	}
}

func TestJobsPageUsesExternalScriptAndServerDataAttributes(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	jobs := readUIFile(t, filepath.Join(root, "web", "templates", "jobs.html"))
	alerts := readUIFile(t, filepath.Join(root, "web", "templates", "components", "alert.html"))
	script := readUIFile(t, filepath.Join(root, "web", "static", "js", "jobs.js"))

	for _, expected := range []string{
		`id="jobs-page"`,
		`data-global-mode="{{.Config.Jobs.Mode}}"`,
		`defer src="/static/js/jobs.js"`,
	} {
		if !strings.Contains(jobs, expected) {
			t.Errorf("jobs.html is missing %s", expected)
		}
	}
	if strings.Contains(jobs, "const globalMode") {
		t.Error("jobs.html still contains inline page state")
	}
	if strings.Contains(jobs, "onclick=") || strings.Contains(jobs, "onchange=") {
		t.Error("jobs.html still contains inline event handlers")
	}
	if strings.Contains(alerts, "onclick=") || strings.Contains(alerts, "<script>") {
		t.Error("shared alerts still contain inline behavior")
	}
	if got := strings.Count(jobs, `role="dialog"`); got != 3 {
		t.Errorf("jobs.html has %d accessible dialogs, want 3", got)
	}
	for _, expected := range []string{"const globalMode", "async function loadData", "function renderJobsList", "function escapeHTML"} {
		if !strings.Contains(script, expected) {
			t.Errorf("jobs.js is missing %s", expected)
		}
	}
	for _, expected := range []string{"function openDialog", "function closeDialog", "function handleDialogKeyboard", "function handleJobsAction"} {
		if !strings.Contains(script, expected) {
			t.Errorf("jobs.js is missing %s", expected)
		}
	}
	for _, expected := range []string{`id="modal-rule-set"`, "function populateJobFilters", "rule_set_id", "/v1/rule-sets", `id="modal-delivery-limit"`, "delivery_limit", `id="import-job-file"`, `data-action="export-job"`, "function importJobBundle", "function exportCurrentJob"} {
		if !strings.Contains(jobs+script, expected) {
			t.Errorf("per-job filter editor is missing %s", expected)
		}
	}
	for _, expected := range []string{
		"fetch(`/v1/jobs/dynamic/${encodeURIComponent(jobId)}`)",
		"job.selection_cycle === true",
		"const savedJob = await response.json()",
	} {
		if !strings.Contains(script, expected) {
			t.Errorf("job editor isolation guard is missing %s", expected)
		}
	}
}

func readUIFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
