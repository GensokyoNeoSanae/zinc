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
english.RegFailPrefix=Files are installed, but registering notifications and PATH failed
english.RegFailSuffix=Run zinc-install.exe from the install folder to retry.
chinesesimplified.SevenZipCaption=选择 7-Zip
chinesesimplified.SevenZipDescription=zinc 需要一个可用的 7-Zip
chinesesimplified.SevenZipSubCaption=zinc 调用 7z.exe 或 7za.exe 完成打包与恢复。请选择 7-Zip 的命令行程序。
chinesesimplified.SevenZipPrompt=7z.exe 或 7za.exe 的路径：
chinesesimplified.SevenZipInvalid=请选择一个有效的 7z.exe 或 7za.exe。zinc 需要它来打包和恢复。
chinesesimplified.RegFailPrefix=文件已安装，但注册通知和 PATH 失败
chinesesimplified.RegFailSuffix=请到安装目录手动运行 zinc-install.exe 重试。

[Files]
Source: "..\zinc.exe";           DestDir: "{app}"; Flags: ignoreversion
Source: "..\zincw.exe";          DestDir: "{app}"; Flags: ignoreversion
Source: "..\zinc-install.exe";   DestDir: "{app}"; Flags: ignoreversion
Source: "..\zinc-uninstall.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\zinc.ico";           DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md";          DestDir: "{app}"; Flags: ignoreversion
Source: "..\README_zh.md";       DestDir: "{app}"; Flags: ignoreversion
Source: "..\LICENSE";            DestDir: "{app}"; Flags: ignoreversion

[UninstallRun]
; 先让 zinc 自己把计划任务、通知注册、PATH 项清掉，再由 Inno 删文件。
Filename: "{app}\zinc-uninstall.exe"; Parameters: "--yes --no-pause"; Flags: runhidden waituntilterminated; RunOnceId: "ZincUnregister"

[Code]
var
  SevenZipPage: TInputFileWizardPage;
  SevenZipPath: string;

function DetectSevenZip(): string;
var
  Names: array[0..1] of string;
  Rest, Dir, Candidate: string;
  P, I: Integer;
begin
  Result := '';
  Names[0] := '7za.exe';
  Names[1] := '7z.exe';

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
    Dir := Trim(RemoveQuotes(Dir));
    if Dir <> '' then
      for I := 0 to 1 do
      begin
        Candidate := AddBackslash(Dir) + Names[I];
        if FileExists(Candidate) then
        begin
          Result := Candidate;
          Exit;
        end;
      end;
  end;

  for I := 0 to 1 do
  begin
    Candidate := ExpandConstant('{pf}\7-Zip\') + Names[I];
    if FileExists(Candidate) then
    begin
      Result := Candidate;
      Exit;
    end;
    Candidate := ExpandConstant('{pf32}\7-Zip\') + Names[I];
    if FileExists(Candidate) then
    begin
      Result := Candidate;
      Exit;
    end;
  end;

  Candidate := 'E:\Tools\7zip\7za.exe';
  if FileExists(Candidate) then
    Result := Candidate;
end;

procedure InitializeWizard();
begin
  SevenZipPage := CreateInputFilePage(wpSelectDir,
    CustomMessage('SevenZipCaption'),
    CustomMessage('SevenZipDescription'),
    CustomMessage('SevenZipSubCaption'));
  SevenZipPage.Add(CustomMessage('SevenZipPrompt'),
    'Executable files (*.exe)|*.exe|All files (*.*)|*.*', '.exe');
  SevenZipPath := DetectSevenZip();
  if SevenZipPath <> '' then
    SevenZipPage.Values[0] := SevenZipPath;
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = SevenZipPage.ID then
  begin
    SevenZipPath := Trim(SevenZipPage.Values[0]);
    if not FileExists(SevenZipPath) then
    begin
      MsgBox(CustomMessage('SevenZipInvalid'), mbError, MB_OK);
      Result := False;
    end;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    if ExpandConstant('{param:NOREG|0}') = '1' then
      Exit;
    if (not Exec(ExpandConstant('{app}\zinc-install.exe'),
        '--7zip "' + SevenZipPath + '" --no-pause', '',
        SW_HIDE, ewWaitUntilTerminated, ResultCode)) or (ResultCode <> 0) then
      MsgBox(CustomMessage('RegFailPrefix') + ' (exit code ' + IntToStr(ResultCode) + ').'
        + #13#10 + CustomMessage('RegFailSuffix'), mbError, MB_OK);
  end;
end;
