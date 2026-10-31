package app

// Messages 是所有用户可见文案的集合。
// 约束：同一字段在 zh/en 中 %s / %d / %v / %q / %w 的数量和顺序必须一致。
type Messages struct {
	Usage string

	// 通用
	ErrPrefix         string
	SaveConfigFailed  string
	LoadConfigFailed  string
	ParseConfigFailed string
	RepoNotFound      string

	// 参数校验
	TooManyArgs          string
	AllWithName          string
	ListWithArgs         string
	ConfirmUnscheduleAll string
	UnscheduleAllAborted string
	LangInvalid          string

	// index
	LoadIndexFailed  string
	ParseIndexFailed string

	// lock
	CreateLockFailed string
	WaitingLock      string
	LockTimeout      string

	// 密码
	PromptPassword      string
	PromptPasswordAgain string
	PasswordEmpty       string
	PasswordMismatch    string
	PasswordTooMany     string

	// root / list / add / remove / set
	RootUnset          string
	NoRepos            string
	ListItem           string
	UsageList          string
	UsageStatus        string
	UsageRoot          string
	RootNotDir         string
	UsageAdd           string
	MissingSource      string
	RepoExists         string
	RootNotSet         string
	WarnSourceMissing  string
	AddedRepo          string
	UsageRemove        string
	RemovedRepo        string
	UsageSet           string
	NeedSourceOrOutput string
	UpdatedRepo        string
	PathInvalid        string
	SourceOutputSame   string
	OutputInsideSource string
	SourceInsideOutput string
	OutputInUse        string
	OutputCaseWarn     string
	RemoveHasTasks     string

	// backup
	FlagAllBackup        string
	FlagAllStatus        string
	FlagPassword         string
	FlagNotify           string
	FlagNotifyEmpty      string
	FlagNoVerify         string
	UsageBackup          string
	NoReposBackup        string
	ScanErrSource        string
	ScanErrGeneric       string
	NoNew                string
	BackupFailed         string
	ToastTitle           string
	ToastHeadChanged     string
	ToastHeadChangedFail string
	ToastHeadFailed      string
	ToastHeadNone        string
	ToastFailedPrefix    string
	NotifyFailed         string
	SourceNotExist       string
	SourceEqualsOutput   string
	CreateOutputFailed   string
	LogNoChange          string
	VerifyFailed         string
	LogVerifyFailed      string
	WriteListfileFailed  string
	CompressFailed       string
	NoVolumes            string

	// --all + 显式密码与配置不一致时的拦截
	BackupPwMismatch  string
	ConfirmPwMismatch string
	BackupPwAborted   string

	// restore
	FlagDest            string
	UsageRestore        string
	MissingDest         string
	NoVolumesRestore    string
	CreateDestFailed    string
	Extracting          string
	ExtractFailed       string
	RestoreDone         string
	RestoreDestRoot     string
	RestoreDestSource   string
	RestoreDestInOutput string

	// verify
	FlagAllVerify       string
	FlagSample          string
	UsageVerify         string
	NoManifest          string
	ManifestBroken      string
	SampleInvalid       string
	MissingFile         string
	CorruptFile         string
	VerifySummary       string
	VerifyFailedMsg     string
	VerifyPwBad         string
	VerifyPwSummary     string
	VerifyPwNo7zip      string
	VerifyPwOnlySummary string
	FlagPasswordOnly    string
	FlagHashOnly        string
	VerifyFlagConflict  string
	VerifyNoPassword    string

	// schedule
	FlagAllSchedule   string
	FlagDaily         string
	FlagWeekly        string
	FlagMonthly       string
	FlagAt            string
	FlagNoNotify      string
	FlagYes           string
	NeedPeriod        string
	NeedAt            string
	UsageSchedule     string
	TimeInvalid       string
	MonthlyDayInvalid string
	MonthlyWarn       string
	UnknownPeriod     string
	TaskCreateFailed  string
	TaskCreated       string
	WeekdayInvalid    string
	QueryFailed       string
	Deleted           string
	NoMatch           string
	FlagAllUnschedule string
	UsageUnschedule   string

	// schedule 取密码（命令行 > 仓库 > settings）
	SchedulePwFromCLI      string
	SchedulePwFromRepo     string
	SchedulePwFromSettings string
	SchedulePwMissing      string
	SchedulePwMixed        string

	// completion
	UsageCompletion           string
	FlagCompletionInstall     string
	FlagListNames             string
	CompletionUnknownShell    string
	CompletionInstallOnlyPwsh string
	CompletionScriptWritten   string
	CompletionProfileUpdated  string
	CompletionProfileHasLine  string
	CompletionInstallFailed   string

	// schedule --list
	FlagListSchedule   string
	ScheduleHeader     string
	NoSchedule         string
	NoScheduleHint     string
	ScheduleItemName   string
	ScheduleItemDetail string
	CronDaily          string
	CronWeekly         string
	CronMonthly        string
	SchedNextNone      string
	StateReady         string
	StateDisabled      string
	StateRunning       string
	StateQueued        string
	StateUnknown       string

	// status
	StatusBrief           string
	StatusSource          string
	StatusOutput          string
	StatusBacked          string
	StatusPending         string
	StatusVolumes         string
	StatusLast            string
	StatusManifestExists  string
	StatusManifestMissing string

	// notify
	RegisterFailed string
	ToastFailed    string

	// init
	UsageInit       string
	FlagForce       string
	InitDir         string
	InitNotifyReady string
	InitNotifyFail  string
	InitPathAdded   string
	InitPathExists  string
	InitPathFail    string
	Init7zipOK      string
	Init7zipMissing string
	InitHint        string

	// install / uninstall（独立可执行文件，与主程序共用这套消息表）
	FlagInstall7zip  string
	FlagNoPause      string
	FlagYesUninstall string
	FlagKeepSchedule string

	InstallHeader    string
	Install7zipOK    string
	PromptSevenZip   string
	SevenZipMissing  string
	SevenZipNotFound string
	InstallDone      string
	PressEnter       string

	UninstallHeader        string
	UninstallTasks         string
	UninstallTaskItem      string
	UninstallNoTasks       string
	UninstallTaskRemoved   string
	UninstallTaskFail      string
	UninstallKeepSchedule  string
	UninstallNoNotify      string
	UninstallNotifyRemoved string
	UninstallNotifySkip    string
	UninstallNotifyFail    string
	UninstallPathRemoved   string
	UninstallPathMissing   string
	UninstallPathFail      string
	UninstallConfirm       string
	UninstallNeedYes       string
	UninstallAbort         string
	UninstallDone          string

	// settings / config
	LoadSettingsFailed  string
	ParseSettingsFailed string
	ConfigUnknownKey    string
	ConfigBadValue      string
	UsageConfig         string
	ConfigPath          string
	ConfigHeader        string
	ConfigItem          string
	ConfigSet           string
	ConfigUnset         string
	ConfigNotSet        string
	ConfigPromptValue   string
	ConfigAuto          string
	ConfigEmpty         string

	// index
	IndexBadHeader        string
	IndexLineBad          string
	IndexTruncated        string
	IndexCountMismatch    string
	IndexWriteFailed      string
	ManifestCorrupt       string
	ManifestCorruptLocked string
	ManifestReadFailed    string
	ManifestWriteFailed   string

	// pack
	PackTooBig       string
	ClockRollback    string
	RenamePackFailed string
	DetailLine       string
	EncFrom          string
	ScanSkipItem     string
	ScanSkipMore     string
	LockedItem       string
	LockedMore       string
	AllLocked        string
	PackSkipItem     string
	PackSkipMore     string
	PartialCommit    string

	// 删除记录
	DeletedLine          string
	DeletingBatch        string
	RestoreDelSummary    string
	RestoreDelMissing    string
	FlagNoDelete         string
	FlagNotifyWarn       string
	ToastRestoreTitle    string
	ToastRestoreBody     string
	CheckListDelItem     string
	DelRecordBadPath     string
	DelRecordBadHeader   string
	DelRecordNoSum       string
	DelRecordAllBad      string
	DelRecordEdited      string
	DelLineSkipped       string
	DelSkipped           string
	DelSkippedAfterFail  string
	PackFailed           string
	PackErrLine          string
	PackPasswordHint     string
	RestoreIncomplete    string
	RestoreFailedPacks   string
	DelRecordReadFailed  string
	DelRecordWriteFailed string
	DelRemoveFailed      string
	VerifyRmMissing      string
	VerifyRmCorrupt      string
	VerifyRmSummary      string

	// restore
	FlagOverwrite        string
	FlagNoOverwrite      string
	FlagCheckList        string
	FlagConflict         string
	ModeOverwrite        string
	ModeRename           string
	DestUnset            string
	RestoreOverwriteWarn string
	RestoreWrapLine      string
	CheckListHeader      string
	CheckListItem        string
	CheckListTotal       string
	PackOrderBroken      string
	NoPacksRestore       string
	FlagOnly             string
	FlagFrom             string
	FromRequiresOnly     string
	RestorePartial       string
	MissingOnly          string
	PathNotBackedUp      string
	PackNotSpecified     string
	OnlyOrigin           string
	OnlyDone             string
	OnlyVersions         string
	OnlyNoVersions       string

	// reindex
	FlagPackStart     string
	FlagPackEnd       string
	FlagYesReindex    string
	UsageReindex      string
	ReindexStart      string
	ReindexPack       string
	ReindexDone       string
	ReindexExists     string
	ReindexListFailed string
	ReindexNoPacks    string
	ReindexBakFailed  string
	ReindexDeleted    string

	// compact
	FlagCompactPack        string
	FlagCompactMinGain     string
	FlagYesCompact         string
	UsageCompact           string
	CompactStart           string
	CompactDryRunHead      string
	CompactDryRunTail      string
	CompactPlanPack        string
	CompactSkipNone        string
	CompactSkipSmall       string
	CompactPlanSummary     string
	CompactNothing         string
	CompactDoing           string
	CompactEmptyDoing      string
	CompactPackDone        string
	CompactDoneLine        string
	CompactUploadHint      string
	CompactFailedMsg       string
	CompactNoPacks         string
	CompactPackFailed      string
	CompactPackNotFound    string
	CompactIndexStale      string
	CompactPackListFailed  string
	CompactExtractFailed   string
	CompactExtractMismatch string
	CompactRebuildFailed   string
	CompactVerifyFailed    string
	CompactCountMismatch   string
	CompactNoSpace         string
	CompactBadGain         string
	PlaceholderRemoved     string
	FlagNotifyCompact      string
	ToastCompactTitle      string
	ToastCompactBody       string
	ToastCompactFailBody   string

	// set flags
	FlagMaxPackByte string
	FlagCompress    string
	FlagSolid       string
	FlagExclude     string

	// status
	StatusPacks   string
	StatusContent string
}

var M *Messages

var zhMessages = &Messages{
	Usage: `zinc - 多仓库、文件级增量、加密独立包备份

用法:
  zinc [--lang zh|en] <command> ...

  zinc init [--force]                   一次性初始化：注册通知 + 把本目录加入用户 PATH
  zinc root [<path>]                    查看/设置全局输出根目录
  zinc list                             列出仓库
  zinc add <name> --source <path> [--output <path>]
  zinc remove <name>
  zinc set <name> [--source <path>] [--output <path>] [--password <pw>]
                  [--max-pack-byte <n>] [--compress <0-9>] [--solid <64M|off>]
                  [--exclude <glob>]
  zinc config [get|set|unset|path] [<key>] [<value>]      全局设置
  zinc backup <name|--all> [--password <pw>] [--notify] [--notify-empty] [--no-verify]
  zinc restore <name> [--dest <path>] [--password <pw>]
                       [--overwrite|--no-overwrite] [--no-delete] [--check-list]
  zinc restore <name> --only <相对路径> [--dest <path>] [--from <包名|序号>]
                       只恢复这一个文件/目录；配 --check-list 可列出它有哪些版本
  zinc reindex <name> [--password <pw>] [--pack-start <名>] [--pack-end <名>] [--yes]
  zinc compact <name> [--password <pw>] [--pack <名|序号>] [--min-gain <5%|8M>] [--yes] [--notify]
                       丢掉包里被新版取代的旧副本；包名不变，云端同名覆盖进历史版本
  zinc verify <name|--all> [--sample <pct>] [--password-only|--hash-only]
  zinc status [--all] [<name>]
  zinc completion <powershell|bash|zsh|fish> [--install]   生成 Tab 补全脚本
  zinc schedule <name|--all> --daily --at HH:MM[,HH:MM...]
  zinc schedule <name|--all> --weekly 1,3 --at HH:MM
  zinc schedule <name|--all> --monthly 1,15 --at HH:MM [--yes]
  zinc schedule --list                 列出当前所有计划任务
  zinc unschedule [<name>|--all]

说明:
  - 每个包都是独立完整的加密 7z，文件名就是打包时刻，恢复按名字升序
  - 固实块默认 64M：块里坏一个字节，该块后续文件就解不出来；--solid off 改成每文件独立块
  - 不给 --password 时按「仓库 > settings」取；backup/restore 取不到才交互输入，schedule 取不到直接报错
  - schedule 会把解析到的密码明文写进计划任务，之后改了密码要重建任务（zinc unschedule <name> 再 schedule）
  - 恢复默认覆盖同名文件（--no-overwrite 则保留两份），--check-list 只列顺序不动手
  - 恢复会执行之后备份记下的删除（--no-delete 保留这些文件）；删除记录是输出目录里的 .rm 旁注
  - compact 删掉的是「已被新版取代 / 已从源目录删除」的副本，那些旧版本随之消失；包被腾空后只留一个占位文件
  - 索引只是缓存，丢了用 zinc reindex 从包本身重建
  - 输出默认落在 <root>\<name>，用 zinc root 设置
  - 解压后先跑一次 zinc init，之后在任意终端直接敲 zinc
  - 安装/卸载用 tools\ 下的 zinc-register.exe / zinc-unregister.exe；zinc init 是轻量版，不强制 7-Zip
  - unschedule <name> 删该仓库的任务；unschedule --all 删所有 zinc 任务；unschedule all 只删 --all 建的那个总任务
`,

	ErrPrefix:         "[错误] ",
	SaveConfigFailed:  "保存配置失败: %v",
	LoadConfigFailed:  "读取 %s 失败: %w",
	ParseConfigFailed: "解析 %s 失败: %w",
	RepoNotFound:      "仓库不存在: %s",

	TooManyArgs:          "多余的参数: %s（用 zinc --help 看用法）",
	AllWithName:          "--all 不能和名字一起用: %s",
	ListWithArgs:         "--list 只查询任务，不能同时带周期/时间参数",
	ConfirmUnscheduleAll: "这会删除所有 zinc 计划任务，确定吗？[y/n]: ",
	UnscheduleAllAborted: "已取消",
	LangInvalid:          "[警告] 未知语言 %q，已回退到 %s",

	LoadIndexFailed:  "读取索引失败: %w",
	ParseIndexFailed: "解析索引失败: %w",

	CreateLockFailed: "创建锁失败: %w",
	WaitingLock:      "[等待] 仓库正被其他任务使用，最多等待 %v...\n",
	LockTimeout:      "仓库被占用超过 %v。\n如果确认没有正在进行的压缩，请手动删除这个文件后重试：\n  %s",

	PromptPassword:      "输入密码: ",
	PromptPasswordAgain: "再次输入密码: ",
	PasswordEmpty:       "密码不能为空",
	PasswordMismatch:    "两次输入不一致，请重新输入",
	PasswordTooMany:     "密码确认失败次数过多",

	RootUnset:          "(未设置)",
	UsageList:          "用法: zinc list [--names]",
	UsageStatus:        "用法: zinc status [--all] [<name>]（--all 不能和仓库名一起用）",
	UsageRoot:          "用法: zinc root [<path>]",
	RootNotDir:         "root 路径已存在且不是目录: %s",
	NoRepos:            "还没有仓库，先 zinc add",
	ListItem:           "%s\n  源:   %s\n  输出: %s\n",
	UsageAdd:           "用法: zinc add <name> --source <path>",
	MissingSource:      "缺少 --source",
	RepoExists:         "仓库 %q 已存在，用 zinc set 修改",
	RootNotSet:         "root 未设置，请先 zinc root <path> 或显式给 --output",
	WarnSourceMissing:  "[警告] 源目录当前不存在: %s\n",
	AddedRepo:          "已添加 %s\n  源:   %s\n  输出: %s\n",
	UsageRemove:        "用法: zinc remove <name>",
	RemovedRepo:        "已移除仓库 %s（包文件和索引未删除）\n",
	UsageSet:           "用法: zinc set <name> [--source <path>] [--output <path>] [--password <pw>]\n                    [--max-pack-byte <n>] [--compress <0-9>] [--solid <64M|off>] [--exclude <glob>]",
	NeedSourceOrOutput: "至少给一个设置项：--source / --output / --password / --max-pack-byte / --compress / --solid / --exclude",
	UpdatedRepo:        "已更新 %s\n  源:   %s\n  输出: %s\n",
	PathInvalid:        "路径无效: %s（%v）",
	SourceOutputSame:   "源目录与输出目录不能相同: %s",
	OutputInsideSource: "输出目录不能位于源目录内: %s 在 %s 内",
	SourceInsideOutput: "源目录不能位于输出目录内: %s 在 %s 内",
	OutputInUse:        "输出目录已被仓库 %q 使用: %s",
	OutputCaseWarn:     "[警告] 输出目录 %s 与仓库 %q 的输出只有大小写不同，Windows 上会落到同一个目录\n",
	RemoveHasTasks:     "[警告] 仓库 %s 还有 %d 个计划任务（%s），可用 zinc unschedule %s 删除\n",

	FlagAllBackup:        "备份全部仓库",
	FlagAllStatus:        "一次展开所有仓库的详细信息",
	FlagPassword:         "备份密码",
	FlagNotify:           "发送 Windows 通知",
	FlagNotifyEmpty:      "无新增也通知",
	FlagNoVerify:         "跳过 7za t 自检",
	UsageBackup:          "用法: zinc backup <name> 或 zinc backup --all",
	NoReposBackup:        "没有仓库",
	ScanErrSource:        "%s · 源目录不存在",
	ScanErrGeneric:       "%s · 扫描失败: %v",
	NoNew:                "%s · 无新增\n",
	BackupFailed:         "%s · 失败: %v\n",
	ToastTitle:           "zinc 备份完成",
	ToastHeadChanged:     "%d 个仓库有新增，%d 个无新增 · 合计 %s",
	ToastHeadChangedFail: "%d 个仓库有新增 · %d 个失败 · 合计 %s",
	ToastHeadFailed:      "%d 个仓库失败",
	ToastHeadNone:        "全部 %d 个仓库无新增",
	ToastFailedPrefix:    "失败: ",
	NotifyFailed:         "[通知失败] %v\n",
	SourceNotExist:       "源目录不存在: %s",
	SourceEqualsOutput:   "源目录与输出目录相同，无法备份: %s",
	CreateOutputFailed:   "创建输出目录失败: %w",
	LogNoChange:          "backup %s: 无变化",
	VerifyFailed:         "自检失败，已删除本批次卷: %w",
	LogVerifyFailed:      "backup %s: 自检失败 %v",
	WriteListfileFailed:  "写 listfile 失败: %w",
	CompressFailed:       "7za 压缩失败: %v\n%s",
	NoVolumes:            "7za 执行成功但未生成任何包",

	BackupPwMismatch:  "[警告] --password 与仓库 %s 的配置密码不一致\n",
	ConfirmPwMismatch: "密码与仓库默认密码不一致！确定要执行吗？[y/n]: ",
	BackupPwAborted:   "已取消：密码与配置不一致",

	FlagDest:            "恢复目标目录",
	UsageRestore:        "用法: zinc restore <name> --dest <path>",
	MissingDest:         "缺少 --dest",
	NoVolumesRestore:    "没有找到任何包: %s",
	CreateDestFailed:    "创建恢复目录失败: %w",
	Extracting:          "[解压] %s\n",
	ExtractFailed:       "解压失败 %s: %v\n%s",
	RestoreDone:         "[完成] 恢复到 %s\n",
	RestoreDestRoot:     "拒绝恢复到盘根目录: %s",
	RestoreDestSource:   "拒绝恢复到源目录（会覆盖源文件）: %s",
	RestoreDestInOutput: "拒绝恢复到输出目录或其中: %s（输出目录 %s）",

	FlagAllVerify:       "校验全部仓库",
	FlagSample:          "抽样百分比 1-100",
	FlagPasswordOnly:    "只验密码，不做 SHA-256",
	FlagHashOnly:        "只算 SHA-256，不验密码",
	VerifyFlagConflict:  "--password-only 和 --hash-only 不能同时用",
	VerifyNoPassword:    "没有可用密码：先设置仓库密码或 settings.default_password",
	UsageVerify:         "用法: zinc verify <name|--all> [--sample <pct>] [--password-only|--hash-only]",
	NoManifest:          "没有 manifest（%s），无法校验",
	ManifestBroken:      "manifest %s 损坏或无法读取: %v",
	SampleInvalid:       "--sample 必须在 1-100 之间: %d",
	MissingFile:         "[缺失] %s\n",
	CorruptFile:         "[损坏] %s\n",
	VerifySummary:       "%s: 通过 %d, 损坏 %d, 缺失 %d\n",
	VerifyPwOnlySummary: "%s: 密码通过 %d, 密码不符 %d, 缺失 %d\n",
	VerifyFailedMsg:     "校验未通过",
	VerifyPwBad:         "[密码不符] %s\n",
	VerifyPwSummary:     "[密码] %d 个包无法用当前配置的密码打开（包本身可能没坏，是密码不对）\n",
	VerifyPwNo7zip:      "[提示] 找不到 7-Zip，跳过密码校验\n",

	FlagAllSchedule:   "为全部仓库建总任务",
	FlagDaily:         "每天",
	FlagWeekly:        "周几，1-7，逗号分隔",
	FlagMonthly:       "每月几号，逗号分隔",
	FlagAt:            "时间 HH:MM[,HH:MM...]",
	FlagNoNotify:      "任务不发通知",
	FlagYes:           "跳过每月日期警告",
	NeedPeriod:        "必须指定 --daily / --weekly / --monthly 之一",
	NeedAt:            "缺少 --at HH:MM",
	UsageSchedule:     "用法: zinc schedule <name|--all> ...",
	TimeInvalid:       "时间格式错误: %q（应为 HH:MM）",
	MonthlyDayInvalid: "每月日期错误: %q（应为 1-31）",
	MonthlyWarn:       "你设置了每月 %d 号备份，但 2 月没有 %d 号，4/6/9/11 月只有 30 天。\n这些月份将不会执行备份。\n继续创建任务请加 --yes",
	UnknownPeriod:     "未知周期: %s",
	TaskCreateFailed:  "创建任务 %s 失败: %v\n%s",
	TaskCreated:       "[任务] %s (%s %s %s)\n",
	WeekdayInvalid:    "周几错误: %q（应为 1-7，1=周一）",
	QueryFailed:       "查询任务失败: %v\n%s",
	Deleted:           "[删除] %s\n",
	NoMatch:           "没有匹配的任务",
	FlagAllUnschedule: "删除全部 zinc 任务",
	UsageUnschedule:   "用法: zinc unschedule <name> 或 zinc unschedule --all",

	SchedulePwFromCLI:      "[密码] 取自命令行\n",
	SchedulePwFromRepo:     "[密码] 取自仓库配置\n",
	SchedulePwFromSettings: "[密码] 取自 settings.json 的默认密码\n",
	SchedulePwMissing:      "找不到可用密码：用 --password 指定，或先 zinc config set default_password <密码>",
	SchedulePwMixed:        "--all 的目标仓库密码不一致，一条任务没法统一；请分开建任务，或先统一各仓库密码",

	UsageCompletion:           "用法: zinc completion <powershell|bash|zsh|fish> [--install]",
	FlagCompletionInstall:     "只为 powershell：写好脚本并往 $PROFILE 追加加载行",
	FlagListNames:             "每行一个仓库名（给脚本/补全用）",
	CompletionUnknownShell:    "不支持的 shell: %s（可选: powershell, bash, zsh, fish）",
	CompletionInstallOnlyPwsh: "--install 只支持 powershell；其它 shell 请用 zinc completion <shell> > 文件",
	CompletionScriptWritten:   "[完成] 补全脚本: %s\n",
	CompletionProfileUpdated:  "[完成] 加载行已写入: %s\n",
	CompletionProfileHasLine:  "[跳过] %s 里已经有加载行\n",
	CompletionInstallFailed:   "安装补全失败: %v",

	FlagListSchedule:   "列出当前所有计划任务",
	ScheduleHeader:     "[计划] %d 个任务\n\n",
	NoSchedule:         "当前没有 zinc 的计划任务",
	NoScheduleHint:     "用 zinc schedule <仓库> --daily --at 03:00 添加",
	ScheduleItemName:   "  %s\n",
	ScheduleItemDetail: "    %s · 下次 %s · %s\n",
	CronDaily:          "每天 %s",
	CronWeekly:         "每周 %s %s",
	CronMonthly:        "每月 %s %s",
	SchedNextNone:      "未安排",
	StateReady:         "就绪",
	StateDisabled:      "已禁用",
	StateRunning:       "运行中",
	StateQueued:        "排队中",
	StateUnknown:       "未知",

	StatusBrief:           "%-12s %6d 文件 / %3d 批次 / %3d 包   最后备份 %-19s\n",
	StatusSource:          "  源:   %s\n",
	StatusOutput:          "  输出: %s\n",
	StatusBacked:          "  已备份: %d 文件\n",
	StatusPending:         "  待备份: %d 个新增, %d 个修改, %d 个待删除, %d 个跳过\n",
	StatusVolumes:         "  卷: %d 批次 / %d 个文件 / %s\n",
	StatusLast:            "  最后备份: %s\n",
	StatusManifestExists:  "  manifest: 存在\n",
	StatusManifestMissing: "  manifest: 缺失\n",

	RegisterFailed: "注册 AUMID 失败: %v\n%s",
	ToastFailed:    "toast 失败: %v",

	UsageInit:       "用法: zinc init [--force]",
	FlagForce:       "即使已注册也重建通知快捷方式（换图标/通知坏了时用）",
	InitDir:         "程序目录: %s\n",
	InitNotifyReady: "[ok] 通知注册: %s\n",
	InitNotifyFail:  "[失败] 通知注册: %v\n",
	InitPathAdded:   "[ok] 已把程序目录加入用户 PATH: %s（新开的终端生效）\n",
	InitPathExists:  "[跳过] 用户 PATH 里已有: %s\n",
	InitPathFail:    "[失败] 写用户 PATH: %v\n",
	Init7zipOK:      "[ok] 7za: %s\n",
	Init7zipMissing: "[警告] 找不到 7z.exe / 7za.exe（可用 zinc config set 7zip_path 指定）\n",
	InitHint:        "\n完成。新开一个终端就能直接敲 zinc；通知不弹就再跑一次 zinc init --force。\n",

	FlagInstall7zip:  "指定 7z.exe / 7za.exe 的路径",
	FlagNoPause:      "结束后不等待按键（给脚本用）",
	FlagYesUninstall: "不再询问确认",
	FlagKeepSchedule: "保留计划任务（默认删除）",

	InstallHeader:    "zinc 安装\n程序目录: %s\n\n",
	Install7zipOK:    "[ok] 7-Zip: %s\n",
	PromptSevenZip:   "没有找到 7-Zip。请先安装（https://www.7-zip.org/），或输入 7z.exe / 7za.exe 的完整路径（留空退出）: ",
	SevenZipMissing:  "没有可用的 7-Zip。zinc 依赖 7z.exe 或 7za.exe；装好后重跑 zinc-register.exe，或用 --7zip <路径> 指定。",
	SevenZipNotFound: "7-Zip 路径不可用: %s",
	InstallDone:      "\n完成。新开一个终端就能直接敲 zinc；接下来设置备份根目录：\n  zinc root D:\\Backup\\zinc\n",
	PressEnter:       "\n按 Enter 关闭窗口……",

	UninstallHeader:        "zinc 卸载\n程序目录: %s\n\n",
	UninstallTasks:         "将删除 %d 个计划任务：\n",
	UninstallTaskItem:      "  %s\n",
	UninstallNoTasks:       "[跳过] 没有指向本目录的计划任务\n",
	UninstallTaskRemoved:   "[ok] 已删除 %d 个计划任务\n",
	UninstallTaskFail:      "[失败] 处理计划任务: %v\n",
	UninstallKeepSchedule:  "[跳过] 按 --keep-schedule 保留计划任务\n",
	UninstallNoNotify:      "[跳过] 没有通知注册\n",
	UninstallNotifyRemoved: "[ok] 已删除通知注册（快捷方式、AppUserModelId、标记文件）\n",
	UninstallNotifySkip:    "[跳过] 通知注册属于另一个目录，未改动: %s\n",
	UninstallNotifyFail:    "[失败] 删除通知注册: %v\n",
	UninstallPathRemoved:   "[ok] 已从用户 PATH 移除: %s（新开的终端生效）\n",
	UninstallPathMissing:   "[跳过] 用户 PATH 里没有: %s\n",
	UninstallPathFail:      "[失败] 写用户 PATH: %v\n",
	UninstallConfirm:       "\n继续卸载？备份数据和设置文件都不会删除 [y/N]: ",
	UninstallNeedYes:       "[失败] 非交互环境需要加 --yes 才会卸载",
	UninstallAbort:         "已取消。\n",
	UninstallDone:          "\n卸载完成。程序文件还在 %s，可以直接删掉这个文件夹；备份数据（输出目录里的 .7z）没有被改动。\n",

	LoadSettingsFailed:  "读取设置文件 %s 失败: %w",
	ParseSettingsFailed: "解析设置文件 %s 失败: %w",
	ConfigUnknownKey:    "未知的设置项: %s",
	ConfigBadValue:      "设置项 %s 的值无效: %q",
	UsageConfig:         "用法: zinc config [get|set|unset|path] [<key>] [<value>]",
	ConfigPath:          "配置文件: %s\n",
	ConfigHeader:        "全局设置（%s）\n",
	ConfigItem:          "  %-22s = %s\n",
	ConfigSet:           "已设置 %s = %s\n",
	ConfigUnset:         "已清除 %s（回落到内置默认）\n",
	ConfigNotSet:        "%s 当前未设置\n",
	ConfigPromptValue:   "输入 %s 的值: ",
	ConfigAuto:          "(自动)",
	ConfigEmpty:         "(空)",

	IndexBadHeader:        "索引文件 %s 的头部不是 %q，可能已损坏；用 zinc reindex 重建",
	IndexLineBad:          "索引文件 %s 第 %d 行格式错误；用 zinc reindex 重建",
	IndexTruncated:        "索引文件 %s 不完整（缺少结尾标记），可能被截断；用 zinc reindex 重建",
	IndexCountMismatch:    "索引文件 %s 计数不符（记录 包%d/文件%d，实际 包%d/文件%d）；用 zinc reindex 重建",
	IndexWriteFailed:      "写索引失败: %w",
	ManifestCorrupt:       "manifest %s 损坏，已改名为 %s；本次备份中止: %v",
	ManifestCorruptLocked: "manifest %s 损坏且无法改名: %v",
	ManifestReadFailed:    "读取 manifest 失败: %w",
	ManifestWriteFailed:   "写入 manifest 失败: %w",

	PackTooBig:       "[警告] %s (%s) 超过单包上限 %s，单独成包\n",
	ClockRollback:    "[警告] 系统时钟似乎回退，包名时刻取 %s\n",
	RenamePackFailed: "重命名成品包失败: %w",
	DetailLine:       "[ok] %s · %d 文件 · %s · %d 包 · %s · scan %dms · hash %s",
	EncFrom:          "enc✓(%s)",
	ScanSkipItem:     "[跳过] %s\n",
	ScanSkipMore:     "  …另有 %d 项跳过\n",
	LockedItem:       "[占用] %s（下次备份会重试）\n",
	LockedMore:       "  …另有 %d 个文件被占用\n",
	AllLocked:        "%s · 变更的 %d 个文件全被占用，本次跳过\n",
	PackSkipItem:     "[跳过] %s（打包时读不到，下次备份会重试）\n",
	PackSkipMore:     "  …另有 %d 个文件本次没能打进包\n",
	PartialCommit:    "%s · 部分完成：已提交 %d 包（%s），%d 个文件留待下次\n",

	DeletedLine:          "[删除] %s · %d 个文件已从源目录消失（记录 %s）\n",
	DeletingBatch:        "[删除] %s · 移除 %d 个文件\n",
	RestoreDelSummary:    "[完成] 按删除记录移除了 %d 个文件\n",
	RestoreDelMissing:    "  删除记录里另有 %d 项在恢复结果中已不存在（跳过；compact 清掉旧副本后属正常）\n",
	FlagNoDelete:         "恢复时不执行备份之后记录的删除",
	FlagNotifyWarn:       "恢复不完整时发 Windows 通知",
	ToastRestoreTitle:    "zinc 恢复不完整",
	ToastRestoreBody:     "%s · %d 个包解压报错，文件可能不完整，用 zinc verify 查",
	CheckListDelItem:     "  %2d. %-44s %8d 项删除\n",
	DelRecordBadPath:     "删除记录里有非法路径 %q（记录 %s）",
	DelRecordBadHeader:   "删除记录 %s 的表头不对（应为 %q）",
	DelRecordNoSum:       "删除记录 %s 缺少校验和行（# sha256 …）",
	DelRecordAllBad:      "删除记录 %s 的每一行都不可用（路径或行校验被改坏）",
	DelRecordEdited:      "[警告] 删除记录 %s 与写入时的哈希不一致（多半被手工加删过行）；每一行都自洽，按现有内容继续\n",
	DelLineSkipped:       "[警告] 删除记录 %s 有一行不可用，已跳过: %s\n",
	DelSkipped:           "[警告] 删除记录不可用，按 --no-delete 忽略、只恢复文件: %v\n",
	DelSkippedAfterFail:  "[跳过] %s · 它所属批次的包解压报错，删除记录一起跳过\n",
	PackFailed:           "[警告] %s · 解压报错（能解出来的已经写入目标），继续解后面的包\n",
	PackErrLine:          "         %s\n",
	PackPasswordHint:     "[提示] 如果每个包都这样，多半是密码不对而不是包坏了；确认 --password 或 settings.json 里的 default_password\n",
	RestoreIncomplete:    "[警告] 有 %d 个包没能解开，恢复结果不完整；报 CRC 错误的文件可能以错误内容留在磁盘上，用 zinc verify 确认\n",
	RestoreFailedPacks:   "%d 个包解压失败",
	DelRecordReadFailed:  "读取删除记录 %s 失败: %v",
	DelRecordWriteFailed: "写入删除记录失败: %w",
	DelRemoveFailed:      "删除 %s 失败: %v",
	VerifyRmMissing:      "[缺失] 删除记录 %s\n",
	VerifyRmCorrupt:      "[损坏] 删除记录 %s\n",
	VerifyRmSummary:      "删除记录: 通过 %d, 损坏 %d, 缺失 %d\n",

	FlagOverwrite:        "恢复时覆盖同名文件",
	FlagNoOverwrite:      "恢复时不覆盖，同名文件重命名为 _1",
	FlagCheckList:        "只列出将解压的包顺序，不动文件",
	FlagConflict:         "不能同时使用 --overwrite 与 --no-overwrite",
	ModeOverwrite:        "覆盖",
	ModeRename:           "不覆盖",
	DestUnset:            "(未指定)",
	RestoreOverwriteWarn: "[警告] 覆盖模式：目标目录中的同名文件将被替换\n",
	RestoreWrapLine:      "[外层] 恢复写入 %s（restore_wrap_folder=on）\n",
	CheckListHeader:      "恢复顺序（%d 个包，%d 份删除记录）· 模式：%s · 目标：%s\n",
	CheckListItem:        "  %2d. %-44s %10s\n",
	CheckListTotal:       "合计 %s\n",
	PackOrderBroken:      "包顺序不单调：%s 与 %s 的顺序无法确定。请检查输出目录里的文件名，恢复已中止",
	NoPacksRestore:       "没有找到任何包: %s",
	FlagOnly:             "只恢复这个相对路径（文件或目录），不动其他文件",
	FlagFrom:             "从指定包恢复：包名、名字前缀或序号（1 起）",
	FromRequiresOnly:     "--from 只能和 --only 一起使用",
	RestorePartial:       "[警告] 恢复已中止，目标目录可能只完成了一部分\n",
	MissingOnly:          "--only 的路径不能为空",
	PathNotBackedUp:      "索引和所有包里都没找到这个路径: %s",
	PackNotSpecified:     "找不到包: %s",
	OnlyOrigin:           "[来自] %s\n",
	OnlyDone:             "[完成] %s → %s\n",
	OnlyVersions:         "包含 %s 的包（%d 个，时间升序）\n",
	OnlyNoVersions:       "没有任何包包含这个路径: %s",

	FlagPackStart:     "只重建从该包名开始的范围",
	FlagPackEnd:       "只重建到该包名为止的范围",
	FlagYesReindex:    "覆盖已有索引（会先备份为 .bak）",
	UsageReindex:      "用法: zinc reindex <name> [--password <pw>] [--pack-start <名>] [--pack-end <名>] [--yes]",
	ReindexStart:      "[重建] 扫描 %d 个包（只读归档头）\n",
	ReindexPack:       "  %s · %d 个文件\n",
	ReindexDone:       "[完成] %d 个包 / %d 个文件 → %s\n",
	ReindexExists:     "索引已存在：%s\n加 --yes 覆盖（会先备份为 .bak）",
	ReindexListFailed: "读取包 %s 的目录失败: %v",
	ReindexNoPacks:    "没有找到任何包: %s",
	ReindexBakFailed:  "备份旧索引失败，已中止: %w",
	ReindexDeleted:    "  按删除记录划掉 %d 个文件\n",

	FlagCompactPack:        "只处理指定的包（包名或序号）",
	FlagCompactMinGain:     "收益低于这个值就跳过重打：5% 或 8M",
	FlagYesCompact:         "真正执行（不给时只报告）",
	UsageCompact:           "用法: zinc compact <name> [--pack <名|序号>] [--min-gain <5%%|8M>] [--yes] [--notify] [--password <pw>]",
	CompactStart:           "[压缩] %s：检查 %d 个包（只读归档头）\n",
	CompactDryRunHead:      "[预演] 没有 --yes，这一步只看不改\n",
	CompactDryRunTail:      "确认无误后加 --yes 执行。\n",
	CompactPlanPack:        "  %s · 可删 %d 个 / %s（%.1f%%）→ 待重写\n",
	CompactSkipNone:        "  %s · 没有可删的文件\n",
	CompactSkipSmall:       "  %s · 可删 %d 个 / %s（%.1f%%），未达阈值 %s → 跳过\n",
	CompactPlanSummary:     "[计划] %d 个包待重写，预计释放 %s\n",
	CompactNothing:         "[完成] 没有需要重写的包（%d 个没有死副本，%d 个低于阈值）\n",
	CompactDoing:           "[重写] %s：删 %d 个文件 / %s\n",
	CompactEmptyDoing:      "[重写] %s：文件全被新版取代，包内只留占位 %s\n",
	CompactPackDone:        "  %s · %s → %s\n",
	CompactDoneLine:        "[完成] 重写 %d 个包，失败 %d 个，%s → %s，释放 %s\n",
	CompactUploadHint:      "提醒：网盘那边要重新上传这些包；同名覆盖会让旧包进「历史版本」。\n",
	CompactFailedMsg:       "%d 个包重写失败，原包未改动",
	CompactNoPacks:         "没有找到任何包: %s",
	CompactPackFailed:      "处理包 %s 失败: %v\n",
	CompactPackNotFound:    "找不到这个包: %s",
	CompactIndexStale:      "索引与包目录对不上（索引 %d 个包，目录 %d 个）：%s\n先跑 zinc reindex %s 把索引重建完整，再来压缩",
	CompactPackListFailed:  "读取包 %s 的目录失败: %v",
	CompactExtractFailed:   "解出包 %s 失败（原包未改动）: %v\n%s",
	CompactExtractMismatch: "包 %s 里的 %s 没能原样解出来（大小不符），跳过这个包（原包未改动）",
	CompactRebuildFailed:   "重打包 %s 失败（原包未改动）: %v",
	CompactVerifyFailed:    "重打的包 %s 自检失败，已丢弃（原包未改动）: %v",
	CompactCountMismatch:   "重打的包 %s 内容不符：应有 %d 个文件，实际 %d 个，已丢弃（原包未改动）",
	CompactNoSpace:         "剩余空间 %s 不够处理 %s（约需 %s），跳过",
	CompactBadGain:         "收益阈值无效: %s（写 5%% 或 8M）",
	PlaceholderRemoved:     "  清掉 %d 个占位文件（%s）\n",
	FlagNotifyCompact:      "完成后发 Windows 通知",
	ToastCompactTitle:      "zinc 压缩完成",
	ToastCompactBody:       "%s · 重写 %d 个包 · 释放 %s",
	ToastCompactFailBody:   "%s · %d 个包重写失败，其余已完成，细节看终端",

	FlagMaxPackByte: "单包上限，如 4G / 50M / 纯字节",
	FlagCompress:    "压缩等级 0-9",
	FlagSolid:       "固实块上限，如 64M / off（off = 每文件独立块，最抗损）",
	FlagExclude:     "排除规则，可重复",

	StatusPacks:   "  包: %d 批次 / %d 个包 / %s\n",
	StatusContent: "  内容日期: %s ~ %s\n",
}

var enMessages = &Messages{
	Usage: `zinc - multi-repo, file-level incremental, encrypted independent packs

Usage:
  zinc [--lang zh|en] <command> ...

  zinc init [--force]                   one-time setup: register notifications + add this dir to user PATH
  zinc root [<path>]                    show/set global output root
  zinc list                             list repos
  zinc add <name> --source <path> [--output <path>]
  zinc remove <name>
  zinc set <name> [--source <path>] [--output <path>] [--password <pw>]
                  [--max-pack-byte <n>] [--compress <0-9>] [--solid <64M|off>]
                  [--exclude <glob>]
  zinc config [get|set|unset|path] [<key>] [<value>]      global settings
  zinc backup <name|--all> [--password <pw>] [--notify] [--notify-empty] [--no-verify]
  zinc restore <name> [--dest <path>] [--password <pw>]
                       [--overwrite|--no-overwrite] [--no-delete] [--check-list]
  zinc restore <name> --only <relative-path> [--dest <path>] [--from <pack|index>]
                       restore just that file/dir; add --check-list to list its versions
  zinc reindex <name> [--password <pw>] [--pack-start <name>] [--pack-end <name>] [--yes]
  zinc compact <name> [--password <pw>] [--pack <name|index>] [--min-gain <5%|8M>] [--yes] [--notify]
                       drop copies a newer version has superseded; same name, cloud overwrite
  zinc verify <name|--all> [--sample <pct>] [--password-only|--hash-only]
  zinc status [--all] [<name>]
  zinc completion <powershell|bash|zsh|fish> [--install]   generate a tab-completion script
  zinc schedule <name|--all> --daily --at HH:MM[,HH:MM...]
  zinc schedule <name|--all> --weekly 1,3 --at HH:MM
  zinc schedule <name|--all> --monthly 1,15 --at HH:MM [--yes]
  zinc schedule --list                 list all scheduled tasks
  zinc unschedule [<name>|--all]

Notes:
  - every pack is a complete encrypted 7z; its name is the packing time, restore goes by name
  - solid blocks default to 64M: one bad byte can cost the rest of that block; --solid off makes one block per file
  - without --password the password comes from repo > settings; backup/restore prompt only if none exists, schedule fails instead
  - schedule stores the resolved password in plain text inside the task; recreate the task after changing the password (zinc unschedule <name>, then schedule again)
  - restore overwrites same-name files by default; --no-overwrite keeps both
  - restore also applies deletions recorded by later backups; --no-delete keeps those files
  - --check-list only prints the order, --dest may be omitted
  - the index is a cache; zinc reindex rebuilds it from the packs themselves
  - output defaults to <root>\<name>; set it with zinc root
  - run zinc init once after unzipping; then just type zinc in any terminal
  - use zinc-register.exe / zinc-unregister.exe from the tools folder for install and removal; zinc init is the light version that does not require 7-Zip
  - unschedule <name> removes that repo's tasks; unschedule --all removes every zinc task; unschedule all removes only the --all summary task
`,

	ErrPrefix:         "[error] ",
	SaveConfigFailed:  "failed to save config: %v",
	LoadConfigFailed:  "failed to read %s: %w",
	ParseConfigFailed: "failed to parse %s: %w",
	RepoNotFound:      "repo not found: %s",

	TooManyArgs:          "unexpected argument: %s (see zinc --help)",
	AllWithName:          "--all cannot be combined with a name: %s",
	ListWithArgs:         "--list only queries tasks; drop the schedule flags",
	ConfirmUnscheduleAll: "this deletes every zinc task, continue? [y/n]: ",
	UnscheduleAllAborted: "cancelled",
	LangInvalid:          "[warn] unknown language %q, fell back to %s",

	LoadIndexFailed:  "failed to read index: %w",
	ParseIndexFailed: "failed to parse index: %w",

	CreateLockFailed: "failed to create lock: %w",
	WaitingLock:      "[wait] repo is in use by another task, waiting up to %v...\n",
	LockTimeout:      "repo has been busy for over %v.\nIf no compression is running, delete this file and retry:\n  %s",

	PromptPassword:      "Password: ",
	PromptPasswordAgain: "Password again: ",
	PasswordEmpty:       "password cannot be empty",
	PasswordMismatch:    "passwords do not match, please retry",
	PasswordTooMany:     "too many failed password confirmations",

	RootUnset:          "(not set)",
	UsageList:          "usage: zinc list [--names]",
	UsageStatus:        "usage: zinc status [--all] [<name>] (--all cannot be combined with a repo name)",
	UsageRoot:          "usage: zinc root [<path>]",
	RootNotDir:         "root path exists and is not a directory: %s",
	NoRepos:            "no repos yet, run zinc add first",
	ListItem:           "%s\n  source: %s\n  output: %s\n",
	UsageAdd:           "usage: zinc add <name> --source <path>",
	MissingSource:      "missing --source",
	RepoExists:         "repo %q already exists, use zinc set",
	RootNotSet:         "root is not set; run zinc root <path> or pass --output",
	WarnSourceMissing:  "[warn] source directory does not exist yet: %s\n",
	AddedRepo:          "added %s\n  source: %s\n  output: %s\n",
	UsageRemove:        "usage: zinc remove <name>",
	RemovedRepo:        "removed repo %s (packs and index are kept)\n",
	UsageSet:           "usage: zinc set <name> [--source <path>] [--output <path>] [--password <pw>]\n                    [--max-pack-byte <n>] [--compress <0-9>] [--solid <64M|off>] [--exclude <glob>]",
	NeedSourceOrOutput: "at least one setting is required: --source / --output / --password / --max-pack-byte / --compress / --solid / --exclude",
	UpdatedRepo:        "updated %s\n  source: %s\n  output: %s\n",
	PathInvalid:        "invalid path: %s (%v)",
	SourceOutputSame:   "source and output must differ: %s",
	OutputInsideSource: "output must not be inside source: %s is inside %s",
	SourceInsideOutput: "source must not be inside output: %s is inside %s",
	OutputInUse:        "output directory is already used by repo %q: %s",
	OutputCaseWarn:     "[warning] output %s differs only by case from repo %q's output; on Windows they are the same directory\n",
	RemoveHasTasks:     "[warning] repo %s still has %d scheduled task(s) (%s); remove them with zinc unschedule %s\n",

	FlagAllBackup:        "back up all repos",
	FlagAllStatus:        "show details for every repo at once",
	FlagPassword:         "backup password",
	FlagNotify:           "send a Windows notification",
	FlagNotifyEmpty:      "notify even when nothing changed",
	FlagNoVerify:         "skip 7za t self-check",
	UsageBackup:          "usage: zinc backup <name> or zinc backup --all",
	NoReposBackup:        "no repos",
	ScanErrSource:        "%s · source directory not found",
	ScanErrGeneric:       "%s · scan failed: %v",
	NoNew:                "%s · nothing new\n",
	BackupFailed:         "%s · failed: %v\n",
	ToastTitle:           "zinc backup complete",
	ToastHeadChanged:     "%d repo(s) changed, %d unchanged · %s",
	ToastHeadChangedFail: "%d repo(s) changed · %d failed · %s",
	ToastHeadFailed:      "%d repo(s) failed",
	ToastHeadNone:        "all %d repos unchanged",
	ToastFailedPrefix:    "failed: ",
	NotifyFailed:         "[notify failed] %v\n",
	SourceNotExist:       "source directory not found: %s",
	SourceEqualsOutput:   "source and output are the same directory, cannot back up: %s",
	CreateOutputFailed:   "failed to create output directory: %w",
	LogNoChange:          "backup %s: no changes",
	VerifyFailed:         "self-check failed, removed this batch: %w",
	LogVerifyFailed:      "backup %s: self-check failed %v",
	WriteListfileFailed:  "failed to write listfile: %w",
	CompressFailed:       "7za compression failed: %v\n%s",
	NoVolumes:            "7za succeeded but produced no packs",

	BackupPwMismatch:  "[warning] --password differs from the configured password of repo %s\n",
	ConfirmPwMismatch: "password differs from the repo/default password! continue? [y/n]: ",
	BackupPwAborted:   "cancelled: password differs from config",

	FlagDest:            "restore destination",
	UsageRestore:        "usage: zinc restore <name> --dest <path>",
	MissingDest:         "missing --dest",
	NoVolumesRestore:    "no packs found: %s",
	CreateDestFailed:    "failed to create restore directory: %w",
	Extracting:          "[extract] %s\n",
	ExtractFailed:       "failed to extract %s: %v\n%s",
	RestoreDone:         "[done] restored to %s\n",
	RestoreDestRoot:     "refusing to restore into a drive root: %s",
	RestoreDestSource:   "refusing to restore into the source directory (it would overwrite source files): %s",
	RestoreDestInOutput: "refusing to restore into the output directory or below it: %s (output %s)",

	FlagAllVerify:       "verify all repos",
	FlagSample:          "sample percentage 1-100",
	FlagPasswordOnly:    "check the password only, skip SHA-256",
	FlagHashOnly:        "check SHA-256 only, skip the password",
	VerifyFlagConflict:  "--password-only and --hash-only cannot be combined",
	VerifyNoPassword:    "no password available: set a repo password or settings.default_password",
	UsageVerify:         "usage: zinc verify <name|--all> [--sample <pct>] [--password-only|--hash-only]",
	NoManifest:          "no manifest (%s), cannot verify",
	ManifestBroken:      "manifest %s is corrupt or unreadable: %v",
	SampleInvalid:       "--sample must be between 1 and 100: %d",
	MissingFile:         "[missing] %s\n",
	CorruptFile:         "[corrupt] %s\n",
	VerifySummary:       "%s: passed %d, corrupt %d, missing %d\n",
	VerifyPwOnlySummary: "%s: password ok %d, wrong password %d, missing %d\n",
	VerifyFailedMsg:     "verification failed",
	VerifyPwBad:         "[wrong password] %s\n",
	VerifyPwSummary:     "[password] %d pack(s) cannot be opened with the configured password (the pack may be fine; the password is wrong)\n",
	VerifyPwNo7zip:      "[note] 7-Zip not found, skipping the password check\n",

	FlagAllSchedule:   "create one task for all repos",
	FlagDaily:         "daily",
	FlagWeekly:        "weekdays, 1-7, comma separated",
	FlagMonthly:       "days of month, comma separated",
	FlagAt:            "time HH:MM[,HH:MM...]",
	FlagNoNotify:      "task sends no notification",
	FlagYes:           "skip the monthly-day warning",
	NeedPeriod:        "one of --daily / --weekly / --monthly is required",
	NeedAt:            "missing --at HH:MM",
	UsageSchedule:     "usage: zinc schedule <name|--all> ...",
	TimeInvalid:       "invalid time: %q (expected HH:MM)",
	MonthlyDayInvalid: "invalid day of month: %q (expected 1-31)",
	MonthlyWarn:       "you set day %d of every month, but February has no day %d, and Apr/Jun/Sep/Nov only have 30 days.\nThose months will be skipped.\nAdd --yes to continue",
	UnknownPeriod:     "unknown period: %s",
	TaskCreateFailed:  "failed to create task %s: %v\n%s",
	TaskCreated:       "[task] %s (%s %s %s)\n",
	WeekdayInvalid:    "invalid weekday: %q (expected 1-7, 1=Monday)",
	QueryFailed:       "failed to query tasks: %v\n%s",
	Deleted:           "[deleted] %s\n",
	NoMatch:           "no matching tasks",
	FlagAllUnschedule: "delete all zinc tasks",
	UsageUnschedule:   "usage: zinc unschedule <name> or zinc unschedule --all",

	SchedulePwFromCLI:      "[password] from the command line\n",
	SchedulePwFromRepo:     "[password] from the repo config\n",
	SchedulePwFromSettings: "[password] from settings.json (default_password)\n",
	SchedulePwMissing:      "no usable password: pass --password, or set one with zinc config set default_password <pw>",
	SchedulePwMixed:        "the target repos do not share one password, so a single task cannot cover them; create separate tasks or unify the passwords",

	UsageCompletion:           "usage: zinc completion <powershell|bash|zsh|fish> [--install]",
	FlagCompletionInstall:     "PowerShell only: write the script and add a loader line to $PROFILE",
	FlagListNames:             "one repo name per line (for scripts and completion)",
	CompletionUnknownShell:    "unsupported shell: %s (choose powershell, bash, zsh or fish)",
	CompletionInstallOnlyPwsh: "--install only supports powershell; for other shells redirect zinc completion <shell> to a file",
	CompletionScriptWritten:   "[done] completion script: %s\n",
	CompletionProfileUpdated:  "[done] loader line written to: %s\n",
	CompletionProfileHasLine:  "[skip] %s already loads the script\n",
	CompletionInstallFailed:   "failed to install completion: %v",

	FlagListSchedule:   "list all scheduled tasks",
	ScheduleHeader:     "[schedule] %d task(s)\n\n",
	NoSchedule:         "no zinc scheduled tasks",
	NoScheduleHint:     "add one with zinc schedule <repo> --daily --at 03:00",
	ScheduleItemName:   "  %s\n",
	ScheduleItemDetail: "    %s · next %s · %s\n",
	CronDaily:          "daily %s",
	CronWeekly:         "weekly %s %s",
	CronMonthly:        "monthly %s %s",
	SchedNextNone:      "not scheduled",
	StateReady:         "ready",
	StateDisabled:      "disabled",
	StateRunning:       "running",
	StateQueued:        "queued",
	StateUnknown:       "unknown",

	StatusBrief:           "%-12s %6d files / %3d batches / %3d packs   last backup %-19s\n",
	StatusSource:          "  source: %s\n",
	StatusOutput:          "  output: %s\n",
	StatusBacked:          "  backed up: %d files\n",
	StatusPending:         "  pending: %d new, %d modified, %d to be deleted, %d skipped\n",
	StatusVolumes:         "  volumes: %d batches / %d files / %s\n",
	StatusLast:            "  last backup: %s\n",
	StatusManifestExists:  "  manifest: present\n",
	StatusManifestMissing: "  manifest: missing\n",

	RegisterFailed: "failed to register AUMID: %v\n%s",
	ToastFailed:    "toast failed: %v",

	UsageInit:       "usage: zinc init [--force]",
	FlagForce:       "rebuild the notification shortcut even if it is already registered",
	InitDir:         "program dir: %s\n",
	InitNotifyReady: "[ok] notification registration: %s\n",
	InitNotifyFail:  "[failed] notification registration: %v\n",
	InitPathAdded:   "[ok] added program dir to user PATH: %s (takes effect in new terminals)\n",
	InitPathExists:  "[skip] user PATH already contains: %s\n",
	InitPathFail:    "[failed] writing user PATH: %v\n",
	Init7zipOK:      "[ok] 7za: %s\n",
	Init7zipMissing: "[warn] 7z.exe / 7za.exe not found (set it with zinc config set 7zip_path)\n",
	InitHint:        "\nDone. Open a new terminal and type zinc; if notifications still fail, run zinc init --force again.\n",

	FlagInstall7zip:  "path to 7z.exe / 7za.exe",
	FlagNoPause:      "do not wait for a key press (for scripts)",
	FlagYesUninstall: "do not ask for confirmation",
	FlagKeepSchedule: "keep scheduled tasks (removed by default)",

	InstallHeader:    "zinc install\nprogram dir: %s\n\n",
	Install7zipOK:    "[ok] 7-Zip: %s\n",
	PromptSevenZip:   "7-Zip was not found. Install it first (https://www.7-zip.org/), or type the full path to 7z.exe / 7za.exe (empty to abort): ",
	SevenZipMissing:  "no usable 7-Zip. zinc needs 7z.exe or 7za.exe; install it and re-run zinc-register.exe, or pass --7zip <path>.",
	SevenZipNotFound: "not a usable 7-Zip path: %s",
	InstallDone:      "\nDone. Open a new terminal and type zinc. Next, pick a backup root:\n  zinc root D:\\Backup\\zinc\n",
	PressEnter:       "\nPress Enter to close this window...",

	UninstallHeader:        "zinc uninstall\nprogram dir: %s\n\n",
	UninstallTasks:         "%d scheduled task(s) will be removed:\n",
	UninstallTaskItem:      "  %s\n",
	UninstallNoTasks:       "[skip] no scheduled tasks point at this directory\n",
	UninstallTaskRemoved:   "[ok] removed %d scheduled task(s)\n",
	UninstallTaskFail:      "[failed] handling scheduled tasks: %v\n",
	UninstallKeepSchedule:  "[skip] scheduled tasks kept (--keep-schedule)\n",
	UninstallNoNotify:      "[skip] no notification registration\n",
	UninstallNotifyRemoved: "[ok] notification registration removed (shortcut, AppUserModelId, marker)\n",
	UninstallNotifySkip:    "[skip] notification registration belongs to another directory, left alone: %s\n",
	UninstallNotifyFail:    "[failed] removing notification registration: %v\n",
	UninstallPathRemoved:   "[ok] removed from user PATH: %s (takes effect in new terminals)\n",
	UninstallPathMissing:   "[skip] user PATH does not contain: %s\n",
	UninstallPathFail:      "[failed] writing user PATH: %v\n",
	UninstallConfirm:       "\nContinue? Backup data and settings are kept [y/N]: ",
	UninstallNeedYes:       "[failed] non-interactive run: pass --yes to uninstall",
	UninstallAbort:         "cancelled.\n",
	UninstallDone:          "\nUninstalled. The program files are still in %s; delete that folder whenever you like. Backup data (the .7z packs) was not touched.\n",

	LoadSettingsFailed:  "failed to read settings file %s: %w",
	ParseSettingsFailed: "failed to parse settings file %s: %w",
	ConfigUnknownKey:    "unknown setting: %s",
	ConfigBadValue:      "invalid value for %s: %q",
	UsageConfig:         "usage: zinc config [get|set|unset|path] [<key>] [<value>]",
	ConfigPath:          "settings file: %s\n",
	ConfigHeader:        "global settings (%s)\n",
	ConfigItem:          "  %-22s = %s\n",
	ConfigSet:           "set %s = %s\n",
	ConfigUnset:         "cleared %s (falls back to built-in default)\n",
	ConfigNotSet:        "%s is not set\n",
	ConfigPromptValue:   "value for %s: ",
	ConfigAuto:          "(auto)",
	ConfigEmpty:         "(empty)",

	IndexBadHeader:        "index file %s does not start with %q; it may be corrupt - run zinc reindex",
	IndexLineBad:          "index file %s has a bad line %d; run zinc reindex",
	IndexTruncated:        "index file %s is incomplete (no end marker); run zinc reindex",
	IndexCountMismatch:    "index file %s count mismatch (recorded %d packs/%d files, found %d/%d); run zinc reindex",
	IndexWriteFailed:      "failed to write index: %w",
	ManifestCorrupt:       "manifest %s is corrupt, renamed to %s; backup aborted: %v",
	ManifestCorruptLocked: "manifest %s is corrupt and could not be renamed: %v",
	ManifestReadFailed:    "failed to read manifest: %w",
	ManifestWriteFailed:   "failed to write manifest: %w",

	PackTooBig:       "[warn] %s (%s) exceeds pack limit %s, packed alone\n",
	ClockRollback:    "[warn] system clock went backwards; using %s as pack time\n",
	RenamePackFailed: "failed to rename finished pack: %w",
	DetailLine:       "[ok] %s · %d files · %s · %d packs · %s · scan %dms · hash %s",
	EncFrom:          "enc✓(%s)",
	ScanSkipItem:     "[skip] %s\n",
	ScanSkipMore:     "  …and %d more skipped\n",
	LockedItem:       "[locked] %s (will retry next backup)\n",
	LockedMore:       "  …and %d more locked files\n",
	AllLocked:        "%s · all %d changed files are locked, skipped this run\n",
	PackSkipItem:     "[skip] %s (unreadable while packing, will retry next backup)\n",
	PackSkipMore:     "  …and %d more files left out of this pack\n",
	PartialCommit:    "%s · partial: committed %d pack(s) (%s), %d files left for next run\n",

	DeletedLine:          "[deleted] %s · %d file(s) disappeared from the source (recorded in %s)\n",
	DeletingBatch:        "[deleted] %s · %d file(s) removed\n",
	RestoreDelSummary:    "[done] %d file(s) removed according to the deletion records\n",
	RestoreDelMissing:    "  %d more path(s) in the deletion records were already absent (skipped; normal after compact drops old copies)\n",
	FlagNoDelete:         "restore without applying deletions recorded after the backup",
	FlagNotifyWarn:       "send a Windows notification when a restore is incomplete",
	ToastRestoreTitle:    "zinc restore incomplete",
	ToastRestoreBody:     "%s · %d pack(s) reported errors; files may be incomplete — run zinc verify",
	CheckListDelItem:     "  %2d. %-44s %8d deletion(s)\n",
	DelRecordBadPath:     "unsafe path %q in deletion record %s",
	DelRecordBadHeader:   "deletion record %s has a bad header (expected %q)",
	DelRecordNoSum:       "deletion record %s has no checksum line (# sha256 ...)",
	DelRecordAllBad:      "every line of deletion record %s is unusable (path or line checksum damaged)",
	DelRecordEdited:      "[warn] deletion record %s no longer matches its recorded hash (probably edited by hand); every line checks out, using the current content\n",
	DelLineSkipped:       "[warn] deletion record %s has an unusable line, skipped: %s\n",
	DelSkipped:           "[warn] deletion records unusable, ignoring them because of --no-delete: %v\n",
	DelSkippedAfterFail:  "[skipped] %s · its batch's pack reported errors, so the deletion record is skipped too\n",
	PackFailed:           "[warn] %s · errors during extraction (whatever could be extracted is in place); continuing with the following packs\n",
	PackErrLine:          "         %s\n",
	PackPasswordHint:     "[hint] if every pack reports this, the password is probably wrong rather than the packs being damaged; check --password or default_password\n",
	RestoreIncomplete:    "[warn] %d pack(s) could not be extracted; the result is incomplete, and files that failed their CRC may be on disk with wrong content — check with zinc verify\n",
	RestoreFailedPacks:   "%d pack(s) failed to extract",
	DelRecordReadFailed:  "failed to read deletion record %s: %v",
	DelRecordWriteFailed: "failed to write deletion record: %w",
	DelRemoveFailed:      "failed to remove %s: %v",
	VerifyRmMissing:      "[missing] deletion record %s\n",
	VerifyRmCorrupt:      "[corrupt] deletion record %s\n",
	VerifyRmSummary:      "deletion records: %d ok, %d corrupt, %d missing\n",

	FlagOverwrite:        "overwrite files with the same name on restore",
	FlagNoOverwrite:      "keep both: same-name files are restored as _1",
	FlagCheckList:        "only list the packs and their order, do not touch files",
	FlagConflict:         "--overwrite and --no-overwrite cannot be used together",
	ModeOverwrite:        "overwrite",
	ModeRename:           "no-overwrite",
	DestUnset:            "(not specified)",
	RestoreOverwriteWarn: "[warn] overwrite mode: same-name files in the destination will be replaced\n",
	RestoreWrapLine:      "[wrap] restoring into %s (restore_wrap_folder=on)\n",
	CheckListHeader:      "restore order (packs %d, deletion records %d) · mode: %s · dest: %s\n",
	CheckListItem:        "  %2d. %-44s %10s\n",
	CheckListTotal:       "total %s\n",
	PackOrderBroken:      "pack order is not monotonic: %s and %s cannot be ordered. Check the file names in the output directory; restore aborted",
	NoPacksRestore:       "no packs found: %s",
	FlagOnly:             "restore only this relative path (file or directory), leave the rest alone",
	FlagFrom:             "restore from a specific pack: name, name prefix or index (1-based)",
	FromRequiresOnly:     "--from can only be used together with --only",
	RestorePartial:       "[warn] restore aborted; the destination may be partially restored\n",
	MissingOnly:          "--only requires a non-empty path",
	PathNotBackedUp:      "path not found in the index or in any pack: %s",
	PackNotSpecified:     "pack not found: %s",
	OnlyOrigin:           "[from] %s\n",
	OnlyDone:             "[done] %s → %s\n",
	OnlyVersions:         "packs containing %s (%d, oldest first)\n",
	OnlyNoVersions:       "no pack contains that path: %s",

	FlagPackStart:     "rebuild only from this pack name onwards",
	FlagPackEnd:       "rebuild only up to this pack name",
	FlagYesReindex:    "overwrite the existing index (a .bak is kept)",
	UsageReindex:      "usage: zinc reindex <name> [--password <pw>] [--pack-start <name>] [--pack-end <name>] [--yes]",
	ReindexStart:      "[reindex] scanning %d packs (headers only)\n",
	ReindexPack:       "  %s · %d files\n",
	ReindexDone:       "[done] %d packs / %d files → %s\n",
	ReindexExists:     "index already exists: %s\nadd --yes to overwrite (a .bak is kept)",
	ReindexListFailed: "failed to list pack %s: %v",
	ReindexNoPacks:    "no packs found: %s",
	ReindexBakFailed:  "failed to back up the old index; aborted: %w",
	ReindexDeleted:    "  %d file(s) dropped by the deletion records\n",

	FlagCompactPack:        "rewrite only this pack (name or index)",
	FlagCompactMinGain:     "skip rewrites that save less than this: 5% or 8M",
	FlagYesCompact:         "actually rewrite (without it, only report)",
	UsageCompact:           "usage: zinc compact <name> [--pack <name|index>] [--min-gain <5%%|8M>] [--yes] [--notify] [--password <pw>]",
	CompactStart:           "[compact] %s: checking %d packs (headers only)\n",
	CompactDryRunHead:      "[dry-run] no --yes, nothing will be changed\n",
	CompactDryRunTail:      "add --yes to do it.\n",
	CompactPlanPack:        "  %s · %d dead file(s) / %s (%.1f%%) → will rewrite\n",
	CompactSkipNone:        "  %s · nothing to drop\n",
	CompactSkipSmall:       "  %s · %d dead file(s) / %s (%.1f%%), below the %s threshold → skipped\n",
	CompactPlanSummary:     "[plan] %d pack(s) to rewrite, about %s to reclaim\n",
	CompactNothing:         "[done] nothing to rewrite (%d pack(s) with no dead copies, %d below threshold)\n",
	CompactDoing:           "[rewrite] %s: dropping %d file(s) / %s\n",
	CompactEmptyDoing:      "[rewrite] %s: every file was superseded, the pack will only hold %s\n",
	CompactPackDone:        "  %s · %s → %s\n",
	CompactDoneLine:        "[done] %d pack(s) rewritten, %d failed, %s → %s, %s reclaimed\n",
	CompactUploadHint:      "note: re-upload these packs to the cloud drive; same-name overwrite moves the old pack into version history.\n",
	CompactFailedMsg:       "%d pack(s) failed to rewrite; the originals were left untouched",
	CompactNoPacks:         "no packs found: %s",
	CompactPackFailed:      "failed to process pack %s: %v\n",
	CompactPackNotFound:    "no such pack: %s",
	CompactIndexStale:      "the index does not match the pack directory (index %d packs, directory %d packs): %s\nrun zinc reindex %s to rebuild it completely first",
	CompactPackListFailed:  "failed to list pack %s: %v",
	CompactExtractFailed:   "failed to extract pack %s (original untouched): %v\n%s",
	CompactExtractMismatch: "%s from pack %s did not extract intact (size mismatch); pack skipped (original untouched)",
	CompactRebuildFailed:   "failed to rebuild %s (original untouched): %v",
	CompactVerifyFailed:    "the rebuilt pack %s failed its self-check and was discarded (original untouched): %v",
	CompactCountMismatch:   "the rebuilt pack %s does not match: %d file(s) expected, %d found; discarded (original untouched)",
	CompactNoSpace:         "not enough free space %s to process %s (about %s needed); skipped",
	CompactBadGain:         "bad gain threshold: %s (use 5%% or 8M)",
	PlaceholderRemoved:     "  removed %d placeholder file(s) (%s)\n",
	FlagNotifyCompact:      "send a Windows notification when done",
	ToastCompactTitle:      "zinc compact complete",
	ToastCompactBody:       "%s · %d pack(s) rewritten · %s reclaimed",
	ToastCompactFailBody:   "%s · %d pack(s) failed to rewrite, the rest went through; see the terminal",

	FlagMaxPackByte: "pack size limit, e.g. 4G / 50M / bytes",
	FlagCompress:    "compression level 0-9",
	FlagSolid:       "solid block limit, e.g. 64M / off (off = one block per file, most damage-tolerant)",
	FlagExclude:     "exclude pattern, repeatable",

	StatusPacks:   "  packs: %d batches / %d packs / %s\n",
	StatusContent: "  content dates: %s ~ %s\n",
}
