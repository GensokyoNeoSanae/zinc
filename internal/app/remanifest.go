package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// batchTimeOfEntry 把包名或删除记录名转成 manifest 的 Time 形式，
// 与 writeManifest 用的 tag（YYYY-MM-DD_HH-MM-SS）保持一致。
func batchTimeOfEntry(name string) string {
	stem := strings.TrimSuffix(strings.TrimSuffix(name, ".7z"), delExt)
	if t, ok := parsePackTime(stem + ".7z"); ok {
		return t.Format("2006-01-02_15-04-05")
	}
	return ""
}

// packDateRange 从归档头算内容日期范围；全是占位文件时 content=false。
func packDateRange(entries []packEntry) (from, to string, content bool) {
	var minT, maxT time.Time
	for _, e := range entries {
		if e.IsDir || e.Path == "" || isPlaceholder(e.Path) {
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
		return "", "", false
	}
	return minT.Format("20060102"), maxT.Format("20060102"), true
}

// manifestPackCount 账本里登记了多少个包（含尚未清理的已消失项）。
func manifestPackCount(batches []ManifestBatch) int {
	n := 0
	for _, b := range batches {
		n += len(b.Packs)
	}
	return n
}

// runRemanifest 对账 manifest 与磁盘：删掉盘上已不存在的包/删除记录、补登孤儿项。
// 批次归属不可推导，所以保留现有批次结构，只做加减；默认预演，--yes 才写回。
func runRemanifest(name string, cfg *Config, set *Settings, password string, yes bool) error {
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
	rms, err := listDelFiles(repo.Output)
	if err != nil {
		return err
	}
	onDiskPack := make(map[string]bool, len(packs))
	for _, p := range packs {
		onDiskPack[p] = true
	}
	onDiskRm := make(map[string]bool, len(rms))
	for _, r := range rms {
		onDiskRm[r] = true
	}

	m := &Manifest{Version: manifestVersion, Repo: name}
	existed := false
	if b, rerr := os.ReadFile(manifestPath(repo.Output)); rerr == nil {
		existed = true
		if m, err = decodeManifest(b); err != nil {
			return fmt.Errorf(M.ManifestBroken, manifestPath(repo.Output), err)
		}
		m.Repo = name
	} else if !os.IsNotExist(rerr) {
		return fmt.Errorf(M.ManifestReadFailed, rerr)
	}
	if !existed {
		fmt.Printf(M.RemanifestNoManifest, manifestPath(repo.Output))
	}
	fmt.Printf(M.RemanifestHeader, name, len(m.Batches), manifestPackCount(m.Batches), len(packs), len(rms))

	listedPacks := make(map[string]bool, len(packs))
	listedRms := make(map[string]bool, len(rms))
	var batches []ManifestBatch
	dropped, added := 0, 0

	// 1) 现有批次：去掉盘上已消失的包/删除记录；整批空的就删掉
	for _, b := range m.Batches {
		nb := ManifestBatch{Time: b.Time}
		for f, h := range b.Packs {
			if !onDiskPack[f] {
				fmt.Printf(M.RemanifestDropPack, f)
				dropped++
				continue
			}
			if nb.Packs == nil {
				nb.Packs = map[string]string{}
			}
			nb.Packs[f] = h
			listedPacks[f] = true
		}
		for rm, h := range b.Rms {
			if !onDiskRm[rm] {
				fmt.Printf(M.RemanifestDropRm, rm)
				dropped++
				continue
			}
			if nb.Rms == nil {
				nb.Rms = map[string]string{}
			}
			nb.Rms[rm] = h
			listedRms[rm] = true
		}
		if len(nb.Packs) == 0 && len(nb.Rms) == 0 {
			continue
		}
		batches = append(batches, nb)
	}

	// 2) 孤儿包：补一批。一包一批——批次归属本来就没法从磁盘还原。
	for _, p := range packs {
		if listedPacks[p] {
			continue
		}
		if !yes {
			// 预演不读包内容（也就不需要密码）；占位包要到执行时才认得出来
			fmt.Printf(M.RemanifestAddPack, p)
			added++
			continue
		}
		entries, lerr := listPackContents(repo.sevenZip(set), filepath.Join(repo.Output, p), password)
		if lerr != nil {
			return fmt.Errorf(M.RemanifestListFailed, p, lerr)
		}
		if _, _, content := packDateRange(entries); !content {
			fmt.Printf(M.RemanifestSkipPlaceholder, p)
			continue
		}
		h, herr := sha256File(filepath.Join(repo.Output, p))
		if herr != nil {
			return herr
		}
		nb := ManifestBatch{Time: batchTimeOfEntry(p), Packs: map[string]string{p: h}}
		// 同名 .rm 跟着这个包走：delRecordName 就是把记录挂在批次最后一个包上
		rm := strings.TrimSuffix(p, ".7z") + delExt
		if onDiskRm[rm] && !listedRms[rm] {
			rh, rerr := sha256File(delPath(repo.Output, rm))
			if rerr != nil {
				return rerr
			}
			nb.Rms = map[string]string{rm: rh}
			listedRms[rm] = true
		}
		fmt.Printf(M.RemanifestAddPack, p)
		added++
		batches = append(batches, nb)
	}

	// 3) 没挂到包上的孤儿删除记录
	for _, rm := range rms {
		if listedRms[rm] {
			continue
		}
		if !yes {
			fmt.Printf(M.RemanifestAddRm, rm)
			added++
			continue
		}
		rh, rerr := sha256File(delPath(repo.Output, rm))
		if rerr != nil {
			return rerr
		}
		fmt.Printf(M.RemanifestAddRm, rm)
		added++
		batches = append(batches, ManifestBatch{
			Time: batchTimeOfEntry(rm),
			Rms:  map[string]string{rm: rh},
		})
	}

	if dropped == 0 && added == 0 {
		fmt.Print(M.RemanifestNoChange)
		return nil
	}
	if !yes {
		fmt.Print(M.RemanifestPreview)
		return nil
	}

	// 时间升序：status 的「最后备份」取最后一批，补登的旧包不能排到后面去
	sort.SliceStable(batches, func(i, j int) bool { return batches[i].Time < batches[j].Time })
	if existed {
		old, rerr := os.ReadFile(manifestPath(repo.Output))
		if rerr != nil {
			return fmt.Errorf(M.RemanifestBakFailed, rerr)
		}
		if werr := writeFileSync(manifestPath(repo.Output)+".bak", old, 0o644); werr != nil {
			return fmt.Errorf(M.RemanifestBakFailed, werr)
		}
	}
	if err := writeManifestFile(repo.Output, &Manifest{Version: manifestVersion, Repo: name, Batches: batches}); err != nil {
		return err
	}
	fmt.Printf(M.RemanifestDone, dropped, added, manifestPath(repo.Output))
	logf("remanifest %s: dropped=%d added=%d batches=%d", name, dropped, added, len(batches))
	return nil
}

// cmdRemanifest 入口：默认预演，--yes 写回。
func cmdRemanifest(args []string) {
	cfg := mustConfig()
	set := mustSettings()
	fs := flag.NewFlagSet("remanifest", flag.ExitOnError)
	password := fs.String("password", "", M.FlagPassword)
	yes := fs.Bool("yes", false, M.FlagYesRemanifest)
	fs.Parse(reorderArgs(args))
	requireArgs(fs, 1, 1, M.UsageRemanifest)
	name := fs.Arg(0)
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}
	pw := *password
	if pw == "" {
		pw = repo.password(set)
	}
	// 预演不读包内容，不需要密码；执行时才要（列归档头认占位包）
	if *yes && pw == "" {
		pw = promptPassword(false)
	}
	if err := runRemanifest(name, cfg, set, pw, *yes); err != nil {
		die("%v", err)
	}
}
