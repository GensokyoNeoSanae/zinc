package app

import (
	"strings"
	"testing"
)

// -h 曾经在 remove 上被当成仓库名去查。这两条守住「用法可问、退出码 0」。
func TestCLIHelpFlags(t *testing.T) {
	bin := buildZincCLI(t)

	out, code := runZinc(t, bin, "remove", "--help")
	if code != 0 {
		t.Errorf("zinc remove --help exited %d, want 0 (out=%q)", code, out)
	}
	if strings.Contains(out, "仓库不存在") {
		t.Errorf("remove treated --help as a repo name: %q", out)
	}
	if !strings.Contains(out, "zinc remove") {
		t.Errorf("remove --help output = %q, want the usage line", out)
	}

	out, code = runZinc(t, bin, "root", "-h")
	if code != 0 || !strings.Contains(out, "zinc root") {
		t.Errorf("root -h = (%q, %d), want the usage line and 0", out, code)
	}
}

// 中英文用法说明的条目数必须一致：一边加了说明、另一边忘了，用户就会
// 在另一种语言里少看到一条规则，这种漂移没人会主动发现。
func TestUsageNotesParity(t *testing.T) {
	zh := countNoteLines(zhMessages.Usage)
	en := countNoteLines(enMessages.Usage)
	if zh != en {
		t.Fatalf("zh usage has %d note lines, en has %d", zh, en)
	}
}

func countNoteLines(usage string) int {
	n := 0
	for _, line := range strings.Split(usage, "\n") {
		if strings.HasPrefix(line, "  - ") {
			n++
		}
	}
	return n
}
