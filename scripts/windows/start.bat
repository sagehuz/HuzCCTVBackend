@echo off
chcp 65001 >nul
setlocal
cd /d %~dp0\..\..
if not exist .env copy .env.example .env
if not exist dist\huzbackend.exe (
  call make build-windows
)
if exist .huzbackend.pid (
  for /f %%i in (.huzbackend.pid) do (
    tasklist /FI "PID eq %%i" 2>nul | findstr /I "%%i" >nul
    if not errorlevel 1 (
      echo Already running with PID %%i
      exit /b 0
    )
  )
)
start "Huz CCTV Server" /B dist\huzbackend.exe > .huzbackend.log 2>&1
for /f "tokens=2 delims= " %%i in ('wmic process where name^="huzbackend.exe" get ProcessId 2^>nul ^| findstr /R "[0-9]" ^| findstr /v "ProcessId"') do (
  echo %%i > .huzbackend.pid
  goto :done
)
:done
for /f "usebackq delims=" %%i in (".env") do (
  echo %%i | findstr /B "PORT=" >nul && set PORT=%%i
)
if not defined PORT set PORT=3300
echo Huz CCTV server started on http://127.0.0.1:%PORT:~5%
