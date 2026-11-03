package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// cmdRepair 是元数据修复的统一入口。原先的 reindex 和 remanifest 是同一个
// 问题的两半——「账本/索引」跟磁盘对不上——用户不该记住哪个文件坏了。
//
// 顺序是先对账 manifest（便宜，只 stat），再重建索引（贵，要读每个包的目录），
// 最后按需重算包哈希（最贵，要读整个包）。
func cmdRepair(args []string) {
	fs := flag.NewFlagSet("repair", flag.ExitOnError)
	password := fs.String("password", "", M.FlagPassword)
	packStart := fs.String("pack-start", "", M.FlagPackStart)
	packEnd := fs.String("pack-end", "", M.FlagPackEnd)
	rehash := fs.Bool("rehash", false, M.FlagRehash)
	check := fs.Bool("check", false, M.FlagCheck)
	fs.Parse(reorderArgs(args))
	requireArgs(fs, 1, 1, M.UsageRepair)
	name := fs.Arg(0)

	cfg := mustConfig()
	set := mustSettings()
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}

	// 预演不读包内容，也就不需要密码；真正执行时要（列归档头才认得出占位包）。
	pw := *password
	if pw == "" {
		pw = repo.password(set)
	}
	if !*check && pw == "" {
		pw = promptPassword(false)
	}

	if !*check && !askConfirm(set, M.RepairConfirmHead) {
		fmt.Print(M.ConfirmCancelled)
		return
	}

	// 1) manifest 对账：盘上没了的删登记、盘上多出来的补登
	if err := runRemanifest(name, cfg, set, pw, !*check); err != nil {
		die("%v", err)
	}
	// 2) index.tsv 重建：从包目录重建「文件 → 当前状态」
	if err := runReindex(name, cfg, set, pw, *packStart, *packEnd, !*check); err != nil {
		die("%v", err)
	}
	// 3) 可选：按磁盘实际内容重算包哈希
	if *rehash {
		if err := runRehash(name, cfg, set, *check); err != nil {
			die("%v", err)
		}
	}
}

// runRehash 按磁盘上的实际内容重算 manifest 里每个包的哈希。
// compact 在「包已替换、manifest 没写成」时会留下旧哈希，verify 会一直
// 报损坏；这是唯一能把它修回来的路径。
func runRehash(name string, cfg *Config, set *Settings, check bool) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	p := manifestPath(repo.Output)
	b, rerr := os.ReadFile(p)
	if os.IsNotExist(rerr) {
		fmt.Printf(M.RemanifestNoManifest, p)
		return nil
	}
	if rerr != nil {
		return fmt.Errorf(M.ManifestReadFailed, rerr)
	}
	m, err := decodeManifest(b)
	if err != nil {
		return fmt.Errorf(M.ManifestBroken, p, err)
	}

	changed := 0
	for i := range m.Batches {
		for pack := range m.Batches[i].Packs {
			full := filepath.Join(repo.Output, pack)
			h, herr := sha256File(full)
			if herr != nil {
				// 包不在盘上：remanifest 那一步已经处理过，这里跳过
				continue
			}
			if m.Batches[i].Packs[pack] != h {
				fmt.Printf(M.RepairRehashPack, pack)
				m.Batches[i].Packs[pack] = h
				changed++
			}
		}
	}

	if changed == 0 {
		fmt.Print(M.RepairRehashNone)
		return nil
	}
	if check {
		fmt.Printf(M.RepairRehashDryRun, changed)
		return nil
	}
	lock, lerr := acquireLock(repo.Output, set.lockTimeout())
	if lerr != nil {
		return lerr
	}
	defer lock.release()
	if err := writeManifestFile(repo.Output, m); err != nil {
		return err
	}
	fmt.Printf(M.RepairRehashDone, changed)
	return nil
}
