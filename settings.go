package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Settings 是 <appDir>\settings.json 的内容：全局默认值。
// 优先级：命令行 > 仓库字段 > settings.json > 内置默认。
type Settings struct {
	Version            int      `json:"version"`
	DefaultPassword    string   `json:"default_password,omitempty"`
	RestoreOverwrite   *bool    `json:"restore_overwrite,omitempty"`
	RestoreWrapFolder  *bool    `json:"restore_wrap_folder,omitempty"`
	MaxPackByte        int64    `json:"max_pack_byte,omitempty"`
	CompressionLevel   *int     `json:"compression_level,omitempty"`
	SevenZipPath       string   `json:"7zip_path,omitempty"`
	Lang               string   `json:"lang,omitempty"`
	LockTimeoutMinutes int      `json:"lock_timeout_minutes,omitempty"`
	Exclude            []string `json:"exclude,omitempty"`
	CompleteNotify     *bool    `json:"complete_notify,omitempty"`
}

func settingsPath() string { return filepath.Join(appDir(), "settings.json") }

// defaultSettings 返回内置默认值（与文件缺失时等价）
func defaultSettings() *Settings {
	return &Settings{Version: 1}
}

func loadSettings() (*Settings, error) {
	p := settingsPath()
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return defaultSettings(), nil
	}
	if err != nil {
		return nil, fmt.Errorf(M.LoadSettingsFailed, p, err)
	}
	var s Settings
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf(M.ParseSettingsFailed, p, err)
	}
	if s.Version == 0 {
		s.Version = 1
	}
	return &s, nil
}

// loadSettingsQuiet 用于 initLang 阶段：失败就退回默认值，不打断启动
func loadSettingsQuiet() *Settings {
	s, err := loadSettings()
	if err != nil || s == nil {
		return defaultSettings()
	}
	return s
}

func saveSettings(s *Settings) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	p := settingsPath()
	tmp := p + ".tmp"
	if err := writeFileSync(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// ---------- 取值回落 ----------

func (s *Settings) defaultPassword() string { return s.DefaultPassword }

func (s *Settings) restoreOverwrite() bool {
	if s.RestoreOverwrite == nil {
		return true
	}
	return *s.RestoreOverwrite
}

// restoreWrapFolder 恢复时是否在外层包一个 restore-<仓库>-<时间> 文件夹（默认开）
func (s *Settings) restoreWrapFolder() bool {
	if s.RestoreWrapFolder == nil {
		return true
	}
	return *s.RestoreWrapFolder
}

func (s *Settings) maxPackByte() int64 {
	if s.MaxPackByte <= 0 {
		return defaultMaxPackByte
	}
	return s.MaxPackByte
}

func (s *Settings) compressLevel() int {
	if s.CompressionLevel == nil {
		return defaultCompressLvl
	}
	return *s.CompressionLevel
}

func (s *Settings) sevenZipPath() string {
	if s.SevenZipPath == "" {
		return defaultSevenZip
	}
	return s.SevenZipPath
}

func (s *Settings) lang() string { return s.Lang }

func (s *Settings) lockTimeout() time.Duration {
	n := s.LockTimeoutMinutes
	if n <= 0 {
		n = 30
	}
	return time.Duration(n) * time.Minute
}

func (s *Settings) exclude() []string { return s.Exclude }

func (s *Settings) completeNotify() bool {
	if s.CompleteNotify == nil {
		return true
	}
	return *s.CompleteNotify
}

// ---------- zinc config ----------

type cfgKind int

const (
	kString cfgKind = iota
	kBool
	kSize
	kInt
	kList
)

type cfgKey struct {
	name string
	kind cfgKind
	help string
}

var cfgKeys = []cfgKey{
	{"default_password", kString, "备份/恢复默认密码，空=每次询问"},
	{"restore_overwrite", kBool, "恢复时同名文件是否覆盖（默认 true）"},
	{"restore_wrap_folder", kBool, "恢复时在外层包一个 restore-<仓库>-<时间> 文件夹（默认 true）"},
	{"max_pack_byte", kSize, "单包上限，支持 4G / 50M / 纯字节"},
	{"compression_level", kInt, "7z 压缩等级 0-9（0=仅存储）"},
	{"7zip_path", kString, "7za.exe 路径"},
	{"lang", kString, "界面语言 zh / en，空=自动"},
	{"lock_timeout_minutes", kInt, "仓库锁最长等待分钟数"},
	{"exclude", kList, "排除名单，glob，可多次 set 追加"},
	{"complete_notify", kBool, "备份完成发 Windows 通知"},
}

func findCfgKey(name string) (cfgKey, bool) {
	for _, k := range cfgKeys {
		if k.name == name {
			return k, true
		}
	}
	return cfgKey{}, false
}

func maskSecret(s string) string {
	if s == "" {
		return M.ConfigEmpty
	}
	r := []rune(s)
	if len(r) <= 4 {
		return "****"
	}
	return string(r[:2]) + "***" + string(r[len(r)-2:])
}

// cfgValueString 返回某个键当前值的可读文本（密码已打码）
func cfgValueString(s *Settings, key string) string {
	switch key {
	case "default_password":
		return maskSecret(s.DefaultPassword)
	case "restore_overwrite":
		return strconv.FormatBool(s.restoreOverwrite())
	case "restore_wrap_folder":
		return strconv.FormatBool(s.restoreWrapFolder())
	case "max_pack_byte":
		return fmt.Sprintf("%d (%s)", s.maxPackByte(), humanSize(s.maxPackByte()))
	case "compression_level":
		return strconv.Itoa(s.compressLevel())
	case "7zip_path":
		return s.sevenZipPath()
	case "lang":
		if s.Lang == "" {
			return M.ConfigAuto
		}
		return s.Lang
	case "lock_timeout_minutes":
		return strconv.Itoa(int(s.lockTimeout() / time.Minute))
	case "exclude":
		if len(s.Exclude) == 0 {
			return M.ConfigEmpty
		}
		return strings.Join(s.Exclude, ", ")
	case "complete_notify":
		return strconv.FormatBool(s.completeNotify())
	}
	return ""
}

// cfgValueRaw 返回适合脚本消费的裸值（不带人类可读后缀）
func cfgValueRaw(s *Settings, key string) string {
	switch key {
	case "max_pack_byte":
		return strconv.FormatInt(s.maxPackByte(), 10)
	case "lock_timeout_minutes":
		return strconv.Itoa(int(s.lockTimeout() / time.Minute))
	}
	return cfgValueString(s, key)
}

// cfgIsSet 判断该键在文件里是否有显式值（用于 unset 提示）
func cfgIsSet(s *Settings, key string) bool {
	switch key {
	case "default_password":
		return s.DefaultPassword != ""
	case "restore_overwrite":
		return s.RestoreOverwrite != nil
	case "restore_wrap_folder":
		return s.RestoreWrapFolder != nil
	case "max_pack_byte":
		return s.MaxPackByte != 0
	case "compression_level":
		return s.CompressionLevel != nil
	case "7zip_path":
		return s.SevenZipPath != ""
	case "lang":
		return s.Lang != ""
	case "lock_timeout_minutes":
		return s.LockTimeoutMinutes != 0
	case "exclude":
		return len(s.Exclude) > 0
	case "complete_notify":
		return s.CompleteNotify != nil
	}
	return false
}

// cfgSet 写入一个键，返回错误描述
func cfgSet(s *Settings, key, val string) error {
	k, ok := findCfgKey(key)
	if !ok {
		return fmt.Errorf(M.ConfigUnknownKey, key)
	}
	switch k.kind {
	case kString:
		switch key {
		case "default_password":
			s.DefaultPassword = val
		case "7zip_path":
			s.SevenZipPath = val
		case "lang":
			if val != "" && !strings.EqualFold(val, "auto") {
				if !setLang(val) {
					return fmt.Errorf(M.ConfigBadValue, key, val)
				}
			}
			s.Lang = val
		}
	case kBool:
		b, err := strconv.ParseBool(strings.ToLower(val))
		if err != nil {
			return fmt.Errorf(M.ConfigBadValue, key, val)
		}
		switch key {
		case "restore_overwrite":
			s.RestoreOverwrite = &b
		case "restore_wrap_folder":
			s.RestoreWrapFolder = &b
		default:
			s.CompleteNotify = &b
		}
	case kSize:
		n, err := parseSize(val)
		if err != nil {
			return fmt.Errorf(M.ConfigBadValue, key, val)
		}
		s.MaxPackByte = n
	case kInt:
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf(M.ConfigBadValue, key, val)
		}
		if key == "compression_level" {
			if n < 0 || n > 9 {
				return fmt.Errorf(M.ConfigBadValue, key, val)
			}
			s.CompressionLevel = &n
		} else {
			if n <= 0 {
				return fmt.Errorf(M.ConfigBadValue, key, val)
			}
			s.LockTimeoutMinutes = n
		}
	case kList:
		if key == "exclude" {
			s.Exclude = append(s.Exclude, val)
		}
	}
	return nil
}

// cfgUnset 清掉一个键（回落到内置默认）
func cfgUnset(s *Settings, key string) error {
	if _, ok := findCfgKey(key); !ok {
		return fmt.Errorf(M.ConfigUnknownKey, key)
	}
	switch key {
	case "default_password":
		s.DefaultPassword = ""
	case "restore_overwrite":
		s.RestoreOverwrite = nil
	case "restore_wrap_folder":
		s.RestoreWrapFolder = nil
	case "max_pack_byte":
		s.MaxPackByte = 0
	case "compression_level":
		s.CompressionLevel = nil
	case "7zip_path":
		s.SevenZipPath = ""
	case "lang":
		s.Lang = ""
	case "lock_timeout_minutes":
		s.LockTimeoutMinutes = 0
	case "exclude":
		s.Exclude = nil
	case "complete_notify":
		s.CompleteNotify = nil
	}
	return nil
}

// parseSize 解析 4G / 50M / 1024 这样的容量写法（1024 进制）
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	mult := int64(1)
	switch s[len(s)-1] {
	case 'k', 'K':
		mult = 1 << 10
		s = s[:len(s)-1]
	case 'm', 'M':
		mult = 1 << 20
		s = s[:len(s)-1]
	case 'g', 'G':
		mult = 1 << 30
		s = s[:len(s)-1]
	case 't', 'T':
		mult = 1 << 40
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	if n > (1<<63-1)/mult {
		return 0, fmt.Errorf("too large")
	}
	return n * mult, nil
}
