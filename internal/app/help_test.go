package app

import (
	"strings"
	"testing"
)

// 折行时标点不能落在行首——中文排版的禁则
func TestWrapLineNoLeadingPunct(t *testing.T) {
	text := "简表的数字来自索引和账本；「源」和「输出」是两边目录的真实遍历——目录项自带大小，不用打开文件，30 GB 量级也就百来毫秒。"
	for _, limit := range []int{20, 30, 40, 50, 60} {
		for _, line := range wrapLine(text, limit) {
			if line == "" {
				continue
			}
			first := []rune(line)[0]
			if strings.ContainsRune(noLineStart, first) {
				t.Errorf("limit=%d: line starts with punctuation %q: %q", limit, string(first), line)
			}
		}
	}
}

// 折行后拼回去必须和原文一致（只允许丢掉断点处的空格）
func TestWrapLineKeepsContent(t *testing.T) {
	text := "任意一段中文文本，里面混着 English words 和数字 123，用来验证折行不吞字。"
	got := strings.Join(wrapLine(text, 20), "")
	want := strings.ReplaceAll(text, " ", "")
	if strings.ReplaceAll(got, " ", "") != want {
		t.Errorf("content changed:\n got %q\nwant %q", got, want)
	}
}
