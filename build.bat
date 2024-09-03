@echo off

cd %~dp0

FOR /F %%i in ('powershell -c "get-date -format yyMMddHHmmss"') do SET BUILD_NO=%%i
FOR /F "tokens=*" %%a in (version\VERSION) do SET VERSION=%%a.%BUILD_NO%
echo OpenNHP version: %VERSION%
FOR /F "tokens=* usebackq" %%a in (`"git show -s "--format^=%%H""`) do SET COMMIT_ID=%%a
echo commit id: %COMMIT_ID%
FOR /F "tokens=* usebackq" %%a in (`"git show -s "--format^=%%cd" "--date^=format:%%Y-%%m-%%d %%H:%%M:%%S""`) do SET COMMIT_TIME=%%a
echo commit time: %COMMIT_TIME%
FOR /F "tokens=1 usebackq" %%a in (`echo %date%`) do SET CURR_DATE=%%a
FOR /F "tokens=1 delims=. usebackq" %%a in (`echo %time%`) do SET CURR_TIME=%%a
SET BUILD_TIME=%CURR_DATE% %CURR_TIME%
echo build time: %BUILD_TIME%

set LD_FLAGS="-X 'github.com/OpenNHP/opennhp/version.Version=%VERSION%' -X 'github.com/OpenNHP/opennhp/version.CommitId=%COMMIT_ID%' -X 'github.com/OpenNHP/opennhp/version.CommitTime=%COMMIT_TIME%' -X 'github.com/OpenNHP/opennhp/version.BuildTime=%BUILD_TIME%'"

go mod tidy

:agentd
go build -trimpath -ldflags %LD_FLAGS% -v -o release\nhp-agent\nhp-agentd.exe agent\main\main.go
IF %ERRORLEVEL% NEQ 0 exit /b 1
copy /y agent\main\etc\*.* release\nhp-agent\etc

:acd
go build -trimpath -ldflags %LD_FLAGS% -v -o release\nhp-ac\nhp-acd.exe ac\main\main.go
IF %ERRORLEVEL% NEQ 0 exit /b 1
copy /y ac\main\etc\*.* release\nhp-ac\etc

:serverd
go build -trimpath -ldflags %LD_FLAGS% -v -o release\nhp-server\nhp-serverd.exe server\main\main.go
IF %ERRORLEVEL% NEQ 0 exit /b 1
copy /y server\main\etc\*.* release\nhp-server\etc

:agentsdk
go build -trimpath -buildmode=c-shared -ldflags %LD_FLAGS% -v -o release\nhp-agent\nhp-agent.dll agent\main\main.go agent\main\export.go
IF %ERRORLEVEL% NEQ 0 exit /b 1
gcc agent\sdkdemo\nhp-agent-demo.c -I release\nhp-agent -l:nhp-agent.dll -Lrelease\nhp-agent -Wl,-rpath=. -o release\nhp-agent\nhp-agent-demo.exe
IF %ERRORLEVEL% NEQ 0 exit /b 1

echo [Done] OpenNHP v%VERSION% for platform %OS% built!