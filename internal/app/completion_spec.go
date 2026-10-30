package app

// 补全用的 CLI 描述表：子命令、参数、位置参数补什么。
// 这是补全脚本的唯一事实来源；改 CLI 时同步改这里，completion_test.go 会拿
// `zinc <cmd> --help` 的输出跟本表对账，漏改会红。
type FlagSpec struct {
	Name     string   // 不带前导横线，与 flag.NewFlagSet 注册名一致
	HasValue bool     // true = 需要值
	Values   []string // 固定取值；空表示自由输入
}

type CmdSpec struct {
	Name            string
	Aliases         []string
	Flags           []FlagSpec
	PositionalKind  string   // "repo" = 仓库名；"path" = 路径；"" = 不补
	PositionalWords []string // 固定词表（config 的动作词、completion 的 shell 名）
}

// globalFlags 是所有子命令都能用的全局参数。
var globalFlags = []FlagSpec{
	{Name: "lang", HasValue: true, Values: []string{"zh", "en"}},
	{Name: "help"},
}

var completionCommands = []CmdSpec{
	{Name: "init", Flags: []FlagSpec{{Name: "force"}}},
	{Name: "root", PositionalKind: "path"},
	{Name: "list", Flags: []FlagSpec{{Name: "names"}}},
	{Name: "add", Flags: []FlagSpec{{Name: "source", HasValue: true}, {Name: "output", HasValue: true}}, PositionalKind: "repo"},
	{Name: "remove", PositionalKind: "repo"},
	{Name: "set", Flags: []FlagSpec{
		{Name: "source", HasValue: true},
		{Name: "output", HasValue: true},
		{Name: "password", HasValue: true},
		{Name: "max-pack-byte", HasValue: true},
		{Name: "compress", HasValue: true},
		{Name: "solid", HasValue: true},
		{Name: "exclude", HasValue: true},
	}, PositionalKind: "repo"},
	{Name: "config", PositionalWords: []string{"get", "set", "unset", "path", "list", "show"}},
	{Name: "backup", Flags: []FlagSpec{
		{Name: "all"},
		{Name: "password", HasValue: true},
		{Name: "notify"},
		{Name: "notify-empty"},
		{Name: "no-verify"},
	}, PositionalKind: "repo"},
	{Name: "restore", Flags: []FlagSpec{
		{Name: "dest", HasValue: true},
		{Name: "password", HasValue: true},
		{Name: "overwrite"},
		{Name: "no-overwrite"},
		{Name: "check-list"},
		{Name: "no-delete"},
		{Name: "notify"},
		{Name: "only", HasValue: true},
		{Name: "from", HasValue: true},
	}, PositionalKind: "repo"},
	{Name: "reindex", Flags: []FlagSpec{
		{Name: "password", HasValue: true},
		{Name: "pack-start", HasValue: true},
		{Name: "pack-end", HasValue: true},
		{Name: "yes"},
	}, PositionalKind: "repo"},
	{Name: "compact", Flags: []FlagSpec{
		{Name: "password", HasValue: true},
		{Name: "pack", HasValue: true},
		{Name: "min-gain", HasValue: true},
		{Name: "yes"},
		{Name: "no-verify"},
		{Name: "notify"},
	}, PositionalKind: "repo"},
	{Name: "verify", Flags: []FlagSpec{{Name: "all"}, {Name: "sample", HasValue: true}}, PositionalKind: "repo"},
	{Name: "status", Flags: []FlagSpec{{Name: "all"}}, PositionalKind: "repo"},
	{Name: "schedule", Flags: []FlagSpec{
		{Name: "list"},
		{Name: "all"},
		{Name: "daily"},
		{Name: "weekly", HasValue: true},
		{Name: "monthly", HasValue: true},
		{Name: "at", HasValue: true},
		{Name: "password", HasValue: true},
		{Name: "no-notify"},
		{Name: "yes"},
	}, PositionalKind: "repo"},
	{Name: "unschedule", Flags: []FlagSpec{{Name: "all"}}, PositionalKind: "repo"},
	{Name: "completion", Flags: []FlagSpec{{Name: "install"}}, PositionalWords: []string{"powershell", "bash", "zsh", "fish"}},
	{Name: "help"},
}

// repoCommands 返回需要补仓库名的子命令。
func repoCommands() []string {
	var out []string
	for _, c := range completionCommands {
		if c.PositionalKind == "repo" {
			out = append(out, c.Name)
		}
	}
	return out
}
