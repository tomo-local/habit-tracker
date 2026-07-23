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

func New(ctx context.Context) *Cmd {
	return &Cmd{ctx: ctx}
}

func (c *Cmd) setup() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	cal, err := c.newCalendarClient()
	if err != nil {
		return err
	}
	c.cfg = cfg
	c.cal = cal
	return nil
}

func (c *Cmd) newCalendarClient() (calendar.Calendar, error) {
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
	return calendar.New(c.ctx, oauthCfg.TokenSource(c.ctx, &token))
}
