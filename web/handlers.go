// Package web is the HTTP/HTMX presentation layer. It owns form parsing
// and template rendering only -- all financial logic lives in the engine
// package.
package web

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/kaizakin/borrower-copilot/engine"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"inr": engine.FormatINR,
}).ParseFS(templateFS, "templates/*.html"))

type fieldVal struct{ Name, Value string }

func echoFields(r *http.Request, questions []engine.Question) []fieldVal {
	out := make([]fieldVal, 0, len(questions))
	for _, q := range questions {
		out = append(out, fieldVal{q.ID, r.FormValue(q.ID)})
	}
	return out
}

// Index renders the tier-1 (must-answer) questionnaire.
func Index(w http.ResponseWriter, r *http.Request) {
	render(w, "index.html", map[string]any{
		"Questions": engine.CoreQuestions(),
	})
}

// Additional parses the tier-1 answers and renders whichever tier-2
// questions are still relevant to this borrower, carrying the tier-1
// answers forward as hidden fields.
func Additional(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form submission", http.StatusBadRequest)
		return
	}
	core := parseCore(r)
	questions := engine.AdditionalQuestions(core)
	render(w, "additional.html", map[string]any{
		"Questions":  questions,
		"CoreEchoes": echoFields(r, engine.CoreQuestions()),
	})
}

// AssessHandler parses whatever tier-1 and tier-2 fields were submitted
// (tier-2 may be entirely absent if the borrower asked for a quick
// assessment straight from tier 1) and renders the final Assessment,
// including the printable negotiation card.
func AssessHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form submission", http.StatusBadRequest)
		return
	}
	profile := parseCore(r)
	applyAdditional(&profile, r)
	assessment := engine.Assess(profile)

	render(w, "results.html", map[string]any{
		"Profile":    profile,
		"Assessment": assessment,
	})
}

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
