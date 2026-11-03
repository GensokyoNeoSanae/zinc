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
