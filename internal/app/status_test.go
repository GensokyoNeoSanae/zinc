package app

import "testing"

func TestDisplayWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{"图片", 4},    // 两个中文字 = 四列
		{"图片abc", 7}, // 4 + 3
		{"文档 1", 6},  // 4 + 空格 + 1
		{"（全角）", 8},  // 全角括号也占两列
	}
	for _, c := range cases {
		if got := displayWidth(c.in); got != c.want {
			t.Errorf("displayWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestPadAlignsByDisplayWidth(t *testing.T) {
	// 「图片」和「abcd」显示宽度相同，补完空格后长度必须一致
	if got, want := displayWidth(padRight("图片", 6)), displayWidth(padRight("abcd", 6)); got != want {
		t.Fatalf("padded widths differ: %d vs %d", got, want)
	}
	if got := padLeft("7", 3); got != "  7" {
		t.Errorf("padLeft(7,3) = %q", got)
	}
}

func TestCommaInt(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		7:       "7",
		999:     "999",
		1000:    "1,000",
		12345:   "12,345",
		1234567: "1,234,567",
	}
	for in, want := range cases {
		if got := commaInt(in); got != want {
			t.Errorf("commaInt(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestSizeDelta(t *testing.T) {
	// 源为空：不知道比例，只能给占位符
	if got := sizeDelta(0, 100); got != M.SizeUnknown {
		t.Errorf("sizeDelta(0,100) = %q, want %q", got, M.SizeUnknown)
	}
	// 相同
	if got := sizeDelta(1000, 1000); got != M.SizeSame {
		t.Errorf("sizeDelta equal = %q, want %q", got, M.SizeSame)
	}
	// 输出更大（括号由调用方加，这里只要数值部分）
	if got := sizeDelta(1000, 1100); got != "+10.0%, 100 B" {
		t.Errorf("sizeDelta(+10%%) = %q", got)
	}
	// 输出更小：说「少占」，不写负数百分比
	if got := sizeDelta(1000, 800); got != "-20.0%, 200 B" {
		t.Errorf("sizeDelta(-20%%) = %q", got)
	}
}

func TestSortStatusDirections(t *testing.T) {
	mk := func() []repoStatus {
		return []repoStatus{
			{name: "乙", outSize: 10, srcSize: 10, last: "2026-01-02"},
			{name: "甲", outSize: 20, srcSize: 20, last: "2026-01-01"},
		}
	}
	// 内部一律升序 + desc 反转：out 降序时大的在前
	list := mk()
	sortStatus(list, "out", true)
	if list[0].name != "甲" {
		t.Errorf("out desc: first = %q, want 甲", list[0].name)
	}
	// out 升序时小的在前
	list = mk()
	sortStatus(list, "out", false)
	if list[0].name != "乙" {
		t.Errorf("out asc: first = %q, want 乙", list[0].name)
	}
	// backup 按时间字符串比较，desc 时新的在前
	list = mk()
	sortStatus(list, "backup", true)
	if list[0].last != "2026-01-02" {
		t.Errorf("backup desc: first = %q, want 2026-01-02", list[0].last)
	}
}
