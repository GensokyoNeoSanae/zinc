package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// packEntry 是 7za l -slt 解析出来的一个归档内条目
type packEntry struct {
	Path  string
	Size  int64
	MTime time.Time
	IsDir bool
}

// listPackContents 只读归档头（不解压内容），列出包内文件
func listPackContents(sevenZip, pack, password string) ([]packEntry, error) {
	cmd := exec.Command(sevenZip, "l", "-slt", "-sccUTF-8", "-p"+password, pack)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	var entries []packEntry
	var cur packEntry
	inList := false

	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if !inList {
			if strings.HasPrefix(line, "----------") {
				inList = true
			}
			continue
		}
		if line == "" {
			if cur.Path != "" {
				entries = append(entries, cur)
				cur = packEntry{}
			}
			continue
		}
		k, v, ok := strings.Cut(line, " = ")
		if !ok {
			continue
		}
		switch k {
		case "Path":
			cur.Path = v
		case "Size":
			cur.Size, _ = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		case "Modified":
			if t, terr := parse7zTime(strings.TrimSpace(v)); terr == nil {
				cur.MTime = t
			}
		case "Attributes":
			if strings.HasPrefix(strings.TrimSpace(v), "D") {
				cur.IsDir = true
			}
		}
	}
	if cur.Path != "" {
		entries = append(entries, cur)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// parse7zTime 解析 7za 输出的本地时间，例如 2026-10-23 12:41:38.9843416
func parse7zTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05.9999999",
		"2006-01-02 15:04:05",
	}
	var lastErr error
	for _, l := range layouts {
		t, err := time.ParseInLocation(l, s, time.Local)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// runReindex 从包本身重建 index.tsv。索引只是缓存，包才是权威数据。
func runReindex(name string, cfg *Config, set *Settings, password, packStart, packEnd string, yes bool) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(repo.Output); statErr != nil {
		return fmt.Errorf(M.CreateOutputFailed, statErr)
	}

	lock, err := acquireLock(repo.Output, set.lockTimeout())
	if err != nil {
		return err
	}
	defer lock.release()

	packs, err := sortedPacks(repo.Output)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		return fmt.Errorf(M.ReindexNoPacks, repo.Output)
	}
	if packStart != "" || packEnd != "" {
		var keep []string
		for _, p := range packs {
			if packStart != "" && p < packStart {
				continue
			}
			if packEnd != "" && p > packEnd {
				continue
			}
			keep = append(keep, p)
		}
		packs = keep
		if len(packs) == 0 {
			return fmt.Errorf(M.ReindexNoPacks, repo.Output)
		}
	}

	p := indexPath(repo.Output)
	if fileExists(p) {
		if !yes {
			return fmt.Errorf(M.ReindexExists, p)
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return fmt.Errorf(M.ReindexBakFailed, rerr)
		}
		if werr := writeFileSync(p+".bak", b, 0644); werr != nil {
			return fmt.Errorf(M.ReindexBakFailed, werr)
		}
	}

	fmt.Printf(M.ReindexStart, len(packs))
	idx := newIndex()
	for _, pk := range packs {
		entries, lerr := listPackContents(repo.sevenZip(set), filepath.Join(repo.Output, pk), password)
		if lerr != nil {
			return fmt.Errorf(M.ReindexListFailed, pk, lerr)
		}

		from, to := "", ""
		var minT, maxT time.Time
		for _, e := range entries {
			if e.IsDir || e.Path == "" {
				continue
			}
			if minT.IsZero() || e.MTime.Before(minT) {
				minT = e.MTime
			}
			if maxT.IsZero() || e.MTime.After(maxT) {
				maxT = e.MTime
			}
		}
		if !minT.IsZero() {
			from = minT.Format("20060102")
			to = maxT.Format("20060102")
		}

		id := idx.addPack(pk, from, to)
		n := 0
		for _, e := range entries {
			if e.IsDir || e.Path == "" {
				continue
			}
			idx.Files[filepath.ToSlash(e.Path)] = &FileEntry{Size: e.Size, MTime: e.MTime, Pack: id}
			n++
		}
		fmt.Printf(M.ReindexPack, pk, n)
	}

	if err := saveIndex(repo.Output, idx); err != nil {
		return err
	}
	fmt.Printf(M.ReindexDone, len(idx.Packs), len(idx.Files), indexPath(repo.Output))
	logf("reindex %s: %d packs, %d files", name, len(idx.Packs), len(idx.Files))
	return nil
}
