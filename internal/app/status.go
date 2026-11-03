package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// repoStatus 是简表一行需要的全部数据
type repoStatus struct {
	name    string
	files   int
	batches int
	packs   int
	srcSize int64
	outSize int64
	last    string
}

// dirSize 累加目录下所有普通文件的大小。
// 目录项里本身带着 size，不需要打开文件；本机实测 33 GB / 6 万文件约 190 ms。
func dirSize(root string) int64 {
	var total int64
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if info, e := d.Info(); e == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// displayWidth 算终端里的显示宽度：CJK 与全角字符占两列。
// 不这么算，中文仓库名永远对不齐。
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x1100 && (r <= 0x115F ||
			(r >= 0x2E80 && r <= 0xA4CF) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE30 && r <= 0xFE6F) ||
			(r >= 0xFF00 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6)) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func padRight(s string, w int) string {
	if d := w - displayWidth(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

func padLeft(s string, w int) string {
	if d := w - displayWidth(s); d > 0 {
		return strings.Repeat(" ", d) + s
	}
	return s
}

// commaInt 给数字加千分位，方便一眼看出量级
func commaInt(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		b.WriteString(",")
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteString(",")
		}
	}
	return b.String()
}

// sizeDelta 把「输出 - 源」写成 +6.7%, 2.0 GB / -3.2%, 0.8 GB / 与源相当。
// 百分比固定以源为分母。
func sizeDelta(src, out int64) string {
	if src <= 0 {
		return M.SizeUnknown
	}
	diff := out - src
	pct := float64(diff) / float64(src) * 100
	switch {
	case diff == 0 || (pct > -0.1 && pct < 0.1):
		return M.SizeSame
	case diff > 0:
		return fmt.Sprintf("+%.1f%%, %s", pct, humanSize(diff))
	default:
		return fmt.Sprintf("-%.1f%%, %s", -pct, humanSize(-diff))
	}
}

// collectStatus 读索引/账本，并量出源目录与输出目录的实际大小
func collectStatus(name string, cfg *Config) (repoStatus, error) {
	repo, err := cfg.repo(name)
	if err != nil {
		return repoStatus{}, err
	}
	st := repoStatus{name: name}
	if idx, ierr := loadIndex(repo.Output); ierr == nil {
		st.files = len(idx.Files)
	}
	if m, merr := loadManifest(repo.Output); merr == nil {
		// 包名去重：旧账本里 compact 会把同一个包再登记一次
		seen := make(map[string]struct{}, len(m.Batches))
		for _, b := range m.Batches {
			st.batches++
			for f := range b.Packs {
				seen[f] = struct{}{}
			}
			if b.Time > st.last {
				st.last = b.Time
			}
		}
		st.packs = len(seen)
	}
	st.srcSize = dirSize(repo.Source)
	st.outSize = dirSize(repo.Output)
	return st, nil
}

// sortStatus 按字段排序，内部一律升序，desc 时整体反转。
// 各字段自己定方向会让「默认方向」和「--asc/--desc」两套逻辑打架——之前
// --sort out 就是在这里反了两次，结果升序。
func sortStatus(list []repoStatus, by string, desc bool) {
	switch by {
	case "src":
		sort.SliceStable(list, func(i, j int) bool { return list[i].srcSize < list[j].srcSize })
	case "out":
		sort.SliceStable(list, func(i, j int) bool { return list[i].outSize < list[j].outSize })
	case "backup":
		sort.SliceStable(list, func(i, j int) bool { return list[i].last < list[j].last })
	default: // name
		// 中文用 collation（拼音）而不是字节序：用户看到「其余备份 图片 学校课件」
		// 不会觉得那是"按名称"。
		c := collate.New(language.Chinese)
		sort.SliceStable(list, func(i, j int) bool {
			return c.CompareString(list[i].name, list[j].name) < 0
		})
	}
	if desc {
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}
}

// printStatusTable 按显示宽度对齐每一列再输出
func printStatusTable(list []repoStatus) {
	wName, wFiles, wBatches, wPacks, wSize := 0, 0, 0, 0, 0
	for _, st := range list {
		if w := displayWidth(st.name); w > wName {
			wName = w
		}
		if w := len(commaInt(st.files)); w > wFiles {
			wFiles = w
		}
		if w := len(strconv.Itoa(st.batches)); w > wBatches {
			wBatches = w
		}
		if w := len(strconv.Itoa(st.packs)); w > wPacks {
			wPacks = w
		}
		for _, s := range []string{humanSize(st.srcSize), humanSize(st.outSize)} {
			if w := displayWidth(s); w > wSize {
				wSize = w
			}
		}
	}
	for _, st := range list {
		fmt.Printf("%s %s 文件 · %s 批次 / %s 包 · 源 %s，输出 %s(%s) · 最后备份 %s\n",
			padRight(st.name, wName),
			padLeft(commaInt(st.files), wFiles),
			padLeft(strconv.Itoa(st.batches), wBatches),
			padLeft(strconv.Itoa(st.packs), wPacks),
			padLeft(humanSize(st.srcSize), wSize),
			padLeft(humanSize(st.outSize), wSize),
			sizeDelta(st.srcSize, st.outSize),
			st.last)
	}
}

// runStatus 输出仓库状态：不带 name 出简表，带 name 出详情
func runStatus(name string, cfg *Config, set *Settings, sortBy string, desc bool) error {
	names := cfg.sortedNames()
	if name != "" {
		names = []string{name}
	}
	if len(names) == 0 {
		fmt.Println(M.NoRepos)
		return nil
	}

	var list []repoStatus
	var firstErr error
	for _, n := range names {
		st, err := collectStatus(n, cfg)
		if err != nil {
			fmt.Printf("%s: %v\n", n, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		list = append(list, st)
	}

	// 详情模式：一次只看一个仓库，排序无意义
	if name != "" && len(list) == 1 {
		return printStatusDetail(name, cfg, set, list[0])
	}
	if len(list) > 0 {
		sortStatus(list, sortBy, desc)
		printStatusTable(list)
	}
	return firstErr
}

// printStatusDetail 是 `zinc status <仓库名>` 的详细视图
func printStatusDetail(name string, cfg *Config, set *Settings, st repoStatus) error {
	repo, err := cfg.repo(name)
	if err != nil {
		return err
	}
	idx, idxErr := loadIndex(repo.Output)
	if idxErr != nil {
		return idxErr
	}

	newCnt, modCnt, delCnt, skipCnt := 0, 0, 0, 0
	if _, statErr := os.Stat(repo.Source); statErr == nil {
		if skip, serr := scanSkipRoot(repo); serr == nil {
			sc, _ := scanSource(repo.Source, idx, repo.exclude(set), skip)
			newCnt, modCnt, delCnt, skipCnt = len(sc.New), len(sc.Mod), len(sc.Del), sc.SkippedN
		}
	}

	from, to := "", ""
	for _, p := range idx.Packs {
		if p.From != "" && (from == "" || p.From < from) {
			from = p.From
		}
		if p.To > to {
			to = p.To
		}
	}

	fmt.Printf("%s\n", name)
	fmt.Printf(M.StatusSource, repo.Source)
	fmt.Printf(M.StatusOutput, repo.Output)
	fmt.Printf(M.StatusBacked, len(idx.Files))
	fmt.Printf(M.StatusPending, newCnt, modCnt, delCnt, skipCnt)
	fmt.Printf(M.StatusPacks, st.batches, st.packs, st.outSize)
	fmt.Printf(M.StatusSizes, humanSize(st.srcSize), humanSize(st.outSize), sizeDelta(st.srcSize, st.outSize))
	if from == "" && to == "" {
		from, to = "-", "-"
	}
	fmt.Printf(M.StatusContent, from, to)
	fmt.Printf(M.StatusLast, st.last)
	if _, merr := loadManifest(repo.Output); merr == nil {
		fmt.Print(M.StatusManifestExists)
	} else {
		fmt.Print(M.StatusManifestMissing)
	}
	return nil
}
