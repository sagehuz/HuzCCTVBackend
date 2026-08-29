@echo off
chcp 65001 >nul
setlocal
cd /d %~dp0\..\..
if not exist dist\huzbackend.exe (
  echo Binary not found; building...
  call make build-windows
)
if not exist dist\.env (
  if exist .env (
    copy /y .env dist\.env >nul
  ) else if exist .env.example (
    copy /y .env.example dist\.env >nul
  )
)
dist\huzbackend.exe start
exit /b %errorlevel%
