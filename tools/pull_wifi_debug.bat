@echo off
:: ============================================================
:: Epitaph WiFi Debug Log Capture Script
:: Jalankan sebagai Administrator di CMD (bukan PowerShell!)
:: ============================================================
:: Pake ADB baru yang udah didownload
set "ADB=d:\Project Coding\2026\4 April\kernel redmi 12\platform-tools\adb.exe"

if not exist "logs" mkdir "logs"

echo =============================================
echo  Epitaph WiFi Debug Log Capture v2
echo =============================================
echo.

echo [1/8] Restart ADB server...
"%ADB%" kill-server
timeout /t 2 >nul
"%ADB%" start-server
timeout /t 3 >nul

echo [2/8] Cek koneksi device...
"%ADB%" devices
echo.

echo [3/8] Cek root access...
"%ADB%" shell "su -c id"
echo.

echo [4/8] Tarik FULL dmesg (kernel log)...
"%ADB%" shell "su -c dmesg" > "logs\dmesg_wifi_debug.log"
echo    Saved: logs\dmesg_wifi_debug.log

echo [5/8] Tarik lsmod (loaded modules list)...
"%ADB%" shell "su -c lsmod" > "logs\lsmod.txt"
echo    Saved: logs\lsmod.txt

echo [6/8] Tarik kernel version + vermagic info...
"%ADB%" shell "su -c 'uname -a'" > "logs\uname.txt"
"%ADB%" shell "su -c 'cat /proc/version'" >> "logs\uname.txt"
echo    Saved: logs\uname.txt

echo [7/8] Tarik WiFi module status...
(
  echo === modinfo wlan_drv_gen4m ===
  "%ADB%" shell "su -c 'modinfo /vendor/lib/modules/wlan_drv_gen4m.ko 2>&1 || echo NOT_FOUND'"
  echo.
  echo === modinfo cfg80211 ===  
  "%ADB%" shell "su -c 'modinfo /vendor/lib/modules/cfg80211.ko 2>&1 || echo NOT_FOUND'"
  echo.
  echo === modinfo mac80211 ===
  "%ADB%" shell "su -c 'modinfo /vendor/lib/modules/mac80211.ko 2>&1 || echo NOT_FOUND'"
  echo.
  echo === Vendor modules directory ===
  "%ADB%" shell "su -c 'ls -la /vendor/lib/modules/ 2>&1'"
  echo.
  echo === vendor_dlkm modules ===
  "%ADB%" shell "su -c 'ls -la /vendor_dlkm/lib/modules/ 2>&1 || echo NO_VENDOR_DLKM'"
  echo.
  echo === /data/adb/wifi_fix modules ===
  "%ADB%" shell "su -c 'ls -la /data/adb/wifi_fix/ 2>&1 || echo NO_WIFI_FIX_DIR'"
  echo.
  echo === Epitaph Tuner Log ===
  "%ADB%" shell "su -c 'cat /data/local/tmp/epitaph_tuner.log 2>&1 || echo NO_TUNER_LOG'"
  echo.
  echo === WiFi Service Status ===
  "%ADB%" shell "su -c 'getprop wlan.driver.status 2>&1'"
  "%ADB%" shell "su -c 'getprop wifi.interface 2>&1'"
  "%ADB%" shell "su -c 'getprop init.svc.wpa_supplicant 2>&1'"
  echo.
  echo === dmesg WiFi/wlan/cfg80211 specific ===
  "%ADB%" shell "su -c 'dmesg | grep -iE \"wlan|wifi|cfg80211|mac80211|vermagic|same_magic|disagrees|version magic\" 2>&1'"
) > "logs\wifi_module_debug.txt"
echo    Saved: logs\wifi_module_debug.txt

echo [8/8] Tarik dmesg errors only...
"%ADB%" shell "su -c 'dmesg | grep -iE \"error|fail|reject|disagrees|panic|killed|denied\"'" > "logs\dmesg_errors.log"
echo    Saved: logs\dmesg_errors.log

echo.
echo =============================================
echo  SELESAI! Semua log ada di folder "logs\"
echo  Bilang ke Antigravity buat lanjut analisis.
echo =============================================
pause
