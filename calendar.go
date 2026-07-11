package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
	"google.golang.org/api/calendar/v3"
)

// listCalendars はアカウントのカレンダー名一覧を返す。
func listCalendars(svc *calendar.Service) ([]string, error) {
	var names []string
	err := svc.CalendarList.List().Pages(nil, func(page *calendar.CalendarList) error {
		for _, item := range page.Items {
			names = append(names, item.Summary)
		}
		return nil
	})
	return names, err
}

// occurrence はカレンダー上の1イベント(タイトルと日付 "2006-01-02")。
type occurrence struct {
	Title string
	Date  string
}

// calendarOccurrences は指定カレンダーの since 以降のイベントを返す。
func calendarOccurrences(svc *calendar.Service, calName string, since time.Time) ([]occurrence, error) {
	calID, err := findCalendarID(svc, calName)
	if err != nil {
		return nil, err
	}

	var occs []occurrence
	err = svc.Events.List(calID).
		SingleEvents(true).
		TimeMin(since.Format(time.RFC3339)).
		MaxResults(2500).
		Pages(nil, func(page *calendar.Events) error {
			for _, ev := range page.Items {
				if ev.Status == "cancelled" || ev.Start == nil {
					continue
				}
				title := strings.TrimSpace(ev.Summary)
				if title == "" {
					title = "(無題)"
				}
				if ev.Start.Date != "" { // 終日イベント
					occs = append(occs, occurrence{title, ev.Start.Date})
				} else if ev.Start.DateTime != "" {
					t, err := time.Parse(time.RFC3339, ev.Start.DateTime)
					if err == nil {
						occs = append(occs, occurrence{title, t.Local().Format("2006-01-02")})
					}
				}
			}
			return nil
		})
	return occs, err
}

func findCalendarID(svc *calendar.Service, name string) (string, error) {
	var id string
	err := svc.CalendarList.List().Pages(nil, func(page *calendar.CalendarList) error {
		for _, item := range page.Items {
			if item.Summary == name || item.SummaryOverride == name {
				id = item.Id
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("カレンダー %q が見つかりません (`habit-tracker list` で確認できます)", name)
	}
	return id, nil
}

// normalizeTitle は照合用にタイトルを正規化する(前後空白・全角半角・大文字小文字)。
func normalizeTitle(s string) string {
	return strings.ToLower(strings.TrimSpace(norm.NFKC.String(s)))
}

// matchHabit はタイトルを正式な習慣名に解決する。
// 正規化後に習慣名を含んでいれば一致とみなし、複数一致時は最長の習慣名を優先。
// 一致しなければ元のタイトルをそのまま返す。
func matchHabit(title string, habits []string) string {
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

// addEntry は習慣の実施記録(終日イベント)を追加する。
// すでに同じタイトルのイベントがその日にあれば何もしない。
func addEntry(svc *calendar.Service, cfg *Config, habitName, date string) error {
	day, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return fmt.Errorf("日付は YYYY-MM-DD 形式で指定してください: %w", err)
	}

	// 習慣名が登録されている場合は正式名に解決し、未登録の名前は弾く(誤字防止)
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

	// group_by_title なら記録先は先頭カレンダー、そうでなければ習慣名のカレンダー
	calName := habitName
	if cfg.GroupByTitle {
		calName = cfg.Calendars[0]
	}
	calID, err := findCalendarID(svc, calName)
	if err != nil {
		return err
	}

	occs, err := calendarOccurrences(svc, calName, day)
	if err != nil {
		return err
	}
	for _, o := range occs {
		if o.Date == date && matchHabit(o.Title, cfg.Habits) == habitName {
			fmt.Printf("%s の %q は記録済みです\n", date, habitName)
			return nil
		}
	}

	_, err = svc.Events.Insert(calID, &calendar.Event{
		Summary: habitName,
		Start:   &calendar.EventDateTime{Date: date},
		End:     &calendar.EventDateTime{Date: day.AddDate(0, 0, 1).Format("2006-01-02")},
	}).Do()
	if err != nil {
		return fmt.Errorf("イベント作成に失敗: %w", err)
	}
	fmt.Printf("記録しました: %s %s (カレンダー: %s)\n", date, habitName, calName)
	return nil
}

// streak は today から遡って連続している日数を返す。
// 今日まだ未実施でも昨日まで続いていればストリークは継続扱い。
func streak(days map[string]bool, today time.Time) int {
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
