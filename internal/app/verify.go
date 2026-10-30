package app

import (
	"encoding/json"
	"errors"
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

// runVerify 对本地包做 SHA-256 校验，并用配置里的密码试开每个包。
// samplePct < 100 时按包随机抽样。
func runVerify(name string, cfg *Config, set *Settings, samplePct int) error {
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

	// 密码检查：7z l 只读加密头，能打开就说明这些包用的就是配置里的密码。
	// 只算 SHA-256 是查不出「包完好、但当初用了别的密码」的。
	// （历史事故：计划任务里固化了一个测试密码，包一直「校验通过」却拿真密码解不开。）
	sz := repo.sevenZip(set)
	pw := repo.password(set)
	if pw != "" && sz == "" {
		fmt.Print(M.VerifyPwNo7zip)
	}
	checkPw := pw != "" && sz != ""

	type packCheck struct {
		name string
		hash string
	}
	// 删除记录也要校验：它坏了，恢复时不是少删就是删错
	rmOK, rmBad, rmMissing := 0, 0, 0
	// 同一个包可能出现两次：备份批次登记一次，compact 重写后再登记一次。
	// 以最后一条为准——账本是追加写的，后写的才是这个包现在的内容。
	var order []string
	latest := make(map[string]string)
	checked := make(map[string]struct{})
	for _, b := range m.Batches {
		if b.Rm != "" {
			// 同一条删除记录也可能被后面的批次重提一次，别重复计数
			if _, done := checked[b.Rm]; !done {
				checked[b.Rm] = struct{}{}
				h, herr := sha256File(filepath.Join(repo.Output, b.Rm))
				switch {
				case os.IsNotExist(herr):
					rmMissing++
					fmt.Printf(M.VerifyRmMissing, b.Rm)
				case herr != nil:
					rmBad++
					fmt.Printf(M.VerifyRmCorrupt, b.Rm)
				case b.RmHash == "":
					rmOK++ // 旧账本没记哈希，只能确认文件还在
				case h != b.RmHash:
					rmBad++
					fmt.Printf(M.VerifyRmCorrupt, b.Rm)
				default:
					rmOK++
				}
			}
		}
		for i, f := range b.Files {
			if i >= len(b.Hashes) {
				continue
			}
			if _, seen := latest[f]; !seen {
				order = append(order, f)
			}
			latest[f] = b.Hashes[i]
		}
	}
	all := make([]packCheck, 0, len(order))
	for _, f := range order {
		all = append(all, packCheck{name: f, hash: latest[f]})
	}

	if samplePct > 0 && samplePct < 100 && len(all) > 0 {
		n := len(all) * samplePct / 100
		if n < 1 {
			n = 1
		}
		rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
		all = all[:n]
	}

	ok, bad, missing, pwBad := 0, 0, 0, 0
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
		if checkPw {
			if _, lerr := listPackContents(sz, full, pw); lerr != nil {
				pwBad++
				fmt.Printf(M.VerifyPwBad, p.name)
			}
		}
		ok++
	}

	if rmOK+rmBad+rmMissing > 0 {
		fmt.Printf(M.VerifyRmSummary, rmOK, rmBad, rmMissing)
	}
	if pwBad > 0 {
		fmt.Printf(M.VerifyPwSummary, pwBad)
	}
	fmt.Printf(M.VerifySummary, name, ok, bad, missing)
	logf("verify %s: ok=%d bad=%d missing=%d pw_bad=%d rm_ok=%d rm_bad=%d rm_missing=%d",
		name, ok, bad, missing, pwBad, rmOK, rmBad, rmMissing)
	if bad > 0 || missing > 0 || pwBad > 0 || rmBad > 0 || rmMissing > 0 {
		return errors.New(M.VerifyFailedMsg)
	}
	return nil
}
