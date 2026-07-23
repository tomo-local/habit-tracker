package cmd

import (
	"context"

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
