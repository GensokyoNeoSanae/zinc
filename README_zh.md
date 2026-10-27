# zinc

[English](README.md) · **简体中文**

[![Go 1.27+](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078D4?logo=windows&logoColor=white)](#环境要求)
[![Build](https://github.com/GensokyoNeoSanae/zinc/actions/workflows/build.yml/badge.svg)](https://github.com/GensokyoNeoSanae/zinc/actions/workflows/build.yml)

**文件级增量备份，产物是普普通通的 7-Zip 归档：每一批都是一个完整的加密 `.7z`，恢复时只需要 7-Zip 和密码。**

zinc 把一个或多个源目录（称为“仓库”）备份到同一个输出根目录下。每次备份只打包上次之后新增或改动过的文件，每一批落地成一个或多个自洽的 `.7z` 压缩包，包内文件名同样加密。它没有私有仓库格式，没有增量链，恢复时也不需要 zinc 本身。

## 目录

- [为什么做这个](#为什么做这个)
- [特性](#特性)
- [工作原理](#工作原理)
- [环境要求](#环境要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [命令](#命令)
- [配置](#配置)
- [恢复](#恢复)
- [定时任务与通知](#定时任务与通知)
- [文件位置](#文件位置)
- [卸载](#卸载)
- [已知限制](#已知限制)
- [常见问题](#常见问题)
- [从源码构建](#从源码构建)
- [参与贡献](#参与贡献)
- [许可证](#许可证)
- [致谢](#致谢)

## 为什么做这个

- **网盘同步的是文件夹，不是仓库。** 把所有源目录挂到同一个输出根下，一个同步客户端就能覆盖全部。
- **分卷本质上仍是一份归档。** 用 `-v4g` 切出来的卷，缺一卷、坏一卷就可能整批打不开；zinc 写的是完整独立的包，丢一个包只丢一批。
- **恢复不应该依赖备份工具。** 每个包都是带加密头的普通 7-Zip 归档，按文件名顺序 `7z x` 一遍就是全部恢复流程。
- **上传量应该跟着改动量走。** 文件级索引意味着今天加一张照片，不会把去年的图库重新传一遍。
- **它是归档，不是镜像。** 写好的包永远不改、不剪。恢复之所以不依赖工具，代价就在这里：删除单独记在包旁边的小文本里，而不是回头改包；由此排除掉的用法见[已知限制](#已知限制)。

## 特性

- **文件级增量**：只压新增和改动过的文件，不重压整个目录。
- **独立包**：每一批是一个或多个完整归档，坏了、丢了只损失那一批。
- **头加密**：`-mhe=on`，没有密码连文件名都看不到。
- **不装 zinc 也能恢复**：7-Zip 加密码就够了，灾难恢复流程见下文。
- **记得住删除**：源目录里删掉的文件，恢复时不会再冒出来，而旧包一个字节都不用改——每批在包旁边留一份小文本记录。
- **单路径恢复**：只把一个文件、一个目录或一批通配符命中项从包链里取出来，不用整仓解压。
- **索引只是缓存**：`index.tsv` 丢了可以从归档头重建，不需要解压内容。
- **多仓库**：图片、视频、文档共用同一个同步根，各自有独立的上限、密码和排除名单。
- **定时任务**：按天、按周或按月创建计划任务，跑完弹 Windows 通知。
- **中英双语**：内置两套消息，默认跟随 Windows 界面语言。
- **零第三方依赖**：只用标准库，四个小程序加一个 7-Zip。

## 工作原理

一次备份就是一条流水线：

```
扫描源目录  →  与索引比对  →  切包计划  →  7za a（先写 .part）  →  7za t 自检  →  改名  →  更新索引、manifest 与删除记录
```

1. **扫描**：源目录只枚举一遍，文件大小和修改时间直接取自目录项，不需要逐个 `stat`。
2. **比对**：索引里没有这个路径，或者大小、修改时间变了，才算“改动过”；没变的文件以后不会再碰。索引里有、这次却扫不到的路径不会被打包——那是删除。
3. **切包**：改动过的文件按修改时间升序排列（同时间按相对路径），依次贪心塞进当前包；塞下一个会超过 `max_pack_byte` 就封包、另起一个。
4. **打包**：每个包都以 `7za a -t7z -p<密码> -mhe=on -mx<N> -scsUTF-8 -sccUTF-8` 写进一个 `.part` 文件，打完用 `7za t` 自检，通过后才改名成最终文件名。一批里任何一个包失败，整批已写出的包会被删除，索引不动。
5. **记删除**：消失的路径写进包旁边一个纯文本旁注，名字跟着这一批最后一个包走。旧包一个字都不改；如果这一批只删了文件、没有新内容，就只留下这份记录，不产生包。
6. **记账**：更新索引和 SHA-256 清单（manifest），都是先写临时文件再改名。删掉的路径同时从索引里划掉，下次不会再被当成“刚消失”。

### 每个包为什么叫这个名字？

```
YYYYMMDD_hhmmss_SSS.7z
20261025_070012_334.7z
```

文件名不是装饰：**它就是恢复顺序**。全量恢复按文件名升序把包逐个解开，后解的覆盖先解的，所以名字必须永远严格递增，而且写下去以后不能再改。

所以名字取的是**这个包开始打包的时刻**，定宽 19 个字符，让字典序等于时间序。重名时追加 `_1`、`_2`（排序时按数字比较）；系统时钟回退时取“已存在最大包名 + 1 毫秒”并打印警告，保证顺序不会倒退。

用打包时刻而不是内容的修改时间命名，是因为恢复的正确性依赖“后备份的覆盖先备份的”。一个文件先备份了新版、之后又把旧版拷回来（软件降级、笔记库回滚），它的内容 mtime 会**倒退**；若用内容时间命名，这个包会被排到前面，全量恢复出来就是新旧混杂的错误状态，而且不报错。内容的 mtime 区间仍然记在索引的 `P` 行里，只是不进文件名。

### `index.tsv` 是干什么的？

它是增量的根据，也是 zinc 判断“哪些文件变了”唯一要读的东西。每次备份先扫一遍源目录，然后对每个文件问一句：这个路径存过没有，大小或修改时间动过没有？答案就来自 `index.tsv`——它把每个已备份的路径映射到装着它的那个包。没有它就没有比对：zinc 会把仓库当成空的，把全部内容重新压一遍。

恢复也要用它：`--only` 查这张表就能直接跳到装着最新版本的包，不用从第一个包翻起。

`<output>/index.tsv` 是缓存，权威数据是包本身。

```
# zinc-index 2
P	1	20261024_110635_081.7z	20261019	20261023
F	笔记/semiconductor.md	2048	20261023T124138.9843416+0800	1
# end	1	1
```

- `P` 行：包编号、包文件名、该包内容的最早与最晚 mtime
- `F` 行：相对路径、字节数、mtime（`20261023T124138.9843416+0800`）、所属包编号
- `# end` 尾行：包数与文件数，用于检测截断

头部不对、某行格式坏了、缺尾行、计数不符，都直接报错，绝不静默当成空索引。索引丢了只损失重建时间，不损失数据：

```powershell
zinc reindex 图片          # 只读归档头重建
zinc reindex 图片 --yes    # 覆盖已有索引（旧文件会留一份 .bak）
```

`reindex` 走的是 `7za l -slt`，不解压任何内容。

### 删除是怎么记下来的？

包只会往里加文件，“这个没了”没法写进包里。所以它另开一份删除记录：包旁边一个纯文本旁注，一行一个相对路径，名字跟着“发现它消失”的那一批走。

```
<输出目录>\
  20261024_110635_081.7z    ← 包
  20261025_070012_334.7z    ← 包
  20261025_070012_334.rm    ← 这一批的删除记录
```

```
# zinc-deleted 1
# sha256 8f14e45fceea167a5a36dedd4bea2543…
notes/旧稿.md
attachments/2026/截图.png
```

- **顺序由文件名决定。** 记录挂在它那一批最后一个包的时间上，所以紧接着那个包排序。`restore` 把包和记录放在同一条时间线上跑，这就是为什么“删掉又加回来”的文件能恢复到正确的那一版。
- **两种文件都带哈希。** 第二行是路径清单自己的 SHA-256，位翻转或手改在读记录时就会被抓住；`manifest.json` 里另记一份整文件哈希，`zinc verify` 两边都重算。任何一项对不上都直接停下，而不是照着一条可疑记录去删文件。
- **记录坏了不会把仓库锁死。** `zinc restore --no-delete` 会跳过记录（给一条提示）照样把文件恢复出来；包本身也一直能手工解压。
- **只记真的没了的文件。** 命中 `exclude` 的不算删除，扫描时读不进去的目录（权限问题）下面的也不算——一次权限报错不该变成恢复时删掉真文件。
- **源目录整个不在，不等于全删了。** 如果 `--source` 不存在，备份直接报“源目录不存在”停下，而不会记下几千条删除。但如果目录在、只是里面空了，那就照实记：里面的东西确实都没了。
- **动手之前先校验。** 恢复会先把所有记录里的路径过一遍：绝对路径、盘符、`..` 段、NTFS 数据流，任何一种都直接中止，一个文件都不解压。
- **历史不会被改写。** 旧包里仍然留着它当初装进去的所有文件，手工 `7z x` 照样能全解出来。想让手工恢复也精确等于最后一次备份，解到最后一个包之后，照着 `.rm` 里列出的路径自己删一遍就行。

### 为什么不能“恢复到某一天”？

因为包不是按天切的。一个包里可能装着横跨好几天的文件，硬按日期截断，要么把太新的文件带进来，要么把本该有的文件丢掉。想拿最新状态，`zinc restore` 就是；想确认数据没坏，用 `zinc verify`。确实需要按时间点回溯历史，那是 kopia、restic 这类仓库型工具的活。

## 环境要求

- **Windows 10 / 11，64 位。** zinc 用到计划任务、Windows 通知和 Win32 控制台接口，不在其他平台上构建或运行。
- **7-Zip。** zinc 调用 `7z.exe` 或 `7za.exe`，装官方 7-Zip 就够。安装包会找到已有的 7-Zip 并把路径写进 `settings.json`；如果它装在奇怪的位置，装的时候手动指一下，或者之后用 `zinc config set 7zip_path <路径>` 指定。
- **Go 1.27 或更新**，仅从源码构建时需要。

## 安装

1. 从 [releases 页面](https://github.com/GensokyoNeoSanae/zinc/releases) 下载 `zinc-<版本>-setup.exe` 并运行。第一页选语言，英文默认，第二项是简体中文。安装包没有代码签名，SmartScreen 可能弹出“Windows 已保护你的电脑”，点 **更多信息 → 仍要运行** 即可。向导不需要管理员权限：默认装到 `%LOCALAPPDATA%\Programs\zinc`（目标目录页可以改，安装包不会沿用上一次的选择），注册通知身份，并把目录加进当前用户的 `PATH`。
2. 确认装了 7-Zip（<https://www.7-zip.org/>）。向导会读 7-Zip 自己的注册表项，再看 `PATH` 和常见安装目录；也可以手动指定别的路径，路径右边的 **自动查找** 会重新搜一遍。没有可用的 7-Zip 不会继续。如果还没装，可以勾选这一页的 **自动安装 7-Zip（需要联网）**，并选择装到哪里（默认 `C:\Program Files\7-Zip`）：向导从官网下载官方安装包并静默安装，过程中会弹一次管理员提权。
3. 开一个**新**终端，选备份根目录：

   ```powershell
   zinc root D:\Backup\zinc        # 示例路径，换成你自己的
   ```

   之后新建的仓库默认输出到 `<root>\<仓库名>`。

向导会把 `zinc.exe`、`zincw.exe`、`tools\` 下的两个辅助程序、这份 README 和 `LICENSE` 装进你选的目录；`settings.json`、`repos.json` 也在那里按需生成，不进备份。卸载程序（`unins000.exe`）由安装包生成，卸载请走 **设置 → 应用**，别指望复制某个文件出去。

想要绿色版的话，release 里另有一份 `zinc-<版本>-windows-amd64.zip`：解压到任意目录，跑里面的 `tools\zinc-register.exe` 即可。它做的事和安装包一样——注册通知身份、写用户 `PATH`、检查 7-Zip，只是不会出现在“设置 → 应用”里。`zinc init` 是更轻的一档：通知和 `PATH` 照做，但 7-Zip 只提示、不强制。

## 快速开始

```powershell
# 建仓库（本文里所有路径都是示例，换成你自己的目录）
zinc add pictures --source D:\Storage\Pictures
zinc add videos   --source D:\Storage\Videos

# 立即备份（没配默认密码时交互输入两遍；没有新增时不问密码）
zinc backup pictures

# 备份全部仓库
zinc backup --all

# 查看状态
zinc list
zinc status
zinc status pictures

# 恢复（默认覆盖同名文件）
zinc restore pictures --dest D:\restore\pictures

# 只看看会按什么顺序解压哪些包，不动文件
zinc restore pictures --check-list
```

## 命令

| 命令 | 说明 |
|---|---|
| `zinc init [--force]` | 重新注册通知、把程序目录写进用户 `PATH`。安装包已经做过，只在通知失效、或 `zinc` 命令找不到时用；绿色版也靠它 |
| `zinc root [<path>]` | 查看或设置全局输出根目录 |
| `zinc list` | 列出仓库 |
| `zinc add <name> --source <path> [--output <path>]` | 新建仓库 |
| `zinc remove <name>` | 移除仓库（不删包） |
| `zinc set <name> [--source <path>] [--output <path>] [--password <pw>] [--max-pack-byte <n>] [--compress <0-9>] [--exclude <glob>]` | 修改仓库 |
| `zinc config [get\|set\|unset\|path] [<key>] [<value>]` | 全局设置 |
| `zinc backup <name\|--all> [--password <pw>] [--notify] [--notify-empty] [--no-verify]` | 备份 |
| `zinc restore <name> [--dest <path>] [--password <pw>] [--overwrite\|--no-overwrite] [--no-delete] [--check-list] [--only <路径>] [--from <包>]` | 恢复全量或单个路径 |
| `zinc reindex <name> [--password <pw>] [--pack-start <name>] [--pack-end <name>] [--yes]` | 从包重建索引 |
| `zinc verify <name\|--all> [--sample <pct>]` | 校验包与删除记录的 SHA-256 |
| `zinc status [<name>]` | 查看状态 |
| `zinc schedule <name\|--all> --daily --at HH:MM` | 每天定时 |
| `zinc schedule <name\|--all> --weekly 1,3 --at HH:MM` | 每周定时（1 = 周一） |
| `zinc schedule <name\|--all> --monthly 1,15 --at HH:MM [--yes]` | 每月定时 |
| `zinc unschedule <name>` | 删除该仓库的定时任务 |
| `zinc unschedule all` | 删除“备份全部”的汇总任务（`zinc-all-*`） |
| `zinc unschedule --all` | 删除所有 zinc 任务（**慎用**） |

`--lang zh|en` 放在子命令前后都行。`zinc verify --sample 20` 只随机抽查 20% 的包文件，而不是全部。

### 安全校验

- `add` 和 `set` 统一存绝对路径，并拒绝源目录与输出目录互相包含的配置。
- 两个仓库不能共用同一个输出目录。只有大小写不同的输出会给出警告，因为 Windows 上它们会落到同一个目录。
- `restore --dest` 拒绝源目录、本仓库的输出目录（及其子目录）和盘根。
- `remove` 时如果还有 `zinc-<name>-*` 计划任务，会提示残留。

## 配置

全局设置在 `<程序目录>\settings.json`，**不进备份、不同步网盘**。优先级：命令行参数 > 仓库字段 > `settings.json` > 内置默认。

| 键 | 默认 | 说明 |
|---|---|---|
| `default_password` | 空 | 空 = 每次询问 |
| `restore_overwrite` | `true` | 恢复时覆盖同名文件 |
| `restore_wrap_folder` | `true` | 整仓恢复时在外层包一个 `restore-<仓库>-<时间>` 文件夹 |
| `max_pack_byte` | `4G` | 单包上限，支持 `4G`、`50M` 或纯字节 |
| `compression_level` | `0` | 7z 压缩等级 0-9；`0` = 仅存储 |
| `7zip_path` | 安装时自动检测 | `7z.exe` / `7za.exe` 的路径；空 = 在 `PATH` 和常见安装目录里找 |
| `lang` | 空 | `zh` 或 `en`；空 = 跟随 Windows |
| `lock_timeout_minutes` | `30` | 仓库锁最长等待分钟数 |
| `exclude` | `[]` | 排除名单，glob；`set` 追加 |
| `complete_notify` | `true` | 备份完成发通知 |

```powershell
zinc config                       # 全部打印（密码打码）
zinc config get max_pack_byte     # 只输出裸值，方便脚本用
zinc config set max_pack_byte 50M
zinc config set exclude '**/node_modules/**'
zinc config unset exclude         # 回落到内置默认
zinc config path                  # 配置文件路径
```

`exclude` 的 glob 对相对路径匹配：`**` 跨目录，`*` 段内，`?` 单字符。命中目录会跳过整棵子树。

### 语言

界面消息内置中英两套。解析顺序：`--lang zh|en`、`ZINC_LANG` 环境变量、`settings.json` 里的 `lang`、编译时默认值（`-X main.defaultLang`）、Windows 界面语言，最后兜底英文。

```powershell
zinc --lang en list
zinc list --lang zh
zinc config set lang zh
```

### 密码

四种来源，优先级：`--password` > 仓库的 `password` > `settings.json` 的 `default_password` > 交互输入。

```powershell
# 设一次，以后无人值守；值写 - 表示交互输入，不进 shell 历史
zinc config set default_password -
```

交互输入是明文可见的（方便发现手滑），`backup` 会要求输两遍。设了默认密码之后，备份汇总行会标注密码来源，例如 `enc✓(settings)`。

密码以明文存在 `settings.json` 和计划任务命令里（`C:\Windows\System32\Tasks\zinc-*`）。这套加密防的是网盘自动内容扫描，不是已经拿到本机权限的定向攻击者。

**换密码后必须重新 `schedule`**，否则任务里还是旧密码。

## 恢复

恢复的语义只有一句话：所有包按文件名（时间）升序全部解压一遍，后压的覆盖先压的。删除记录排在同一个时间线上，紧跟它名字所对应的那个包执行，所以拿回来的就是“最后一次备份时的目录”，而不是“历史上出现过的所有文件的并集”。

| 参数 | 行为 |
|---|---|
| 不填 | 按 `restore_overwrite`（默认 `true` = 覆盖） |
| `--overwrite` | 强制覆盖 |
| `--no-overwrite` | 保留两份，同名文件解压成 `_1`、`_2` |
| `--no-delete` | 只恢复文件，不执行后续备份记下的删除 |
| `--check-list` | 只打印执行顺序和体积（包与删除记录），不动文件；此时可以不写 `--dest` |
| `--only <路径>` | 只恢复这一个文件或目录，其他文件不动；不套用删除记录 |
| `--from <包>` | 配合 `--only`，从指定的包恢复：完整包名、唯一前缀，或该仓库**全部包**里的 1 起序号；不是 `--check-list` 打印的那套编号 |

覆盖模式启动时会打印警告。恢复严格按包名顺序逐个解压，所以改名或手工拷入的包会改变最后胜出的版本——请保留 zinc 生成的包名。

### 恢复外层文件夹

默认恢复会在 `--dest` 下面再建一层 `restore-<仓库名>-<YYYYMMDD_hhmmss_SSS>`（时间取本次恢复开始的时刻），文件解压进这层，不直接散落在 `--dest` 根上：

```
<dest>\
  restore-pictures-20261025_170012_334\
    2024\...
```

- 关掉（恢复成直接平铺到 `--dest`）：`zinc config set restore_wrap_folder false`
- 恢复中断后重跑会得到新的时间戳文件夹，不会和上次的半成品混在一起
- 只作用于整仓恢复；`--only` 单路径恢复仍然直接落到 `--dest`

### 只恢复一个文件

不用解压整个仓库：

```powershell
# 找回某个文件的最新版本（自动定位它在哪个包里）
zinc restore docs --only 'notes/chapter1.md' --dest D:\recovered

# 整个目录
zinc restore docs --only 'notes/archive' --dest D:\recovered

# 通配符
zinc restore docs --only '*.md' --dest D:\recovered

# 先看这个文件分布在哪些包里（不动文件）
zinc restore docs --only 'notes/chapter1.md' --check-list

# 从某个旧版本里取（用 --check-list 打印出来的包名）
zinc restore docs --only 'notes/chapter1.md' --from 20261020_070012_334.7z --dest D:\old
```

定位规则：

- 精确文件路径：直接查 `index.tsv`，一次命中最新版本所在包
- 目录或通配符：从旧到新遍历所有包含匹配项的包逐个解压，后压的覆盖先压的，结果与全量恢复一致
- 索引缺失也不影响：自动退回翻包目录（`7za l`，不解压）

通配符按 7-Zip 的匹配语义，`*` 不保证跨目录：`--only '*.md'` 只命中根层文件。要恢复整棵子树就直接给目录路径（`--only 'notes/archive'`），别用通配符猜。

`--only --check-list` 需要密码，因为列包内文件要读加密的归档头。

## 定时任务与通知

```powershell
# 每天 20:00 备份全部仓库，完成后弹通知
zinc schedule --all --daily --at 20:00 --password yourpassword

# pictures 每天 10:00 和 20:00 各一次
zinc schedule pictures --daily --at 10:00,20:00 --password yourpassword

# 每周一、三 20:00
zinc schedule pictures --weekly 1,3 --at 20:00 --password yourpassword

# 每月 1 号、15 号 20:00（写 29–31 号要加 --yes，否则直接拒绝创建）
zinc schedule pictures --monthly 1,15 --at 20:00 --password yourpassword
```

计划任务通过任务计划程序创建，指向 `zincw.exe`（GUI 子系统），执行时不会闪控制台窗口。通知规则：

- `zinc backup` 在 `complete_notify` 开着（默认）时会弹窗；手动跑想安静就加 `--notify=false`
- `zinc schedule` 创建的任务命令行里固定带 `--notify`（建任务时用了 `--no-notify` 除外），`complete_notify` 对它无效
- `--all` 弹一条汇总；单仓库任务各弹一条
- 全部无新增时不弹（`--notify-empty` 强制）

通知机制：

Windows 通知要求应用有 AUMID（AppUserModelID），未打包的 exe 默认没有。安装包在安装时就把下面三样建好；绿色版则是在第一次 `--notify` 时建：

1. `%APPDATA%\Microsoft\Windows\Start Menu\Programs\zinc.lnk`，写入 AUMID `zinc.backup`
2. `HKCU\Software\Classes\AppUserModelId\zinc.backup`（DisplayName = zinc）
3. 程序目录下的 `.aumid_ok` 标记，之后不再重复

**三样都在，通知才能弹。** 如果通知突然不弹了，让 zinc 自己重建一遍，不用到处找文件删：

```powershell
zinc init --force
```

它会一次性重写快捷方式、注册表项和标记文件，不管程序装在哪个目录。

图标：放一个 `zinc.ico` 在 `zinc.exe` 旁边，再跑一次 `zinc init --force` 就会用上。

### 删除任务：`all` 和 `--all` 差两个横线，差的是全部

```powershell
zinc unschedule pictures   # 只删 pictures 仓库的任务
zinc unschedule all        # 只删“备份全部”的汇总任务
zinc unschedule --all      # 删掉所有 zinc 任务，包括各仓库的
```

想删汇总任务却敲成 `--all`，会把所有仓库的定时备份一起删掉。

## 文件位置

输出根目录（把这个目录同步到网盘）：

```
<root>\
  pictures\
    20261024_110635_081.7z    ← 每个包都是完整归档，按名字升序恢复
    20261025_070012_334.7z
    20261025_070012_334.rm    ← 这一批删掉的文件（纯文本）
    index.tsv                 ← 增量检测缓存，丢了用 zinc reindex 重建
    manifest.json             ← SHA-256 账本，verify 用
  videos\
```

| 文件 | 位置 | 上传网盘 | 用途 |
|---|---|---|---|
| `*.7z` | `<output>/` | 是 | 数据本身，每个都是完整归档 |
| `*.rm` | `<output>/` | 是 | 这一批的删除记录及其校验和；恢复要靠它删掉文件，必须跟包一起上传 |
| `index.tsv` | `<output>/` | 是 | 增量检测缓存，丢了用 `zinc reindex` 重建 |
| `manifest.json` | `<output>/` | 是 | 每批的 SHA-256 记录，`verify` 用；必须随包上传，否则换机后没法校验旧包 |

运行时文件（程序目录，不同步）：

| 文件 | 用途 |
|---|---|
| `zinc.exe` | 控制台版，手动跑 |
| `zincw.exe` | GUI 版，计划任务用 |
| `tools\zinc-register.exe` | 注册辅助程序，由安装包调用：检查 7-Zip、注册通知、写用户 `PATH` |
| `tools\zinc-unregister.exe` | 反过来：清计划任务、通知注册、`PATH` 项 |
| `unins000.exe` | 安装包生成的卸载程序，**设置 → 应用**里点卸载跑的就是它 |
| `zinc.ico` | 可选，通知快捷方式图标 |
| `repos.json` | 仓库注册表，首次 `add` 时生成 |
| `settings.json` | 全局默认值，安装时写入 7-Zip 路径，之后由 `config set` 维护 |
| `logs\zinc-YYYYMMDD.log` | 按天一个文件，保留最近 10 份 |
| `.aumid_ok` | 通知注册标记 |

### 灾难恢复

电脑炸了之后不需要 `zinc.exe`：

1. 把某个仓库的 `.7z` 包全部下到一个目录
2. 按文件名升序逐个解压：

```powershell
7z x 20261024_110635_081.7z -pPASSWORD -oRESTORE_DIR
7z x 20261024_120901_552.7z -pPASSWORD -oRESTORE_DIR
# ……按名字顺序解压完所有包，后解压的覆盖先解压的
```

每个包都是完整独立的归档，所以坏一个只缺那一次备份的内容，其余照常解。

手工解压得到的是超集：后来被删掉的文件还在，因为包没法反过来做删除。想在没有 `zinc.exe` 的情况下也还原到“最后一次备份”的状态，就按包的执行顺序把对应的 `.rm` 也过一遍，照着里面列的路径自己删——它是纯文本，一行一个相对路径，文件名本身就说明了它在时间线上的位置。

### 换机

备份数据本身不用操心：包就是普通归档。跟着机器走的只有设置。

1. 在新机器上跑一遍安装包。7-Zip 检查、通知注册、`PATH` 它都会做，不用手工复制程序目录。
2. 用 `zinc root <根目录>` 指到同一份备份位置；再把旧机器安装目录里的 `repos.json`、`settings.json` 复制到新目录，各仓库的源目录、输出目录和密码就都回来了。手工重新 `add`、`config` 一遍效果相同。
3. 重新建计划任务（`zinc schedule ...`）：任务计划程序里的条目属于本机，不会跟着走。

## 卸载

打开 **设置 → 应用 → 安装的应用**，找到 `zinc` 点卸载。实际跑的是 Inno Setup 生成的卸载程序（安装目录里的 `unins000.exe`）：先问“确定要完全卸载吗”，确认之后再问是否保留设置，然后才开始删文件。

删文件之前它会调用 `tools\zinc-unregister.exe` 撤销注册：

- 停掉并删除命令行指向本目录的计划任务——程序没了，备份不会还在后台跑
- 清掉通知注册：开始菜单快捷方式、`HKCU\Software\Classes\AppUserModelId\zinc.backup`、`.aumid_ok`
- 从用户 `PATH` 里移除这个目录

保留设置的问题在确认卸载之后才弹：选“否”连 `settings.json`、`repos.json`、`logs` 一起删，目录整个清空；选“是”只删程序，留下仓库列表和默认密码，方便以后重装。静默卸载（`/VERYSILENT`）按“否”处理。`.aumid_ok` 这类注册标记不算用户数据，总是删掉；备份数据——输出目录里的 `.7z` 包和 manifest——一律不碰。

如果装完之后移动过程序目录，计划任务还指着旧路径，卸载时匹配不到：要么把目录放回原处，要么先 `zinc unschedule --all` 清掉。

## 已知限制

zinc 回答的是“这个目录每次备份时长什么样”，恢复能把它还原成最后一次备份的样子，包括那时已经删掉的文件。它做不到的是回到任意一个历史时刻、对改名去重、以及回收空间。经常整理、大量改名和删除的目录，比如笔记、Obsidian 库、源码树，仍然不适合它；这类场景用 [kopia](https://kopia.io/)、[restic](https://restic.net/) 这种仓库型工具更合适，代价是恢复时必须用它的工具。

- **删除会记下来，但不会抹掉。** 删掉的文件不再打包，索引里的记录会划掉，这一批把它写进一份 `.rm` 记录；旧包一个字都不改，字节还在。恢复会套用这些记录，所以拿回来的就是最后一次备份的目录，但空间只有新建输出目录重备一遍才省得下来。
- **移动或改名会重新打包。** 索引按相对路径记账，文件名或目录一变就是删一个加一个，内容会完整重压一遍。重命名一整棵附件目录，可能要整棵重新上传。（删的那一半只是 `.rm` 里的一行，不会让旧包变小。）
- **备份只增不减。** zinc 没有任何收缩手段：包不会被重写，所以删掉源目录也不会让已经装着它的包变小。真要省空间，只能新建一个输出目录重新备一遍，再把旧仓库删掉。
- **不导入旧的分卷布局。** `vol-*.7z.001` 这类归档不读取。请新建仓库，或把旧目录整体当归档留着。
- **计划任务需要真实终端。** `schtasks` 在受限环境（如沙箱）里跑不了。
- **任务里密码是明文。** 无人值守的必然代价；威胁模型是网盘内容扫描，不是定向攻击者。
- **`manifest.json` 必须随包上传。** 它记录备份时算出的哈希，`verify` 靠它在换机后识别传输损坏；丢了就没法校验历史批次。
- **`reindex` 依赖本地时区。** `7za l` 打印的是本地时间且不带偏移，重建时按当前时区解释。
- **`reindex --pack-start` / `--pack-end` 会用子集覆盖整份索引。** 范围外的包下次备份会被当成新增重压一遍。这两个参数是灾难恢复用的，日常直接跑 `zinc reindex <仓库>`。
- **符号链接和 junction 不会被扫描。** 源目录本身是 junction 时扫描结果为空（表现为永远“无新增”），源目录内部的 junction 子树也不进索引。请把 `--source` 配成真实路径。
- **强杀后 `.lock` 残留。** 仓库锁不会自动接管。进程被强杀后要手动删掉 `<output>\.lock`，否则之后每次备份都会等满 `lock_timeout_minutes`（默认 30 分钟）再失败。
- **索引缺失等于全量重压。** `index.tsv` 被删掉之后如果不去重建，下次备份会把它当空仓库，把所有文件重压一遍（不会漏数据，但会多传一遍）。丢了先跑一次 `zinc reindex <仓库>`，这一次重压就省了。**损坏**的索引是另一回事：那是硬错误，备份会停下，直到你跑 `zinc reindex`。
- **换密码后旧包打不开。** 旧包仍用旧密码加密。
- **只支持 Windows。** 64 位 Windows 10 或 11。

## 常见问题

| 现象 | 原因与处理 |
|---|---|
| 装完敲 `zinc` 提示不是内部或外部命令 | `PATH` 只对新开的终端生效；如果一直不行，跑一次 `zinc init --force` |
| 安装包报 `unable to create the directory ... is-XXXX.tmp` / `Error 5` | 安装器写不了 `%TEMP%`（被重定向到别的盘，或被安全软件挡住）。在 `cmd` 里先 `set TEMP=%LOCALAPPDATA%\Temp`、`set TMP=%LOCALAPPDATA%\Temp` 再启动安装包，或把这些变量指到可写目录 |
| 提示 `[警告] 找不到 7z.exe / 7za.exe` | 7-Zip 不在 zinc 会去找的位置；装一个，或者 `zinc config set 7zip_path <7z.exe 路径>` |
| 备份卡很久然后失败 | 残留的 `<output>\.lock`；手动删掉再跑一次 |
| 通知突然不弹了 | 跑一次 `zinc init --force`，它会重建快捷方式、注册表项和标记文件 |
| `restore --from` 说不认识这个包 | 用 `--check-list` 打印的包名；裸数字按整个仓库的包顺序编址，不是那个列表的编号 |
| 恢复出来不是想要的版本 | 先 `zinc restore <仓库> --only <路径> --check-list` 看版本分布，再用 `--from` 指定 |
| 某个文件备份不进去 | 被独占打开（例如 Excel）。它会被跳过且**不写索引**，下次备份自动重试 |

## 从源码构建

```powershell
git clone https://github.com/GensokyoNeoSanae/zinc.git
cd zinc
go build -o zinc.exe ./cmd/zinc
go build -ldflags "-s -w -H=windowsgui" -o zincw.exe ./cmd/zinc
go build -o zinc-register.exe ./cmd/install
go build -o zinc-unregister.exe ./cmd/uninstall
```

`build.ps1` 把上面四条包成一条命令。安装包是 [Inno Setup 6](https://jrsoftware.org/isinfo.php) 脚本 `installer/zinc.iss`；装一次 Inno Setup，构建脚本会在 `_tools\InnoSetup6`、`Program Files` 或 `-Iscc` 指定的位置找 `ISCC.exe`：

```powershell
.\build.ps1                          # 四个 zinc 可执行文件
.\installer\build.ps1 -Version 0.1.0 # dist\zinc-0.1.0-setup.exe
.\package.ps1 -Version 0.1.0         # 安装包 + 绿色版 zip
```

`package.ps1` 会调上面两个，产出 `dist\zinc-<版本>-setup.exe`（向导）和 `dist\zinc-<版本>-windows-amd64.zip`（绿色版）。安装包是单文件：Inno 把四个可执行文件、图标、文档和它自己的卸载程序一起压进去。第一页是语言选择，英文默认，简体中文是第二项。

`cmd/` 是四个入口，`internal/app` 是共用代码。Inno 脚本不自己实现注册：装完文件后调 `tools\zinc-register.exe`，卸载前调 `tools\zinc-unregister.exe`，所以注册逻辑只有一份。两个主程序只差子系统：`zinc.exe` 是控制台程序，手动跑能看输出；`zincw.exe` 是 GUI 子系统，计划任务执行时不闪窗口、不抢焦点。项目没有第三方依赖，`go.mod` 只声明了 Go 版本。

`zinc.ico` 是脚本生成的，不是手绘：改脚本顶部的颜色常量后跑 `python tools\make-icon.py`（需要 Pillow）。它按十个尺寸分别绘制 `Zn` 再自己拼 ICO，所以 16px 那一帧不是从 256px 缩下来的。

## 参与贡献

欢迎提 issue 和 PR。请保持改动聚焦，提交前跑一下 `gofmt`、`go vet`，并把四个 exe 都编译一遍。细节见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[MIT](LICENSE)。

## 致谢

zinc 的想法来自 [zbak](https://github.com/imaegoo/zbak)——一个更早的、基于 7-Zip 的 NAS 备份工具。它让人看到：把目录树打成一堆普通 7z 文件，几年后仍然能只靠 7-Zip 恢复。谢谢 [@iMaeGoo](https://github.com/imaegoo)。

也感谢 7-Zip 项目——真正干活的命令行是它——以及 Go 标准库，它是本项目唯一的依赖。
