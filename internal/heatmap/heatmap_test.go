package heatmap

import (
	"fmt"
	"testing"
)

func Test_colorBlock(t *testing.T) {
	tests := []struct {
		count    int
		wantCode int
	}{
		{0, 236},
		{1, 22},
		{2, 34},
		{3, 46},
		{10, 46},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count=%d", tt.count), func(t *testing.T) {
			want := fmt.Sprintf("\033[48;5;%dm  \033[0m", tt.wantCode)
			if got := colorBlock(tt.count); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
