package heatmap

import (
	"fmt"
	"testing"
)

func Test_colorBlock(t *testing.T) {
	tests := []struct {
		minutes  int
		wantCode int
	}{
		{0, 236},   // colorNone
		{15, 236},  // below 30m threshold → still colorNone
		{30, 22},   // colorLight
		{60, 34},   // colorMedium
		{120, 46},  // colorDark
		{180, 46},  // above 120m → colorDark
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%dmin", tt.minutes), func(t *testing.T) {
			want := fmt.Sprintf("\033[48;5;%dm  \033[0m", tt.wantCode)
			if got := colorBlock(tt.minutes); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
