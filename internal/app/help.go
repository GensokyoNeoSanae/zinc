package app

import (
	"fmt"
	"strings"
)

// HelpTopic 是一条命令的详细帮助。主面板（zinc --help）放不下这些文字，
// 所以 `zinc help <命令>` 单独展开：目的、原理、参数、注意事项。
type HelpTopic struct {
	Usage   string   // 用法行，不含 "zinc " 前缀
	Purpose string   // 指令目的：干什么用、什么时候用
	How     string   // 指令原理：内部怎么做，为什么这么设计
	Params  []string // 参数逐条说明，按重要性排，--password 永远最后
	Notes   []string // 注意事项、坑、与其它命令的关系
}

// render 把一条帮助排版成终端文本。节标题空一行、缩进两格，
// 让 `zinc help compact` 的输出能一眼扫出来。标题也要跟随语言。
func (t HelpTopic) render() string {
	var b strings.Builder
	b.WriteString(M.HelpUsage)
	b.WriteString(":\n  zinc ")
	b.WriteString(t.Usage)
	b.WriteString("\n")
	if t.Purpose != "" {
		b.WriteString("\n" + M.HelpPurpose + ":\n")
		b.WriteString(indent(t.Purpose, "  "))
	}
	if t.How != "" {
		b.WriteString("\n" + M.HelpHow + ":\n")
		b.WriteString(indent(t.How, "  "))
	}
	if len(t.Params) > 0 {
		b.WriteString("\n" + M.HelpParams + ":\n")
		for _, p := range t.Params {
			b.WriteString(indent(p, "  "))
		}
	}
	if len(t.Notes) > 0 {
		b.WriteString("\n" + M.HelpNotes + ":\n")
		for _, n := range t.Notes {
			b.WriteString(indent("- "+n, "  "))
		}
	}
	return b.String()
}

// indent 给每一行加前缀，保持多行段落整齐
func indent(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	var b strings.Builder
	for _, l := range lines {
		if l == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(prefix)
		b.WriteString(l)
		b.WriteString("\n")
	}
	return b.String()
}

// helpTopic 按当前语言取一条命令的详细帮助
func helpTopic(cmd string) (HelpTopic, bool) {
	m := zhHelp
	if M == enMessages {
		m = enHelp
	}
	t, ok := m[cmd]
	return t, ok
}

// printHelp 是 `zinc help <命令>` 的入口，展开那一条的详细说明。
func printHelp(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	t, ok := helpTopic(args[0])
	if !ok {
		die(M.HelpUnknownTopic, args[0])
	}
	fmt.Print(t.render())
}

var zhHelp = map[string]HelpTopic{
	"add": {
		Usage: "add <仓库名> --source <路径> [--output <路径>]",
		Purpose: `新建一个备份仓库：登记要备份的源目录，和存放包的位置。
仓库名是你之后所有命令用的代号，可以带中文。源目录必须是已存在的目录；输出目录
不存在会自动创建。`,
		How: `只写 repos.json，不碰磁盘上的任何数据。输出目录默认是 <root>\<仓库名>，
root 由 zinc root 设置。源目录当前不存在只会警告，不会拒绝——挂载盘还没插上时
也能先把仓库配好。`,
		Params: []string{
			"--source <路径>：要备份哪个目录。",
			"--output <路径>：包放哪儿。不给就用全局 root 下的同名子目录。",
		},
		Notes: []string{
			"同一个输出目录不能给两个仓库用，会直接拒绝。",
			"大小写不同的两个路径会被提示可能是同一个目录（Windows 上它们是）。",
		},
	},
	"remove": {
		Usage:   "remove <仓库名>",
		Purpose: `删掉一个仓库的配置。只动 repos.json，输出目录里的包和索引一个都不会删。`,
		How:     `确认后从配置里移除这个条目。如果它还有计划任务，会提示你去 unschedule。`,
		Params:  []string{"（无参数，仓库名是唯一的位置参数）"},
		Notes: []string{
			"想连数据一起删，得自己去输出目录手动删。",
			"删错了可以用 add 重新加回来，只要输出目录还在。",
		},
	},
	"set": {
		Usage: "set <仓库名> [--source <路径>] [--output <路径>] [--compress <0-9>] [--solid <块大小|off>] [--mhe on|off>] [--max-pack-byte <数量>] [--exclude <通配符>] [--password <密码>]",
		Purpose: `修改某个仓库的参数。只给想改的那几项，其余保持原样。这些参数优先级高于
config 里的同名全局设置。`,
		How: `写进 repos.json 的对应字段。压缩等级、固实块、头部加密只影响之后打的新包；
已经存在的包不会被重打——compact 不做转换器，它只处理有旧副本可丢的包。`,
		Params: []string{
			"--compress <0-9>：7z 压缩等级，0=仅存储。默认 0，抗损最好也最快。",
			"--solid <块大小|off>：固实块上限（如 64M）或 off。压缩等级为 0 时不生效。",
			"--mhe on|off：是否加密 7z 头部（隐藏包内文件名）。",
			"--max-pack-byte <数量>：单包上限，如 4G、500M。",
			"--exclude <通配符>：追加排除规则，可重复给。",
			"--password <密码>：这个仓库专用密码，覆盖全局默认。",
		},
		Notes: []string{
			"改密码不会重打旧包：旧包还是老密码，新旧混用时 restore 会提示打不开的包。",
			"改 source 不会让旧包失效，但之后扫描的是新目录。",
		},
	},
	"list": {
		Usage:   "list [--names]",
		Purpose: `列出所有仓库，以及它们的源目录和输出目录。`,
		How:     `读 repos.json，按名字排序输出。`,
		Params:  []string{"--names：每行只输出仓库名，给脚本和补全用。"},
		Notes:   []string{"没配任何仓库时输出一句提示，不是错误。"},
	},
	"root": {
		Usage: "root [<路径>]",
		Purpose: `查看或设置全局输出根目录：新建仓库时没给 --output，包就放在
<root>\<仓库名> 下。`,
		How:    `写在 repos.json 的 root 字段。给路径时会先校验是不是目录。`,
		Params: []string{"<路径>：要设置的目录。不给就是查看当前值。"},
		Notes:  []string{"已经存在的仓库不受影响，它们各自记着自己的输出目录。"},
	},
	"config": {
		Usage:   "config [get|set|unset|path|list|show] [<键>] [<值>]",
		Purpose: `读写全局设置（settings.json）。这些是没被仓库覆盖时的默认值。`,
		How: `get/set/unset 是单键操作；list 列出所有键和当前值；show 显示文件路径。
取值优先级永远是：命令行 > 仓库字段 > settings.json > 内置默认。`,
		Params: []string{
			"get <键>：读一个键。set <键> <值>：写一个键。unset <键>：删掉，回落到内置默认。",
			"path：打印 settings.json 的位置。",
		},
		Notes: []string{"布尔键接受 on/off/true/false/yes/no。"},
	},
	"verify": {
		Usage: "verify <仓库名|--all> [--sample <百分比>] [--password-only|--hash-only]",
		Purpose: `检查每个包是否还和账本记的一样：重算 SHA-256 比对，并用配置的密码试着打开它。

它是「事后体检」，跟 backup 打包时的自检不是一回事：那个只证明写进去的那一刻是对的，
verify 证明的是现在还是对的（磁盘位腐、网盘下载出错都靠它发现）。`,
		How: `读 manifest 里每个包的最新哈希，逐个算文件 SHA-256 比对。默认还会用密码
开一次包（只读归档头），确认密码没错、头没坏。

删除记录（.rm 旁注）也一起校验——它坏了恢复时不是少删就是删错。`,
		Params: []string{
			"--all：验证所有仓库。",
			"--sample <百分比>：只随机抽查这么多包，大仓库省时间。",
			"--password-only：只试密码，不算哈希（快）。",
			"--hash-only：只算哈希，不试密码。",
		},
		Notes: []string{
			"报「损坏」只说明文件变了，不一定真坏——但复算哈希不通过就不能信任它。",
			"verify 不改任何东西，是安全的只读操作。",
		},
	},
	"status": {
		Usage:   "status [--all] [<仓库名>]",
		Purpose: `看一个仓库的概况：多少包、总体积、最近一次备份、索引和账本的状态。`,
		How:     `读索引和 manifest，不读包内容，所以很快。`,
		Params:  []string{"--all：一次展开所有仓库。", "<仓库名>：只看一个。"},
		Notes:   []string{"它不校验数据完整性，那是 verify 的事。"},
	},
	"schedule": {
		Usage:   "schedule <仓库名|--all> --daily|--weekly <1-7>|--monthly <1-31> --at HH:MM[,HH:MM...] [--check] [--no-notify] [--password <密码>]",
		Purpose: `给某个仓库（或全部）建一个 Windows 计划任务，到点自动跑 backup。`,
		How: `调用 schtasks 创建任务，命令行里带的是 zinc backup 和你指定的密码。
计划任务本身不联网、不需要你在场，所以密码必须能提前解析出来——解析不到直接报错，
不会留一个到点才失败的坑。

一次可以给多个时间，用逗号分隔，会建多个任务。每月 29-31 号在某些月份不存在，
它会在创建前问你一次。`,
		Params: []string{
			"--daily / --weekly 1,3 / --monthly 1,15：周期。",
			"--at HH:MM[,HH:MM...]：执行时间，可多个。",
			"--check：只打印会创建哪些任务，不真建。",
			"--no-notify：任务里不带 --notify。",
			"--password <密码>：写进任务里，之后改密码要重建任务。",
		},
		Notes: []string{
			"密码会明文存在计划任务里，这是 schtasks 的机制，不是 zinc 的选择。",
			"zinc schedule --list 查看现有任务。",
		},
	},
	"unschedule": {
		Usage:   "unschedule [<仓库名>|--all]",
		Purpose: `删除计划任务。不带 --all 时删这个仓库的全部任务。`,
		How:     `用 PowerShell 查任务名再删（schtasks 的输出是 GBK，直接读会乱码）。`,
		Params:  []string{"--all：删所有 zinc 任务。"},
		Notes:   []string{"只删任务，不删仓库配置，也不删包。"},
	},
	"completion": {
		Usage:   "completion <shell> [--install]",
		Purpose: `生成 Tab 补全脚本，让终端能补全命令名、仓库名和参数。`,
		How:     `按命令表生成对应 shell 的脚本，写到 stdout；加 --install 则直接装进用户配置。`,
		Params: []string{
			"<shell>：powershell / bash / zsh / fish。",
			"--install：自动写进当前 shell 的启动配置（目前只对 PowerShell 实现）。",
		},
		Notes: []string{"装完要重开终端才生效：补全靠 dot-source 加载，当前进程不会重读配置。"},
	},
	"hash-compare": {
		Usage: "hash-compare --source1 <路径> --source2 <路径>\n       zinc hash-compare --source <路径> --hash-value <SHA-256>",
		Purpose: `手动比对哈希，不依赖仓库。两种用法：比两棵树/两个文件，或拿一个文件对
一个已知的 SHA-256。`,
		How: `第一种逐个文件算 SHA-256，输出不同 / 缺失 / 多出三类；第二种只算一个文件。
有差异时退出码是 1，方便脚本判断。`,
		Params: []string{
			"--source1 / --source2：两边必须同为文件或同为目录。",
			"--source / --hash-value：单文件对期望值，不能配目录。",
		},
		Notes: []string{"跟备份仓库无关，纯粹是附带的工具。"},
	},
	"init": {
		Usage: "init [--force]",
		Purpose: `一次性初始化：注册通知、把程序目录写进用户 PATH，让你在任意终端能直接敲
zinc。绿色版和手动解压的用户需要跑一次；用安装包装过的已经做过了。`,
		How: `写 HKCU 的 PATH 和通知注册，不碰任何备份数据。用 --force 可以重做一遍，
适合通知失效或 PATH 被清掉的情况。`,
		Params: []string{"--force：即使已经注册过也重做。"},
		Notes:  []string{"改完 PATH 要重开终端才生效。"},
	},
	"repair": {
		Usage: "repair <仓库名> [--pack-start <包名>] [--pack-end <包名>] [--rehash] [--check] [--password <密码>]",
		Purpose: `修复「元数据」与「磁盘上的包」对不上的情况。它不碰包的内容，只重写两个记账文件：

  index.tsv      —— 每个文件当前在哪个包、多大、什么时间（增量备份靠它判断谁变了）
  manifest.json  —— 每个包的 SHA-256，以及每次备份记下的删除清单

什么时候会需要它：

  - 你手动删掉了 verify 报损坏的那个包，账本里还留着它的登记；
  - 换机器/换目录，只从网盘下载了一部分包过来，账本却是全量的；
  - compact 跑到一半被中断，包已经换新、manifest 还记着旧哈希，verify 从此一直报损坏；
  - 手动把包改了名或挪了位置；
  - 杀毒软件隔离了某个包，或者 chkdsk 动了文件；
  - index.tsv 被误删、被截断。

包本身是唯一权威数据。这些元数据丢了都能从这个目录里重建出来。`,
		How: `按「便宜到贵」的顺序做三件事：

  1. 对账 manifest：扫描输出目录里的包和 .rm 删除记录，删掉账本里已经不存在
     的登记，补登盘上有、账上没有的孤儿包。这一步只做 stat，很快。
  2. 重建 index.tsv：逐个读包的归档头（不解压内容），重建「文件 → 当前状态」，
     并按删除记录划掉已经消失的文件。这一步要读每个包，包多时慢。
  3. --rehash：把每个包的 SHA-256 按磁盘实际内容重算一遍。这一步要读整个包，
     最慢，所以默认不做——它专治 compact 中断留下的旧哈希。

第 3 步是 repair 存在的核心理由：除了它，没有任何命令能把 manifest 里过时的
哈希改回来，verify 会永远报那个包损坏。`,
		Params: []string{
			"--pack-start <包名> / --pack-end <包名>：只重建这一段的索引，不全量读一遍。",
			"--rehash：重算每个已登记包的哈希。修 compact 中途失败留下的旧哈希。",
			"--check：只打印计划、不写盘。写之前想看一遍会改什么就用它。",
			"--password <密码>：读归档头需要密码；不给时按「仓库 > settings」取。",
		},
		Notes: []string{
			"它不修复包本身。包真的损坏了，只能靠别的副本或重新备份。",
			"重建 index.tsv 时旧索引会备份成 index.tsv.bak，覆写前不会直接丢。",
			"repair 会拿仓库锁，备份进行时无法同时修复。",
		},
	},
	"backup": {
		Usage: "backup <仓库名|--all> [--notify] [--notify-empty] [--no-verify] [--password <密码>]",
		Purpose: `把源目录里「新增或改动过」的文件打成一个或多个独立的加密 7z 包。

它不是镜像：仓库只增不减，每个包都是当时那一批变更的完整快照。删掉源目录里的
文件不会让仓库变小（用 compact 处理旧副本），改过的文件会以新包的形式再存一份。`,
		How: `先用 index.tsv 和磁盘实际状态对比，找出这个源目录里「新的、改过的、大小变了的」
文件——所以没变化时它几乎不花时间，输出「无新增」。

然后按 max_pack_byte（默认 4G）把这些文件切成若干包，逐包用 7za 打包、校验
（默认跑一次 7za t 确认 CRC）、算 SHA-256，再更新 manifest 与索引。

逐包提交是刻意的：每成功一个包就落账，后面的包出问题也不影响前面已经提交的。
一次备份跑几小时，中断时你至少拿到了中间那部分。

打包时读不到的文件（被占用、权限不足）会被跳过并在终端列出来，不算失败——下次
备份会重试。被占用的行数不设上限，几次重跑就能把清单清干净。`,
		Params: []string{
			"--all：备份所有仓库，任何一个失败不影响其余。",
			"--notify：完成后发 Windows 通知（默认取 settings.complete_notify）。",
			"--notify-empty：没有新增也发通知，定时任务里用来确认它还活着。",
			"--no-verify：跳过打包后的 7za t 自检。只建议在确认磁盘可靠时用，它能省掉一次全包读取。",
			"--password <密码>：不给时按「仓库 > settings」取，取不到才交互询问。",
		},
		Notes: []string{
			"备份是安全的：它只写输出目录，从不删改源目录里的东西。",
			"包名就是打包时刻，恢复按名字升序重放，所以时钟回拨会被显式修正并提示。",
			"源目录里的 junction 会被跳过（它会表现为「永远无新增」），这是已知限制。",
		},
	},
	"compact": {
		Usage: "compact <仓库名> [--pack <包名>] [--min-gain <阈值>] [--check] [--password <密码>]",
		Purpose: `释放输出目录占用的空间。它删掉的是包里「已被新版取代」或「已从源目录消失」的副本，
那些旧版本随之消失，磁盘占用才会真正降下来。

最典型的用法：你大批移动或重命名了图片、视频这类大体量文件之后。zinc 按路径追踪
文件，不追踪「同一个文件换了位置」——移动之后，旧路径记为被删除、新路径记为新增，
于是同一份内容在仓库里出现了两次（甚至更多）。备份没错，但空间被吃掉了。`,
		How: `它扫描输出目录里的删除记录（.rm 旁注），找出「含有待删内容的包」，把这些包解压、
删掉对应条目、再按同样的参数重新打包成同名压缩包。

这里有两个刻意的设计：

其一，包被清空时不删包，而是塞一个 0 字节的占位文件 placeholder_please_delete_this，
保证仍然能打出一个同名压缩包。原因是网盘：多数网盘在检测到目录变更时执行的是「覆盖
并保留历史版本」，同名覆盖才会把更新推到云端；直接删包的话，网盘那边只会认为你删了
一个文件，历史版本的语义就断了。

其二，packArgs 用的是与备份完全相同的参数（压缩等级、固实块、头部加密），所以重打
出来的包跟原来结构一致，恢复路径不需要任何特殊处理。`,
		Params: []string{
			"--pack <包名>：只处理指定的包，可给包名或序号。不给就是全仓库范围。",
			"--min-gain <阈值>：收益低于这个值就跳过重打，避免为一个包来回解压又打包。\n" +
				"  支持两种写法：百分比（5%）表示「能省下这个包大小的 5% 才动手」；\n" +
				"  绝对大小（8M）表示「至少省 8 MB 才动手」，K/M/G/T 都可以。\n" +
				"  不给就用 config 里的 compact_min_gain（默认 5%）。",
			"--check：只打印计划、不执行。想先看看会重打哪些包时用它。",
			"--password <密码>：不给时按「仓库 > settings」取，取不到会交互询问。",
		},
		Notes: []string{
			"重写是有代价的：解压加重新压缩的耗时与整个包的大小成正比，临时空间需要约两倍单包大小。",
			"重写会改变包文件的字节内容但保持文件名不变，所以需要把更新的包重新上传到网盘。",
			"它只处理「有删除记录可依」的包。单纯因为改压缩参数而想重打历史包，不适用。",
		},
	},
	"restore": {
		Usage: "restore <仓库名> <目标路径> [--only <相对路径>] [--from <包名>]\n" +
			"                 [--no-overwrite] [--no-delete] [--check] [--password <密码>]",
		Purpose: `把备份包里的内容还原到磁盘。默认还原整个仓库的全部文件，也可以只还原一个路径
（--only），或从某个包开始还原（--from）。

这是唯一会直接改动你磁盘上已有文件的命令，所以它在动手之前会把要做的事列一遍，
目标位置已经有文件或有删除要被重放时，会停下来问一次。

不建议直接恢复到源目录。更稳的做法是先恢复到一个空目录，用资源管理器逐项看一遍，
确认没问题再手动覆盖回去——源目录在此期间始终保持完好，恢复逻辑万一有意外也不会
伤到唯一的副本。`,
		How: `它按包名的时间顺序（升序）依次解压，后一个包覆盖前一个，最后得到「最新状态」。

删除是单独一层语义：每次备份不仅记录「有哪些文件」，还记录「哪些文件不见了」
（输出目录里的 .rm 旁注）。恢复时会按时间线重放这些删除——也就是说，如果你在
10 月 5 日删过某个文件，10 月 10 日恢复时，目标目录里那份文件也会被删掉。这才
是「还原到那个时间点」的完整含义。用 --no-delete 可以只补不删。`,
		Params: []string{
			"--only <相对路径>：只恢复这一个文件或目录。它同样会重放与该路径相关的删除记录，\n" +
				"  所以恢复出来的结果与整仓恢复一致，只是范围小。",
			"--from <包名>：只从这个包开始恢复，不给就从第一个包开始。",
			"--no-overwrite：目标已存在同名文件时保留它，不覆盖。",
			"--no-delete：不执行删除记录，目标目录里多出来的旧文件统统保留。",
			"--check：只打印将要执行的清单，不询问、不写盘；此时可以不写目标路径。",
			"--password <密码>：不给时按「仓库 > settings」取，取不到会交互询问。",
		},
		Notes: []string{
			"默认覆盖同名文件。这是恢复的预期语义，但意味着目标目录里同名的旧内容会消失。",
			"删除记录是备份时写下的旁注，不是实时状态；用 --check 先看一遍最稳。",
			"恢复过程中会拿仓库锁，备份进行时无法同时恢复。",
		},
	},
}

var enHelp = map[string]HelpTopic{
	"add": {
		Usage: "add <repo_name> --source <path> [--output <path>]",
		Purpose: `Create a repo: register which directory to back up and where the packs go.

The name is the handle every later command uses; it can contain any characters. The
source must already exist; the output directory is created if missing.`,
		How: `Writes repos.json only, touching nothing on disk. The output defaults to
<root>\<repo_name>, where root comes from zinc root. A missing source is only a
warning, not a refusal, so you can configure a repo before plugging the drive in.`,
		Params: []string{
			"--source <path>: the directory to back up.",
			"--output <path>: where the packs go; defaults to a subdirectory of the global root.",
		},
		Notes: []string{
			"The same output directory cannot be shared by two repos.",
			"Two paths differing only in case are flagged as probably the same directory (on Windows they are).",
		},
	},
	"remove": {
		Usage: "remove <repo_name>",
		Purpose: `Drop a repo definition. Only repos.json changes; the packs and index in the output
directory are left alone.`,
		How:    `Asks once, then removes the entry. If scheduled tasks reference it, it tells you to unschedule them.`,
		Params: []string{"(none; the repo name is the only positional argument)"},
		Notes: []string{
			"To delete the data too, remove the output directory by hand.",
			"Add the repo back with add if you removed it by mistake and the output directory still exists.",
		},
	},
	"set": {
		Usage: "set <repo_name> [--source <path>] [--output <path>] [--compress <0-9>] [--solid <block|off>] [--mhe on|off>] [--max-pack-byte <quantity>] [--exclude <glob>] [--password <password>]",
		Purpose: `Change one repo's settings. Pass only what you want to change; the rest stays as it is.
These override the same-named global settings in config.`,
		How: `Written to repos.json. Compression level, solid blocks and header encryption only affect
packs built afterwards; existing packs are not rebuilt — compact is not a converter, it
only touches packs that have old copies to drop.`,
		Params: []string{
			"--compress <0-9>: 7z level, 0 = store only. Default 0, fastest and most damage-tolerant.",
			"--solid <block|off>: solid block limit (e.g. 64M) or off. Ignored at level 0.",
			"--mhe on|off: encrypt the 7z header, hiding the file names inside.",
			"--max-pack-byte <quantity>: pack size limit, e.g. 4G, 500M.",
			"--exclude <glob>: add an exclusion rule; repeatable.",
			"--password <password>: repo-specific password, overriding the global default.",
		},
		Notes: []string{
			"Changing the password does not re-encrypt old packs; restore will report the ones it cannot open.",
			"Changing source does not invalidate old packs, but later scans look at the new directory.",
		},
	},
	"list": {
		Usage:   "list [--names]",
		Purpose: `List every repo with its source and output directories.`,
		How:     `Reads repos.json and prints in name order.`,
		Params:  []string{"--names: print only the names, one per line, for scripts and completion."},
		Notes:   []string{"An empty config prints a hint, not an error."},
	},
	"root": {
		Usage: "root [<path>]",
		Purpose: `Show or set the global output root: a new repo without --output puts its packs in
<root>\<repo_name>.`,
		How:    `Stored in repos.json. A path argument is validated as a directory first.`,
		Params: []string{"<path>: the directory to set. Omit to just print the current one."},
		Notes:  []string{"Existing repos keep their own output directories and are unaffected."},
	},
	"config": {
		Usage: "config [get|set|unset|path|list|show] [<key>] [<value>]",
		Purpose: `Read and write global settings (settings.json): the defaults used when a repo does not
override them.`,
		How: `get/set/unset work on one key; list prints every key and its current value; show prints the
file location. Precedence is always: command line > repo field > settings.json > built-in.`,
		Params: []string{
			"get <key>: read one key. set <key> <value>: write one. unset <key>: drop it, back to the built-in.",
			"path: print where settings.json lives.",
		},
		Notes: []string{"Boolean keys accept on/off/true/false/yes/no."},
	},
	"verify": {
		Usage: "verify <repo_name|--all> [--sample <percent>] [--password-only|--hash-only]",
		Purpose: `Check that every pack still matches the manifest: recompute its SHA-256, and try to open
it with the configured password.

This is the after-the-fact check. backup's own 7za t only proves the pack was correct
when written; verify proves it is correct now — which is how disk rot and a bad cloud
download get caught.`,
		How: `Reads the latest hash for each pack from the manifest and compares it to the file on
disk. By default it also opens each pack (reading only the header) to confirm the password
and the header are still fine.

Deletion records (.rm sidecars) are checked too: a broken one makes restore delete the
wrong things.`,
		Params: []string{
			"--all: verify every repo.",
			"--sample <percent>: check a random sample of packs, for big repos.",
			"--password-only: only test the password, skip hashing (fast).",
			"--hash-only: only hash, skip the password test.",
		},
		Notes: []string{
			"A corrupt report means the bytes changed; it may not be physically damaged, but it cannot be trusted.",
			"verify changes nothing; it is read-only.",
		},
	},
	"status": {
		Usage: "status [--all] [<repo_name>]",
		Purpose: `Show a repo at a glance: pack count, total size, the last backup, and the state of the
index and manifest.`,
		How:    `Reads the index and manifest without opening packs, so it is fast.`,
		Params: []string{"--all: expand every repo.", "<repo_name>: just one."},
		Notes:  []string{"It does not verify data integrity; that is verify's job."},
	},
	"schedule": {
		Usage: "schedule <repo_name|--all> --daily|--weekly <1-7>|--monthly <1-31> --at HH:MM[,HH:MM...] [--check] [--no-notify] [--password <password>]",
		Purpose: `Create a Windows scheduled task that runs backup for a repo (or all of them) at fixed
times.`,
		How: `Calls schtasks with a command line containing zinc backup and the password you resolved.
A scheduled task cannot ask questions later, so the password must resolve now — if it
cannot, this fails instead of leaving a task that breaks at 3 AM.

Several times can be given, comma separated, creating one task each. Days 29-31 do not
exist in every month, so it asks once before creating those.`,
		Params: []string{
			"--daily / --weekly 1,3 / --monthly 1,15: the period.",
			"--at HH:MM[,HH:MM...]: when to run, one or more times.",
			"--check: print the tasks it would create, create nothing.",
			"--no-notify: leave --notify out of the task.",
			"--password <password>: baked into the task; changing the password later means recreating it.",
		},
		Notes: []string{
			"The password is stored in plain text inside the task; that is schtasks, not a zinc choice.",
			"zinc schedule --list shows the existing tasks.",
		},
	},
	"unschedule": {
		Usage:   "unschedule [<repo_name>|--all]",
		Purpose: `Delete scheduled tasks. Without --all it removes every task of that repo.`,
		How:     `Queries task names with PowerShell, then deletes them (schtasks output is GBK and garbles in a UTF-8 read).`,
		Params:  []string{"--all: remove every zinc task."},
		Notes:   []string{"Only tasks are removed; the repo definition and the packs stay."},
	},
	"completion": {
		Usage:   "completion <shell> [--install]",
		Purpose: `Generate a tab-completion script so the shell completes command names, repo names and flags.`,
		How:     `Renders a script for the requested shell to stdout; --install writes it into the user profile instead.`,
		Params: []string{
			"<shell>: powershell / bash / zsh / fish.",
			"--install: write it into the current shell's startup config (PowerShell only for now).",
		},
		Notes: []string{"Reopen the terminal afterwards: completion is loaded by dot-sourcing, and the running shell will not reread its config."},
	},
	"hash-compare": {
		Usage: "hash-compare --source1 <path> --source2 <path>\n       zinc hash-compare --source <path> --hash-value <SHA-256>",
		Purpose: `Compare hashes by hand, without a repo. Either two files/trees against each other, or one
file against a known SHA-256.`,
		How: `The first form hashes every file and reports changed / missing / extra. The second hashes
one file. Any difference exits 1, so scripts can branch on it.`,
		Params: []string{
			"--source1 / --source2: both must be files, or both directories.",
			"--source / --hash-value: one file against an expected value; not valid for directories.",
		},
		Notes: []string{"Unrelated to the backup repos; it is a bundled utility."},
	},
	"init": {
		Usage: "init [--force]",
		Purpose: `One-time setup: register notifications and put the program directory on the user PATH so
zinc works in any terminal. Needed for the portable zip; the installer already did it.`,
		How: `Writes the HKCU PATH and the notification registration. It never touches backup data.
--force redoes it, for when notifications stopped working or PATH was cleared.`,
		Params: []string{"--force: run the setup again even if it was already done."},
		Notes:  []string{"Reopen the terminal after changing PATH."},
	},
	"repair": {
		Usage: "repair <repo_name> [--pack-start <pack_name>] [--pack-end <pack_name>] [--rehash] [--check] [--password <password>]",
		Purpose: `Fix the case where the metadata no longer matches the packs on disk. It never touches
pack content; it only rewrites two bookkeeping files:

  index.tsv      - which pack holds the current version of each file (incremental backup reads it)
  manifest.json  - the SHA-256 of every pack, plus the deletion records written by each backup

When you need it:

  - you deleted a pack that verify reported as corrupt, and the manifest still lists it;
  - you moved to another machine and only downloaded some of the packs from the cloud;
  - a compact was interrupted after replacing a pack but before updating the manifest, so
    verify has reported that pack as corrupt ever since;
  - you renamed or moved a pack by hand;
  - antivirus quarantined a pack, or chkdsk touched the directory;
  - index.tsv was deleted or truncated.

The packs are the only source of truth. Any of this metadata can be rebuilt from them.`,
		How: `Three steps, cheapest first:

  1. Reconcile the manifest: scan the packs and .rm deletion records in the output
     directory, drop entries whose file is gone, register packs the manifest never saw.
     This is just stat calls, so it is fast.
  2. Rebuild index.tsv: read each pack's archive header (no extraction), rebuild
     "file -> current state", and cross out files that the deletion records removed.
     This reads every pack, so it is slow on a large repo.
  3. --rehash: recompute each pack's SHA-256 from what is actually on disk. This reads
     whole packs, so it is slow and off by default.

Step 3 is the reason the command exists: nothing else can repair a stale manifest hash,
and without it verify reports that pack as corrupt forever.`,
		Params: []string{
			"--pack-start <pack_name> / --pack-end <pack_name>: rebuild only that slice of the index.",
			"--rehash: recompute every registered pack's hash; repairs a compact that died mid-way.",
			"--check: print the plan, write nothing.",
			"--password <password>: reading archive headers needs one; repo > settings, else prompt.",
		},
		Notes: []string{
			"It does not repair a damaged pack. Only another copy or a fresh backup can do that.",
			"The old index is kept as index.tsv.bak before being overwritten.",
			"repair takes the repo lock, so it cannot run while a backup is in progress.",
		},
	},
	"backup": {
		Usage: "backup <repo_name|--all> [--notify] [--notify-empty] [--no-verify] [--password <password>]",
		Purpose: `Pack the files that are new or changed in the source directory into one or more
independent encrypted 7z archives.

It is not a mirror: the repo only grows. Every pack is a complete snapshot of that
batch of changes. Deleting a file from the source does not shrink the repo (compact
handles old copies), and an edited file gets stored again as a new pack.`,
		How: `It compares index.tsv against the source directory to find files that are new, edited
or resized — so a run with nothing to do costs almost nothing and prints "no changes".

Those files are split into packs by max_pack_byte (4G by default), and each pack is
built with 7za, tested once (7za t, checking CRCs), hashed, then recorded in the
manifest and the index.

Committing pack by pack is deliberate: each finished pack is recorded immediately, so a
failure later in the run does not throw away the packs already written. A backup can run
for hours; you keep the middle part either way.

Files that cannot be read while packing (locked, permission denied) are skipped and
listed in the terminal instead of failing the run — the next backup retries them. The
list is not truncated, so a couple of reruns clear it out.`,
		Params: []string{
			"--all: back up every repo; one failure does not stop the others.",
			"--notify: toast when it finishes (default from settings.complete_notify).",
			"--notify-empty: toast even when nothing changed, useful to see a scheduled task alive.",
			"--no-verify: skip the post-pack 7za t test. It saves a full read of every pack; only use it when you trust the disk.",
			"--password <password>: repo > settings, and only prompt if neither has one.",
		},
		Notes: []string{
			"Backup is harmless: it only writes to the output directory, never to the source.",
			"Pack names are timestamps and restore replays them in name order, which is why a clock rollback is corrected and reported.",
			"Directory junctions inside the source are skipped (they look like \"nothing new\" forever); a known limitation.",
		},
	},
	"compact": {
		Usage: "compact <repo_name> [--pack <pack_name>] [--min-gain <threshold>] [--check] [--password <password>]",
		Purpose: `Reclaim disk space in the output directory. It drops the copies a newer version has
superseded, and the copies of files that no longer exist in the source. Those old
versions disappear, and only then does the space actually come back.

The typical case is moving or renaming a large batch of media files. zinc tracks files
by path, not by identity: after a move, the old path looks deleted and the new path looks
new, so the same content ends up stored twice (or more). The backup is correct, but the
space is doubled.`,
		How: `It scans the deletion records (.rm sidecars) in the output directory, finds the packs
that contain the doomed entries, extracts each one, removes those entries, and repacks it
as the same file name with the same packing parameters.

Two details are deliberate:

A pack that loses every entry is not deleted. Instead it gets a zero-byte placeholder
file named placeholder_please_delete_this, so the pack can still be rebuilt under the
same name. This matters for network drives: most of them treat a directory change as
"overwrite and keep a version history", so a same-name overwrite pushes the new bytes
to the cloud while a deletion would just look like a file vanished.

The repack uses the exact same arguments as a backup (compression level, solid block,
header encryption), so a compacted pack has the same structure as the original and
restoring needs no special handling.`,
		Params: []string{
			"--pack <pack_name>: limit the work to one pack, by name or index. Omit to scan the whole repo.",
			"--min-gain <threshold>: skip packs whose savings would be too small to justify the work.\n" +
				"  A percentage (5%) means \"only if it saves at least 5% of this pack\";\n" +
				"  an absolute size (8M) means \"only if it saves at least 8 MB\". K/M/G/T all work.\n" +
				"  Omit to use compact_min_gain from settings (default 5%).",
			"--check: print the plan and stop. Use it to see which packs would be rewritten.",
			"--password <password>: omit to use repo > settings; prompt if neither has one.",
		},
		Notes: []string{
			"Rewriting costs real time: extract plus recompress scales with the whole pack, and the temporary space needed is about twice one pack.",
			"The pack keeps its name but its bytes change, so the rewritten packs need uploading again.",
			"It only touches packs that have deletion records to act on. Repacking history just to change compression settings is not what this does.",
		},
	},
	"restore": {
		Usage: "restore <repo_name> <dest_path> [--only <rel_path>] [--from <pack_name>]\n" +
			"                 [--no-overwrite] [--no-delete] [--check] [--password <password>]",
		Purpose: `Put the contents of the packs back on disk. By default it restores everything in the repo;
--only narrows it to one path, and --from starts at a given pack.

This is the only command that touches files you already have, so it prints what it is
about to do first, and stops to ask when the destination already has files or when
deletions are about to be replayed.

Restoring straight into the source directory is not recommended. The safer route is to
restore into an empty directory, look through it in Explorer, and then copy it back by
hand — the source stays intact the whole time, so a surprise in the restore logic cannot
damage your only copy.`,
		How: `Packs are extracted in name order (ascending); each later pack overwrites the earlier
one, which leaves you with the latest state.

Deletion is a separate layer. Every backup records not just the files that exist but the
files that vanished (as .rm sidecars in the output directory). A restore replays those
deletions along the timeline: delete a file on Oct 5, restore on Oct 10, and that file
is removed from the destination as well. That is what "restore to that moment" means.
--no-delete keeps those files instead.`,
		Params: []string{
			"--only <rel_path>: restore just this file or directory. It still replays the deletions\n" +
				"  related to that path, so the result matches a full restore, only narrower.",
			"--from <pack_name>: start at this pack instead of the first one.",
			"--no-overwrite: keep an existing same-name file instead of overwriting it.",
			"--no-delete: do not replay deletions; keep the extra old files in the destination.",
			"--check: print the plan, ask nothing, write nothing; the destination may be omitted.",
			"--password <password>: omit to use repo > settings; prompt if neither has one.",
		},
		Notes: []string{
			"Same-name files are overwritten by default. That is what restore means, but it also means old content in the destination disappears.",
			"Deletion records are notes written at backup time, not live state; use --check when unsure.",
			"Restore takes the repo lock, so it cannot run while a backup is in progress.",
		},
	},
}
