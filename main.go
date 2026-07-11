package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config は ~/.config/habit-tracker/config.json の内容。
// Calendars には習慣として扱うカレンダー名(summary)を列挙する。
// GroupByTitle が true の場合、カレンダー内のイベントタイトルごとに別習慣として扱う。
// Habits に正式な習慣名を登録すると、タイトルは正規化+部分一致でこの名前に集約され、
// add コマンドもこの中の名前しか受け付けなくなる(誤字対策)。
type Config struct {
	Calendars    []string `json:"calendars"`
	GroupByTitle bool     `json:"group_by_title"`
	Habits       []string `json:"habits"`
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot resolve home dir:", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "habit-tracker")
}

func loadConfig() (*Config, error) {
	b, err := os.ReadFile(filepath.Join(configDir(), "config.json"))
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func main() {
	port := flag.Int("port", 8391, "port for the web view")
	weeks := flag.Int("weeks", 26, "number of weeks to show in the grid")
	date := flag.String("date", time.Now().Format("2006-01-02"), "date for `add` (YYYY-MM-DD)")
	flag.Parse()

	ctx := context.Background()

	switch flag.Arg(0) {
	case "list":
		svc, err := calendarService(ctx)
		if err != nil {
			fatal(err)
		}
		cals, err := listCalendars(svc)
		if err != nil {
			fatal(err)
		}
		fmt.Println("利用可能なカレンダー:")
		for _, c := range cals {
			fmt.Printf("  %s\n", c)
		}
		fmt.Printf("\n習慣として使うカレンダー名を %s に列挙してください:\n", filepath.Join(configDir(), "config.json"))
		fmt.Println(`  {"calendars": ["筋トレ", "読書"]}`)
	case "serve", "":
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config.json が読めません: %v\n", err)
			fmt.Fprintln(os.Stderr, "まず `habit-tracker list` でカレンダー名を確認し、config.json を作成してください。")
			os.Exit(1)
		}
		if len(cfg.Calendars) == 0 {
			fatal(fmt.Errorf("config.json の calendars が空です"))
		}
		svc, err := calendarService(ctx)
		if err != nil {
			fatal(err)
		}
		if err := serve(svc, cfg, *port, *weeks); err != nil {
			fatal(err)
		}
	case "add":
		cfg, err := loadConfig()
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
		svc, err := calendarService(ctx)
		if err != nil {
			fatal(err)
		}
		for _, name := range names {
			if err := addEntry(svc, cfg, name, *date); err != nil {
				fatal(err)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: habit-tracker [-port N] [-weeks N] [-date YYYY-MM-DD] [list|serve|add <習慣名>]\n", flag.Arg(0))
		os.Exit(1)
	}
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
