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

if (-not $Iscc) {
    $candidates = @(
        (Join-Path $root '..\_tools\InnoSetup6\ISCC.exe'),
        (Join-Path $root '_tools\InnoSetup6\ISCC.exe'),
        (Join-Path ${env:ProgramFiles(x86)} 'Inno Setup 6\ISCC.exe'),
        (Join-Path $env:ProgramFiles 'Inno Setup 6\ISCC.exe')
    )
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
& $Iscc "/DAppVersion=$Version" (Join-Path $PSScriptRoot 'zinc.iss')
if ($LASTEXITCODE -ne 0) { throw 'Inno Setup compile failed' }

Write-Host "built dist\zinc-$Version-setup.exe"
