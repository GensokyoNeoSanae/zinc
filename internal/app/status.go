package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// runStatus 输出仓库状态：不带 name 出简表，带 name 出详情
func runStatus(name string, cfg *Config, set *Settings) error {
	names := cfg.sortedNames()
	if name != "" {
		names = []string{name}
	}
	if len(names) == 0 {
		fmt.Println(M.NoRepos)
		return nil
	}

	var firstErr error
	for _, n := range names {
		repo, err := cfg.repo(n)
		if err != nil {
			fmt.Printf("%s: %v\n", n, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		idx, idxErr := loadIndex(repo.Output)
		if idxErr != nil {
			fmt.Printf("%s: %v\n", n, idxErr)
			if firstErr == nil {
				firstErr = idxErr
			}
			continue
		}
		m, mErr := loadManifest(repo.Output)

		batches, packs, last := 0, 0, ""
		var total int64
		if mErr == nil {
			// 包名去重：旧账本里 compact 会把同一个包再登记一次，别算成两个包。
			seen := make(map[string]struct{}, len(m.Batches))
			for _, b := range m.Batches {
				batches++
				for f := range b.Packs {
					if _, ok := seen[f]; ok {
						continue
					}
					seen[f] = struct{}{}
					if st, e := os.Stat(filepath.Join(repo.Output, f)); e == nil {
						total += st.Size()
					}
				}
				if b.Time > last {
					last = b.Time
				}
			}
			packs = len(seen)
		}

		// 内容日期范围从索引的 P 行取（manifest 只存哈希，不再存 from/to）
		from, to := "", ""
		for _, p := range idx.Packs {
			if p.From != "" && (from == "" || p.From < from) {
				from = p.From
			}
			if p.To > to {
				to = p.To
			}
		}

		if name == "" {
			fmt.Printf(M.StatusBrief, n, len(idx.Files), batches, packs, last)
			continue
		}

		newCnt, modCnt, delCnt, skipCnt := 0, 0, 0, 0
		if _, statErr := os.Stat(repo.Source); statErr == nil {
			if skip, serr := scanSkipRoot(repo); serr == nil {
				sc, _ := scanSource(repo.Source, idx, repo.exclude(set), skip)
				newCnt, modCnt, delCnt, skipCnt = len(sc.New), len(sc.Mod), len(sc.Del), sc.SkippedN
			}
		}

		fmt.Printf("%s\n", n)
		fmt.Printf(M.StatusSource, repo.Source)
		fmt.Printf(M.StatusOutput, repo.Output)
		fmt.Printf(M.StatusBacked, len(idx.Files))
		fmt.Printf(M.StatusPending, newCnt, modCnt, delCnt, skipCnt)
		fmt.Printf(M.StatusPacks, batches, packs, humanSize(total))
		if from == "" && to == "" {
			from, to = "-", "-"
		}
		fmt.Printf(M.StatusContent, from, to)
		fmt.Printf(M.StatusLast, last)
		if mErr == nil {
			fmt.Print(M.StatusManifestExists)
		} else {
			fmt.Print(M.StatusManifestMissing)
		}
	}
	return firstErr
}
