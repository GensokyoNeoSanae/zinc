package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AUMID：Windows 通知用它识别「这是哪个应用发的」
const appID = "zinc.backup"

// shortcutHelper 用 COM 给 .lnk 快捷方式设置 System.AppUserModel.ID。
// 未打包的 exe 必须有带 AUMID 的开始菜单快捷方式，WinRT Toast 才能工作。
const shortcutHelper = `using System;
using System.Runtime.InteropServices;
using System.Text;

[ComImport, Guid("00021401-0000-0000-C000-000000000046")]
internal class ShellLink { }

[ComImport, InterfaceType(ComInterfaceType.InterfaceIsIUnknown), Guid("000214F9-0000-0000-C000-000000000046")]
internal interface IShellLinkW
{
    void GetPath([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder pszFile, int cch, IntPtr pfd, uint fFlags);
    void GetIDList(out IntPtr ppidl);
    void SetIDList(IntPtr pidl);
    void GetDescription([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder pszName, int cch);
    void SetDescription([MarshalAs(UnmanagedType.LPWStr)] string pszName);
    void GetWorkingDirectory([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder pszDir, int cch);
    void SetWorkingDirectory([MarshalAs(UnmanagedType.LPWStr)] string pszDir);
    void GetArguments([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder pszArgs, int cch);
    void SetArguments([MarshalAs(UnmanagedType.LPWStr)] string pszArgs);
    void GetHotkey(out short pwHotkey);
    void SetHotkey(short wHotkey);
    void GetShowCmd(out int piShowCmd);
    void SetShowCmd(int iShowCmd);
    void GetIconLocation([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder pszIconPath, int cch, out int piIcon);
    void SetIconLocation([MarshalAs(UnmanagedType.LPWStr)] string pszIconPath, int iIcon);
    void SetRelativePath([MarshalAs(UnmanagedType.LPWStr)] string pszPathRel, uint dwReserved);
    void Resolve(IntPtr hwnd, uint fFlags);
    void SetPath([MarshalAs(UnmanagedType.LPWStr)] string pszFile);
}

[ComImport, InterfaceType(ComInterfaceType.InterfaceIsIUnknown), Guid("886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99")]
internal interface IPropertyStore
{
    void GetCount(out uint cProps);
    void GetAt(uint iProp, out PROPERTYKEY pkey);
    void GetValue(ref PROPERTYKEY key, out PROPVARIANT pv);
    void SetValue(ref PROPERTYKEY key, ref PROPVARIANT pv);
    void Commit();
}

[StructLayout(LayoutKind.Sequential, Pack = 4)]
internal struct PROPERTYKEY
{
    public Guid fmtid;
    public uint pid;
}

[StructLayout(LayoutKind.Explicit)]
internal struct PROPVARIANT
{
    [FieldOffset(0)] public ushort vt;
    [FieldOffset(8)] public IntPtr pointerValue;
}

public static class ZincShortcut
{
    public static void SetAppId(string shortcutPath, string appId)
    {
        var link = (IShellLinkW)new ShellLink();
        ((System.Runtime.InteropServices.ComTypes.IPersistFile)link).Load(shortcutPath, 2); // STGM_READWRITE
        var store = (IPropertyStore)link;
        var key = new PROPERTYKEY { fmtid = new Guid("9F4C2855-9F79-4B39-A8D0-E1D42DE1D5F3"), pid = 5 };
        var pv = new PROPVARIANT { vt = 31, pointerValue = Marshal.StringToCoTaskMemUni(appId) };
        store.SetValue(ref key, ref pv);
        store.Commit();
        ((System.Runtime.InteropServices.ComTypes.IPersistFile)link).Save(shortcutPath, true);
    }
}`

func shortcutPath() string {
	return filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "zinc.lnk")
}

func psQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// ensureShortcut 首次发通知前创建带 AUMID 的开始菜单快捷方式。
func ensureShortcut() error { return ensureShortcutForce(false) }

// ensureShortcutForce 注册 AUMID（快捷方式 + 注册表 + .aumid_ok）。
// force 为真时即使三样齐全也重建一次（换图标、修坏通知时用）。
func ensureShortcutForce(force bool) error {
	lnk := shortcutPath()
	marker := filepath.Join(appDir(), ".aumid_ok")
	if !force && fileExists(lnk) && fileExists(marker) {
		return nil
	}
	if force {
		os.Remove(marker) // 重建失败时别留下“已注册”的假标记
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// 先写到同目录的临时 .lnk，成功后再替换正式文件；失败不会碰已有的 zinc.lnk。
	tmp := filepath.Join(filepath.Dir(lnk), fmt.Sprintf("zinc-%d.tmp.lnk", os.Getpid()))
	os.Remove(tmp)
	// 优先指向同目录的 zinc.exe，双击快捷方式能看到用法
	if zi := filepath.Join(filepath.Dir(exe), "zinc.exe"); fileExists(zi) {
		exe = zi
	}
	// 同目录有 zinc.ico 就用它做通知图标
	iconLine := ""
	if ico := filepath.Join(filepath.Dir(exe), "zinc.ico"); fileExists(ico) {
		iconLine = fmt.Sprintf("$sc.IconLocation = '%s,0'", psQuote(ico))
	}
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$code = @'
%s
'@
Add-Type -TypeDefinition $code
$lnk = '%s'
$ws = New-Object -ComObject WScript.Shell
$sc = $ws.CreateShortcut($lnk)
$sc.TargetPath = '%s'
%s
$sc.Save()
[ZincShortcut]::SetAppId($lnk, '%s')
$key = 'HKCU:\Software\Classes\AppUserModelId\%s'
New-Item -Path $key -Force | Out-Null
Set-ItemProperty -Path $key -Name 'DisplayName' -Value 'zinc'
Set-ItemProperty -Path $key -Name 'ShowInSettings' -Value 1 -Type DWord
`, shortcutHelper, psQuote(tmp), psQuote(exe), iconLine, appID, appID)

	out, err := exec.Command(powershellPath(), "-NoProfile", "-Command", script).CombinedOutput()
	if err != nil {
		os.Remove(tmp) // 只清理临时文件，保留已有的 zinc.lnk
		return fmt.Errorf(M.RegisterFailed, err, string(out))
	}
	if err := os.Rename(tmp, lnk); err != nil {
		os.Remove(tmp)
		return fmt.Errorf(M.RegisterFailed, err, "rename temp shortcut failed")
	}
	os.WriteFile(marker, []byte("ok"), 0644)
	return nil
}

// sendToast 用 WinRT Toast 发通知，标题显示为 zinc。
func sendToast(title, body string) error {
	if err := ensureShortcut(); err != nil {
		logf("shortcut: %v", err)
	}

	esc := func(s string) string {
		r := strings.NewReplacer(
			"&", "&amp;",
			"<", "&lt;",
			">", "&gt;",
			`"`, "&quot;",
			"'", "&apos;",
		)
		return r.Replace(s)
	}

	var texts strings.Builder
	texts.WriteString("<text>" + esc(title) + "</text>")
	for _, line := range strings.Split(body, "\n") {
		texts.WriteString("<text>" + esc(line) + "</text>")
	}
	xml := `<toast><visual><binding template="ToastGeneric">` + texts.String() + `</binding></visual></toast>`

	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$doc = New-Object Windows.Data.Xml.Dom.XmlDocument
$doc.LoadXml('%s')
$toast = New-Object Windows.UI.Notifications.ToastNotification $doc
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('%s').Show($toast)
`, psQuote(xml), appID)

	out, err := exec.Command(powershellPath(), "-NoProfile", "-Command", script).CombinedOutput()
	if err != nil {
		logf("toast failed: %v %s", err, string(out))
		return fmt.Errorf(M.ToastFailed, err)
	}
	return nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
