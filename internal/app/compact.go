package app

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// compact 是「重写已有包、丢掉死副本」的那条路径。它跟备份完全解耦：
// 备份只往新包里写，compact 只动已经躺在盘上的包，包名保持不变。
//
// 为什么包名必须保持不变：云端（夸克）对同名文件做原地覆盖，旧内容进「历史版本」。
// 改个名字就等于在云端多留一份，省下来的空间原样还回去。
//
// 一个包里的文件只分两种：索引里指到这个包的（当前版本，留），
// 以及指向别的包或根本没有条目的（已被新版取代 / 已从源目录删除，丢）。
// 索引就是判定依据，所以 compact 前必须保证索引是完整的（见 checkIndexCoversPacks）。

// placeholderName 是「这个包已经空了」的占位文件。
// 7z 打不出真正的空包，而删掉整个包又会打乱「按名字顺序解压」这条恢复依据，
// 所以留一个 0 字节、名字一眼看得出该删的文件顶位。
// 恢复时 zinc 会自己把它清掉（见 restore.go），手工 7z x 也一眼知道怎么处理。
const placeholderName = "placeholder-Please_Delete_This"

const (
	compactTmpPrefix  = "compact-"
	compactTmpSuffix  = "-tmp"
	defaultMinGainAbs = 8 << 20 // 收益下限 8 MiB：免得为几 MB 重传 1 GB
)

var procGetDiskFreeSpaceEx = kernel32.NewProc("GetDiskFreeSpaceExW")

// compactPlan 描述一个包该怎么重写
type compactPlan struct {
	Pack   string
	ID     int // 这个包在索引里的编号
	Keep   []packEntry
	Drop   []packEntry
	KeepSz int64
	DropSz int64
	PackSz int64
}

// compactResult 是一个重写完成的包
type compactResult struct {
	Pack    string
	Hash    string
	OldSize int64
	NewSize int64
	From    string
	To      string
	Empty   bool
}

func isPlaceholder(rel string) bool {
	return strings.EqualFold(filepath.ToSlash(rel), placeholderName)
}

// compactTmpDir 是重写一个包时的工作目录。必须在输出目录里（同卷），
// 这样最后那一步 os.Rename 才是原子的目录项替换，而不是跨卷复制。
func compactTmpDir(output, pack string) string {
	stem := strings.TrimSuffix(pack, ".7z")
	return filepath.Join(output, compactTmpPrefix+stem+compactTmpSuffix)
}

// isCompactTmpName 判断是不是 compact 留下的工作目录
func isCompactTmpName(name string) bool {
	return strings.HasPrefix(name, compactTmpPrefix) && strings.HasSuffix(name, compactTmpSuffix)
}

// cleanupCompactTmp 清掉中断的 compact 留下的工作目录。调用前必须已持有仓库锁。
func cleanupCompactTmp(output string) {
	ents, err := os.ReadDir(output)
	if err != nil {
		return
	}
	for _, e := range ents {
		if !e.IsDir() || !isCompactTmpName(e.Name()) {
			continue
		}
		rmTreeForce(filepath.Join(output, e.Name()))
	}
}

// rmTreeForce 删目录；Windows 上只读属性会让 RemoveAll 删不干净，
// 而 7z 解出来的文件可能带着源文件的只读属性。
func rmTreeForce(dir string) {
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && info != nil && info.Mode().Perm()&0200 == 0 {
			os.Chmod(p, 0644)
		}
		return nil
	})
	os.RemoveAll(dir)
}

// freeSpaceAt 返回该路径所在卷的可用字节数
func freeSpaceAt(dir string) (int64, error) {
	vol := filepath.VolumeName(dir)
	if vol == "" {
		return 0, fmt.Errorf("no volume in %s", dir)
	}
	root, err := syscall.UTF16PtrFromString(vol + `\`)
	if err != nil {
		return 0, err
	}
	var avail, total, free uint64
	r, _, e := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(root)),
		uintptr(unsafe.Pointer(&avail)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&free)),
	)
	if r == 0 {
		return 0, e
	}
	return int64(avail), nil
}

// buildCompactPlan 用索引判定包内哪些文件仍是当前版本
func buildCompactPlan(repo *Repo, set *Settings, idx *Index, password, pack string) (compactPlan, error) {
	sz := repo.sevenZip(set)
	if sz == "" {
		return compactPlan{}, fmt.Errorf("%s", M.SevenZipMissing)
	}
	full := filepath.Join(repo.Output, pack)
	entries, err := listPackContents(sz, full, password)
	if err != nil {
		return compactPlan{}, fmt.Errorf(M.CompactPackListFailed, pack, err)
	}

	plan := compactPlan{Pack: pack}
	if st, serr := os.Stat(full); serr == nil {
		plan.PackSz = st.Size()
	}
	for _, p := range idx.Packs {
		if p.Name == pack {
			plan.ID = p.ID
			break
		}
	}
	for _, e := range entries {
		if e.IsDir || e.Path == "" || isPlaceholder(e.Path) {
			continue
		}
		rel := filepath.ToSlash(e.Path)
		// 只有「索引说这个文件现在就在这个包里、而且大小对得上」才算当前版本。
		// 大小对不上说明包里的这一份已经不是索引描述的内容，宁可不删。
		cur := idx.Files[rel]
		if cur != nil && cur.Pack == plan.ID && cur.Size == e.Size {
			plan.Keep = append(plan.Keep, e)
			plan.KeepSz += e.Size
			continue
		}
		plan.Drop = append(plan.Drop, e)
		plan.DropSz += e.Size
	}
	return plan, nil
}

// parseMinGain 解析收益阈值：5% / 8M / 0。
// 百分比按包大小换算，但不会低于 8 MiB——重打一个包要读一遍写一遍，
// 云端还要重传一遍，省几 MB 不值当。
func parseMinGain(s string, packSize int64) (int64, error) {
	v := strings.TrimSpace(s)
	if v == "" || v == "0" {
		return 0, nil
	}
	if strings.HasSuffix(v, "%") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64)
		if err != nil || f < 0 {
			return 0, fmt.Errorf(M.CompactBadGain, s)
		}
		n := int64(float64(packSize) * f / 100)
		if n < defaultMinGainAbs {
			n = defaultMinGainAbs
		}
		return n, nil
	}
	n, err := parseSize(v)
	if err != nil {
		return 0, fmt.Errorf(M.CompactBadGain, s)
	}
	return n, nil
}

// checkIndexCoversPacks 确认索引里登记的包与实际目录里的一致。
// 索引是 compact 判断「谁还是当前版本」的唯一依据：如果索引只覆盖了一部分包
// （比如用 reindex --pack-start 重建过），另外那些包里的文件会全部被当成
// 「已删除」——那是能把备份清空的方向，必须挡在动手之前。
func checkIndexCoversPacks(idx *Index, packs []string, name, indexFile string) error {
	have := make(map[string]struct{}, len(idx.Packs))
	for _, p := range idx.Packs {
		have[p.Name] = struct{}{}
	}
	missing := 0
	for _, n := range packs {
		if _, ok := have[n]; !ok {
			missing++
		}
	}
	if missing > 0 || len(idx.Packs) != len(packs) {
		return fmt.Errorf(M.CompactIndexStale, len(idx.Packs), len(packs), indexFile, name)
	}
	return nil
}

// rewritePack 重写一个包：解出保留的文件 → 删掉死副本（自然就没解出来）→ 同参数重打 → 原子替换。
// 原包在最后一步之前一直不动，任何一步失败都退回原状。
func rewritePack(repo *Repo, set *Settings, plan compactPlan, password string, noVerify bool) (compactResult, error) {
	sz := repo.sevenZip(set)
	if sz == "" {
		return compactResult{}, fmt.Errorf("%s", M.SevenZipMissing)
	}
	tmp := compactTmpDir(repo.Output, plan.Pack)
	rmTreeForce(tmp) // 上一轮中断的残留
	if err := os.MkdirAll(tmp, 0755); err != nil {
		return compactResult{}, fmt.Errorf(M.CompactRebuildFailed, plan.Pack, err)
	}
	defer rmTreeForce(tmp)

	rels := make([]string, 0, len(plan.Keep))
	for _, e := range plan.Keep {
		rels = append(rels, filepath.ToSlash(e.Path))
	}
	sort.Strings(rels)

	empty := len(rels) == 0
	if empty {
		// 空包：只放一个 0 字节占位文件，别留下一个读得出内容、却什么都不含的假包
		if err := os.WriteFile(filepath.Join(tmp, placeholderName), nil, 0644); err != nil {
			return compactResult{}, fmt.Errorf(M.CompactRebuildFailed, plan.Pack, err)
		}
		rels = []string{placeholderName}
	} else {
		// 只解出要留的文件：把死副本解到盘上是纯粹的浪费和风险
		if err := writeListfile(filepath.Join(tmp, ".zinc-keep.txt"), rels); err != nil {
			return compactResult{}, fmt.Errorf(M.WriteListfileFailed, err)
		}
		args := []string{"x", "-y", "-o" + tmp, "-p" + password, "-spd", "-scsUTF-16LE", "-sccUTF-8",
			filepath.Join(repo.Output, plan.Pack), "@.zinc-keep.txt"}
		cmd := hiddenCmd(sz, args...)
		cmd.Dir = tmp
		out, err := cmd.CombinedOutput()
		if err != nil {
			return compactResult{}, fmt.Errorf(M.CompactExtractFailed, plan.Pack, err, string(out))
		}
		// 解出来的东西必须跟保留清单一一对上：少一个就说明有文件名在
		// Windows 上落不了地（大小写冲突、保留名、超长路径），这时候重打只会把文件弄丢。
		for _, e := range plan.Keep {
			st, serr := os.Stat(filepath.Join(tmp, filepath.FromSlash(e.Path)))
			if serr != nil || st.Size() != e.Size {
				return compactResult{}, fmt.Errorf(M.CompactExtractMismatch, plan.Pack, filepath.ToSlash(e.Path))
			}
		}
	}

	// 重打：参数与备份时完全一致（同样的 -mx / -ms / -mhe），
	// 否则重打出来的包和原来的包不是一类东西，恢复和校验都会变味。
	part := filepath.Join(tmp, plan.Pack+".part")
	if err := writeListfile(filepath.Join(tmp, ".zinc-pack.txt"), rels); err != nil {
		return compactResult{}, fmt.Errorf(M.WriteListfileFailed, err)
	}
	args := packArgs(repo, set, password, part, ".zinc-pack.txt")
	cmd := hiddenCmd(sz, args...)
	cmd.Dir = tmp
	if out, err := cmd.CombinedOutput(); err != nil {
		return compactResult{}, fmt.Errorf(M.CompactRebuildFailed, plan.Pack, fmt.Errorf("%v\n%s", err, string(out)))
	}

	if !noVerify {
		if err := verifyArchive(sz, part, password); err != nil {
			return compactResult{}, fmt.Errorf(M.CompactVerifyFailed, plan.Pack, err)
		}
	}
	// 归档头对账：重打出来的包必须恰好装着保留清单，多一个少一个都不能替换。
	got, err := archivePaths(sz, part, password)
	if err != nil {
		return compactResult{}, fmt.Errorf(M.CompactPackListFailed, plan.Pack, err)
	}
	if len(got) != len(rels) {
		return compactResult{}, fmt.Errorf(M.CompactCountMismatch, plan.Pack, len(rels), len(got))
	}
	for _, rel := range rels {
		if _, ok := got[rel]; !ok {
			return compactResult{}, fmt.Errorf(M.CompactCountMismatch, plan.Pack, len(rels), len(got))
		}
	}

	h, err := sha256File(part)
	if err != nil {
		return compactResult{}, fmt.Errorf(M.CompactRebuildFailed, plan.Pack, err)
	}
	var newSz int64
	if st, serr := os.Stat(part); serr == nil {
		newSz = st.Size()
	}
	// 落盘后再改名：rename 不保证目录项落盘，断电会留下「账上有、包不在」。
	if err := syncFile(part); err != nil {
		return compactResult{}, fmt.Errorf(M.CompactRebuildFailed, plan.Pack, err)
	}
	// 同卷 rename，原子替换：要么还是旧包，要么已经是新包，不存在半个。
	if err := os.Rename(part, filepath.Join(repo.Output, plan.Pack)); err != nil {
		return compactResult{}, fmt.Errorf(M.CompactRebuildFailed, plan.Pack, err)
	}

	res := compactResult{Pack: plan.Pack, Hash: h, OldSize: plan.PackSz, NewSize: newSz, Empty: empty}
	res.From, res.To = keepRange(plan.Keep)
	return res, nil
}

// keepRange 算剩下的内容覆盖的 mtime 日期范围
func keepRange(keep []packEntry) (string, string) {
	var minT, maxT time.Time
	for _, e := range keep {
		if e.MTime.IsZero() {
			continue
		}
		if minT.IsZero() || e.MTime.Before(minT) {
			minT = e.MTime
		}
		if maxT.IsZero() || e.MTime.After(maxT) {
			maxT = e.MTime
		}
	}
	if minT.IsZero() {
		return "", ""
	}
	return minT.Format("20060102"), maxT.Format("20060102")
}

// runCompact 是 compact 命令的主体
func runCompact(name string, cfg *Config, set *Settings, password, packSpec, minGain string, yes, noVerify, notify bool) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(repo.Output); statErr != nil {
		return fmt.Errorf(M.CompactNoPacks, repo.Output)
	}
	// 先把阈值语法过一遍：一个死副本都没有时不会再解析到它，别让错参数静默通过
	if _, gerr := parseMinGain(minGain, 1<<30); gerr != nil {
		return gerr
	}

	lock, err := acquireLock(repo.Output, set.lockTimeout())
	if err != nil {
		return err
	}
	defer lock.release()
	cleanupStaleParts(repo.Output)
	cleanupCompactTmp(repo.Output)

	idx, err := loadIndex(repo.Output)
	if err != nil {
		return err
	}
	packs, err := sortedPacks(repo.Output)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		return fmt.Errorf(M.CompactNoPacks, repo.Output)
	}
	if err := checkIndexCoversPacks(idx, packs, name, indexPath(repo.Output)); err != nil {
		return err
	}

	if packSpec != "" {
		one := resolvePackSpec(repo.Output, packSpec)
		if one == "" {
			return fmt.Errorf(M.CompactPackNotFound, packSpec)
		}
		packs = []string{one}
	}

	fmt.Printf(M.CompactStart, name, len(packs))
	if !yes {
		fmt.Print(M.CompactDryRunHead)
	}

	var plans []compactPlan
	var skippedNone, skippedSmall int
	var totalDrop int64
	for _, p := range packs {
		plan, perr := buildCompactPlan(repo, set, idx, password, p)
		if perr != nil {
			fmt.Printf(M.CompactPackFailed, p, perr)
			continue
		}
		if len(plan.Drop) == 0 {
			fmt.Printf(M.CompactSkipNone, p)
			skippedNone++
			continue
		}
		need, gerr := parseMinGain(minGain, plan.PackSz)
		if gerr != nil {
			return gerr
		}
		pct := 0.0
		if plan.PackSz > 0 {
			pct = float64(plan.DropSz) / float64(plan.PackSz) * 100
		}
		if plan.DropSz < need {
			fmt.Printf(M.CompactSkipSmall, p, len(plan.Drop), humanSize(plan.DropSz), pct, humanSize(need))
			skippedSmall++
			continue
		}
		fmt.Printf(M.CompactPlanPack, p, len(plan.Drop), humanSize(plan.DropSz), pct)
		plans = append(plans, plan)
		totalDrop += plan.DropSz
	}

	if len(plans) == 0 {
		fmt.Printf(M.CompactNothing, skippedNone, skippedSmall)
		return nil
	}

	if !yes {
		fmt.Printf(M.CompactPlanSummary, len(plans), humanSize(totalDrop))
		fmt.Print(M.CompactDryRunTail)
		return nil
	}

	fmt.Printf(M.CompactPlanSummary, len(plans), humanSize(totalDrop))

	free, ferr := freeSpaceAt(repo.Output)
	var done []compactResult
	var failed int
	var oldTotal, newTotal int64
	for _, plan := range plans {
		// 重写期间会同时存在：原包 + 解出来的文件 + 新包，最坏要两份包的空间。
		if ferr == nil {
			need := plan.KeepSz + plan.PackSz + (64 << 20)
			if free < need {
				fmt.Printf(M.CompactNoSpace, plan.Pack, humanSize(free), humanSize(need))
				failed++
				continue
			}
		}
		if len(plan.Keep) == 0 {
			fmt.Printf(M.CompactEmptyDoing, plan.Pack, placeholderName)
		} else {
			fmt.Printf(M.CompactDoing, plan.Pack, len(plan.Drop), humanSize(plan.DropSz))
		}
		res, rerr := rewritePack(repo, set, plan, password, noVerify)
		if rerr != nil {
			fmt.Printf(M.CompactPackFailed, plan.Pack, rerr)
			logf("compact %s: %s failed: %v", name, plan.Pack, rerr)
			failed++
			continue
		}
		if err := writeManifest(repo.Output, name, time.Now().Format("2006-01-02_15-04-05"), batchMeta{
			Packs:  []string{res.Pack},
			Hashes: []string{res.Hash},
			Hash:   res.Hash,
			From:   res.From,
			To:     res.To,
			Op:     "compact",
		}); err != nil {
			fmt.Printf(M.CompactPackFailed, plan.Pack, err)
			logf("compact %s: manifest update failed for %s: %v", name, plan.Pack, err)
			failed++
			continue
		}
		// 内容日期跟着剩下的东西走；空了就保留原范围（它仍然代表那个时间段）
		if res.From != "" {
			for i := range idx.Packs {
				if idx.Packs[i].Name == plan.Pack {
					idx.Packs[i].From, idx.Packs[i].To = res.From, res.To
					break
				}
			}
		}
		fmt.Printf(M.CompactPackDone, res.Pack, humanSize(res.OldSize), humanSize(res.NewSize))
		logf("compact %s: %s rewritten, %s -> %s, dropped %d file(s) / %s",
			name, res.Pack, humanSize(res.OldSize), humanSize(res.NewSize),
			len(plan.Drop), humanSize(plan.DropSz))
		done = append(done, res)
		oldTotal += res.OldSize
		newTotal += res.NewSize
		free -= (plan.KeepSz + plan.PackSz)
	}

	if len(done) > 0 {
		if err := saveIndex(repo.Output, idx); err != nil {
			return err
		}
	}
	fmt.Printf(M.CompactDoneLine, len(done), failed, humanSize(oldTotal), humanSize(newTotal), humanSize(oldTotal-newTotal))
	logf("compact %s: %d packs rewritten, %d failed, %s -> %s, saved %s",
		name, len(done), failed, humanSize(oldTotal), humanSize(newTotal), humanSize(oldTotal-newTotal))
	if len(done) > 0 {
		fmt.Print(M.CompactUploadHint)
	}
	// 通知只在真的干了活时才发：dry-run 和「没东西可压」都不该弹窗
	if notify && (len(done) > 0 || failed > 0) {
		body := fmt.Sprintf(M.ToastCompactBody, name, len(done), humanSize(oldTotal-newTotal))
		if failed > 0 {
			body += "\n" + fmt.Sprintf(M.ToastCompactFailBody, name, failed)
		}
		if err := sendToast(M.ToastCompactTitle, body); err != nil {
			fmt.Printf(M.NotifyFailed, err)
		}
	}
	if failed > 0 {
		return fmt.Errorf(M.CompactFailedMsg, failed)
	}
	return nil
}

// placeholderPathIn 返回恢复目录里占位文件的位置
func placeholderPathIn(dest string) string {
	return filepath.Join(dest, path.Base(placeholderName))
}

// removePlaceholder 清掉恢复出来的占位文件：只认「名字完全一致且 0 字节」，
// 万一用户真有一个同名文件（内容不为空），不动它。
func removePlaceholder(dest string) (int, error) {
	p := placeholderPathIn(dest)
	st, err := os.Lstat(p)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if st.IsDir() || st.Size() != 0 {
		return 0, nil
	}
	if err := os.Remove(p); err != nil {
		return 0, err
	}
	return 1, nil
}
