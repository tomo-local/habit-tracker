package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"

	"golang.org/x/oauth2"
)

type Auth interface {
	Authorize() (*oauth2.Token, error)
}

type auth struct {
	oauthConfig *oauth2.Config
}

func New(oauthConfig *oauth2.Config) Auth {
	return &auth{oauthConfig: oauthConfig}
}

func (a *auth) Authorize() (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	a.oauthConfig.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch: possible CSRF attack")
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			return
		}
		fmt.Fprintln(w, "Authorization complete. You can close this tab.")
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(listener)
	defer srv.Close()

	authURL := a.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println("Open this URL to authorize:", authURL)
	openBrowser(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return nil, err
	}

	token, err := a.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	return token, nil
}

var openBrowser = func(url string) {
	exec.Command("open", url).Start()
}
