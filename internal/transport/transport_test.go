package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSOriginDefaultsAndOverrides(t *testing.T) {
	for _, tt := range []struct {
		name         string
		origins      []string
		origin, want string
	}{
		{"omitted", nil, "https://app.example", "*"},
		{"empty", []string{}, "https://other.example", "*"},
		{"wildcard", []string{"*"}, "null", "*"},
		{"explicit allowed", []string{"https://app.example"}, "https://app.example", "https://app.example"},
		{"explicit denied", []string{"https://app.example"}, "https://other.example", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			cors := CORS(tt.origins, []string{"GET", "OPTIONS"}, []string{"Authorization"})
			r.Use(cors)
			r.GET("/resource", func(c *Context) { c.Status(http.StatusOK) })
			r.OPTIONS("/*", cors)
			for _, method := range []string{http.MethodGet, http.MethodOptions} {
				req := httptest.NewRequest(method, "/resource", nil)
				req.Header.Set("Origin", tt.origin)
				req.Header.Set("Access-Control-Request-Method", "GET")
				req.Header.Set("Access-Control-Request-Headers", "Authorization")
				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)
				if got := res.Header().Get("Access-Control-Allow-Origin"); got != tt.want {
					t.Fatalf("%s allow origin = %q, want %q", method, got, tt.want)
				}
				if method == http.MethodOptions && res.Code != http.StatusNoContent {
					t.Fatalf("preflight status = %d, want 204", res.Code)
				}
			}
		})
	}
}
