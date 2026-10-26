package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	defaultSevenZip    = `E:\Tools\7zip\7za.exe`
	defaultMaxPackByte = int64(4 * 1024 * 1024 * 1024) // 4 GB
	defaultCompressLvl = 0                             // 存储模式
)

// Repo 是一个备份仓库的配置。字段留空表示回落到 settings.json。
type Repo struct {
	Source        string   `json:"source"`
	Output        string   `json:"output"`
	MaxPackByte   int64    `json:"max_pack_byte,omitempty"`
	CompressLevel *int     `json:"compression_level,omitempty"`
	SevenZipPath  string   `json:"7zip_path,omitempty"`
	Password      string   `json:"password,omitempty"`
	Exclude       []string `json:"exclude,omitempty"`
}

func (r *Repo) maxPackByte(s *Settings) int64 {
	if r.MaxPackByte > 0 {
		return r.MaxPackByte
	}
	return s.maxPackByte()
}

func (r *Repo) compressLevel(s *Settings) int {
	if r.CompressLevel != nil {
		return *r.CompressLevel
	}
	return s.compressLevel()
}

func (r *Repo) sevenZip(s *Settings) string {
	if r.SevenZipPath != "" {
		return r.SevenZipPath
	}
	return s.sevenZipPath()
}

func (r *Repo) password(s *Settings) string {
	if r.Password != "" {
		return r.Password
	}
	return s.defaultPassword()
}

func (r *Repo) exclude(s *Settings) []string {
	if len(r.Exclude) > 0 {
		return r.Exclude
	}
	return s.exclude()
}

// Config 是 repos.json 的整体结构
type Config struct {
	Version int              `json:"version"`
	Root    string           `json:"root"`
	Repos   map[string]*Repo `json:"repos"`
}

func loadConfig() (*Config, error) {
	p := configPath()
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{Version: 2, Repos: map[string]*Repo{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf(M.LoadConfigFailed, p, err)
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf(M.ParseConfigFailed, p, err)
	}
	if c.Repos == nil {
		c.Repos = map[string]*Repo{}
	}
	return &c, nil
}

func saveConfig(c *Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	p := configPath()
	tmp := p + ".tmp"
	if err := writeFileSync(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (c *Config) repo(name string) (*Repo, error) {
	r, ok := c.Repos[name]
	if !ok {
		return nil, fmt.Errorf(M.RepoNotFound, name)
	}
	return r, nil
}

func (c *Config) sortedNames() []string {
	names := make([]string, 0, len(c.Repos))
	for n := range c.Repos {
		names = append(names, n)
	}
	// 简单插入排序，仓库数量很少
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
	return names
}

// outputOwner 返回使用同一输出目录的其它仓库（大小写敏感比较，按用户要求）。
func (c *Config) outputOwner(self, output string) (string, bool) {
	o, err := absClean(output)
	if err != nil {
		return "", false
	}
	for name, r := range c.Repos {
		if name == self {
			continue
		}
		ro, err := absClean(r.Output)
		if err != nil {
			continue
		}
		if ro == o {
			return name, true
		}
	}
	return "", false
}

// outputOwnerFold 大小写不敏感地找“物理上可能是同一个目录”的仓库，用于警告。
func (c *Config) outputOwnerFold(self, output string) (string, bool) {
	o, err := absClean(output)
	if err != nil {
		return "", false
	}
	for name, r := range c.Repos {
		if name == self {
			continue
		}
		ro, err := absClean(r.Output)
		if err != nil {
			continue
		}
		if strings.EqualFold(ro, o) {
			return name, true
		}
	}
	return "", false
}
