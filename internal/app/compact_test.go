package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMinGain(t *testing.T) {
	const gb = int64(1 << 30)

	// 百分比：1 GB 的 5% = 51 MB，高于 8 MiB 下限，按 5% 算
	if got, err := parseMinGain("5%", gb); err != nil || got != gb*5/100 {
		t.Fatalf("5%% of 1G = %d, want %d (err=%v)", got, gb*5/100, err)
	}
	// 小包按百分比算出来远低于 8 MiB，要抬到下限，否则为几 MB 重传整包
	if got, err := parseMinGain("5%", 1<<20); err != nil || got != defaultMinGainAbs {
		t.Fatalf("5%% of 1M = %d, want %d (下限)", got, defaultMinGainAbs)
	}
	// 绝对值
	if got, err := parseMinGain("8M", gb); err != nil || got != 8<<20 {
		t.Fatalf("8M = %d, err=%v", got, err)
	}
	// 空 / 0 = 不设阈值
	for _, s := range []string{"", "0"} {
		if got, err := parseMinGain(s, gb); err != nil || got != 0 {
			t.Fatalf("%q = %d, want 0 (err=%v)", s, got, err)
		}
	}
	// 非法
	for _, s := range []string{"abc", "-5%", "8Q"} {
		if _, err := parseMinGain(s, gb); err == nil {
			t.Fatalf("%q 应该报错", s)
		}
	}
}

func TestCompactTmpNaming(t *testing.T) {
	dir := compactTmpDir(`E:\out`, "20260101_000000_000.7z")
	if want := filepath.Join(`E:\out`, "compact-20260101_000000_000-tmp"); dir != want {
		t.Fatalf("tmp dir = %s, want %s", dir, want)
	}
	if !isCompactTmpName("compact-20260101_000000_000-tmp") {
		t.Fatal("应该认出 compact 工作目录")
	}
	for _, n := range []string{"20260101_000000_000.7z", "compact-x", "compactx-tmp", "xcompact-1-tmp"} {
		if isCompactTmpName(n) {
			t.Fatalf("%q 不该被当成 compact 工作目录", n)
		}
	}
}

// 占位文件只在「同名且 0 字节」时才清，别误删用户真正的同名文件
func TestRemovePlaceholder(t *testing.T) {
	dir := t.TempDir()
	if n, err := removePlaceholder(dir); err != nil || n != 0 {
		t.Fatalf("空目录 = %d, %v", n, err)
	}
	p := filepath.Join(dir, placeholderName)
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := removePlaceholder(dir); err != nil || n != 1 {
		t.Fatalf("0 字节占位 = %d, %v, want 1", n, err)
	}
	if fileExists(p) {
		t.Fatal("占位文件应该被删掉")
	}
	// 有内容就不动
	if err := os.WriteFile(p, []byte("real data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := removePlaceholder(dir); err != nil || n != 0 {
		t.Fatalf("非空同名文件 = %d, %v, want 0", n, err)
	}
	if !fileExists(p) {
		t.Fatal("非空的同名文件不该被删")
	}
}

func TestCheckIndexCoversPacks(t *testing.T) {
	idx := newIndex()
	idx.addPack("a.7z", "", "")
	idx.addPack("b.7z", "", "")
	if err := checkIndexCoversPacks(idx, []string{"a.7z", "b.7z"}, "r", "idx"); err != nil {
		t.Fatalf("一致时报错: %v", err)
	}
	// 目录里多了一个包：索引没登记，说明索引不完整，compact 会把它的内容当死文件
	if err := checkIndexCoversPacks(idx, []string{"a.7z", "b.7z", "c.7z"}, "r", "idx"); err == nil {
		t.Fatal("索引缺包时应该拒绝")
	}
	// 索引里多登记了一个已不存在的包
	idx.addPack("c.7z", "", "")
	if err := checkIndexCoversPacks(idx, []string{"a.7z", "b.7z"}, "r", "idx"); err == nil {
		t.Fatal("索引多包时应该拒绝")
	}
}

// 一个包里：2 个还是当前版本、1 个被新包取代、1 个已从源目录删除。
// 重写后包里只该剩前两个，且 mtime 原样不变。
func TestRewritePackKeepsOnlyCurrent(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	files := writeFiles(t, src, "f", 4, 4096)
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz}
	pack := "20260101_000000_000.7z"
	if _, err := packOne(repo, set, packPlan{Files: files}, pack, "TESTPW", false); err != nil {
		t.Fatalf("packOne: %v", err)
	}

	idx := newIndex()
	id := idx.addPack(pack, "", "")
	id2 := idx.addPack("20260102_000000_000.7z", "", "")
	for i := 0; i < 2; i++ {
		idx.Files[files[i].Rel] = &FileEntry{Size: files[i].Size, MTime: files[i].MTime, Pack: id}
	}
	// f2 已经被新包取代，f3 没有条目（已从源目录删除）
	idx.Files[files[2].Rel] = &FileEntry{Size: files[2].Size, MTime: files[2].MTime, Pack: id2}

	plan, err := buildCompactPlan(repo, set, idx, "TESTPW", pack)
	if err != nil {
		t.Fatalf("buildCompactPlan: %v", err)
	}
	if len(plan.Keep) != 2 || len(plan.Drop) != 2 {
		t.Fatalf("Keep=%d Drop=%d, want 2/2", len(plan.Keep), len(plan.Drop))
	}
	if plan.KeepSz != files[0].Size+files[1].Size {
		t.Fatalf("KeepSz = %d", plan.KeepSz)
	}

	oldSize := plan.PackSz
	res, err := rewritePack(repo, set, plan, "TESTPW", false)
	if err != nil {
		t.Fatalf("rewritePack: %v", err)
	}
	if res.Empty {
		t.Fatal("还有两个文件，不该被当成空包")
	}
	if res.NewSize >= oldSize {
		t.Fatalf("重写后应该更小: %d -> %d", oldSize, res.NewSize)
	}
	if res.From != files[0].MTime.Format("20060102") {
		t.Fatalf("From = %s", res.From)
	}

	entries, err := listPackContents(sz, filepath.Join(out, pack), "TESTPW")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := map[string]packEntry{}
	for _, e := range entries {
		if !e.IsDir {
			got[filepath.ToSlash(e.Path)] = e
		}
	}
	if len(got) != 2 {
		t.Fatalf("重写后包内有 %d 个文件, want 2: %v", len(got), got)
	}
	for i := 0; i < 2; i++ {
		e, ok := got[files[i].Rel]
		if !ok {
			t.Fatalf("%s 丢了", files[i].Rel)
		}
		if e.Size != files[i].Size {
			t.Fatalf("%s 大小 %d, want %d", files[i].Rel, e.Size, files[i].Size)
		}
		// 解出来再压回去，mtime 必须原样：恢复出来的文件时间应该还是源文件的时间
		if !e.MTime.Equal(files[i].MTime) {
			t.Fatalf("%s mtime %v, want %v", files[i].Rel, e.MTime, files[i].MTime)
		}
	}
	// 工作目录不能留在输出目录里
	if fileExists(compactTmpDir(out, pack)) {
		t.Fatal("compact 工作目录应该被清掉")
	}
	// 输出目录里不该留下 .list / .part 之类的碎片
	ents, _ := os.ReadDir(out)
	for _, e := range ents {
		n := e.Name()
		if filepath.Ext(n) == ".part" || (len(n) > 6 && n[:6] == ".list-") {
			t.Fatalf("残留临时文件: %s", n)
		}
	}
}

// 包里的文件全被新版取代：重写成只含占位文件的包，包本身留着（不删包）
func TestRewritePackEmptyBecomesPlaceholder(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)
	mustMkdir(t, out)

	files := writeFiles(t, src, "f", 3, 1024)
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz}
	pack := "20260101_000000_000.7z"
	if _, err := packOne(repo, set, packPlan{Files: files}, pack, "TESTPW", false); err != nil {
		t.Fatalf("packOne: %v", err)
	}

	idx := newIndex()
	idx.addPack(pack, "", "")
	id2 := idx.addPack("20260102_000000_000.7z", "", "")
	for i := range files {
		idx.Files[files[i].Rel] = &FileEntry{Size: files[i].Size, MTime: files[i].MTime, Pack: id2}
	}

	plan, err := buildCompactPlan(repo, set, idx, "TESTPW", pack)
	if err != nil {
		t.Fatalf("buildCompactPlan: %v", err)
	}
	if len(plan.Keep) != 0 || len(plan.Drop) != 3 {
		t.Fatalf("Keep=%d Drop=%d, want 0/3", len(plan.Keep), len(plan.Drop))
	}
	res, err := rewritePack(repo, set, plan, "TESTPW", false)
	if err != nil {
		t.Fatalf("rewritePack: %v", err)
	}
	if !res.Empty {
		t.Fatal("应该标记为空包")
	}
	if !fileExists(filepath.Join(out, pack)) {
		t.Fatal("包本身必须留着：删包会打乱按名字解压的时间线")
	}

	entries, err := listPackContents(sz, filepath.Join(out, pack), "TESTPW")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir {
			names = append(names, filepath.ToSlash(e.Path))
		}
	}
	if len(names) != 1 || names[0] != placeholderName {
		t.Fatalf("空包内容 = %v, want [%s]", names, placeholderName)
	}
}

// 〔你提〕场景：A 改过几版 → 从源目录删除（留下一条 .rm）→ 某天又放回来。
// compact 把 A 的旧副本全清掉之后，恢复走到那条 .rm 时已经「删不到东西」了。
// 这里要证明的是：这是良性跳过，不中止、也不把后面重新放回来的 A 误删掉。
func TestRestoreRmMissingAfterCompact(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	dest := filepath.Join(dir, "restore")
	mustMkdir(t, src)

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}

	backup := func() BackupResult {
		t.Helper()
		sc, err := scanRepo(repo, set)
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		return runBackup("t", cfg, set, sc, "TESTPW", "cli", false)
	}

	// b.txt 全程不动：保证第一个包不会被腾空，好顺带验证「该留的还在」
	write("b.txt", "bee v1")
	write("a.txt", "alpha v1")
	if r := backup(); r.Err != nil {
		t.Fatalf("backup1: %v", r.Err)
	}
	write("a.txt", "alpha v2, longer content")
	if r := backup(); r.Err != nil {
		t.Fatalf("backup2: %v", r.Err)
	}

	// 删掉 a.txt：只写删除记录，不打新包
	if err := os.Remove(filepath.Join(src, "a.txt")); err != nil {
		t.Fatal(err)
	}
	rDel := backup()
	if rDel.Err != nil {
		t.Fatalf("backup3: %v", rDel.Err)
	}
	if rDel.DelRecord == "" {
		t.Fatal("删除应该留下一条 .rm 记录")
	}

	// 又放回来，进第三个包
	write("a.txt", "alpha v3 back again")
	if r := backup(); r.Err != nil {
		t.Fatalf("backup4: %v", r.Err)
	}

	// compact：前两个包里的 a.txt 旧副本全清掉
	if err := runCompact("t", cfg, set, "TESTPW", "", "0", true, false, false); err != nil {
		t.Fatalf("runCompact: %v", err)
	}

	// 恢复：走到 .rm 时 a.txt 还不存在，必须良性跳过而不是报错中止
	if err := runRestore("t", cfg, set, restoreOpts{
		Dest:      dest,
		Password:  "TESTPW",
		Overwrite: true,
		Wrap:      false,
	}); err != nil {
		t.Fatalf("「删不到」不该让恢复失败: %v", err)
	}

	// 最终结果：a.txt 是最后放回来那一版，b.txt 还在，占位文件没漏进来
	got, err := os.ReadFile(filepath.Join(dest, "a.txt"))
	if err != nil {
		t.Fatalf("a.txt 应该被恢复出来: %v", err)
	}
	if string(got) != "alpha v3 back again" {
		t.Fatalf("a.txt = %q, 应该是最后放回来那一版", string(got))
	}
	if _, err := os.Stat(filepath.Join(dest, "b.txt")); err != nil {
		t.Fatalf("b.txt 应该还在: %v", err)
	}
	if fileExists(filepath.Join(dest, placeholderName)) {
		t.Fatal("占位文件不该出现在恢复结果里")
	}
}

// 删除记录里的路径在恢复目录里本就不存在时，applyDeletions 只记 missing，不报错
func TestApplyDeletionsMissingIsBenign(t *testing.T) {
	dir := t.TempDir()
	removed, missing, err := applyDeletions(dir, []string{"a.txt", "sub/b.txt"})
	if err != nil {
		t.Fatalf("删不到不该报错: %v", err)
	}
	if removed != 0 || missing != 2 {
		t.Fatalf("removed=%d missing=%d, want 0/2", removed, missing)
	}
}

// 端到端：备份 → 改一个文件 → 再备份 → compact。老包里那个被取代的副本要消失，
// 新包不动，verify 仍然通过（manifest 以最后一次登记的哈希为准）。
func TestCompactEndToEnd(t *testing.T) {
	sz := test7z(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	out := filepath.Join(dir, "out")
	mustMkdir(t, src)

	files := writeFiles(t, src, "f", 3, 2048)
	repo := &Repo{Source: src, Output: out}
	set := &Settings{SevenZipPath: sz}
	cfg := &Config{Repos: map[string]*Repo{"t": repo}}

	scan1, err := scanRepo(repo, set)
	if err != nil {
		t.Fatalf("scan1: %v", err)
	}
	r1 := runBackup("t", cfg, set, scan1, "TESTPW", "cli", false)
	if r1.Err != nil || len(r1.Packs) != 1 {
		t.Fatalf("第一次备份: packs=%v err=%v", r1.Packs, r1.Err)
	}
	oldPack := r1.Packs[0]

	// 改 f0，第二次备份会把它写进新包
	if err := os.WriteFile(filepath.Join(src, files[0].Rel), []byte("changed content"), 0o644); err != nil {
		t.Fatal(err)
	}
	scan2, err := scanRepo(repo, set)
	if err != nil {
		t.Fatalf("scan2: %v", err)
	}
	r2 := runBackup("t", cfg, set, scan2, "TESTPW", "cli", false)
	if r2.Err != nil || len(r2.Packs) != 1 {
		t.Fatalf("第二次备份: packs=%v err=%v", r2.Packs, r2.Err)
	}
	newPack := r2.Packs[0]

	// compact：min-gain 0 → 不看阈值，有死副本就重写
	if err := runCompact("t", cfg, set, "TESTPW", "", "0", true, false, false); err != nil {
		t.Fatalf("runCompact: %v", err)
	}

	// 老包里 f0 的旧副本没了，另外两个还在
	old, err := listPackContents(sz, filepath.Join(out, oldPack), "TESTPW")
	if err != nil {
		t.Fatalf("list old: %v", err)
	}
	var oldNames []string
	for _, e := range old {
		if !e.IsDir {
			oldNames = append(oldNames, filepath.ToSlash(e.Path))
		}
	}
	if len(oldNames) != 2 {
		t.Fatalf("老包应剩 2 个文件, got %v", oldNames)
	}
	for _, n := range oldNames {
		if n == files[0].Rel {
			t.Fatalf("老包里不该再有 %s 的旧副本", n)
		}
	}

	// 新包一个字都不该动
	ne, err := listPackContents(sz, filepath.Join(out, newPack), "TESTPW")
	if err != nil {
		t.Fatalf("list new: %v", err)
	}
	n := 0
	for _, e := range ne {
		if !e.IsDir {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("新包应有 1 个文件, got %d", n)
	}

	// 索引没变：f0 仍然指向新包
	idx, err := loadIndex(out)
	if err != nil {
		t.Fatalf("loadIndex: %v", err)
	}
	if fe := idx.Files[files[0].Rel]; fe == nil {
		t.Fatal("f0 不该从索引里消失")
	} else if idx.packName(fe.Pack) != newPack {
		t.Fatalf("f0 应指向新包 %s, 实际 %s", newPack, idx.packName(fe.Pack))
	}

	// manifest 里新增了一条 compact 批次，哈希是新包的
	m, err := loadManifest(out)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	compacted := 0
	for _, b := range m.Batches {
		if b.Op == "compact" {
			compacted++
			if len(b.Files) != 1 || b.Files[0] != oldPack {
				t.Fatalf("compact 批次记的是 %v", b.Files)
			}
		}
	}
	if compacted != 1 {
		t.Fatalf("compact 批次 %d 条, want 1", compacted)
	}

	// 校验必须通过：同一个包出现两次时以最后一次的哈希为准
	if err := runVerify("t", cfg, set, 100); err != nil {
		t.Fatalf("verify 失败: %v", err)
	}

	// 密码不对时必须报错：包哈希对、但 7z 打不开。
	// 历史事故里 verify 只算 SHA-256，查不出「包完好、密码却是别的」。
	if err := runVerify("t", cfg, &Settings{SevenZipPath: sz, DefaultPassword: "WRONGPW"}, 100); err == nil {
		t.Fatal("密码不对时 verify 应该失败")
	}
}
