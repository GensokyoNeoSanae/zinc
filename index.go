package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// index.tsv 是「文件 → 当前状态」的缓存，权威数据在包本身里。
// 丢失或损坏都能用 zinc reindex 从包目录重建。
//
// 格式（制表符分隔）：
//   # zinc-index 2
//   # P = pack: id, 包文件名, 内容 mtime 起, 内容 mtime 止
//   # F = file: 相对路径, 字节数, mtime(本地 YYYYMMDDThhmmss.fffffff+zzzz), 包编号
//   # end = 包数, 文件数
//   P	1	20261024_110635_081.7z	20261019	20261023
//   F	a/b.png	78388	20261023T124138.9843416+0800	1
//   # end	1	1

const (
	indexMagic  = "# zinc-index 2"
	idxMTimeFmt = "20060102T150405.0000000-0700"
)

// FileEntry 是某个文件最后一次被备份时的状态
type FileEntry struct {
	Size  int64
	MTime time.Time
	Pack  int
}

// PackEntry 是一个备份包的登记
type PackEntry struct {
	ID   int
	Name string
	From string // 内容 mtime 最早日期 YYYYMMDD
	To   string // 内容 mtime 最晚日期 YYYYMMDD
}

// Index 是 <output>/index.tsv
type Index struct {
	Files  map[string]*FileEntry
	Packs  []PackEntry
	nextID int
}

func indexPath(output string) string { return filepath.Join(output, "index.tsv") }

func newIndex() *Index {
	return &Index{Files: map[string]*FileEntry{}, nextID: 1}
}

// addPack 登记一个包并返回它的编号
func (idx *Index) addPack(name, from, to string) int {
	id := idx.nextID
	idx.nextID++
	idx.Packs = append(idx.Packs, PackEntry{ID: id, Name: name, From: from, To: to})
	return id
}

// packName 返回包编号对应的文件名
func (idx *Index) packName(id int) string {
	for i := range idx.Packs {
		if idx.Packs[i].ID == id {
			return idx.Packs[i].Name
		}
	}
	return ""
}

func loadIndex(output string) (*Index, error) {
	p := indexPath(output)
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return newIndex(), nil
	}
	if err != nil {
		return nil, fmt.Errorf(M.LoadIndexFailed, err)
	}
	defer f.Close()

	idx := newIndex()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	lineNo := 0
	sawEnd := false
	wantPacks, wantFiles := -1, -1
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if lineNo == 1 {
			if line != indexMagic {
				return nil, fmt.Errorf(M.IndexBadHeader, p, indexMagic)
			}
			continue
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "# end\t") {
				fields := strings.Split(line, "\t")
				if len(fields) != 3 {
					return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
				}
				a, e1 := strconv.Atoi(fields[1])
				b, e2 := strconv.Atoi(fields[2])
				if e1 != nil || e2 != nil {
					return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
				}
				wantPacks, wantFiles = a, b
				sawEnd = true
			}
			continue
		}

		fields := strings.Split(line, "\t")
		switch fields[0] {
		case "P":
			if len(fields) != 5 {
				return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
			}
			id, e := strconv.Atoi(fields[1])
			if e != nil {
				return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
			}
			idx.Packs = append(idx.Packs, PackEntry{ID: id, Name: fields[2], From: fields[3], To: fields[4]})
			if id >= idx.nextID {
				idx.nextID = id + 1
			}
		case "F":
			// 兼容曾经写过的 6 列（带哈希）索引：读的时候忽略第 6 列，下次保存会写回 5 列。
			if len(fields) != 5 && len(fields) != 6 {
				return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
			}
			size, e1 := strconv.ParseInt(fields[2], 10, 64)
			mt, e2 := parseIdxTime(fields[3])
			pack, e3 := strconv.Atoi(fields[4])
			if e1 != nil || e2 != nil || e3 != nil {
				return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
			}
			idx.Files[fields[1]] = &FileEntry{Size: size, MTime: mt, Pack: pack}
		default:
			return nil, fmt.Errorf(M.IndexLineBad, p, lineNo)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf(M.LoadIndexFailed, err)
	}
	if !sawEnd {
		return nil, fmt.Errorf(M.IndexTruncated, p)
	}
	if wantPacks != len(idx.Packs) || wantFiles != len(idx.Files) {
		return nil, fmt.Errorf(M.IndexCountMismatch, p, wantPacks, wantFiles, len(idx.Packs), len(idx.Files))
	}
	return idx, nil
}

func saveIndex(output string, idx *Index) error {
	var b strings.Builder
	b.WriteString(indexMagic)
	b.WriteString("\n")
	b.WriteString("# P = pack: id, 包文件名, 内容 mtime 起, 内容 mtime 止\n")
	b.WriteString("# F = file: 相对路径, 字节数, mtime(本地 YYYYMMDDThhmmss.fffffff+zzzz), 包编号\n")
	b.WriteString("# end = 包数, 文件数\n")

	for _, p := range idx.Packs {
		fmt.Fprintf(&b, "P\t%d\t%s\t%s\t%s\n", p.ID, p.Name, p.From, p.To)
	}

	rels := make([]string, 0, len(idx.Files))
	for rel := range idx.Files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	written := 0
	for _, rel := range rels {
		f := idx.Files[rel]
		if strings.ContainsAny(rel, "\t\n") {
			continue
		}
		fmt.Fprintf(&b, "F\t%s\t%d\t%s\t%d\n", rel, f.Size, f.MTime.Local().Format(idxMTimeFmt), f.Pack)
		written++
	}

	fmt.Fprintf(&b, "# end\t%d\t%d\n", len(idx.Packs), written)

	p := indexPath(output)
	tmp := p + ".tmp"
	if err := writeFileSync(tmp, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf(M.IndexWriteFailed, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf(M.IndexWriteFailed, err)
	}
	return nil
}

// parseIdxTime 解析定宽时间串 20261023T124138.9843416+0800。
// 手工切片而不是 time.Parse：十万条时差一个数量级。
func parseIdxTime(s string) (time.Time, error) {
	// 20261023T124138.9843416+0800
	if len(s) != 28 || s[8] != 'T' || s[15] != '.' {
		return time.Time{}, fmt.Errorf("bad time %q", s)
	}
	num := func(a, b int) (int, bool) {
		n := 0
		for i := a; i < b; i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return 0, false
			}
			n = n*10 + int(c-'0')
		}
		return n, true
	}
	y, ok1 := num(0, 4)
	mo, ok2 := num(4, 6)
	d, ok3 := num(6, 8)
	h, ok4 := num(9, 11)
	mi, ok5 := num(11, 13)
	se, ok6 := num(13, 15)
	frac, ok7 := num(16, 23)
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7) {
		return time.Time{}, fmt.Errorf("bad time %q", s)
	}
	// 时区偏移 ±hhmm
	sign := 1
	switch s[23] {
	case '+':
	case '-':
		sign = -1
	default:
		return time.Time{}, fmt.Errorf("bad time %q", s)
	}
	oh, ok8 := num(24, 26)
	om, ok9 := num(26, 28)
	if !ok8 || !ok9 {
		return time.Time{}, fmt.Errorf("bad time %q", s)
	}
	off := sign * (oh*3600 + om*60)
	t := time.Date(y, time.Month(mo), d, h, mi, se, frac*100, time.UTC)
	return t.Add(-time.Duration(off) * time.Second), nil
}
