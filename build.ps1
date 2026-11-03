# 编译四个可执行文件：
#   zinc.exe          控制台版主程序
#   zincw.exe         GUI 子系统版主程序（计划任务用，不弹窗口）
#   zinc-register.exe  安装：检查 7-Zip、注册通知、写用户 PATH
#   zinc-unregister.exe 卸载：删计划任务、通知注册、PATH 项（不碰备份数据）
# 用法：.\build.ps1                用 PATH 里的 go
#       .\build.ps1 -Go C:\Go\bin\go.exe
#       .\build.ps1 -Version 0.7.0  指定注入 zinc --version 的版本号
param(
    [string]$Go = 'go',
    [string]$Version = ''
)

$ErrorActionPreference = 'Stop'

# 版本号优先级：显式传入 > git describe > dev。
# 只给纯数字开头的补 v 前缀，让 `zinc --version` 与 tag 名一致（v0.7.0），
# 而 git describe 的 "v0.7.0-3-gabc123" 和 "dev" 原样保留。
if (-not $Version) {
    $Version = (& git -C $PSScriptRoot describe --tags --always --dirty 2>$null)
    if (-not $Version) { $Version = 'dev' }
}
if ($Version -match '^\d') { $Version = "v$Version" }
$ldflags = "-s -w -X github.com/GensokyoNeoSanae/zinc/internal/app.Version=$Version"
# Go 构建缓存放到临时目录，别在仓库里堆一个几十 MB 的 .gocache
if ($env:TEMP) { $env:GOCACHE = Join-Path $env:TEMP 'zinc-gocache' }
else { $env:GOCACHE = Join-Path $PSScriptRoot '.gocache' }
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null

& $Go build -C $PSScriptRoot -ldflags $ldflags -o zinc.exe ./cmd/zinc
if ($LASTEXITCODE -ne 0) { throw 'zinc build failed' }

& $Go build -C $PSScriptRoot -ldflags "$ldflags -H=windowsgui" -o zincw.exe ./cmd/zinc
if ($LASTEXITCODE -ne 0) { throw 'GUI build failed' }

& $Go build -C $PSScriptRoot -ldflags $ldflags -o zinc-register.exe ./cmd/install
if ($LASTEXITCODE -ne 0) { throw 'installer build failed' }

& $Go build -C $PSScriptRoot -ldflags $ldflags -o zinc-unregister.exe ./cmd/uninstall
if ($LASTEXITCODE -ne 0) { throw 'uninstaller build failed' }

# 产物如果带 Internet 标记（Zone.Identifier），双击会被 SmartScreen 当成下载文件拦一下
Get-ChildItem -Path $PSScriptRoot -Filter *.exe | Unblock-File

Write-Host 'built zinc.exe, zincw.exe, zinc-register.exe, zinc-unregister.exe'
