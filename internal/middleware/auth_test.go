package middleware

import (
	"net/http/httptest"
	"testing"

	"go-api-starter/internal/transport"
)

func TestSkipAuthRequiresSecretAndConfiguredPath(t *testing.T) {
	for _, tc := range []struct {
		name       string
		secret     string
		header     string
		path       string
		wantStatus int
	}{
		{name: "authorized", secret: "test-secret", header: "test-secret", path: "/api/v1/test/callback", wantStatus: 200},
		{name: "wrong secret", secret: "test-secret", header: "wrong", path: "/api/v1/test/callback", wantStatus: 401},
		{name: "missing secret disables bypass", header: "test-secret", path: "/api/v1/test/callback", wantStatus: 401},
		{name: "unlisted path", secret: "test-secret", header: "test-secret", path: "/api/v1/users/me", wantStatus: 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mw := &AuthMiddleware{
				skipAuthHeader: "X-Skip-Auth",
				skipAuthSecret: tc.secret,
				skipAuthPaths:  []string{"/api/v1/test/callback"},
			}
			r := transport.New()
			r.GET(tc.path, mw.RequireAuth(), func(c *transport.Context) {
				if id, ok := c.Get("userID"); !ok || id != uint(1) {
					t.Fatalf("unexpected integration identity: %v", id)
				}
				if token, _ := c.Get("token"); token != "skip-auth" {
					t.Fatalf("unexpected integration token: %v", token)
				}
				c.Status(200)
			})
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Header.Set("X-Skip-Auth", tc.header)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			if res.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tc.wantStatus)
			}
		})
	}
}

func TestSkipAuthPathMatching(t *testing.T) {
	for _, tc := range []struct {
		path, pattern string
		want          bool
	}{
		{"/api/v1/test/callback/retry", "/api/v1/test/callback", true},
		{"/api/v1/test/hooks/a/complete", "/api/v1/test/hooks/*/complete", true},
		{"/api/v1/test/hooks/a/b/complete", "/api/v1/test/hooks/*/complete", false},
		{"/api/v1/test/hooks/a/b/other", "/api/v1/test/hooks/*/complete", false},
		{"/api/v1/test/callbackish", "/api/v1/test/callback", false},
		{"/api/v1/users/me", "", false},
	} {
		if got := matchSkipAuthPath(tc.path, tc.pattern); got != tc.want {
			t.Errorf("matchSkipAuthPath(%q, %q) = %v, want %v", tc.path, tc.pattern, got, tc.want)
		}
	}
}
