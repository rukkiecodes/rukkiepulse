@echo off

:: Find Git Bash
set "GITBASH="
if exist "C:\Program Files\Git\git-bash.exe"       set "GITBASH=C:\Program Files\Git\git-bash.exe"
if exist "C:\Program Files (x86)\Git\git-bash.exe" set "GITBASH=C:\Program Files (x86)\Git\git-bash.exe"

if "%GITBASH%"=="" (
  for /f "tokens=2*" %%A in ('reg query "HKLM\SOFTWARE\GitForWindows" /v InstallPath 2^>nul') do (
    if exist "%%B\git-bash.exe" set "GITBASH=%%B\git-bash.exe"
  )
)

if not "%GITBASH%"=="" (
  start "" "%GITBASH%" --login -i
  exit
)

:: Fallback: cmd stays open
echo.
echo  RukkiePulse - CLI Observability Tool
echo  Docs: https://rukkiepulse.netlify.app
echo  Run: rukkie login
echo.
cmd /K
