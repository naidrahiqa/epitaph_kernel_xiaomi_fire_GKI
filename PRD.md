# PRD — Epitaph Kernel
**Product Requirements Document**
> This document is the **source of truth** for all technical decisions in the Epitaph Kernel project.
> Anyone (human or AI) who wants to touch this repository **must read this document first**.
> If there is a conflict between this document and other instructions, this document wins.

---

## 1. Product Overview

**Epitaph Kernel** is a custom GKI 6.6 kernel designed for the **Xiaomi Redmi 12 (codename: fire)** running **Android 15 HyperOS 2.0**.

Built from Google's `common-android15-6.6` branch, it is automatically compiled via a GitHub Actions multi-toolchain pipeline, and shipped as an AnyKernel3 ZIP package to be flashed via **KernelFlasher** (no custom recovery like TWRP/OrangeFox is available or supported for this device).

### Who are the users?
- **Developer (maintainer):** Faqih Ardian Syah (@naidrahiqa) — the sole maintainer.
- **AI pair programmers:** Antigravity, Claude, Gemini, DeepSeek, Qwen — who must read this PRD as a mandatory context helper.
- **End users:** Redmi 12 owners who wish to install a highly optimized custom kernel.

### What is the value proposition?
The stock HyperOS 2.0 kernel lacks root compatibility and performance tuning. Epitaph delivers:
1. Kernel-level root via KernelSU-Next (safer and cleaner than Magisk).
2. Optional root-hiding capabilities via SUSFS (for banking/corporate apps).
3. Optimized performance: TCP BBR, BFQ I/O scheduling, custom tuned schedutil governor, ZRAM ZSTD.
4. Highly stable WiFi & Hotspot (a historical issue for GKI builds on this device).
5. Post-boot tuner (Epitaph Schedutil Performance) with 3 runtime profiles.

---

## 2. Device Context — MANDATORY TO UNDERSTAND

| Field | Value |
|---|---|
| Device | Xiaomi Redmi 12 4G |
| Codename | `fire` |
| Chipset | MediaTek Helio G88 (MT6769), 12nm |
| CPU | 2×Cortex-A75 @ 2.0GHz + 6×Cortex-A55 @ 1.8GHz |
| GPU | Mali-G52 MC2 |
| RAM | 4 / 6 / 8 GB LPDDR4x |
| Target OS | Android 15 HyperOS 2.0 **ONLY** |
| Kernel Branch | `common-android15-6.6` (always tip of branch) |
| KMI Version | android15-8 |
| Partitions | A/B seamless, Dynamic (super) |
| Page Size | 4K (mandatory, as vendor modules are compiled for 4K page size) |

### Panel Variants (CRITICAL)
This device ships with 4 different LCD panel variants:
- **LC0A / LC0B** — kernel source code is available and fully supported ✅.
- **LC0C / LC0D** — Xiaomi has NOT yet released source code (GPL violation), hence not supported ❌.

Users can identify their panel variant using: `adb shell getprop ro.boot.lcm_name`

---

## 3. System Architecture

### 3.1 Repository Structure

```
epitaph_kernel/
├── .github/workflows/
│   ├── _build_kernel_core.yml      ← Core compilation workflow recipe
│   ├── build_manager_gki.yml       ← Dispatcher matrix workflow
│   └── build_debug_bootimg.yml     ← Rescue kernel builder
├── scripts/
│   ├── prepare_kernel_build.sh     ← CI disk setup, dependencies, sync, and KSU setup
│   └── epitaph_tuner.sh            ← Post-boot performance script packaged in AnyKernel3
├── workflow_scripts/
│   ├── patch_build_system.py       ← Registers WiFi modules inside BUILD.bazel
│   ├── patch_vermagic.py           ← Bypasses vermagic for stock Xiaomi modules
│   └── patch_kbuild.py             ← Injects a static KernelSU-Next version into Kbuild
├── patches/                        ← Custom patch files (applied via patch -p1)
│   └── epitaph_schedutil.patch     ← Unlocks the schedutil rate limit minimum to 100µs
└── guidelines/                     ← Topically organized developer guidelines
```

### 3.2 CI/CD Pipeline

```
Trigger: workflow_dispatch (manual)
         └── build_manager_gki.yml
               ├── prepare: generate matrix (toolchain × SUSFS variant)
               ├── notify_start: Telegram start notification
               ├── trigger: _build_kernel_core.yml (parallel execution, max 4 runs)
               │     ├── prepare_kernel_build.sh
               │     │     ├── maximize_disk
               │     │     ├── setup_swap (16GB)
               │     │     ├── install_deps
               │     │     ├── install_repo
               │     │     ├── download_toolchain (if custom compiler is active)
               │     │     ├── sync_kernel (repo sync common-android15-6.6)
               │     │     ├── set_kmi
               │     │     ├── setup_ksu (pershoot/KernelSU-Next branch next-susfs)
               │     │     ├── apply_patches
               │     │     └── patch_build_system
               │     ├── Setup SUSFS (if with_susfs=true)
               │     ├── Configure Kernel (defconfig manipulation)
               │     ├── Build (Bazel OR Custom Clang)
               │     ├── Extract Build Output
               │     ├── Verify Build Correctness
               │     ├── Package AnyKernel3
               │     ├── Upload Artifacts
               │     ├── Create GitHub Release
               │     └── Telegram notify (success/failure)
               └── summary: final overall build status report
```

### 3.3 Toolchain Matrix

| Toolchain | Build System | Status | Notes |
|---|---|---|---|
| `bazel-default` | Bazel/Kleaf | ✅ **Production** | The only officially production-tested system |
| `aosp-latest` | make | ⚠️ Experimental | crdroidandroid prebuilt Clang |
| `zyc-latest` | make | ⚠️ Experimental | ZyClang toolchain |
| `weebx-latest` | make | ⚠️ Experimental | WeebX Clang toolchain |
| `neutron-latest` | make | ⚠️ Experimental | Neutron Clang toolchain |

**Crucial:** Bazel and custom Clang make-based compilations must be kept isolated. Do not symlink or inject custom compilers into Bazel prebuilt compiler paths.

---

## 4. Features & Project Status

### 4.1 Root & Security

| Feature | Status | Implementation Details |
|---|---|---|
| KernelSU-Next | ✅ Always Included | `pershoot/KernelSU-Next` branch `next-susfs` |
| SUSFS for KSU | ✅ Optional Variant | `simonpunk/susfs4ksu` branch `gki-android15-6.6` |
| Vermagic bypass | ✅ Always Active | `workflow_scripts/patch_vermagic.py` |

**Correct SUSFS Integration Setup:**
- KSU side: `pershoot/KernelSU-Next` branch `next-susfs` is pre-patched; `10_enable_susfs_for_ksu.patch` must be SKIPPED.
- Kernel side: `simonpunk/susfs4ksu` — manually apply `50_add_susfs_in_kernel.patch`.
- Staged commits: Always run `git commit` after adding SUSFS changes, as the Bazel sandbox only tracks files committed in HEAD.

### 4.2 Performance Tuning

| Feature | Kernel Config | Status |
|---|---|---|
| CPU Governor | `CONFIG_CPU_FREQ_GOV_SCHEDUTIL=y` | ✅ Enabled |
| TCP BBR | `CONFIG_TCP_CONG_BBR=y` + `CONFIG_NET_SCH_FQ=y` | ✅ Enabled |
| I/O BFQ | `CONFIG_IOSCHED_BFQ=y` | ✅ Enabled |
| I/O Kyber | `CONFIG_MQ_IOSCHED_KYBER=y` | ✅ Enabled |
| Timer HZ=300 | `CONFIG_HZ_300=y` | ✅ Enabled |
| WireGuard | `CONFIG_WIREGUARD=y` | ✅ Enabled |
| MGLRU | `CONFIG_LRU_GEN=y` | ✅ Enabled |
| ZRAM ZSTD | `CONFIG_CRYPTO_ZSTD=y` + `CONFIG_ZRAM_MULTI_COMP=y` | ✅ Enabled |
| PStore/RAMoops | `CONFIG_PSTORE_RAM=y` @ `0x4d010000` | ✅ Enabled |

### 4.3 Epitaph Schedutil Performance Profiles

Managed runtime options via `/data/adb/epitaph/mode`:

| Profile | up_rate | down_rate | GPU Tuning | Swappiness | Uclamp.min | Ideal Use Case |
|---|---|---|---|---|---|---|
| `performance` | 100µs | 40ms | always_on + GED boost | 200 | 180 (aggressive) | High-end Gaming |
| `balanced` | 500µs | 10ms | dynamic + GED boost | 180 | 64 (smooth UI) | Daily Driver (Default) |
| `battery` | 2ms | 1ms | coarse_demand | 160 | 0 (battery save) | Standby / Long battery life |

Apply profiles at runtime without reboots:
```sh
echo "performance" > /data/adb/epitaph/mode && sh /data/adb/epitaph/apply
```

### 4.4 WiFi & Network Fixes

| Feature | Status | Notes |
|---|---|---|
| cfg80211 + mac80211 | ✅ Modular (`=m`) | Must remain modular; registered inside BUILD.bazel |
| Netfilter NAT IPv4 | ✅ Enabled | `CONFIG_NF_NAT=y`, `CONFIG_IP_NF_TARGET_MASQUERADE=y` |
| Netfilter NAT IPv6 | ✅ Enabled | `CONFIG_IP6_NF_NAT=y`, `CONFIG_IP6_NF_TARGET_MASQUERADE=y` |
| WiFi fallback loader | ✅ Enabled | Loaded via `epitaph_tuner.sh` if systemless loader fails |

---

## 5. Known Issues & Troubleshooting

### 5.1 SUSFS Build Failures (v1–v129)
**Status:** Fully Resolved.

**Root causes:**
1. **Incorrect KSU source** — standard KernelSU dev branch lacked SUSFS hooks. Fixed by switching to `pershoot/KernelSU-Next` branch `next-susfs`.
2. **Bazel sandboxing limits** — Bazel built from HEAD, ignoring unstaged SUSFS patches. Fixed by executing `git commit` directly after staging files.
3. **Falsified `SUSFS_INTEGRATED` flag** — flag was set via self-written configuration checks rather than actual patch success. Fixed by verifying files directly (e.g., `fs/susfs.c`).

### 5.2 Flash-induced Bootloops
**Primary Causes:**
1. Unsupported LCD panels (LC0C/LC0D) lacking kernel-side display drivers.
2. Disabling debugging symbols (`CONFIG_DEBUG_INFO_NONE=y`), which crashes the BPF subsystem on Android 15.
3. Activating MediaTek combo WiFi config (`CONFIG_MTK_COMBO_WIFI=y`), leading to system crashes.
4. Using raw `Image` formatting (MediaTek bootloaders require compressed `Image.gz`).

**Emergency Recovery Procedure:**
```bash
# Step 1: Flash official stock boot image via PC CMD
fastboot flash boot boot_stock.img && fastboot reboot

# Step 2: Extract crash log
adb shell "su -c cat /sys/fs/pstore/console-ramoops-0" > last_kmsg.txt
```
*Note: Never flash multiple boots sequentially in Fastboot as it wipes out the volatile RAMoops cache.*

### 5.3 WiFi/Hotspot Failure
**Causes:**
1. `CONFIG_DEBUG_INFO_NONE=y` — BTF metadata lost, BPF for network management breaks.
2. WiFi modules not included in `module_outs` inside `BUILD.bazel` — thus not packaged.
3. Leftover modules from previous kernel remaining in `/vendor_dlkm` after flashing rescue boot via Fastboot.
4. `patch_build_system.py` fails to inject `cfg80211.ko`/`mac80211.ko`.

**Fix for leftover WiFi modules in `/vendor_dlkm`:**
1. Flash stock boot via Fastboot.
2. Flash stock boot ONCE MORE via KernelFlasher (not Fastboot) -> KernelFlasher restores original stock Xiaomi modules to `/vendor_dlkm`.
3. Flash the new Epitaph ZIP via KernelFlasher.

### 5.4 CI/CD Build Failures
**Common Causes:**
1. Bazel OOM — runner only has 7GB RAM, requires `--lto=none` + `--local_resources=memory=6144`.
2. repo sync timeout — retried 3 times, but googlesource connections occasionally drop.
3. SUSFS branch does not exist — `gki-android15-6.6` is not always present in every `susfs4ksu` version.
4. Stale Bazel cache — manually clear cache via GitHub Actions UI if suspected.

---

## 6. Technical Constraints — NON-NEGOTIABLE

Absolute constraints that must never be broken by any developer (human or AI).

### 6.1 Defconfig Configuration

| Config Parameter | Constraint | Rationale |
|---|---|---|
| `CONFIG_DEBUG_INFO_NONE` | ❌ MUST BE `=n` | Prevents BPF/BTF symbol losses which breaks WiFi on Android 15 |
| `CONFIG_MTK_COMBO_WIFI` | ❌ MUST BE `=n` | Prevents hardware combo clashes resulting in instant bootloops |
| `CONFIG_MTK_COMBO_BT` | ❌ MUST BE `=n` | Same as above |
| `CONFIG_ZSMALLOC` | ✅ MUST BE `=m` | Bazel expects it as a compiled module |
| `CONFIG_ZRAM` | ✅ MUST BE `=m` | Bazel expects it as a compiled module |
| `CONFIG_CFG80211` | ✅ MUST BE `=m` | Kept modular for systemless integration |
| `CONFIG_MAC80211` | ✅ MUST BE `=m` | Kept modular for systemless integration |
| `CONFIG_KPROBES` | ✅ MUST BE `=y` | Prerequisite for KernelSU-Next hooks |
| `CONFIG_HAVE_KPROBES` | ✅ MUST BE `=y` | Prerequisite for KernelSU-Next hooks |
| `CONFIG_KPROBE_EVENTS` | ✅ MUST BE `=y` | Prerequisite for KernelSU-Next hooks |
| `CONFIG_ARM64_4K_PAGES` | ✅ MUST BE `=y` | Mandatory as vendor drivers are compiled with 4K alignment |
| `CONFIG_MODVERSIONS` | ✅ MUST BE `=y` | Ensures Xiaomi proprietary modules load successfully |

### 6.2 Compilation Pipelines
- Always use `--lto=none` in Bazel to prevent Out-Of-Memory errors on restricted runners.
- Set `--local_resources=memory=6144` (never use deprecated `--local_ram_resources`).
- Limit parallel compilation steps to `--jobs=2`.
- Target the top of `common-android15-6.6` branch rather than pinning older commits.
- Commit all staged patch files to Git prior to running Bazel (Bazel sandbox reads HEAD only).
- Keep Bazel and custom Clang make-based compilations 100% isolated.
- Do not remove `patch_vermagic.py` — bypasses vermagic for stock Xiaomi modules.

### 6.3 AnyKernel3 Rules
- `supported.versions=15` only — GKI 6.6 is incompatible with Android 14.
- Image priority: `Image.gz` → `Image.lz4` → `Image` — MTK bootloader often rejects raw `Image`.
- `cfg80211.ko` and `mac80211.ko` must be packaged in the ZIP.

### 6.4 Recovery Rules
- No custom recovery (TWRP/OrangeFox) exists — never suggest it.
- Pull logs via PStore: `adb shell "su -c cat /sys/fs/pstore/console-ramoops-0"`.
- Rescue kernel (`build_debug_bootimg.yml`) — always boots, PStore enabled.
- Never flash multiple boot images sequentially via Fastboot — wipes RAMoops.

---

## 7. Debugging Guide

### 7.1 Decision Tree Bootloop

```
Phone bootloops after flashing Epitaph
│
├── Reboots immediately to Fastboot?
│   └── Flash stock boot → boot Android → pull last_kmsg.txt
│       └── Search for: "Kernel panic", "Call Trace", "init: Service killed"
│
├── Stuck on logo (infinite loop)?
│   └── Likely cause: display panel driver (LC0C/LC0D) or KSU init crash.
│
└── Boots but reboots immediately?
    └── Likely cause: incomplete SUSFS patching or BPF crash.
```

### 7.2 How to Pull Crash Logs
```bash
# From PC (NOT from within adb shell):
adb shell "su -c cat /sys/fs/pstore/console-ramoops-0" > last_kmsg.txt

# If the file does not exist, try:
adb shell "su -c cat /sys/fs/pstore/dmesg-ramoops-0" > last_kmsg.txt
```

### 7.3 Critical Keywords in Logs

| Keyword | Definition | Diagnostic Step |
|---|---|---|
| `Kernel panic` | Fatal crash | Read lines above to find the failing driver/function |
| `Call Trace` | Stack trace | Top-most lines show the source of the crash |
| `BUG: scheduling while atomic` | Driver race condition | Usually SUSFS hook conflict |
| `init: Service '...' killed` | Android process killed | Check `dmesg \| grep avc` for SELinux issues |
| `module verification failed` | KMI/vermagic mismatch | Check if patch_vermagic.py ran successfully |
| `KSU: ` | KernelSU init log | Ensure hooks are mounted correctly |
| `cfg80211:` | WiFi subsystem | Check module loading status |

### 7.4 Debug Build CI/CD
Add at the end of Setup SUSFS step for debugging:
```bash
echo "=== SUSFS DEBUG ==="
echo "SUSFS_PATCH_APPLIED: $SUSFS_PATCH_APPLIED"
echo "SUSFS_INTEGRATED: $SUSFS_INTEGRATED"
echo "fs/susfs.c: $([ -f fs/susfs.c ] && echo EXISTS || echo MISSING)"
echo "core_hook.c susfs lines: $(grep -c susfs drivers/kernelsu/core_hook.c 2>/dev/null || echo 0)"
echo "hooks.c susfs lines: $(grep -c susfs drivers/kernelsu/hooks.c 2>/dev/null || echo 0)"
git log --oneline -3
git status --short | head -20
```

---

## 8. Roadmap

### Sprint 1 — Fix SUSFS (URGENT - COMPLETED)
- [x] Switch KSU source to `pershoot/KernelSU-Next` branch `next-susfs` in `prepare_kernel_build.sh`
- [x] Add `git commit` after SUSFS `git add` in `_build_kernel_core.yml`
- [x] Fix `SUSFS_INTEGRATED` logic — check from source directly (expanded to core_hook.c & hooks.c)
- [x] Verification: first successful compilation of the SUSFS variant

### Sprint 2 — Defconfig Completeness (COMPLETED)
- [x] TCP BBR + FQ
- [x] HZ=300
- [x] BFQ + Kyber
- [x] WireGuard
- [x] MGLRU
- [x] ZRAM ZSTD multi-comp
- [x] Verify all configurations are correctly written in actual defconfig

### Sprint 3 — Epitaph Schedutil Performance (COMPLETED)
- [x] Kernel patch (`patches/epitaph_schedutil.patch`) — unlock rate limit to 100µs
- [x] Tuner script (`scripts/epitaph_tuner.sh`) — 3 profiles with logging
- [x] Apply script in AnyKernel3 (`_build_kernel_core.yml`)
- [x] Test performance/balanced/battery profiles on device

### Sprint 4 — Polish (COMPLETED)
- [x] Pin/Fork AnyKernel3 to dedicated repository
- [x] Auto-changelog in Release Body
- [x] Dynamic LOCALVERSION with Build Number

---

## 9. Changelog of Technical Decisions

| Date | Technical Decision | Rationale |
|---|---|---|
| v70 | Use schedutil instead of performance/powersave | Most stable for daily drivers, fully EAS-aware |
| v71 | Remove CONFIG_DEBUG_INFO_NONE | Fixed ZyClang v71 bootloops — debug info is required by Android 15 BPF |
| v72 | Migrate parsers to python scripts | Fixed heredoc inline indentation errors in YAML files |
| v72 | Drop Azure compiler toolchain | Incompatible with GKI 6.6 Android 15 toolchain specs |
| v73 | Integrate full Netfilter NAT support | Restores stable IPv4/IPv6 hotspot sharing functionality |
| v73 | Add Epitaph Tuner script | Fixes stock MediaTek GPU throttles and CPU frequency transitions |
| v129 | Identify root cause of SUSFS failures | Pinpointed incorrect KSU branch, untracked sandbox files, and false validation flags |
| 2026-05 | Switch KSU source to pershoot/next-susfs | Pre-patched SUSFS; allows skipping manual 10_enable_susfs patches |

---

## 10. Quick Reference

### Triggering Builds
```
GitHub Actions → 🎛️ GKI Control Center → Run workflow
- release_tag: v1.x
- susfs_variant: no-susfs | susfs | both
- toolchain: bazel-default | all
```

### Output File ZIP Name
```
Epitaph-{Toolchain}-kernelsu-next[-SUSFS]-{DDMMYYYY}-AnyKernel3.zip
```

### Changing Tuner Profile at Runtime (No Reboots)
```bash
echo "performance" > /data/adb/epitaph/mode
sh /data/adb/epitaph/apply
cat /data/adb/epitaph/tuner.log  # Verify status
```

### Checking Tuner Logs
```bash
adb shell "cat /data/adb/epitaph/tuner.log"
adb shell "cat /data/adb/epitaph/status"
```

### Checking Device Panel LCD Variant
```bash
adb shell getprop ro.boot.lcm_name
```

---

*This document is dynamically updated as new technical decisions are made.*
*Last updated: May 2026 — v148*