package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// templateDir is set at startup by LoadTemplatesFromDir.
var templateDir string

// pageCache caches per-page parsed templates (layout + page + pagination).
var pageCache = map[string]*template.Template{}

// partialCache caches per-partial parsed templates.
var partialCache = map[string]*template.Template{}

// LoadTemplatesFromDir sets the template root and warms the caches.
func LoadTemplatesFromDir(dir string) error {
	templateDir = dir
	pages := []string{"dashboard", "artists", "albums", "tracks", "playlists", "playlist_detail"}
	partials := []string{"artists", "albums", "tracks", "playlists"}

	for _, name := range pages {
		files := []string{
			filepath.Join(dir, "layout.html"),
			filepath.Join(dir, "pagination.html"),
			filepath.Join(dir, name+".html"),
		}
		// Some pages also need their partial for the initial render.
		if name != "dashboard" && name != "playlist_detail" {
			files = append(files, filepath.Join(dir, name+"_partial.html"))
		}
		t, err := template.New("layout.html").Funcs(funcMap()).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parse page %s: %w", name, err)
		}
		pageCache[name] = t
	}

	for _, name := range partials {
		files := []string{
			filepath.Join(dir, "pagination.html"),
			filepath.Join(dir, name+"_partial.html"),
		}
		t, err := template.New(name+"_partial.html").Funcs(funcMap()).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parse partial %s: %w", name, err)
		}
		partialCache[name] = t
	}

	return nil
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"formatDuration": func(ms int) string {
			total := ms / 1000
			m, s := total/60, total%60
			return fmt.Sprintf("%d:%02d", m, s)
		},
		"prevOffset": func(page, limit int) int {
			return (page - 2) * limit
		},
		"nextOffset": func(page, limit int) int {
			return page * limit
		},
	}
}

// renderPage renders a full page (layout + content).
// The page template must define {{define "content"}}...{{end}}.
func renderPage(w http.ResponseWriter, page string, data any) {
	t, ok := pageCache[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// renderPartial renders an HTMX partial fragment.
func renderPartial(w http.ResponseWriter, partial string, data any) {
	t, ok := partialCache[partial]
	if !ok {
		http.Error(w, "partial not found: "+partial, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, partial+"_partial.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// isHTMX returns true if the request came from HTMX.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
