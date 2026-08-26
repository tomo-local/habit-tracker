package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"habit-tracker/internal/auth"
	"habit-tracker/internal/calendar"
	"habit-tracker/internal/config"
)

func (c *Cmd) RunAuth(args []string) error {
	if len(args) == 0 || args[0] != "login" {
		return fmt.Errorf("usage: habit auth login")
	}

	oauthCfg := &oauth2.Config{
		ClientID:     c.cfg.ClientID(),
		ClientSecret: c.cfg.ClientSecret(),
		Endpoint:     google.Endpoint,
		Scopes:       calendar.Scopes,
	}
	_, err := authorizeAndSaveToken(oauthCfg)
	return err
}

// authorizeAndSaveToken runs the interactive OAuth flow and persists the
// resulting token so other commands can reuse it via c.cfg.LoadToken().
func authorizeAndSaveToken(oauthCfg *oauth2.Config) (*oauth2.Token, error) {
	a := auth.New(oauthCfg)
	token, err := a.Authorize()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(config.ConfigDir(), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(config.TokenPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return nil, err
	}
	return token, nil
}

// loadOrAuthorizeToken returns the persisted token, running the interactive
// OAuth flow to obtain and save one if it doesn't exist yet.
func loadOrAuthorizeToken(cfg *config.Config, oauthCfg *oauth2.Config) (*oauth2.Token, error) {
	token, err := cfg.LoadToken()
	if err == nil {
		return token, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read token: %w", err)
	}
	return authorizeAndSaveToken(oauthCfg)
}
