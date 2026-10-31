package app

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fileSHA256 返回文件的 SHA-256（小写十六进制）。
func fileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashTree 递归计算目录下每个文件的 SHA-256，键是斜杠分隔的相对路径。
func hashTree(root string) (map[string]string, error) {
	m := make(map[string]string)
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		h, herr := fileSHA256(p)
		if herr != nil {
			return herr
		}
		m[filepath.ToSlash(rel)] = h
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// hashDiff 是两个目录树的差异。
type hashDiff struct {
	same    int
	changed []string
	missing []string // 只在左边
	extra   []string // 只在右边
}

func compareHashTrees(left, right map[string]string) hashDiff {
	var d hashDiff
	for rel, lh := range left {
		rh, ok := right[rel]
		switch {
		case !ok:
			d.missing = append(d.missing, rel)
		case rh != lh:
			d.changed = append(d.changed, rel)
		default:
			d.same++
		}
	}
	for rel := range right {
		if _, ok := left[rel]; !ok {
			d.extra = append(d.extra, rel)
		}
	}
	sort.Strings(d.changed)
	sort.Strings(d.missing)
	sort.Strings(d.extra)
	return d
}

// hashComparePaths 比较两个文件或两个目录，打印结果并返回是否一致。
func hashComparePaths(a, b string) (bool, error) {
	sa, err := os.Stat(a)
	if err != nil {
		return false, fmt.Errorf(M.PathInvalid, a, err)
	}
	sb, err := os.Stat(b)
	if err != nil {
		return false, fmt.Errorf(M.PathInvalid, b, err)
	}
	if sa.IsDir() != sb.IsDir() {
		return false, errors.New(M.HashCompareType)
	}

	if !sa.IsDir() {
		ha, err := fileSHA256(a)
		if err != nil {
			return false, err
		}
		hb, err := fileSHA256(b)
		if err != nil {
			return false, err
		}
		if ha == hb {
			fmt.Printf(M.HashCompareFileSame, a)
			return true, nil
		}
		fmt.Printf(M.HashCompareFileDiff, a, ha, hb)
		return false, nil
	}

	ta, err := hashTree(a)
	if err != nil {
		return false, err
	}
	tb, err := hashTree(b)
	if err != nil {
		return false, err
	}
	d := compareHashTrees(ta, tb)
	for _, rel := range d.changed {
		fmt.Printf(M.HashCompareDiff, rel)
	}
	for _, rel := range d.missing {
		fmt.Printf(M.MissingFile, rel)
	}
	for _, rel := range d.extra {
		fmt.Printf(M.HashCompareExtra, rel)
	}
	if len(d.changed)+len(d.missing)+len(d.extra) == 0 {
		fmt.Printf(M.HashCompareSame, d.same)
		return true, nil
	}
	fmt.Printf(M.HashCompareSummary, len(d.changed), len(d.missing), len(d.extra))
	return false, nil
}

// hashFileMatches 计算单个文件的 SHA-256 并与期望值比较。
func hashFileMatches(path, want string) (bool, error) {
	st, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf(M.PathInvalid, path, err)
	}
	if st.IsDir() {
		return false, errors.New(M.HashCompareNeedFile)
	}
	actual, err := fileSHA256(path)
	if err != nil {
		return false, err
	}
	w := strings.ToLower(strings.TrimSpace(want))
	w = strings.TrimPrefix(w, "sha256:")
	w = strings.ReplaceAll(w, " ", "")
	if actual == w {
		fmt.Printf(M.HashCompareFileSame, path)
		return true, nil
	}
	fmt.Printf(M.HashCompareFileDiff, path, w, actual)
	return false, nil
}

func cmdHashCompare(args []string) {
	fs := flag.NewFlagSet("hash-compare", flag.ExitOnError)
	s1 := fs.String("source1", "", M.FlagHashSource1)
	s2 := fs.String("source2", "", M.FlagHashSource2)
	src := fs.String("source", "", M.FlagHashSource)
	want := fs.String("hash-value", "", M.FlagHashValue)
	fs.Parse(reorderArgs(args))
	requireArgs(fs, 0, 0, M.UsageHashCompare)

	switch {
	case *s1 != "" || *s2 != "":
		if *s1 == "" || *s2 == "" || *src != "" || *want != "" {
			die(M.HashCompareMode)
		}
		same, err := hashComparePaths(*s1, *s2)
		if err != nil {
			die("%v", err)
		}
		if !same {
			os.Exit(1)
		}
	case *src != "" || *want != "":
		if *src == "" || *want == "" {
			die(M.HashCompareMode)
		}
		same, err := hashFileMatches(*src, *want)
		if err != nil {
			die("%v", err)
		}
		if !same {
			os.Exit(1)
		}
	default:
		die(M.HashCompareMode)
	}
}
