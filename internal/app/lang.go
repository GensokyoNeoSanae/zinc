package app

import (
	"os"
	"strings"
	"syscall"
)

// defaultLang 可通过 -ldflags "-X main.defaultLang=zh|en" 在编译时指定
var defaultLang = "auto"

// initLang 提取并剥离 --lang 参数，选定全局 M，返回剩余参数
func initLang(args []string) []string {
	// 先给 M 一个默认值：读 settings.json 时可能要报错，M 不能是 nil
	M = enMessages

	lang := ""
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--lang" && i+1 < len(args) {
			lang = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(a, "--lang=") {
			lang = strings.TrimPrefix(a, "--lang=")
			continue
		}
		rest = append(rest, a)
	}

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
		// 非法语言码回落到系统语言，而不是直接写死 en
		if !setLang(systemLang()) {
			setLang("en")
		}
	}
	return rest
}

func setLang(lang string) bool {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "zh", "zh-cn", "zh-hans", "chinese":
		M = zhMessages
		return true
	case "en", "en-us", "english":
		M = enMessages
		return true
	}
	return false
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
