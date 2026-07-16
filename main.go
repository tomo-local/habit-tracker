package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Config struct {
	CredentialsPath string
	TokenPath       string
	CalendarID      string
}

func loadConfig() Config {
	home := os.Getenv("HOME")
	return Config{
		CredentialsPath: getenv("HABIT_CREDENTIALS", "./credentials.json"),
		TokenPath:       getenv("HABIT_TOKEN", home+"/.config/habit-tracker/token.json"),
		CalendarID:      os.Getenv("HABIT_CALENDAR_ID"),
	}
}

func getenv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	if len(os.Args) < 2 {
		fatal(fmt.Errorf("command required. usage: app <command> [flags]"))
	}

	cfg := loadConfig()
	if cfg.CalendarID == "" {
		fatal(fmt.Errorf("HABIT_CALENDAR_ID is required"))
	}

	service, err := NewCalendarClient(context.Background(), cfg.CredentialsPath, cfg.TokenPath)
	if err != nil {
		fatal(err)
	}

	switch os.Args[1] {
	case "add":
		cmd, err := NewAddCmd(os.Args[2:])
		if err != nil {
			fatal(err)
		}
		if *cmd.help {
			cmd.fs.Usage()
			return
		}

		end := time.Now()
		start := end.Add(-time.Minute * 30)

		event := &calendar.Event{
			Summary: cmd.name,
			Start: &calendar.EventDateTime{
				DateTime: start.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: end.Format(time.RFC3339),
			},
		}
		_, err = service.Events.Insert(cfg.CalendarID, event).Do()
		if err != nil {
			fatal(fmt.Errorf("insert event: %w", err))
		}
		fmt.Println("成功")
	default:
		fatal(fmt.Errorf("unknown command: %q", os.Args[1]))
	}
}

type AddCmd struct {
	fs   *flag.FlagSet
	help *bool
	name string
}

func NewAddCmd(args []string) (*AddCmd, error) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	help := fs.Bool("h", false, "show help")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if len(fs.Args()) == 0 {
		fs.Usage()
		return nil, fmt.Errorf("title is required: habit add <title>")
	}

	return &AddCmd{
		fs:   fs,
		help: help,
		name: fs.Args()[0],
	}, nil
}

func NewCalendarClient(ctx context.Context, credentialsPath, tokenPath string) (*calendar.Service, error) {
	credBytes, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("read token: %w", err)
	}

	// credentials.json からアプリの client_id/secret を読み込んで OAuth2 config を作る
	// CalendarScope はカレンダーの読み書き権限を要求するスコープ
	config, err := google.ConfigFromJSON(credBytes, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	// token.json に保存された access_token と refresh_token を復元する
	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	// TokenSource は access_token が期限切れのとき refresh_token で自動更新する
	ts := config.TokenSource(ctx, &token)

	svc, err := calendar.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}

	return svc, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
