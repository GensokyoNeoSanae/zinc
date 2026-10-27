package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

func loadManifest(output string) (*Manifest, error) {
	p := manifestPath(output)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// runVerify 对本地包做 SHA-256 校验。samplePct < 100 时按包随机抽样。
func runVerify(name string, cfg *Config, samplePct int) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	m, err := loadManifest(repo.Output)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(M.NoManifest, manifestPath(repo.Output))
		}
		return fmt.Errorf(M.ManifestBroken, manifestPath(repo.Output), err)
	}

	type packCheck struct {
		name string
		hash string
	}
	var all []packCheck
	for _, b := range m.Batches {
		// 删除记录不参与哈希校验，但缺了就意味着恢复时不会删掉那些文件
		if b.Rm != "" && !fileExists(filepath.Join(repo.Output, b.Rm)) {
			fmt.Printf(M.VerifyRmMissing, b.Rm)
		}
		for i, f := range b.Files {
			if i >= len(b.Hashes) {
				continue
			}
			all = append(all, packCheck{name: f, hash: b.Hashes[i]})
		}
	}

	if samplePct > 0 && samplePct < 100 && len(all) > 0 {
		n := len(all) * samplePct / 100
		if n < 1 {
			n = 1
		}
		rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
		all = all[:n]
	}

	ok, bad, missing := 0, 0, 0
	for _, p := range all {
		full := filepath.Join(repo.Output, p.name)
		if _, statErr := os.Stat(full); statErr != nil {
			missing++
			fmt.Printf(M.MissingFile, p.name)
			continue
		}
		h, hashErr := sha256File(full)
		if hashErr != nil || h != p.hash {
			bad++
			fmt.Printf(M.CorruptFile, p.name)
			continue
		}
		ok++
	}

	fmt.Printf(M.VerifySummary, name, ok, bad, missing)
	logf("verify %s: ok=%d bad=%d missing=%d", name, ok, bad, missing)
	if bad > 0 || missing > 0 {
		return fmt.Errorf(M.VerifyFailedMsg, name)
	}
	return nil
}
