package web

import (
	"bytes"
	"strings"
	"testing"
)

func TestPageRender(t *testing.T) {
	data := &pageData{
		UpdatedAt: "2026-07-11 10:00",
		Habits: []habitView{{
			Name: "筋トレ", Streak: 3, Total: 10,
			Weeks: [][7]cell{{{Date: "2026-07-05", Done: true}, {Date: "2026-07-06"}, {Date: "2026-07-07", Done: true}, {Date: "2026-07-08"}, {Date: "2026-07-09"}, {Date: "2026-07-10"}, {Date: "2026-07-11", Today: true, Future: false}}},
		}},
	}
	var buf bytes.Buffer
	if err := page.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"筋トレ", "🔥 3日連続", "計 10回", `class="cell done"`, `class="cell today"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}
