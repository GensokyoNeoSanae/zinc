package app

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestFileURI(t *testing.T) {
	cases := map[string]string{
		`C:\Users\asus\zinc\zinc.png`: "file:///C:/Users/asus/zinc/zinc.png",
		`E:\中文 目录\zinc.png`:           "file:///E:/%E4%B8%AD%E6%96%87%20%E7%9B%AE%E5%BD%95/zinc.png",
		"":                            "",
	}
	for in, want := range cases {
		if got := fileURI(in); got != want {
			t.Errorf("fileURI(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToastXML(t *testing.T) {
	// 没有 PNG：不出 image，别给系统一个加载不了的地址
	plain := toastXML("zinc 备份完成", "t1 · 3 文件", "")
	if strings.Contains(plain, "<image") {
		t.Error("no png should mean no logo element")
	}
	if !strings.Contains(plain, "<text>zinc 备份完成</text>") {
		t.Error("title missing")
	}

	// 有 PNG：应用头像要带上，并且路径里的空格被转义
	withLogo := toastXML("标题", "正文", `E:\My Apps\zinc.png`)
	if !strings.Contains(withLogo, `<image placement="appLogoOverride"`) {
		t.Error("logo element missing")
	}
	if !strings.Contains(withLogo, "file:///E:/My%20Apps/zinc.png") {
		t.Errorf("logo uri not escaped: %s", withLogo)
	}

	// 正文里的 XML 特殊字符要转义，否则 LoadXml 直接报错
	escaped := toastXML(`a & b`, `<c> "d" 'e'`, "")
	for _, bad := range []string{"a & b", "<c>", `"d"`} {
		if strings.Contains(escaped, bad) {
			t.Errorf("not escaped: %q in %s", bad, escaped)
		}
	}
	if !strings.Contains(escaped, "a &amp; b") {
		t.Error("ampersand not escaped")
	}

	// 整体必须是合法 XML：LoadXml 一崩，通知就彻底发不出来
	for _, doc := range []string{plain, withLogo, escaped} {
		var v struct{}
		if err := xml.Unmarshal([]byte(doc), &v); err != nil {
			t.Errorf("toast xml is not well-formed: %v\n%s", err, doc)
		}
	}
}

func TestNotifyFingerprint(t *testing.T) {
	base := notifyFingerprint(`C:\app\zinc.exe`, `C:\app\zinc.ico`, `C:\app\zinc.png`)
	if base == notifyFingerprint(`C:\app\zinc.exe`, `C:\app\zinc.ico`, "") {
		t.Error("adding the png must change the fingerprint (that is what re-registers the shortcut)")
	}
	if base == notifyFingerprint(`D:\other\zinc.exe`, `C:\app\zinc.ico`, `C:\app\zinc.png`) {
		t.Error("moving the program must change the fingerprint")
	}
}

func TestShortcutScriptCarriesIcon(t *testing.T) {
	s := shortcutScript(`C:\app\zinc.exe`, `C:\sm\zinc.lnk`, `C:\app\zinc.ico`, `C:\app\zinc.png`)
	for _, want := range []string{
		`$sc.IconLocation = 'C:\app\zinc.ico,0'`,
		`Set-ItemProperty -Path $key -Name 'IconUri' -Value 'C:\app\zinc.png'`,
		`Set-ItemProperty -Path $key -Name 'DisplayName' -Value 'zinc'`,
		`[ZincShortcut]::SetAppId($lnk, 'zinc.backup')`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing from script: %s", want)
		}
	}

	// 没有 PNG 时退回 ICO，也好过什么都不写
	noPng := shortcutScript(`C:\app\zinc.exe`, `C:\sm\zinc.lnk`, `C:\app\zinc.ico`, ``)
	if !strings.Contains(noPng, `-Name 'IconUri' -Value 'C:\app\zinc.ico'`) {
		t.Error("without a png the ico should still be registered as IconUri")
	}

	// 一个图标都没有：不该出现半截属性行（PowerShell 会语法错）
	// 注意别拿 "IconLocation" 判断——C# 那段 COM 接口里本来就有 Get/SetIconLocation。
	bare := shortcutScript(`C:\app\zinc.exe`, `C:\sm\zinc.lnk`, ``, ``)
	if strings.Contains(bare, "$sc.IconLocation") || strings.Contains(bare, "-Name 'IconUri'") {
		t.Error("no icon file should mean no icon properties at all")
	}
	if !strings.Contains(bare, "CreateShortcut") {
		t.Error("script still has to build the shortcut")
	}
}

// 计划任务弹通知时不能闪控制台窗口：参数要带 -WindowStyle Hidden，
// 进程还要带 CREATE_NO_WINDOW——GUI 子系统里只靠参数仍可能先闪一下。
func TestHiddenPowerShell(t *testing.T) {
	cmd := hiddenPowerShellCmd("Write-Output 1")
	joined := strings.Join(cmd.Args, " ")
	for _, want := range []string{"-NoProfile", "-WindowStyle", "Hidden"} {
		if !strings.Contains(joined, want) {
			t.Errorf("缺少 %q: %q", want, joined)
		}
	}
	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr 为空：没有 CREATE_NO_WINDOW，计划任务里会闪窗")
	}
	if cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Errorf("CreationFlags=%#x 不含 CREATE_NO_WINDOW", cmd.SysProcAttr.CreationFlags)
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow 没设")
	}

	// 真的跑一次：CREATE_NO_WINDOW 不该影响 stdout 捕获，否则通知脚本会静默失败
	out, err := hiddenPowerShellCmd("Write-Output hidden-ok").CombinedOutput()
	if err != nil {
		t.Fatalf("隐藏窗口的 PowerShell 执行失败: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "hidden-ok") {
		t.Fatalf("输出没拿到: %q", string(out))
	}
}
