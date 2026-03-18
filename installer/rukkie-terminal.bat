@echo off

:: Prefer Git Bash — check common install locations
set "GITBASH="
if exist "C:\Program Files\Git\git-bash.exe"     set "GITBASH=C:\Program Files\Git\git-bash.exe"
if exist "C:\Program Files (x86)\Git\git-bash.exe" set "GITBASH=C:\Program Files (x86)\Git\git-bash.exe"

:: Try registry location for custom Git installs
if "%GITBASH%"=="" (
  for /f "tokens=2*" %%A in ('reg query "HKLM\SOFTWARE\GitForWindows" /v InstallPath 2^>nul') do (
    if exist "%%B\git-bash.exe" set "GITBASH=%%B\git-bash.exe"
  )
)

if not "%GITBASH%"=="" (
  :: Launch Git Bash and show welcome message on open
  start "" "%GITBASH%" --login -i -c "echo; echo '  RukkiePulse — CLI Observability Tool'; echo '  Docs: https://rukkiepulse.netlify.app'; echo; rukkie; exec bash"
  exit
)

:: Fallback: plain cmd with colors via ANSI
echo.
echo  RukkiePulse - CLI Observability Tool
echo  Docs: https://rukkiepulse.netlify.app
echo.
"%~dp0rukkie.exe"
echo.
cmd /K
