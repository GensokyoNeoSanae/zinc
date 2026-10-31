package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// v1 账本（files/hashes + rm/rm_hash）读进来要升成 v2，下一次写入落成 packs/rms。
func TestManifestV1Migration(t *testing.T) {
	out := t.TempDir()
	v1 := `{
  "repo": "t",
  "batches": [
    {
      "time": "2026-01-01_00-00-00",
      "op": "compact",
      "batch_hash": "ignored",
      "files": ["20260101_000000_000.7z"],
      "hashes": ["aaa"],
      "from": "20260101",
      "to": "20260101",
      "rm": "20260101_000000_000.rm",
      "rm_hash": "bbb"
    }
  ]
}`
	if err := os.WriteFile(manifestPath(out), []byte(v1), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(out)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if m.Version != manifestVersion {
		t.Errorf("version = %d, want %d", m.Version, manifestVersion)
	}
	if got := m.Batches[0].Packs["20260101_000000_000.7z"]; got != "aaa" {
		t.Errorf("包哈希迁移 = %q", got)
	}
	if got := m.Batches[0].Rms["20260101_000000_000.rm"]; got != "bbb" {
		t.Errorf("删除记录哈希迁移 = %q", got)
	}

	// 再写一个批次，整份账本应该升级成 v2 的写法
	if err := writeManifest(out, "t", "2026-01-02_00-00-00", batchMeta{
		Packs:  []string{"20260102_000000_000.7z"},
		Hashes: []string{"ccc"},
	}); err != nil {
		t.Fatalf("writeManifest: %v", err)
	}
	raw, err := os.ReadFile(manifestPath(out))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"files"`) {
		t.Errorf("写回后不该再有 v1 的 files 字段:\n%s", raw)
	}
	if !strings.Contains(string(raw), `"packs"`) {
		t.Errorf("写回后应该是 v2 的 packs 映射:\n%s", raw)
	}
	// 升级前必须留下一份旧格式原样副本，降级/迁移出错时能回去
	bak, err := os.ReadFile(manifestPath(out) + ".v1.bak")
	if err != nil {
		t.Fatalf("升级前应该留下 .v1.bak: %v", err)
	}
	if !strings.Contains(string(bak), `"files"`) {
		t.Errorf(".v1.bak 应该是旧格式原样:\n%s", bak)
	}
}

// 对账：删掉盘上已消失的幽灵包、补登孤儿包和孤儿删除记录，`--yes` 前不写文件。
func TestRemanifestReconcile(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	files := writeFiles(t, src, "f", 2, 128)
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz, DefaultPassword: "PACKPW"}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}

	packA := "20260101_000000_000.7z"
	packB := "20260102_000000_000.7z"
	rA, err := packOne(repo, set, packPlan{Files: files[:1]}, packA, "PACKPW", false)
	if err != nil {
		t.Fatalf("packOne A: %v", err)
	}
	rB, err := packOne(repo, set, packPlan{Files: files[1:]}, packB, "PACKPW", false)
	if err != nil {
		t.Fatalf("packOne B: %v", err)
	}
	// 账本里只有 A，外加一个盘上不存在的幽灵包
	if err := writeManifest(out, "t", "2026-01-01_00-00-00", batchMeta{
		Packs: []string{packA}, Hashes: []string{rA.Hash},
	}); err != nil {
		t.Fatal(err)
	}
	if err := writeManifest(out, "t", "2026-01-03_00-00-00", batchMeta{
		Packs: []string{"20990101_000000_000.7z"}, Hashes: []string{"deadbeef"},
	}); err != nil {
		t.Fatal(err)
	}
	// 孤儿删除记录，与 B 同名
	rmName := strings.TrimSuffix(packB, ".7z") + ".rm"
	if err := writeDeletions(out, rmName, []string{"gone.txt"}); err != nil {
		t.Fatalf("writeDeletions: %v", err)
	}

	before, err := os.ReadFile(manifestPath(out))
	if err != nil {
		t.Fatal(err)
	}
	// 预演：只报告，不落盘
	if err := runRemanifest("t", cfg, set, "PACKPW", false); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	after, err := os.ReadFile(manifestPath(out))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("预演不该写 manifest")
	}

	// 执行
	if err := runRemanifest("t", cfg, set, "PACKPW", true); err != nil {
		t.Fatalf("remanifest: %v", err)
	}
	m, err := loadManifest(out)
	if err != nil {
		t.Fatal(err)
	}
	packHashes := map[string]string{}
	rmHashes := map[string]string{}
	for _, b := range m.Batches {
		for p, h := range b.Packs {
			packHashes[p] = h
		}
		for rm, h := range b.Rms {
			rmHashes[rm] = h
		}
	}
	if _, ok := packHashes["20990101_000000_000.7z"]; ok {
		t.Error("幽灵包应该被删掉")
	}
	if packHashes[packA] != rA.Hash || packHashes[packB] != rB.Hash {
		t.Errorf("补登后的包哈希不对: %v", packHashes)
	}
	if _, ok := rmHashes[rmName]; !ok {
		t.Errorf("孤儿删除记录 should be added: %v", rmHashes)
	}
	// 批次按时间升序
	for i := 1; i < len(m.Batches); i++ {
		if m.Batches[i-1].Time > m.Batches[i].Time {
			t.Errorf("批次时间没有升序: %s > %s", m.Batches[i-1].Time, m.Batches[i].Time)
		}
	}
	// 对账后 verify 必须干净
	if err := runVerify("t", cfg, set, 100, true, true); err != nil {
		t.Fatalf("对账后 verify 失败: %v", err)
	}
}

// manifest 丢了也能从盘上重建（一包一批），但「哪个包曾经在」的记录就没了。
func TestRemanifestRebuildsMissingManifest(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	// out 必须存在；manifest 故意不建，让 remanifest 走「按磁盘重建」的分支
	mustMkdir(t, out)

	files := writeFiles(t, src, "f", 1, 64)
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz, DefaultPassword: "PACKPW"}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}
	pack := "20260101_000000_000.7z"
	r, err := packOne(repo, set, packPlan{Files: files}, pack, "PACKPW", false)
	if err != nil {
		t.Fatalf("packOne: %v", err)
	}
	if err := runRemanifest("t", cfg, set, "PACKPW", true); err != nil {
		t.Fatalf("remanifest: %v", err)
	}
	m, err := loadManifest(out)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if got := m.Batches[0].Packs[pack]; got != r.Hash {
		t.Errorf("重建后的包哈希 = %q, want %q", got, r.Hash)
	}
	if err := runVerify("t", cfg, set, 100, true, true); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

// 盘上只剩占位文件的包（compact 腾空的）不该被补登成一批。
func TestRemanifestSkipsPlaceholderPack(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	ph := filepath.Join(src, placeholderName)
	if err := os.WriteFile(ph, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(ph)
	if err != nil {
		t.Fatal(err)
	}
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz, DefaultPassword: "PACKPW"}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}
	pack := "20260101_000000_000.7z"
	if _, err := packOne(repo, set, packPlan{Files: []fileInfo{{
		Rel: placeholderName, Size: st.Size(), MTime: st.ModTime(),
	}}}, pack, "PACKPW", false); err != nil {
		t.Fatalf("packOne: %v", err)
	}
	if err := runRemanifest("t", cfg, set, "PACKPW", true); err != nil {
		t.Fatalf("remanifest: %v", err)
	}
	// 唯一要补的是占位包，跳过后没有改动，账本根本不会生成
	if _, err := os.Stat(manifestPath(out)); err == nil {
		m, lerr := loadManifest(out)
		if lerr != nil {
			t.Fatalf("loadManifest: %v", lerr)
		}
		for _, b := range m.Batches {
			if _, ok := b.Packs[pack]; ok {
				t.Fatalf("占位包不该被补登: %v", b.Packs)
			}
		}
	}
}
