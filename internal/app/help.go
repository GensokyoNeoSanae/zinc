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

// render 把一条帮助排版成终端文本。正文按终端宽度自动折行、整段缩进两格，
// 所以 help 文本里只管写连续的一段，不要手工折行；空行才是分段。
func (t HelpTopic) render() string {
	width := termWidth()
	var b strings.Builder
	b.WriteString(M.HelpUsage)
	b.WriteString(":\n  zinc ")
	b.WriteString(t.Usage)
	b.WriteString("\n")
	if t.Purpose != "" {
		b.WriteString("\n" + M.HelpPurpose + ":\n")
		b.WriteString(wrapBlock(t.Purpose, "  ", "  ", width))
	}
	if t.How != "" {
		b.WriteString("\n" + M.HelpHow + ":\n")
		b.WriteString(wrapBlock(t.How, "  ", "  ", width))
	}
	if len(t.Params) > 0 {
		b.WriteString("\n" + M.HelpParams + ":\n")
		for _, p := range t.Params {
			b.WriteString(wrapBlock(p, "  ", "  ", width))
		}
	}
	if len(t.Notes) > 0 {
		b.WriteString("\n" + M.HelpNotes + ":\n")
		for _, n := range t.Notes {
			b.WriteString(wrapBlock(n, "  - ", "    ", width))
		}
	}
	return b.String()
}

// wrapBlock 把一个段落按终端宽度折行，首行用 first 前缀、续行用 rest。
// 空行原样保留，用来分段。这是「Word 里整段缩进」的效果。
func wrapBlock(s, first, rest string, width int) string {
	var b strings.Builder
	limit := width - displayWidth(rest)
	if limit < 20 {
		limit = 20
	}
	firstLine := true
	for _, para := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if strings.TrimSpace(para) == "" {
			b.WriteString("\n")
			firstLine = true
			continue
		}
		for _, line := range wrapLine(para, limit) {
			if firstLine {
				b.WriteString(first)
				firstLine = false
			} else {
				b.WriteString(rest)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// isWordRune 判断字符是否属于一个英文单词（含点号，manifest.json 要当成一个整体）
func isWordRune(r rune) bool {
	return r < 128 && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_')
}

// noLineStart 是不能出现在行首的字符（中文排版禁则）。
const noLineStart = "，。、；：？！）〕】」』’”%.,;:?!)"

// wrapLine 按显示宽度折行：CJK 可以在任意字符处断开，ASCII 尽量在空格处断，
// 免得把一个单词劈成两半。
func wrapLine(s string, limit int) []string {
	var out []string
	cur := make([]rune, 0, len(s))
	curW := 0

	for _, r := range s {
		w := runeWidth(r)
		if curW+w > limit && len(cur) > 0 {
			// 优先退到最后一个空格断行，没有空格才硬断
			cut := len(cur)
			for i := len(cur) - 1; i > 0; i-- {
				if cur[i] == ' ' {
					cut = i
					break
				}
			}
			// 中文段落里空格很稀疏，最后一个空格可能在很远的前面。为它回退
			// 会把这一行砍掉一大半（「写进 repos.json」然后换行），不如就在
			// 当前位置硬断——中文本来就可以在任意两个字之间断开。
			if len(cur)-cut > 8 {
				cut = len(cur)
			}
			out = append(out, strings.TrimRight(string(cur[:cut]), " "))
			rest := append([]rune(nil), cur[cut:]...)
			if len(rest) > 0 && rest[0] == ' ' {
				rest = rest[1:] // 断点处的空格不带到下一行开头
			}
			// 别把英文单词劈成两半（manifest.json 变成 manifest.j / son）。
			// 只在硬断时需要处理：软断本来就落在空格上，词是完整的，再挪一次
			// 反而会把它和下一行粘起来（disk.It）。
			if len(rest) == 0 && isWordRune(r) && len(out) > 0 {
				last := []rune(out[len(out)-1])
				k := len(last)
				for k > 0 && isWordRune(last[k-1]) {
					k--
				}
				if k < len(last) {
					moved := []rune(string(last[k:]))
					out[len(out)-1] = strings.TrimRight(string(last[:k]), " ")
					rest = moved
				}
			}
			// 中文排版禁则：标点不落行首。宁可让上一行多占一格，
			// 也不要出现「…大小」/「，不用打开文件」这种断法。
			for len(rest) > 0 && strings.ContainsRune(noLineStart, rest[0]) {
				out[len(out)-1] += string(rest[0])
				rest = rest[1:]
			}
			cur = rest
			curW = 0
			for _, c := range cur {
				curW += runeWidth(c)
			}
			// 硬断时 rest 是空的，当前字符会直接开新行——如果它本身是标点，
			// 补到上一行末尾，别让新行以「」」或「，」开头。
			if len(cur) == 0 && strings.ContainsRune(noLineStart, r) && len(out) > 0 {
				out[len(out)-1] += string(r)
				continue
			}
		}
		cur = append(cur, r)
		curW += w
	}
	if len(cur) > 0 {
		out = append(out, strings.TrimRight(string(cur), " "))
	}
	return out
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
		Usage:   "add <仓库名> --source <路径> [--output <路径>]",
		Purpose: `新建一个备份仓库：登记要备份的源目录和存放包的位置。它只是登记，不会产生任何包——真正备份要跑 zinc backup <仓库名>。仓库名是之后所有命令用的代号，可带中文。输出目录不存在时会自动创建。`,
		How:     `只写 repos.json，不碰磁盘上的任何数据。输出目录默认是 <root>\<仓库名>，root 由 zinc root 设置。源目录当前不存在只会警告，不会拒绝——挂载盘还没插上时也能先把仓库配好。`,
		Params: []string{
			"--source <路径>：要备份哪个目录。",
			"--output <路径>：包放哪儿。不给就用全局 root 下的同名子目录。",
		},
		Notes: []string{
			"同一个输出目录不能给两个仓库用，会直接拒绝。",
			"大小写不同的两个路径会被提示可能是同一个目录（至少在Windows 上是这样的）。",
		},
	},
	"remove": {
		Usage:   "remove <仓库名>",
		Purpose: `删掉一个仓库的配置。只动 repos.json，不会删除输出目录里的包和 index.tsv（索引）。`,
		How:     `确认后从配置里移除这个条目。如果它还有计划任务，会提示你去使用指令 unschedule。`,
		Params:  []string{"（无参数，仓库名是唯一的位置参数）"},
		Notes: []string{
			"想连数据一起删，得自己去输出目录手动删。",
			"删错了可以用指令 add 重新加回来，只要输出目录还在。",
		},
	},
	"set": {
		Usage:   "set <仓库名> [--source <路径>] [--output <路径>] [--compress <0-9>] [--solid <块大小|off>] [--mhe on|off>] [--max-pack-byte <数量>] [--exclude <通配符>] [--password <密码>]",
		Purpose: `修改某个仓库的参数，只改动给出的项。仓库参数优先于 config 里的同名全局设置。`,
		How:     `写进 repos.json 的对应字段。这几项只对之后新打的包生效：已经存在的包不会因为改了参数就被重新压缩一遍。compact 也不做这种转换——它只会把包里「已经被新版取代」的旧副本删掉，再按同样的参数把剩下的内容重新打包，目的是省空间，不是换格式。`,
		Params: []string{
			"--compress <0-9>：7z 压缩等级，0=仅存储。默认 0，抗损最好也最快。",
			"--solid <块大小|off>：固实块上限（如 64M）或 off。压缩等级为 0 时不生效。",
			"--mhe on|off：是否加密 7z 头部（隐藏包内文件名）。",
			"--max-pack-byte <数量>：单包上限，如 4G、500M。",
			"--exclude <通配符>：追加排除规则，可重复添加。",
			"--password <密码>：这个仓库的专用密码，覆盖全局默认。",
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
		Usage:   "root [<路径>]",
		Purpose: `查看或设置全局输出根目录。新建仓库没给 --output 时，包放在 <root>\<仓库名> 下。`,
		How:     `写在 repos.json 的 root 字段。给路径时会先校验是不是目录。`,
		Params:  []string{"<路径>：要设置的目录。不给就是查看当前值。"},
		Notes:   []string{"已经存在的仓库不受影响，它们各自记着自己的输出目录。"},
	},
	"config": {
		Usage:   "config [get|set|unset|path|list|show] [<键>] [<值>]",
		Purpose: `读写全局设置（settings.json）。这些是没被仓库覆盖时的默认值。`,
		How:     `get/set/unset 是单键操作；list 列出所有键和当前值；show 显示文件路径。取值优先级永远是：命令行 > 仓库字段 > settings.json > 内置默认。`,
		Params: []string{
			"get <键>：读一个键。set <键> <值>：写一个键。unset <键>：删掉，回落到内置默认。",
			"path：打印 settings.json 的位置。",
			"可用的键（默认值）：default_password（空，强制非空）、restore_overwrite（true）、restore_wrap_folder（true）、max_pack_byte（4G）、compression_level（0）、solid_block（64M）、header_encryption（true）、confirm_tips（true）、7zip_path（自动检测）、lang（自动）、lock_timeout_minutes（30）、exclude（空）、complete_notify（true）、compact_min_gain（5%）。",
			"default_password：全局默认密码。zinc 强制每个包都有密码，这个键不能留空——仓库可以用 zinc set <仓库名> --password 单独指定。",
			"restore_overwrite：恢复时是否覆盖同名文件。默认 true；仓库级的 --no-overwrite 只影响那一次。",
			"restore_wrap_folder：恢复时是否在目标下多套一层 restore-<仓库名>-<时间戳>。默认 true，写进新目录没有覆盖风险，所以不询问；设成 false 才会往目标目录里直接写，并在目录非空时问 y/N。",
			"max_pack_byte：单个包的上限，如 4G、500M。",
			"compression_level：7z 压缩等级 0-9，0 = 仅存储。",
			"solid_block：固实块大小，如 64M；off = 每文件独立块。压缩等级 0 时不生效。",
			"header_encryption：是否加密 7z 头部（-mhe=on），隐藏包内文件名。",
			"confirm_tips：危险操作执行前是否询问 y/N。脚本里可以设 false。",
			"7zip_path：7z.exe / 7za.exe 的路径。留空 = 在 PATH 和常见安装目录里找。",
			"lang：界面语言 zh / en。空 = 跟随 Windows。",
			"lock_timeout_minutes：仓库锁最长等待分钟数。",
			"exclude：全局排除规则，glob 写法，可重复。",
			"complete_notify：备份完成后是否弹系统通知。",
			"compact_min_gain：compact 的最小收益阈值。裸数字表示百分比（5 = 5%），带单位表示绝对大小（8M = 8 MB）。",
		},
		Notes: []string{"布尔键接受 on/off/true/false/yes/no。"},
	},
	"verify": {
		Usage:   "verify <仓库名|--all> [--sample <百分比>] [--password-only|--hash-only]",
		Purpose: `检查每个包（包括.rm 文件，即删除记录）是否在创建后变得不一致或缺失，用于发现随机位翻转、网盘下载出错等意外情况。另外还检验包的密码是否与默认密码相同。`,
		How:     `读 manifest.json（账本）里每个包的最新哈希，逐个算文件 SHA-256 比对。默认还会用密码开一次包（只读归档头），确认密码没错、头没坏。`,
		Params: []string{
			"--all：验证所有仓库。",
			"--sample <百分比>：只随机抽查这么多包，大仓库省时间。",
			"--password-only：只试密码，不算哈希（快）。",
			"--hash-only：只算哈希，不试密码。",
		},
		Notes: []string{
			"报「不一致」只说明包内容和账本记的哈希对不上，不一定真坏——但复算哈希不通过就不能信任它。",
			"verify 不改任何东西，是安全的只读操作。",
		},
	},
	"status": {
		Usage:   "status [<仓库名>] [--expand] [--sort <字段>] [--asc|--desc]",
		Purpose: `查看仓库状态。不带参数时每行一个仓库，含文件数、批次、包数、源与输出目录的实际占用、最后备份时间；带仓库名时展开详情，含待备份的新增、修改、删除数量。`,
		How:     `简表的数字来自 index.tsv（索引）和 manifest.json（账本）；「源」和「输出」是两边目录的真实遍历——目录项自带大小，不用打开文件，30 GB 量级也就百来毫秒。「输出比源多」是正常的：多出来的是历史版本、7z 头开销，以及 compact 还没清掉的旧副本。按名称排序走中文 collation（拼音），不是字节序；按大小排序默认大的在前。`,
		Params: []string{
			"<仓库名>：只看一个，展开成详情。",
			"--expand：每个仓库都展开详情，不再只给一行。",
			"--sort <字段>：name / src / out / backup，分别按名称、源大小、输出大小、最后备份时间排。",
			"--asc / --desc：覆盖默认方向（name 和 backup 升序，src 和 out 降序）。",
		},
		Notes: []string{
			"「源 / 输出」是两个目录的真实大小，跟压缩率无关。",
			"旧副本会一直计入「输出」，直到跑 compact 才真正消失。",
			"它不校验数据完整性，那是 verify 的事。",
		},
	},
	"schedule": {
		Usage:   "schedule <仓库名|--all> --daily|--weekly <1-7>|--monthly <1-31> --at HH:MM[,HH:MM...] [--check] [--no-notify] [--password <密码>]",
		Purpose: `给某个仓库（或全部）创建 Windows 计划任务，到点自动执行 backup。`,
		How: `调用 schtasks 创建任务，命令行里带的是 zinc backup 和你指定的密码。计划任务本身不联网、不需要你在场，所以密码必须能提前解析出来——解析不到直接报错，不会留一个到点才失败的坑。

一次可以给多个时间，用逗号分隔，会建多个任务。每月 29-31 号在某些月份不存在，它会在创建前问你一次。`,
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
		Usage:   "hash-compare --source1 <路径> --source2 <路径>\n       zinc hash-compare --source <路径> --hash-value <SHA-256>",
		Purpose: `手动比对哈希，不依赖仓库。两种用法：比两棵树/两个文件，或拿一个文件对一个已知的 SHA-256。`,
		How:     `第一种逐个文件算 SHA-256，输出不同 / 缺失 / 多出三类；第二种只算一个文件。有差异时退出码是 1，方便脚本判断。`,
		Params: []string{
			"--source1 / --source2：两边必须同为文件或同为目录。",
			"--source / --hash-value：单文件对期望值，不能配目录。",
		},
		Notes: []string{"跟备份仓库无关，纯粹是附带的工具。"},
	},
	"init": {
		Usage:   "init [--force]",
		Purpose: `一次性初始化：注册通知，并把程序目录写进用户 PATH，使 zinc 能在任意终端直接运行。绿色版和手动解压的用户需要跑一次；安装包装过的已经做过。`,
		How:     `写进当前用户的 PATH（注册表里的用户环境变量）并注册通知，不碰任何备份数据。用 --force 可以重做一遍，适合通知失效或 PATH 被清掉的情况。`,
		Params:  []string{"--force：即使已经注册过也重做。"},
		Notes:   []string{"改完 PATH 要重开终端才生效。"},
	},
	"fixpack": {
		Usage:   "fixpack <仓库名> [--check] [--password <密码>]",
		Purpose: `修复损坏的包：保留还能读取的文件，为读不出来的文件找替代来源，然后重新打包成同名文件。它只处理包内容，不修改 manifest.json（账本）。`,
		How:     `先对每个包跑一遍 7z t，拿到坏掉的文件名（非固实包下 7-Zip 会逐文件报出具体是哪个）。对每个坏文件按优先级找替代：先看源目录（还在、大小和时间跟 index.tsv 记的一样，就是那一份），再按从新到旧的顺序翻其他历史包，并且用 7z t 确认那份没坏。全都没有就标记为无法恢复。确认后逐个重写坏包：解压出好文件、补上替代来源、按原参数重新打包。`,
		Params: []string{
			"<仓库名>：要修哪个仓库。",
			"--check：只检查并列出计划，不动任何文件。",
			"--password <密码>：不给时按「仓库 > settings」取，取不到会交互询问。",
		},
		Notes: []string{
			"非固实包坏了不等于全废：7z t 只报真正坏掉的文件，其余文件仍能正常提取。",
			"没有任何替代来源的文件会被跳过，那个包保留原样不动——它可能是你最后一份数据。",
			"重打包后账本会重建，不需要再手动跑 fixmeta。",
			"zinc 只能确认源文件在、大小和时间对得上，不能确认它本身没坏。",
		},
	},
	"fixmeta": {
		Usage: "fixmeta <仓库名> [--pack-start <包名>] [--pack-end <包名>] [--rehash] [--check] [--password <密码>]",
		Purpose: `修复两个记账文件和磁盘上的包对不上的情况。index.tsv（索引）记录每个文件当前在哪个包、多大、什么时间；manifest.json（账本）记录每个包的 SHA-256 以及每次备份记下的删除清单。它不修改包内容。

常见原因：verify 报不一致的包被手动删除、换机器后只下载了部分包、compact 中断导致账本仍是旧哈希、包被改名或移动、杀毒软件隔离、index.tsv 被删除或截断。这些元数据都能从输出目录重建。`,
		How: `按「便宜到贵」的顺序做三件事。第一，对账 manifest：扫描输出目录里的包和 .rm 删除记录，删掉账本里已经不存在 的登记，补登盘上有、账上没有的孤儿包，这一步只做 stat，很快。第二，重建 index.tsv：逐个读包的归档头（不解压内容），重建「文件 → 当前状态」，并按删除记录划掉已经消失的文件，这一步要读每个包，包多时慢。第三，--rehash：把每个包的 SHA-256 按磁盘实际内容重算一遍，最慢，所以默认不做——它专治 compact 中断留下的旧哈希。

第 3 步是 fixmeta 存在的核心理由：除了它，没有任何命令能把 manifest 里过时的哈希改回来，verify 会永远把那个包报成「不一致」。`,
		Params: []string{
			"--pack-start <包名> / --pack-end <包名>：只重建这一段的索引，不全量读一遍。",
			"--rehash：重算每个已登记包的哈希。修 compact 中途失败留下的旧哈希。",
			"--check：只打印计划、不写盘。写之前想看一遍会改什么就用它。",
			"--password <密码>：读归档头需要密码；不给时按「仓库 > settings」取。",
		},
		Notes: []string{
			"它不修复包本身。包真的损坏了，只能靠别的副本或重新备份。",
			"重建 index.tsv 时旧索引会备份成 index.tsv.bak，覆写前不会直接丢。",
			"fixmeta 会拿仓库锁，备份进行时无法同时修复。",
		},
	},
	"backup": {
		Usage:   "backup <仓库名|--all> [--notify] [--notify-empty] [--no-verify] [--password <密码>]",
		Purpose: `把源目录里新增或改动过的文件打成一批独立的加密 7z 包。这是增量备份：每个包只包含那一批变更，已备份且未改动的文件不会重复打包。删掉源目录里的文件不会让仓库变小，旧副本要靠 compact 清理。`,
		How: `先用 index.tsv 和磁盘实际状态对比，找出这个源目录里「新的、改过的、大小变了的」文件——所以没变化时它几乎不花时间，输出「无新增」。

然后按 max_pack_byte（默认 4G）把这些文件切成若干包，逐包用 7za 打包、校验（默认跑一次 7za t 确认 CRC）、算 SHA-256，再把结果写回 manifest.json（账本）和 index.tsv（索引）。

逐包提交是刻意的：每成功一个包就落账，后面的包出问题也不影响前面已经提交的。一次备份跑几小时，中断时你至少拿到了中间那部分。

打包时读不到的文件（被占用、权限不足）会被跳过并在终端列出来，不算失败——下次备份会重试。被占用的行数不设上限，几次重跑就能把清单清干净。`,
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
		Usage:   "compact <仓库名> [--pack <包名>] [--min-gain <阈值>] [--check] [--password <密码>]",
		Purpose: `释放输出目录占用的空间：删掉包里已被新版取代、或对应文件已从源目录消失的副本，然后重写这些包。文件大批移动或重命名之后最有用——zinc 按路径追踪文件，移动后旧路径记为删除、新路径记为新增，同一份内容会在仓库里存两份。`,
		How: `它扫描输出目录里的删除记录（.rm 文件，每次备份把「源目录里消失的文件」记在里面），找出「含有待删内容的包」，把这些包解压、删掉对应条目、再按同样的参数重新打包成同名压缩包。

这里有两个刻意的设计：

其一，包被清空时不删包，而是塞一个 0 字节的占位文件 placeholder_please_delete_this，保证仍然能打出一个同名压缩包。原因是网盘：多数网盘在检测到目录变更时执行的是「覆盖并保留历史版本」，同名覆盖才会把更新推到云端；直接删包的话，网盘那边只会认为你删了一个文件，历史版本的语义就断了。

其二，重新打包时用的参数和备份时完全一样（压缩等级、固实块、头部加密），所以重打出来的包结构和原来一致，恢复流程不需要任何特殊处理。`,
		Params: []string{
			"--pack <包名>：只处理指定的包，可给包名或序号。不给就是全仓库范围。",
			"--min-gain <阈值>：收益低于这个值就跳过重打，避免为一个包来回解压又打包。\n" +
				"  裸数字表示百分比：5 = 至少省下这个包大小的 5% 才动手。\n" +
				"  带单位表示绝对大小：8M = 至少省 8 MB 才动手，K/M/G/T 都可以。\n" +
				"  百分号不认，写成 5% 会被当成绝对值。不给就用 config 里的 compact_min_gain（默认 5%）。",
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
		Usage: "restore <仓库名> --dest <路径> [--only <相对路径>] [--from <包名>]\n" +
			"                 [--no-overwrite] [--no-delete] [--check] [--password <密码>]",
		Purpose: `把备份包里的内容还原到磁盘。默认还原整个仓库；--only 只还原一个路径，--from 只解压指定的那一个包。

这是唯一会直接改动目标目录的命令。目标目录非空、或要重放删除记录时会询问一次；但如果 restore_wrap_folder 为 true（默认），zinc 会在目标下多套一层 restore-<仓库名>-<时间戳>，写进新目录没有覆盖风险，就不再询问。不建议直接恢复到源目录：先恢复到空目录、确认无误后再手动覆盖回去，源目录在此期间保持完好。`,
		How: `默认按包名的时间顺序（升序）依次解压，后一个包覆盖前一个，最后得到「最新状态」。

--from 是这条规则的例外：它只解压指定的那一个包，既不往后累积，也不重放删除记录。所以它表达的是「从历史里取某个时刻的那一份」，不是「从这个包开始往后恢复」。它必须配合 --only 使用。

删除是单独一层语义：每次备份不仅记录「有哪些文件」，还记录「哪些文件不见了」（输出目录里与包同名的 .rm 文件）。恢复时会按时间线重放这些删除——也就是说，如果你在 10 月 5 日删过某个文件，10 月 10 日恢复时，目标目录里那份文件也会被删掉。这才是「还原到那个时间点」的完整含义。用 --no-delete 可以只补不删。`,
		Params: []string{
			"--dest <路径>：恢复到哪个目录。必填，只有 --check 时可以省。",
			"--only <相对路径>：只恢复这一个文件或目录。它同样会重放与该路径相关的删除记录，所以结果与整仓恢复一致，只是范围小。",
			"--from <包名>：只解压指定的这一个包，不往后累积，也不重放删除记录。必须配合 --only 使用。",
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
		Usage:   "add <repo_name> --source <path> [--output <path>]",
		Purpose: `Create a repo: register which directory to back up and where the packs go. This only registers paths and produces no packs — run zinc backup <repo_name> to actually back things up. The repo name is the identifier every later command uses; it may contain Chinese. The output directory is created if missing.`,
		How:     `Writes repos.json only, touching nothing on disk. The output defaults to<root>\<repo_name>, where root comes from zinc root. A missing source is only awarning, not a refusal, so you can configure a repo before plugging the drive in.`,
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
		Usage:   "remove <repo_name>",
		Purpose: `Drop a repo definition. Only repos.json changes; the packs and index.tsv (the index) in the output directory are left alone.`,
		How:     `Asks once, then removes the entry. If scheduled tasks reference it, it tells you to unschedule them.`,
		Params:  []string{"(none; the repo name is the only positional argument)"},
		Notes: []string{
			"To delete the data too, remove the output directory by hand.",
			"Add the repo back with add if you removed it by mistake and the output directory still exists.",
		},
	},
	"set": {
		Usage:   "set <repo_name> [--source <path>] [--output <path>] [--compress <0-9>] [--solid <block|off>] [--mhe on|off>] [--max-pack-byte <quantity>] [--exclude <glob>] [--password <password>]",
		Purpose: `Change one repo's settings, passing only what you want to change. Repo settings override the same-named global settings in config.`,
		How:     `Written to repos.json. These settings only affect packs built afterwards: an existing pack is never recompressed because a setting changed. compact does not convert them either — it drops the copies a newer version has superseded and repacks the rest with the same settings, to save space rather than to change the format.`,
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
		Usage:   "root [<path>]",
		Purpose: `Show or set the global output root. A new repo without --output puts its packs in <root>\<repo_name>.`,
		How:     `Stored in repos.json. A path argument is validated as a directory first.`,
		Params:  []string{"<path>: the directory to set. Omit to just print the current one."},
		Notes:   []string{"Existing repos keep their own output directories and are unaffected."},
	},
	"config": {
		Usage:   "config [get|set|unset|path|list|show] [<key>] [<value>]",
		Purpose: `Read and write global settings (settings.json): the defaults used when a repo does not override them.`,
		How:     `get/set/unset work on one key; list prints every key and its current value; show prints thefile location. Precedence is always: command line > repo field > settings.json > built-in.`,
		Params: []string{
			"get <key>: read one key. set <key> <value>: write one. unset <key>: drop it, back to the built-in.",
			"path: print where settings.json lives.",
		},
		Notes: []string{"Boolean keys accept on/off/true/false/yes/no."},
	},
	"verify": {
		Usage:   "verify <repo_name|--all> [--sample <percent>] [--password-only|--hash-only]",
		Purpose: `Check every pack (including the .rm deletion records) for changes or missing files since it was created, to catch bit rot, a bad cloud download and similar accidents. Also checks that each pack's password matches the configured one.`,
		How: `Reads the latest hash for each pack from manifest.json (the ledger) and compares it to the file on disk. By default it also opens each pack (reading only the header) to confirm the passwordand the header are still fine.

Deletion records (.rm sidecars) are checked too: a broken one makes restore delete thewrong things.`,
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
		Usage:   "status [<repo_name>] [--expand] [--sort <field>] [--asc|--desc]",
		Purpose: `Show repo status. With no argument, one line per repo: file count, batches, packs, the real footprint of the source and output directories, and the last backup. With a repo name, the detail view, including how many files are pending.`,
		How:     `The table numbers come from index.tsv (the index) and manifest.json (the ledger); the source and output figures are real directory walks. Directory entries carry their size, so nothing is opened and 30 GB takes a few hundred milliseconds. Output being larger than the source is normal: it holds history, 7z headers, and old copies compact has not cleared yet. Sorting by name uses Chinese collation (pinyin), not byte order; sorting by size puts the largest first.`,
		Params: []string{
			"<repo_name>: show that repo alone, expanded.",
			"--expand: expand every repo instead of one line each.",
			"--sort <field>: name / src / out / backup.",
			"--asc / --desc: override the default direction (name and backup ascending, src and out descending).",
		},
		Notes: []string{
			"The source/output sizes are real, and unrelated to the compression ratio.",
			"Old copies keep counting toward the output until compact clears them.",
			"It does not verify data integrity; that is verify's job.",
		},
	},
	"schedule": {
		Usage:   "schedule <repo_name|--all> --daily|--weekly <1-7>|--monthly <1-31> --at HH:MM[,HH:MM...] [--check] [--no-notify] [--password <password>]",
		Purpose: `Create Windows scheduled tasks that run backup for a repo (or all of them) at fixed times.`,
		How: `Calls schtasks with a command line containing zinc backup and the password you resolved.A scheduled task cannot ask questions later, so the password must resolve now — if itcannot, this fails instead of leaving a task that breaks at 3 AM.

Several times can be given, comma separated, creating one task each. Days 29-31 do notexist in every month, so it asks once before creating those.`,
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
		Usage:   "hash-compare --source1 <path> --source2 <path>\n       zinc hash-compare --source <path> --hash-value <SHA-256>",
		Purpose: `Compare hashes by hand, without a repo. Either two files/trees against each other, or onefile against a known SHA-256.`,
		How:     `The first form hashes every file and reports changed / missing / extra. The second hashesone file. Any difference exits 1, so scripts can branch on it.`,
		Params: []string{
			"--source1 / --source2: both must be files, or both directories.",
			"--source / --hash-value: one file against an expected value; not valid for directories.",
		},
		Notes: []string{"Unrelated to the backup repos; it is a bundled utility."},
	},
	"init": {
		Usage:   "init [--force]",
		Purpose: `One-time setup: register notifications and put the program directory on the user PATH so zinc runs in any terminal. Needed for the portable zip; the installer has already done it.`,
		How:     `Writes the current user's PATH (a per-user environment variable in the registry) and registers notifications. It never touches backup data.--force redoes it, for when notifications stopped working or PATH was cleared.`,
		Params:  []string{"--force: run the setup again even if it was already done."},
		Notes:   []string{"Reopen the terminal after changing PATH."},
	},
	"fixpack": {
		Usage:   "fixpack <repo_name> [--check] [--password <password>]",
		Purpose: `Repair damaged packs: keep the files that still read, find a replacement for the ones that do not, and repack under the same name. It only touches pack content, never manifest.json (the ledger).`,
		How:     `Run 7z t on every pack first; on a non-solid archive 7-Zip reports each unreadable file by name. For every bad file, look for a replacement in order: the source directory first (still there, size and mtime matching index.tsv = that copy), then other packs from newest to oldest, each confirmed with 7z t so a damaged history copy is skipped. Nothing found means unrecoverable. After confirmation each bad pack is rewritten: extract what reads, drop in the replacements, repack with the same settings.`,
		Params: []string{
			"<repo_name>: which repo to repair.",
			"--check: only inspect and print the plan; change nothing.",
			"--password <password>: omit to use repo > settings; prompt if neither has one.",
		},
		Notes: []string{
			"A damaged non-solid pack is not a total loss: 7z t names the bad files, the rest still extract.",
			"Files with no replacement are skipped and the pack is left alone — it may be the last copy of that data.",
			"The ledger is rebuilt afterwards, so there is no need to run fixmeta by hand.",
			"zinc can confirm the source file exists with matching size and mtime; it cannot confirm the file itself is intact.",
		},
	},
	"fixmeta": {
		Usage: "fixmeta <repo_name> [--pack-start <pack_name>] [--pack-end <pack_name>] [--rehash] [--check] [--password <password>]",
		Purpose: `Reconcile the two bookkeeping files with the packs on disk. index.tsv (the index) records which pack holds the current version of each file, its size and time; manifest.json (the ledger) records the SHA-256 of every pack plus the deletion records written by each backup. It never modifies pack content.

Typical causes: a pack reported as changed by verify was deleted by hand, only part of the repo was downloaded on a new machine, compact died mid-way leaving stale hashes, packs were renamed or moved, antivirus quarantined one, or index.tsv was deleted or truncated. All of it can be rebuilt from the output directory.`,
		How: `Three steps, cheapest first:

1. Reconcile the manifest: scan the packs and .rm deletion records in the output     directory, drop entries whose file is gone, register packs the manifest never saw.     This is just stat calls, so it is fast.  2. Rebuild index.tsv: read each pack's archive header (no extraction), rebuild     "file -> current state", and cross out files that the deletion records removed.     This reads every pack, so it is slow on a large repo.  3. --rehash: recompute each pack's SHA-256 from what is actually on disk. This reads     whole packs, so it is slow and off by default.

Step 3 is the reason the command exists: nothing else can repair a stale manifest hash,and without it verify reports that pack as corrupt forever.`,
		Params: []string{
			"--pack-start <pack_name> / --pack-end <pack_name>: rebuild only that slice of the index.",
			"--rehash: recompute every registered pack's hash; repairs a compact that died mid-way.",
			"--check: print the plan, write nothing.",
			"--password <password>: reading archive headers needs one; repo > settings, else prompt.",
		},
		Notes: []string{
			"It does not repair a damaged pack. Only another copy or a fresh backup can do that.",
			"The old index is kept as index.tsv.bak before being overwritten.",
			"fixmeta takes the repo lock, so it cannot run while a backup is in progress.",
		},
	},
	"backup": {
		Usage:   "backup <repo_name|--all> [--notify] [--notify-empty] [--no-verify] [--password <password>]",
		Purpose: `Pack the files that are new or changed in the source directory into a batch of independent encrypted 7z archives. This is incremental: a pack holds only that batch of changes, and files already backed up and unchanged are not packed again. Deleting a file from the source does not shrink the repo; old copies are reclaimed by compact.`,
		How: `It compares index.tsv against the source directory to find files that are new, editedor resized — so a run with nothing to do costs almost nothing and prints "no changes".

Those files are split into packs by max_pack_byte (4G by default), and each pack isbuilt with 7za, tested once (7za t, checking CRCs), hashed, then recorded in themanifest and the index.

Committing pack by pack is deliberate: each finished pack is recorded immediately, so afailure later in the run does not throw away the packs already written. A backup can runfor hours; you keep the middle part either way.

Files that cannot be read while packing (locked, permission denied) are skipped andlisted in the terminal instead of failing the run — the next backup retries them. Thelist is not truncated, so a couple of reruns clear it out.`,
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
		Usage:   "compact <repo_name> [--pack <pack_name>] [--min-gain <threshold>] [--check] [--password <password>]",
		Purpose: `Reclaim space in the output directory: drop the copies a newer version has superseded, and the copies of files that no longer exist in the source, then rewrite those packs. Most useful after moving or renaming a large batch of files — zinc tracks files by path, so after a move the old path looks deleted and the new one looks new, and the same content ends up stored twice.`,
		How: `It scans the deletion records (.rm files written by each backup, noting which source files had disappeared), finds the packsthat contain the doomed entries, extracts each one, removes those entries, and repacks itas the same file name with the same packing parameters.

Two details are deliberate:

A pack that loses every entry is not deleted. Instead it gets a zero-byte placeholderfile named placeholder_please_delete_this, so the pack can still be rebuilt under thesame name. This matters for network drives: most of them treat a directory change as"overwrite and keep a version history", so a same-name overwrite pushes the new bytesto the cloud while a deletion would just look like a file vanished.

The repack uses the exact same arguments as a backup (compression level, solid block,header encryption), so a compacted pack has the same structure as the original andrestoring needs no special handling.`,
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
		Usage: "restore <repo_name> --dest <path> [--only <rel_path>] [--from <pack_name>]\n" +
			"                 [--no-overwrite] [--no-delete] [--check] [--password <password>]",
		Purpose: `Put the contents of the packs back on disk. By default the whole repo; --only narrows it to one path, --from extracts that single pack only.

This is the only command that modifies the destination. It asks when the destination is non-empty or deletions are about to be replayed — unless restore_wrap_folder is true (the default), in which case zinc nests a restore-<repo_name>-<timestamp> directory inside the destination; writing into a fresh directory has no overwrite risk, so it does not ask. Restoring straight into the source directory is not recommended: restore into an empty directory, check it, then copy back by hand, keeping the source intact meanwhile.`,
		How: `By default packs are extracted in name order (ascending); each later pack overwrites the earlier one, which leaves you with the latest state.

--from is the exception: it extracts only that one pack — it does not continue forward and does not replay deletions. Its meaning is "take the copy from that moment in history", not "start restoring from this pack onward". It must be combined with --only.

Deletion is a separate layer. Every backup records not just the files that exist but the files that vanished (the .rm files sitting next to each pack in the output directory). A restore replays those deletions along the timeline: delete a file on Oct 5, restore on Oct 10, and that file is removed from the destination as well. That is what "restore to that moment" means. --no-delete keeps those files instead.`,
		Params: []string{
			"--dest <path>: destination directory. Required; only --check may omit it.",
			"--only <rel_path>: restore just this file or directory. It still replays the deletions related to that path, so the result matches a full restore, only narrower.",
			"--from <pack_name>: extract only that one pack — no forward accumulation, no deletion replay. Must be combined with --only.",
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
