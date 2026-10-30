package app

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// packSortKey 把包名拆成（定宽前缀, 重名序号）。
// 前缀是 22 字符的 YYYYMMDD_hhmmss_SSS，定宽所以字典序就是时间序；
// 重名序号单独按数字比，避免 _10 排在 _2 前面。
func packSortKey(name string) (string, int) {
	stem := strings.TrimSuffix(name, ".7z")
	if len(stem) > packStemLen && stem[packStemLen] == '_' {
		n, err := strconv.Atoi(stem[packStemLen+1:])
		if err == nil {
			return stem[:packStemLen], n
		}
		return stem[:packStemLen], 0
	}
	return stem, 0
}

// sortedPacks 列出包并按打包时刻升序排序，同时校验严格单调。
// 顺序是恢复正确性的唯一依据，所以宁可直接报错也不硬着头皮解压。
func sortedPacks(output string) ([]string, error) {
	names, err := listPackFiles(output)
	if err != nil {
		return nil, err
	}
	sort.Slice(names, func(i, j int) bool {
		pi, ni := packSortKey(names[i])
		pj, nj := packSortKey(names[j])
		if pi != pj {
			return pi < pj
		}
		return ni < nj
	})
	for i := 1; i < len(names); i++ {
		pi, ni := packSortKey(names[i-1])
		pj, nj := packSortKey(names[i])
		if pi > pj || (pi == pj && ni >= nj) {
			return nil, fmt.Errorf(M.PackOrderBroken, names[i-1], names[i])
		}
	}
	return names, nil
}

// restoreDest 计算实际解压目标目录。wrap 为真时在外层再包一个
// restore-<仓库>-<时间> 文件夹（时间取恢复开始时刻），避免把文件直接平铺到 --dest 根上。
func restoreDest(name, dest string, wrap bool) (string, error) {
	base, err := absClean(dest)
	if err != nil {
		return "", fmt.Errorf(M.PathInvalid, dest, err)
	}
	if !wrap {
		return base, nil
	}
	stem := packStem(time.Now())
	for i := 0; ; i++ {
		folder := fmt.Sprintf("restore-%s-%s", name, stem)
		if i > 0 {
			folder = fmt.Sprintf("restore-%s-%s_%d", name, stem, i)
		}
		full := filepath.Join(base, folder)
		if _, statErr := os.Stat(full); os.IsNotExist(statErr) {
			return full, nil
		}
	}
}

// restoreOpts 是全量恢复的开关集合，参数多了以后按名字传更不容易搞错
type restoreOpts struct {
	Dest      string
	Password  string
	Overwrite bool
	CheckList bool
	Wrap      bool
	NoDelete  bool
	Notify    bool
}

// runRestore 全量恢复：包和删除记录合成一条时间线，包按打包时刻升序解压，
// 后压的覆盖先压的；记录里删掉的文件在该批次最后一个包解压之后移除，
// 于是恢复出来的目录等于“最后一次备份时的源目录”。
// 某个包解压报错不会中止整场恢复：能解出来的已经落盘，剩下的继续解，最后统一报。
func runRestore(name string, cfg *Config, set *Settings, opt restoreOpts) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	// 输出目录还不存在（从没备份成功过）：别让 os.ReadDir 的原始错误漏给用户
	if _, serr := os.Stat(repo.Output); serr != nil {
		return fmt.Errorf(M.NoPacksRestore, repo.Output)
	}
	// 先把所有删除记录读完并校验：坏记录要在动目标目录之前就暴露出来。
	dels, derr := loadDelRecords(repo.Output)
	if derr != nil {
		if !opt.NoDelete {
			return derr
		}
		// 用户明确要求不动删除，那记录坏了也不该挡住恢复文件
		fmt.Printf(M.DelSkipped, derr)
		dels = map[string][]string{}
	}
	delNames := make([]string, 0, len(dels))
	for n := range dels {
		delNames = append(delNames, n)
	}
	entries, err := buildTimeline(repo.Output, delNames)
	if err != nil {
		return err
	}
	packCount := 0
	for _, e := range entries {
		if e.Pack != "" {
			packCount++
		}
	}
	if packCount == 0 {
		return fmt.Errorf(M.NoPacksRestore, repo.Output)
	}

	mode := M.ModeRename
	if opt.Overwrite {
		mode = M.ModeOverwrite
	}

	destText := opt.Dest
	if destText != "" && opt.Wrap {
		if t, terr := restoreDest(name, opt.Dest, opt.Wrap); terr == nil {
			destText = t
		}
	}
	if destText == "" {
		destText = M.DestUnset
	}

	if opt.CheckList {
		var total int64
		fmt.Printf(M.CheckListHeader, packCount, len(dels), mode, destText)
		if opt.Dest != "" && opt.Wrap {
			fmt.Printf(M.RestoreWrapLine, destText)
		}
		for i, e := range entries {
			if e.Pack == "" {
				fmt.Printf(M.CheckListDelItem, i+1, e.Del, len(dels[e.Del]))
				continue
			}
			var sz int64
			if st, serr := os.Stat(filepath.Join(repo.Output, e.Pack)); serr == nil {
				sz = st.Size()
				total += sz
			}
			fmt.Printf(M.CheckListItem, i+1, e.Pack, humanSize(sz))
		}
		fmt.Printf(M.CheckListTotal, humanSize(total))
		return nil
	}

	if opt.Dest == "" {
		return errors.New(M.MissingDest)
	}
	if err := checkRestoreDest(repo, opt.Dest); err != nil {
		return err
	}
	target, err := restoreDest(name, opt.Dest, opt.Wrap)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf(M.CreateDestFailed, err)
	}
	sz := repo.sevenZip(set)
	if sz == "" {
		return errors.New(M.SevenZipMissing)
	}
	if opt.Overwrite {
		fmt.Print(M.RestoreOverwriteWarn)
	}
	if opt.Wrap {
		fmt.Printf(M.RestoreWrapLine, target)
	}

	ao := "-aou"
	if opt.Overwrite {
		ao = "-aoa"
	}

	deletedTotal := 0
	delMissingTotal := 0
	var deletedPaths []string
	var failedPacks []string
	failedStems := make(map[string]bool, 0)
	pwHinted := false
	for _, e := range entries {
		if e.Pack != "" {
			full := filepath.Join(repo.Output, e.Pack)
			fmt.Printf(M.Extracting, e.Pack)
			// 7z 是“能解的先解、坏的报 CRC”，所以报错时包里其余文件照样落盘，
			// 这里不能当成整包放弃：记下来、继续解后面的包，最后统一报。
			cmd := hiddenCmd(sz, "x", "-y", ao, "-sccUTF-8", "-p"+opt.Password, full, "-o"+target,
				"-x!"+placeholderName)
			out, xerr := cmd.CombinedOutput()
			if xerr != nil {
				failedPacks = append(failedPacks, e.Pack)
				failedStems[e.Stem] = true
				fmt.Printf(M.PackFailed, e.Pack)
				if line := sevenZipErrorLine(string(out)); line != "" {
					fmt.Printf(M.PackErrLine, line)
					// 只有“整包都打不开”才像是密码问题；CRC 错误里也带 “Wrong password?”，别误报
					low := strings.ToLower(line)
					unopen := strings.Contains(low, "cannot open") || strings.Contains(low, "can not open") || strings.Contains(low, "can't open")
					if !pwHinted && unopen && strings.Contains(low, "password") {
						fmt.Print(M.PackPasswordHint)
						pwHinted = true
					}
				}
				logf("restore %s: pack %s failed: %v", name, e.Pack, xerr)
			}
			continue
		}
		if opt.NoDelete {
			continue
		}
		// 这一批的包解压报错，源目录当时是什么状态就不知道了；
		// 这时候照着记录删文件只会把损失叠上去，跳过不删。
		if failedStems[e.Stem] {
			fmt.Printf(M.DelSkippedAfterFail, e.Del)
			continue
		}
		rels := dels[e.Del]
		removed, missing, derr := applyDeletions(target, rels)
		if derr != nil {
			fmt.Print(M.RestorePartial)
			return derr
		}
		// 「删不到」是正常的：compact 清掉旧副本之后，这条记录要删的东西本来就没被解出来。
		// 只计数，不中止，更不能把后面重新放回来的同名文件当成余党。
		delMissingTotal += missing
		if removed > 0 {
			fmt.Printf(M.DeletingBatch, e.Del, removed)
			deletedTotal += removed
			deletedPaths = append(deletedPaths, rels...)
		}
	}
	if len(deletedPaths) > 0 {
		pruneEmptyDirs(target, deletedPaths)
	}
	if deletedTotal > 0 {
		fmt.Printf(M.RestoreDelSummary, deletedTotal)
	}
	if delMissingTotal > 0 {
		fmt.Printf(M.RestoreDelMissing, delMissingTotal)
	}
	// compact 把腾空的包换成只含占位文件的包，占位文件不该出现在恢复结果里
	if n, perr := removePlaceholder(target); perr == nil && n > 0 {
		fmt.Printf(M.PlaceholderRemoved, n, placeholderName)
	}
	fmt.Printf(M.RestoreDone, target)
	logf("restore %s -> %s (%d packs, %d deleted, %d failed, overwrite=%v, wrap=%v)",
		name, target, packCount, deletedTotal, len(failedPacks), opt.Overwrite, opt.Wrap)
	if len(failedPacks) > 0 {
		fmt.Printf(M.RestoreIncomplete, len(failedPacks))
		if opt.Notify {
			if err := sendToast(M.ToastRestoreTitle, fmt.Sprintf(M.ToastRestoreBody, name, len(failedPacks))); err != nil {
				fmt.Printf(M.NotifyFailed, err)
			}
		}
		return fmt.Errorf(M.RestoreFailedPacks, len(failedPacks))
	}
	return nil
}

// sevenZipErrorLine 从 7z 的输出里挑一行能读懂的错误，
// 别把整屏横幅（版本号、扫描、统计）倒给用户。
// 7z 打错包时先打一行 "ERROR: <包路径>"，那行只是定位不是原因，跳过。
func sevenZipErrorLine(out string) string {
	fallback := ""
	for _, raw := range strings.Split(out, "\n") {
		t := strings.TrimSpace(raw)
		if t == "" || strings.HasPrefix(t, "7-Zip") || strings.HasPrefix(t, "Scanning") ||
			strings.HasPrefix(t, "Extracting archive") || strings.HasPrefix(t, "Path = ") ||
			strings.HasSuffix(t, ".7z") {
			continue
		}
		switch {
		case strings.Contains(t, "CRC Failed"), strings.Contains(t, "Data Error"),
			strings.Contains(t, "Headers Error"), strings.Contains(t, "Wrong password"),
			strings.HasPrefix(t, "Cannot open"), strings.HasPrefix(t, "Can not open"),
			strings.HasPrefix(t, "Can't open"):
			return t
		}
		if fallback == "" && strings.HasPrefix(t, "ERROR:") {
			fallback = t
		}
	}
	return fallback
}

// checkRestoreDest 拒绝会把源文件或备份自身覆盖掉的恢复目标。
func checkRestoreDest(repo *Repo, dest string) error {
	d, err := absClean(dest)
	if err != nil {
		return fmt.Errorf(M.PathInvalid, dest, err)
	}
	if isVolumeRoot(d) {
		return fmt.Errorf(M.RestoreDestRoot, d)
	}
	if src, err := absClean(repo.Source); err == nil && samePathFold(d, src) {
		return fmt.Errorf(M.RestoreDestSource, d)
	}
	if out, err := absClean(repo.Output); err == nil && withinFold(d, out) {
		return fmt.Errorf(M.RestoreDestInOutput, d, out)
	}
	return nil
}

// matchPath 判断包内路径是否命中用户给的 spec（支持目录前缀和 * ? 通配）
func matchPath(p, spec string) bool {
	if p == spec || strings.HasPrefix(p, spec+"/") {
		return true
	}
	if strings.ContainsAny(spec, "*?") {
		if ok, err := path.Match(spec, p); err == nil && ok {
			return true
		}
	}
	return false
}

// packHasPath 只读归档头，判断某个包里有没有这个路径
func packHasPath(sevenZip, pack, password, spec string) (bool, error) {
	entries, err := listPackContents(sevenZip, pack, password)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir {
			continue
		}
		if matchPath(filepath.ToSlash(e.Path), spec) {
			return true, nil
		}
	}
	return false, nil
}

// resolvePackSpec 把 序号 / 完整名 / 名字前缀 解析成包名
func resolvePackSpec(output, spec string) string {
	packs, err := sortedPacks(output)
	if err != nil {
		return ""
	}
	if n, e := strconv.Atoi(spec); e == nil {
		if n >= 1 && n <= len(packs) {
			return packs[n-1]
		}
		return ""
	}
	for _, p := range packs {
		if p == spec {
			return p
		}
	}
	for _, p := range packs {
		if strings.HasPrefix(p, spec) {
			return p
		}
	}
	return ""
}

// runRestoreOnly 只恢复一个相对路径（文件或目录），不动其他文件。
// 不给 --from 时取索引里记录的最新版本；checkList 则列出包含它的所有包。
func runRestoreOnly(name string, cfg *Config, set *Settings, dest, password, only, from string, overwrite, checkList bool) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	spec := strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(only)), "./")
	spec = strings.TrimPrefix(spec, "/")
	if spec == "" {
		return errors.New(M.MissingOnly)
	}
	sevenZip := repo.sevenZip(set)

	if checkList {
		packs, perr := sortedPacks(repo.Output)
		if perr != nil {
			return perr
		}
		var hits []string
		for _, p := range packs {
			ok, e := packHasPath(sevenZip, filepath.Join(repo.Output, p), password, spec)
			if e != nil {
				continue
			}
			if ok {
				hits = append(hits, p)
			}
		}
		if len(hits) == 0 {
			return fmt.Errorf(M.OnlyNoVersions, spec)
		}
		fmt.Printf(M.OnlyVersions, spec, len(hits))
		for i, p := range hits {
			var sz int64
			if st, e := os.Stat(filepath.Join(repo.Output, p)); e == nil {
				sz = st.Size()
			}
			fmt.Printf(M.CheckListItem, i+1, p, humanSize(sz))
		}
		return nil
	}

	if dest == "" {
		return errors.New(M.MissingDest)
	}
	if err := checkRestoreDest(repo, dest); err != nil {
		return err
	}

	var targets []string
	if from != "" {
		pack := resolvePackSpec(repo.Output, from)
		if pack == "" {
			return fmt.Errorf(M.PackNotSpecified, from)
		}
		targets = []string{pack}
	} else {
		// 精确文件路径：索引直接给出最新版本所在包
		if idx, idxErr := loadIndex(repo.Output); idxErr == nil {
			if fe := idx.Files[spec]; fe != nil {
				if name := idx.packName(fe.Pack); name != "" {
					targets = []string{name}
				}
			}
		}
		if len(targets) == 0 {
			// 目录或索引缺失：收集所有包含它的包，按时间升序全部解压，后压的赢
			packs, perr := sortedPacks(repo.Output)
			if perr != nil {
				return perr
			}
			for _, p := range packs {
				ok, e := packHasPath(sevenZip, filepath.Join(repo.Output, p), password, spec)
				if e != nil {
					continue
				}
				if ok {
					targets = append(targets, p)
				}
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf(M.PathNotBackedUp, spec)
		}
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf(M.CreateDestFailed, err)
	}
	ao := "-aou"
	if overwrite {
		ao = "-aoa"
	}
	patterns := []string{spec}
	if !strings.ContainsAny(spec, "*?") {
		patterns = append(patterns, spec+"/*")
	}
	for _, pack := range targets {
		fmt.Printf(M.Extracting, pack)
		args := []string{"x", "-y", ao, "-sccUTF-8", "-p" + password, filepath.Join(repo.Output, pack), "-o" + dest,
			"-x!" + placeholderName}
		args = append(args, patterns...)
		cmd := hiddenCmd(sevenZip, args...)
		out, xerr := cmd.CombinedOutput()
		if xerr != nil {
			fmt.Print(M.RestorePartial)
			return fmt.Errorf(M.ExtractFailed, pack, xerr, string(out))
		}
	}
	if strings.ContainsAny(spec, "*?") {
		fmt.Printf(M.OnlyDone, spec, dest)
	} else {
		fmt.Printf(M.OnlyDone, spec, filepath.Join(dest, filepath.FromSlash(spec)))
	}
	logf("restore-only %s: %s from %d pack(s) -> %s", name, spec, len(targets), dest)
	return nil
}
