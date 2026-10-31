package app

import "os"

func Main() {
	rotateLogs()
	args := initLang(os.Args[1:])
	if len(args) < 1 {
		usage()
		// 双击快捷方式时这是唯一一次能看到用法的机会；从终端里跑则不停
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
	case "reindex":
		cmdReindex(rest)
	case "remanifest":
		cmdRemanifest(rest)
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
		usage()
	default:
		usage()
		os.Exit(1)
	}
}
