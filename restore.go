package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// runRestore 全量恢复：所有包按打包时刻升序解压，后压的覆盖先压的。
func runRestore(name string, cfg *Config, set *Settings, dest, password string, overwrite, checkList, wrap bool) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	packs, err := sortedPacks(repo.Output)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		return fmt.Errorf(M.NoPacksRestore, repo.Output)
	}

	mode := M.ModeRename
	if overwrite {
		mode = M.ModeOverwrite
	}

	if checkList {
		var total int64
		sizes := make([]int64, len(packs))
		for i, p := range packs {
			if st, e := os.Stat(filepath.Join(repo.Output, p)); e == nil {
				sizes[i] = st.Size()
				total += st.Size()
			}
		}
		destText := dest
		if destText != "" && wrap {
			if t, terr := restoreDest(name, dest, wrap); terr == nil {
				destText = t
			}
		}
		if destText == "" {
			destText = M.DestUnset
		}
		fmt.Printf(M.CheckListHeader, len(packs), mode, destText)
		if dest != "" && wrap {
			fmt.Printf(M.RestoreWrapLine, destText)
		}
		for i, p := range packs {
			fmt.Printf(M.CheckListItem, i+1, p, humanSize(sizes[i]))
		}
		fmt.Printf(M.CheckListTotal, humanSize(total))
		return nil
	}

	if dest == "" {
		return errors.New(M.MissingDest)
	}
	if err := checkRestoreDest(repo, dest); err != nil {
		return err
	}
	target, err := restoreDest(name, dest, wrap)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf(M.CreateDestFailed, err)
	}
	if overwrite {
		fmt.Print(M.RestoreOverwriteWarn)
	}
	if wrap {
		fmt.Printf(M.RestoreWrapLine, target)
	}

	ao := "-aou"
	if overwrite {
		ao = "-aoa"
	}
	for _, p := range packs {
		full := filepath.Join(repo.Output, p)
		fmt.Printf(M.Extracting, p)
		cmd := exec.Command(repo.sevenZip(set), "x", "-y", ao, "-sccUTF-8", "-p"+password, full, "-o"+target)
		out, xerr := cmd.CombinedOutput()
		if xerr != nil {
			fmt.Print(M.RestorePartial)
			return fmt.Errorf(M.ExtractFailed, p, xerr, string(out))
		}
	}
	fmt.Printf(M.RestoreDone, target)
	logf("restore %s -> %s (%d packs, overwrite=%v, wrap=%v)", name, target, len(packs), overwrite, wrap)
	return nil
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
		args := []string{"x", "-y", ao, "-sccUTF-8", "-p" + password, filepath.Join(repo.Output, pack), "-o" + dest}
		args = append(args, patterns...)
		cmd := exec.Command(sevenZip, args...)
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
