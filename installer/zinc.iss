; zinc 安装包（Inno Setup 6）
; 编译：ISCC.exe /DAppVersion=0.1.0 installer\zinc.iss
; 第一页是语言选择：英文默认，第二项简体中文。

#ifndef AppVersion
  #define AppVersion "0.1.0"
#endif

#define AppName "zinc"
#define AppPublisher "GensokyoNeoSanae"
#define AppURL "https://github.com/GensokyoNeoSanae/zinc"

[Setup]
AppId={{A3F1C0B7-5D2E-4B8C-9A41-6E7F2D90C1B4}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppURL}
AppUpdatesURL={#AppURL}/releases
VersionInfoVersion={#AppVersion}.0
VersionInfoCompany={#AppPublisher}
VersionInfoDescription={#AppName} setup
DefaultDirName={localappdata}\Programs\zinc
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
DisableDirPage=no
UsePreviousAppDir=no
DisableWelcomePage=no
LicenseFile=..\LICENSE
OutputDir=..\dist
OutputBaseFilename=zinc-{#AppVersion}-setup
SetupIconFile=..\zinc.ico
UninstallDisplayIcon={app}\zinc.ico
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
MinVersion=10.0
CloseApplications=no
RestartApplications=no
SetupLogging=yes
ShowLanguageDialog=yes
LanguageDetectionMethod=none

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "chinesesimplified"; MessagesFile: "ChineseSimplified.isl"

[CustomMessages]
english.SevenZipCaption=7-Zip
english.SevenZipDescription=zinc needs a usable 7-Zip
english.SevenZipSubCaption=zinc calls 7z.exe or 7za.exe to compress and restore. Select the executable.
english.SevenZipPrompt=Path to 7z.exe or 7za.exe:
english.SevenZipInvalid=Select a valid 7z.exe or 7za.exe. zinc needs it to pack and restore.
english.RegFailPrefix=Registering notifications and PATH failed
english.RegBusy=Registering notifications and PATH...
english.RegFailSuffix=Setup will roll back. Please report the text above.
english.Download7zip=Install 7-Zip automatically (needs internet)
english.Download7zipDir=Install to:
english.Download7zipBrowse=Browse…
english.Download7zipBrowseTitle=Select the 7-Zip folder
english.AutoFind7zip=Auto-detect
english.AutoFind7zipNotFound=No 7z.exe or 7za.exe found. Pick it manually.
english.ReadySevenZip=7-Zip:
english.ReadySevenZipAuto=7-Zip will be installed to
english.UninstallKeepSettings=Keep settings and logs (repo list, default password)? Yes keeps them, No deletes them.
english.Download7zipHint=Downloads the official installer from www.7-zip.org. Windows asks for administrator approval once.
english.Download7zipBusy=Downloading and installing 7-Zip...
english.Download7zipFail=Could not download or install 7-Zip. Setup will roll back; install 7-Zip from https://www.7-zip.org/ and run the setup again.
english.Download7zipDone=Installed 7-Zip at %1
chinesesimplified.SevenZipCaption=选择 7-Zip
chinesesimplified.SevenZipDescription=zinc 需要一个可用的 7-Zip
chinesesimplified.SevenZipSubCaption=zinc 调用 7z.exe 或 7za.exe 完成打包与恢复。请选择 7-Zip 的命令行程序。
chinesesimplified.SevenZipPrompt=7z.exe 或 7za.exe 的路径：
chinesesimplified.SevenZipInvalid=请选择一个有效的 7z.exe 或 7za.exe。zinc 需要它来打包和恢复。
chinesesimplified.RegFailPrefix=注册通知和 PATH 失败
chinesesimplified.RegBusy=正在注册通知与 PATH...
chinesesimplified.RegFailSuffix=安装将回滚，请把上面的内容反馈到项目 issue。
chinesesimplified.Download7zip=自动安装 7-Zip（需要联网）
chinesesimplified.Download7zipDir=安装到：
chinesesimplified.Download7zipBrowse=浏览…
chinesesimplified.Download7zipBrowseTitle=选择 7-Zip 安装目录
chinesesimplified.AutoFind7zip=自动查找
chinesesimplified.AutoFind7zipNotFound=没找到 7z.exe 或 7za.exe，请手动选择。
chinesesimplified.ReadySevenZip=7-Zip：
chinesesimplified.ReadySevenZipAuto=7-Zip 将安装到
chinesesimplified.UninstallKeepSettings=是否保留设置与日志（仓库列表、默认密码）？选“是”保留，选“否”一并删除。
chinesesimplified.Download7zipHint=从官网 www.7-zip.org 下载官方安装包，安装时会弹一次管理员提权。
chinesesimplified.Download7zipBusy=正在下载并安装 7-Zip...
chinesesimplified.Download7zipFail=下载或安装 7-Zip 失败，安装将回滚。请到 https://www.7-zip.org/ 手动装好 7-Zip 后重跑安装包。
chinesesimplified.Download7zipDone=已安装 7-Zip：%1

[Files]
Source: "..\zinc.exe";           DestDir: "{app}"; Flags: ignoreversion
Source: "..\zincw.exe";          DestDir: "{app}"; Flags: ignoreversion
Source: "..\zinc-register.exe";   DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\zinc-unregister.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\zinc.ico";           DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md";          DestDir: "{app}"; Flags: ignoreversion
Source: "..\README_zh.md";       DestDir: "{app}"; Flags: ignoreversion
Source: "..\LICENSE";            DestDir: "{app}"; Flags: ignoreversion

[UninstallDelete]
; 设置与日志是否删除由卸载时的确认框决定（见 CurUninstallStepChanged），这里只负责
; 收拾程序自己生成的杂项和空目录。.aumid_ok 是通知注册标记，不算用户数据，直接删。
Type: files; Name: "{app}\.aumid_ok"
Type: dirifempty; Name: "{app}\logs"
Type: dirifempty; Name: "{app}\manifests"
Type: dirifempty; Name: "{app}\tools"
[UninstallRun]
; 先让 zinc 自己把计划任务、通知注册、PATH 项清掉，再由 Inno 删文件。
Filename: "{app}\tools\zinc-unregister.exe"; Parameters: "--yes --no-pause"; Flags: runhidden waituntilterminated; RunOnceId: "ZincUnregister"

[Code]
var
  SevenZipPage: TInputFileWizardPage;
  SevenZipPath: string;
  DownloadSevenZip: Boolean;
  DownloadCheck: TNewCheckBox;
  DownloadHint: TNewStaticText;
  DownloadDirLabel: TNewStaticText;
  DownloadDirEdit: TNewEdit;
  DownloadBrowse: TNewButton;
  AutoFindBtn: TNewButton;
  KeepSettings: Boolean;

function SevenZipAt(const Dir: string): string;
var
  P: string;
begin
  Result := '';
  if Dir = '' then
    Exit;
  P := AddBackslash(Dir) + '7z.exe';
  if FileExists(P) then
  begin
    Result := P;
    Exit;
  end;
  P := AddBackslash(Dir) + '7za.exe';
  if FileExists(P) then
    Result := P;
end;

// 真正的查找：官方安装器写的注册表、App Paths、PATH、再退到常见安装目录。
function DetectSevenZip(): string;
var
  Rest, Dir, Value: string;
  P: Integer;
begin
  Result := '';

  if RegQueryStringValue(HKLM, 'SOFTWARE\7-Zip', 'Path', Value) then
  begin
    Result := SevenZipAt(Value);
    if Result <> '' then Exit;
  end;
  if RegQueryStringValue(HKLM, 'SOFTWARE\WOW6432Node\7-Zip', 'Path', Value) then
  begin
    Result := SevenZipAt(Value);
    if Result <> '' then Exit;
  end;
  if RegQueryStringValue(HKCU, 'Software\7-Zip', 'Path', Value) then
  begin
    Result := SevenZipAt(Value);
    if Result <> '' then Exit;
  end;

  if RegQueryStringValue(HKLM, 'SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\7zFM.exe', '', Value) then
  begin
    Result := SevenZipAt(ExtractFileDir(Value));
    if Result <> '' then Exit;
  end;

  Rest := GetEnv('PATH');
  while Rest <> '' do
  begin
    P := Pos(';', Rest);
    if P = 0 then
    begin
      Dir := Rest;
      Rest := '';
    end
    else
    begin
      Dir := Copy(Rest, 1, P - 1);
      Delete(Rest, 1, P);
    end;
    Result := SevenZipAt(Trim(RemoveQuotes(Dir)));
    if Result <> '' then Exit;
  end;

  Result := SevenZipAt(ExpandConstant('{pf}\7-Zip'));
  if Result <> '' then Exit;
  Result := SevenZipAt(ExpandConstant('{pf32}\7-Zip'));
  if Result <> '' then Exit;
  Result := SevenZipAt(ExpandConstant('{localappdata}\Programs\7-Zip'));
end;

procedure UpdateDownloadUi();
begin
  DownloadSevenZip := DownloadCheck.Checked;
  if DownloadSevenZip then
    DownloadHint.Font.Color := $00000000
  else
    DownloadHint.Font.Color := $00808080;
  SevenZipPage.Edits[0].Enabled := not DownloadSevenZip;
  SevenZipPage.Buttons[0].Enabled := not DownloadSevenZip;
  DownloadDirLabel.Enabled := DownloadSevenZip;
  DownloadDirEdit.Enabled := DownloadSevenZip;
  DownloadBrowse.Enabled := DownloadSevenZip;
end;

procedure DownloadCheckClick(Sender: TObject);
begin
  UpdateDownloadUi();
end;

procedure AutoFindClick(Sender: TObject);
var
  Found: string;
begin
  Found := DetectSevenZip();
  if Found = '' then
    MsgBox(CustomMessage('AutoFind7zipNotFound'), mbInformation, MB_OK)
  else
    SevenZipPage.Values[0] := Found;
end;

procedure DownloadBrowseClick(Sender: TObject);
var
  Dir: string;
begin
  Dir := DownloadDirEdit.Text;
  if BrowseForFolder(CustomMessage('Download7zipBrowseTitle'), Dir, True) then
    DownloadDirEdit.Text := Dir;
end;

procedure InitializeWizard();
begin
  SevenZipPage := CreateInputFilePage(wpSelectDir,
    CustomMessage('SevenZipCaption'),
    '',
    CustomMessage('SevenZipSubCaption'));
  SevenZipPage.Add(CustomMessage('SevenZipPrompt'),
    'Executable files (*.exe)|*.exe|All files (*.*)|*.*', '.exe');

  DownloadCheck := TNewCheckBox.Create(WizardForm);
  DownloadCheck.Parent := SevenZipPage.Surface;
  DownloadCheck.Left := 0;
  DownloadCheck.Top := SevenZipPage.Surface.Height - 118;
  DownloadCheck.Width := SevenZipPage.Surface.Width;
  DownloadCheck.Caption := CustomMessage('Download7zip');
  DownloadCheck.OnClick := @DownloadCheckClick;

  DownloadDirLabel := TNewStaticText.Create(WizardForm);
  DownloadDirLabel.Parent := SevenZipPage.Surface;
  DownloadDirLabel.Left := 18;
  DownloadDirLabel.Top := SevenZipPage.Surface.Height - 82;
  DownloadDirLabel.Caption := CustomMessage('Download7zipDir');

  DownloadDirEdit := TNewEdit.Create(WizardForm);
  DownloadDirEdit.Parent := SevenZipPage.Surface;
  DownloadDirEdit.Left := 18 + DownloadDirLabel.Width + 6;
  DownloadDirEdit.Top := SevenZipPage.Surface.Height - 84;
  DownloadDirEdit.Width := SevenZipPage.Surface.Width - DownloadDirEdit.Left - 92;
  DownloadDirEdit.Text := ExpandConstant('{sd}\Program Files\7-Zip');

  DownloadBrowse := TNewButton.Create(WizardForm);
  DownloadBrowse.Parent := SevenZipPage.Surface;
  DownloadBrowse.Width := 80;
  DownloadBrowse.Left := SevenZipPage.Surface.Width - DownloadBrowse.Width;
  DownloadBrowse.Top := DownloadDirEdit.Top - 1;
  DownloadBrowse.Caption := CustomMessage('Download7zipBrowse');
  DownloadBrowse.OnClick := @DownloadBrowseClick;

  DownloadHint := TNewStaticText.Create(WizardForm);
  DownloadHint.Parent := SevenZipPage.Surface;
  DownloadHint.Left := 18;
  DownloadHint.Top := SevenZipPage.Surface.Height - 50;
  DownloadHint.Width := SevenZipPage.Surface.Width - 18;
  DownloadHint.AutoSize := False;
  DownloadHint.WordWrap := True;
  DownloadHint.Caption := CustomMessage('Download7zipHint');

  SevenZipPath := DetectSevenZip();
  if SevenZipPath <> '' then
    SevenZipPage.Values[0] := SevenZipPath;

  AutoFindBtn := TNewButton.Create(WizardForm);
  AutoFindBtn.Parent := SevenZipPage.Surface;
  AutoFindBtn.Width := 96;
  AutoFindBtn.Left := SevenZipPage.Buttons[0].Left - AutoFindBtn.Width - 6;
  AutoFindBtn.Top := SevenZipPage.Buttons[0].Top;
  AutoFindBtn.Caption := CustomMessage('AutoFind7zip');
  AutoFindBtn.OnClick := @AutoFindClick;
  SevenZipPage.Edits[0].Width := AutoFindBtn.Left - SevenZipPage.Edits[0].Left - 6;

  DownloadCheck.Checked := False;
  UpdateDownloadUi();
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = SevenZipPage.ID then
  begin
    if DownloadSevenZip then
      Exit;
    SevenZipPath := Trim(SevenZipPage.Values[0]);
    if not FileExists(SevenZipPath) then
    begin
      MsgBox(CustomMessage('SevenZipInvalid'), mbError, MB_OK);
      Result := False;
    end;
  end;
end;

function FindInstalledSevenZip(const TargetDir: string): string;
var
  Cand: string;
begin
  Result := '';
  Cand := AddBackslash(TargetDir) + '7z.exe';
  if FileExists(Cand) then begin Result := Cand; Exit; end;
  Cand := AddBackslash(TargetDir) + '7za.exe';
  if FileExists(Cand) then begin Result := Cand; Exit; end;
  Cand := ExpandConstant('{sd}\Program Files\7-Zip\7z.exe');
  if FileExists(Cand) then begin Result := Cand; Exit; end;
  Cand := ExpandConstant('{sd}\Program Files (x86)\7-Zip\7z.exe');
  if FileExists(Cand) then begin Result := Cand; Exit; end;
  Cand := ExpandConstant('{pf}\7-Zip\7z.exe');
  if FileExists(Cand) then begin Result := Cand; Exit; end;
  Cand := ExpandConstant('{pf32}\7-Zip\7z.exe');
  if FileExists(Cand) then begin Result := Cand; Exit; end;
end;

function DownloadAndInstallSevenZip(const TargetDir: string): string;
var
  Ps1, Script, SetupExe: string;
  Rc: Integer;
begin
  Result := '';
  // 目标目录里已经有 7z.exe 就不要再下载/安装一遍
  if FileExists(AddBackslash(TargetDir) + '7z.exe') then
  begin
    Result := AddBackslash(TargetDir) + '7z.exe';
    Exit;
  end;
  SetupExe := ExpandConstant('{tmp}\7z-setup.exe');
  Ps1 := ExpandConstant('{tmp}\get-7zip.ps1');
  Script :=
    '$ProgressPreference = ''SilentlyContinue''' + #13#10 +
    '[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12' + #13#10 +
    'try {' + #13#10 +
    '  $h = (Invoke-WebRequest -UseBasicParsing ''https://www.7-zip.org/download.html'').Content' + #13#10 +
    '  $m = [regex]::Match($h, ''a/(7z\d+-x64\.exe)'')' + #13#10 +
    '  if (-not $m.Success) { exit 2 }' + #13#10 +
    '  Invoke-WebRequest -UseBasicParsing (''https://www.7-zip.org/a/'' + $m.Groups[1].Value) -OutFile ''' + SetupExe + '''' + #13#10 +
    '} catch { exit 3 }' + #13#10;
  if not SaveStringToFile(Ps1, Script, False) then
    Exit;
  if not Exec('powershell.exe', '-NoProfile -ExecutionPolicy Bypass -File "' + Ps1 + '"',
      '', SW_HIDE, ewWaitUntilTerminated, Rc) then
    Exit;
  if (Rc <> 0) or (not FileExists(SetupExe)) then
    Exit;
  if not ShellExec('runas', SetupExe, '/S /D=' + TargetDir, '', SW_SHOW, ewWaitUntilTerminated, Rc) then
    Exit;
  Sleep(1500);
  Result := FindInstalledSevenZip(TargetDir);
end;

function UpdateReadyMemo(Space, NewLine, MemoUserInfoInfo, MemoDirInfo, MemoTypeInfo,
  MemoComponentsInfo, MemoGroupInfo, MemoTasksInfo: String): String;
var
  ZipLine: string;
begin
  if DownloadSevenZip then
    ZipLine := CustomMessage('ReadySevenZipAuto') + ' ' + Trim(DownloadDirEdit.Text)
  else
    ZipLine := CustomMessage('ReadySevenZip') + ' ' + SevenZipPath;
  Result := MemoDirInfo + NewLine + ZipLine + NewLine + MemoGroupInfo + MemoTasksInfo;
end;

function RunZincInstall(const SevenZip: string; var LogText: string): Integer;
var
  CmdFile, LogFile, Script: string;
  AnsiLog: AnsiString;
  Rc: Integer;
begin
  CmdFile := ExpandConstant('{tmp}\register.cmd');
  LogFile := ExpandConstant('{tmp}\register.log');
  // --force：每次安装都重建通知注册（开始菜单快捷方式 + AppUserModelId），
  // 免得重装到旧目录时沿用上一次装的快捷方式（旧版本目标还指错到 tools\ 下）。
  Script := '@echo off' + #13#10
    + '"' + ExpandConstant('{app}\tools\zinc-register.exe') + '" --7zip "' + SevenZip + '" --no-pause --lang en --force > "' + LogFile + '" 2>&1' + #13#10
    + 'exit /b %ERRORLEVEL%' + #13#10;
  if not SaveStringToFile(CmdFile, Script, False) then
  begin
    Result := -1;
    LogText := 'cannot write ' + CmdFile;
    Exit;
  end;
  if not Exec('cmd.exe', '/c "' + CmdFile + '"', '', SW_HIDE, ewWaitUntilTerminated, Rc) then
  begin
    Result := -1;
    LogText := 'cannot start cmd.exe';
    Exit;
  end;
  LogText := '';
  LoadStringFromFile(LogFile, AnsiLog);
  LogText := AnsiLog;
  Result := Rc;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep <> usUninstall then
    Exit;
  // 问在 Inno 自己的「确定要完全卸载吗」之后：先确认要卸，再问要不要留设置。
  // 放在 InitializeUninstall 会把顺序反过来，看着很别扭。
  KeepSettings := False;
  if not UninstallSilent then
    KeepSettings := MsgBox(CustomMessage('UninstallKeepSettings'), mbConfirmation, MB_YESNO) = IDYES;
  if KeepSettings then
    Exit;
  DeleteFile(ExpandConstant('{app}\settings.json'));
  DeleteFile(ExpandConstant('{app}\repos.json'));
  DeleteFile(ExpandConstant('{app}\zinc.log'));
  DelTree(ExpandConstant('{app}\logs'), True, True, True);
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Rc, FailCode: Integer;
  LogText, ZipDir: string;
begin
  if CurStep <> ssPostInstall then
    Exit;
  if ExpandConstant('{param:NOREG|0}') = '1' then
    Exit;

  if DownloadSevenZip then
  begin
    WizardForm.StatusLabel.Caption := CustomMessage('Download7zipBusy');
    ZipDir := Trim(DownloadDirEdit.Text);
    if ZipDir = '' then
      ZipDir := ExpandConstant('{sd}\Program Files\7-Zip');
    SevenZipPath := DownloadAndInstallSevenZip(ZipDir);
    if SevenZipPath = '' then
      RaiseException(CustomMessage('Download7zipFail'));
  end;

  WizardForm.StatusLabel.Caption := CustomMessage('RegBusy');
  Rc := RunZincInstall(SevenZipPath, LogText);
  if Rc <> 0 then
  begin
    FailCode := Rc;
    // 注册失败就当作安装失败：先撤掉可能写了一半的注册，再让 Inno 回滚文件。
    Exec(ExpandConstant('{app}\tools\zinc-unregister.exe'), '--yes --no-pause', '', SW_HIDE, ewWaitUntilTerminated, Rc);
    RaiseException(CustomMessage('RegFailPrefix') + ' (exit code ' + IntToStr(FailCode) + ').'
      + #13#10 + LogText + #13#10 + CustomMessage('RegFailSuffix'));
  end;
end;
