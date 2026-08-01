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

func TestSharedUIFunctionsAndDynamicStylesAreCompiled(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	app := readUIFile(t, filepath.Join(root, "web", "static", "js", "app.js"))
	css := readUIFile(t, filepath.Join(root, "web", "static", "css", "app.css"))

	for _, function := range []string{"togglePassword", "showNotification", "renderLucideIcons", "toggleAdvanced", "testConnection"} {
		if !strings.Contains(app, "window."+function) {
			t.Errorf("shared app.js is missing window.%s", function)
		}
	}
	for _, class := range []string{".bg-green-500", ".bg-red-500", ".bg-yellow-500", ".from-yellow-900\\/50", ".to-slate-800\\/50"} {
		if !strings.Contains(css, class) {
			t.Errorf("compiled CSS is missing dynamic class %s", class)
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
}

func readUIFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
