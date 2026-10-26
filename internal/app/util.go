package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
	procSetConsoleCP       = kernel32.NewProc("SetConsoleCP")
)

func init() {
	// 控制台按 UTF-8 显示中文
	procSetConsoleOutputCP.Call(65001)
	procSetConsoleCP.Call(65001)
}

// appDir 返回 zinc.exe 所在目录（repos.json、manifests、日志都放这里）。
// tools\ 下的 register / unregister 辅助程序要认回上一级，否则 PATH 和计划任务会指到 tools。
func appDir() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if strings.EqualFold(filepath.Base(dir), "tools") {
			parent := filepath.Dir(dir)
			if fileExists(filepath.Join(parent, "zinc.exe")) {
				dir = parent
			}
		}
		if _, statErr := os.Stat(dir); statErr == nil {
			return dir
		}
	}
	wd, _ := os.Getwd()
	return wd
}

func configPath() string   { return filepath.Join(appDir(), "repos.json") }
func logPath() string      { return filepath.Join(logsDir(), "zinc-"+time.Now().Format("20060102")+".log") }
func manifestsDir() string { return filepath.Join(appDir(), "manifests") }
func logsDir() string      { return filepath.Join(appDir(), "logs") }

// stdinReader 复用同一个缓冲读取器：连续两次 ReadString 不会吞掉第二行。
var stdinReader = bufio.NewReader(os.Stdin)

// readLine 读一整行（保留空格，去掉行尾 CR/LF）；EOF 时返回已读到的部分。
func readLine() string {
	s, err := stdinReader.ReadString('\n')
	if err != nil && s == "" {
		return ""
	}
	return strings.TrimRight(s, "\r\n")
}

// writeFileSync 写文件并 fsync，用于配置/索引/manifest 这类小块关键文件。
func writeFileSync(p string, b []byte, perm os.FileMode) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// powershellPath 优先用系统目录下的 powershell.exe，避免 PATH 被改坏时找不到。
func powershellPath() string {
	if root := os.Getenv("SystemRoot"); root != "" {
		p := filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "powershell.exe"
}

// absClean 返回绝对化并清理过的路径。
func absClean(p string) (string, error) {
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(a), nil
}

// samePathFold 按 Windows 的大小写不敏感语义比较两个路径。
func samePathFold(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// withinFold 判断 child 是否等于 parent 或位于 parent 之下（大小写不敏感）。
func withinFold(child, parent string) bool {
	c := strings.ToLower(filepath.Clean(child))
	p := strings.ToLower(strings.TrimSuffix(filepath.Clean(parent), string(filepath.Separator)))
	if c == p {
		return true
	}
	return strings.HasPrefix(c, p+string(filepath.Separator))
}

// isVolumeRoot 判断路径是否是盘根（C:\ 或 \\server\share）。
func isVolumeRoot(p string) bool {
	c := filepath.Clean(p)
	vol := filepath.VolumeName(c)
	if vol == "" {
		return false
	}
	rest := strings.TrimPrefix(c, vol)
	return rest == "" || rest == string(filepath.Separator)
}

// validateRepoPaths 校验源目录与输出目录不互相包含。
func validateRepoPaths(source, output string) error {
	s, err := absClean(source)
	if err != nil {
		return fmt.Errorf(M.PathInvalid, source, err)
	}
	o, err := absClean(output)
	if err != nil {
		return fmt.Errorf(M.PathInvalid, output, err)
	}
	if samePathFold(s, o) {
		return fmt.Errorf(M.SourceOutputSame, s)
	}
	if withinFold(o, s) {
		return fmt.Errorf(M.OutputInsideSource, o, s)
	}
	if withinFold(s, o) {
		return fmt.Errorf(M.SourceInsideOutput, s, o)
	}
	return nil
}

// logKeep 是 logs 目录最多保留的日志份数（含当天）。
// 当天文件还不存在时只留 logKeep-1 份历史，给今天留一个位置。
const logKeep = 10

// rotateLogs 在启动时整理 logs：迁移旧的单文件 zinc.log，按天只保留最近若干份。
func rotateLogs() {
	dir := logsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	migrateLegacyLog(dir)

	today := "zinc-" + time.Now().Format("20060102") + ".log"
	ents, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	files := make([]string, 0, len(ents))
	todayExists := false
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !isLogName(name) {
			continue
		}
		files = append(files, name)
		if name == today {
			todayExists = true
		}
	}

	limit := logKeep - 1
	if todayExists {
		limit = logKeep
	}
	if len(files) <= limit {
		return
	}

	sort.Strings(files) // zinc-YYYYMMDD.log 的字典序就是日期序
	n := len(files)
	for _, name := range files {
		if n <= limit {
			break
		}
		if name == today {
			continue
		}
		if os.Remove(filepath.Join(dir, name)) == nil {
			n--
		}
	}
}

// isLogName 判断是否是本程序生成的按天日志文件名。
func isLogName(name string) bool {
	if !strings.HasPrefix(name, "zinc-") || !strings.HasSuffix(name, ".log") {
		return false
	}
	d := strings.TrimSuffix(strings.TrimPrefix(name, "zinc-"), ".log")
	if len(d) != 8 {
		return false
	}
	_, err := time.ParseInLocation("20060102", d, time.Local)
	return err == nil
}

// migrateLegacyLog 把旧版放在程序目录的 zinc.log 迁进 logs。
// 目标日期文件不存在时直接改名（原子、不会重复追加）；已存在才追加后删除。
func migrateLegacyLog(dir string) {
	old := filepath.Join(appDir(), "zinc.log")
	st, err := os.Stat(old)
	if err != nil {
		return
	}
	target := filepath.Join(dir, "zinc-"+st.ModTime().Format("20060102")+".log")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		if os.Rename(old, target) == nil {
			return
		}
	}
	b, err := os.ReadFile(old)
	if err != nil {
		return
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, werr := f.Write(b)
	f.Close()
	if werr == nil {
		os.Remove(old)
	}
}

// logf 追加一条日志；失败静默，不影响主流程
func logf(format string, args ...any) {
	line := fmt.Sprintf("%s %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	if err := os.MkdirAll(logsDir(), 0755); err != nil {
		return
	}
	f, err := os.OpenFile(logPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line)
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}

// isConsole 判断标准输入是不是控制台（双击 exe 时是，管道或重定向时不是）。
func isConsole() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// pauseUnless 在交互式控制台里等一次回车，双击运行时窗口才不会一闪而过。
func pauseUnless(noPause bool) {
	if noPause || !isConsole() {
		return
	}
	fmt.Print(M.PressEnter)
	readLine()
}
