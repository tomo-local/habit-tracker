package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func Test_Authorize(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test_access_token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer ts.Close()

	tests := []struct {
		name       string
		code       string // empty = no code in callback
		wantToken  string
		wantErrMsg string
	}{
		{name: "success", code: "test_code", wantToken: "test_access_token"},
		{name: "no code in callback", code: "", wantErrMsg: "no code in callback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openBrowser = func(authURL string) {
				u, _ := url.Parse(authURL)
				redirectURI := u.Query().Get("redirect_uri")
				state := u.Query().Get("state")
				go func() {
					time.Sleep(20 * time.Millisecond)
					if tt.code != "" {
						http.Get(redirectURI + "?code=" + tt.code + "&state=" + state)
					} else {
						http.Get(redirectURI + "?state=" + state)
					}
				}()
			}

			a := &auth{oauthConfig: &oauth2.Config{
				ClientID:     "client_id",
				ClientSecret: "client_secret",
				Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
			}}

			token, err := a.Authorize()
			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token.AccessToken != tt.wantToken {
				t.Errorf("got %q, want %q", token.AccessToken, tt.wantToken)
			}
		})
	}
}
