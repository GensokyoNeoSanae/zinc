package app

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
)

func loadManifest(output string) (*Manifest, error) {
	b, err := os.ReadFile(manifestPath(output))
	if err != nil {
		return nil, err
	}
	return decodeManifest(b)
}

// runVerify 校验本地包：doHash 时比对 SHA-256，doPw 时用配置密码试开每个包。
// samplePct < 100 时按包随机抽样。
func runVerify(name string, cfg *Config, set *Settings, samplePct int, doHash, doPw bool) error {
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
	if doPw && sz == "" {
		// 显式只要密码检查时，缺 7-Zip 不能假装通过
		if !doHash {
			return errors.New(M.SevenZipMissing)
		}
		fmt.Print(M.VerifyPwNo7zip)
		doPw = false
	}
	if doPw && pw == "" {
		// 显式只要密码检查时，没有密码就没得验
		if !doHash {
			return errors.New(M.VerifyNoPassword)
		}
		doPw = false
	}

	type packCheck struct {
		name string
		hash string
	}
	// 删除记录也要校验：它坏了，恢复时不是少删就是删错
	rmOK, rmBad, rmMissing := 0, 0, 0
	// 同一个包可能在多个批次里出现（旧账本里 compact 会再登记一次）：
	// 以最后一条为准，后写的才是这个包现在的内容。
	var order []string
	latest := make(map[string]string)
	seenRm := make(map[string]bool)
	for _, b := range m.Batches {
		if doHash {
			for rm, want := range b.Rms {
				if _, done := seenRm[rm]; done {
					continue
				}
				seenRm[rm] = true
				h, herr := sha256File(filepath.Join(repo.Output, rm))
				switch {
				case os.IsNotExist(herr):
					rmMissing++
					fmt.Printf(M.VerifyRmMissing, rm)
				case herr != nil:
					rmBad++
					fmt.Printf(M.VerifyRmCorrupt, rm)
				case want == "":
					rmOK++ // 旧账本没记哈希，只能确认文件还在
				case h != want:
					rmBad++
					fmt.Printf(M.VerifyRmCorrupt, rm)
				default:
					rmOK++
				}
			}
		}
		for f, h := range b.Packs {
			if _, seen := latest[f]; !seen {
				order = append(order, f)
			}
			latest[f] = h
		}
	}
	// 映射没有顺序，排一下让 missing/corrupt 的输出稳定
	sort.Strings(order)
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
		if doHash {
			h, hashErr := sha256File(full)
			if hashErr != nil || h != p.hash {
				bad++
				fmt.Printf(M.CorruptFile, p.name)
				continue
			}
		}
		pwOK := true
		if doPw {
			if _, lerr := listPackContents(sz, full, pw); lerr != nil {
				pwBad++
				pwOK = false
				fmt.Printf(M.VerifyPwBad, p.name)
			}
		}
		if pwOK {
			ok++
		}
	}

	if doHash && rmOK+rmBad+rmMissing > 0 {
		fmt.Printf(M.VerifyRmSummary, rmOK, rmBad, rmMissing)
	}
	if doPw && pwBad > 0 {
		fmt.Printf(M.VerifyPwSummary, pwBad)
	}
	if doHash {
		fmt.Printf(M.VerifySummary, name, ok, bad, missing)
	} else {
		fmt.Printf(M.VerifyPwOnlySummary, name, ok, pwBad, missing)
	}
	mode := "both"
	if !doHash {
		mode = "password"
	} else if !doPw {
		mode = "hash"
	}
	logf("verify %s: mode=%s ok=%d bad=%d missing=%d pw_bad=%d rm_ok=%d rm_bad=%d rm_missing=%d",
		name, mode, ok, bad, missing, pwBad, rmOK, rmBad, rmMissing)
	if bad > 0 || missing > 0 || pwBad > 0 || rmBad > 0 || rmMissing > 0 {
		return errors.New(M.VerifyFailedMsg)
	}
	return nil
}
