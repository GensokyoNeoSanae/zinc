using System;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.Text;
using System.Threading;
using System.Windows.Forms;
using Microsoft.Win32;

namespace ZincUninstall
{
    internal static class Program
    {
        [STAThread]
        private static int Main(string[] args)
        {
            bool silent = false;
            bool keepSettings = true;
            bool keepTasks = false;
            string cleanupDir = null;
            bool cleanupKeep = false;
            for (int i = 0; i < args.Length; i++)
            {
                string a = args[i].ToLowerInvariant();
                if (a == "/s" || a == "/silent") silent = true;
                else if (a == "/cleanup")
                {
                    if (i + 1 < args.Length) cleanupDir = args[++i];
                    if (i + 1 < args.Length) cleanupKeep = args[++i] == "1";
                }
                else if (a == "/keepsettings") keepSettings = true;
                else if (a == "/deletesettings") keepSettings = false;
                else if (a == "/keeptasks") keepTasks = true;
            }

            if (cleanupDir != null)
            {
                Cleaner.Run(cleanupDir, cleanupKeep);
                return 0;
            }

            if (silent)
            {
                Uninstaller.Run(keepSettings, keepTasks, null);
                return 0;
            }

            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            Application.Run(new UninstallForm(keepTasks));
            return 0;
        }
    }

    internal static class Uninstaller
    {
        public static string InstallDir()
        {
            return Path.GetDirectoryName(Application.ExecutablePath);
        }

        /// <summary>
        /// 逆着安装做四件事：清任务、清通知注册、清 PATH、删文件；数据目录一律不碰。
        /// </summary>
        public static bool Run(bool keepSettings, bool keepTasks, Action<string> log)
        {
            string dir = InstallDir();
            try
            {
                string child = Path.Combine(dir, "zinc-uninstall.exe");
                if (File.Exists(child))
                {
                    string args = "--yes --no-pause" + (keepTasks ? " --keep-schedule" : "");
                    ProcessStartInfo psi = new ProcessStartInfo(child, args);
                    psi.UseShellExecute = false;
                    psi.CreateNoWindow = true;
                    psi.RedirectStandardOutput = true;
                    psi.RedirectStandardError = true;
                    psi.StandardOutputEncoding = Encoding.UTF8;
                    psi.StandardErrorEncoding = Encoding.UTF8;
                    using (Process p = Process.Start(psi))
                    {
                        string text = p.StandardOutput.ReadToEnd() + p.StandardError.ReadToEnd();
                        p.WaitForExit();
                        if (log != null) log("已移除计划任务、通知注册与 PATH 项");
                        if (log != null && text.Trim().Length > 0) log(text.Trim());
                    }
                }

                try
                {
                    Registry.CurrentUser.DeleteSubKeyTree(
                        @"Software\Microsoft\Windows\CurrentVersion\Uninstall\zinc", false);
                }
                catch { }

                // 自己在运行，删不掉自己：复制到临时目录，由临时副本负责收尾。
                string temp = Path.Combine(Path.GetTempPath(),
                    "zinc-unins-" + Guid.NewGuid().ToString("N") + ".exe");
                File.Copy(Application.ExecutablePath, temp, true);
                ProcessStartInfo start = new ProcessStartInfo(temp,
                    "/cleanup \"" + dir + "\" " + (keepSettings ? "1" : "0"));
                start.UseShellExecute = false;
                start.CreateNoWindow = true;
                Process.Start(start);
                return true;
            }
            catch (Exception ex)
            {
                if (log != null) log("卸载失败: " + ex.Message);
                return false;
            }
        }
    }

    internal static class Cleaner
    {
        public static void Run(string dir, bool keepSettings)
        {
            if (string.IsNullOrEmpty(dir)) return;
            Thread.Sleep(900);
            for (int attempt = 0; attempt < 10; attempt++)
            {
                try
                {
                    if (keepSettings)
                        DeleteExcept(dir, new string[] { "settings.json", "repos.json", "logs", "manifests" });
                    else
                        DeleteAll(dir);
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
            foreach (string f in Directory.GetFiles(dir))
            {
                string name = Path.GetFileName(f).ToLowerInvariant();
                bool keepIt = false;
                foreach (string k in keep) if (k == name) keepIt = true;
                if (keepIt) continue;
                try { File.SetAttributes(f, FileAttributes.Normal); File.Delete(f); } catch { }
            }
            foreach (string d in Directory.GetDirectories(dir))
            {
                string name = Path.GetFileName(d).ToLowerInvariant();
                bool keepIt = false;
                foreach (string k in keep) if (k == name) keepIt = true;
                if (keepIt) continue;
                try { Directory.Delete(d, true); } catch { }
            }
            // 保留项不存在时目录已经空了，顺手删掉，别留一个空壳。
            try { Directory.Delete(dir, true); } catch { }
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

    internal sealed class UninstallForm : Form
    {
        private readonly CheckBox _keepSettings;
        private readonly CheckBox _keepTasks;
        private readonly Button _uninstall;
        private readonly Button _cancel;
        private readonly Label _status;
        private readonly ProgressBar _progress;

        public UninstallForm(bool keepTasksDefault)
        {
            Text = "卸载 zinc";
            ClientSize = new Size(500, 300);
            FormBorderStyle = FormBorderStyle.FixedDialog;
            MaximizeBox = false;
            MinimizeBox = false;
            StartPosition = FormStartPosition.CenterScreen;
            AutoScaleMode = AutoScaleMode.None;
            Font = new Font("Segoe UI", 9F);
            BackColor = Color.White;
            try { Icon = Icon.ExtractAssociatedIcon(Application.ExecutablePath); } catch { }

            Label title = new Label();
            title.Text = "卸载 zinc";
            title.Font = new Font("Segoe UI", 13F, FontStyle.Bold);
            title.Location = new Point(20, 18);
            title.AutoSize = true;
            Controls.Add(title);

            Label path = new Label();
            path.Text = "程序目录：" + Uninstaller.InstallDir();
            path.Location = new Point(22, 52);
            path.Size = new Size(456, 40);
            Controls.Add(path);

            _keepSettings = new CheckBox();
            _keepSettings.Text = "保留设置与日志（settings.json / repos.json / logs）";
            _keepSettings.Location = new Point(22, 100);
            _keepSettings.Size = new Size(456, 24);
            _keepSettings.Checked = true;
            Controls.Add(_keepSettings);

            _keepTasks = new CheckBox();
            _keepTasks.Text = "保留计划任务（默认会删除）";
            _keepTasks.Location = new Point(22, 128);
            _keepTasks.Size = new Size(456, 24);
            _keepTasks.Checked = keepTasksDefault;
            Controls.Add(_keepTasks);

            Label note = new Label();
            note.Text = "备份数据、输出目录里的 .7z 和 manifest 不会被删除。";
            note.ForeColor = Color.FromArgb(90, 90, 90);
            note.Location = new Point(22, 164);
            note.Size = new Size(456, 24);
            Controls.Add(note);

            _status = new Label();
            _status.Text = "";
            _status.Location = new Point(22, 196);
            _status.Size = new Size(456, 24);
            Controls.Add(_status);

            _progress = new ProgressBar();
            _progress.Style = ProgressBarStyle.Marquee;
            _progress.MarqueeAnimationSpeed = 25;
            _progress.Location = new Point(22, 224);
            _progress.Size = new Size(456, 14);
            _progress.Visible = false;
            Controls.Add(_progress);

            _uninstall = new Button();
            _uninstall.Text = "卸载";
            _uninstall.Location = new Point(316, 254);
            _uninstall.Size = new Size(76, 28);
            _uninstall.FlatStyle = FlatStyle.System;
            _uninstall.Click += new EventHandler(OnUninstall);
            Controls.Add(_uninstall);

            _cancel = new Button();
            _cancel.Text = "取消";
            _cancel.Location = new Point(400, 254);
            _cancel.Size = new Size(76, 28);
            _cancel.FlatStyle = FlatStyle.System;
            _cancel.Click += new EventHandler(delegate(object s, EventArgs e) { Close(); });
            Controls.Add(_cancel);
        }

        private void OnUninstall(object sender, EventArgs e)
        {
            bool keepSettings = _keepSettings.Checked;
            bool keepTasks = _keepTasks.Checked;
            _uninstall.Enabled = false;
            _cancel.Enabled = false;
            _keepSettings.Enabled = false;
            _keepTasks.Enabled = false;
            _progress.Visible = true;
            _status.Text = "正在卸载…";

            Thread t = new Thread(delegate()
            {
                bool ok = Uninstaller.Run(keepSettings, keepTasks, delegate(string line)
                {
                    try { BeginInvoke((Action)delegate() { _status.Text = line; }); } catch { }
                });
                try
                {
                    BeginInvoke((Action)delegate()
                    {
                        _progress.Visible = false;
                        if (!ok)
                        {
                            _status.Text = "卸载失败";
                            MessageBox.Show(this, "卸载没有完成。可以关闭窗口后重新运行 uninstall.exe。", "zinc",
                                MessageBoxButtons.OK, MessageBoxIcon.Warning);
                            _uninstall.Enabled = true;
                            _cancel.Enabled = true;
                            return;
                        }
                        string extra = keepSettings
                            ? "\r\n\r\n已保留设置与日志，目录仍在：" + Uninstaller.InstallDir()
                            : "";
                        MessageBox.Show(this, "zinc 已卸载。" + extra, "zinc",
                            MessageBoxButtons.OK, MessageBoxIcon.Information);
                        Close();
                    });
                }
                catch { }
            });
            t.IsBackground = true;
            t.Start();
        }
    }
}
