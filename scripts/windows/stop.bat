@echo off
chcp 65001 >nul
setlocal
cd /d %~dp0\..\..
if not exist .huzbackend.pid (
  echo No PID file found
  exit /b 0
)
for /f %%i in (.huzbackend.pid) do (
  taskkill /PID %%i /F >nul 2>&1
  echo Stopped Huz CCTV server (PID %%i)
)
del /f /q .huzbackend.pid >nul 2>&1
