package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"habit-tracker/config"
	"habit-tracker/internal/web"
	"habit-tracker/pkg/auth"
	"habit-tracker/pkg/calendar"
)

func main() {
	port := flag.Int("port", 8391, "port for the web view")
	weeks := flag.Int("weeks", 26, "number of weeks to show in the grid")
	date := flag.String("date", time.Now().Format("2006-01-02 15:04"), "end time for `add` (YYYY-MM-DD HH:MM); start is 30 min before")
	flag.Parse()

	ctx := context.Background()

	switch flag.Arg(0) {
	case "list":
		svc, err := newCalendarClient(ctx)
		if err != nil {
			fatal(err)
		}
		cals, err := svc.ListCalendars()
		if err != nil {
			fatal(err)
		}
		fmt.Println("利用可能なカレンダー:")
		for _, c := range cals {
			fmt.Printf("  %s\n", c)
		}
		fmt.Printf("\n習慣として使うカレンダー名を %s に列挙してください:\n", filepath.Join(config.ConfigDir(), "config.json"))
		fmt.Println(`  {"calendars": ["筋トレ", "読書"]}`)
	case "serve", "":
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config.json が読めません: %v\n", err)
			fmt.Fprintln(os.Stderr, "まず `habit-tracker list` でカレンダー名を確認し、config.json を作成してください。")
			os.Exit(1)
		}
		if len(cfg.Calendars) == 0 {
			fatal(fmt.Errorf("config.json の calendars が空です"))
		}
		svc, err := newCalendarClient(ctx)
		if err != nil {
			fatal(err)
		}
		if err := web.Serve(svc, cfg, *port, *weeks); err != nil {
			fatal(err)
		}
	case "add":
		cfg, err := config.LoadConfig()
		if err != nil {
			fatal(err)
		}
		if len(cfg.Calendars) == 0 {
			fatal(fmt.Errorf("config.json の calendars が空です"))
		}
		names := flag.Args()[1:]
		if len(names) == 0 {
			names, err = selectHabits(cfg.Habits)
			if err != nil {
				fatal(err)
			}
		}
		svc, err := newCalendarClient(ctx)
		if err != nil {
			fatal(err)
		}
		for _, name := range names {
			if err := calendar.AddEntry(svc, cfg, name, *date); err != nil {
				fatal(err)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: habit-tracker [-port N] [-weeks N] [-date YYYY-MM-DD] [list|serve|add <習慣名>]\n", flag.Arg(0))
		os.Exit(1)
	}
}

// newCalendarClient は credentials を読み込み、認証して calendar.Client を返す。
func newCalendarClient(ctx context.Context) (calendar.Client, error) {
	creds, err := readCredentials()
	if err != nil {
		return nil, err
	}
	ts, err := auth.Authenticate(ctx, creds)
	if err != nil {
		return nil, err
	}
	return calendar.NewClient(ctx, ts)
}

// readCredentials は credentials.json を探して読み込む。
// 実行ファイルの隣 → カレントディレクトリ → ~/.config/habit-tracker/ の順で検索する。
func readCredentials() ([]byte, error) {
	var paths []string
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), "credentials.json"))
	}
	paths = append(paths,
		"credentials.json",
		filepath.Join(config.ConfigDir(), "credentials.json"),
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

// selectHabits は登録済み習慣を番号選択で選ばせる(スペース/カンマ区切りで複数可)。
func selectHabits(habits []string) ([]string, error) {
	if len(habits) == 0 {
		return nil, fmt.Errorf("config.json の habits が空です。選択式で使うには習慣名を登録してください")
	}
	fmt.Println("どの習慣を記録しますか?(番号、複数はスペース区切り)")
	for i, h := range habits {
		fmt.Printf("  %d) %s\n", i+1, h)
	}
	fmt.Print("> ")

	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return nil, fmt.Errorf("入力がありません")
	}
	var names []string
	for _, tok := range strings.FieldsFunc(sc.Text(), func(r rune) bool { return r == ' ' || r == ',' || r == '　' }) {
		n, err := strconv.Atoi(tok)
		if err != nil || n < 1 || n > len(habits) {
			return nil, fmt.Errorf("不正な選択: %q (1〜%d の番号を入力してください)", tok, len(habits))
		}
		names = append(names, habits[n-1])
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("何も選択されませんでした")
	}
	return names, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
