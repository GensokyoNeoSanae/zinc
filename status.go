package main

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

		batches, packs, last, from, to := 0, 0, "", "", ""
		var total int64
		if mErr == nil {
			batches = len(m.Batches)
			for _, b := range m.Batches {
				packs += len(b.Files)
				for _, f := range b.Files {
					if st, e := os.Stat(filepath.Join(repo.Output, f)); e == nil {
						total += st.Size()
					}
				}
				if b.From != "" && (from == "" || b.From < from) {
					from = b.From
				}
				if b.To > to {
					to = b.To
				}
			}
			if batches > 0 {
				last = m.Batches[batches-1].Time
			}
		}

		if name == "" {
			fmt.Printf(M.StatusBrief, n, len(idx.Files), batches, packs, last)
			continue
		}

		newCnt, modCnt, skipCnt := 0, 0, 0
		if _, statErr := os.Stat(repo.Source); statErr == nil {
			if skip, serr := scanSkipRoot(repo); serr == nil {
				sc, _ := scanSource(repo.Source, idx, repo.exclude(set), skip)
				newCnt, modCnt, skipCnt = len(sc.New), len(sc.Mod), sc.SkippedN
			}
		}

		fmt.Printf("%s\n", n)
		fmt.Printf(M.StatusSource, repo.Source)
		fmt.Printf(M.StatusOutput, repo.Output)
		fmt.Printf(M.StatusBacked, len(idx.Files))
		fmt.Printf(M.StatusPending, newCnt, modCnt, skipCnt)
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
