package main

import (
	"fmt"
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
