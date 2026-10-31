package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeHashFile(t *testing.T, dir, rel, body string) string {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestHashCompareFiles(t *testing.T) {
	dir := t.TempDir()
	a := writeHashFile(t, dir, "a.txt", "hello")
	b := writeHashFile(t, dir, "b.txt", "hello")
	same, err := hashComparePaths(a, b)
	if err != nil || !same {
		t.Fatalf("相同内容应一致: same=%v err=%v", same, err)
	}
	if err := os.WriteFile(b, []byte("hello!"), 0o644); err != nil {
		t.Fatal(err)
	}
	same, err = hashComparePaths(a, b)
	if err != nil || same {
		t.Fatalf("内容不同应报不一致: same=%v err=%v", same, err)
	}
}

func TestHashCompareTrees(t *testing.T) {
	dir := t.TempDir()
	left := filepath.Join(dir, "left")
	right := filepath.Join(dir, "right")
	writeHashFile(t, left, "x/one.txt", "1")
	writeHashFile(t, left, "two.txt", "2")
	writeHashFile(t, right, "x/one.txt", "1")
	writeHashFile(t, right, "two.txt", "2")

	same, err := hashComparePaths(left, right)
	if err != nil || !same {
		t.Fatalf("相同目录树应一致: %v %v", same, err)
	}

	// 改一个、缺一个、多一个
	if err := os.WriteFile(filepath.Join(right, "two.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(right, "x", "one.txt")); err != nil {
		t.Fatal(err)
	}
	writeHashFile(t, right, "extra.txt", "e")

	same, err = hashComparePaths(left, right)
	if err != nil {
		t.Fatal(err)
	}
	if same {
		t.Fatal("有差异时不该返回一致")
	}
	d := compareHashTrees(mustHashTree(t, left), mustHashTree(t, right))
	if len(d.changed) != 1 || d.changed[0] != "two.txt" {
		t.Errorf("changed = %v", d.changed)
	}
	if len(d.missing) != 1 || d.missing[0] != "x/one.txt" {
		t.Errorf("missing = %v", d.missing)
	}
	if len(d.extra) != 1 || d.extra[0] != "extra.txt" {
		t.Errorf("extra = %v", d.extra)
	}
}

func mustHashTree(t *testing.T, root string) map[string]string {
	t.Helper()
	m, err := hashTree(root)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestHashFileMatches(t *testing.T) {
	dir := t.TempDir()
	p := writeHashFile(t, dir, "f.txt", "abc")
	actual, err := fileSHA256(p)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := hashFileMatches(p, strings.ToUpper(actual))
	if err != nil || !ok {
		t.Fatalf("哈希应一致（大小写不敏感）: %v %v", ok, err)
	}
	ok, err = hashFileMatches(p, strings.Repeat("0", 64))
	if err != nil || ok {
		t.Fatalf("错误哈希应不一致: %v %v", ok, err)
	}
	if _, err := hashFileMatches(dir, actual); err == nil {
		t.Fatal("目录不应走 --hash-value 模式")
	}
}
