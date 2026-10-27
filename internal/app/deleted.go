package app

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 删除记录（.rm）与包同名同位，落在输出目录里：
//
//	20261024_110635_081.7z    ← 一个包
//	20261024_110635_081.rm    ← 这一批次里从源目录消失的文件
//
// 恢复时把包和记录合成一条时间线：记录挂在批次最后一个包的时间上，
// 所以它在这个批次的所有包解压完之后执行。包本身永不改写，
// 手工按名字顺序 7z x 仍然能拿到所有曾经备份过的文件。
//
// 记录是纯文本，第二行是路径清单自己的 SHA-256，用来挡位翻转和手改：
//
//	# zinc-deleted 1
//	# sha256 8f14e45fceea167a5a36dedd4bea2543…
//	notes/old-draft.md
//	attachments/2026/screenshot.png

const (
	delMagic     = "# zinc-deleted 1"
	delSumPrefix = "# sha256 "
	delExt       = ".rm"
)

var delNameRe = regexp.MustCompile(`^\d{8}_\d{6}_\d{3}(_\d+)?\.rm$`)

func isDelName(name string) bool { return delNameRe.MatchString(name) }

func delPath(output, name string) string { return filepath.Join(output, name) }

// delChecksum 算路径清单的哈希：一行一个，含行尾换行，顺序就是文件里的顺序。
// 不含表头，所以换行符从 LF 变成 CRLF 不影响校验。
func delChecksum(rels []string) string {
	h := sha256.New()
	for _, rel := range rels {
		h.Write([]byte(rel))
		h.Write([]byte("\n"))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// listDelFiles 列出输出目录里的删除记录（字典序）
func listDelFiles(output string) ([]string, error) {
	ents, err := os.ReadDir(output)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() || !isDelName(e.Name()) {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

// safeRel 校验记录里的相对路径。
// Windows 上 "C:x"、"\\server\share\x"、"a\..\..\b" 都能被 filepath.Join
// 拼到目标目录外面去，所以在动手删之前一律挡下来。
func safeRel(rel string) (string, error) {
	s := strings.ReplaceAll(rel, "\\", "/")
	if s == "" {
		return "", fmt.Errorf("empty path")
	}
	if strings.HasPrefix(s, "/") {
		return "", fmt.Errorf("absolute path")
	}
	if strings.Contains(s, ":") {
		return "", fmt.Errorf("colon in path")
	}
	if strings.ContainsAny(s, "\x00\t\n\r") {
		return "", fmt.Errorf("control character in path")
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", fmt.Errorf("bad segment %q", seg)
		}
	}
	if filepath.VolumeName(filepath.FromSlash(s)) != "" {
		return "", fmt.Errorf("volume path")
	}
	return s, nil
}

// writeDeletions 写一份删除记录：排序去重，先写 .tmp 再改名。
func writeDeletions(output, name string, rels []string) error {
	seen := make(map[string]struct{}, len(rels))
	clean := make([]string, 0, len(rels))
	for _, rel := range rels {
		s, err := safeRel(rel)
		if err != nil {
			return fmt.Errorf(M.DelRecordBadPath, rel, name)
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		clean = append(clean, s)
	}
	sort.Strings(clean)

	var b strings.Builder
	b.WriteString(delMagic)
	b.WriteString("\n")
	b.WriteString(delSumPrefix)
	b.WriteString(delChecksum(clean))
	b.WriteString("\n")
	for _, rel := range clean {
		b.WriteString(rel)
		b.WriteString("\n")
	}

	p := delPath(output, name)
	tmp := p + ".tmp"
	if err := writeFileSync(tmp, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf(M.DelRecordWriteFailed, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return fmt.Errorf(M.DelRecordWriteFailed, err)
	}
	return nil
}

// readDeletions 读一份删除记录。表头不对、校验和对不上、出现非法路径都直接报错：
// 恢复时宁可停下，也不要照着一条看不懂的记录去删文件。
func readDeletions(output, name string) ([]string, error) {
	b, err := os.ReadFile(delPath(output, name))
	if err != nil {
		return nil, fmt.Errorf(M.DelRecordReadFailed, name, err)
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != delMagic {
		return nil, fmt.Errorf(M.DelRecordBadHeader, name, delMagic)
	}
	if len(lines) < 2 || !strings.HasPrefix(lines[1], delSumPrefix) {
		return nil, fmt.Errorf(M.DelRecordNoSum, name)
	}
	want := strings.TrimSpace(strings.TrimPrefix(lines[1], delSumPrefix))

	var rels []string
	for _, line := range lines[2:] {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		s, serr := safeRel(line)
		if serr != nil {
			return nil, fmt.Errorf(M.DelRecordBadPath, line, name)
		}
		rels = append(rels, s)
	}
	if got := delChecksum(rels); got != want {
		return nil, fmt.Errorf(M.DelRecordCorrupt, name, shortHash(want), shortHash(got))
	}
	return rels, nil
}

// loadDelRecords 一次读完所有删除记录。恢复前先把它们全部校验完，
// 这样一条坏记录不会等到删了一半才发现。
func loadDelRecords(output string) (map[string][]string, error) {
	names, err := listDelFiles(output)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(names))
	for _, n := range names {
		rels, rerr := readDeletions(output, n)
		if rerr != nil {
			return nil, rerr
		}
		out[n] = rels
	}
	return out, nil
}

// timeEntry 是恢复时间线上的一项：一个包，或一份删除记录。
type timeEntry struct {
	Pack string
	Del  string
	Stem string
	Seq  int
}

func (e timeEntry) name() string {
	if e.Pack != "" {
		return e.Pack
	}
	return e.Del
}

// maxSeq 让不带 _N 后缀的删除记录排在同一个时间戳的所有包之后。
const maxSeq = int(^uint(0) >> 1)

// buildTimeline 把包和删除记录合成一条按时间升序的执行序列。
// 排序依据与包名一致（定宽前缀 + 数字序号）；同一个前缀下先所有包、后删除记录。
func buildTimeline(output string, dels []string) ([]timeEntry, error) {
	packs, err := listPackFiles(output)
	if err != nil {
		return nil, err
	}
	entries := make([]timeEntry, 0, len(packs)+len(dels))
	for _, p := range packs {
		stem, seq := packSortKey(p)
		entries = append(entries, timeEntry{Pack: p, Stem: stem, Seq: seq})
	}
	for _, d := range dels {
		stem := strings.TrimSuffix(d, delExt)
		if len(stem) > packStemLen && stem[packStemLen] == '_' {
			if n, e := strconv.Atoi(stem[packStemLen+1:]); e == nil {
				entries = append(entries, timeEntry{Del: d, Stem: stem[:packStemLen], Seq: n})
				continue
			}
		}
		entries = append(entries, timeEntry{Del: d, Stem: stem, Seq: maxSeq})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Stem != b.Stem {
			return a.Stem < b.Stem
		}
		if a.Seq != b.Seq {
			return a.Seq < b.Seq
		}
		// 同一时刻：包先解压，删除记录后执行
		return a.Pack != "" && b.Pack == ""
	})

	// 包的顺序必须是严格递增的，这是恢复正确性的唯一依据。
	lastName := ""
	lastStem := ""
	lastSeq := 0
	for _, e := range entries {
		if e.Pack == "" {
			continue
		}
		if lastName != "" && (e.Stem < lastStem || (e.Stem == lastStem && e.Seq <= lastSeq)) {
			return nil, fmt.Errorf(M.PackOrderBroken, lastName, e.Pack)
		}
		lastName, lastStem, lastSeq = e.Pack, e.Stem, e.Seq
	}
	return entries, nil
}

// applyDeletions 在 dest 里删掉记录中的路径，返回实际删除数与本就不存在的数量。
// 先把所有路径校验一遍再动手：一条记录要么全做，要么一个文件都不动。
func applyDeletions(dest string, rels []string) (removed, missing int, err error) {
	full := make([]string, 0, len(rels))
	for _, rel := range rels {
		s, serr := safeRel(rel)
		if serr != nil {
			return 0, 0, fmt.Errorf(M.DelRecordBadPath, rel, dest)
		}
		f := filepath.Join(dest, filepath.FromSlash(s))
		if !withinFold(f, dest) || samePathFold(f, dest) {
			return 0, 0, fmt.Errorf(M.DelRecordBadPath, rel, dest)
		}
		full = append(full, f)
	}
	for _, f := range full {
		st, statErr := os.Lstat(f)
		if os.IsNotExist(statErr) {
			missing++
			continue
		}
		if statErr != nil {
			return removed, missing, fmt.Errorf(M.DelRemoveFailed, f, statErr)
		}
		if st.IsDir() {
			// 记录里只有文件；空目录由 pruneEmptyDirs 收尾
			missing++
			continue
		}
		if rmErr := os.Remove(f); rmErr != nil {
			if os.IsNotExist(rmErr) {
				missing++
				continue
			}
			return removed, missing, fmt.Errorf(M.DelRemoveFailed, f, rmErr)
		}
		removed++
	}
	return removed, missing, nil
}

// pruneEmptyDirs 删掉被删空的目录：从最深的开始，非空自然失败，止于 dest 本身。
func pruneEmptyDirs(dest string, rels []string) {
	set := make(map[string]struct{}, len(rels))
	for _, rel := range rels {
		d := path.Dir(rel)
		for d != "." && d != "/" && d != "" {
			set[d] = struct{}{}
			d = path.Dir(d)
		}
	}
	dirs := make([]string, 0, len(set))
	for d := range set {
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool {
		if len(dirs[i]) != len(dirs[j]) {
			return len(dirs[i]) > len(dirs[j])
		}
		return dirs[i] > dirs[j]
	})
	for _, d := range dirs {
		f := filepath.Join(dest, filepath.FromSlash(d))
		if samePathFold(f, dest) || !withinFold(f, dest) {
			continue
		}
		os.Remove(f) // 非空就失败，正好
	}
}

// delRecordName 决定本批次删除记录的文件名：有包就挂到最后一个包的时间上，
// 保证它在这个批次的所有包之后执行；只删了文件、没打新包时另取一个更晚的时间。
func delRecordName(output string, packs []packResult) string {
	if len(packs) > 0 {
		return strings.TrimSuffix(packs[len(packs)-1].Name, ".7z") + delExt
	}
	return uniqueDelName(output, time.Now())
}

// uniqueDelName 生成不与现存条目冲突的记录名。
// 时间不够晚就往后挪 1 毫秒——定宽时间戳，所以字典序就是时间序。
func uniqueDelName(output string, ts time.Time) string {
	if last := latestEntryTime(output); !ts.After(last) {
		ts = last.Add(time.Millisecond)
	}
	for {
		stem := packStem(ts)
		if !fileExists(delPath(output, stem+delExt)) && !fileExists(delPath(output, stem+".7z")) {
			return stem + delExt
		}
		ts = ts.Add(time.Millisecond)
	}
}

// latestEntryTime 取输出目录里包和删除记录中最大的时间
func latestEntryTime(output string) time.Time {
	names, _ := listPackFiles(output)
	if dels, err := listDelFiles(output); err == nil {
		names = append(names, dels...)
	}
	var latest time.Time
	for _, n := range names {
		stem := strings.TrimSuffix(strings.TrimSuffix(n, ".7z"), delExt)
		if t, ok := parsePackTime(stem + ".7z"); ok && t.After(latest) {
			latest = t
		}
	}
	return latest
}
