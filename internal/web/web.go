package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"sort"
	"time"

	gcal "google.golang.org/api/calendar/v3"
	"habit-tracker/config"
	"habit-tracker/pkg/calendar"
)

type cell struct {
	Date   string
	Done   bool
	Future bool
	Today  bool
}

type habitView struct {
	Name   string
	Streak int
	Total  int
	Weeks  [][7]cell // 週ごと、日曜始まり
}

type pageData struct {
	Habits    []habitView
	UpdatedAt string
}

func Serve(svc *gcal.Service, cfg *config.Config, port, weeks int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := buildPage(svc, cfg, weeks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := page.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Println("習慣トラッカー起動:", url)
	exec.Command("open", url).Start()
	return http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
}

func buildPage(svc *gcal.Service, cfg *config.Config, weeks int) (*pageData, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	// グリッドの起点: weeks 週前の日曜日
	start := today.AddDate(0, 0, -int(today.Weekday())-(weeks-1)*7)

	data := &pageData{UpdatedAt: now.Format("2006-01-02 15:04")}
	for _, name := range cfg.Calendars {
		occs, err := calendar.CalendarOccurrences(svc, name, start)
		if err != nil {
			return nil, err
		}
		for _, h := range groupHabits(name, occs, cfg) {
			data.Habits = append(data.Habits, buildHabitView(h.name, h.days, start, today, weeks))
		}
	}
	return data, nil
}

type habitDays struct {
	name string
	days map[string]bool
}

// groupHabits はイベント列を習慣単位にまとめる。
// group_by_title が true ならタイトルごと(habits 登録があれば正式名に集約)、
// false ならカレンダー全体で1習慣。
// habits に登録された習慣はイベント0件でも行を出す。
// どの習慣にも一致しないタイトルは独立した行になる(誤字に気づけるように)。
func groupHabits(calName string, occs []calendar.Occurrence, cfg *config.Config) []habitDays {
	grouped := map[string]map[string]bool{}
	var unmatched []string

	if cfg.GroupByTitle {
		for _, h := range cfg.Habits {
			grouped[h] = map[string]bool{}
		}
	}
	for _, o := range occs {
		key := calName
		if cfg.GroupByTitle {
			key = calendar.MatchHabit(o.Title, cfg.Habits)
		}
		if grouped[key] == nil {
			grouped[key] = map[string]bool{}
			unmatched = append(unmatched, key)
		}
		grouped[key][o.Date] = true
	}

	// 表示順: config の habits 順 → 未登録タイトル(名前順)
	sort.Strings(unmatched)
	var order []string
	if cfg.GroupByTitle {
		order = append(order, cfg.Habits...)
	} else if grouped[calName] != nil {
		order = append(order, calName)
	}
	order = append(order, unmatched...)

	var habits []habitDays
	for _, key := range order {
		habits = append(habits, habitDays{key, grouped[key]})
	}
	return habits
}

func buildHabitView(name string, days map[string]bool, start, today time.Time, weeks int) habitView {
	hv := habitView{Name: name, Streak: calendar.Streak(days, today)}
	for w := 0; w < weeks; w++ {
		var week [7]cell
		for d := 0; d < 7; d++ {
			day := start.AddDate(0, 0, w*7+d)
			key := day.Format("2006-01-02")
			c := cell{
				Date:   key,
				Done:   days[key],
				Future: day.After(today),
				Today:  day.Equal(today),
			}
			if c.Done {
				hv.Total++
			}
			week[d] = c
		}
		hv.Weeks = append(hv.Weeks, week)
	}
	return hv
}

var page = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>習慣トラッカー</title>
<style>
  :root { color-scheme: dark; }
  body {
    background: #0d1117; color: #e6edf3;
    font-family: -apple-system, "Hiragino Sans", sans-serif;
    max-width: 900px; margin: 40px auto; padding: 0 20px;
  }
  h1 { font-size: 20px; }
  .updated { color: #8b949e; font-size: 12px; }
  .habit { margin: 28px 0; }
  .habit-header { display: flex; align-items: baseline; gap: 14px; margin-bottom: 8px; }
  .habit-name { font-size: 16px; font-weight: 600; }
  .streak { color: #f0883e; font-size: 14px; }
  .total { color: #8b949e; font-size: 13px; }
  .grid { display: flex; gap: 3px; overflow-x: auto; padding-bottom: 4px; }
  .labels { display: grid; grid-template-rows: repeat(7, 12px); gap: 3px;
            font-size: 9px; color: #8b949e; margin-right: 4px; }
  .labels span { line-height: 12px; }
  .week { display: grid; grid-template-rows: repeat(7, 12px); gap: 3px; }
  .cell {
    width: 12px; height: 12px; border-radius: 3px;
    background: #161b22; outline: 1px solid rgba(255,255,255,0.05); outline-offset: -1px;
  }
  .cell.done { background: #39d353; }
  .cell.future { background: transparent; outline: none; }
  .cell.today { outline: 1px solid #e6edf3; }
</style>
</head>
<body>
<h1>習慣トラッカー</h1>
<div class="updated">更新: {{.UpdatedAt}} / リロードで再取得</div>
{{range .Habits}}
<div class="habit">
  <div class="habit-header">
    <span class="habit-name">{{.Name}}</span>
    <span class="streak">🔥 {{.Streak}}日連続</span>
    <span class="total">計 {{.Total}}回</span>
  </div>
  <div class="grid">
    <div class="labels"><span>日</span><span>月</span><span>火</span><span>水</span><span>木</span><span>金</span><span>土</span></div>
    {{range .Weeks}}
    <div class="week">
      {{range .}}<div class="cell{{if .Done}} done{{end}}{{if .Future}} future{{end}}{{if .Today}} today{{end}}" title="{{.Date}}"></div>{{end}}
    </div>
    {{end}}
  </div>
</div>
{{end}}
</body>
</html>`))
