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
Source: "..\rukkie.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\RukkiePulse"; Filename: "{app}\{#AppExeName}"

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
Filename: "cmd.exe"; \
  Parameters: "/C echo RukkiePulse installed. Open a new terminal and run: rukkie login"; \
  Flags: runhidden

[UninstallDelete]
Type: filesandordirs; Name: "{app}"
