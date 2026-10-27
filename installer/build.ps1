# 用 Inno Setup 6 编译安装包。
# 需要 ISCC.exe：可以装在工作区（_tools\InnoSetup6）或默认安装位置，也可以用 -Iscc 指定。
param(
    [string]$Version = '0.1.0',
    [string]$Go = 'go',
    [string]$Iscc = ''
)

$ErrorActionPreference = 'Stop'
if ($Version.StartsWith('v') -or $Version.StartsWith('V')) { $Version = $Version.Substring(1) }
$root = Split-Path -Parent $PSScriptRoot

# VersionInfoVersion 只吃 1-4 段纯数字：0.0.0-dev → 0.0.0.0，v0.1.1 → 0.1.1.0。
# 展示版本（AppVersion）可以是任意字符串，文件版本必须干净，否则 ISCC 直接报 invalid。
$numeric = '0.0.0.0'
if ($Version -match '^(\d+)\.(\d+)\.(\d+)(?:\.(\d+))?') {
    $build = if ($Matches[4]) { $Matches[4] } else { '0' }
    $numeric = "$($Matches[1]).$($Matches[2]).$($Matches[3]).$build"
}

if (-not $Iscc) {
    $candidates = @(
        (Join-Path $root '..\_tools\InnoSetup6\ISCC.exe'),
        (Join-Path $root '_tools\InnoSetup6\ISCC.exe'),
        (Join-Path ${env:ProgramFiles(x86)} 'Inno Setup 6\ISCC.exe'),
        (Join-Path $env:ProgramFiles 'Inno Setup 6\ISCC.exe'),
        (Join-Path $env:LOCALAPPDATA 'Programs\Inno Setup 6\ISCC.exe')
    )
    # 装在非默认位置（例如 E:\Tools\Inno Setup 6）时，靠卸载登记里的 InstallLocation 找。
    foreach ($k in @(
        'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1',
        'HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1',
        'HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1')) {
        try {
            $loc = (Get-ItemProperty -Path $k -ErrorAction Stop).InstallLocation
            if ($loc) { $candidates += (Join-Path $loc 'ISCC.exe') }
        } catch { }
    }
    foreach ($c in $candidates) {
        if ($c -and (Test-Path $c)) { $Iscc = $c; break }
    }
}
if (-not $Iscc -or -not (Test-Path $Iscc)) {
    throw '找不到 Inno Setup 的 ISCC.exe。先安装 Inno Setup 6，或用 -Iscc <path> 指定。'
}

# 1. 先出主程序的四个 exe
& (Join-Path $root 'build.ps1') -Go $Go

# 2. 编译安装包（输出到 dist\zinc-<版本>-setup.exe）
& $Iscc "/DAppVersion=$Version" "/DAppVersionNumeric=$numeric" (Join-Path $PSScriptRoot 'zinc.iss')
if ($LASTEXITCODE -ne 0) { throw 'Inno Setup compile failed' }

# 清掉可能带上的 Internet 标记，避免双击时多一次 SmartScreen
Unblock-File (Join-Path $root "dist\zinc-$Version-setup.exe") -ErrorAction SilentlyContinue

Write-Host "built dist\zinc-$Version-setup.exe"
