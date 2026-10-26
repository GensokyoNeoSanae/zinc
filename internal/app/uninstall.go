package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UninstallMain 是 zinc-uninstall.exe 的入口。
//
// 只撤销“注册”：计划任务、通知注册（快捷方式 + AppUserModelId + 标记）、用户 PATH。
// 备份数据（各个输出目录里的 .7z）、settings.json、repos.json、logs 一律不动——
// 删数据这种事不该藏在卸载命令里。
//
// 计划任务按“指向本目录”过滤，不按 zinc- 前缀一把抓：从临时目录跑一次卸载，
// 不会误删另一个 zinc 安装的任务。
func UninstallMain() {
	args := initLang(os.Args[1:])
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	yes := fs.Bool("yes", false, M.FlagYesUninstall)
	keepSchedule := fs.Bool("keep-schedule", false, M.FlagKeepSchedule)
	noPause := fs.Bool("no-pause", false, M.FlagNoPause)
	fs.Parse(reorderArgs(args))

	dir := appDir()
	fmt.Printf(M.UninstallHeader, dir)

	if !*yes && !isConsole() {
		fmt.Println(M.UninstallNeedYes)
		os.Exit(1)
	}

	var tasks []string
	if *keepSchedule {
		fmt.Print(M.UninstallKeepSchedule)
	} else {
		tasks = scheduledTasksForDir(dir)
		if len(tasks) == 0 {
			fmt.Print(M.UninstallNoTasks)
		} else {
			fmt.Printf(M.UninstallTasks, len(tasks))
			for _, t := range tasks {
				fmt.Printf(M.UninstallTaskItem, t)
			}
		}
	}

	if !*yes {
		fmt.Print(M.UninstallConfirm)
		answer := strings.ToLower(strings.TrimSpace(readLine()))
		if answer != "y" && answer != "yes" && answer != "是" {
			fmt.Print(M.UninstallAbort)
			pauseUnless(*noPause)
			return
		}
	}

	failed := false

	if !*keepSchedule && len(tasks) > 0 {
		n, err := removeScheduledTasksForDir(dir)
		if err != nil {
			fmt.Printf(M.UninstallTaskFail, err)
			failed = true
		} else {
			fmt.Printf(M.UninstallTaskRemoved, n)
		}
	}

	owner, err := shortcutOwnerDir()
	switch {
	case err != nil:
		fmt.Printf(M.UninstallTaskFail, err)
		failed = true
	case owner == "":
		fmt.Print(M.UninstallNoNotify)
	case samePathFold(owner, dir):
		os.Remove(shortcutPath())
		os.Remove(markerPath())
		if err := removeAumidKey(); err != nil {
			fmt.Printf(M.UninstallNotifyFail, err)
			failed = true
		} else {
			fmt.Print(M.UninstallNotifyRemoved)
		}
	default:
		fmt.Printf(M.UninstallNotifySkip, owner)
	}

	removed, err := removeUserPath(dir)
	switch {
	case err != nil:
		fmt.Printf(M.UninstallPathFail, err)
		failed = true
	case removed:
		fmt.Printf(M.UninstallPathRemoved, dir)
	default:
		fmt.Printf(M.UninstallPathMissing, dir)
	}

	fmt.Printf(M.UninstallDone, dir)
	pauseUnless(*noPause)
	if failed {
		os.Exit(1)
	}
}

// ---------- 计划任务 ----------

const (
	uninstallTaskListScript = `
$ErrorActionPreference = 'SilentlyContinue'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$dir = $env:ZINC_EXE_DIR.TrimEnd('\')
foreach ($t in (Get-ScheduledTask -TaskName 'zinc-*')) {
    foreach ($a in $t.Actions) {
        if (-not $a.Execute) { continue }
        $d = [IO.Path]::GetDirectoryName($a.Execute)
        if ($d -and $d.TrimEnd('\') -ieq $dir) { $t.TaskName; break }
    }
}
`

	uninstallTaskRemoveScript = `
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$dir = $env:ZINC_EXE_DIR.TrimEnd('\')
$n = 0
foreach ($t in (Get-ScheduledTask -TaskName 'zinc-*' -ErrorAction SilentlyContinue)) {
    $hit = $false
    foreach ($a in $t.Actions) {
        if (-not $a.Execute) { continue }
        $d = [IO.Path]::GetDirectoryName($a.Execute)
        if ($d -and $d.TrimEnd('\') -ieq $dir) { $hit = $true; break }
    }
    if ($hit) {
        Unregister-ScheduledTask -TaskName $t.TaskName -Confirm:$false
        $n++
    }
}
$n
`

	uninstallShortcutOwnerScript = `
$ErrorActionPreference = 'SilentlyContinue'
$lnk = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\zinc.lnk'
if (-not (Test-Path $lnk)) { 'NONE'; exit 0 }
$sh = New-Object -ComObject WScript.Shell
$t = $sh.CreateShortcut($lnk).TargetPath
if (-not $t) { 'NONE'; exit 0 }
[IO.Path]::GetDirectoryName($t).TrimEnd('\')
`

	uninstallPathScript = `
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$dir = $env:ZINC_INIT_DIR.TrimEnd('\')
$key = Get-Item 'HKCU:\Environment'
$kind = 'String'
try { $kind = $key.GetValueKind('Path').ToString() } catch { }
$p = ''
try { $p = [string]$key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames) } catch { }
$parts = @()
if ($p) { $parts = @($p -split ';' | Where-Object { $_ -ne '' }) }
$keep = @()
$found = $false
foreach ($x in $parts) {
    $t = $x.Trim().Trim('"').TrimEnd('\')
    if ($t -ieq $dir) { $found = $true } else { $keep += $x }
}
if (-not $found) { Write-Output 'NOTFOUND'; exit 0 }
$new = ($keep -join ';')
# 同 init：Get-Item 的句柄只读，必须重新以可写方式打开
$wkey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
$wkey.SetValue('Path', $new, [Microsoft.Win32.RegistryValueKind]::$kind)
$wkey.Close()
Add-Type -Namespace ZincInit -Name Native -MemberDefinition '[DllImport("user32.dll", CharSet=CharSet.Auto, SetLastError=true)] public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);'
$r = [UIntPtr]::Zero
[ZincInit.Native]::SendMessageTimeout([IntPtr]0xffff, 0x1A, [UIntPtr]::Zero, 'Environment', 2, 5000, [ref]$r) | Out-Null
Write-Output 'REMOVED'
`
)

// runPS 固定用 -NoProfile -Command，参数走环境变量，不依赖执行策略。
func runPS(script string, env ...string) (string, error) {
	cmd := exec.Command(powershellPath(), "-NoProfile", "-Command", script)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// scheduledTasksForDir 返回命令行指向 dir 下 exe 的 zinc 任务名。
func scheduledTasksForDir(dir string) []string {
	out, err := runPS(uninstallTaskListScript, "ZINC_EXE_DIR="+dir)
	if err != nil || out == "" {
		return nil
	}
	var tasks []string
	for _, line := range strings.Split(out, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			tasks = append(tasks, s)
		}
	}
	return tasks
}

// removeScheduledTasksForDir 删除这些任务并返回实际删掉的数量。
func removeScheduledTasksForDir(dir string) (int, error) {
	out, err := runPS(uninstallTaskRemoveScript, "ZINC_EXE_DIR="+dir)
	if err != nil {
		return 0, fmt.Errorf("%v: %s", err, out)
	}
	var n int
	if _, err := fmt.Sscanf(out, "%d", &n); err != nil {
		return 0, fmt.Errorf("unexpected output: %s", out)
	}
	return n, nil
}

// ---------- 通知注册 ----------

// markerPath 是通知已注册的标记文件。
func markerPath() string { return filepath.Join(appDir(), ".aumid_ok") }

// shortcutOwnerDir 读开始菜单快捷方式的目标目录；不存在时返回空串。
func shortcutOwnerDir() (string, error) {
	out, err := runPS(uninstallShortcutOwnerScript)
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}
	if out == "" || out == "NONE" {
		return "", nil
	}
	return out, nil
}

// removeAumidKey 删掉 AppUserModelId 注册表项。用 reg.exe 避免再拼一段 PowerShell。
func removeAumidKey() error {
	key := `HKCU\Software\Classes\AppUserModelId\` + appID
	out, err := exec.Command("reg", "delete", key, "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ---------- 用户 PATH ----------

// removeUserPath 从 HKCU\Environment 的 Path 里精确删掉 dir，保留原有值类型。
func removeUserPath(dir string) (bool, error) {
	out, err := runPS(uninstallPathScript, "ZINC_INIT_DIR="+dir)
	if err != nil {
		return false, fmt.Errorf("%v: %s", err, out)
	}
	switch {
	case strings.Contains(out, "REMOVED"):
		return true, nil
	case strings.Contains(out, "NOTFOUND"):
		return false, nil
	}
	return false, fmt.Errorf("unexpected output: %s", out)
}
