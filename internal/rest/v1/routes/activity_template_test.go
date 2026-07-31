package routes

import (
	"html/template"
	"io"
	"testing"

	"github.com/mahcks/blockbusterr/internal/database"
)

func TestActivityTableSupportsLanguage(t *testing.T) {
	tmpl, err := template.New("activity_table").Funcs(template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}).ParseFiles("../../../../web/templates/activity_table.html")
	if err != nil {
		t.Fatal(err)
	}

	data := map[string]any{
		"Page":       1,
		"TotalPages": 1,
		"Logs": []activityLogGroup{{
			Log: database.ActivityLog{Title: "Test", Language: "en"},
		}},
	}
	if err := tmpl.ExecuteTemplate(io.Discard, "activity_table.html", data); err != nil {
		t.Fatal(err)
	}
}
