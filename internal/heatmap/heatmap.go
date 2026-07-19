package heatmap

import (
	"fmt"
	"time"
)

var dayLabels = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

func Render(counts map[string]int, now time.Time) {
	start := now.AddDate(0, 0, -(52*7 - 1))
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}

	for wd := 0; wd < 7; wd++ {
		fmt.Printf("%s  ", dayLabels[wd])
		for week := 0; week < 52; week++ {
			d := start.AddDate(0, 0, week*7+wd)
			if d.After(now) {
				fmt.Print("  ")
			} else {
				fmt.Print(colorBlock(counts[d.Format("2006-01-02")]))
			}
		}
		fmt.Println()
	}
}

const (
	colorNone   = 0
	colorLight  = 1
	colorMedium = 2
	colorDark   = 3
)

// colorThresholds defines minute thresholds in ascending order (30m, 60m, 120m)
var colorThresholds = []int{30, 60, 120}

var colorCodes = map[int]int{
	colorNone:   236,
	colorLight:  22,
	colorMedium: 34,
	colorDark:   46,
}

func colorBlock(minutes int) string {
	level := colorNone
	for i, threshold := range colorThresholds {
		if minutes >= threshold {
			level = i + 1
		}
	}
	return fmt.Sprintf("\033[48;5;%dm  \033[0m", colorCodes[level])
}
