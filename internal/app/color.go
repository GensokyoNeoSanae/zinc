package app

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ANSI 颜色。只在真正的终端里生效：重定向到文件或管道时全部退化为纯文本，
// 否则日志和 `zinc list --names | ...` 里会混进转义码。NO_COLOR 优先。
const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiCyan   = "\x1b[36m"
)

const enableVirtualTerminalProcessing = 0x0004

// kernel32 在 util.go 里已经声明，这里只补两个控制台模式的过程
var (
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// colorEnabled 在启动时就定下来，别每次打印都问一遍系统。
var colorEnabled = detectColor()

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	h := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	if r, _, _ := procGetConsoleMode.Call(uintptr(h), uintptr(unsafe.Pointer(&mode))); r == 0 {
		// 不是控制台：被重定向了，一律不上色
		return false
	}
	// 打开 ANSI 转义支持；老系统上这一步失败，那就不上色
	if r, _, _ := procSetConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminalProcessing)); r == 0 {
		return false
	}
	return true
}

func paint(code, s string) string {
	if !colorEnabled || s == "" {
		return s
	}
	return code + s + ansiReset
}

func green(s string) string  { return paint(ansiGreen, s) }
func yellow(s string) string { return paint(ansiYellow, s) }
func red(s string) string    { return paint(ansiRed, s) }
func cyan(s string) string   { return paint(ansiCyan, s) }

// statusLineColor 按行首的方括号标记挑颜色。
// 只染标记那一段，不染整行——长列表里整行上色非常刺眼，也会盖掉对齐。
// 中英标记都列在这里：这些字符串出现在装配好的输出里，拿不到当前语言。
var colorRules = []struct {
	prefixes []string
	paint    func(string) string
}{
	{[]string{"[ok]", "[完成]", "[done]"}, green},
	{[]string{
		"[警告]", "[warn]", "[占用]", "[locked]", "[跳过]", "[skip]",
		"[预演]", "[dry-run]", "[删除]", "[deleted]", "[提示]", "[note]",
	}, yellow},
	{[]string{
		"[损坏]", "[corrupt]", "[密码不符]", "[错误]", "[error]", "[失败]", "[failed]",
	}, red},
}

func statusLineColor(line string) string {
	for _, r := range colorRules {
		for _, p := range r.prefixes {
			if strings.HasPrefix(line, p) {
				return r.paint(p) + line[len(p):]
			}
		}
	}
	return line
}

// printColored 打印一行并上色。只用在人读的输出上；
// `--names` 这类给脚本吃的格式不要走这里。
func printColored(format string, args ...any) {
	fmt.Print(statusLineColor(fmt.Sprintf(format, args...)))
}
