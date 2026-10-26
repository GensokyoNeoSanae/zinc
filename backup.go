package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// fileInfo 是一个待备份文件的快照
type fileInfo struct {
	Rel   string
	Size  int64
	MTime time.Time
}

// ScanResult 是预扫描结果：先扫描再决定要不要密码
type ScanResult struct {
	New      []fileInfo
	Mod      []fileInfo
	Alive    int
	SkippedN int
	Skipped  []string // 跳过路径样本，最多 10 条
	Dur      time.Duration
}

func (s ScanResult) Changed() int { return len(s.New) + len(s.Mod) }

// scanRepo 读取索引并扫描源目录，不修改任何东西
func scanRepo(repo *Repo, set *Settings) (ScanResult, error) {
	idx, err := loadIndex(repo.Output)
	if err != nil {
		return ScanResult{}, err
	}
	skip, err := scanSkipRoot(repo)
	if err != nil {
		return ScanResult{}, err
	}
	return scanSource(repo.Source, idx, repo.exclude(set), skip)
}

// scanSkipRoot 返回需要整棵跳过的输出目录（当它位于源目录内时）。
// 源目录与输出目录相同属于配置错误，直接报错。
func scanSkipRoot(repo *Repo) (string, error) {
	src, err := absClean(repo.Source)
	if err != nil {
		return "", err
	}
	out, err := absClean(repo.Output)
	if err != nil {
		return "", err
	}
	if samePathFold(src, out) {
		return "", fmt.Errorf(M.SourceEqualsOutput, src)
	}
	if withinFold(out, src) {
		return out, nil
	}
	return "", nil
}

// scanSource 遍历源目录，返回新增/修改文件与统计。
// 拒绝访问的目录/文件只跳过并计数，绝不中断整次备份。
func scanSource(source string, idx *Index, excludes []string, skipRoot string) (ScanResult, error) {
	res := ScanResult{}
	start := time.Now()

	if abs, err := filepath.Abs(source); err == nil {
		source = abs
	}

	note := func(p string) {
		res.SkippedN++
		if len(res.Skipped) < 10 {
			res.Skipped = append(res.Skipped, p)
		}
	}

	err := filepath.WalkDir(source, func(full string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			note(full)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if skipRoot != "" && withinFold(full, skipRoot) {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(source, full)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if matchAny(excludes, rel) {
			note(rel)
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			note(rel)
			return nil
		}
		fi := fileInfo{Rel: rel, Size: info.Size(), MTime: info.ModTime()}
		rec := idx.Files[rel]
		if rec == nil {
			res.New = append(res.New, fi)
			return nil
		}
		// 增量判定只看大小和 mtime：任何一项不同都按已修改处理。
		if rec.Size != fi.Size || !rec.MTime.Equal(fi.MTime) {
			res.Mod = append(res.Mod, fi)
			return nil
		}
		res.Alive++
		return nil
	})
	res.Dur = time.Since(start)
	return res, err
}

// matchAny 用 glob 匹配相对路径（斜杠分隔）。支持 ** 跨目录、* 段内、? 单字符。
func matchAny(patterns []string, rel string) bool {
	if len(patterns) == 0 || rel == "" {
		return false
	}
	segs := strings.Split(rel, "/")
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if matchSegs(strings.Split(p, "/"), segs) {
			return true
		}
	}
	return false
}

func matchSegs(pat, seg []string) bool {
	if len(pat) == 0 {
		return len(seg) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(seg); i++ {
			if matchSegs(pat[1:], seg[i:]) {
				return true
			}
		}
		return false
	}
	if len(seg) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], seg[0])
	if err != nil || !ok {
		return false
	}
	return matchSegs(pat[1:], seg[1:])
}

// BackupResult 汇总一次备份的关键指标，供通知和 --all 汇总使用
type BackupResult struct {
	Repo       string
	NewCount   int
	ModCount   int
	Alive      int
	Locked     int
	Skipped    int
	Packs      []string
	TotalSize  int64
	BatchHash  string
	ScanDur    time.Duration
	SkippedAll bool
	Err        error
}

func (r BackupResult) Changed() int { return r.NewCount + r.ModCount }

// packPlan 是一批待打进同一个包的文件
type packPlan struct {
	Files []fileInfo
	Start string // 内容 mtime 最早日期 YYYYMMDD
	End   string // 内容 mtime 最晚日期 YYYYMMDD
}

// packResult 是一个已经落盘的包
type packResult struct {
	Name  string
	Files []fileInfo
	Hash  string
	Size  int64
	Start string
	End   string
}

var packNameRe = regexp.MustCompile(`^\d{8}_\d{6}_\d{3}(_\d+)?\.7z$`)

func isPackName(name string) bool { return packNameRe.MatchString(name) }

// listPackFiles 列出输出目录里的包（字典序）
func listPackFiles(output string) ([]string, error) {
	ents, err := os.ReadDir(output)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() || !isPackName(e.Name()) {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

// packStemLen 是包名主体的定宽长度：YYYYMMDD_hhmmss_SSS
const packStemLen = 19

// packStem 返回包名主体。注意 Go 的格式化布局只有 '.' 后面跟 0 才是小数秒，
// "_000" 会原样输出，所以毫秒必须自己拼。
func packStem(ts time.Time) string {
	return fmt.Sprintf("%s_%03d", ts.Format("20060102_150405"), ts.Nanosecond()/1_000_000)
}

// parsePackTime 从包名解出打包时刻
func parsePackTime(name string) (time.Time, bool) {
	if !isPackName(name) || len(name) < packStemLen {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("20060102_150405", name[:15], time.Local)
	if err != nil {
		return time.Time{}, false
	}
	ms, err := strconv.Atoi(name[16:packStemLen])
	if err != nil {
		return time.Time{}, false
	}
	return t.Add(time.Duration(ms) * time.Millisecond), true
}

// latestPackTime 返回已存在包中最大的打包时刻
func latestPackTime(output string) time.Time {
	names, _ := listPackFiles(output)
	var latest time.Time
	for _, n := range names {
		if t, ok := parsePackTime(n); ok && t.After(latest) {
			latest = t
		}
	}
	return latest
}

// uniquePackName 生成不冲突的包名（同时避开残留的 .part）
func uniquePackName(output string, ts time.Time) string {
	stem := packStem(ts)
	for i := 0; ; i++ {
		name := stem + ".7z"
		if i > 0 {
			name = fmt.Sprintf("%s_%d.7z", stem, i)
		}
		full := filepath.Join(output, name)
		if !fileExists(full) && !fileExists(full+".part") {
			return name
		}
	}
}

// sortByMTime 按 mtime 升序、路径升序排序（保证确定性）
func sortByMTime(files []fileInfo) {
	sort.Slice(files, func(i, j int) bool {
		if !files[i].MTime.Equal(files[j].MTime) {
			return files[i].MTime.Before(files[j].MTime)
		}
		return files[i].Rel < files[j].Rel
	})
}

// planPacks 贪心填包：逐个累加，超过上限就切新包
func planPacks(files []fileInfo, maxBytes int64) []packPlan {
	if maxBytes <= 0 {
		maxBytes = defaultMaxPackByte
	}
	var plans []packPlan
	var cur []fileInfo
	var curSize int64

	flush := func() {
		if len(cur) == 0 {
			return
		}
		plans = append(plans, packPlan{
			Files: cur,
			Start: cur[0].MTime.Format("20060102"),
			End:   cur[len(cur)-1].MTime.Format("20060102"),
		})
		cur = nil
		curSize = 0
	}
	for _, f := range files {
		if len(cur) > 0 && curSize+f.Size > maxBytes {
			flush()
		}
		cur = append(cur, f)
		curSize += f.Size
	}
	flush()
	return plans
}

// packOne 把一个 plan 打成独立的 .7z：先写 .part，自检通过再改名
func packOne(repo *Repo, set *Settings, plan packPlan, name, password string, noVerify bool) (packResult, error) {
	finalPath := filepath.Join(repo.Output, name)
	partPath := finalPath + ".part"
	listfile := filepath.Join(repo.Output, ".list-"+name+".txt")

	rels := make([]string, 0, len(plan.Files))
	for _, f := range plan.Files {
		rels = append(rels, f.Rel)
	}
	if err := os.WriteFile(listfile, []byte(strings.Join(rels, "\n")+"\n"), 0644); err != nil {
		return packResult{}, fmt.Errorf(M.WriteListfileFailed, err)
	}
	defer os.Remove(listfile)

	args := []string{
		"a", "-y", "-t7z",
		"-p" + password,
		"-mhe=on",
		fmt.Sprintf("-mx=%d", repo.compressLevel(set)),
		"-scsUTF-8",
		"-sccUTF-8",
		partPath,
		"@" + listfile,
	}
	cmd := exec.Command(repo.sevenZip(set), args...)
	cmd.Dir = repo.Source // listfile 里的相对路径按源目录解析
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(partPath)
		return packResult{}, fmt.Errorf(M.CompressFailed, err, string(out))
	}

	if !noVerify {
		if err := verifyArchive(repo.sevenZip(set), partPath, password); err != nil {
			os.Remove(partPath)
			return packResult{}, fmt.Errorf(M.VerifyFailed, err)
		}
	}

	if err := os.Rename(partPath, finalPath); err != nil {
		os.Remove(partPath)
		return packResult{}, fmt.Errorf(M.RenamePackFailed, err)
	}

	h, err := sha256File(finalPath)
	if err != nil {
		os.Remove(finalPath)
		return packResult{}, err
	}
	var size int64
	if st, serr := os.Stat(finalPath); serr == nil {
		size = st.Size()
	}
	return packResult{Name: name, Files: plan.Files, Hash: h, Size: size, Start: plan.Start, End: plan.End}, nil
}

// cleanupStaleParts 清掉上次中断留下的 .part。
// 调用前必须已持有仓库锁，否则可能删掉正在写的临时文件。
func cleanupStaleParts(output string) {
	ents, err := os.ReadDir(output)
	if err != nil {
		return
	}
	for _, e := range ents {
		n := e.Name()
		if strings.HasSuffix(n, ".7z.part") && isPackName(strings.TrimSuffix(n, ".part")) {
			os.Remove(filepath.Join(output, n))
		}
	}
}

// packAll 规划并打包全部文件。任一步失败则清掉本次已生成的所有包。
func packAll(repo *Repo, set *Settings, files []fileInfo, password string, noVerify bool) ([]packResult, error) {
	cleanupStaleParts(repo.Output)
	sortByMTime(files)
	limit := repo.maxPackByte(set)
	plans := planPacks(files, limit)

	last := latestPackTime(repo.Output)
	var results []packResult
	cleanup := func() {
		for _, r := range results {
			os.Remove(filepath.Join(repo.Output, r.Name))
		}
	}

	for _, plan := range plans {
		if len(plan.Files) == 1 && plan.Files[0].Size > limit {
			fmt.Printf(M.PackTooBig, plan.Files[0].Rel, humanSize(plan.Files[0].Size), humanSize(limit))
		}
		ts := time.Now()
		if !ts.After(last) {
			ts = last.Add(time.Millisecond)
			fmt.Printf(M.ClockRollback, ts.Format("2006-01-02 15:04:05.000"))
		}
		last = ts
		name := uniquePackName(repo.Output, ts)
		r, err := packOne(repo, set, plan, name, password, noVerify)
		if err != nil {
			cleanup()
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// runBackup 执行单仓库备份。返回结果而非直接退出，方便 --all 汇总。
func runBackup(name string, cfg *Config, set *Settings, scan ScanResult, password, encFrom string, noVerify bool) BackupResult {
	res := BackupResult{
		Repo:     name,
		NewCount: len(scan.New),
		ModCount: len(scan.Mod),
		Alive:    scan.Alive,
		Skipped:  scan.SkippedN,
		ScanDur:  scan.Dur,
	}
	started := time.Now()

	repo, err := cfg.repo(name)
	if err != nil {
		res.Err = err
		return res
	}
	if _, statErr := os.Stat(repo.Source); statErr != nil {
		res.Err = fmt.Errorf(M.SourceNotExist, repo.Source)
		return res
	}
	if err := os.MkdirAll(repo.Output, 0755); err != nil {
		res.Err = fmt.Errorf(M.CreateOutputFailed, err)
		return res
	}

	lock, err := acquireLock(repo.Output, set.lockTimeout())
	if err != nil {
		res.Err = err
		return res
	}
	defer lock.release()

	idx, err := loadIndex(repo.Output)
	if err != nil {
		res.Err = err
		return res
	}

	// 占用预检：被 Excel 之类独占打开的文件先剔除，索引不更新，下次自动重试
	changed := make([]fileInfo, 0, len(scan.New)+len(scan.Mod))
	changed = append(changed, scan.New...)
	changed = append(changed, scan.Mod...)

	ready := make([]fileInfo, 0, len(changed))
	var locked []string
	for _, f := range changed {
		// 扫描发生在拿锁之前：如果等锁期间另一个进程已经备份了同样的版本，这里跳过。
		if rec := idx.Files[f.Rel]; rec != nil && rec.Size == f.Size && rec.MTime.Equal(f.MTime) {
			continue
		}
		full := filepath.Join(repo.Source, filepath.FromSlash(f.Rel))
		fh, oerr := os.Open(full)
		if oerr != nil {
			locked = append(locked, f.Rel)
			continue
		}
		fh.Close()
		ready = append(ready, f)
	}
	res.Locked = len(locked)

	shown := 0
	for _, p := range scan.Skipped {
		if shown >= 3 {
			break
		}
		fmt.Printf(M.ScanSkipItem, p)
		shown++
	}
	if scan.SkippedN > shown {
		fmt.Printf(M.ScanSkipMore, scan.SkippedN-shown)
	}
	shown = 0
	for _, p := range locked {
		if shown >= 3 {
			break
		}
		fmt.Printf(M.LockedItem, p)
		shown++
	}
	if len(locked) > shown {
		fmt.Printf(M.LockedMore, len(locked)-shown)
	}

	if len(ready) == 0 {
		res.SkippedAll = true
		if len(locked) > 0 {
			fmt.Printf(M.AllLocked, name, len(locked))
			logf("backup %s: all %d changed files locked", name, len(locked))
		} else {
			fmt.Printf(M.NoNew, name)
			logf(M.LogNoChange, name)
		}
		return res
	}

	packs, err := packAll(repo, set, ready, password, noVerify)
	if err != nil {
		res.Err = err
		return res
	}

	hashes := make([]string, 0, len(packs))
	names := make([]string, 0, len(packs))
	var total int64
	from, to := "", ""
	for _, p := range packs {
		id := idx.addPack(p.Name, p.Start, p.End)
		for _, f := range p.Files {
			idx.Files[f.Rel] = &FileEntry{Size: f.Size, MTime: f.MTime, Pack: id}
		}
		hashes = append(hashes, p.Hash)
		names = append(names, p.Name)
		total += p.Size
		if from == "" || p.Start < from {
			from = p.Start
		}
		if p.End > to {
			to = p.End
		}
	}

	sum := sha256.Sum256([]byte(strings.Join(hashes, "|")))
	res.BatchHash = fmt.Sprintf("%x", sum)
	res.TotalSize = total
	res.Packs = names

	tag := started.Format("2006-01-02_15-04-05")
	// manifest 先写：失败就把本批次的包清掉，索引不动，整批当作没发生过。
	if err := writeManifest(repo.Output, name, tag, names, hashes, res.BatchHash, from, to); err != nil {
		for _, n := range names {
			os.Remove(filepath.Join(repo.Output, n))
		}
		res.Err = err
		return res
	}

	if err := saveIndex(repo.Output, idx); err != nil {
		res.Err = err
		return res
	}

	fmt.Printf(M.DetailLine+"\n", name, len(ready), humanSize(total), len(packs),
		fmt.Sprintf(M.EncFrom, encFrom), res.ScanDur.Milliseconds(), shortHash(res.BatchHash))
	logf("backup %s: %d files, %d packs, %s, hash=%s", name, len(ready), len(packs), humanSize(total), res.BatchHash)
	return res
}

// verifyArchive 用 7za t 校验包内所有数据块（CRC32）
func verifyArchive(sevenZip, pack, password string) error {
	cmd := exec.Command(sevenZip, "t", "-t7z", "-sccUTF-8", "-p"+password, pack)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("7za t: %v\n%s", err, string(out))
	}
	return nil
}

func sha256File(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func baseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

// Manifest 记录每批次的包哈希，供事后 verify
type Manifest struct {
	Repo    string          `json:"repo"`
	Batches []ManifestBatch `json:"batches"`
}

type ManifestBatch struct {
	Time   string   `json:"time"`
	Hash   string   `json:"batch_hash"`
	Files  []string `json:"files"`
	Hashes []string `json:"hashes"`
	From   string   `json:"from,omitempty"`
	To     string   `json:"to,omitempty"`
}

func manifestPath(output string) string {
	return filepath.Join(output, "manifest.json")
}

// writeManifest 追加一个批次到 manifest.json。
// 先写临时文件再改名；读到损坏的旧 manifest 时把它改名保留并报错，
// 绝不把损坏内容当成空账本继续追加。
func writeManifest(output, repo, tag string, packs, hashes []string, batchHash, from, to string) error {
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	p := manifestPath(output)
	m := Manifest{Repo: repo}
	if b, err := os.ReadFile(p); err == nil {
		if err := json.Unmarshal(b, &m); err != nil {
			bad := p + ".corrupt-" + time.Now().Format("20060102-150405.000")
			if rerr := os.Rename(p, bad); rerr != nil {
				return fmt.Errorf(M.ManifestCorruptLocked, p, err)
			}
			return fmt.Errorf(M.ManifestCorrupt, p, bad, err)
		}
		m.Repo = repo
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(M.ManifestReadFailed, err)
	}
	m.Batches = append(m.Batches, ManifestBatch{
		Time:   tag,
		Hash:   batchHash,
		Files:  baseNames(packs),
		Hashes: hashes,
		From:   from,
		To:     to,
	})
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	tmp := p + ".tmp"
	if err := writeFileSync(tmp, b, 0644); err != nil {
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	return nil
}
