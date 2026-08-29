@echo off
chcp 65001 >nul
cd /d %~dp0\..\..
call scripts\windows\stop.bat
call scripts\windows\start.bat
