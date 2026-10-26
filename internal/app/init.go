package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// initPathScript 把程序目录追加到当前用户的 PATH（HKCU\Environment）。
// 不落 .ps1，也不用 setx（setx 会把 PATH 截断到 1024 字符）；保留原值的类型
// （REG_SZ / REG_EXPAND_SZ），写完广播 WM_SETTINGCHANGE，让新开的终端生效。
const initPathScript = `
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$dir = $env:ZINC_INIT_DIR.TrimEnd('\')
if ($dir -eq '') { Write-Output 'ERROR:empty'; exit 1 }
$key = Get-Item 'HKCU:\Environment'
$kind = 'String'
try { $kind = $key.GetValueKind('Path').ToString() } catch { }
$p = ''
try { $p = [string]$key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames) } catch { }
$parts = @()
if ($p) { $parts = @($p -split ';' | Where-Object { $_ -ne '' }) }
foreach ($x in $parts) {
    $t = $x.Trim().Trim('"').TrimEnd('\')
    if ($t -ieq $dir) { Write-Output 'EXISTS'; exit 0 }
}
$new = (($parts + $dir) -join ';')
# Get-Item 拿到的键是只读句柄，必须重新以可写方式打开。
# 覆盖安装时 OpenSubKey 偶尔会被拒（Access denied），退一步用 reg.exe 写。
$written = $false
try {
    $wkey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
    $wkey.SetValue('Path', $new, [Microsoft.Win32.RegistryValueKind]::$kind)
    $wkey.Close()
    $written = $true
} catch {
    Start-Sleep -Milliseconds 200
    try {
        $wkey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
        $wkey.SetValue('Path', $new, [Microsoft.Win32.RegistryValueKind]::$kind)
        $wkey.Close()
        $written = $true
    } catch { }
}
if (-not $written) {
    $regType = 'REG_SZ'
    if ($kind -eq 'ExpandString') { $regType = 'REG_EXPAND_SZ' }
    & reg.exe add 'HKCU\Environment' /v Path /t $regType /d $new /f | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Output 'ERROR:write-path'; exit 1 }
}
Add-Type -Namespace ZincInit -Name Native -MemberDefinition '[DllImport("user32.dll", CharSet=CharSet.Auto, SetLastError=true)] public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);'
$r = [UIntPtr]::Zero
[ZincInit.Native]::SendMessageTimeout([IntPtr]0xffff, 0x1A, [UIntPtr]::Zero, 'Environment', 2, 5000, [ref]$r) | Out-Null
Write-Output 'ADDED'
`

// addUserPath 把 dir 追加到用户 PATH（幂等）。返回是否真的新增。
func addUserPath(dir string) (bool, error) {
	if strings.ContainsAny(dir, `;"`) {
		return false, fmt.Errorf("路径里含 ; 或双引号，无法安全写入 PATH: %s", dir)
	}
	cmd := exec.Command(powershellPath(), "-NoProfile", "-Command", initPathScript)
	cmd.Env = append(os.Environ(), "ZINC_INIT_DIR="+dir)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return false, fmt.Errorf("%v: %s", err, text)
	}
	switch {
	case strings.Contains(text, "ADDED"):
		return true, nil
	case strings.Contains(text, "EXISTS"):
		return false, nil
	}
	return false, fmt.Errorf("PowerShell 返回了意外结果: %s", text)
}

// registerSetup 注册通知并把程序目录写进用户 PATH；返回是否全部成功。
// zinc init 与 zinc-register.exe 共用这一份逻辑，安装程序只是多了 7-Zip 的强制检查。
func registerSetup(force bool) bool {
	dir := appDir()
	failed := false

	if err := ensureShortcutForce(force); err != nil {
		fmt.Printf(M.InitNotifyFail, err)
		failed = true
	} else {
		fmt.Printf(M.InitNotifyReady, shortcutPath())
	}

	if added, err := addUserPath(dir); err != nil {
		fmt.Printf(M.InitPathFail, err)
		failed = true
	} else if added {
		fmt.Printf(M.InitPathAdded, dir)
	} else {
		fmt.Printf(M.InitPathExists, dir)
	}

	return !failed
}

// cmdInit 一次性初始化：注册通知所需的 AUMID（快捷方式 + 注册表 + 标记），
// 并把程序目录加进用户 PATH，之后直接敲 zinc 即可。
// 7-Zip 只提示不阻断；需要强制的安装流程用 zinc-register.exe。
func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	force := fs.Bool("force", false, M.FlagForce)
	fs.Parse(reorderArgs(args))
	if fs.NArg() > 0 {
		die(M.UsageInit)
	}
	set := mustSettings()

	fmt.Printf(M.InitDir, appDir())
	ok := registerSetup(*force)

	if sz := set.sevenZipPath(); fileExists(sz) {
		fmt.Printf(M.Init7zipOK, sz)
	} else {
		fmt.Printf(M.Init7zipMissing, sz)
	}

	fmt.Print(M.InitHint)
	if !ok {
		os.Exit(1)
	}
}
