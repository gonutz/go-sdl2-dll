set GOOS=windows
set GOARCH=amd64
go test -args SDL2-2_0_10-amd64.dll
if errorlevel 1 pause
