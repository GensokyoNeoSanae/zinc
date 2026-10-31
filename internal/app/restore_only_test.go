package app

import (
	"os"
	"path/filepath"
	"testing"
)

// 局部恢复也要走删除记录：删掉的文件在恢复结果里不该复活。
func TestRestoreOnlyAppliesDeletions(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	a := filepath.Join(src, "A")
	mustMkdir(t, a)
	mustMkdir(t, out)
	if err := os.WriteFile(filepath.Join(a, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz, DefaultPassword: "PW"}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}

	sc1, err := scanRepo(repo, set)
	if err != nil {
		t.Fatalf("scan1: %v", err)
	}
	r1 := runBackup("t", cfg, set, sc1, "PW", "settings", false)
	if r1.Err != nil {
		t.Fatalf("backup1: %v", r1.Err)
	}

	// 删 b、加 c：第二次备份产生新包 + 删除记录
	if err := os.Remove(filepath.Join(a, "b.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "c.txt"), []byte("c"), 0o644); err != nil {
		t.Fatal(err)
	}
	sc2, err := scanRepo(repo, set)
	if err != nil {
		t.Fatalf("scan2: %v", err)
	}
	r2 := runBackup("t", cfg, set, sc2, "PW", "settings", false)
	if r2.Err != nil {
		t.Fatalf("backup2: %v", r2.Err)
	}
	if r2.DelRecord == "" {
		t.Fatal("第二次备份应该留下删除记录")
	}

	dest := filepath.Join(dir, "restore")
	if err := runRestoreOnly("t", cfg, set, dest, "PW", "A", "", true, false); err != nil {
		t.Fatalf("restore-only: %v", err)
	}
	if !fileExists(filepath.Join(dest, "A", "a.txt")) || !fileExists(filepath.Join(dest, "A", "c.txt")) {
		t.Error("恢复结果应该包含 a.txt 和 c.txt")
	}
	if fileExists(filepath.Join(dest, "A", "b.txt")) {
		t.Error("b.txt 已被删除记录删掉，不该在局部恢复里复活")
	}

	// --from 指定旧包时不套删除记录：应该能拿回 b.txt
	destOld := filepath.Join(dir, "restore-old")
	if err := runRestoreOnly("t", cfg, set, destOld, "PW", "A", r1.Packs[0], true, false); err != nil {
		t.Fatalf("restore-only --from: %v", err)
	}
	if !fileExists(filepath.Join(destOld, "A", "b.txt")) {
		t.Error("--from 旧包应该拿到 b.txt")
	}
}

// 没有删除记录时，局部恢复的行为和以前一致。
func TestRestoreOnlyNoDeletions(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, filepath.Join(src, "A"))
	mustMkdir(t, out)
	if err := os.WriteFile(filepath.Join(src, "A", "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz, DefaultPassword: "PW"}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}
	sc, err := scanRepo(repo, set)
	if err != nil {
		t.Fatal(err)
	}
	if r := runBackup("t", cfg, set, sc, "PW", "settings", false); r.Err != nil {
		t.Fatalf("backup: %v", r.Err)
	}
	dest := filepath.Join(dir, "restore")
	if err := runRestoreOnly("t", cfg, set, dest, "PW", "A", "", true, false); err != nil {
		t.Fatalf("restore-only: %v", err)
	}
	if !fileExists(filepath.Join(dest, "A", "x.txt")) {
		t.Error("x.txt 应该被恢复")
	}
}
