package calendar

import (
	"context"

	"golang.org/x/oauth2"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar interface {
	GetClient(ctx context.Context, oauthCfg *oauth2.Config, token *oauth2.Token) (Client, error)
}

type service struct{}

func New() Calendar {
	return &service{}
}

func (s *service) GetClient(ctx context.Context, oauthCfg *oauth2.Config, token *oauth2.Token) (Client, error) {
	svc, err := gcal.NewService(ctx, option.WithTokenSource(oauthCfg.TokenSource(ctx, token)))
	if err != nil {
		return nil, err
	}
	return &client{svc: svc}, nil
}
