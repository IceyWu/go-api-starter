package docs

import (
	"encoding/json"
	"testing"
)

func TestReadDocContainsSystemPing(t *testing.T) {
	var document struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal([]byte(ReadDoc()), &document); err != nil {
		t.Fatalf("checked-in OpenAPI document is invalid JSON: %v", err)
	}
	if _, ok := document.Paths["/api/v1/system/ping"]; !ok {
		t.Fatal("checked-in OpenAPI document is missing system ping")
	}
}

func TestReadDocForRequestUsesPublicHost(t *testing.T) {
	var document struct {
		Host     string   `json:"host"`
		BasePath string   `json:"basePath"`
		Schemes  []string `json:"schemes"`
	}
	if err := json.Unmarshal([]byte(ReadDocForRequest("10.0.30.194:9527", "/api", "http")), &document); err != nil {
		t.Fatalf("request OpenAPI document is invalid JSON: %v", err)
	}
	if document.Host != "10.0.30.194:9527" || document.BasePath != "/api" || len(document.Schemes) != 1 || document.Schemes[0] != "http" {
		t.Fatalf("unexpected request OpenAPI metadata: %+v", document)
	}
}

func TestReadDocContainsPaginationDefaultsAndDescriptions(t *testing.T) {
	var document struct {
		Paths map[string]map[string]struct {
			Description string `json:"description"`
			Parameters  []struct {
				Name    string `json:"name"`
				In      string `json:"in"`
				Default any    `json:"default"`
			} `json:"parameters"`
		} `json:"paths"`
	}
	if err := json.Unmarshal([]byte(ReadDoc()), &document); err != nil {
		t.Fatalf("checked-in OpenAPI document is invalid JSON: %v", err)
	}

	for _, path := range []string{"/api/v1/files", "/api/v1/users"} {
		operation, ok := document.Paths[path]["get"]
		if !ok {
			t.Fatalf("missing GET operation for %s", path)
		}
		defaults := map[string]any{"page": float64(1), "page_size": float64(10), "sort": "id desc"}
		for name, want := range defaults {
			found := false
			for _, parameter := range operation.Parameters {
				if parameter.In == "query" && parameter.Name == name {
					found = true
					if parameter.Default != want {
						t.Errorf("%s %s default = %v, want %v", path, name, parameter.Default, want)
					}
				}
			}
			if !found {
				t.Errorf("%s is missing query parameter %q", path, name)
			}
		}
	}

	for _, path := range []string{"/api/v1/users/me"} {
		for method, operation := range document.Paths[path] {
			if operation.Description == "" {
				t.Errorf("%s %s is missing an operation description", method, path)
			}
		}
	}
}
