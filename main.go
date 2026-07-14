package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	// ctx := context.Background()
	flag.Parse()

	c := Config{}
	cmd := flag.Arg(0)

	switch cmd {
	case "add":
		token, err := c.GetGoogleCalendarToken()
		if err != nil {
			fatal(err)
		}

		title := flag.Arg(1)
		if title == "" {
			fatal(fmt.Errorf("required: title"))
		}

		client, err := NewCalendarClient(context.Background(), token.OAuth2, *token.Credentials)
		if err != nil {
			fatal(err)
		}

		if err := client.AddEvent(title); err != nil {
			fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: habit-tracker [-port N] [-weeks N] [-date YYYY-MM-DD] [list|serve|add <習慣名>]\n", flag.Arg(0))
		os.Exit(1)

	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

type Config struct {
	googleCalendarTokenPath string
	googleCredentialsPath   string
}

type GoogleCredentials struct {
	Installed *GoogleCredentialsConfig `json:"installed"`
	Web       *GoogleCredentialsConfig `json:"web"`
}

type GoogleCredentialsConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

type GoogleCalendarToken struct {
	OAuth2      *oauth2.Token
	Credentials *GoogleCredentials
}

type CalendarClient struct {
	ctx context.Context
	svc *calendar.Service
}

func NewCalendarClient(ctx context.Context, token *oauth2.Token, credentials GoogleCredentials) (*CalendarClient, error) {
	credsConfig := credentials.Installed
	if credsConfig == nil {
		credsConfig = credentials.Web
	}
	if credsConfig == nil {
		return nil, fmt.Errorf("no credentials found")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     credsConfig.ClientID,
		ClientSecret: credsConfig.ClientSecret,
		RedirectURL:  credsConfig.RedirectURIs[0],
		Endpoint: oauth2.Endpoint{
			AuthURL:  credsConfig.AuthURI,
			TokenURL: credsConfig.TokenURI,
		},
		Scopes: []string{calendar.CalendarScope},
	}

	httpClient := oauthConfig.Client(ctx, token)
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return &CalendarClient{ctx: ctx, svc: svc}, nil
}

func (c *CalendarClient) AddEvent(title string) error {
	// TODO(human): カレンダーにイベントを追加する
	return nil
}

func (c *Config) getGoogleCredentials() (*GoogleCredentials, error) {
	path, err := c.resolveConfigPath(c.googleCredentialsPath)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(filepath.Join(path, "credentials.json"))
	if err != nil {
		return nil, err
	}

	var credential GoogleCredentials
	if err := json.Unmarshal(b, &credential); err != nil {
		return nil, err
	}

	return &credential, nil
}

func (c *Config) GetGoogleCalendarToken() (*GoogleCalendarToken, error) {
	path, err := c.resolveConfigPath(c.googleCalendarTokenPath)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(filepath.Join(path, "token.json"))
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(b, &token); err != nil {
		return nil, err
	}

	credentials, err := c.getGoogleCredentials()
	if err != nil {
		return nil, err
	}

	return &GoogleCalendarToken{
		OAuth2:      &token,
		Credentials: credentials,
	}, nil
}

func (c *Config) resolveConfigPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "habit-tracker"), nil
}
