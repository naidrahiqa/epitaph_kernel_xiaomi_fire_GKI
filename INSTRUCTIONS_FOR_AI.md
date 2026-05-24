# 🤖 CRITICAL INSTRUCTIONS FOR AI PAIR PROGRAMMER
> [!IMPORTANT]
> **READ THIS ENTIRE FILE BEFORE PERFORMING ANY ACTION OR INITIATING ANY KERNEL BUILD COMPILATION.**

This document contains absolute architectural rules and environment constraints for compiling the **Epitaph GKI 6.6 Kernel** for **Redmi 12 (*fire*)** running **Android 15 HyperOS 2.0**. Failure to adhere to these rules will result in bootloops and failed workflows.

---

## 🚫 Rule 1: Recovery & Debugging Constraints
* **No Custom Recovery:** GKI on Android 15 HyperOS 2.0 for Redmi 12 (*fire*) **does not have custom recovery (TWRP/OrangeFox) available**.
* **Do NOT suggest booting to TWRP:** Any suggestion to boot into custom recovery to retrieve logs or flash zips is completely invalid and incorrect.
* **Fastboot RAM Wiping:** Flashing multiple boot images via Fastboot or doing cold reboots wipes the volatile RAMoops storage (`/sys/fs/pstore/`).
* **Correct Log Retrieval Workflow:**
  1. Boot into Android using the **Rescue Kernel** (which always boots successfully).
  2. Pull the crash log of the previously failed main kernel directly from PC ADB shell:
     ```powershell
     adb shell "cat /sys/fs/pstore/console-ramoops-0" > last_kmsg.txt
     ```

---

## 🛠️ Rule 2: Clean Toolchain Isolation (No Mixing!)
* **Bazel is Bazel:** The `bazel-default` toolchain must compile strictly under the **Bazel Kleaf build system** (`tools/bazel run //common:kernel_aarch64_dist`).
* **Custom Toolchains are Custom:** Custom compilers (`weebx-latest`, `zyc-latest`, `neutron-latest`, `aosp-latest`) must compile strictly under the **classic make-based workflow** (`Build Kernel (Custom Toolchain)`) using direct `make ARCH=arm64 CC=clang`.
* **Zero Symlink Hacks:** Under **no circumstances** should you attempt to symlink, override, or inject custom compilers into Bazel's prebuilt compiler paths. Bazel and custom toolchains must remain 100% separate.

---

## 📶 Rule 3: GKI Defconfig & Bootloop Prevention
To prevent immediate device bootloops on boot, the kernel configuration (`gki_defconfig`) must strictly match the working, stable Rescue Kernel setup:

1. **Modular Connectivity Only (CRITICAL):**
   * Do **NOT** enable obsolete MediaTek built-in configs:
     `CONFIG_MTK_COMBO_WIFI=y` (WRONG)
     `CONFIG_MTK_COMBO_BT=y` (WRONG)
     `CONFIG_MTK_COMBO_GPS=y` (WRONG)
   * On GKI 6.6, MediaTek connectivity drivers must remain strictly **modular** (`CONFIG_CFG80211=m`, `CONFIG_MAC80211=m`). Forcing built-in registrations (`=y`) causes direct hardware conflicts and immediate bootloops!
2. **Minimal and Safe Defconfig:**
   * Revert all experimental memory compressions (`CONFIG_ZRAM_MULTI_COMP=y`, `CONFIG_KSM=y`) and advanced network configurations unless explicitly tested and verified.
   * Keep modules, schedutil governor, ZRAM, netfilter NAT, and PStore RAMoops at their safe default configurations.

---

## 🔒 Rule 4: Vermagic Check Bypass (Stock Modules Compatibility)
* **The WiFi/Hotspot Issue:** Precompiled proprietary vendor modules (like MediaTek's `wlan.ko` for WiFi) are compiled against the stock Xiaomi kernel version (e.g. `6.6.86-android15-8-gXXXXXX`). When compiling Epitaph from Google's GKI source (`common-android15-6.6`), the kernel version mismatch (`6.6.138` vs `6.6.86`) causes the module loader to reject these modules, breaking WiFi and Hotspot completely!
* **Bypass via Source Patching:** To allow stock Xiaomi drivers to load on custom compiled kernels, both the Main Kernel and Rescue Kernel workflows execute **`patch_vermagic.py`**.
* **Do NOT remove `patch_vermagic.py`:** This python script automatically overrides the `same_magic` helper function inside the kernel's `kernel/module/internal.h` or `kernel/module/main.c` to always return `1` (success). This completely bypasses the vermagic string mismatch checks and lets stock modular drivers load safely.

---

## 📦 File Reference
* [_build_kernel_core.yml](file:///d:/Project%20Coding/2026/4%20April/kernel%20redmi%2012/.github/workflows/_build_kernel_core.yml): Core GitHub Actions compilation workflow recipe.
* [build_manager_gki.yml](file:///d:/Project%20Coding/2026/4%20April/kernel%20redmi%2012/.github/workflows/build_manager_gki.yml): Dispatcher matrix control.
* [build_debug_bootimg.yml](file:///d:/Project%20Coding/2026/4%20April/kernel%20redmi%2012/.github/workflows/build_debug_bootimg.yml): Safe Rescue boot compiler recipe.
* [patch_vermagic.py](file:///d:/Project%20Coding/2026/4%20April/kernel%20redmi%2012/workflow_scripts/patch_vermagic.py): Python script that automatically bypasses the kernel vermagic checks.
