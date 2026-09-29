package handlers

import (
	"regexp"
	"strings"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func initTestFilters() {
	if !pongo2.FilterExists("truncate") {
		pongo2.RegisterFilter("truncate", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			length := param.Integer()
			if length <= 0 {
				length = 150
			}
			s := in.String()
			runes := []rune(s)
			if len(runes) > length {
				return pongo2.AsValue(string(runes[:length]) + "..."), nil
			}
			return pongo2.AsValue(s), nil
		})
	}
	if !pongo2.FilterExists("striptags") {
		re := regexp.MustCompile(`<[^>]*>`)
		pongo2.RegisterFilter("striptags", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return pongo2.AsValue(re.ReplaceAllString(in.String(), "")), nil
		})
	}
	if !pongo2.FilterExists("split") {
		pongo2.RegisterFilter("split", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			sep := param.String()
			if sep == "" {
				sep = ","
			}
			rawParts := strings.Split(in.String(), sep)
			var res []string
			for _, p := range rawParts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					res = append(res, trimmed)
				}
			}
			return pongo2.AsValue(res), nil
		})
	}
	if !pongo2.FilterExists("strip") {
		pongo2.RegisterFilter("strip", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return pongo2.AsValue(strings.TrimSpace(in.String())), nil
		})
	}
	if !pongo2.FilterExists("trim") {
		pongo2.RegisterFilter("trim", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return pongo2.AsValue(strings.TrimSpace(in.String())), nil
		})
	}

	pongo2.Globals["static_url"] = func(filename string) string {
		return "/static/" + strings.TrimPrefix(filename, "/")
	}
	pongo2.Globals["asset_url"] = func(path string) string {
		return path
	}
	pongo2.Globals["url_for"] = func(endpoint string, args ...string) string {
		return "/" + strings.TrimPrefix(endpoint, "/")
	}
}

func TestProjectsTemplate_Render(t *testing.T) {
	initTestFilters()

	tpl, err := pongo2.FromFile("../../views/projects.html")
	if err != nil {
		t.Fatalf("failed to load projects.html: %v", err)
	}

	out, err := tpl.Execute(pongo2.Context{
		"projects": []map[string]interface{}{
			{
				"Title":       "High-Speed System",
				"Description": "Engineered with Go",
				"TechStack":   "Go Fiber, Redis, C++",
				"Category":    "Case Study",
			},
		},
		"project": map[string]interface{}{
			"Title":       "High-Speed System",
			"Description": "Engineered with Go",
			"TechStack":   "Go Fiber, Redis, C++",
			"Category":    "Case Study",
		},
	})
	if err != nil {
		t.Fatalf("failed to render projects.html: %v", err)
	}

	if !strings.Contains(out, "Go Fiber") || !strings.Contains(out, "Redis") || !strings.Contains(out, "C++") {
		t.Errorf("rendered output missing expected tech tags: %s", out)
	}
}

func TestIndexTemplate_Render(t *testing.T) {
	initTestFilters()

	tpl, err := pongo2.FromFile("../../views/index.html")
	if err != nil {
		t.Fatalf("failed to load index.html: %v", err)
	}

	out, err := tpl.Execute(pongo2.Context{
		"current_path": "/",
		"projects": []map[string]interface{}{
			{
				"Title":       "Fastest Web Engine",
				"Description": "Engineered with Go",
				"TechStack":   "Go, Redis",
				"Category":    "Systems",
			},
		},
		"services": []map[string]interface{}{
			{
				"Name":        "Web Development",
				"Description": "Custom high-performance web development",
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to render index.html: %v", err)
	}

	if !strings.Contains(out, "Inlined Critical CSS") {
		t.Errorf("rendered output missing inlined critical CSS comment: %s", out)
	}
	if !strings.Contains(out, "display-hero") {
		t.Errorf("rendered output missing display-hero class: %s", out)
	}
}

