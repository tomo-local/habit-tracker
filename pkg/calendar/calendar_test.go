package calendar_test

import (
	"habit-tracker/pkg/calendar"
	"testing"
	"time"
)

func TestMatchHabit(t *testing.T) {
	habits := []string{"筋トレ", "読書", "情報処理の資格勉強"}
	tests := []struct {
		title, want string
	}{
		{"筋トレ", "筋トレ"},
		{" 筋トレ ", "筋トレ"},  // 前後空白
		{"筋トレ30分", "筋トレ"}, // 部分一致
		{"情報処理の資格勉強", "情報処理の資格勉強"},
		{"きんトレ", "きんトレ"}, // 誤字は集約されず独立行(気づける)
		{"散歩", "散歩"},     // 未登録タイトルはそのまま
	}
	for _, tt := range tests {
		if got := calendar.MatchHabit(tt.title, habits); got != tt.want {
			t.Errorf("MatchHabit(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestStreak(t *testing.T) {
	today := time.Date(2026, 7, 11, 0, 0, 0, 0, time.Local)
	day := func(offset int) string { return today.AddDate(0, 0, offset).Format("2006-01-02") }

	tests := []struct {
		name string
		days []int // today からのオフセット
		want int
	}{
		{"実施なし", nil, 0},
		{"今日のみ", []int{0}, 1},
		{"3日連続(今日含む)", []int{0, -1, -2}, 3},
		{"今日未実施でも昨日まで継続", []int{-1, -2}, 2},
		{"一昨日で途切れ", []int{-2, -3}, 0},
		{"途中に穴", []int{0, -1, -3, -4}, 2},
	}
	for _, tt := range tests {
		days := map[string]bool{}
		for _, o := range tt.days {
			days[day(o)] = true
		}
		if got := calendar.Streak(days, today); got != tt.want {
			t.Errorf("%s: streak = %d, want %d", tt.name, got, tt.want)
		}
	}
}
