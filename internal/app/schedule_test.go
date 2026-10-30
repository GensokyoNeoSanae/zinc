package app

import (
	"bytes"
	"testing"
)

func TestTriggerKind(t *testing.T) {
	cases := map[string]string{
		"MSFT_TaskDailyTrigger":   "daily",
		"MSFT_TaskWeeklyTrigger":  "weekly",
		"MSFT_TaskMonthlyTrigger": "monthly",
		"MSFT_TaskTimeTrigger":    "",
		"":                        "",
		// 只认后缀：换个前缀照样归一
		"Other_DailyTrigger": "daily",
	}
	for in, want := range cases {
		if got := triggerKind(in); got != want {
			t.Errorf("triggerKind(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSchedDescFallback(t *testing.T) {
	saved := M
	M = zhMessages
	defer func() { M = saved }()

	// 类名认不出来时，靠属性推断类型
	r := schedRow{start: "2026-10-28T10:00:00", month: "1,15"}
	if got := schedDesc(r); got != "每月 1,15 10:00" {
		t.Errorf("monthly fallback = %q", got)
	}
	r = schedRow{start: "2026-10-28T20:00:00", week: "10"}
	if got := schedDesc(r); got != "每周 1,3 20:00" {
		t.Errorf("weekly fallback = %q", got)
	}
	r = schedRow{start: "2026-10-28T03:00:00"}
	if got := schedDesc(r); got != "每天 03:00" {
		t.Errorf("daily fallback = %q", got)
	}
}

func TestTriggerClock(t *testing.T) {
	// 优先 StartBoundary
	r := schedRow{name: "zinc-demo-daily-0300", start: "2026-10-28T03:00:00"}
	if got := triggerClock(r); got != "03:00" {
		t.Errorf("from StartBoundary = %q, want 03:00", got)
	}
	// 没有 StartBoundary 时退回任务名末尾的 HHMM
	r = schedRow{name: "zinc-demo-weekly-2130"}
	if got := triggerClock(r); got != "21:30" {
		t.Errorf("from name = %q, want 21:30", got)
	}
	// 两边都拿不到
	r = schedRow{name: "zinc-demo"}
	if got := triggerClock(r); got != "?" {
		t.Errorf("no clock = %q, want ?", got)
	}
	// 含 T 但长度不够
	r = schedRow{name: "zinc-demo-daily-0300", start: "2026-10-28T"}
	if got := triggerClock(r); got != "03:00" {
		t.Errorf("short StartBoundary = %q, want fallback 03:00", got)
	}
}

func TestSchedWeekdays(t *testing.T) {
	cases := map[string]string{
		"2":   "1",   // Monday
		"10":  "1,3", // Monday + Wednesday
		"127": "1,2,3,4,5,6,7",
		"1":   "7", // Sunday
		"":    "?",
		"0":   "?",
		"x":   "?",
	}
	for in, want := range cases {
		if got := schedWeekdays(in); got != want {
			t.Errorf("schedWeekdays(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSchedDesc(t *testing.T) {
	saved := M
	M = zhMessages
	defer func() { M = saved }()

	daily := schedRow{cls: "MSFT_TaskDailyTrigger", start: "2026-10-28T03:00:00"}
	if got := schedDesc(daily); got != "每天 03:00" {
		t.Errorf("daily desc = %q", got)
	}
	weekly := schedRow{cls: "MSFT_TaskWeeklyTrigger", week: "10", start: "2026-10-28T20:00:00"}
	if got := schedDesc(weekly); got != "每周 1,3 20:00" {
		t.Errorf("weekly desc = %q", got)
	}
	monthly := schedRow{cls: "MSFT_TaskMonthlyTrigger", month: "1,15", start: "2026-10-28T10:00:00"}
	if got := schedDesc(monthly); got != "每月 1,15 10:00" {
		t.Errorf("monthly desc = %q", got)
	}
	// 不认识的触发器原样显示类名，至少别丢信息
	other := schedRow{cls: "MSFT_TaskTimeTrigger", start: "2026-10-28T10:00:00"}
	if got := schedDesc(other); got != "MSFT_TaskTimeTrigger" {
		t.Errorf("unknown desc = %q", got)
	}
}

func TestSchedState(t *testing.T) {
	saved := M
	M = zhMessages
	defer func() { M = saved }()

	cases := map[string]string{
		"Ready":    "就绪",
		"Disabled": "已禁用",
		"Running":  "运行中",
		"Queued":   "排队中",
		"":         "未知",
		"Weird":    "Weird",
	}
	for in, want := range cases {
		if got := schedState(schedRow{state: in}); got != want {
			t.Errorf("schedState(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseTaskLines(t *testing.T) {
	out := "zinc-demo-daily-0300\tReady\t2026-10-29 03:00\tMSFT_TaskDailyTrigger\t2026-10-28T03:00:00\t\t\r\n" +
		"zinc-demo-weekly-2000\tDisabled\t\tMSFT_TaskWeeklyTrigger\t2026-10-28T20:00:00\t10\t\t\n" +
		"garbage line\n" +
		"zinc-broken\tonly\tthree\tfields\n"
	rows := parseTaskLines(out)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	if rows[0].name != "zinc-demo-daily-0300" || rows[0].state != "Ready" || rows[0].month != "" {
		t.Errorf("row0 = %+v", rows[0])
	}
	if rows[0].next != "2026-10-29 03:00" || rows[0].week != "" {
		t.Errorf("row0 fields = %+v", rows[0])
	}
	if rows[1].name != "zinc-demo-weekly-2000" || rows[1].week != "10" || rows[1].next != "" {
		t.Errorf("row1 = %+v", rows[1])
	}
}

func TestRenderSchedules(t *testing.T) {
	saved := M
	M = zhMessages
	defer func() { M = saved }()

	var buf bytes.Buffer
	renderSchedules(&buf, nil)
	want := "当前没有 zinc 的计划任务\n用 zinc schedule <仓库> --daily --at 03:00 添加\n"
	if buf.String() != want {
		t.Errorf("empty render = %q, want %q", buf.String(), want)
	}

	buf.Reset()
	renderSchedules(&buf, []schedRow{{
		name:  "zinc-demo-daily-0300",
		state: "Ready",
		next:  "2026-10-29 03:00",
		cls:   "MSFT_TaskDailyTrigger",
		start: "2026-10-28T03:00:00",
	}})
	want = "[计划] 1 个任务\n\n" +
		"  zinc-demo-daily-0300\n" +
		"    每天 03:00 · 下次 2026-10-29 03:00 · 就绪\n"
	if buf.String() != want {
		t.Errorf("render = %q, want %q", buf.String(), want)
	}
}

// 计划任务密码只从配置里取，绝不再交互询问（2026-10-29 的 ndmm 事故）
func TestSchedulePassword(t *testing.T) {
	set := &Settings{DefaultPassword: "DEF"}
	cfg := &Config{Repos: map[string]*Repo{"a": {}, "b": {}}}

	pw, from := schedulePassword(cfg, set, []string{"a", "b"})
	if pw != "DEF" || from != "settings" {
		t.Fatalf("settings 回落 = %q/%q, want DEF/settings", pw, from)
	}

	cfg.Repos["a"] = &Repo{Password: "OWN"}
	pw, from = schedulePassword(cfg, set, []string{"a", "b"})
	if pw != "" || from != "mixed" {
		t.Fatalf("混合密码 = %q/%q, want empty/mixed", pw, from)
	}

	pw, from = schedulePassword(cfg, set, []string{"a"})
	if pw != "OWN" || from != "repo" {
		t.Fatalf("仓库密码 = %q/%q, want OWN/repo", pw, from)
	}

	empty := &Config{Repos: map[string]*Repo{"x": {}}}
	if pw, from := schedulePassword(empty, &Settings{}, []string{"x"}); pw != "" || from != "" {
		t.Fatalf("无密码 = %q/%q, want empty/empty", pw, from)
	}
}

func TestPasswordMismatch(t *testing.T) {
	set := &Settings{DefaultPassword: "DEF"}
	cfg := &Config{Repos: map[string]*Repo{"a": {}, "b": {}}}
	if rn, ok := passwordMismatch(cfg, set, []string{"a", "b"}, "DEF"); ok {
		t.Fatalf("配置密码相同却报冲突: %s", rn)
	}
	if rn, ok := passwordMismatch(cfg, set, []string{"a", "b"}, "OTHER"); !ok || rn != "a" {
		t.Fatalf("密码不同应报冲突, got %q/%v", rn, ok)
	}
	// 配置里根本没密码时，--password 是唯一来源，不算冲突
	if _, ok := passwordMismatch(&Config{Repos: map[string]*Repo{"c": {}}}, &Settings{}, []string{"c"}, "X"); ok {
		t.Fatal("没有配置密码时不该拦截")
	}
}
