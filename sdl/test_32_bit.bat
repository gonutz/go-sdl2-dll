set GOOS=windows
set GOARCH=386
go test -args SDL2-2_0_10-386.dll
if errorlevel 1 pause
