package cmd

import (
	"context"

	"habit-tracker/internal/calendar"
	"habit-tracker/internal/config"
	"habit-tracker/internal/prompt"
)

type Cmd struct {
	ctx    context.Context
	cfg    *config.Config
	cal    calendar.Calendar
	prompt prompt.Prompter
}

func New(ctx context.Context, cfg *config.Config, cal calendar.Calendar, p prompt.Prompter) *Cmd {
	return &Cmd{ctx: ctx, cfg: cfg, cal: cal, prompt: p}
}
