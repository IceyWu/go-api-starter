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
