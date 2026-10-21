package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moviehub/auth"
)

func TestRequireLoginAndRole(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("0123456789abcdef0123456789abcdef"),
		TTL:    time.Hour,
	}
	memberToken, err := issuer.Issue(17, "member")
	if err != nil {
		t.Fatal(err)
	}
	adminToken, err := issuer.Issue(23, "admin")
	if err != nil {
		t.Fatal(err)
	}

	handlerCalled := false
	protected := RequireLogin(issuer, RequireRole("admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		who := WhoFrom(r.Context())
		if who.ID != 23 || who.Role != "admin" {
			t.Errorf("authenticated identity = %+v, want admin 23", who)
		}
		w.WriteHeader(http.StatusNoContent)
	})))

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized},
		{name: "member token", token: memberToken, wantStatus: http.StatusForbidden},
		{name: "admin token", token: adminToken, wantStatus: http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
			if test.token != "" {
				req.Header.Set("Authorization", "Bearer "+test.token)
			}
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)
			if rec.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, test.wantStatus)
			}
			if got := handlerCalled; got != (test.wantStatus == http.StatusNoContent) {
				t.Fatalf("handler called = %t, want %t", got, test.wantStatus == http.StatusNoContent)
			}
		})
	}
}
