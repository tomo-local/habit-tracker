package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// calendarService は OAuth 認証済みの Calendar API クライアントを返す。
// 初回はブラウザで同意フローを実行し、トークンを token.json に保存する。
func calendarService(ctx context.Context) (*calendar.Service, error) {
	b, err := readCredentials()
	if err != nil {
		return nil, err
	}
	// 読み取り(カレンダー一覧・イベント)+ add コマンド用のイベント書き込み
	conf, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("credentials.json の解析に失敗: %w", err)
	}

	tokPath := filepath.Join(configDir(), "token.json")
	tok, err := tokenFromFile(tokPath)
	if err != nil {
		tok, err = tokenFromWeb(ctx, conf)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokPath, tok); err != nil {
			return nil, err
		}
	}
	return calendar.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx, tok)))
}

// readCredentials はリポジトリ直下(実行ファイルの隣→カレントディレクトリ)、
// 次に ~/.config/habit-tracker/ の順で credentials.json を探す。
func readCredentials() ([]byte, error) {
	var paths []string
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), "credentials.json"))
	}
	paths = append(paths,
		"credentials.json",
		filepath.Join(configDir(), "credentials.json"),
	)
	for _, p := range paths {
		if b, err := os.ReadFile(p); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf(
		"credentials.json が見つかりません(探した場所: %v)。\n"+
			"GCPコンソールで OAuth クライアント(デスクトップアプリ)を作成し、JSONを配置してください", paths)
}

func tokenFromFile(path string) (*oauth2.Token, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(b, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func saveToken(path string, tok *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// tokenFromWeb はローカルにコールバック用サーバーを立ててブラウザで認可を受ける。
func tokenFromWeb(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer ln.Close()
	conf.RedirectURL = fmt.Sprintf("http://%s/", ln.Addr().String())

	codeCh := make(chan string, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "no code", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "認証が完了しました。このタブは閉じて構いません。")
		codeCh <- code
	})}
	go srv.Serve(ln)
	defer srv.Close()

	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("ブラウザで認証してください:")
	fmt.Println("  " + url)
	exec.Command("open", url).Start()

	select {
	case code := <-codeCh:
		return conf.Exchange(ctx, code)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
