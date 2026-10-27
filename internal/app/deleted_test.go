package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	M = zhMessages
	os.Exit(m.Run())
}

const (
	tStem1 = "20261024_110635_081"
	tStem2 = "20261025_070012_334"
)

func writeEmpty(t *testing.T, p string) {
	t.Helper()
	if err := os.WriteFile(p, nil, 0644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

func TestSafeRel(t *testing.T) {
	ok := []string{
		"a.txt",
		"notes/2026/a.md",
		"deep/a/b/c.png",
		`back\slash\a.txt`, // 反斜杠按路径分隔符处理
	}
	for _, rel := range ok {
		if _, err := safeRel(rel); err != nil {
			t.Errorf("safeRel(%q) unexpected error: %v", rel, err)
		}
	}

	bad := []string{
		"",
		"/abs/path",
		`C:\Windows\notepad.exe`,
		"C:relative",
		"../outside.txt",
		"a/../../outside.txt",
		"a/./b",
		"a//b",
		"file.txt:stream",
		"tab\there.txt",
		`\\server\share\x.txt`,
	}
	for _, rel := range bad {
		if _, err := safeRel(rel); err == nil {
			t.Errorf("safeRel(%q) should have failed", rel)
		}
	}
}

func TestDeletionRecordRoundTrip(t *testing.T) {
	out := t.TempDir()
	rels := []string{"b.txt", "a/z.md", "a/y.md", "b.txt"}
	if err := writeDeletions(out, tStem1+delExt, rels); err != nil {
		t.Fatalf("writeDeletions: %v", err)
	}
	got, err := readDeletions(out, tStem1+delExt)
	if err != nil {
		t.Fatalf("readDeletions: %v", err)
	}
	want := []string{"a/y.md", "a/z.md", "b.txt"}
	if len(got) != len(want) {
		t.Fatalf("got %d paths, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("path %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDeletionRecordRejectsBadHeaderAndPaths(t *testing.T) {
	out := t.TempDir()

	if err := os.WriteFile(delPath(out, tStem1+delExt), []byte("a.txt\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readDeletions(out, tStem1+delExt); err == nil {
		t.Error("missing header should be rejected")
	}

	body := delMagic + "\n../escape.txt\n"
	if err := os.WriteFile(delPath(out, tStem2+delExt), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readDeletions(out, tStem2+delExt); err == nil {
		t.Error("path traversal should be rejected")
	}
}

func TestBuildTimelineOrder(t *testing.T) {
	out := t.TempDir()
	// 同一批次三个包 + 一份记录；再加一个只删文件、没有包的批次
	writeEmpty(t, filepath.Join(out, tStem1+".7z"))
	writeEmpty(t, filepath.Join(out, tStem1+"_1.7z"))
	writeEmpty(t, filepath.Join(out, tStem1+"_2.7z"))
	writeEmpty(t, filepath.Join(out, tStem1+delExt))
	writeEmpty(t, filepath.Join(out, tStem2+delExt))

	entries, err := buildTimeline(out, []string{tStem1 + delExt, tStem2 + delExt})
	if err != nil {
		t.Fatalf("buildTimeline: %v", err)
	}
	want := []string{
		tStem1 + ".7z",
		tStem1 + "_1.7z",
		tStem1 + "_2.7z",
		tStem1 + delExt,
		tStem2 + delExt,
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d", len(entries), len(want))
	}
	for i := range want {
		if entries[i].name() != want[i] {
			t.Errorf("entry %d = %q, want %q", i, entries[i].name(), want[i])
		}
	}

	// 顺序坏掉必须报错，不能硬着头皮恢复：X.7z 与 X_0.7z 撞到同一个排序键
	bad := t.TempDir()
	writeEmpty(t, filepath.Join(bad, tStem1+".7z"))
	writeEmpty(t, filepath.Join(bad, tStem1+"_0.7z"))
	if _, err := buildTimeline(bad, nil); err == nil {
		t.Error("two packs sharing one sort key should be rejected")
	}
}

func TestDelRecordName(t *testing.T) {
	out := t.TempDir()
	writeEmpty(t, filepath.Join(out, tStem1+".7z"))

	// 有包：挂在最后一个包的时间上
	got := delRecordName(out, []packResult{{Name: tStem1 + "_1.7z"}})
	if want := tStem1 + "_1" + delExt; got != want {
		t.Errorf("delRecordName with packs = %q, want %q", got, want)
	}

	// 没包：取一个严格晚于现存条目的时间
	got = delRecordName(out, nil)
	if !isDelName(got) {
		t.Fatalf("bad record name %q", got)
	}
	if got <= tStem1+delExt {
		t.Errorf("record name %q should sort after %q", got, tStem1)
	}
}

func TestApplyDeletions(t *testing.T) {
	dest := t.TempDir()
	mustWrite := func(rel string) {
		full := filepath.Join(dest, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("keep.txt")
	mustWrite("notes/a.md")
	mustWrite("notes/sub/b.md")

	removed, missing, err := applyDeletions(dest, []string{"notes/a.md", "notes/sub/b.md", "never-existed.txt"})
	if err != nil {
		t.Fatalf("applyDeletions: %v", err)
	}
	if removed != 2 || missing != 1 {
		t.Errorf("removed=%d missing=%d, want 2/1", removed, missing)
	}
	if _, err := os.Stat(filepath.Join(dest, "keep.txt")); err != nil {
		t.Errorf("keep.txt should survive: %v", err)
	}

	pruneEmptyDirs(dest, []string{"notes/a.md", "notes/sub/b.md"})
	if _, err := os.Stat(filepath.Join(dest, "notes")); !os.IsNotExist(err) {
		t.Errorf("empty notes/ should be pruned, stat err = %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("destination root must survive: %v", err)
	}

	// 越界路径一个文件都不能动
	mustWrite("keep.txt")
	if _, _, err := applyDeletions(dest, []string{"../outside.txt"}); err == nil {
		t.Error("traversal must be refused")
	}
}

func TestScanSourceReportsDeletions(t *testing.T) {
	src := t.TempDir()
	write := func(rel string) {
		full := filepath.Join(src, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(rel), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt")
	write("sub/b.txt")
	write("skip/keep.tmp")

	idx := newIndex()
	for _, rel := range []string{"a.txt", "sub/b.txt", "gone/c.txt", "skip/keep.tmp", "skip/other.tmp"} {
		idx.Files[rel] = &FileEntry{Size: 1, Pack: 1}
	}

	res, err := scanSource(src, idx, []string{"skip/*.tmp"}, "")
	if err != nil {
		t.Fatalf("scanSource: %v", err)
	}
	want := []string{"gone/c.txt"}
	if len(res.Del) != len(want) || res.Del[0] != want[0] {
		t.Errorf("deletions = %v, want %v", res.Del, want)
	}
	// 大小不一样所以算修改；被排除的文件既不算修改也不算删除
	for _, f := range res.Mod {
		if f.Rel == "skip/keep.tmp" {
			t.Error("excluded file must not be treated as modified")
		}
	}
	if len(res.Mod) != 2 {
		t.Errorf("modified = %d, want 2 (index sizes are all wrong)", len(res.Mod))
	}
	if res.SkippedN != 1 {
		t.Errorf("skipped = %d, want 1 (skip/ dir)", res.SkippedN)
	}
}
