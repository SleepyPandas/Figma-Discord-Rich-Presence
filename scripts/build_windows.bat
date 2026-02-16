@echo off
echo ==========================================
echo  Figma Discord Rich Presence - Build
echo ==========================================
echo.
if not defined APP_PUBLISHER set "APP_PUBLISHER=Sleepy Pandas / Anthony Hua"

:: Step 1: Build the executable
echo [1/2] Building figma-rpc.exe (no console window)...
cd /d "%~dp0..\src"
go build -ldflags "-H windowsgui" -o ..\figma-rpc.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)
echo       Done!
echo.

:: Step 2: Compile the installer
echo [2/2] Compiling installer with Inno Setup...
:: Find Inno Setup compiler
set "ISCC="
where iscc >nul 2>nul && set "ISCC=iscc"
if not defined ISCC if exist "D:\Programs\Inno Setup 6\ISCC.exe" set "ISCC=D:\Programs\Inno Setup 6\ISCC.exe"
if not defined ISCC if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" set "ISCC=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
if not defined ISCC if exist "C:\Program Files\Inno Setup 6\ISCC.exe" set "ISCC=C:\Program Files\Inno Setup 6\ISCC.exe"
if not defined ISCC (
    echo ERROR: Inno Setup not found! Install it from https://jrsoftware.org/isdl.php
    pause
    exit /b 1
)
"%ISCC%" "%~dp0installer.iss"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ==========================================
    echo  Build complete!
    echo  Installer: dist\FigmaRPC_Setup.exe
    echo ==========================================
) else (
    echo Installer compilation failed!
    pause
    exit /b 1
)
pause
