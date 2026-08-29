@echo off
chcp 65001 >nul
setlocal
cd /d %~dp0\..\..
if not exist dist\huzbackend.exe (
  echo Binary not found: dist\huzbackend.exe >&2
  exit /b 1
)
dist\huzbackend.exe restart
exit /b %errorlevel%
