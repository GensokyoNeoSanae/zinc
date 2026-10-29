package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var weekdayMap = map[int]string{
	1: "MON", 2: "TUE", 3: "WED", 4: "THU", 5: "FRI", 6: "SAT", 7: "SUN",
}

// normalizeTime 校验 HH:MM 并补零成 schtasks 能接受的格式（9:5 → 09:05）。
func normalizeTime(t string) (string, error) {
	parts := strings.Split(strings.TrimSpace(t), ":")
	if len(parts) != 2 {
		return "", fmt.Errorf(M.TimeInvalid, t)
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return "", fmt.Errorf(M.TimeInvalid, t)
	}
	return fmt.Sprintf("%02d:%02d", h, m), nil
}

func exePath() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return "zinc.exe"
}

// scheduledExePath 优先返回同目录的 zincw.exe（GUI 子系统，任务计划执行时无控制台窗口）
func scheduledExePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "zincw.exe"
	}
	w := filepath.Join(filepath.Dir(exe), "zincw.exe")
	if _, statErr := os.Stat(w); statErr == nil {
		return w
	}
	return exe
}

// runSchedule 创建计划任务。kind 为 daily/weekly/monthly。
func runSchedule(name string, all bool, kind, days, times, password string, noNotify, yes bool) error {
	var normTimes []string
	for _, t := range strings.Split(times, ",") {
		n, err := normalizeTime(t)
		if err != nil {
			return err
		}
		normTimes = append(normTimes, n)
	}

	if kind == "monthly" {
		for _, d := range strings.Split(days, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(d))
			if err != nil || n < 1 || n > 31 {
				return fmt.Errorf(M.MonthlyDayInvalid, d)
			}
			if n > 28 && !yes {
				return fmt.Errorf(M.MonthlyWarn, n, n)
			}
		}
	}

	target := name
	if all {
		target = "--all"
	}
	cmdArgs := []string{scheduledExePath(), "backup", target, "--password", password}
	if !noNotify {
		cmdArgs = append(cmdArgs, "--notify")
	}
	escaped := make([]string, len(cmdArgs))
	for i, a := range cmdArgs {
		escaped[i] = syscall.EscapeArg(a)
	}
	baseCmd := strings.Join(escaped, " ")

	for _, t := range normTimes {
		taskName := fmt.Sprintf("zinc-%s-%s-%s", strings.TrimPrefix(target, "--"), kind, strings.ReplaceAll(t, ":", ""))

		args := []string{"/Create", "/TN", taskName, "/TR", baseCmd, "/ST", t, "/F"}
		switch kind {
		case "daily":
			args = append(args, "/SC", "DAILY")
		case "weekly":
			ds, err := weekdayList(days)
			if err != nil {
				return err
			}
			args = append(args, "/SC", "WEEKLY", "/D", ds)
		case "monthly":
			args = append(args, "/SC", "MONTHLY", "/D", strings.ReplaceAll(days, " ", ""))
		default:
			return fmt.Errorf(M.UnknownPeriod, kind)
		}

		out, err := exec.Command("schtasks", args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf(M.TaskCreateFailed, taskName, err, string(out))
		}
		fmt.Printf(M.TaskCreated, taskName, kind, days, t)
		logf("schedule %s: %s %s %s", taskName, kind, days, t)
	}
	return nil
}

func weekdayList(days string) (string, error) {
	var out []string
	for _, d := range strings.Split(days, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil || n < 1 || n > 7 {
			return "", fmt.Errorf(M.WeekdayInvalid, d)
		}
		out = append(out, weekdayMap[n])
	}
	return strings.Join(out, ","), nil
}

// runPowerShell 用固定脚本 + 环境变量传参。
// 不落 .ps1、不依赖执行策略，用户数据只出现在 $env:ZINC_TASK_PREFIX 里。
const taskPrefixEnv = "ZINC_TASK_PREFIX"

func runPowerShell(script, prefix string) ([]byte, error) {
	cmd := exec.Command(powershellPath(), "-NoProfile", "-Command", script)
	cmd.Env = append(os.Environ(), taskPrefixEnv+"="+prefix)
	return cmd.CombinedOutput()
}

// runUnschedule 用 PowerShell 查询并删除任务。
// schtasks 的输出是 GBK，Go 按 UTF-8 读会让中文任务名变乱码、匹配失败。
func runUnschedule(name string, all bool) error {
	prefix := "zinc-" + name + "-"
	if all || name == "" {
		prefix = "zinc-"
	}

	const script = `
$ErrorActionPreference = 'SilentlyContinue'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$p = $env:ZINC_TASK_PREFIX
$tasks = Get-ScheduledTask -TaskName 'zinc-*' | Where-Object { $_.TaskName.StartsWith($p, [System.StringComparison]::OrdinalIgnoreCase) }
if (-not $tasks) { Write-Output 'NONE'; exit 0 }
foreach ($t in $tasks) {
    Write-Output $t.TaskName
    Unregister-ScheduledTask -TaskName $t.TaskName -Confirm:$false -ErrorAction SilentlyContinue
}
`
	out, err := runPowerShell(script, prefix)
	if err != nil {
		return fmt.Errorf(M.QueryFailed, err, string(out))
	}

	deleted := 0
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "NONE" || !strings.HasPrefix(line, "zinc-") {
			continue
		}
		fmt.Printf(M.Deleted, line)
		deleted++
	}
	if deleted == 0 {
		fmt.Println(M.NoMatch)
	}
	logf("unschedule prefix=%s deleted=%d", prefix, deleted)
	return nil
}

// findScheduledTasks 返回匹配前缀的 zinc 任务名，供 remove 时提示残留任务。
func findScheduledTasks(prefix string) ([]string, error) {
	const script = `
$ErrorActionPreference = 'SilentlyContinue'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$p = $env:ZINC_TASK_PREFIX
Get-ScheduledTask -TaskName 'zinc-*' | Where-Object { $_.TaskName.StartsWith($p, [System.StringComparison]::OrdinalIgnoreCase) } | ForEach-Object { $_.TaskName }
`
	out, err := runPowerShell(script, prefix)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		if s := strings.TrimSpace(line); strings.HasPrefix(s, "zinc-") {
			names = append(names, s)
		}
	}
	return names, nil
}

// ---------- schedule --list ----------

// schedRow 是 zinc schedule --list 里的一行（一个计划任务）。
type schedRow struct {
	name  string
	state string
	next  string
	cls   string // 触发器 CIM 类名，如 MSFT_TaskDailyTrigger
	start string // StartBoundary，如 2026-10-28T03:00:00
	week  string // 周触发器的 bitmask
	month string // 月触发器的日号，逗号分隔
}

// listTasksScript 每个任务输出一行 TSV：
// Name / State / NextRunTime / TriggerClass / StartBoundary / DaysOfWeek / DaysOfMonth
// 用 -join [char]9 而不是反引号转义，Go 的 raw string 才能原样传给 PowerShell。
const listTasksScript = `
$ErrorActionPreference = 'SilentlyContinue'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
$p = $env:ZINC_TASK_PREFIX
$tasks = @(Get-ScheduledTask -TaskName 'zinc-*' | Where-Object { $_.TaskName.StartsWith($p, [System.StringComparison]::OrdinalIgnoreCase) })
foreach ($t in $tasks) {
    $info = Get-ScheduledTaskInfo -TaskName $t.TaskName
    $tr = @($t.Triggers)
    $cls = ''
    $sb = ''
    $dw = ''
    $dm = ''
    if ($tr.Count -gt 0) {
        $x = $tr[0]
        $cls = [string]$x.CimClass.CimClassName
        $sb = [string]$x.StartBoundary
        if ($x.DaysOfWeek) { $dw = [string]$x.DaysOfWeek }
        if ($x.DaysOfMonth) { $dm = ($x.DaysOfMonth -join ',') }
    }
    $next = ''
    if ($info -and $info.NextRunTime) { $next = $info.NextRunTime.ToString('yyyy-MM-dd HH:mm') }
    Write-Output ((@($t.TaskName, [string]$t.State, $next, $cls, $sb, $dw, $dm)) -join [char]9)
}
`

// parseTaskLines 解析 listTasksScript 的 TSV 输出。字段不足的行直接跳过。
func parseTaskLines(out string) []schedRow {
	var rows []schedRow
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(line, "zinc-") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 7 {
			continue
		}
		rows = append(rows, schedRow{
			name:  f[0],
			state: f[1],
			next:  f[2],
			cls:   f[3],
			start: f[4],
			week:  f[5],
			month: f[6],
		})
	}
	return rows
}

// runListSchedules 打印当前所有 zinc 计划任务。
func runListSchedules() error {
	out, err := runPowerShell(listTasksScript, "zinc-")
	if err != nil {
		return fmt.Errorf(M.QueryFailed, err, string(out))
	}
	renderSchedules(os.Stdout, parseTaskLines(string(out)))
	return nil
}

// renderSchedules 把任务清单写到 w；抽出来是为了能脱离 PowerShell 断言格式。
func renderSchedules(w io.Writer, rows []schedRow) {
	if len(rows) == 0 {
		fmt.Fprintln(w, M.NoSchedule)
		fmt.Fprintln(w, M.NoScheduleHint)
		return
	}
	fmt.Fprintf(w, M.ScheduleHeader, len(rows))
	for _, r := range rows {
		fmt.Fprintf(w, M.ScheduleItemName, r.name)
		fmt.Fprintf(w, M.ScheduleItemDetail, schedDesc(r), schedNext(r), schedState(r))
	}
}

// triggerKind 把 CIM 类名归一成 daily/weekly/monthly；不认识的返回空串。
// 用后缀而不是全名匹配：CIM 类名带 MSFT_Task 前缀，但不同 Windows 版本的
// 命名空间/前缀可能不完全一致，末尾的 Daily/Weekly/Monthly 是稳的。
func triggerKind(cls string) string {
	switch {
	case strings.HasSuffix(cls, "DailyTrigger"):
		return "daily"
	case strings.HasSuffix(cls, "WeeklyTrigger"):
		return "weekly"
	case strings.HasSuffix(cls, "MonthlyTrigger"):
		return "monthly"
	}
	return ""
}

// triggerClock 取触发时刻 HH:MM：优先 StartBoundary，退回任务名末尾的 HHMM。
func triggerClock(r schedRow) string {
	if i := strings.IndexByte(r.start, 'T'); i >= 0 && len(r.start) >= i+6 {
		return r.start[i+1 : i+6]
	}
	// 任务名以「-HHMM」结尾；只有这四位是合法时间才算数，
	// 否则 zinc-demo 这种名字会被切出 de:mo。
	if i := strings.LastIndexByte(r.name, '-'); i >= 0 && len(r.name) == i+5 {
		if n, err := strconv.Atoi(r.name[i+1:]); err == nil {
			if hh, mm := n/100, n%100; hh < 24 && mm < 60 {
				return fmt.Sprintf("%02d:%02d", hh, mm)
			}
		}
	}
	return "?"
}

// schedWeekdays 把 MSFT_TaskWeeklyTrigger 的 bitmask 转回 zinc 的 1-7（1=周一）。
func schedWeekdays(mask string) string {
	n, err := strconv.Atoi(strings.TrimSpace(mask))
	if err != nil || n <= 0 {
		return "?"
	}
	var out []string
	for _, p := range []struct{ bit, num int }{
		{2, 1}, {4, 2}, {8, 3}, {16, 4}, {32, 5}, {64, 6}, {1, 7},
	} {
		if n&p.bit != 0 {
			out = append(out, strconv.Itoa(p.num))
		}
	}
	if len(out) == 0 {
		return "?"
	}
	return strings.Join(out, ",")
}

// schedDesc 拼出人类可读的计划，如「每天 03:00」「每周 1,3 03:00」。
func schedDesc(r schedRow) string {
	clock := triggerClock(r)
	kind := triggerKind(r.cls)
	// 只有类名完全没拿到时才按属性推断；拿到了但不认识就原样显示，
	// 免得把用户手建的 TimeTrigger 说成「每天」。
	if kind == "" && r.cls == "" {
		switch {
		case strings.TrimSpace(r.month) != "":
			kind = "monthly"
		case strings.TrimSpace(r.week) != "":
			kind = "weekly"
		case strings.TrimSpace(r.start) != "":
			kind = "daily"
		}
	}
	switch kind {
	case "daily":
		return fmt.Sprintf(M.CronDaily, clock)
	case "weekly":
		return fmt.Sprintf(M.CronWeekly, schedWeekdays(r.week), clock)
	case "monthly":
		days := strings.TrimSpace(r.month)
		if days == "" {
			days = "?"
		}
		return fmt.Sprintf(M.CronMonthly, days, clock)
	}
	if r.cls == "" {
		return "?"
	}
	return r.cls
}

func schedNext(r schedRow) string {
	if strings.TrimSpace(r.next) == "" {
		return M.SchedNextNone
	}
	return r.next
}

func schedState(r schedRow) string {
	switch strings.ToLower(strings.TrimSpace(r.state)) {
	case "ready":
		return M.StateReady
	case "disabled":
		return M.StateDisabled
	case "running":
		return M.StateRunning
	case "queued":
		return M.StateQueued
	case "":
		return M.StateUnknown
	}
	return r.state
}
