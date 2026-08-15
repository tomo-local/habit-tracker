package heatmap

import (
	"fmt"
	"time"
)

var dayLabels = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

func renderMonthRow(start, now time.Time, weeks int) {
	buf := make([]byte, weeks*2)
	for i := range buf {
		buf[i] = ' '
	}
	lastMonth := time.Month(0)
	for week := 0; week < weeks; week++ {
		d := start.AddDate(0, 0, week*7)
		if d.After(now) {
			break
		}
		if d.Month() != lastMonth {
			label := []byte(d.Format("Jan"))
			pos := week * 2
			for i, c := range label {
				if pos+i < len(buf) {
					buf[pos+i] = c
				}
			}
			lastMonth = d.Month()
		}
	}
	fmt.Printf("     %s\n", buf)
}

func Render(counts map[string]int, now time.Time, weeks int) {
	start := now.AddDate(0, 0, -(weeks-1)*7)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}

	renderMonthRow(start, now, weeks)

	for wd := 0; wd < 7; wd++ {
		fmt.Printf("%s  ", dayLabels[wd])
		for week := 0; week < weeks; week++ {
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
	colorMax    = 4
)

// colorThresholds defines minute thresholds in ascending order (30m, 60m, 120m)
var colorThresholds = []int{30, 60, 120}

// colorCodes maps each level to an ANSI 256-color background code.
var colorCodes = map[int]int{
	colorNone:   236, // near-black (no record)
	colorLight:  22,  // dark green (>0m, below 30m threshold)
	colorMedium: 34,  // medium green (30m+)
	colorDark:   46,  // bright green (60m+)
	colorMax:    82,  // brightest green (120m+)
}

func colorBlock(minutes int) string {
	if minutes == 0 {
		return fmt.Sprintf("\033[48;5;%dm  \033[0m", colorCodes[colorNone])
	}
	level := colorLight // any >0 gets at least colorLight
	for i, threshold := range colorThresholds {
		if minutes >= threshold {
			level = i + 2 // 30m→colorMedium, 60m→colorDark, 120m→colorMax
		}
	}
	return fmt.Sprintf("\033[48;5;%dm  \033[0m", colorCodes[level])
}
