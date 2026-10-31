package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// 命令表本身的自检：名字不重复、flag 不带横线、必须有仓库名补全的入口。
func TestCompletionSpecTable(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range completionCommands {
		if c.Name == "" || seen[c.Name] {
			t.Errorf("命令名重复或为空: %q", c.Name)
		}
		seen[c.Name] = true
		for _, f := range c.Flags {
			if f.Name == "" || strings.HasPrefix(f.Name, "-") {
				t.Errorf("%s: flag 名要存不带横线的形式: %q", c.Name, f.Name)
			}
		}
	}
	for _, want := range []string{"backup", "restore", "schedule", "config", "completion"} {
		if !seen[want] {
			t.Errorf("命令表漏了 %q", want)
		}
	}
	if len(repoCommands()) == 0 {
		t.Error("repoCommands 为空")
	}
}

// 四个 shell 的脚本都要生成得出来，并且带上关键注册语句与动态仓库名调用。
func TestCompletionScripts(t *testing.T) {
	markers := map[string]string{
		"powershell": "Register-ArgumentCompleter -Native -CommandName zinc",
		"bash":       "complete -F _zinc zinc",
		"zsh":        "#compdef zinc",
		"fish":       "complete -c zinc",
	}
	for shell, marker := range markers {
		s, ok := completionScript(shell)
		if !ok || s == "" {
			t.Fatalf("%s: 没生成脚本", shell)
		}
		if !strings.Contains(s, marker) {
			t.Errorf("%s: 缺注册语句 %q", shell, marker)
		}
		if !strings.Contains(s, "zinc list --names") {
			t.Errorf("%s: 没有用 zinc list --names 取仓库名", shell)
		}
		if !strings.Contains(s, "schedule") {
			t.Errorf("%s: 命令表没进脚本", shell)
		}
		if shell == "powershell" && !strings.Contains(s, "OutputEncoding") {
			t.Errorf("powershell: 取仓库名时没临时切 UTF-8，中文仓库名会变「鏂囨。」")
		}
	}
	if _, ok := completionScript("nu"); ok {
		t.Error("不支持的 shell 应该返回 false")
	}
}

// 防漂移：把 zinc 真编出来，对每条命令跑 --help，和命令表双向对账。
// 以后加参数忘了补表，或者表里写了不存在的参数，这条测试会红。
func TestCompletionMatchesHelp(t *testing.T) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go 不在 PATH，跳过对账")
	}
	bin := filepath.Join(t.TempDir(), "zinc.exe")
	build := exec.Command(goExe, "build", "-o", bin, "../../cmd/zinc")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}

	// list --names 是补全的数据源，顺手做一次端到端：配置放在 exe 同目录（appDir）
	cfg := `{"version":2,"repos":{` +
		`"图片":{"source":"C:\\src1","output":"C:\\out1"},` +
		`"文档":{"source":"C:\\src2","output":"C:\\out2"}}}`
	if err := os.WriteFile(filepath.Join(filepath.Dir(bin), "repos.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _ := exec.Command(bin, "list", "--names").Output()
	if got := strings.TrimSpace(string(out)); got != "图片\n文档" {
		t.Errorf("list --names = %q, want 图片/文档 两行", got)
	}

	flagLine := regexp.MustCompile(`(?m)^\s{2}-([A-Za-z0-9][A-Za-z0-9-]*)`)
	for _, c := range completionCommands {
		if c.Name == "help" {
			continue // help 走整份 usage，没有自己的 flag 表
		}
		raw, _ := exec.Command(bin, c.Name, "--help").CombinedOutput()
		help := string(raw)
		got := map[string]bool{}
		for _, m := range flagLine.FindAllStringSubmatch(help, -1) {
			got[m[1]] = true
		}

		want := map[string]bool{}
		for _, f := range c.Flags {
			want[f.Name] = true
			if !got[f.Name] {
				t.Errorf("%s: --%s 在命令表里，但 --help 里没有", c.Name, f.Name)
			}
		}
		for name := range got {
			if name == "h" || name == "help" {
				continue
			}
			if !want[name] {
				t.Errorf("%s: --help 里有 --%s，命令表没登记", c.Name, name)
			}
		}
	}
}
