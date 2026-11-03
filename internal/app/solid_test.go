package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseSolid(t *testing.T) {
	ok := []struct {
		in   string
		want int64
	}{
		{"off", 0},
		{"OFF", 0},
		{"none", 0},
		{"0", 0},
		{"64M", 64 << 20},
		{"64m", 64 << 20},
		{"4G", 4 << 30},
		{"1024", 1024},
		{" 64M ", 64 << 20},
	}
	for _, c := range ok {
		got, err := parseSolid(c.in)
		if err != nil {
			t.Errorf("parseSolid(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseSolid(%q) = %d, want %d", c.in, got, c.want)
		}
	}

	bad := []string{"", "abc", "-1", "0M", "offf"}
	for _, in := range bad {
		if _, err := parseSolid(in); err == nil {
			t.Errorf("parseSolid(%q) should fail", in)
		}
	}
}

func TestSolidArg(t *testing.T) {
	off := solidArg(0)
	if off != "-ms=off" {
		t.Errorf("solidArg(0) = %q, want -ms=off", off)
	}
	if got := solidArg(-1); got != "-ms=off" {
		t.Errorf("solidArg(-1) = %q, want -ms=off", got)
	}
	if got := solidArg(64 << 20); got != "-ms=64m" {
		t.Errorf("solidArg(64M) = %q, want -ms=64m", got)
	}
	if got := solidArg(1 << 20); got != "-ms=1m" {
		t.Errorf("solidArg(1M) = %q, want -ms=1m", got)
	}
	// 非整 MB 向上取整：7-Zip 的 -ms 不带单位会报参数错误。
	if got := solidArg((1 << 20) + 1); got != "-ms=2m" {
		t.Errorf("solidArg(1M+1) = %q, want -ms=2m", got)
	}
}

func TestSolidFallback(t *testing.T) {
	s := defaultSettings()
	if got := s.solidBlock(); got != defaultSolidBlock {
		t.Errorf("default solidBlock() = %d, want %d", got, defaultSolidBlock)
	}

	zero := int64(0)
	s.SolidBlock = &zero
	if got := s.solidBlock(); got != 0 {
		t.Errorf("explicit off solidBlock() = %d, want 0", got)
	}

	// 仓库字段优先于 settings.json；仓库没设才回落到 settings。
	r := &Repo{}
	if got := r.solidBlock(s); got != 0 {
		t.Errorf("repo inheriting settings = %d, want 0", got)
	}
	v := int64(32 << 20)
	r.SolidBlock = &v
	if got := r.solidBlock(s); got != v {
		t.Errorf("repo override = %d, want %d", got, v)
	}
}

func TestPackArgsSolid(t *testing.T) {
	set := defaultSettings()
	repo := &Repo{}

	// 压缩等级 0（默认）一律非固实：Copy 模式下 7-Zip 本来就不建固实块，
	// 再传 -ms 只是改容器结构，实测对体积和损坏恢复都没有影响。
	args := packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-ms=off") {
		t.Errorf("level 0 should force -ms=off: %v", args)
	}

	// 就算显式配了固实块，等级 0 也不该把它传下去。
	big := int64(64 << 20)
	repo.SolidBlock = &big
	args = packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-ms=off") {
		t.Errorf("level 0 with solid_block set should still be off: %v", args)
	}
	repo.SolidBlock = nil

	// 开了压缩才按配置走，默认 64M。
	five := 5
	repo.CompressLevel = &five
	args = packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-ms=64m") {
		t.Errorf("level 5 should use the default 64M blocks: %v", args)
	}

	off := int64(0)
	repo.SolidBlock = &off
	args = packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-ms=off") {
		t.Errorf("level 5 with solid off: %v", args)
	}
}

func TestPackArgsHeaderCrypt(t *testing.T) {
	set := defaultSettings()
	repo := &Repo{}

	// 默认加密头部。
	args := packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-mhe=on") {
		t.Errorf("default pack args missing -mhe=on: %v", args)
	}

	// 仓库级关掉优先于默认。
	f := false
	repo.HeaderCrypt = &f
	args = packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-mhe=off") {
		t.Errorf("repo header off should yield -mhe=off: %v", args)
	}

	// 仓库没设时回落到 settings。
	repo.HeaderCrypt = nil
	set.HeaderCrypt = &f
	args = packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-mhe=off") {
		t.Errorf("settings header off should yield -mhe=off: %v", args)
	}
}

func TestCfgHeaderCrypt(t *testing.T) {
	// 旧版 settings.json / repos.json 里没有 header_encryption 字段，
	// 反序列化必须是 nil 并回落到「开」，否则老用户升级后包会突然变成明文头。
	var oldSet Settings
	if err := json.Unmarshal([]byte(`{"version":1,"compression_level":0}`), &oldSet); err != nil {
		t.Fatalf("unmarshal old settings: %v", err)
	}
	if oldSet.HeaderCrypt != nil {
		t.Fatal("old settings must leave HeaderCrypt nil")
	}
	if !oldSet.headerCrypt() {
		t.Fatal("old settings without the key must default to header crypt on")
	}
	var oldRepo Repo
	if err := json.Unmarshal([]byte(`{"source":"x","output":"y"}`), &oldRepo); err != nil {
		t.Fatalf("unmarshal old repo: %v", err)
	}
	if !oldRepo.headerCrypt(&oldSet) {
		t.Fatal("old repo without the key must inherit header crypt on")
	}

	s := defaultSettings()
	if !s.headerCrypt() {
		t.Fatal("header crypt should default to true")
	}
	if err := cfgSet(s, "header_encryption", "off"); err != nil {
		t.Fatalf("cfgSet off: %v", err)
	}
	if s.headerCrypt() {
		t.Fatal("header crypt should be off after set")
	}
	if got := cfgValueString(s, "header_encryption"); got != "false" {
		t.Fatalf("cfgValueString = %q, want false", got)
	}
	if err := cfgSet(s, "header_encryption", "on"); err != nil {
		t.Fatalf("cfgSet on: %v", err)
	}
	if !s.headerCrypt() {
		t.Fatal("header crypt should be on after set on")
	}
	if err := cfgSet(s, "header_encryption", "wat"); err == nil {
		t.Fatal("cfgSet wat should fail")
	}
	if err := cfgUnset(s, "header_encryption"); err != nil {
		t.Fatalf("cfgUnset: %v", err)
	}
	if s.HeaderCrypt != nil {
		t.Fatal("after unset the field should be nil")
	}
}

func TestCfgSolid(t *testing.T) {
	s := defaultSettings()
	if err := cfgSet(s, "solid_block", "off"); err != nil {
		t.Fatalf("cfgSet off: %v", err)
	}
	if s.SolidBlock == nil || *s.SolidBlock != 0 {
		t.Fatalf("after set off: %v", s.SolidBlock)
	}
	if err := cfgSet(s, "solid_block", "32M"); err != nil {
		t.Fatalf("cfgSet 32M: %v", err)
	}
	if s.SolidBlock == nil || *s.SolidBlock != 32<<20 {
		t.Fatalf("after set 32M: %v", s.SolidBlock)
	}
	if err := cfgSet(s, "solid_block", "wat"); err == nil {
		t.Fatal("cfgSet wat should fail")
	}
	if err := cfgUnset(s, "solid_block"); err != nil {
		t.Fatalf("cfgUnset: %v", err)
	}
	if s.SolidBlock != nil {
		t.Fatalf("after unset: %v", s.SolidBlock)
	}
}

// hasArg 判断参数列表里是否有某一项（避免 strings.Contains 把前缀误判）。
func hasArg(args []string, want string) bool {
	for _, a := range args {
		if strings.EqualFold(a, want) {
			return true
		}
	}
	return false
}
