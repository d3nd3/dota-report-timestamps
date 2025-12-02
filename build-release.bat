@echo off
setlocal enabledelayedexpansion

set GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

echo Building release binaries for all platforms...
echo.

set DIST_DIR=dist
if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"
mkdir "%DIST_DIR%"
mkdir "%DIST_DIR%\windows"
mkdir "%DIST_DIR%\mac"
mkdir "%DIST_DIR%\linux"

echo Building Windows binaries (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o "%DIST_DIR%\windows\bot.exe" ./cmd/bot
go build -ldflags="-s -w" -o "%DIST_DIR%\windows\server.exe" ./cmd/server

echo Building Mac binaries (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o "%DIST_DIR%\mac\bot" ./cmd/bot
go build -ldflags="-s -w" -o "%DIST_DIR%\mac\server" ./cmd/server

echo Building Mac binaries (arm64)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o "%DIST_DIR%\mac\bot-arm64" ./cmd/bot
go build -ldflags="-s -w" -o "%DIST_DIR%\mac\server-arm64" ./cmd/server

echo Building Linux binaries (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o "%DIST_DIR%\linux\bot" ./cmd/bot
go build -ldflags="-s -w" -o "%DIST_DIR%\linux\server" ./cmd/server

echo.
echo Build complete! Binaries are in the dist\ directory.

