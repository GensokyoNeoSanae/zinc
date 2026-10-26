// 开发期辅助：离屏渲染安装/卸载向导的每一页，输出 PNG 供人工检查布局。
// 不参与发布，只在 installer/build.ps1 之外单独编译。
using System;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Reflection;
using System.Windows.Forms;

namespace ZincSetup
{
    internal static class ShotHarness
    {
        [STAThread]
        private static void Main(string[] args)
        {
            string outDir = args.Length > 0 ? args[0] : Path.Combine(Path.GetTempPath(), "zinc-shots");
            Directory.CreateDirectory(outDir);
            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            RenderSetup(outDir);
            RenderUninstall(outDir);
            Console.WriteLine("shots in " + outDir);
        }

        private static void Shot(Form f, string path)
        {
            f.CreateControl();
            f.PerformLayout();
            Application.DoEvents();
            using (Bitmap bmp = new Bitmap(f.Width, f.Height))
            {
                f.DrawToBitmap(bmp, new Rectangle(0, 0, bmp.Width, bmp.Height));
                bmp.Save(path, ImageFormat.Png);
            }
        }

        private static void RenderSetup(string outDir)
        {
            using (SetupForm f = new SetupForm(new Options()))
            {
                f.StartPosition = FormStartPosition.Manual;
                f.Location = new Point(-4000, -4000);
                f.ShowInTaskbar = false;
                f.Opacity = 0;
                f.Show();
                Application.DoEvents();
                MethodInfo setPage = typeof(SetupForm).GetMethod("SetPage", BindingFlags.NonPublic | BindingFlags.Instance);
                for (int i = 0; i < 5; i++)
                {
                    setPage.Invoke(f, new object[] { i });
                    Application.DoEvents();
                    Shot(f, Path.Combine(outDir, "setup-" + i + ".png"));
                }
                f.Hide();
                f.Dispose();
            }
        }

        private static void RenderUninstall(string outDir)
        {
            using (ZincUninstall.UninstallForm f = new ZincUninstall.UninstallForm(false))
            {
                f.StartPosition = FormStartPosition.Manual;
                f.Location = new Point(-4000, -4000);
                f.ShowInTaskbar = false;
                f.Opacity = 0;
                f.Show();
                Application.DoEvents();
                Shot(f, Path.Combine(outDir, "uninstall.png"));
                f.Hide();
                f.Dispose();
            }
        }
    }
}
