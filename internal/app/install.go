package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallMain 是 zinc-register.exe 的入口。
//
// 与 zinc init 的区别只有一处：7-Zip 是硬性依赖。找不到 7z.exe / 7za.exe 时要求用户
// 指定路径，指定不成功就直接退出，且不写任何系统状态（PATH、通知注册都不动），
// 所以失败的安装不会留下半套配置。
func InstallMain() {
	args := initLang(os.Args[1:])
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	szFlag := fs.String("7zip", "", M.FlagInstall7zip)
	force := fs.Bool("force", false, M.FlagForce)
	noPause := fs.Bool("no-pause", false, M.FlagNoPause)
	fs.Parse(reorderArgs(args))

	dir := appDir()
	fmt.Printf(M.InstallHeader, dir)

	sz, err := ensureSevenZip(*szFlag)
	if err != nil {
		fmt.Printf("%s\n", err)
		pauseUnless(*noPause)
		os.Exit(1)
	}
	fmt.Printf(M.Install7zipOK, sz)

	if !registerSetup(*force) {
		pauseUnless(*noPause)
		os.Exit(1)
	}

	fmt.Print(M.InstallDone)
	pauseUnless(*noPause)
}

// ensureSevenZip 确定 7z 可执行文件，顺序是：
// 命令行 --7zip > settings.json 里已有且仍存在的路径 > PATH 与常见安装目录 > 交互输入。
// 找到后写回 settings.json，之后 zinc.exe 直接可用。
func ensureSevenZip(explicit string) (string, error) {
	set, err := loadSettings()
	if err != nil {
		return "", err
	}

	if explicit != "" {
		p, err := checkSevenZip(explicit)
		if err != nil {
			return "", err
		}
		return p, saveSevenZipPath(set, p)
	}

	if p := set.SevenZipPath; p != "" && fileExists(p) {
		return p, nil
	}

	for _, c := range sevenZipCandidates() {
		if fileExists(c) {
			return c, saveSevenZipPath(set, c)
		}
	}

	fmt.Print(M.PromptSevenZip)
	answer := strings.Trim(strings.TrimSpace(readLine()), `"'`)
	if answer == "" {
		return "", fmt.Errorf("%s", M.SevenZipMissing)
	}
	p, err := checkSevenZip(answer)
	if err != nil {
		return "", err
	}
	return p, saveSevenZipPath(set, p)
}

// checkSevenZip 校验用户给的路径确实是个可执行文件。
func checkSevenZip(p string) (string, error) {
	abs, err := absClean(p)
	if err != nil {
		return "", fmt.Errorf(M.SevenZipNotFound, p)
	}
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return "", fmt.Errorf(M.SevenZipNotFound, p)
	}
	if !strings.HasSuffix(strings.ToLower(abs), ".exe") {
		return "", fmt.Errorf(M.SevenZipNotFound, p)
	}
	return abs, nil
}

// sevenZipCandidates 按“最可能先被找到”的顺序列出候选。
// 7z.exe 与 7za.exe 的参数集对 zinc 用到的部分完全一致，装完整版 7-Zip 的用户也能直接用。
// 这里只找约定位置（PATH、Program Files、LOCALAPPDATA）；装在别处就靠 --7zip 或交互输入。
func sevenZipCandidates() []string {
	var cands []string
	for _, name := range []string{"7za.exe", "7z.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			cands = append(cands, p)
		}
	}
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
		if root := os.Getenv(env); root != "" {
			cands = append(cands,
				filepath.Join(root, "7-Zip", "7z.exe"),
				filepath.Join(root, "7-Zip", "7za.exe"))
		}
	}
	if root := os.Getenv("LOCALAPPDATA"); root != "" {
		cands = append(cands,
			filepath.Join(root, "Programs", "7-Zip", "7z.exe"),
			filepath.Join(root, "Programs", "7-Zip", "7za.exe"))
	}
	return cands
}

// saveSevenZipPath 只在值变化时写盘，避免装一次就多一次无意义的设置文件写入。
func saveSevenZipPath(set *Settings, p string) error {
	if set.SevenZipPath == p {
		return nil
	}
	set.SevenZipPath = p
	if err := saveSettings(set); err != nil {
		return fmt.Errorf(M.SaveConfigFailed, err)
	}
	return nil
}
