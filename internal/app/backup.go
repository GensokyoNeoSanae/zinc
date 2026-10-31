package app

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
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
	"unicode/utf16"
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
	Del      []string // 自上次备份以来从源目录消失的文件（相对路径）
	Alive    int
	SkippedN int
	Skipped  []string // 跳过路径样本，最多 10 条
	Dur      time.Duration
}

func (s ScanResult) Changed() int { return len(s.New) + len(s.Mod) + len(s.Del) }

// Packable 是这次要真正打包的文件数（不含删除）
func (s ScanResult) Packable() int { return len(s.New) + len(s.Mod) }

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

// scanSource 遍历源目录，返回新增/修改文件、消失的文件与统计。
// 拒绝访问的目录/文件只跳过并计数，绝不中断整次备份。
func scanSource(source string, idx *Index, excludes []string, skipRoot string) (ScanResult, error) {
	res := ScanResult{}
	start := time.Now()

	if abs, err := filepath.Abs(source); err == nil {
		source = abs
	}

	// 索引里的路径先全算“待定消失”，走到的当场划掉，剩下的才可能是删除。
	pending := make(map[string]struct{}, len(idx.Files))
	for rel := range idx.Files {
		pending[rel] = struct{}{}
	}
	// 内容无法确认的目录：拒绝访问、被排除、整棵跳过的输出目录。
	// 它们下面的索引条目一律不算删除——一次权限问题不该让恢复去删真文件。
	var unknown []string

	note := func(p string) {
		res.SkippedN++
		if len(res.Skipped) < 10 {
			res.Skipped = append(res.Skipped, p)
		}
	}

	err := filepath.WalkDir(source, func(full string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			note(full)
			if rel, relErr := filepath.Rel(source, full); relErr == nil && rel != "." {
				unknown = append(unknown, filepath.ToSlash(rel))
			}
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
		if skipRoot != "" && withinFold(full, skipRoot) {
			unknown = append(unknown, rel)
			if d.IsDir() {
				return filepath.SkipDir
			}
			delete(pending, rel)
			return nil
		}
		if matchAny(excludes, rel) {
			note(rel)
			if d.IsDir() {
				unknown = append(unknown, rel)
				return filepath.SkipDir
			}
			delete(pending, rel)
			return nil
		}
		delete(pending, rel)
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
	if err != nil {
		return res, err
	}

	var gone []string
	for rel := range pending {
		if underAny(rel, unknown) || matchAny(excludes, rel) {
			continue
		}
		gone = append(gone, rel)
	}
	sort.Strings(gone)
	res.Del = gone
	return res, nil
}

// underAny 判断 rel 是否等于这些相对目录之一，或位于其下
func underAny(rel string, dirs []string) bool {
	for _, d := range dirs {
		if rel == d || strings.HasPrefix(rel, d+"/") {
			return true
		}
	}
	return false
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
	DelCount   int
	Alive      int
	Locked     int
	Skipped    int
	Packs      []string
	TotalSize  int64
	BatchHash  string
	DelRecord  string
	ScanDur    time.Duration
	SkippedAll bool
	Err        error
}

func (r BackupResult) Changed() int { return r.NewCount + r.ModCount + r.DelCount }

// packPlan 是一批待打进同一个包的文件
type packPlan struct {
	Files []fileInfo
	Start string // 内容 mtime 最早日期 YYYYMMDD
	End   string // 内容 mtime 最晚日期 YYYYMMDD
}

// packResult 是一个已经落盘的包
type packResult struct {
	Name    string
	Files   []fileInfo // 计划打进这个包的文件
	Skipped []string   // 其中实际没能进去的（打包时读不到），不写索引，下次自动重试
	Hash    string
	Size    int64
	Start   string
	End     string
}

// Kept 返回真正进了包的文件（计划清单减去被跳过的）
func (r packResult) Kept() []fileInfo {
	if len(r.Skipped) == 0 {
		return r.Files
	}
	bad := make(map[string]struct{}, len(r.Skipped))
	for _, s := range r.Skipped {
		bad[s] = struct{}{}
	}
	out := make([]fileInfo, 0, len(r.Files))
	for _, f := range r.Files {
		if _, skip := bad[f.Rel]; skip {
			continue
		}
		out = append(out, f)
	}
	return out
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

// writeListfile 写 7za 的清单文件：UTF-16LE 带 BOM、CRLF 换行。
// 不用 UTF-8：7-Zip 23.01 的 UTF-8 清单解析不了非 BMP 字符（文件名里带 emoji
// 就整批失败，报 “Incorrect item in listfile”），而 UTF-16LE 清单能吃下任意字符。
func writeListfile(p string, rels []string) error {
	var buf bytes.Buffer
	buf.Write([]byte{0xFF, 0xFE}) // BOM
	for _, rel := range rels {
		for _, u := range utf16.Encode([]rune(rel)) {
			buf.Write([]byte{byte(u), byte(u >> 8)})
		}
		buf.Write([]byte{'\r', 0x00, '\n', 0x00}) // CRLF 也必须是 UTF-16LE 码元
	}
	return os.WriteFile(p, buf.Bytes(), 0644)
}

// solidArg 生成 7z 的 -ms 参数：n<=0 表示非固实（每个文件独立块），
// 否则给固实块上限。块越小，单点损坏波及的后续文件越少，代价是跨块
// 不再互相引用、压缩率略降；7z 的默认是一整包一个块，那正是最脆的。
// 单位必须带后缀：实测 -ms=67108864（纯字节）会被 7-Zip 判为参数错误，
// -ms=64m 才行；这里统一向上取整到 MB，兼容老版本也读得懂。
func solidArg(n int64) string {
	if n <= 0 {
		return "-ms=off"
	}
	mb := (n + (1 << 20) - 1) >> 20
	return fmt.Sprintf("-ms=%dm", mb)
}

// packArgs 组装 7z 打包命令行。-mx 和 -ms 都按仓库 > settings 的优先级取值；
// 单拎出来是为了能脱离 7za 直接断言参数。
func packArgs(repo *Repo, set *Settings, password, partPath, listfile string) []string {
	return []string{
		"a", "-y", "-t7z",
		"-p" + password,
		"-mhe=on",
		fmt.Sprintf("-mx=%d", repo.compressLevel(set)),
		solidArg(repo.solidBlock(set)),
		"-scsUTF-16LE",
		"-sccUTF-8",
		partPath,
		"@" + listfile,
	}
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
	if err := writeListfile(listfile, rels); err != nil {
		return packResult{}, fmt.Errorf(M.WriteListfileFailed, err)
	}
	defer os.Remove(listfile)

	args := packArgs(repo, set, password, partPath, listfile)
	// 没有可用的 7-Zip 时 exec.Command("") 只会报一句没人看得懂的话，这里先拦下来。
	sz := repo.sevenZip(set)
	if sz == "" {
		return packResult{}, fmt.Errorf("%s", M.SevenZipMissing)
	}
	cmd := hiddenCmd(sz, args...)
	cmd.Dir = repo.Source // listfile 里的相对路径按源目录解析
	out, err := cmd.CombinedOutput()
	var skipped []string
	if err != nil {
		// 退出码 1 是 7-Zip 的 WARNING：个别文件读不到被跳过，归档本身照样完整可用
		// （实测 23.01：报 "Cannot open N file"，其余文件全部写入，包能正常列出和校验）。
		// 这跟「打包前预检发现文件被占」是同一件事，处理也必须一样：跳过、不写索引、
		// 下次重试。退出码 >=2 才是真的没打出来（写盘失败、参数错误、内存不足）。
		if exitCodeOf(err) != 1 {
			os.Remove(partPath)
			return packResult{}, fmt.Errorf(M.CompressFailed, err, string(out))
		}
		// 谁没进去，以归档头为准，不解析 7-Zip 的告警文案（文案随语言变）。
		skipped, err = missingFromArchive(sz, partPath, password, rels)
		if err != nil {
			os.Remove(partPath)
			return packResult{}, fmt.Errorf(M.CompressFailed, err, string(out))
		}
		// 一个都没打进去，说明不是「个别文件占用」，按真失败处理，别把它当成功。
		if len(skipped) >= len(rels) {
			os.Remove(partPath)
			return packResult{}, fmt.Errorf(M.CompressFailed, err, string(out))
		}
	}

	if !noVerify {
		if err := verifyArchive(sz, partPath, password); err != nil {
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
	return packResult{Name: name, Files: plan.Files, Skipped: skipped, Hash: h, Size: size, Start: plan.Start, End: plan.End}, nil
}

// cleanupStaleParts 清掉上次中断留下的 .part、.rm.tmp 与 .list 清单。
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
			continue
		}
		if strings.HasSuffix(n, ".rm.tmp") && isDelName(strings.TrimSuffix(n, ".tmp")) {
			os.Remove(filepath.Join(output, n))
			continue
		}
		// packOne 的清单文件：正常路径靠 defer 删，但 Ctrl+C / 崩溃时 defer 不执行，
		// 会一直留在输出目录里（用户实测见过）。
		if strings.HasPrefix(n, ".list-") && strings.HasSuffix(n, ".txt") {
			os.Remove(filepath.Join(output, n))
		}
	}
}

// packAll 规划并打包全部文件。每打完一个包就交给 onPack 提交一次：
// 已经提交的包不会因为后面某个包失败而消失——这就是「逐包提交」的全部意义。
// 返回已提交的包；中途失败时也把已提交的那部分一起返回，由调用方决定怎么记账。
func packAll(repo *Repo, set *Settings, files []fileInfo, password string, noVerify bool, onPack func(packResult) error) ([]packResult, error) {
	sortByMTime(files)
	limit := repo.maxPackByte(set)
	plans := planPacks(files, limit)

	last := latestPackTime(repo.Output)
	submitted := make([]packResult, 0, len(plans))

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
			return submitted, err
		}
		if err := onPack(r); err != nil {
			return submitted, err
		}
		submitted = append(submitted, r)
	}
	return submitted, nil
}

// commitPack 把一个已经落盘的包登记进账本：包 fsync → manifest → 索引。
// 顺序不能反：manifest 在前，崩了最多重复登记一次；索引在前，会留下
// verify 永远看不见的幽灵包。包必须先落盘，否则断电后索引会指向一个不存在的
// 文件，下次扫描误以为已备份而静默跳过——那是唯一会丢数据的方向。
func commitPack(repo *Repo, repoName, tag string, idx *Index, r packResult) error {
	full := filepath.Join(repo.Output, r.Name)
	kept := r.Kept()
	// 整包文件全被跳过：这不是有用的包，别登记，直接扔掉。
	if len(kept) == 0 {
		os.Remove(full)
		return nil
	}
	if err := syncFile(full); err != nil {
		return err
	}
	if err := writeManifest(repo.Output, repoName, tag, batchMeta{
		Packs:  []string{r.Name},
		Hashes: []string{r.Hash},
	}); err != nil {
		return err
	}
	id := idx.addPack(r.Name, r.Start, r.End)
	for _, f := range kept {
		idx.Files[f.Rel] = &FileEntry{Size: f.Size, MTime: f.MTime, Pack: id}
	}
	return saveIndex(repo.Output, idx)
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
	cleanupStaleParts(repo.Output)

	idx, err := loadIndex(repo.Output)
	if err != nil {
		res.Err = err
		return res
	}

	// 扫描同样早于拿锁：等锁期间文件可能又回来了，删除候选要在锁内复核一遍。
	del := make([]string, 0, len(scan.Del))
	for _, rel := range scan.Del {
		if !fileExists(filepath.Join(repo.Source, filepath.FromSlash(rel))) {
			del = append(del, rel)
		}
	}
	res.DelCount = len(del)

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

	if len(ready) == 0 && len(del) == 0 {
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

	tag := started.Format("2006-01-02_15-04-05")

	// 逐包提交：每成功一个包就 fsync → manifest → 索引。后面哪个包炸了，前面已提交的
	// 包留在盘上也留在账上，下次接着压 —— 一个文件被占用不再毁掉整批。
	var packs []packResult
	var packSkipped []string
	var packErr error
	if len(ready) > 0 {
		packs, packErr = packAll(repo, set, ready, password, noVerify, func(r packResult) error {
			if err := commitPack(repo, name, tag, idx, r); err != nil {
				return err
			}
			packSkipped = append(packSkipped, r.Skipped...)
			return nil
		})
	}

	names := make([]string, 0, len(packs))
	hashes := make([]string, 0, len(packs))
	var total int64
	for _, p := range packs {
		names = append(names, p.Name)
		hashes = append(hashes, p.Hash)
		total += p.Size
	}
	res.TotalSize = total
	res.Packs = names
	res.BatchHash = batchHashOf(hashes, "")
	// 打包时读不到的文件：跟预检跳过的一样，不写索引、下次重试。
	res.Skipped += len(packSkipped)

	shown = 0
	for _, p := range packSkipped {
		if shown >= 3 {
			break
		}
		fmt.Printf(M.PackSkipItem, p)
		shown++
	}
	if len(packSkipped) > shown {
		fmt.Printf(M.PackSkipMore, len(packSkipped)-shown)
	}

	// 中途失败：已提交的包留下；删除记录一律不写 —— 下次扫描会重新发现这些「消失」，
	// 自愈。写了一半的删除记录比没写危险得多。
	if packErr != nil {
		res.Err = packErr
		res.DelCount = 0
		if len(packs) > 0 {
			fmt.Printf(M.PartialCommit, name, len(packs), humanSize(total), len(packSkipped))
		}
		logf("backup %s: partial, %d packs committed, %s, %d files left, err=%v",
			name, len(packs), humanSize(total), len(packSkipped), packErr)
		return res
	}

	// 删除记录是整个 run 的最后一个动作，而且必须「先落 .rm、再从索引删 del 条目」。
	// 顺序反了就是永久漏备：下次扫描看不到这些路径，「消失」永远不会被记录。
	rmName := ""
	if len(del) > 0 {
		rmName = delRecordName(repo.Output, packs)
		if err := writeDeletions(repo.Output, rmName, del); err != nil {
			res.Err = err
			return res
		}
		rmHash, herr := sha256File(delPath(repo.Output, rmName))
		if herr != nil {
			res.Err = herr
			return res
		}
		if err := writeManifest(repo.Output, name, tag, batchMeta{
			Rm:     rmName,
			RmHash: rmHash,
		}); err != nil {
			res.Err = err
			return res
		}
		// 索引跟着源目录走：删掉的文件不再登记，否则下次还会被当成「刚消失」。
		for _, rel := range del {
			delete(idx.Files, rel)
		}
		if err := saveIndex(repo.Output, idx); err != nil {
			res.Err = err
			return res
		}
		res.DelCount = len(del)
		res.BatchHash = batchHashOf(hashes, rmHash)
	}
	res.DelRecord = rmName

	filesPacked := len(ready) - len(packSkipped)
	if len(packs) > 0 {
		fmt.Printf(M.DetailLine+"\n", name, filesPacked, humanSize(total), len(packs),
			fmt.Sprintf(M.EncFrom, encFrom), res.ScanDur.Milliseconds(), shortHash(res.BatchHash))
	}
	if len(del) > 0 {
		fmt.Printf(M.DeletedLine, name, len(del), rmName)
	}
	logf("backup %s: %d files, %d packs, %d deleted, %s, hash=%s",
		name, filesPacked, len(packs), len(del), humanSize(total), res.BatchHash)
	return res
}

// exitCodeOf 取出子进程的退出码。进程根本没起来（比如路径不存在）时返回 -1，
// 调用方会把它当成「非 1」，也就是硬失败——这是对的，那种情况不该静默跳过。
func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// archivePaths 读归档头，返回包内实际存在的相对路径集合（斜杠分隔）。
// 只读头不解压；加密包的头部要密码。用归档本身当裁判，而不是解析 7-Zip 的
// 告警文案 —— 文案随界面语言变，中文文件名里还可能有冒号、等号之类。
func archivePaths(sevenZip, archive, password string) (map[string]struct{}, error) {
	cmd := hiddenCmd(sevenZip, "l", "-slt", "-p"+password, "-sccUTF-8", archive)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("7za l: %v\n%s", err, string(out))
	}
	got := make(map[string]struct{})
	body := false
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "----------" { // 这行之前是归档自身的元信息，那里也有一个 Path = 归档自己
			body = true
			continue
		}
		if !body || !strings.HasPrefix(line, "Path = ") {
			continue
		}
		got[filepath.ToSlash(strings.TrimPrefix(line, "Path = "))] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return got, nil
}

// missingFromArchive 返回「计划里有、包里没有」的文件（相对路径、斜杠分隔）。
func missingFromArchive(sevenZip, archive, password string, want []string) ([]string, error) {
	got, err := archivePaths(sevenZip, archive, password)
	if err != nil {
		return nil, err
	}
	var missing []string
	for _, rel := range want {
		if _, ok := got[rel]; !ok {
			missing = append(missing, rel)
		}
	}
	return missing, nil
}

// syncFile 把包内容刷到盘上，提交索引前必须做：Windows 的 rename 不保证目录项
// 落盘，断电后可能留下「索引登记了、包不在」——那会让下次扫描误以为已备份而静默跳过。
func syncFile(p string) error {
	f, err := os.OpenFile(p, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// verifyArchive 用 7za t 校验包内所有数据块（CRC32）
func verifyArchive(sevenZip, pack, password string) error {
	cmd := hiddenCmd(sevenZip, "t", "-t7z", "-sccUTF-8", "-p"+password, pack)
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

// batchHashOf 是汇总行/日志里用的批次哈希：包哈希按顺序拼接再取 sha256；
// 一个包都没有（只有删除记录）时退化成删除记录自己的哈希。
// 注意：它只用于展示，不写进 manifest —— 逐包 SHA-256 已经足够 verify 用。
func batchHashOf(hashes []string, rmHash string) string {
	if len(hashes) == 0 {
		return rmHash
	}
	sum := sha256.Sum256([]byte(strings.Join(hashes, "|")))
	return fmt.Sprintf("%x", sum)
}

// manifestVersion 2：批次里只留「名字 → SHA-256」两张映射，不再有批次级哈希、
// 内容日期范围和 compact 标记——那些要么能推出来，要么根本没人读。
const manifestVersion = 2

// Manifest 是「校验账本」：记录每个包/删除记录在写完那一刻的 SHA-256，供 verify 事后比对。
type Manifest struct {
	Version int             `json:"version"`
	Repo    string          `json:"repo"`
	Batches []ManifestBatch `json:"batches"`
}

// ManifestBatch 是一次备份产生的东西。Time 同时是「同一次备份」的合并键。
type ManifestBatch struct {
	Time  string            `json:"time"`
	Packs map[string]string `json:"packs,omitempty"` // 包文件名 → SHA-256
	Rms   map[string]string `json:"rms,omitempty"`   // 删除记录名 → 文件 SHA-256
}

// batchMeta 是写 manifest 时要登记的这批信息
type batchMeta struct {
	Packs  []string
	Hashes []string
	Rm     string
	RmHash string
}

func manifestPath(output string) string {
	return filepath.Join(output, "manifest.json")
}

// manifestWire 兼容两代格式：v1 是 files/hashes 平行数组 + rm/rm_hash，
// v2 是 packs/rms 映射。读 v1 时在内存里升级成 v2，下一次写入就落成 v2。
type manifestWire struct {
	Version int    `json:"version"`
	Repo    string `json:"repo"`
	Batches []struct {
		Time   string            `json:"time"`
		Packs  map[string]string `json:"packs"`
		Rms    map[string]string `json:"rms"`
		Files  []string          `json:"files"`
		Hashes []string          `json:"hashes"`
		Rm     string            `json:"rm"`
		RmHash string            `json:"rm_hash"`
	} `json:"batches"`
}

func decodeManifest(b []byte) (*Manifest, error) {
	var w manifestWire
	if err := json.Unmarshal(b, &w); err != nil {
		return nil, err
	}
	m := &Manifest{Version: manifestVersion, Repo: w.Repo}
	for _, wb := range w.Batches {
		mb := ManifestBatch{Time: wb.Time}
		if len(wb.Packs) > 0 {
			mb.Packs = wb.Packs
		}
		if len(wb.Rms) > 0 {
			mb.Rms = wb.Rms
		}
		// v1 迁移：files 第 i 个包对应 hashes 第 i 个
		for i, f := range wb.Files {
			if i >= len(wb.Hashes) {
				break
			}
			if mb.Packs == nil {
				mb.Packs = map[string]string{}
			}
			mb.Packs[filepath.Base(f)] = wb.Hashes[i]
		}
		if wb.Rm != "" {
			if mb.Rms == nil {
				mb.Rms = map[string]string{}
			}
			mb.Rms[wb.Rm] = wb.RmHash
		}
		m.Batches = append(m.Batches, mb)
	}
	return m, nil
}

// manifestFormatVersion 只看文件头的 version，用来判断要不要留升级前的副本。
func manifestFormatVersion(b []byte) int {
	var v struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return 0
	}
	return v.Version
}

// readManifestForWrite 读旧账本准备追加。损坏时改名保留并报错，
// 绝不把损坏内容当成空账本继续写。
func readManifestForWrite(output, repo string) (*Manifest, error) {
	p := manifestPath(output)
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Manifest{Version: manifestVersion, Repo: repo}, nil
	}
	if err != nil {
		return nil, fmt.Errorf(M.ManifestReadFailed, err)
	}
	// 旧格式（v1）第一次被写回前留一份原样副本：万一迁移出问题、或者想降级，
	// 还能拿回旧账本。只在需要升级时做一次，之后 version=2 就不再触发。
	if manifestFormatVersion(b) < manifestVersion {
		bak := p + ".v1.bak"
		if !fileExists(bak) {
			if werr := writeFileSync(bak, b, 0o644); werr != nil {
				return nil, fmt.Errorf(M.RemanifestBakFailed, werr)
			}
		}
	}
	m, derr := decodeManifest(b)
	if derr != nil {
		bad := p + ".corrupt-" + time.Now().Format("20060102-150405.000")
		if rerr := os.Rename(p, bad); rerr != nil {
			return nil, fmt.Errorf(M.ManifestCorruptLocked, p, derr)
		}
		return nil, fmt.Errorf(M.ManifestCorrupt, p, bad, derr)
	}
	m.Version = manifestVersion
	m.Repo = repo
	return m, nil
}

// writeManifestFile 原子写回整份账本。
func writeManifestFile(output string, m *Manifest) error {
	m.Version = manifestVersion
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	p := manifestPath(output)
	tmp := p + ".tmp"
	if err := writeFileSync(tmp, b, 0o644); err != nil {
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	return nil
}

// writeManifest 追加一个批次到 manifest.json。同一次备份（tag 相同）的多个包
// 合并进同一批：逐包提交不该把「一批 = 一次备份」变成「一批 = 一个包」。
func writeManifest(output, repo, tag string, batch batchMeta) error {
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf(M.ManifestWriteFailed, err)
	}
	m, err := readManifestForWrite(output, repo)
	if err != nil {
		return err
	}
	names := baseNames(batch.Packs)
	n := len(m.Batches)
	if n > 0 && tag != "" && m.Batches[n-1].Time == tag {
		b := &m.Batches[n-1]
		if b.Packs == nil {
			b.Packs = map[string]string{}
		}
		for i, name := range names {
			if i < len(batch.Hashes) {
				b.Packs[name] = batch.Hashes[i]
			}
		}
		if batch.Rm != "" {
			if b.Rms == nil {
				b.Rms = map[string]string{}
			}
			b.Rms[batch.Rm] = batch.RmHash
		}
	} else {
		nb := ManifestBatch{Time: tag}
		if len(names) > 0 {
			nb.Packs = map[string]string{}
			for i, name := range names {
				if i < len(batch.Hashes) {
					nb.Packs[name] = batch.Hashes[i]
				}
			}
		}
		if batch.Rm != "" {
			nb.Rms = map[string]string{batch.Rm: batch.RmHash}
		}
		m.Batches = append(m.Batches, nb)
	}
	return writeManifestFile(output, m)
}

// updateManifestPackHash 在 compact 重写包之后原地更新它的哈希。
// 批次归属不变，也不追加「compact 批次」——重写不是一次备份。
func updateManifestPackHash(output, repo, pack, hash string) error {
	m, err := readManifestForWrite(output, repo)
	if err != nil {
		return err
	}
	for i := range m.Batches {
		if _, ok := m.Batches[i].Packs[pack]; ok {
			m.Batches[i].Packs[pack] = hash
			return writeManifestFile(output, m)
		}
	}
	// 账本里没有这个包（比如它本来就没登记）：补一批，别让新哈希丢掉
	m.Batches = append(m.Batches, ManifestBatch{
		Time:  time.Now().Format("2006-01-02_15-04-05"),
		Packs: map[string]string{pack: hash},
	})
	return writeManifestFile(output, m)
}
