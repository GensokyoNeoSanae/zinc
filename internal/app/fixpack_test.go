package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBadFiles(t *testing.T) {
	out := `ERROR: CRC Failed in encrypted file. Wrong password? : f1.txt
ERROR: Data Error : dir/f2.bin
Scanning the drive for archives:
Everything is Ok
ERROR: CRC Failed in encrypted file. Wrong password? : f1.txt`
	got := parseBadFiles(out)
	want := []string{"f1.txt", "dir/f2.bin"}
	if len(got) != len(want) {
		t.Fatalf("parseBadFiles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseBadFiles[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseBadFilesIgnoresNoise(t *testing.T) {
	for _, out := range []string{"", "Everything is Ok\n", "ERROR: something without a filename"} {
		if got := parseBadFiles(out); len(got) != 0 {
			t.Errorf("parseBadFiles(%q) = %v, want empty", out, got)
		}
	}
}

func TestSrcFileMatches(t *testing.T) {
	src := t.TempDir()
	repo := &Repo{Source: src}
	idx := newIndex()
	rel := "a.txt"
	p := filepath.Join(src, rel)
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}

	// 索引里没有这个文件：存在但不算「一致」
	if exists, same := srcFileMatches(repo, idx, rel); !exists || same {
		t.Errorf("no index entry: exists=%v same=%v, want true/false", exists, same)
	}

	// 登记了且大小、时间都对得上
	idx.Files[rel] = &FileEntry{Size: st.Size(), MTime: st.ModTime()}
	if exists, same := srcFileMatches(repo, idx, rel); !exists || !same {
		t.Errorf("matching entry: exists=%v same=%v, want true/true", exists, same)
	}

	// 大小变了
	idx.Files[rel] = &FileEntry{Size: st.Size() + 1, MTime: st.ModTime()}
	if exists, same := srcFileMatches(repo, idx, rel); !exists || same {
		t.Errorf("size mismatch: exists=%v same=%v, want true/false", exists, same)
	}

	// 文件不在
	idx.Files["gone.txt"] = &FileEntry{Size: 1}
	if exists, _ := srcFileMatches(repo, idx, "gone.txt"); exists {
		t.Error("missing file reported as existing")
	}
}

// 端到端：造一个坏包，fixpack 之后 verify 必须通过
func TestFixpackEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)

	for i := 1; i <= 3; i++ {
		name := filepath.Join(src, fmt.Sprintf("f%d.txt", i))
		if err := os.WriteFile(name, []byte(strings.Repeat("payload", 300)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repo := &Repo{Source: src, Output: out}
	set := noConfirmSettings(sz)
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}

	sc, err := scanRepo(repo, set)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if r := runBackup("t", cfg, set, sc, "PW", "cli", false); r.Err != nil {
		t.Fatalf("backup: %v", r.Err)
	}

	packs, err := sortedPacks(out)
	if err != nil || len(packs) == 0 {
		t.Fatalf("no packs: %v", err)
	}
	packPath := filepath.Join(out, packs[0])
	b, err := os.ReadFile(packPath)
	if err != nil {
		t.Fatal(err)
	}
	b[len(b)/2] ^= 0xFF
	if err := os.WriteFile(packPath, b, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runFixpack("t", cfg, set, "PW", false); err != nil {
		t.Fatalf("fixpack: %v", err)
	}
	// 修完必须能通过哈希校验
	if err := runVerify("t", cfg, set, 100, true, false); err != nil {
		t.Fatalf("verify after fixpack: %v", err)
	}
}
