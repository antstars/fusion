package auth

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func newTestOIDCAuthenticator() *OIDCAuthenticator {
	return &OIDCAuthenticator{
		oauth2Config: oauth2.Config{
			ClientID:    "client",
			RedirectURL: "https://app.example.com/api/oidc/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://issuer.example.com/auth",
			},
		},
		states: make(map[string]stateEntry),
	}
}

func TestOIDCAuthURLRejectsWhenStateStoreIsFull(t *testing.T) {
	authenticator := newTestOIDCAuthenticator()

	for i := 0; i < maxOIDCStates; i++ {
		if _, err := authenticator.AuthURL(); err != nil {
			t.Fatalf("AuthURL() before cap: %v", err)
		}
	}

	if _, err := authenticator.AuthURL(); !errors.Is(err, ErrOIDCStateStoreFull) {
		t.Fatalf("AuthURL() error = %v, want ErrOIDCStateStoreFull", err)
	}
}

func TestOIDCAuthURLCleansExpiredStatesBeforeCapacityCheck(t *testing.T) {
	authenticator := newTestOIDCAuthenticator()
	for i := 0; i < maxOIDCStates; i++ {
		authenticator.states[string(rune(i))] = stateEntry{
			codeVerifier: "verifier",
			createdAt:    time.Now().Add(-stateMaxAge - time.Minute),
		}
	}

	if _, err := authenticator.AuthURL(); err != nil {
		t.Fatalf("AuthURL() after expired states: %v", err)
	}
	if got := len(authenticator.states); got != 1 {
		t.Fatalf("state count = %d, want 1", got)
	}
}
