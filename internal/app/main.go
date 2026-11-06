package app

import (
	"fmt"
	"os"
)

func Main() {
	rotateLogs()
	args := initLang(os.Args[1:])
	// --lang 就是「把语言改掉」：写进 settings.json，之后一直是它。
	// 不存在「这次用英文、下次用中文」的场景，别为它留开关。
	// 非法语言已经在 initLang 里回退并警告，这里不写进配置。
	if n := normalizeLang(cliLang); n == "zh" || n == "en" {
		if err := persistLang(n); err != nil {
			fmt.Fprintf(os.Stderr, M.LangSaveFailed+"\n", err)
		} else {
			fmt.Printf(M.LangSaved+"\n", n)
		}
	}
	if len(args) < 1 {
		// 双击快捷方式时这是唯一一次能看到版本的机会；别再糊一整篇帮助上去。
		fmt.Println(versionLine())
		fmt.Println(M.BareHint)
		if startedFromExplorer() {
			pauseUnless(false)
		}
		return
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "init":
		cmdInit(rest)
	case "root":
		cmdRoot(rest)
	case "list":
		cmdList(rest)
	case "add":
		cmdAdd(rest)
	case "remove":
		cmdRemove(rest)
	case "set":
		cmdSet(rest)
	case "config":
		cmdConfig(rest)
	case "backup":
		cmdBackup(rest)
	case "restore":
		cmdRestore(rest)
	case "fixmeta":
		cmdFixmeta(rest)
	case "fixrepo":
		cmdFixrepo(rest)
	case "compact":
		cmdCompact(rest)
	case "verify":
		cmdVerify(rest)
	case "status":
		cmdStatus(rest)
	case "schedule":
		cmdSchedule(rest)
	case "unschedule":
		cmdUnschedule(rest)
	case "completion":
		cmdCompletion(rest)
	case "hash-compare":
		cmdHashCompare(rest)
	case "help", "--help", "-h":
		// `zinc help <命令>` 展开单条命令的详细说明；不带参数才是主面板
		if len(rest) > 0 {
			printHelp(rest)
			return
		}
		usage()
	case "version", "--version", "-v":
		// 跟其他命令一样，多余参数直接报错，不静默忽略
		if len(rest) > 0 {
			die(M.TooManyArgs, rest[0])
		}
		printVersion()
	case "readme":
		cmdReadme()
	default:
		// 打错命令时别把整篇帮助糊上来——先明说是哪个词不认识
		die(M.UnknownCommand, cmd)
	}
}
