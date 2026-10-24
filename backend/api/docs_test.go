package api

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
)

func loadSpec(t *testing.T) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal(Spec(), &doc); err != nil {
		t.Fatalf("openapi.yaml is not valid YAML: %v", err)
	}
	return doc
}

// walk calls fn for every value in the document.
func walk(v any, fn func(any)) {
	fn(v)
	switch x := v.(type) {
	case map[string]any:
		for _, child := range x {
			walk(child, fn)
		}
	case []any:
		for _, child := range x {
			walk(child, fn)
		}
	}
}

func TestSpecHasTheExpectedShape(t *testing.T) {
	doc := loadSpec(t)
	if v, _ := doc["openapi"].(string); !strings.HasPrefix(v, "3.1") {
		t.Fatalf("openapi version = %q, want 3.1.x", v)
	}
	if paths, _ := doc["paths"].(map[string]any); len(paths) == 0 {
		t.Fatal("the spec documents no paths")
	}
}

func TestEveryReferenceResolves(t *testing.T) {
	doc := loadSpec(t)
	walk(doc, func(v any) {
		m, ok := v.(map[string]any)
		if !ok {
			return
		}
		ref, ok := m["$ref"].(string)
		if !ok {
			return
		}
		node := any(doc)
		for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
			next, ok := node.(map[string]any)[part]
			if !ok {
				t.Errorf("unresolved reference %s", ref)
				return
			}
			node = next
		}
	})
}

func TestOperationsAreIdentifiedAndDocumented(t *testing.T) {
	doc := loadSpec(t)
	seen := map[string]string{}
	for path, item := range doc["paths"].(map[string]any) {
		for method, op := range item.(map[string]any) {
			if method == "parameters" || method == "servers" {
				continue
			}
			where := strings.ToUpper(method) + " " + path
			o := op.(map[string]any)
			id, _ := o["operationId"].(string)
			if id == "" {
				t.Errorf("%s has no operationId", where)
			} else if prev, dup := seen[id]; dup {
				t.Errorf("operationId %q is used by both %s and %s", id, prev, where)
			} else {
				seen[id] = where
			}
			if s, _ := o["summary"].(string); s == "" {
				t.Errorf("%s has no summary", where)
			}
			if tags, _ := o["tags"].([]any); len(tags) == 0 {
				t.Errorf("%s has no tag", where)
			}
			if r, _ := o["responses"].(map[string]any); len(r) == 0 {
				t.Errorf("%s documents no responses", where)
			}
		}
	}
}

func TestPublicEndpointsAreExplicit(t *testing.T) {
	doc := loadSpec(t)
	// Everything is protected by default; these are the only operations that opt out.
	want := []string{
		"POST /auth/login", "POST /auth/logout", "POST /auth/register",
		"POST /auth/verify-email/{token}",
		"GET /health", "GET /ready", "GET /ws",
	}
	var got []string
	for path, item := range doc["paths"].(map[string]any) {
		for method, op := range item.(map[string]any) {
			o, ok := op.(map[string]any)
			if !ok {
				continue
			}
			if sec, ok := o["security"].([]any); ok && len(sec) == 0 {
				got = append(got, strings.ToUpper(method)+" "+path)
			}
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("public operations changed.\n got: %v\nwant: %v", got, want)
	}
}

func TestDocsRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterDocs(router)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Body.String(), "openapi: 3.1") {
		t.Fatalf("/openapi.yaml: status %d, body starts %q", rec.Code, rec.Body.String()[:20])
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "data-url=\"/openapi.yaml\"") || !strings.Contains(body, "@scalar/api-reference@"+scalarVersion) {
		t.Fatalf("/docs: status %d, page does not load Scalar with the contract", rec.Code)
	}
}
