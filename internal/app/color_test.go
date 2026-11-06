package app

import (
	"strings"
	"testing"
)

func TestStatusLineColorRules(t *testing.T) {
	old := colorEnabled
	colorEnabled = true
	defer func() { colorEnabled = old }()

	cases := []struct {
		line     string
		wantCode string
	}{
		{"[ok] demo", ansiGreen},
		{"[完成] demo", ansiGreen},
		{"[警告] demo", ansiYellow},
		{"[占用] demo", ansiYellow},
		{"[跳过] demo", ansiYellow},
		{"[不一致] demo", ansiRed},
		{"[inconsistent] demo", ansiRed},
		{"[损坏] demo", ansiRed},
		{"[密码不符] demo", ansiRed},
		{"普通文字", ""},
		{"", ""},
	}
	for _, c := range cases {
		got := statusLineColor(c.line)
		if c.wantCode == "" {
			if got != c.line {
				t.Errorf("statusLineColor(%q) = %q, want it unchanged", c.line, got)
			}
			continue
		}
		if !strings.HasPrefix(got, c.wantCode) {
			t.Errorf("statusLineColor(%q) = %q, want prefix %q", c.line, got, c.wantCode)
		}
		// 颜色只包住标记本身，reset 紧跟在标记后面，不染整行
		if !strings.Contains(got, ansiReset) {
			t.Errorf("statusLineColor(%q) = %q, want a reset after the marker", c.line, got)
		}
		if strings.HasSuffix(got, ansiReset) {
			t.Errorf("statusLineColor(%q) = %q, reset should not wrap the whole line", c.line, got)
		}
	}
}

func TestNoColorWhenDisabled(t *testing.T) {
	old := colorEnabled
	colorEnabled = false
	defer func() { colorEnabled = old }()

	for _, line := range []string{"[ok] demo", "[警告] x", "[损坏] y"} {
		if got := statusLineColor(line); got != line {
			t.Errorf("with color disabled, statusLineColor(%q) = %q", line, got)
		}
	}
}
