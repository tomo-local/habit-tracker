package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"

	"habit-tracker/internal/auth"
	"habit-tracker/internal/config"
)

func (c *Cmd) RunAuth(args []string) error {
	if len(args) == 0 || args[0] != "login" {
		return fmt.Errorf("usage: habit-tracker auth login")
	}

	oauthCfg := &oauth2.Config{
		ClientID:     c.cfg.ClientID(),
		ClientSecret: c.cfg.ClientSecret(),
		Endpoint:     google.Endpoint,
		Scopes:       []string{gcal.CalendarScope},
	}

	a := auth.New(oauthCfg)
	token, err := a.Authorize()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(config.TokenPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
