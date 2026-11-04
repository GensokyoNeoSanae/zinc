package app

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// defaultLang 可通过 -ldflags "-X main.defaultLang=zh|en" 在编译时指定
var defaultLang = "auto"

// cliLang 是命令行显式给出的 --lang 值，空表示没给。
// 单独使用（`zinc --lang en`，不带子命令）时它表示「改设置」而不是「临时看一眼」，
// 由 Main 负责写进 settings.json。
var cliLang string

// initLang 提取并剥离 --lang 参数，选定全局 M，返回剩余参数
func initLang(args []string) []string {
	// 先给 M 一个默认值：读 settings.json 时可能要报错，M 不能是 nil
	M = enMessages

	lang := ""
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if (a == "--lang" || a == "-lang") && i+1 < len(args) {
			lang = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(a, "--lang=") {
			lang = strings.TrimPrefix(a, "--lang=")
			continue
		}
		if strings.HasPrefix(a, "-lang=") {
			lang = strings.TrimPrefix(a, "-lang=")
			continue
		}
		rest = append(rest, a)
	}
	cliLang = lang

	if lang == "" {
		lang = os.Getenv("ZINC_LANG")
	}
	if lang == "" {
		lang = loadSettingsQuiet().Lang
	}
	if lang == "" {
		lang = defaultLang
	}
	if lang == "" || strings.EqualFold(lang, "auto") {
		lang = systemLang()
	}
	if !setLang(lang) {
		// 非法语言码回落到系统语言，但要把它说出来——静默改语言是歧义
		sys := systemLang()
		if !setLang(sys) {
			sys = "en"
			setLang(sys)
		}
		fmt.Fprintf(os.Stderr, M.LangInvalid+"\n", lang, sys)
	}
	return rest
}

func setLang(lang string) bool {
	switch normalizeLang(lang) {
	case "zh":
		M = zhMessages
		return true
	case "en":
		M = enMessages
		return true
	}
	return false
}

// normalizeLang 把各种写法收敛成 zh / en，认不出来就原样返回
func normalizeLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "zh", "zh-cn", "zh-hans", "chinese":
		return "zh"
	case "en", "en-us", "english":
		return "en"
	}
	return lang
}

// persistLang 把语言写进 settings.json。只在 `zinc --lang en` 单独使用时调用：
// 带子命令的写法仍然是「这次用这个语言跑」。
func persistLang(lang string) error {
	s, err := loadSettings()
	if err != nil {
		return err
	}
	s.Lang = normalizeLang(lang)
	return saveSettings(s)
}

// systemLang 读 Windows UI 语言（LANG_CHINESE = 0x04）
func systemLang() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetUserDefaultUILanguage")
	r, _, _ := proc.Call()
	if r == 0 {
		return "en"
	}
	primary := uint16(r) & 0x03FF
	if primary == 0x04 {
		return "zh"
	}
	return "en"
}
