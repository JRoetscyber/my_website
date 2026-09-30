package handlers

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/flosch/pongo2/v6"
)

func initTestFilters() {
	if !pongo2.FilterExists("markdown") {
		pongo2.RegisterFilter("markdown", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return pongo2.AsSafeValue("<p>" + in.String() + "</p>"), nil
		})
	}
	if !pongo2.FilterExists("from_json") {
		pongo2.RegisterFilter("from_json", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return pongo2.AsValue([]interface{}{}), nil
		})
	}
	if !pongo2.FilterExists("tojson") {
		pongo2.RegisterFilter("tojson", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return pongo2.AsSafeValue("{}"), nil
		})
	}
	if !pongo2.FilterExists("selectattr") {
		pongo2.RegisterFilter("selectattr", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
			return in, nil
		})
	}
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

func TestAdminAndPublicTemplates_Render(t *testing.T) {
	initTestFilters()

	// 1. login.html
	if tpl, err := pongo2.FromFile("../../views/login.html"); err != nil {
		t.Fatalf("failed to load login.html: %v", err)
	} else if _, err := tpl.Execute(pongo2.Context{"current_path": "/login"}); err != nil {
		t.Fatalf("failed to render login.html: %v", err)
	}

	// 2. order_display.html
	if tpl, err := pongo2.FromFile("../../views/order_display.html"); err != nil {
		t.Fatalf("failed to load order_display.html: %v", err)
	} else if _, err := tpl.Execute(pongo2.Context{"preparing_orders": []map[string]interface{}{{"SequenceNumber": 101}}, "ready_orders": []map[string]interface{}{{"SequenceNumber": 100}}}); err != nil {
		t.Fatalf("failed to render order_display.html: %v", err)
	}

	// 3. order_tracker.html
	if tpl, err := pongo2.FromFile("../../views/order_tracker.html"); err != nil {
		t.Fatalf("failed to load order_tracker.html: %v", err)
	} else if _, err := tpl.Execute(pongo2.Context{
		"order": map[string]interface{}{
			"SequenceNumber": 105,
			"CustomerName":   "Alice",
			"Status":         "in_progress",
			"TrackingCode":   "trk-12345",
			"NotifyCount":    0,
			"Items":          "1x Custom Web App",
			"TotalAmount":    5000.0,
			"CreatedAt":      time.Now(),
		},
	}); err != nil {
		t.Fatalf("failed to render order_tracker.html: %v", err)
	}

	// 4. app_development.html
	if tpl, err := pongo2.FromFile("../../views/app_development.html"); err != nil {
		t.Fatalf("failed to load app_development.html: %v", err)
	} else if _, err := tpl.Execute(pongo2.Context{"current_path": "/app-development"}); err != nil {
		t.Fatalf("failed to render app_development.html: %v", err)
	}

	// 5. admin views
	loader := pongo2.MustNewLocalFileSystemLoader("../../views")
	set := pongo2.NewSet("adminViewsSet", loader)

	adminViews := []string{
		"admin/orders.html",
		"admin/automation.html",
		"admin/blogs.html",
		"admin/booking.html",
		"admin/faqs.html",
		"admin/services.html",
	}

	for _, v := range adminViews {
		tpl, err := set.FromFile(v)
		if err != nil {
			t.Fatalf("failed to load %s: %v", v, err)
		}
		_, err = tpl.Execute(pongo2.Context{
			"orders": []map[string]interface{}{
				{"ID": 1, "SequenceNumber": 101, "CustomerName": "Bob", "Status": "received", "CreatedAt": time.Now(), "UpdatedAt": time.Now()},
			},
			"blogs": []map[string]interface{}{
				{"ID": 1, "Title": "Test Blog", "Slug": "test-blog", "Status": "published", "PublishedAt": time.Now(), "CreatedAt": time.Now(), "UpdatedAt": time.Now()},
			},
			"faqs": []map[string]interface{}{
				{"ID": 1, "Question": "What is Go?", "Answer": "A fast language", "CreatedAt": time.Now(), "UpdatedAt": time.Now()},
			},
			"services": []map[string]interface{}{
				{"ID": 1, "Name": "Cloud Architecture", "Slug": "cloud", "CreatedAt": time.Now(), "UpdatedAt": time.Now()},
			},
		})
		if err != nil {
			t.Fatalf("failed to render %s: %v", v, err)
		}
	}
}


