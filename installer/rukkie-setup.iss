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
SetupIconFile=favicon.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes

; Minimum Windows 10
MinVersion=10.0

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "rukkie.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "rukkie-terminal.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "favicon.ico"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu — launches bat file which handles spaces in path correctly
Name: "{group}\RukkiePulse Terminal"; \
  Filename: "{app}\rukkie-terminal.bat"; \
  IconFilename: "{app}\favicon.ico"; \
  WorkingDir: "%USERPROFILE%"; \
  Comment: "Open RukkiePulse terminal"

; Desktop shortcut
Name: "{commondesktop}\RukkiePulse Terminal"; \
  Filename: "{app}\rukkie-terminal.bat"; \
  IconFilename: "{app}\favicon.ico"; \
  WorkingDir: "%USERPROFILE%"; \
  Comment: "Open RukkiePulse terminal"

[Registry]
; System PATH (when installer is run as admin)
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
  ValueType: expandsz; ValueName: "Path"; \
  ValueData: "{olddata};{app}"; \
  Check: NeedsAddPathSystem(ExpandConstant('{app}')); \
  Flags: preservestringtype

; User PATH (fallback — works without admin, picked up by all new terminals)
Root: HKCU; Subkey: "Environment"; \
  ValueType: expandsz; ValueName: "Path"; \
  ValueData: "{olddata};{app}"; \
  Check: NeedsAddPathUser(ExpandConstant('{app}')); \
  Flags: preservestringtype

[Code]
function NeedsAddPathSystem(Param: string): boolean;
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

function NeedsAddPathUser(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;

[Run]
; Launch the bat file after install — it uses %~dp0 so spaces in path are safe
Filename: "{app}\rukkie-terminal.bat"; \
  Flags: nowait postinstall skipifsilent shellexec; \
  Description: "Open RukkiePulse terminal"

[UninstallDelete]
Type: filesandordirs; Name: "{app}"
