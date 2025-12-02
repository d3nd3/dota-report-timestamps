@echo off
setlocal enabledelayedexpansion

title Dota 2 Report Timestamp Tool

if not exist "bot.exe" (
    echo ERROR: bot.exe not found!
    echo Please make sure you extracted all files from the zip.
    pause
    exit /b 1
)

if not exist "server.exe" (
    echo ERROR: server.exe not found!
    echo Please make sure you extracted all files from the zip.
    pause
    exit /b 1
)

echo Starting Dota 2 Report Timestamp Tool...
echo.

set BOT_PORT=8082
start "Dota2 Bot" /min cmd /c "bot.exe"
timeout /t 2 /nobreak >nul

echo Starting server on http://localhost:8081
echo.
echo IMPORTANT: Keep this window open while using the tool!
echo Press Ctrl+C to stop the server.
echo.

start http://localhost:8081

server.exe
set SERVER_EXIT=%ERRORLEVEL%

echo.
echo Stopping bot...
for /f "tokens=2" %%a in ('tasklist /FI "WINDOWTITLE eq Dota2 Bot*" /FO LIST ^| findstr "PID"') do taskkill /PID %%a /T /F >nul 2>&1
taskkill /FI "IMAGENAME eq bot.exe" /T /F >nul 2>&1
echo Done!
exit /b %SERVER_EXIT%

