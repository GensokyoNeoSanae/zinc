using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.IO.Compression;
using System.Reflection;
using System.Text;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Win32;

namespace ZincSetup
{
    /// <summary>命令行选项。GUI 模式忽略除 /S 以外的项；/S 用于自动化安装。</summary>
    internal sealed class Options
    {
        public bool Silent;
        public bool NoReg;          // 测试用：跳过注册（不跑 zinc-install，不写卸载项）
        public bool Cleanup;        // 内部：由卸载器复制到临时目录后调用
        public string Dir = "";
        public string SevenZip = "";
        public string CleanupDir = "";
        public bool CleanupKeepSettings;

        public static Options Parse(string[] args)
        {
            Options o = new Options();
            for (int i = 0; i < args.Length; i++)
            {
                string a = args[i];
                string lower = a.ToLowerInvariant();
                if (lower == "/s" || lower == "/silent") o.Silent = true;
                else if (lower == "/noreg") o.NoReg = true;
                else if (lower == "/cleanup")
                {
                    o.Cleanup = true;
                    if (i + 1 < args.Length) o.CleanupDir = args[++i];
                    if (i + 1 < args.Length) o.CleanupKeepSettings = args[++i] == "1";
                }
                else if (lower.StartsWith("/d=")) o.Dir = a.Substring(3).Trim('"');
                else if (lower.StartsWith("/dir=")) o.Dir = a.Substring(5).Trim('"');
                else if (lower.StartsWith("/7zip=")) o.SevenZip = a.Substring(6).Trim('"');
            }
            return o;
        }
    }

    internal static class Program
    {
        [STAThread]
        private static int Main(string[] args)
        {
            Options opt = Options.Parse(args);
            if (opt.Cleanup)
            {
                Cleaner.Run(opt.CleanupDir, opt.CleanupKeepSettings);
                return 0;
            }
            if (opt.Silent)
            {
                Log.Reset();
                if (string.IsNullOrEmpty(opt.Dir)) opt.Dir = Installer.DefaultDir();
                if (string.IsNullOrEmpty(opt.SevenZip)) opt.SevenZip = Installer.DetectSevenZip();
                bool ok = Installer.InstallCore(opt, Log.Write);
                return ok ? 0 : 1;
            }
            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            Application.Run(new SetupForm(opt));
            return 0;
        }
    }

    /// <summary>静默模式的日志，写到 %TEMP%\zinc-setup.log；GUI 模式不用它。</summary>
    internal static class Log
    {
        private static string _path;

        public static void Reset()
        {
            _path = Path.Combine(Path.GetTempPath(), "zinc-setup.log");
            try { File.WriteAllText(_path, ""); } catch { }
        }

        public static void Write(string s)
        {
            try { File.AppendAllText(_path, s + Environment.NewLine); } catch { }
        }
    }

    internal sealed class RunResult
    {
        public int Exit;
        public string Output = "";
    }

    internal static class Installer
    {
        public static string DefaultDir()
        {
            return Path.Combine(
                Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
                Path.Combine("Programs", "zinc"));
        }

        private static string[] SevenZipNames = new string[] { "7za.exe", "7z.exe" };

        /// <summary>与 zinc-install 相同的查找顺序：PATH → Program Files → 旧默认路径。</summary>
        public static string DetectSevenZip()
        {
            string path = Environment.GetEnvironmentVariable("PATH") ?? "";
            foreach (string raw in path.Split(';'))
            {
                string dir = raw.Trim().Trim('"');
                if (dir.Length == 0) continue;
                foreach (string n in SevenZipNames)
                {
                    try
                    {
                        string p = Path.Combine(dir, n);
                        if (File.Exists(p)) return p;
                    }
                    catch { }
                }
            }
            string[] roots = new string[]
            {
                Environment.GetEnvironmentVariable("ProgramFiles"),
                Environment.GetEnvironmentVariable("ProgramFiles(x86)")
            };
            foreach (string root in roots)
            {
                if (string.IsNullOrEmpty(root)) continue;
                foreach (string n in SevenZipNames)
                {
                    string p = Path.Combine(root, Path.Combine("7-Zip", n));
                    if (File.Exists(p)) return p;
                }
            }
            string fallback = @"E:\Tools\7zip\7za.exe";
            if (File.Exists(fallback)) return fallback;
            return "";
        }

        public static bool InstallCore(Options o, Action<string> log)
        {
            try
            {
                if (string.IsNullOrEmpty(o.Dir)) o.Dir = DefaultDir();
                Directory.CreateDirectory(o.Dir);

                log("释放文件到 " + o.Dir);
                ExtractZip("payload.zip", o.Dir, log);
                ExtractFile("uninstall.exe", Path.Combine(o.Dir, "uninstall.exe"), log);

                if (o.NoReg)
                {
                    log("已跳过注册（/noreg）");
                    return true;
                }

                log("注册通知与用户 PATH");
                RunResult r = RunCapture(Path.Combine(o.Dir, "zinc-install.exe"),
                    "--7zip \"" + o.SevenZip + "\" --no-pause");
                foreach (string line in r.Output.Split('\n'))
                {
                    string t = line.TrimEnd('\r');
                    if (t.Length > 0) log(t);
                }
                if (r.Exit != 0)
                {
                    log("注册失败（退出码 " + r.Exit + "）");
                    return false;
                }

                WriteUninstallKey(o);
                log("已写入卸载信息");
                return true;
            }
            catch (Exception ex)
            {
                log("安装失败: " + ex.Message);
                return false;
            }
        }

        public static void ExtractFile(string name, string targetPath, Action<string> log)
        {
            Assembly asm = Assembly.GetExecutingAssembly();
            using (Stream res = asm.GetManifestResourceStream(name))
            {
                if (res == null) throw new InvalidOperationException("缺少内置资源: " + name);
                using (FileStream fs = File.Create(targetPath)) res.CopyTo(fs);
                log("  " + name + " -> " + targetPath + " (" + new FileInfo(targetPath).Length + " bytes)");
            }
        }

        public static void ExtractZip(string name, string targetDir, Action<string> log)
        {
            Assembly asm = Assembly.GetExecutingAssembly();
            using (Stream res = asm.GetManifestResourceStream(name))
            {
                if (res == null) throw new InvalidOperationException("缺少内置资源: " + name);
                string tmp = Path.Combine(Path.GetTempPath(), "zinc-" + Guid.NewGuid().ToString("N") + ".tmp");
                using (FileStream fs = File.Create(tmp)) res.CopyTo(fs);
                try
                {
                    log("  " + name + " -> " + tmp + " (" + new FileInfo(tmp).Length + " bytes)");
                    using (ZipArchive zip = ZipFile.OpenRead(tmp))
                    {
                        foreach (ZipArchiveEntry entry in zip.Entries)
                        {
                            if (string.IsNullOrEmpty(entry.Name)) continue;
                            string dest = Path.Combine(targetDir, entry.FullName.Replace('/', Path.DirectorySeparatorChar));
                            Directory.CreateDirectory(Path.GetDirectoryName(dest));
                            entry.ExtractToFile(dest, true);
                        }
                    }
                }
                finally
                {
                    try { File.Delete(tmp); } catch { }
                }
            }
        }

        public static RunResult RunCapture(string exe, string args)
        {
            RunResult r = new RunResult();
            ProcessStartInfo psi = new ProcessStartInfo(exe, args);
            psi.UseShellExecute = false;
            psi.CreateNoWindow = true;
            psi.RedirectStandardOutput = true;
            psi.RedirectStandardError = true;
            psi.StandardOutputEncoding = Encoding.UTF8;
            psi.StandardErrorEncoding = Encoding.UTF8;
            using (Process p = Process.Start(psi))
            {
                string stdout = p.StandardOutput.ReadToEnd();
                string stderr = p.StandardError.ReadToEnd();
                p.WaitForExit();
                r.Exit = p.ExitCode;
                r.Output = stdout + stderr;
            }
            return r;
        }

        public static void WriteUninstallKey(Options o)
        {
            string uninstaller = Path.Combine(o.Dir, "uninstall.exe");
            using (RegistryKey k = Registry.CurrentUser.CreateSubKey(
                @"Software\Microsoft\Windows\CurrentVersion\Uninstall\zinc"))
            {
                k.SetValue("DisplayName", "zinc");
                k.SetValue("DisplayVersion", Build.Version);
                k.SetValue("Publisher", "GensokyoNeoSanae");
                k.SetValue("DisplayIcon", Path.Combine(o.Dir, "zinc.ico"));
                k.SetValue("InstallLocation", o.Dir);
                k.SetValue("UninstallString", "\"" + uninstaller + "\"");
                k.SetValue("QuietUninstallString", "\"" + uninstaller + "\" /S");
                k.SetValue("URLInfoAbout", "https://github.com/GensokyoNeoSanae/zinc");
                k.SetValue("NoModify", 1, RegistryValueKind.DWord);
                k.SetValue("NoRepair", 1, RegistryValueKind.DWord);
                k.SetValue("EstimatedSize", EstimateKb(o.Dir), RegistryValueKind.DWord);
            }
        }

        private static int EstimateKb(string dir)
        {
            try
            {
                long total = 0;
                foreach (string f in Directory.GetFiles(dir)) total += new FileInfo(f).Length;
                return (int)(total / 1024);
            }
            catch { return 0; }
        }
    }

    /// <summary>从临时目录执行的收尾：删文件、删目录、删自己。</summary>
    internal static class Cleaner
    {
        public static void Run(string dir, bool keepSettings)
        {
            if (string.IsNullOrEmpty(dir)) return;
            Thread.Sleep(800);
            for (int attempt = 0; attempt < 10; attempt++)
            {
                try
                {
                    if (keepSettings) DeleteExcept(dir, new string[] { "settings.json", "repos.json", "logs", "manifests" });
                    else DeleteAll(dir);
                    break;
                }
                catch
                {
                    Thread.Sleep(400);
                }
            }
            SelfDelete();
        }

        private static void DeleteAll(string dir)
        {
            if (!Directory.Exists(dir)) return;
            foreach (string f in Directory.GetFiles(dir))
            {
                try { File.SetAttributes(f, FileAttributes.Normal); File.Delete(f); } catch { }
            }
            foreach (string d in Directory.GetDirectories(dir))
            {
                try { Directory.Delete(d, true); } catch { }
            }
            try { Directory.Delete(dir, true); } catch { }
        }

        private static void DeleteExcept(string dir, string[] keep)
        {
            if (!Directory.Exists(dir)) return;
            List<string> keepList = new List<string>(keep);
            foreach (string f in Directory.GetFiles(dir))
            {
                if (keepList.Contains(Path.GetFileName(f).ToLowerInvariant())) continue;
                try { File.SetAttributes(f, FileAttributes.Normal); File.Delete(f); } catch { }
            }
            foreach (string d in Directory.GetDirectories(dir))
            {
                if (keepList.Contains(Path.GetFileName(d).ToLowerInvariant())) continue;
                try { Directory.Delete(d, true); } catch { }
            }
        }

        private static void SelfDelete()
        {
            try
            {
                string self = Application.ExecutablePath;
                ProcessStartInfo psi = new ProcessStartInfo("cmd.exe",
                    "/c ping -n 3 127.0.0.1 >nul & del /f /q \"" + self + "\"");
                psi.UseShellExecute = false;
                psi.CreateNoWindow = true;
                psi.WindowStyle = ProcessWindowStyle.Hidden;
                Process.Start(psi);
            }
            catch { }
        }
    }

    internal sealed class SetupForm : Form
    {
        private readonly Options _opt;
        private readonly Panel[] _pages;
        private int _page;
        private bool _installing;

        private readonly Label _headerTitle;
        private readonly Label _headerSub;
        private readonly Button _back;
        private readonly Button _next;
        private readonly Button _cancel;

        private readonly CheckBox _accept;
        private readonly TextBox _dirBox;
        private readonly TextBox _sevenZipBox;
        private readonly Label _sevenZipStatus;
        private readonly ProgressBar _progress;
        private readonly TextBox _log;
        private readonly Label _finishText;

        private string _headerTitleText;
        private string _headerSubText;

        public SetupForm(Options opt)
        {
            _opt = opt;
            Text = "zinc " + Build.Version + " 安装向导";
            ClientSize = new Size(560, 420);
            FormBorderStyle = FormBorderStyle.FixedSingle;
            MaximizeBox = false;
            MinimizeBox = false;
            StartPosition = FormStartPosition.CenterScreen;
            AutoScaleMode = AutoScaleMode.None;
            Font = new Font("Segoe UI", 9F);
            BackColor = Color.White;
            try { Icon = Icon.ExtractAssociatedIcon(Application.ExecutablePath); } catch { }

            Panel header = new Panel();
            header.Location = new Point(0, 0);
            header.Size = new Size(560, 64);
            header.BackColor = Color.White;
            header.Paint += new PaintEventHandler(delegate(object s, PaintEventArgs e)
            {
                using (Pen p = new Pen(Color.FromArgb(230, 230, 230)))
                {
                    e.Graphics.DrawLine(p, 0, header.Height - 1, header.Width, header.Height - 1);
                }
            });

            _headerTitle = new Label();
            _headerTitle.Text = "";
            _headerTitle.Font = new Font("Segoe UI", 14F, FontStyle.Bold);
            _headerTitle.Location = new Point(20, 12);
            _headerTitle.AutoSize = true;
            _headerTitleText = "";

            _headerSub = new Label();
            _headerSub.Text = "";
            _headerSub.ForeColor = Color.FromArgb(90, 90, 90);
            _headerSub.Location = new Point(22, 40);
            _headerSub.AutoSize = true;

            header.Controls.Add(_headerTitle);
            header.Controls.Add(_headerSub);
            Controls.Add(header);

            Panel body = new Panel();
            body.Location = new Point(0, 64);
            body.Size = new Size(560, 296);
            body.BackColor = Color.White;
            body.Padding = new Padding(16);
            Controls.Add(body);

            Panel footer = new Panel();
            footer.Location = new Point(0, 360);
            footer.Size = new Size(560, 60);
            footer.BackColor = Color.FromArgb(246, 246, 246);
            Controls.Add(footer);

            _back = MakeButton("上一步", new Point(312, 16));
            _next = MakeButton("下一步", new Point(392, 16));
            _cancel = MakeButton("取消", new Point(472, 16));
            _back.Click += new EventHandler(delegate(object s, EventArgs e) { SetPage(_page - 1); });
            _next.Click += new EventHandler(OnNext);
            _cancel.Click += new EventHandler(delegate(object s, EventArgs e) { Close(); });
            footer.Controls.Add(_back);
            footer.Controls.Add(_next);
            footer.Controls.Add(_cancel);

            _accept = new CheckBox();
            _accept.Text = "我已阅读并同意 MIT 许可证";
            _accept.Location = new Point(0, 226);
            _accept.Size = new Size(520, 24);
            _accept.CheckedChanged += new EventHandler(delegate(object s, EventArgs e) { if (_page == 0) _next.Enabled = _accept.Checked; });

            TextBox license = new TextBox();
            license.Multiline = true;
            license.ReadOnly = true;
            license.ScrollBars = ScrollBars.Vertical;
            license.Location = new Point(0, 26);
            license.Size = new Size(520, 190);
            license.Text = ReadTextResource("license.txt");

            _dirBox = new TextBox();
            _dirBox.Location = new Point(0, 0);
            _dirBox.Size = new Size(430, 26);
            _dirBox.Text = opt.Dir.Length > 0 ? opt.Dir : Installer.DefaultDir();

            Button browseDir = MakeButton("浏览…", new Point(440, 0));
            browseDir.Click += new EventHandler(delegate(object s, EventArgs e)
            {
                using (FolderBrowserDialog dlg = new FolderBrowserDialog())
                {
                    dlg.Description = "选择安装目录";
                    dlg.SelectedPath = _dirBox.Text;
                    if (dlg.ShowDialog(this) == DialogResult.OK) _dirBox.Text = dlg.SelectedPath;
                }
            });

            _sevenZipBox = new TextBox();
            _sevenZipBox.Location = new Point(0, 0);
            _sevenZipBox.Size = new Size(430, 26);
            _sevenZipBox.Text = Installer.DetectSevenZip();

            Button browseZip = MakeButton("浏览…", new Point(440, 0));
            browseZip.Click += new EventHandler(delegate(object s, EventArgs e)
            {
                using (OpenFileDialog dlg = new OpenFileDialog())
                {
                    dlg.Title = "选择 7z.exe 或 7za.exe";
                    dlg.Filter = "7-Zip 命令行|7z*.exe|可执行文件|*.exe";
                    if (dlg.ShowDialog(this) == DialogResult.OK) { _sevenZipBox.Text = dlg.FileName; UpdateSevenZipStatus(); }
                }
            });
            _sevenZipBox.TextChanged += new EventHandler(delegate(object s, EventArgs e) { UpdateSevenZipStatus(); });

            LinkLabel download = new LinkLabel();
            download.Text = "没有装 7-Zip？打开官网下载";
            download.Location = new Point(0, 38);
            download.AutoSize = true;
            download.LinkClicked += new LinkLabelLinkClickedEventHandler(delegate(object s, LinkLabelLinkClickedEventArgs e)
            {
                try { Process.Start("https://www.7-zip.org/"); } catch { }
            });

            _sevenZipStatus = new Label();
            _sevenZipStatus.Location = new Point(0, 64);
            _sevenZipStatus.Size = new Size(520, 40);
            _sevenZipStatus.ForeColor = Color.FromArgb(90, 90, 90);

            _progress = new ProgressBar();
            _progress.Location = new Point(0, 12);
            _progress.Size = new Size(520, 18);
            _progress.Style = ProgressBarStyle.Marquee;
            _progress.MarqueeAnimationSpeed = 25;

            _log = new TextBox();
            _log.Multiline = true;
            _log.ReadOnly = true;
            _log.ScrollBars = ScrollBars.Vertical;
            _log.Location = new Point(0, 42);
            _log.Size = new Size(520, 200);
            _log.BackColor = Color.FromArgb(250, 250, 250);

            _finishText = new Label();
            _finishText.Location = new Point(0, 0);
            _finishText.Size = new Size(520, 240);
            _finishText.Text = "";

            _pages = new Panel[5];
            _pages[0] = MakePage(new Control[] { Header("许可证"), license, _accept });
            _pages[1] = MakePage(new Control[] { _dirBox, browseDir, Note("程序目录里会保存 settings.json、repos.json 和 logs。请选择你有写权限的位置；默认在 LocalAppData 下。") });
            _pages[2] = MakePage(new Control[] { _sevenZipBox, browseZip, download, _sevenZipStatus });
            _pages[3] = MakePage(new Control[] { _progress, _log });
            _pages[4] = MakePage(new Control[] { _finishText });

            foreach (Panel p in _pages) body.Controls.Add(p);
            // Dock 的布局顺序看 z-order：把 Top/Bottom 提到前面，Fill 才不会盖住标题栏和按钮栏。
            header.BringToFront();
            footer.BringToFront();
            UpdateSevenZipStatus();
            SetPage(0);
            ActiveControl = _accept;
        }

        private static Label Header(string text)
        {
            Label l = new Label();
            l.Text = text;
            l.Font = new Font("Segoe UI", 10F, FontStyle.Bold);
            l.Location = new Point(0, 0);
            l.AutoSize = true;
            return l;
        }

        private static Label Note(string text)
        {
            Label l = new Label();
            l.Text = text;
            l.Location = new Point(0, 76);
            l.Size = new Size(520, 60);
            l.ForeColor = Color.FromArgb(90, 90, 90);
            return l;
        }

        private static Button MakeButton(string text, Point at)
        {
            Button b = new Button();
            b.Text = text;
            b.Location = at;
            b.Size = new Size(72, 28);
            b.FlatStyle = FlatStyle.System;
            return b;
        }

        private static Panel MakePage(Control[] controls)
        {
            Panel p = new Panel();
            p.Dock = DockStyle.Fill;
            p.BackColor = Color.White;
            foreach (Control c in controls) p.Controls.Add(c);
            return p;
        }

        private static string ReadTextResource(string name)
        {
            using (Stream s = Assembly.GetExecutingAssembly().GetManifestResourceStream(name))
            {
                if (s == null) return "";
                using (StreamReader r = new StreamReader(s, Encoding.UTF8))
                {
                    string text = r.ReadToEnd().Replace("\r\n", "\n").Replace("\r", "\n");
                    return text.Replace("\n", "\r\n");
                }
            }
        }

        private void UpdateSevenZipStatus()
        {
            string p = _sevenZipBox.Text.Trim();
            if (p.Length == 0)
            {
                _sevenZipStatus.ForeColor = Color.FromArgb(200, 60, 60);
                _sevenZipStatus.Text = "zinc 需要 7z.exe 或 7za.exe。装好 7-Zip 后点“浏览…”指定，或把路径粘贴到这里。";
            }
            else if (!File.Exists(p))
            {
                _sevenZipStatus.ForeColor = Color.FromArgb(200, 60, 60);
                _sevenZipStatus.Text = "这个路径不存在：" + p;
            }
            else
            {
                _sevenZipStatus.ForeColor = Color.FromArgb(30, 120, 30);
                _sevenZipStatus.Text = "将使用：" + p;
            }
        }

        private void SetPage(int index)
        {
            if (index < 0 || index >= _pages.Length) return;
            _page = index;
            for (int i = 0; i < _pages.Length; i++) _pages[i].Visible = i == index;

            string title, sub;
            switch (index)
            {
                case 0:
                    title = "欢迎使用 zinc";
                    sub = "文件级增量备份，产物是普通 7-Zip 归档。";
                    _back.Visible = false;
                    _next.Text = "下一步";
                    _next.Enabled = _accept.Checked;
                    _cancel.Visible = true;
                    break;
                case 1:
                    title = "安装位置";
                    sub = "程序会装进这个目录，并加入当前用户的 PATH。";
                    _back.Visible = true;
                    _next.Text = "下一步";
                    _next.Enabled = true;
                    break;
                case 2:
                    title = "7-Zip";
                    sub = "zinc 调用 7z.exe 或 7za.exe 完成压缩与恢复。";
                    _back.Visible = true;
                    _next.Text = "安装";
                    _next.Enabled = File.Exists(_sevenZipBox.Text.Trim());
                    UpdateSevenZipStatus();
                    break;
                case 3:
                    title = "正在安装";
                    sub = "正在释放文件并注册通知与 PATH。";
                    _back.Visible = false;
                    _next.Text = "请稍候";
                    _next.Enabled = false;
                    _cancel.Enabled = false;
                    break;
                default:
                    title = "安装完成";
                    sub = "zinc 已经可以使用。";
                    _back.Visible = false;
                    _next.Text = "完成";
                    _next.Enabled = true;
                    _cancel.Visible = false;
                    break;
            }
            SetHeader(title, sub);
        }

        private void SetHeader(string title, string sub)
        {
            _headerTitleText = title;
            _headerSubText = sub;
            _headerTitle.Text = title;
            _headerSub.Text = sub;
        }

        private void OnNext(object sender, EventArgs e)
        {
            if (_installing) return;
            switch (_page)
            {
                case 0:
                    SetPage(1);
                    break;
                case 1:
                    string dir = _dirBox.Text.Trim();
                    if (dir.Length == 0) { Warn("请填写安装目录。"); return; }
                    try { Directory.CreateDirectory(dir); }
                    catch (Exception ex) { Warn("这个目录不可写：" + ex.Message); return; }
                    _opt.Dir = dir;
                    SetPage(2);
                    break;
                case 2:
                    string sz = _sevenZipBox.Text.Trim();
                    if (!File.Exists(sz)) { Warn("7-Zip 路径无效，请重新选择。"); return; }
                    _opt.SevenZip = sz;
                    SetPage(3);
                    StartInstall();
                    break;
                case 4:
                    Close();
                    break;
            }
        }

        private void StartInstall()
        {
            _installing = true;
            _log.Text = "";
            Options o = _opt;
            Thread t = new Thread(delegate()
            {
                bool ok = Installer.InstallCore(o, delegate(string line) { AppendLog(line); });
                try { BeginInvoke((Action)delegate() { FinishInstall(ok); }); }
                catch { }
            });
            t.IsBackground = true;
            t.Start();
        }

        private void AppendLog(string line)
        {
            if (InvokeRequired)
            {
                try { BeginInvoke((Action)delegate() { AppendLog(line); }); } catch { }
                return;
            }
            _log.AppendText(line + "\r\n");
            _log.SelectionStart = _log.Text.Length;
            _log.ScrollToCaret();
        }

        private void FinishInstall(bool ok)
        {
            _installing = false;
            if (ok)
            {
                _finishText.Text =
                    "安装完成。\r\n\r\n" +
                    "程序目录：\r\n" + _opt.Dir + "\r\n\r\n" +
                    "下一步（新开一个终端）：\r\n" +
                    "    zinc root D:\\Backup\\zinc   # 选择备份输出根目录\r\n" +
                    "    zinc add pictures --source D:\\Storage\\Pictures\r\n" +
                    "    zinc backup pictures\r\n\r\n" +
                    "“应用和功能”里可以随时卸载；备份数据不会被卸载程序删除。";
                SetPage(4);
            }
            else
            {
                _cancel.Enabled = true;
                Warn("安装没有完成，详情见下方日志。\r\n\r\n" + _log.Text);
                SetPage(2);
            }
        }

        private void Warn(string text)
        {
            MessageBox.Show(this, text, "zinc", MessageBoxButtons.OK, MessageBoxIcon.Warning);
        }
    }
}
