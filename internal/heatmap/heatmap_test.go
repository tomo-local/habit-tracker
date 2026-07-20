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
		{0, 236},   // colorNone — gray only for zero
		{15, 22},   // colorLight — any >0 gets at least light
		{30, 34},   // colorMedium
		{60, 46},   // colorDark
		{120, 82},  // colorMax
		{180, 82},  // above 120m → colorMax
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
