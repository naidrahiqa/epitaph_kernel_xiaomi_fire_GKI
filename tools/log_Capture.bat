@echo off
:: Set path ke ADB lo
set "ADB_PATH=E:\OPREK\INSTALLER RECOVERY\Minimal ADB Fastboot\adb.exe"

if not exist "logs" mkdir "logs"

echo [TEST] Cek koneksi adb...
"%ADB_PATH%" devices
timeout /t 3 >nul

echo [TEST] Cek status root...
"%ADB_PATH%" shell "su -c id"

echo [1/4] Mencoba dmesg...
"%ADB_PATH%" shell "su -c 'dmesg'" > "logs\last_dmesg.log"

echo [2/4] Mencoba lsmod...
"%ADB_PATH%" shell "su -c 'lsmod'" > "logs\last_lsmod.txt"

echo [3/4] Mencoba meminfo...
"%ADB_PATH%" shell "cat /proc/meminfo" > "logs\last_meminfo.txt"

:: Copy buat Antigravity
copy /y "logs\last_dmesg.log" "booting\dmesg_live.log" >nul

echo.
echo [DONE] Log ditarik pake ADB dari folder E:\OPREK\...
echo Cek lagi jendela ini, ada tulisan 'uid=0(root)' nggak?
pause
