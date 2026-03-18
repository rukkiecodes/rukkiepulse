; RukkiePulse Windows Installer Script
; Build with Inno Setup: https://jrsoftware.org/isinfo.php

#define AppName "RukkiePulse"
#define AppVersion "1.0.0"
#define AppPublisher "Rukkiecodes"
#define AppURL "https://github.com/rukkiecodes/rukkiepulse"
#define AppExeName "rukkie.exe"

[Setup]
AppId={{B3F2A1E4-7C8D-4F5A-9B2E-1D3C6E8F0A2B}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppURL}
DefaultDirName={autopf}\RukkiePulse
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
OutputDir=output
OutputBaseFilename=rukkie-setup
SetupIconFile=
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes

; Minimum Windows 10
MinVersion=10.0

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
; The rukkie.exe binary — build it first with: go build -o rukkie.exe ./cmd/rukkie/
Source: "rukkie.exe"; DestDir: "{app}"; Flags: ignoreversion
; Launcher batch file — opens a cmd window that stays open
Source: "rukkie-terminal.bat"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu shortcut — opens cmd.exe using full path, keeps window open
Name: "{group}\RukkiePulse Terminal"; \
  Filename: "{sys}\cmd.exe"; \
  Parameters: "/K ""{app}\rukkie.exe"""; \
  WorkingDir: "%USERPROFILE%"; \
  Comment: "Open RukkiePulse terminal"

; Desktop shortcut
Name: "{commondesktop}\RukkiePulse Terminal"; \
  Filename: "{sys}\cmd.exe"; \
  Parameters: "/K ""{app}\rukkie.exe"""; \
  WorkingDir: "%USERPROFILE%"; \
  Comment: "Open RukkiePulse terminal"

[Registry]
; Add install directory to system PATH
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
  ValueType: expandsz; ValueName: "Path"; \
  ValueData: "{olddata};{app}"; \
  Check: NeedsAddPath(ExpandConstant('{app}'))

[Code]
function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(
    HKEY_LOCAL_MACHINE,
    'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
    'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;

[Run]
; After install — open a cmd window showing the welcome message, stays open
Filename: "{sys}\cmd.exe"; \
  Parameters: "/K ""echo. && echo  RukkiePulse installed successfully! && echo. && echo  Run: rukkie login && echo  Docs: https://rukkiepulse.netlify.app && echo. && {app}\rukkie.exe"""; \
  Flags: nowait postinstall skipifsilent; \
  Description: "Open RukkiePulse terminal"

[UninstallDelete]
Type: filesandordirs; Name: "{app}"
