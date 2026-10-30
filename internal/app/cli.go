package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	fmt.Print(M.Usage)
}

var flagNeedsValue = map[string]bool{
	"source": true, "output": true, "dest": true, "password": true,
	"sample": true, "at": true, "weekly": true, "monthly": true,
	"max-pack-byte": true, "compress": true, "solid": true, "exclude": true,
	"pack-start": true, "pack-end": true, "only": true, "from": true,
	"7zip": true, "pack": true, "min-gain": true,
}

// multiFlag 是可重复的字符串 flag（--exclude a --exclude b）
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

// reorderArgs 把 flags 移到位置参数前面。
// Go 的 flag 包遇到第一个非 flag 参数就停止解析，导致 "add 名字 --source x" 失效。
func reorderArgs(args []string) []string {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 0 && a[0] == '-' {
			flags = append(flags, a)
			name := strings.TrimLeft(a, "-")
			if strings.Contains(name, "=") {
				continue
			}
			if flagNeedsValue[name] && i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			pos = append(pos, a)
		}
	}
	return append(flags, pos...)
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, M.ErrPrefix+format+"\n", args...)
	os.Exit(1)
}

func mustConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		die("%v", err)
	}
	return cfg
}

func mustSettings() *Settings {
	set, err := loadSettings()
	if err != nil {
		die("%v", err)
	}
	return set
}

// pickPassword 按 命令行 > 仓库 > settings > 交互输入 的顺序取密码，
// 第二个返回值是来源，用于备份汇总行标注。
func pickPassword(flagVal, prompted string, repo *Repo, set *Settings) (string, string) {
	if flagVal != "" {
		return flagVal, "cli"
	}
	if repo != nil && repo.Password != "" {
		return repo.Password, "repo"
	}
	if set.DefaultPassword != "" {
		return set.DefaultPassword, "settings"
	}
	return prompted, "prompt"
}

// passwordMismatch 返回第一个「配置里已有密码、且与 flagPw 不同」的仓库名。
// 只看配置非空的仓库：配置里根本没密码时，--password 是唯一来源，不算冲突。
func passwordMismatch(cfg *Config, set *Settings, names []string, flagPw string) (string, bool) {
	for _, n := range names {
		r, err := cfg.repo(n)
		if err != nil {
			continue
		}
		if v := r.password(set); v != "" && v != flagPw {
			return n, true
		}
	}
	return "", false
}

// schedulePassword 解析计划任务要嵌入的密码：仓库 > settings。
// 返回来源（repo/settings）；没有可用密码返回空串；目标仓库密码不一致返回 "mixed"。
// 一条任务只能带一个密码，所以 --all 下必须所有仓库解析出同一个值。
func schedulePassword(cfg *Config, set *Settings, names []string) (string, string) {
	pw, from := "", ""
	for _, n := range names {
		r, err := cfg.repo(n)
		if err != nil {
			continue
		}
		v := r.password(set)
		if v == "" {
			continue
		}
		if pw == "" {
			pw = v
			from = "settings"
			if r.Password != "" {
				from = "repo"
			}
			continue
		}
		if v != pw {
			return "", "mixed"
		}
	}
	return pw, from
}

func promptPassword(twice bool) string {
	fmt.Print(M.PromptPassword)
	p := readLine()
	if p == "" {
		die(M.PasswordEmpty)
	}
	if !twice {
		return p
	}
	for i := 0; i < 3; i++ {
		fmt.Print(M.PromptPasswordAgain)
		p2 := readLine()
		if p == p2 {
			return p
		}
		fmt.Println(M.PasswordMismatch)
		fmt.Print(M.PromptPassword)
		p = readLine()
	}
	die(M.PasswordTooMany)
	return ""
}

func cmdRoot(args []string) {
	cfg := mustConfig()
	if len(args) == 0 {
		if cfg.Root == "" {
			fmt.Println(M.RootUnset)
		} else {
			fmt.Println(cfg.Root)
		}
		return
	}
	p := args[0]
	if strings.HasPrefix(p, "-") && !strings.ContainsAny(p, `\/`) {
		die(M.UsageRoot)
	}
	abs, perr := absClean(p)
	if perr != nil {
		die(M.PathInvalid, p, perr)
	}
	if st, statErr := os.Stat(abs); statErr == nil && !st.IsDir() {
		die(M.RootNotDir, abs)
	}
	cfg.Root = abs
	if err := saveConfig(cfg); err != nil {
		die(M.SaveConfigFailed, err)
	}
	fmt.Printf("root = %s\n", cfg.Root)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	names := fs.Bool("names", false, M.FlagListNames)
	fs.Parse(reorderArgs(args))
	if fs.NArg() > 0 {
		die(M.UsageList)
	}
	cfg := mustConfig()
	if *names {
		// 机器可读：一行一个名字，不缩进、不带路径，补全脚本直接拿
		for _, n := range cfg.sortedNames() {
			fmt.Println(n)
		}
		return
	}
	if len(cfg.Repos) == 0 {
		fmt.Println(M.NoRepos)
		return
	}
	for _, n := range cfg.sortedNames() {
		r := cfg.Repos[n]
		fmt.Printf(M.ListItem, n, r.Source, r.Output)
	}
}

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	source := fs.String("source", "", "source directory")
	output := fs.String("output", "", "output directory")
	fs.Parse(reorderArgs(args))
	if fs.NArg() < 1 {
		die(M.UsageAdd)
	}
	name := fs.Arg(0)
	if *source == "" {
		die(M.MissingSource)
	}
	cfg := mustConfig()
	if _, exists := cfg.Repos[name]; exists {
		die(M.RepoExists, name)
	}
	src, perr := absClean(*source)
	if perr != nil {
		die(M.PathInvalid, *source, perr)
	}
	out := *output
	if out == "" {
		if cfg.Root == "" {
			die(M.RootNotSet)
		}
		out = filepath.Join(cfg.Root, name)
	}
	outAbs, perr := absClean(out)
	if perr != nil {
		die(M.PathInvalid, out, perr)
	}
	if err := validateRepoPaths(src, outAbs); err != nil {
		die("%v", err)
	}
	if owner, ok := cfg.outputOwner("", outAbs); ok {
		die(M.OutputInUse, owner, outAbs)
	}
	if owner, ok := cfg.outputOwnerFold("", outAbs); ok {
		fmt.Printf(M.OutputCaseWarn, outAbs, owner)
	}
	if _, err := os.Stat(src); err != nil {
		fmt.Printf(M.WarnSourceMissing, src)
	}
	cfg.Repos[name] = &Repo{Source: src, Output: outAbs}
	if err := saveConfig(cfg); err != nil {
		die(M.SaveConfigFailed, err)
	}
	fmt.Printf(M.AddedRepo, name, src, outAbs)
}

func cmdRemove(args []string) {
	if len(args) < 1 {
		die(M.UsageRemove)
	}
	name := args[0]
	cfg := mustConfig()
	if _, exists := cfg.Repos[name]; !exists {
		die(M.RepoNotFound, name)
	}
	delete(cfg.Repos, name)
	if err := saveConfig(cfg); err != nil {
		die(M.SaveConfigFailed, err)
	}
	fmt.Printf(M.RemovedRepo, name)
	if tasks, err := findScheduledTasks("zinc-" + name + "-"); err == nil && len(tasks) > 0 {
		fmt.Printf(M.RemoveHasTasks, name, len(tasks), strings.Join(tasks, ", "), name)
	}
}

func cmdSet(args []string) {
	fs := flag.NewFlagSet("set", flag.ExitOnError)
	source := fs.String("source", "", "new source directory")
	output := fs.String("output", "", "new output directory")
	password := fs.String("password", "", M.FlagPassword)
	maxPack := fs.String("max-pack-byte", "", M.FlagMaxPackByte)
	compress := fs.Int("compress", -1, M.FlagCompress)
	solid := fs.String("solid", "", M.FlagSolid)
	var excludes multiFlag
	fs.Var(&excludes, "exclude", M.FlagExclude)
	fs.Parse(reorderArgs(args))
	if fs.NArg() < 1 {
		die(M.UsageSet)
	}
	name := fs.Arg(0)
	cfg := mustConfig()
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}
	if *source == "" && *output == "" && *password == "" && *maxPack == "" && *compress < 0 && *solid == "" && len(excludes) == 0 {
		die(M.NeedSourceOrOutput)
	}
	if *source != "" {
		s, perr := absClean(*source)
		if perr != nil {
			die(M.PathInvalid, *source, perr)
		}
		repo.Source = s
	}
	if *output != "" {
		o, perr := absClean(*output)
		if perr != nil {
			die(M.PathInvalid, *output, perr)
		}
		repo.Output = o
	}
	if *source != "" || *output != "" {
		if err := validateRepoPaths(repo.Source, repo.Output); err != nil {
			die("%v", err)
		}
	}
	if *output != "" {
		if owner, ok := cfg.outputOwner(name, repo.Output); ok {
			die(M.OutputInUse, owner, repo.Output)
		}
		if owner, ok := cfg.outputOwnerFold(name, repo.Output); ok {
			fmt.Printf(M.OutputCaseWarn, repo.Output, owner)
		}
	}
	if *password != "" {
		repo.Password = *password
	}
	if *maxPack != "" {
		n, perr := parseSize(*maxPack)
		if perr != nil {
			die(M.ConfigBadValue, "max_pack_byte", *maxPack)
		}
		repo.MaxPackByte = n
	}
	if *compress >= 0 {
		if *compress > 9 {
			die(M.ConfigBadValue, "compression_level", *compress)
		}
		v := *compress
		repo.CompressLevel = &v
	}
	if *solid != "" {
		n, perr := parseSolid(*solid)
		if perr != nil {
			die(M.ConfigBadValue, "solid_block", *solid)
		}
		repo.SolidBlock = &n
	}
	if len(excludes) > 0 {
		repo.Exclude = append(repo.Exclude, excludes...)
	}
	if err := saveConfig(cfg); err != nil {
		die(M.SaveConfigFailed, err)
	}
	fmt.Printf(M.UpdatedRepo, name, repo.Source, repo.Output)
}

func cmdBackup(args []string) {
	cfg := mustConfig()
	set := mustSettings()

	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	all := fs.Bool("all", false, M.FlagAllBackup)
	password := fs.String("password", "", M.FlagPassword)
	notify := fs.Bool("notify", set.completeNotify(), M.FlagNotify)
	notifyEmpty := fs.Bool("notify-empty", false, M.FlagNotifyEmpty)
	noVerify := fs.Bool("no-verify", false, M.FlagNoVerify)
	fs.Parse(reorderArgs(args))

	var names []string
	if *all {
		names = cfg.sortedNames()
	} else {
		if fs.NArg() < 1 {
			die(M.UsageBackup)
		}
		names = []string{fs.Arg(0)}
	}
	if len(names) == 0 {
		die(M.NoReposBackup)
	}

	// --all + 显式 --password：与某个仓库配置的密码不一致时先拦一下。
	// 计划任务没有控制台（isConsole 为假），读不到回答就按取消处理——宁可响亮失败，也不要静默用错密码加密。
	if *all && *password != "" {
		if rn, ok := passwordMismatch(cfg, set, names, *password); ok {
			fmt.Printf(M.BackupPwMismatch, rn)
			ans := ""
			if isConsole() {
				fmt.Print(M.ConfirmPwMismatch)
				ans = strings.ToLower(strings.TrimSpace(readLine()))
			}
			if ans != "y" && ans != "yes" {
				logf("backup --all: cancelled, --password differs from repo/default password")
				if *notify {
					if err := sendToast(M.ToastTitle, M.BackupPwAborted); err != nil {
						fmt.Printf(M.NotifyFailed, err)
					}
				}
				die(M.BackupPwAborted)
			}
		}
	}

	// 先扫描，全部无变化时不需要密码
	scans := make(map[string]ScanResult)
	var scanErrs []string
	anyChange := false
	for _, n := range names {
		repo, err := cfg.repo(n)
		if err != nil {
			scanErrs = append(scanErrs, fmt.Sprintf("%s · %v", n, err))
			continue
		}
		if _, statErr := os.Stat(repo.Source); statErr != nil {
			scanErrs = append(scanErrs, fmt.Sprintf(M.ScanErrSource, n))
			continue
		}
		s, err := scanRepo(repo, set)
		if err != nil {
			scanErrs = append(scanErrs, fmt.Sprintf(M.ScanErrGeneric, n, err))
			continue
		}
		scans[n] = s
		if s.Changed() > 0 {
			anyChange = true
		}
	}

	if !anyChange {
		for _, n := range names {
			if _, ok := scans[n]; ok {
				fmt.Printf(M.NoNew, n)
			}
		}
		for _, e := range scanErrs {
			fmt.Printf("%s\n", e)
		}
		if *notify && *notifyEmpty {
			var results []BackupResult
			for _, n := range names {
				results = append(results, BackupResult{Repo: n, SkippedAll: true})
			}
			sendSummary(results, true)
		}
		if len(scanErrs) > 0 {
			os.Exit(1)
		}
		return
	}

	// 只有真的要压包时才需要密码：只记录删除的批次不碰 7-Zip。
	needPrompt := false
	if *password == "" && set.DefaultPassword == "" {
		for _, n := range names {
			r, e := cfg.repo(n)
			if e != nil || r.Password != "" {
				continue
			}
			if s, ok := scans[n]; ok && s.Packable() > 0 {
				needPrompt = true
				break
			}
		}
	}
	prompted := ""
	if needPrompt {
		prompted = promptPassword(true)
	}

	var results []BackupResult
	failed := len(scanErrs) > 0
	for _, e := range scanErrs {
		fmt.Printf("%s\n", e)
	}
	for _, n := range names {
		s, ok := scans[n]
		if !ok {
			continue
		}
		repo, _ := cfg.repo(n)
		pw, encFrom := pickPassword(*password, prompted, repo, set)
		r := runBackup(n, cfg, set, s, pw, encFrom, *noVerify)
		if r.Err != nil {
			fmt.Printf(M.BackupFailed, n, r.Err)
			// 失败必须进日志：否则事后只知道「没备份成」，不知道死在哪一步、
			// 已提交了多少、还剩多少没压，全是黑盒。
			logf("backup %s: FAILED after %d packs (%s), %d files left, err=%v",
				n, len(r.Packs), humanSize(r.TotalSize), r.Skipped, r.Err)
			failed = true
		}
		results = append(results, r)
	}

	if *notify {
		sendSummary(results, *notifyEmpty)
	}
	if failed {
		os.Exit(1)
	}
}

func sendSummary(results []BackupResult, notifyEmpty bool) {
	var lines []string
	any := false
	for _, r := range results {
		if r.Err != nil {
			lines = append(lines, r.Repo+M.SummaryFailed)
			any = true
			continue
		}
		if r.Changed() == 0 {
			lines = append(lines, r.Repo+M.SummaryNoNew)
			continue
		}
		any = true
		lines = append(lines, fmt.Sprintf(M.SummaryLine,
			r.Repo, r.Changed(), humanSize(r.TotalSize), shortHash(r.BatchHash)))
	}
	if !any && !notifyEmpty {
		return
	}
	body := strings.Join(lines, "\n")
	if len(lines) > 8 {
		body = strings.Join(lines[:8], "\n") + fmt.Sprintf(M.SummaryMore, len(lines)-8)
	}
	if err := sendToast(M.ToastTitle, body); err != nil {
		fmt.Printf(M.NotifyFailed, err)
	}
}

func cmdRestore(args []string) {
	cfg := mustConfig()
	set := mustSettings()

	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	dest := fs.String("dest", "", M.FlagDest)
	password := fs.String("password", "", M.FlagPassword)
	overwrite := fs.Bool("overwrite", false, M.FlagOverwrite)
	noOverwrite := fs.Bool("no-overwrite", false, M.FlagNoOverwrite)
	checkList := fs.Bool("check-list", false, M.FlagCheckList)
	noDelete := fs.Bool("no-delete", false, M.FlagNoDelete)
	notify := fs.Bool("notify", set.completeNotify(), M.FlagNotifyWarn)
	only := fs.String("only", "", M.FlagOnly)
	from := fs.String("from", "", M.FlagFrom)
	fs.Parse(reorderArgs(args))
	if fs.NArg() < 1 {
		die(M.UsageRestore)
	}
	name := fs.Arg(0)
	if *overwrite && *noOverwrite {
		die(M.FlagConflict)
	}
	if *from != "" && *only == "" {
		die(M.FromRequiresOnly)
	}

	ow := set.restoreOverwrite()
	if *overwrite {
		ow = true
	}
	if *noOverwrite {
		ow = false
	}

	// 选择性恢复：只解压指定路径
	if *only != "" {
		// --check-list 也要密码：列包内文件需要读加密的归档头
		pw := restorePassword(name, cfg, set, *password)
		if *checkList {
			if err := runRestoreOnly(name, cfg, set, "", pw, *only, *from, ow, true); err != nil {
				die("%v", err)
			}
			return
		}
		if *dest == "" {
			die(M.MissingDest)
		}
		if err := runRestoreOnly(name, cfg, set, *dest, pw, *only, *from, ow, false); err != nil {
			die("%v", err)
		}
		return
	}

	opt := restoreOpts{
		Dest:      *dest,
		Overwrite: ow,
		CheckList: *checkList,
		Wrap:      set.restoreWrapFolder(),
		NoDelete:  *noDelete,
		Notify:    *notify,
	}

	if *checkList {
		if err := runRestore(name, cfg, set, opt); err != nil {
			die("%v", err)
		}
		return
	}
	if *dest == "" {
		die(M.MissingDest)
	}
	opt.Password = restorePassword(name, cfg, set, *password)

	if err := runRestore(name, cfg, set, opt); err != nil {
		die("%v", err)
	}
}

// restorePassword 解析恢复用的密码：命令行 > 仓库 > settings > 交互输入
func restorePassword(name string, cfg *Config, set *Settings, flagVal string) string {
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}
	pw, _ := pickPassword(flagVal, "", repo, set)
	if pw == "" {
		pw = promptPassword(false)
	}
	return pw
}

func cmdVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	all := fs.Bool("all", false, M.FlagAllVerify)
	sample := fs.Int("sample", 100, M.FlagSample)
	fs.Parse(reorderArgs(args))

	if *sample < 1 || *sample > 100 {
		die(M.SampleInvalid, *sample)
	}
	cfg := mustConfig()
	set := mustSettings()
	var names []string
	if *all {
		names = cfg.sortedNames()
	} else {
		if fs.NArg() < 1 {
			die(M.UsageVerify)
		}
		names = []string{fs.Arg(0)}
	}

	failed := false
	for _, n := range names {
		if err := runVerify(n, cfg, set, *sample); err != nil {
			fmt.Printf("%s: %v\n", n, err)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

func cmdStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	all := fs.Bool("all", false, M.FlagAllStatus)
	fs.Parse(reorderArgs(args))

	cfg := mustConfig()
	set := mustSettings()

	// --all：每个仓库都展开成详情，仓库之间空一行。
	if *all {
		if fs.NArg() > 0 {
			die(M.UsageStatus)
		}
		names := cfg.sortedNames()
		if len(names) == 0 {
			fmt.Println(M.NoRepos)
			return
		}
		var firstErr error
		for i, n := range names {
			if i > 0 {
				fmt.Println()
			}
			if err := runStatus(n, cfg, set); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if firstErr != nil {
			os.Exit(1)
		}
		return
	}

	name := ""
	if fs.NArg() > 0 {
		name = fs.Arg(0)
	}
	if err := runStatus(name, cfg, set); err != nil {
		os.Exit(1)
	}
}

func cmdConfig(args []string) {
	set, err := loadSettings()
	if err != nil {
		die("%v", err)
	}
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "", "list", "show":
		fmt.Printf(M.ConfigHeader, settingsPath())
		for _, k := range cfgKeys {
			fmt.Printf(M.ConfigItem, k.name, cfgValueString(set, k.name))
		}
	case "path":
		fmt.Printf(M.ConfigPath, settingsPath())
	case "get":
		if len(args) < 2 {
			die(M.UsageConfig)
		}
		if _, ok := findCfgKey(args[1]); !ok {
			die(M.ConfigUnknownKey, args[1])
		}
		fmt.Println(cfgValueRaw(set, args[1]))
	case "set":
		if len(args) < 3 {
			die(M.UsageConfig)
		}
		key, val := args[1], args[2]
		if _, ok := findCfgKey(key); !ok {
			die(M.ConfigUnknownKey, key)
		}
		if val == "-" {
			fmt.Printf(M.ConfigPromptValue, key)
			val = readLine()
		}
		if err := cfgSet(set, key, val); err != nil {
			die("%v", err)
		}
		if err := saveSettings(set); err != nil {
			die(M.SaveConfigFailed, err)
		}
		fmt.Printf(M.ConfigSet, key, cfgValueString(set, key))
	case "unset":
		if len(args) < 2 {
			die(M.UsageConfig)
		}
		if err := cfgUnset(set, args[1]); err != nil {
			die("%v", err)
		}
		if err := saveSettings(set); err != nil {
			die(M.SaveConfigFailed, err)
		}
		fmt.Printf(M.ConfigUnset, args[1])
	default:
		die(M.UsageConfig)
	}
}

// cmdCompact 重写已有包、丢掉被新版取代的旧副本。包名不变，云端同名覆盖。
func cmdCompact(args []string) {
	cfg := mustConfig()
	set := mustSettings()
	fs := flag.NewFlagSet("compact", flag.ExitOnError)
	password := fs.String("password", "", M.FlagPassword)
	packSpec := fs.String("pack", "", M.FlagCompactPack)
	minGain := fs.String("min-gain", "5%", M.FlagCompactMinGain)
	yes := fs.Bool("yes", false, M.FlagYesCompact)
	noVerify := fs.Bool("no-verify", false, M.FlagNoVerify)
	notify := fs.Bool("notify", set.completeNotify(), M.FlagNotifyCompact)
	fs.Parse(reorderArgs(args))
	if fs.NArg() < 1 {
		die(M.UsageCompact)
	}
	name := fs.Arg(0)
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}
	pw, _ := pickPassword(*password, "", repo, set)
	if pw == "" {
		pw = promptPassword(false)
	}
	if err := runCompact(name, cfg, set, pw, *packSpec, *minGain, *yes, *noVerify, *notify); err != nil {
		logf("compact %s: failed: %v", name, err)
		die("%v", err)
	}
}

func cmdReindex(args []string) {
	fs := flag.NewFlagSet("reindex", flag.ExitOnError)
	password := fs.String("password", "", M.FlagPassword)
	packStart := fs.String("pack-start", "", M.FlagPackStart)
	packEnd := fs.String("pack-end", "", M.FlagPackEnd)
	yes := fs.Bool("yes", false, M.FlagYesReindex)
	fs.Parse(reorderArgs(args))
	if fs.NArg() < 1 {
		die(M.UsageReindex)
	}
	name := fs.Arg(0)

	cfg := mustConfig()
	set := mustSettings()
	repo, err := cfg.repo(name)
	if err != nil {
		die("%v", err)
	}
	pw, _ := pickPassword(*password, "", repo, set)
	if pw == "" {
		pw = promptPassword(false)
	}
	if err := runReindex(name, cfg, set, pw, *packStart, *packEnd, *yes); err != nil {
		die("%v", err)
	}
}

func cmdSchedule(args []string) {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)
	list := fs.Bool("list", false, M.FlagListSchedule)
	all := fs.Bool("all", false, M.FlagAllSchedule)
	daily := fs.Bool("daily", false, M.FlagDaily)
	weekly := fs.String("weekly", "", M.FlagWeekly)
	monthly := fs.String("monthly", "", M.FlagMonthly)
	at := fs.String("at", "", M.FlagAt)
	password := fs.String("password", "", M.FlagPassword)
	noNotify := fs.Bool("no-notify", false, M.FlagNoNotify)
	yes := fs.Bool("yes", false, M.FlagYes)
	fs.Parse(reorderArgs(args))

	// --list 是只读查询，不需要仓库名、周期、时间，也不需要密码。
	if *list {
		if err := runListSchedules(); err != nil {
			die("%v", err)
		}
		return
	}

	var kind, days string
	switch {
	case *daily:
		kind = "daily"
	case *weekly != "":
		kind, days = "weekly", *weekly
	case *monthly != "":
		kind, days = "monthly", *monthly
	default:
		die(M.NeedPeriod)
	}
	if *at == "" {
		die(M.NeedAt)
	}

	cfg := mustConfig()
	set := mustSettings()
	name := ""
	var targets []string
	if *all {
		targets = cfg.sortedNames()
		if len(targets) == 0 {
			die(M.NoReposBackup)
		}
	} else {
		if fs.NArg() < 1 {
			die(M.UsageSchedule)
		}
		name = fs.Arg(0)
		if _, err := cfg.repo(name); err != nil {
			die("%v", err)
		}
		targets = []string{name}
	}

	// 密码按「命令行 > 仓库 > settings」取。计划任务会把密码明文存进任务的 /TR，
	// 所以这里绝不能再无条件交互询问——一次手滑就会把测试密码永久固化进任务。
	pw, pwFrom := *password, "cli"
	if pw == "" {
		pw, pwFrom = schedulePassword(cfg, set, targets)
	}
	if pwFrom == "mixed" {
		die(M.SchedulePwMixed)
	}
	if pw == "" {
		die(M.SchedulePwMissing)
	}
	switch pwFrom {
	case "repo":
		fmt.Print(M.SchedulePwFromRepo)
	case "settings":
		fmt.Print(M.SchedulePwFromSettings)
	default:
		fmt.Print(M.SchedulePwFromCLI)
	}

	if err := runSchedule(name, *all, kind, days, *at, pw, *noNotify, *yes); err != nil {
		logf("schedule %s: failed: %v", name, err)
		die("%v", err)
	}
}

func cmdUnschedule(args []string) {
	fs := flag.NewFlagSet("unschedule", flag.ExitOnError)
	all := fs.Bool("all", false, M.FlagAllUnschedule)
	fs.Parse(reorderArgs(args))
	name := ""
	if !*all {
		if fs.NArg() < 1 {
			die(M.UsageUnschedule)
		}
		name = fs.Arg(0)
	}
	if err := runUnschedule(name, *all); err != nil {
		die("%v", err)
	}
}
