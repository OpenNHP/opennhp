@echo off

cd %~dp0

FOR /F %%i in ('powershell -c "get-date -format yyMMddHHmmss"') do SET BUILD_NO=%%i
FOR /F "tokens=*" %%a in (nhp\version\VERSION) do SET VERSION=%%a.%BUILD_NO%
echo OpenNHP version: %VERSION%
FOR /F "tokens=* usebackq" %%a in (`"git show -s "--format^=%%H""`) do SET COMMIT_ID=%%a
echo Commit id: %COMMIT_ID%
FOR /F "tokens=* usebackq" %%a in (`"git show -s "--format^=%%cd" "--date^=format:%%Y-%%m-%%d %%H:%%M:%%S""`) do SET COMMIT_TIME=%%a
echo Commit time: %COMMIT_TIME%
FOR /F "tokens=1 usebackq" %%a in (`echo %date%`) do SET CURR_DATE=%%a
FOR /F "tokens=1 delims=. usebackq" %%a in (`echo %time%`) do SET CURR_TIME=%%a
SET BUILD_TIME=%CURR_DATE% %CURR_TIME%
echo Build time: %BUILD_TIME%

set LD_FLAGS="-X 'github.com/OpenNHP/opennhp/nhp/version.Version=%VERSION%' -X 'github.com/OpenNHP/opennhp/nhp/version.CommitId=%COMMIT_ID%' -X 'github.com/OpenNHP/opennhp/nhp/version.CommitTime=%COMMIT_TIME%' -X 'github.com/OpenNHP/opennhp/nhp/version.BuildTime=%BUILD_TIME%'"
set CGO_ENABLED=1

cd nhp
go mod tidy
cd ../endpoints
go mod tidy

:agentd
go build -trimpath -ldflags %LD_FLAGS% -v -o ..\release\nhp-agent\nhp-agentd.exe agent\main\main.go
IF %ERRORLEVEL% NEQ 0 goto :exit
if not exist ..\release\nhp-agent\etc mkdir ..\release\nhp-agent\etc
copy agent\main\etc\*.* ..\release\nhp-agent\etc

:acd
go build -trimpath -ldflags %LD_FLAGS% -v -o ..\release\nhp-ac\nhp-acd.exe ac\main\main.go
IF %ERRORLEVEL% NEQ 0 goto :exit
if not exist ..\release\nhp-ac\etc mkdir ..\release\nhp-ac\etc
copy  ac\main\etc\*.* ..\release\nhp-ac\etc

:serverd
go build -trimpath -ldflags %LD_FLAGS% -v -o ..\release\nhp-server\nhp-serverd.exe server\main\main.go
IF %ERRORLEVEL% NEQ 0 goto :exit
if not exist ..\release\nhp-server\etc mkdir ..\release\nhp-server\etc
copy  server\main\etc\*.* ..\release\nhp-server\etc

:db
go build -trimpath -ldflags %LD_FLAGS% -v -o ..\release\nhp-db\nhp-db.exe db\main\main.go
IF %ERRORLEVEL% NEQ 0 goto :exit
if not exist ..\release\nhp-db\etc mkdir ..\release\nhp-db\etc
copy  db\main\etc\*.* ..\release\nhp-db\etc

:kgc
go build -trimpath -ldflags %LD_FLAGS% -v -o ..\release\nhp-kgc\nhp-kgc.exe kgc\main\main.go
IF %ERRORLEVEL% NEQ 0 goto :exit
if not exist ..\release\nhp-kgc\etc mkdir ..\release\nhp-kgc\etc
copy  kgc\main\etc\*.* ..\release\nhp-kgc\etc

:agentsdk
go build -trimpath -buildmode=c-shared -ldflags %LD_FLAGS% -v -o ..\release\nhp-agent\nhp-agent.dll agent\main\main.go agent\main\export.go
IF %ERRORLEVEL% NEQ 0 goto :exit
gcc agent\sdkdemo\nhp-agent-demo.c -I ..\release\nhp-agent -l:nhp-agent.dll -L..\release\nhp-agent -Wl,-rpath=. -o ..\release\nhp-agent\nhp-agent-demo.exe
IF %ERRORLEVEL% NEQ 0 goto :exit
@REM :devicesdk
@REM go build -trimpath -buildmode=c-shared -ldflags %LD_FLAGS% -v -o release\nhp-device\nhpdevice.dll core\main\main.go core\main\nhpdevice.go
@REM IF %ERRORLEVEL% NEQ 0 exit /b 1
@REM REM gcc nhp\sdkdemo\nhp-device-demo.c -I nhp\main -I release\nhp-device -l:nhpdevice.dll -Lrelease\nhp-device -Wl,-rpath=. -o release\nhp-device\nhp-device-demo.exe
@REM IF %ERRORLEVEL% NEQ 0 exit /b 1
@REM cd release\nhp-device
@REM call "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvarsall.bat" x64
@REM lib /def:./nhpdevice.def /name:nhpdevice.dll /out:./nhpdevice.lib /MACHINE:X64
@REM cd ..\..

:exit
IF %ERRORLEVEL% NEQ 0 (
    echo [Error] %ERRORLEVEL%
) ELSE (
    echo [Done] OpenNHP v%VERSION% for platform %OS% built!
)
cd ..
