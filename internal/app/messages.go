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
	FlagAllBackup       string
	FlagPassword        string
	FlagNotify          string
	FlagNotifyEmpty     string
	FlagNoVerify        string
	UsageBackup         string
	NoReposBackup       string
	ScanErrSource       string
	ScanErrGeneric      string
	NoNew               string
	BackupFailed        string
	SummaryFailed       string
	SummaryNoNew        string
	SummaryLine         string
	SummaryMore         string
	ToastTitle          string
	NotifyFailed        string
	SourceNotExist      string
	SourceEqualsOutput  string
	CreateOutputFailed  string
	LogNoChange         string
	VerifyFailed        string
	LogVerifyFailed     string
	WriteListfileFailed string
	CompressFailed      string
	NoVolumes           string

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
	FlagAllVerify   string
	FlagSample      string
	UsageVerify     string
	NoManifest      string
	ManifestBroken  string
	SampleInvalid   string
	MissingFile     string
	CorruptFile     string
	VerifySummary   string
	VerifyFailedMsg string

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

	// 删除记录
	DeletedLine          string
	DeletingBatch        string
	RestoreDelSummary    string
	FlagNoDelete         string
	CheckListDelItem     string
	DelRecordBadPath     string
	DelRecordBadHeader   string
	DelRecordReadFailed  string
	DelRecordWriteFailed string
	DelRemoveFailed      string
	VerifyRmMissing      string

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

	// set flags
	FlagMaxPackByte string
	FlagCompress    string
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
                  [--max-pack-byte <n>] [--compress <0-9>] [--exclude <glob>]
  zinc config [get|set|unset|path] [<key>] [<value>]      全局设置
  zinc backup <name|--all> [--password <pw>] [--notify] [--notify-empty] [--no-verify]
  zinc restore <name> [--dest <path>] [--password <pw>]
                       [--overwrite|--no-overwrite] [--no-delete] [--check-list]
  zinc restore <name> --only <相对路径> [--dest <path>] [--from <包名|序号>]
                       只恢复这一个文件/目录；配 --check-list 可列出它有哪些版本
  zinc reindex <name> [--password <pw>] [--pack-start <名>] [--pack-end <名>] [--yes]
  zinc verify <name|--all> [--sample <pct>]
  zinc status [<name>]
  zinc schedule <name|--all> --daily --at HH:MM[,HH:MM...]
  zinc schedule <name|--all> --weekly 1,3 --at HH:MM
  zinc schedule <name|--all> --monthly 1,15 --at HH:MM [--yes]
  zinc unschedule [<name>|--all]

说明:
  - 每个包都是独立完整的加密 7z，文件名就是打包时刻，恢复按名字升序
  - 不给 --password 时交互输入；settings.json 里设了 default_password 就不再询问
  - 恢复默认覆盖同名文件（--no-overwrite 则保留两份），--check-list 只列顺序不动手
  - 恢复会执行之后备份记下的删除（--no-delete 保留这些文件）；删除记录是输出目录里的 .rm 旁注
  - 索引只是缓存，丢了用 zinc reindex 从包本身重建
  - 输出默认落在 <root>\<name>，用 zinc root 设置
  - 解压后先跑一次 zinc init，之后在任意终端直接敲 zinc
  - 安装/卸载用 tools\ 下的 zinc-register.exe / zinc-unregister.exe；zinc init 是轻量版，不强制 7-Zip
`,

	ErrPrefix:         "[错误] ",
	SaveConfigFailed:  "保存配置失败: %v",
	LoadConfigFailed:  "读取 %s 失败: %w",
	ParseConfigFailed: "解析 %s 失败: %w",
	RepoNotFound:      "仓库不存在: %s",

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
	UsageList:          "用法: zinc list",
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
	UsageSet:           "用法: zinc set <name> [--source <path>] [--output <path>]",
	NeedSourceOrOutput: "至少给一个 --source 或 --output",
	UpdatedRepo:        "已更新 %s\n  源:   %s\n  输出: %s\n",
	PathInvalid:        "路径无效: %s（%v）",
	SourceOutputSame:   "源目录与输出目录不能相同: %s",
	OutputInsideSource: "输出目录不能位于源目录内: %s 在 %s 内",
	SourceInsideOutput: "源目录不能位于输出目录内: %s 在 %s 内",
	OutputInUse:        "输出目录已被仓库 %q 使用: %s",
	OutputCaseWarn:     "[警告] 输出目录 %s 与仓库 %q 的输出只有大小写不同，Windows 上会落到同一个目录\n",
	RemoveHasTasks:     "[警告] 仓库 %s 还有 %d 个计划任务（%s），可用 zinc unschedule %s 删除\n",

	FlagAllBackup:       "备份全部仓库",
	FlagPassword:        "备份密码",
	FlagNotify:          "发送 Windows 通知",
	FlagNotifyEmpty:     "无新增也通知",
	FlagNoVerify:        "跳过 7za t 自检",
	UsageBackup:         "用法: zinc backup <name> 或 zinc backup --all",
	NoReposBackup:       "没有仓库",
	ScanErrSource:       "%s · 源目录不存在",
	ScanErrGeneric:      "%s · 扫描失败: %v",
	NoNew:               "%s · 无新增\n",
	BackupFailed:        "%s · 失败: %v\n",
	SummaryFailed:       " · 失败",
	SummaryNoNew:        " · 无新增",
	SummaryLine:         "%s · %d 文件 · %s · %s",
	SummaryMore:         "\n等 %d 个仓库",
	ToastTitle:          "zinc 备份完成",
	NotifyFailed:        "[通知失败] %v\n",
	SourceNotExist:      "源目录不存在: %s",
	SourceEqualsOutput:  "源目录与输出目录相同，无法备份: %s",
	CreateOutputFailed:  "创建输出目录失败: %w",
	LogNoChange:         "backup %s: 无变化",
	VerifyFailed:        "自检失败，已删除本批次卷: %w",
	LogVerifyFailed:     "backup %s: 自检失败 %v",
	WriteListfileFailed: "写 listfile 失败: %w",
	CompressFailed:      "7za 压缩失败: %v\n%s",
	NoVolumes:           "7za 执行成功但未生成任何包",

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

	FlagAllVerify:   "校验全部仓库",
	FlagSample:      "抽样百分比 1-100",
	UsageVerify:     "用法: zinc verify <name> 或 zinc verify --all",
	NoManifest:      "没有 manifest（%s），无法校验",
	ManifestBroken:  "manifest %s 损坏或无法读取: %v",
	SampleInvalid:   "--sample 必须在 1-100 之间: %d",
	MissingFile:     "[缺失] %s\n",
	CorruptFile:     "[损坏] %s\n",
	VerifySummary:   "%s: 通过 %d, 损坏 %d, 缺失 %d\n",
	VerifyFailedMsg: "%s 校验未通过",

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
	SevenZipMissing:  "[失败] 没有可用的 7-Zip。zinc 依赖 7z.exe 或 7za.exe；装好后重跑 zinc-register.exe，或用 --7zip <路径> 指定。",
	SevenZipNotFound: "[失败] 7-Zip 路径不可用: %s",
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

	DeletedLine:          "[删除] %s · %d 个文件已从源目录消失（记录 %s）\n",
	DeletingBatch:        "[删除] %s · 移除 %d 个文件\n",
	RestoreDelSummary:    "[完成] 按删除记录移除了 %d 个文件\n",
	FlagNoDelete:         "恢复时不执行备份之后记录的删除",
	CheckListDelItem:     "  %2d. %-44s %8d 项删除\n",
	DelRecordBadPath:     "删除记录里有非法路径 %q（记录 %s），已中止",
	DelRecordBadHeader:   "删除记录 %s 的表头不对（应为 %q），已中止",
	DelRecordReadFailed:  "读取删除记录 %s 失败: %v",
	DelRecordWriteFailed: "写入删除记录失败: %w",
	DelRemoveFailed:      "删除 %s 失败: %v",
	VerifyRmMissing:      "[警告] 删除记录缺失: %s\n",

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

	FlagMaxPackByte: "单包上限，如 4G / 50M / 纯字节",
	FlagCompress:    "压缩等级 0-9",
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
                  [--max-pack-byte <n>] [--compress <0-9>] [--exclude <glob>]
  zinc config [get|set|unset|path] [<key>] [<value>]      global settings
  zinc backup <name|--all> [--password <pw>] [--notify] [--notify-empty] [--no-verify]
  zinc restore <name> [--dest <path>] [--password <pw>]
                       [--overwrite|--no-overwrite] [--no-delete] [--check-list]
  zinc restore <name> --only <relative-path> [--dest <path>] [--from <pack|index>]
                       restore just that file/dir; add --check-list to list its versions
  zinc reindex <name> [--password <pw>] [--pack-start <name>] [--pack-end <name>] [--yes]
  zinc verify <name|--all> [--sample <pct>]
  zinc status [<name>]
  zinc schedule <name|--all> --daily --at HH:MM[,HH:MM...]
  zinc schedule <name|--all> --weekly 1,3 --at HH:MM
  zinc schedule <name|--all> --monthly 1,15 --at HH:MM [--yes]
  zinc unschedule [<name>|--all]

Notes:
  - every pack is a complete encrypted 7z; its name is the packing time, restore goes by name
  - without --password it prompts; a default_password in settings.json skips the prompt
  - restore overwrites same-name files by default; --no-overwrite keeps both
  - restore also applies deletions recorded by later backups; --no-delete keeps those files
  - --check-list only prints the order, --dest may be omitted
  - the index is a cache; zinc reindex rebuilds it from the packs themselves
  - output defaults to <root>\<name>; set it with zinc root
  - run zinc init once after unzipping; then just type zinc in any terminal
  - use zinc-register.exe / zinc-unregister.exe from the tools folder for install and removal; zinc init is the light version that does not require 7-Zip
`,

	ErrPrefix:         "[error] ",
	SaveConfigFailed:  "failed to save config: %v",
	LoadConfigFailed:  "failed to read %s: %w",
	ParseConfigFailed: "failed to parse %s: %w",
	RepoNotFound:      "repo not found: %s",

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
	UsageList:          "usage: zinc list",
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
	UsageSet:           "usage: zinc set <name> [--source <path>] [--output <path>]",
	NeedSourceOrOutput: "at least one of --source or --output is required",
	UpdatedRepo:        "updated %s\n  source: %s\n  output: %s\n",
	PathInvalid:        "invalid path: %s (%v)",
	SourceOutputSame:   "source and output must differ: %s",
	OutputInsideSource: "output must not be inside source: %s is inside %s",
	SourceInsideOutput: "source must not be inside output: %s is inside %s",
	OutputInUse:        "output directory is already used by repo %q: %s",
	OutputCaseWarn:     "[warning] output %s differs only by case from repo %q's output; on Windows they are the same directory\n",
	RemoveHasTasks:     "[warning] repo %s still has %d scheduled task(s) (%s); remove them with zinc unschedule %s\n",

	FlagAllBackup:       "back up all repos",
	FlagPassword:        "backup password",
	FlagNotify:          "send a Windows notification",
	FlagNotifyEmpty:     "notify even when nothing changed",
	FlagNoVerify:        "skip 7za t self-check",
	UsageBackup:         "usage: zinc backup <name> or zinc backup --all",
	NoReposBackup:       "no repos",
	ScanErrSource:       "%s · source directory not found",
	ScanErrGeneric:      "%s · scan failed: %v",
	NoNew:               "%s · nothing new\n",
	BackupFailed:        "%s · failed: %v\n",
	SummaryFailed:       " · failed",
	SummaryNoNew:        " · nothing new",
	SummaryLine:         "%s · %d files · %s · %s",
	SummaryMore:         "\nand %d more repos",
	ToastTitle:          "zinc backup complete",
	NotifyFailed:        "[notify failed] %v\n",
	SourceNotExist:      "source directory not found: %s",
	SourceEqualsOutput:  "source and output are the same directory, cannot back up: %s",
	CreateOutputFailed:  "failed to create output directory: %w",
	LogNoChange:         "backup %s: no changes",
	VerifyFailed:        "self-check failed, removed this batch: %w",
	LogVerifyFailed:     "backup %s: self-check failed %v",
	WriteListfileFailed: "failed to write listfile: %w",
	CompressFailed:      "7za compression failed: %v\n%s",
	NoVolumes:           "7za succeeded but produced no packs",

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

	FlagAllVerify:   "verify all repos",
	FlagSample:      "sample percentage 1-100",
	UsageVerify:     "usage: zinc verify <name> or zinc verify --all",
	NoManifest:      "no manifest (%s), cannot verify",
	ManifestBroken:  "manifest %s is corrupt or unreadable: %v",
	SampleInvalid:   "--sample must be between 1 and 100: %d",
	MissingFile:     "[missing] %s\n",
	CorruptFile:     "[corrupt] %s\n",
	VerifySummary:   "%s: passed %d, corrupt %d, missing %d\n",
	VerifyFailedMsg: "%s verification failed",

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
	SevenZipMissing:  "[failed] no usable 7-Zip. zinc needs 7z.exe or 7za.exe; install it and re-run zinc-register.exe, or pass --7zip <path>.",
	SevenZipNotFound: "[failed] not a usable 7-Zip path: %s",
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

	DeletedLine:          "[deleted] %s · %d file(s) disappeared from the source (recorded in %s)\n",
	DeletingBatch:        "[deleted] %s · %d file(s) removed\n",
	RestoreDelSummary:    "[done] %d file(s) removed according to the deletion records\n",
	FlagNoDelete:         "restore without applying deletions recorded after the backup",
	CheckListDelItem:     "  %2d. %-44s %8d deletion(s)\n",
	DelRecordBadPath:     "unsafe path %q in deletion record %s; aborted",
	DelRecordBadHeader:   "deletion record %s has a bad header (expected %q); aborted",
	DelRecordReadFailed:  "failed to read deletion record %s: %v",
	DelRecordWriteFailed: "failed to write deletion record: %w",
	DelRemoveFailed:      "failed to remove %s: %v",
	VerifyRmMissing:      "[warn] deletion record is missing: %s\n",

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

	FlagMaxPackByte: "pack size limit, e.g. 4G / 50M / bytes",
	FlagCompress:    "compression level 0-9",
	FlagExclude:     "exclude pattern, repeatable",

	StatusPacks:   "  packs: %d batches / %d packs / %s\n",
	StatusContent: "  content dates: %s ~ %s\n",
}
