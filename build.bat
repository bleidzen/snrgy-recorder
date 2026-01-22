@echo off
echo Building SNRGY Recorder...

:: Download dependencies
go mod tidy

:: Build for Windows (no console window)
echo Building Windows executable...
go build -ldflags="-H windowsgui -s -w" -o snrgy-recorder.exe .

echo Done! Output: snrgy-recorder.exe
pause
