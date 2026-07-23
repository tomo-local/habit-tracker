package calendar

import (
	"context"

	"golang.org/x/oauth2"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar interface {
	GetClient(ts oauth2.TokenSource) (Client, error)
}

type service struct {
	ctx context.Context
}

func New(ctx context.Context) Calendar {
	return &service{ctx: ctx}
}

func (s *service) GetClient(ts oauth2.TokenSource) (Client, error) {
	svc, err := gcal.NewService(s.ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return &client{svc: svc}, nil
}
