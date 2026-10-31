package app

import (
	"path/filepath"
	"strings"
	"testing"
)

// --password-only / --hash-only 要真的改变检查范围，而不是换个说法。
func TestVerifyModes(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	files := writeFiles(t, src, "f", 2, 256)
	repo := &Repo{Source: src, Output: out}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}
	good := &Settings{SevenZipPath: sz, DefaultPassword: "PACKPW"}
	bad := &Settings{SevenZipPath: sz, DefaultPassword: "WRONGPW"}

	pack := "20260101_000000_000.7z"
	res, err := packOne(repo, good, packPlan{Files: files}, pack, "PACKPW", false)
	if err != nil {
		t.Fatalf("packOne: %v", err)
	}
	if err := writeManifest(out, "t", "2026-01-01_00-00-00", batchMeta{
		Packs:  []string{pack},
		Hashes: []string{res.Hash},
		Hash:   res.Hash,
	}); err != nil {
		t.Fatalf("writeManifest: %v", err)
	}

	// 默认：哈希 + 密码都查，密码不对必须失败
	if err := runVerify("t", cfg, bad, 100, true, true); err == nil {
		t.Fatal("默认模式：错密码应该失败")
	}
	// --hash-only：不碰密码，包哈希对就通过
	if err := runVerify("t", cfg, bad, 100, true, false); err != nil {
		t.Fatalf("--hash-only 不该关心密码: %v", err)
	}
	// --password-only：错密码必须失败
	if err := runVerify("t", cfg, bad, 100, false, true); err == nil {
		t.Fatal("--password-only：错密码应该失败")
	}
	// 密码正确时两种模式都通过
	if err := runVerify("t", cfg, good, 100, false, true); err != nil {
		t.Fatalf("--password-only 正确密码: %v", err)
	}
	if err := runVerify("t", cfg, good, 100, true, false); err != nil {
		t.Fatalf("--hash-only: %v", err)
	}
	// 没有配置密码时，--password-only 不能假装通过
	if err := runVerify("t", cfg, &Settings{SevenZipPath: sz}, 100, false, true); err == nil {
		t.Fatal("没有密码时 --password-only 应该报错")
	}
}

// toast 正文固定压成两行：Windows 只渲染 3 个 <text>（标题 + 2 行）。
func TestSummaryBodyIsTwoLines(t *testing.T) {
	saved := M
	M = zhMessages
	defer func() { M = saved }()

	results := []BackupResult{
		{Repo: "图片", NewCount: 2, TotalSize: 3 << 20},
		{Repo: "文档", NewCount: 1, TotalSize: 1 << 20},
		{Repo: "其余备份"},
		{Repo: "学校课件"},
		{Repo: "肥鱼记忆"},
		{Repo: "肥鱼技能"},
	}
	body, ok := summaryBody(results, false)
	if !ok {
		t.Fatal("有变化时应该有通知")
	}
	lines := strings.Split(body, "\n")
	if len(lines) != 2 {
		t.Fatalf("正文应压成 2 行，实际 %d 行:\n%s", len(lines), body)
	}
	if !strings.Contains(lines[0], "2 个仓库有新增") || !strings.Contains(lines[0], "4 个无新增") {
		t.Errorf("第一行统计不对: %q", lines[0])
	}
	if !strings.Contains(lines[1], "图片") || !strings.Contains(lines[1], "文档") {
		t.Errorf("第二行缺少有变化的仓库: %q", lines[1])
	}

	// 有失败时：第一行给失败数，第二行带上失败仓库名
	body, _ = summaryBody(append(results, BackupResult{Repo: "视频", Err: errTest}), false)
	lines = strings.Split(body, "\n")
	if len(lines) != 2 {
		t.Fatalf("有失败时正文应压成 2 行，实际 %d 行:\n%s", len(lines), body)
	}
	if !strings.Contains(lines[0], "1 个失败") {
		t.Errorf("第一行缺少失败数: %q", lines[0])
	}
	if !strings.Contains(lines[1], "失败: 视频") {
		t.Errorf("第二行缺少失败仓库: %q", lines[1])
	}

	// 全部无新增且没开 --notify-empty：不发通知
	if _, ok := summaryBody([]BackupResult{{Repo: "a"}, {Repo: "b"}}, false); ok {
		t.Error("全部无新增时不该发通知")
	}
	// 开了 --notify-empty：发一行“全部无新增”
	body, ok = summaryBody([]BackupResult{{Repo: "a"}, {Repo: "b"}}, true)
	if !ok || !strings.Contains(body, "全部 2 个仓库无新增") {
		t.Errorf("notify-empty 正文不对: %q", body)
	}
}

// errTest 给 summaryBody 造一个失败结果。
var errTest = &testError{}

type testError struct{}

func (*testError) Error() string { return "boom" }
