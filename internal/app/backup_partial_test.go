package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// test7z 找一个可用的 7-Zip。CI 上没装就跳过，不让测试变成环境依赖。
func test7z(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("ZINC_TEST_7ZIP"); p != "" && fileExists(p) {
		return p
	}
	if p := defaultSettings().sevenZipPath(); p != "" {
		return p
	}
	t.Skip("skipping: 没找到 7-Zip（可设 ZINC_TEST_7ZIP 指向 7z.exe）")
	return ""
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
}

// writeFiles 造 n 个固定大小的小文件，并返回它们的 fileInfo。
func writeFiles(t *testing.T, dir, prefix string, n, size int) []fileInfo {
	t.Helper()
	out := make([]fileInfo, 0, n)
	for i := 0; i < n; i++ {
		rel := fmt.Sprintf("%s%d.bin", prefix, i)
		full := filepath.Join(dir, rel)
		if err := os.WriteFile(full, bytes.Repeat([]byte{byte('a' + i)}, size), 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
		st, err := os.Stat(full)
		if err != nil {
			t.Fatalf("stat %s: %v", full, err)
		}
		out = append(out, fileInfo{Rel: rel, Size: st.Size(), MTime: st.ModTime()})
	}
	return out
}

func TestPackResultKept(t *testing.T) {
	files := []fileInfo{{Rel: "a"}, {Rel: "b"}, {Rel: "c"}}

	// 没有跳过时原样返回，并且不该额外分配。
	r := packResult{Files: files}
	if got := r.Kept(); len(got) != 3 {
		t.Fatalf("no skip: got %d, want 3", len(got))
	}

	r = packResult{Files: files, Skipped: []string{"b"}}
	got := r.Kept()
	if len(got) != 2 || got[0].Rel != "a" || got[1].Rel != "c" {
		t.Fatalf("Kept() = %v, want [a c]", got)
	}

	// 全被跳过 → 空
	r = packResult{Files: files, Skipped: []string{"a", "b", "c"}}
	if got := r.Kept(); len(got) != 0 {
		t.Fatalf("Kept() = %v, want empty", got)
	}
}

func TestExitCodeOf(t *testing.T) {
	// 正常退出
	if code := exitCodeOf(nil); code != 0 {
		t.Errorf("nil err -> %d, want 0", code)
	}
	// 进程起来但非 0 退出
	err := exec.Command("cmd", "/c", "exit", "3").Run()
	if code := exitCodeOf(err); code != 3 {
		t.Errorf("exit 3 -> %d, want 3", code)
	}
	// 进程根本没起来（不是 ExitError）：-1，调用方会当硬失败
	if code := exitCodeOf(errors.New("boom")); code != -1 {
		t.Errorf("plain error -> %d, want -1", code)
	}
}

// 一个计划里的文件在「扫描之后、打包之前」消失了：7-Zip 报 WARNING 并跳过它，
// 但归档必须留下、其余文件必须进包。这是「被占用不再毁掉整批」的地基。
func TestPackOneKeepsArchiveWhenAFileVanished(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	files := writeFiles(t, src, "f", 5, 4096)
	// 第 2 号文件消失（等价于打包瞬间被别人独占/删掉）
	if err := os.Remove(filepath.Join(src, "f2.bin")); err != nil {
		t.Fatal(err)
	}

	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz}

	r, err := packOne(repo, set, packPlan{Files: files}, "20260101_000000_000.7z", "TESTPW", false)
	if err != nil {
		t.Fatalf("packOne: %v（一个文件读不到不该让整包失败）", err)
	}
	if len(r.Skipped) != 1 || r.Skipped[0] != "f2.bin" {
		t.Fatalf("Skipped = %v, want [f2.bin]", r.Skipped)
	}
	if kept := r.Kept(); len(kept) != 4 {
		t.Fatalf("Kept() = %d 个, want 4", len(kept))
	}
	if !fileExists(filepath.Join(out, r.Name)) {
		t.Fatal("归档应该留在盘上")
	}
}

// 中途某个包提交失败：此前已提交的包必须还在，绝不能被「整批回退」带走。
func TestPackAllKeepsCommittedPacksOnLaterFailure(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	mustMkdir(t, src)

	files := writeFiles(t, src, "g", 8, 4096)

	out := filepath.Join(dir, "out")
	mustMkdir(t, out)
	repo := &Repo{Source: src, Output: out, MaxPackByte: 10 * 1024} // 每包约 2 个文件
	set := &Settings{SevenZipPath: sz}

	calls := 0
	submitted, err := packAll(repo, set, files, "TESTPW", true, func(r packResult) error {
		calls++
		if calls == 2 {
			return errors.New("模拟第 2 个包提交失败（磁盘满）")
		}
		return nil
	})
	if err == nil {
		t.Fatal("应该返回错误")
	}
	if len(submitted) != 1 {
		t.Fatalf("已提交包数 = %d, want 1", len(submitted))
	}
	if !fileExists(filepath.Join(out, submitted[0].Name)) {
		t.Fatalf("已提交的包 %s 不该被删", submitted[0].Name)
	}
}

// commitPack 只登记真正进包的文件；被跳过的必须留在索引外，等下次重试。
func TestCommitPackRegistersOnlyKeptFiles(t *testing.T) {
	out := t.TempDir()
	const packName = "20260101_000000_000.7z"
	if err := os.WriteFile(filepath.Join(out, packName), []byte("fake-archive"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := &Repo{Output: out}
	idx := newIndex()
	r := packResult{
		Name:    packName,
		Hash:    "deadbeef",
		Size:    12,
		Start:   "20260101",
		End:     "20260101",
		Files:   []fileInfo{{Rel: "a.txt", Size: 10}, {Rel: "b.txt", Size: 20}},
		Skipped: []string{"b.txt"},
	}
	if err := commitPack(repo, "t", "2026-01-01_00-00-00", idx, r); err != nil {
		t.Fatalf("commitPack: %v", err)
	}
	if _, ok := idx.Files["a.txt"]; !ok {
		t.Error("a.txt 进了包，应该登记")
	}
	if _, ok := idx.Files["b.txt"]; ok {
		t.Error("b.txt 被跳过了，不该登记（否则下次不会重试）")
	}

	// 落盘的索引要能原样读回来（# end 计数校验）
	back, err := loadIndex(out)
	if err != nil {
		t.Fatalf("loadIndex: %v", err)
	}
	if len(back.Files) != 1 {
		t.Fatalf("索引里 %d 个文件, want 1", len(back.Files))
	}

	// manifest 必须已经登记这个包，否则 verify 永远看不见它
	m, err := loadManifest(out)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if len(m.Batches) != 1 {
		t.Fatalf("manifest 批次数 = %d, want 1", len(m.Batches))
	}
	if len(m.Batches[0].Files) != 1 || m.Batches[0].Files[0] != packName {
		t.Fatalf("manifest 里的包 = %v, want [%s]", m.Batches[0].Files, packName)
	}
}

// 扫描之后、打包之前文件消失：runBackup 必须照常完成，把这个文件排除在索引外，
// 并在结果里计入（下次自动重试）。
//
// 注意这里走的是**预检**那条路（os.Open 失败 → locked），因为预检就紧跟在扫描之后。
// 7-Zip 的 warning 路径覆盖的是「预检那一刻还能读、轮到它时读不到」，由
// TestPackOneKeepsArchiveWhenAFileVanished 单独钉住。两条路的出口完全一样：
// 不进索引、下次重试、其余照常。
func TestRunBackupSkipsVanishedFileAndKeepsGoing(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	writeFiles(t, src, "h", 4, 4096)

	set := &Settings{SevenZipPath: sz, DefaultPassword: "PW", MaxPackByte: 10 * 1024}
	cfg := &Config{Repos: map[string]*Repo{"t": {Source: src, Output: out}}}

	idx, err := loadIndex(out)
	if err != nil {
		t.Fatal(err)
	}
	scan, err := scanSource(src, idx, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if scan.Changed() != 4 {
		t.Fatalf("扫描到 %d 个变更, want 4", scan.Changed())
	}

	// 扫描已完成，这时候文件被删掉（或被人独占）——预检和扫描都看不见它了。
	if err := os.Remove(filepath.Join(src, "h2.bin")); err != nil {
		t.Fatal(err)
	}

	res := runBackup("t", cfg, set, scan, "PW", "settings", true)
	if res.Err != nil {
		t.Fatalf("runBackup: %v（一个文件消失不该让整次备份失败）", res.Err)
	}
	if res.Locked != 1 {
		t.Fatalf("res.Locked = %d, want 1（预检会先兜住「扫描后消失」）", res.Locked)
	}
	if res.Skipped != 0 {
		t.Fatalf("res.Skipped = %d, want 0（这个文件没进 ready，不算包内跳过）", res.Skipped)
	}
	if len(res.Packs) == 0 {
		t.Fatal("其余文件应该照常成包")
	}

	back, err := loadIndex(out)
	if err != nil {
		t.Fatalf("loadIndex: %v", err)
	}
	if _, ok := back.Files["h2.bin"]; ok {
		t.Error("消失的文件不该进索引，否则下次不会重试")
	}
	if len(back.Files) != 3 {
		t.Fatalf("索引里 %d 个文件, want 3", len(back.Files))
	}
}

// 同一次备份的多个包必须合并进同一个 manifest 批次，否则 status 会把
// 22 个包显示成 22 次备份。
func TestManifestMergesSameRunBatches(t *testing.T) {
	out := t.TempDir()
	const tag = "2026-01-01_00-00-00"

	for _, name := range []string{"20260101_000000_000.7z", "20260101_000001_000.7z"} {
		if err := writeManifest(out, "t", tag, batchMeta{
			Packs:  []string{name},
			Hashes: []string{name},
			Hash:   name,
		}); err != nil {
			t.Fatalf("writeManifest %s: %v", name, err)
		}
	}
	m, err := loadManifest(out)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if len(m.Batches) != 1 {
		t.Fatalf("同一次备份应该合并成 1 批，实际 %d", len(m.Batches))
	}
	if len(m.Batches[0].Files) != 2 {
		t.Fatalf("批内包数 = %d, want 2", len(m.Batches[0].Files))
	}
	if m.Batches[0].Hash != batchHashOf([]string{"20260101_000000_000.7z", "20260101_000001_000.7z"}, "") {
		t.Error("合并后的 batch_hash 应该是所有包哈希拼接的 sha256")
	}
}
