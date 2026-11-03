package app

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
)

const enableVirtualTerminalProcessing = 0x0004

// kernel32 在 util.go 里已经声明，这里只补需要的几个过程
var (
	procGetConsoleMode       = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode       = kernel32.NewProc("SetConsoleMode")
	procGetConsoleBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	fallbackTerminalWidth    = 80
)

type coord struct {
	X, Y int16
}

type smallRect struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

// colorEnabled 在启动时就定下来，别每次打印都问一遍系统。
var colorEnabled = detectColor()

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	h := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	if r, _, _ := procGetConsoleMode.Call(uintptr(h), uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	if r, _, _ := procSetConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminalProcessing)); r == 0 {
		return false
	}
	return true
}

// termWidth 返回当前终端宽度。拿不到（被重定向、非常规控制台）就退回 80。
func termWidth() int {
	h := syscall.Handle(os.Stdout.Fd())
	var info consoleScreenBufferInfo
	if r, _, _ := procGetConsoleBufferInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&info))); r == 0 {
		return fallbackTerminalWidth
	}
	w := int(info.Size.X)
	if w < 20 {
		return fallbackTerminalWidth
	}
	return w
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

// blue 用于中性的「接下来该做什么」提示：不是状态，是指导。
func blue(s string) string { return paint(ansiBlue, s) }

// statusLineColor 按行首的方括号标记挑颜色。
// 只染标记那一段，不染整行——长列表里整行上色非常刺眼，也会盖掉对齐。
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
		"[更改]", "[changed]", "[损坏]", "[corrupt]",
		"[密码不符]", "[错误]", "[error]", "[失败]", "[failed]",
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
