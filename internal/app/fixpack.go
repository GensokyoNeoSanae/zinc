package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// fixpack 修的是「包」，fixmeta 修的是「账」。这里只关心包内容：
// 找出坏掉的包、列出坏掉的文件、给每个文件找一份可用的替代来源、重打包。
//
// 一条重要的边界：非固实包坏了不等于全废。7z t 只会报出真正坏掉的文件，
// 其余文件仍然能正常提取，所以修复是「捞回能捞的」，不是「整包放弃」。

// badFile 是一个等待替换的文件
type badFile struct {
	rel    string // 包内相对路径
	source string // 人类可读的来源说明（源目录 / 历史包名）
	path   string // 实际要读的文件路径
	found  bool   // 是否找到可用的替代
}

// parseBadFiles 从 `7z t` 的输出里挑出坏掉的文件名。
// 7-Zip 的报错形如：ERROR: CRC Failed in encrypted file. Wrong password? : f1.txt
// Windows 文件名不允许冒号，所以取最后一个 " : " 之后的部分是安全的。
func parseBadFiles(out string) []string {
	var bad []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "ERROR:") {
			continue
		}
		i := strings.LastIndex(line, " : ")
		if i < 0 {
			continue
		}
		name := strings.TrimSpace(line[i+3:])
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		bad = append(bad, name)
	}
	return bad
}

// packTestResult 是 `7z t` 的一次结果，带缓存：同一个包在一次 fixpack 里只测一次
type packTestResult struct {
	ok      bool
	badFile []string
}

func testPackCached(sz, pack, password string, cache map[string]packTestResult) packTestResult {
	if r, hit := cache[pack]; hit {
		return r
	}
	cmd := hiddenCmd(sz, "t", "-sccUTF-8", "-p"+password, pack)
	out, err := cmd.CombinedOutput()
	r := packTestResult{ok: err == nil}
	if err != nil {
		r.badFile = parseBadFiles(string(out))
	}
	cache[pack] = r
	return r
}

// srcFileMatches 判断源目录里的文件是否就是 index 里记的那一份。
// 只看大小和 mtime——zinc 没有内容哈希可比，这两个是唯一能用的证据。
func srcFileMatches(repo *Repo, idx *Index, rel string) (exists, same bool) {
	st, err := os.Stat(filepath.Join(repo.Source, filepath.FromSlash(rel)))
	if err != nil {
		return false, false
	}
	e, ok := idx.Files[rel]
	if !ok {
		return true, false
	}
	if e.Size != st.Size() {
		return true, false
	}
	// mtime 允许 2 秒误差：不同文件系统的时间精度不一样
	diff := st.ModTime().Sub(e.MTime)
	if diff < 0 {
		diff = -diff
	}
	return true, diff <= 2*time.Second
}

// findReplacement 按优先级给一个坏文件找替代来源：
//  1. 源目录（当前数据，首选）
//  2. 其他包里没坏的那一份，从新往旧找（用户要「默认用最新的」）
//
// 每个候选包都先跑一次 7z t 并且缓存结果：包整体通过说明它里面的文件都好；
// 包整体失败时才看目标文件在不在坏名单里。
func findReplacement(repo *Repo, set *Settings, idx *Index, packs []string, rel, skipPack string,
	password string, cache map[string]packTestResult) (src, desc string, ok bool) {
	sz := repo.sevenZip(set)

	if exists, same := srcFileMatches(repo, idx, rel); exists {
		if same {
			return filepath.Join(repo.Source, filepath.FromSlash(rel)), M.FixSrcSame, true
		}
		return filepath.Join(repo.Source, filepath.FromSlash(rel)), M.FixSrcChanged, true
	}

	// 源目录没有：翻历史包，从新到旧
	for i := len(packs) - 1; i >= 0; i-- {
		p := packs[i]
		if p == skipPack {
			continue
		}
		r := testPackCached(sz, filepath.Join(repo.Output, p), password, cache)
		if !r.ok {
			// 这个包整体是坏的，看看目标文件是不是也坏在里面
			bad := false
			for _, b := range r.badFile {
				if b == rel {
					bad = true
					break
				}
			}
			if bad {
				continue
			}
		}
		return filepath.Join(repo.Output, p), fmt.Sprintf(M.FixFromPack, p), true
	}
	return "", "", false
}

// cmdFixpack 入口：默认预演，--check 以外的写法都会真的动手
func cmdFixpack(args []string) {
	fs := flag.NewFlagSet("fixpack", flag.ExitOnError)
	password := fs.String("password", "", M.FlagPassword)
	check := fs.Bool("check", false, M.FlagCheck)
	fs.Parse(reorderArgs(args))
	requireArgs(fs, 1, 1, M.UsageFixpack)
	name := fs.Arg(0)

	cfg := mustConfig()
	set := mustSettings()
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}
	pw := *password
	if pw == "" {
		pw = repo.password(set)
	}
	if pw == "" {
		pw = promptPassword(false)
	}
	if err := runFixpack(name, cfg, set, pw, *check); err != nil {
		die("%v", err)
	}
}

// runFixpack 的骨架：先找出坏包和坏文件，再决定能不能修、怎么修
func runFixpack(name string, cfg *Config, set *Settings, password string, check bool) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	lock, lerr := acquireLock(repo.Output, set.lockTimeout())
	if lerr != nil {
		return lerr
	}
	defer lock.release()

	packs, err := sortedPacks(repo.Output)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		return fmt.Errorf(M.FixNoPacks, repo.Output)
	}
	idx, err := loadIndex(repo.Output)
	if err != nil {
		return err
	}

	sz := repo.sevenZip(set)
	if sz == "" {
		return fmt.Errorf("%s", M.SevenZipMissing)
	}

	cache := make(map[string]packTestResult)
	badPacks := make([]string, 0)
	byPack := make(map[string][]badFile, 0)

	for _, p := range packs {
		full := filepath.Join(repo.Output, p)
		printColored(M.FixTesting, p)
		r := testPackCached(sz, full, password, cache)
		if r.ok {
			printColored(M.FixPackOK, p)
			continue
		}
		badPacks = append(badPacks, p)
		printColored(M.FixPackBad, p, len(r.badFile))
		files := make([]badFile, 0, len(r.badFile))
		for _, rel := range r.badFile {
			printColored(M.FixLooking, rel)
			src, desc, ok := findReplacement(repo, set, idx, packs, rel, p, password, cache)
			if ok {
				printColored(M.FixFound, rel, desc)
				files = append(files, badFile{rel: rel, source: desc, path: src, found: true})
			} else {
				printColored(M.FixNotFound, rel)
				files = append(files, badFile{rel: rel, found: false})
			}
		}
		byPack[p] = files
	}

	if len(badPacks) == 0 {
		printColored(M.FixAllOK, name)
		return nil
	}

	printColored(M.FixSummary, len(badPacks), len(packs))
	if check {
		printColored("%s", M.FixDryRun)
		return nil
	}
	if !askConfirm(set, fmt.Sprintf(M.FixConfirm, len(badPacks))) {
		fmt.Print(M.ConfirmCancelled)
		return nil
	}

	// 真正重写：逐个坏包，把能提取的文件提出来、补上替代来源、重新打包
	for _, p := range badPacks {
		files := byPack[p]
		repairable := make([]badFile, 0, len(files))
		unrecoverable := make([]badFile, 0)
		for _, f := range files {
			if f.found {
				repairable = append(repairable, f)
			} else {
				unrecoverable = append(unrecoverable, f)
			}
		}
		if len(repairable) == 0 {
			printColored(M.FixPackSkip, p, len(unrecoverable))
			continue
		}
		if err := rewritePackFixed(repo, set, p, repairable, password); err != nil {
			printColored(M.FixPackFail, p, err)
			continue
		}
		// 包字节变了，账本里的哈希必须跟着变，否则 verify 会一直报「更改」
		h, herr := sha256File(filepath.Join(repo.Output, p))
		if herr != nil {
			printColored(M.FixPackFail, p, herr)
			continue
		}
		if uerr := updateManifestPackHash(repo.Output, name, p, h); uerr != nil {
			printColored(M.FixPackFail, p, uerr)
			continue
		}
		printColored(M.FixPackDone, p)
		for _, f := range unrecoverable {
			printColored(M.FixStillLost, f.rel)
		}
	}

	// 包变了，账本要跟上
	printColored(M.FixSyncing, name)
	if err := runRemanifest(name, cfg, set, password, true); err != nil {
		return err
	}
	if err := runReindex(name, cfg, set, password, "", "", true); err != nil {
		return err
	}
	printColored(M.FixDone, name)
	return nil
}

// rewritePackFixed 把包解出来、用替代来源覆盖坏文件、再打回去。
// 包名不变，跟 compact 的重写策略一致（网盘同名覆盖才认得出是更新）。
func rewritePackFixed(repo *Repo, set *Settings, pack string, fixed []badFile, password string) error {
	sz := repo.sevenZip(set)
	tmp, err := os.MkdirTemp("", "zinc-fixpack-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// 解压：坏文件会报错，但其余文件照样落盘，这里只关心能出来的那些
	cmd := hiddenCmd(sz, "x", "-y", "-sccUTF-8", "-p"+password, filepath.Join(repo.Output, pack),
		"-o"+tmp, "-x!"+placeholderName)
	if _, xerr := cmd.CombinedOutput(); xerr != nil {
		printColored(M.FixExtractWarn, pack)
	}

	for _, f := range fixed {
		dst := filepath.Join(tmp, filepath.FromSlash(f.rel))
		if mkErr := os.MkdirAll(filepath.Dir(dst), 0o755); mkErr != nil {
			return mkErr
		}
		b, rerr := os.ReadFile(f.path)
		if rerr != nil {
			return fmt.Errorf(M.FixReadFail, f.rel, rerr)
		}
		if werr := os.WriteFile(dst, b, 0o644); werr != nil {
			return werr
		}
	}

	// 打包 tmp 目录下的全部内容（原包的好文件 + 刚补进来的替代文件）。
	// 这里刻意不走 packArgs 的 listfile：那份清单只列坏文件，会打出一个
	// 只剩坏文件的新包。
	part := filepath.Join(repo.Output, pack+".part")
	level := repo.compressLevel(set)
	args := []string{
		"a", "-y", "-t7z",
		"-p" + password,
		"-mhe=" + onOff(repo.headerCrypt(set)),
		fmt.Sprintf("-mx=%d", level),
	}
	if level > 0 {
		args = append(args, solidArg(repo.solidBlock(set)))
	} else {
		args = append(args, "-ms=off")
	}
	args = append(args, "-scsUTF-16LE", "-sccUTF-8", part, ".")
	cmd2 := hiddenCmd(sz, args...)
	cmd2.Dir = tmp
	if out, err := cmd2.CombinedOutput(); err != nil {
		os.Remove(part)
		return fmt.Errorf(M.CompressFailed, err, string(out))
	}
	if err := verifyArchive(sz, part, password); err != nil {
		os.Remove(part)
		return fmt.Errorf(M.VerifyFailed, err)
	}
	if err := syncFile(part); err != nil {
		os.Remove(part)
		return err
	}
	if err := os.Rename(part, filepath.Join(repo.Output, pack)); err != nil {
		os.Remove(part)
		return err
	}
	return nil
}
