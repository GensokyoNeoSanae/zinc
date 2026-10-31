package app

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildZincCLI 编译出真 exe：参数解析和退出码只有跑真进程才算数。
func buildZincCLI(t *testing.T) string {
	t.Helper()
	goExe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go 不在 PATH，跳过 CLI 级测试")
	}
	bin := filepath.Join(t.TempDir(), "zinc.exe")
	cmd := exec.Command(goExe, "build", "-o", bin, "../../cmd/zinc")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build zinc: %v", err)
	}
	return bin
}

// writeCLIConfig 在 exe 同目录（appDir）放一份最小配置。
func writeCLIConfig(t *testing.T, bin string) {
	t.Helper()
	dir := filepath.Dir(bin)
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := map[string]any{
		"version": 2,
		"repos": map[string]any{
			"demo": map[string]any{"source": src, "output": filepath.Join(dir, "out")},
			"图片":   map[string]any{"source": src, "output": filepath.Join(dir, "out2")},
		},
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "repos.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	set := map[string]any{
		"version": 1, "default_password": "PW",
		"7zip_path": `E:\Tools\7-Zip\7z.exe`, "complete_notify": false,
	}
	b2, _ := json.MarshalIndent(set, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), b2, 0o644); err != nil {
		t.Fatal(err)
	}
}

func runZinc(t *testing.T, bin string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	// 断言的是中文文案；CI 的系统语言是英文，不钉住语言就会跟着环境变。
	cmd.Env = append(os.Environ(), "ZINC_LANG=zh")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run %v: %v", args, err)
	return "", -1
}

// 歧义写法必须显式报错，而不是静默走某一边。
func TestCLIRejectsAmbiguousArgs(t *testing.T) {
	bin := buildZincCLI(t)
	writeCLIConfig(t, bin)
	dest := filepath.Dir(bin)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"backup 多余参数", []string{"backup", "demo", "extra"}, "多余的参数: extra"},
		{"backup --all 带名字", []string{"backup", "demo", "--all"}, "--all 不能和名字一起用: demo"},
		// 布尔参数误写成空格 + 值：false 变成多余位置参数，必须报错而不是静默忽略
		{"backup 布尔空格", []string{"backup", "demo", "--notify", "false"}, "多余的参数: false"},
		{"verify 多余参数", []string{"verify", "demo", "extra"}, "多余的参数: extra"},
		{"verify --all 带名字", []string{"verify", "demo", "--all"}, "--all 不能和名字一起用: demo"},
		{"status 多余参数", []string{"status", "demo", "extra"}, "多余的参数: extra"},
		{"remove 多余参数", []string{"remove", "demo", "extra"}, "多余的参数: extra"},
		{"set 多余参数", []string{"set", "demo", "extra"}, "多余的参数: extra"},
		{"restore 多余参数", []string{"restore", "demo", "extra", "--dest", dest}, "多余的参数: extra"},
		{"reindex 多余参数", []string{"reindex", "demo", "extra"}, "多余的参数: extra"},
		{"compact 多余参数", []string{"compact", "demo", "extra"}, "多余的参数: extra"},
		{"schedule 多余参数", []string{"schedule", "demo", "--daily", "--at", "20:00", "extra"}, "多余的参数: extra"},
		{"schedule --list 带周期", []string{"schedule", "--list", "--daily"}, "只查询任务"},
		{"config get 多余参数", []string{"config", "get", "lang", "extra"}, "多余的参数: extra"},
		{"config set 多余参数", []string{"config", "set", "lang", "zh", "extra"}, "多余的参数: extra"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, code := runZinc(t, bin, c.args...)
			if code != 1 {
				t.Fatalf("退出码 = %d, want 1\n%s", code, out)
			}
			if !strings.Contains(out, c.want) {
				t.Errorf("输出里没有 %q:\n%s", c.want, out)
			}
		})
	}
}

// 收紧参数后，这些合法写法不能误伤。
func TestCLIStillAcceptsLegitimateForms(t *testing.T) {
	bin := buildZincCLI(t)
	writeCLIConfig(t, bin)

	// --notify=false 是正确写法，不能当成多余参数
	if out, code := runZinc(t, bin, "backup", "demo", "--notify=false"); code != 0 {
		t.Fatalf("--notify=false 被误杀: %d\n%s", code, out)
	}
	// config set 的 value 恰好是 false，必须能写进去
	if out, code := runZinc(t, bin, "config", "set", "complete_notify", "false"); code != 0 {
		t.Fatalf("config set false 被误杀: %d\n%s", code, out)
	}
	// -h 的语义要和其它命令一致：打印用法、退出码 0
	for _, args := range [][]string{{"root", "--help"}, {"config", "--help"}} {
		out, code := runZinc(t, bin, args...)
		if code != 0 || !strings.Contains(out, "用法") {
			t.Errorf("%v: code=%d, out=%s", args, code, out)
		}
	}
	// 非法 --lang 会回退到系统语言，所以警告本身可能是中英任一；
	// 用语言无关的断言：退出码 0、把非法值说出来。
	out, code := runZinc(t, bin, "--lang", "xx", "list")
	if code != 0 || !strings.Contains(out, "xx") ||
		(!strings.Contains(out, "语言") && !strings.Contains(out, "language")) {
		t.Errorf("--lang xx: code=%d, out=%s", code, out)
	}
	// 非控制台跑 unschedule --all：读不到回答就取消，不能默默删任务
	out, code = runZinc(t, bin, "unschedule", "--all")
	if code != 1 || !strings.Contains(out, "已取消") {
		t.Errorf("unschedule --all: code=%d, out=%s", code, out)
	}
}
