package app

import (
	"strings"
	"testing"
)

func TestVersionLine(t *testing.T) {
	old := Version
	Version = "v9.9.9"
	defer func() { Version = old }()

	got := versionLine()
	if !strings.HasPrefix(got, "zinc ") {
		t.Fatalf("versionLine() = %q, want it to start with zinc", got)
	}
	if !strings.Contains(got, "v9.9.9") {
		t.Fatalf("versionLine() = %q, want it to contain the injected version", got)
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("versionLine() must stay a single line, got %q", got)
	}
}

// 版本号三种写法都要能用，且退出码必须是 0：查版本不是错误。
func TestCLIVersion(t *testing.T) {
	bin := buildZincCLI(t)

	for _, form := range []string{"version", "--version", "-v"} {
		out, code := runZinc(t, bin, form)
		if code != 0 {
			t.Errorf("zinc %s exited %d, want 0 (out=%q)", form, code, out)
			continue
		}
		line := strings.TrimSpace(out)
		if !strings.HasPrefix(line, "zinc ") {
			t.Errorf("zinc %s output = %q, want it to start with zinc", form, out)
		}
		// 测试二进制不带 -ldflags，所以版本应当落在 dev 上。
		if !strings.Contains(line, "dev") {
			t.Errorf("zinc %s output = %q, want the uninjected build to say dev", form, out)
		}
	}

	// 不带参数时先打版本再打用法，双击快捷方式的人也能看见版本。
	out, _ := runZinc(t, bin)
	first := strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]
	if !strings.HasPrefix(first, "zinc ") {
		t.Fatalf("first line without args = %q, want the version line", first)
	}
}
