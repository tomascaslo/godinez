package godinez

import (
	"html/template"
	"path/filepath"
	"time"
)

type TemplateData struct {
	CSRFToken       string // Used to add CSRFToken to template
	CurrentYear     int // The current year e.g. 2019
	Flash           string // Flash message to show on website
	IsAuthenticated bool
}

func HumanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": HumanDate,
}

func NewTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Need to check if there are layout files first
		globFilepath := filepath.Join(dir, "*.layout.tmpl")
		matches, err := filepath.Glob(globFilepath)
		if err != nil {
			return nil, err
		}
		if matches != nil && len(matches) != 0 {
			ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
			if err != nil {
				return nil, err
			}
		}

		// Need to check if there are partial files first
		globFilepath = filepath.Join(dir, "*.partial.tmpl")
		matches, err = filepath.Glob(globFilepath)
		if err != nil {
			return nil, err
		}
		if matches != nil && len(matches) != 0 {
			ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
			if err != nil {
				return nil, err
			}
		}

		cache[name] = ts
	}

	return cache, nil
}
