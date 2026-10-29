package app

import (
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

	// 没配任何东西时用内置默认 64M。
	args := packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-ms=64m") {
		t.Errorf("default pack args missing -ms=64m: %v", args)
	}

	off := int64(0)
	repo.SolidBlock = &off
	args = packArgs(repo, set, "pw", "x.7z.part", ".list-x.txt")
	if !hasArg(args, "-ms=off") {
		t.Errorf("off pack args missing -ms=off: %v", args)
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
