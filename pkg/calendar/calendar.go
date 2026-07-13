package calendar

import (
	"fmt"
	"habit-tracker/config"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
	gcal "google.golang.org/api/calendar/v3"
)

// Occurrence はカレンダー上の1イベント(タイトルと日付 "2006-01-02")。
type Occurrence struct {
	Title   string
	Date    string
	Minutes int
}

// ToOccurrences は Google Calendar のイベント一覧を Occurrence に変換する。
// キャンセル済みや開始日時のないイベントは除外する。
func ToOccurrences(items []*gcal.Event) []Occurrence {
	var occs []Occurrence
	for _, ev := range items {
		if ev.Status == "cancelled" || ev.Start == nil {
			continue
		}
		title := strings.TrimSpace(ev.Summary)
		if title == "" {
			title = "(無題)"
		}
		if ev.Start.Date != "" {
			occs = append(occs, Occurrence{title, ev.Start.Date, 60})
		} else if ev.Start.DateTime != "" {
			t, err := time.Parse(time.RFC3339, ev.Start.DateTime)
			if err == nil {
				minutes := 0
				if ev.End != nil && ev.End.DateTime != "" {
					if endT, err2 := time.Parse(time.RFC3339, ev.End.DateTime); err2 == nil {
						minutes = int(endT.Sub(t).Minutes())
					}
				}
				occs = append(occs, Occurrence{title, t.Local().Format("2006-01-02"), minutes})
			}
		}
	}
	return occs
}

// normalizeTitle は照合用にタイトルを正規化する(前後空白・全角半角・大文字小文字)。
func normalizeTitle(s string) string {
	return strings.ToLower(strings.TrimSpace(norm.NFKC.String(s)))
}

// MatchHabit はタイトルを正式な習慣名に解決する。
// 正規化後に習慣名を含んでいれば一致とみなし、複数一致時は最長の習慣名を優先。
// 一致しなければ元のタイトルをそのまま返す。
func MatchHabit(title string, habits []string) string {
	nt := normalizeTitle(title)
	best := ""
	for _, h := range habits {
		if strings.Contains(nt, normalizeTitle(h)) && len(h) > len(best) {
			best = h
		}
	}
	if best == "" {
		return title
	}
	return best
}

// AddEntry は習慣の実施記録(時間付きイベント)を追加する。
// 習慣名が登録済みリストにある場合は正式名に解決し、未登録の名前は弾く(誤字防止)。
func AddEntry(c Client, cfg *config.Config, habitName, date string) error {
	end, err := time.ParseInLocation("2006-01-02 15:04", date, time.Local)
	if err != nil {
		return fmt.Errorf("日付は YYYY-MM-DD HH:MM 形式で指定してください: %w", err)
	}
	start := end.Add(-30 * time.Minute)
	dateOnly := end.Format("2006-01-02")

	if len(cfg.Habits) > 0 {
		resolved := ""
		for _, h := range cfg.Habits {
			if normalizeTitle(h) == normalizeTitle(habitName) {
				resolved = h
				break
			}
		}
		if resolved == "" {
			return fmt.Errorf("未登録の習慣名 %q です。config.json の habits に登録済み: %s",
				habitName, strings.Join(cfg.Habits, ", "))
		}
		habitName = resolved
	}

	calName := habitName
	if cfg.GroupByTitle {
		calName = cfg.Calendars[0]
	}

	if err := c.AddEvent(calName, habitName, start, end); err != nil {
		return err
	}
	fmt.Printf("記録しました: %s %s (カレンダー: %s)\n", dateOnly, habitName, calName)
	return nil
}

// Streak は today から遡って連続している日数を返す。
// 今日まだ未実施でも昨日まで続いていればストリークは継続扱い。
func Streak(days map[string]bool, today time.Time) int {
	d := today
	if !days[d.Format("2006-01-02")] {
		d = d.AddDate(0, 0, -1)
	}
	n := 0
	for days[d.Format("2006-01-02")] {
		n++
		d = d.AddDate(0, 0, -1)
	}
	return n
}
