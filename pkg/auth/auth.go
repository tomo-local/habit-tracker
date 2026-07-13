package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"habit-tracker/config"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Authenticate は credentials JSON から OAuth トークンソースを返す。
// トークンが未保存の場合はブラウザで同意フローを実行し、token.json に保存する。
func Authenticate(ctx context.Context, credentials []byte) (oauth2.TokenSource, error) {
	conf, err := google.ConfigFromJSON(credentials, calendar.CalendarReadonlyScope, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("credentials の解析に失敗: %w", err)
	}

	tokPath := filepath.Join(config.ConfigDir(), "token.json")
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
	return conf.TokenSource(ctx, tok), nil
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
