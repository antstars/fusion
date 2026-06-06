package handler

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestOIDCLoginRateLimit(t *testing.T) {
	h := &Handler{
		oidcAuth:         &fakeOIDCAuth{},
		oidcStartLimiter: newRequestLimiter(2, 60, 60),
	}

	r := newTestRouter()
	r.GET("/api/oidc/login", h.oidcLogin)

	for i := 0; i < 2; i++ {
		w := performRequest(r, http.MethodGet, "/api/oidc/login", nil, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, w.Code)
		}
	}

	w := performRequest(r, http.MethodGet, "/api/oidc/login", nil, nil)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("limited request status = %d, want 429", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got == "" {
		t.Fatal("expected Retry-After header")
	}
}

type fakeOIDCAuth struct{}

func (f *fakeOIDCAuth) AuthURL() (string, error) {
	return "https://issuer.example.com/auth", nil
}

func (f *fakeOIDCAuth) Callback(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func TestRequestLimiterWindowReset(t *testing.T) {
	limiter := newRequestLimiter(1, 10, 10)
	now := time.Unix(100, 0)

	if allowed, _ := limiter.allow("ip", now); !allowed {
		t.Fatal("first request should be allowed")
	}
	if allowed, _ := limiter.allow("ip", now.Add(time.Second)); allowed {
		t.Fatal("second request in window should be blocked")
	}
	if allowed, _ := limiter.allow("ip", now.Add(20*time.Second)); !allowed {
		t.Fatal("request after block should be allowed")
	}
}
