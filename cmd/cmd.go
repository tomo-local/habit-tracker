package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"

	"habit-tracker/internal/calendar"
	"habit-tracker/internal/config"
)

type Cmd struct {
	ctx context.Context
	cfg *config.Config
	cal calendar.Calendar
}

func New(ctx context.Context, cfg *config.Config, cal calendar.Calendar) *Cmd {
	return &Cmd{ctx: ctx, cfg: cfg, cal: cal}
}

func (c *Cmd) getClient() (calendar.Client, error) {
	credPath, err := config.CredentialsPath()
	if err != nil {
		return nil, err
	}
	credBytes, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}
	oauthCfg, err := google.ConfigFromJSON(credBytes, gcal.CalendarScope)
	if err != nil {
		return nil, err
	}
	tokenBytes, err := os.ReadFile(config.TokenPath())
	if err != nil {
		return nil, fmt.Errorf("read token (run auth login first): %w", err)
	}
	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	return c.cal.GetClient(oauthCfg.TokenSource(c.ctx, &token))
}
