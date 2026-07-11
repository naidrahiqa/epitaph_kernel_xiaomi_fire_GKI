# Changelog — Epitaph Kernel

All notable changes are documented here.
Format: [vX.Y] — YYYY-MM-DD

---

## [v330] — 2026-07-11
### Added
- NoMount file injection framework integration (kernel-based path redirection)
- Epitaph custom governors and core management tools for Redmi 12
- Epitaph cpufreq governor framework with thermal-aware frequency scaling and input boost
- Epitaph Kernel stats dashboard and data collection scripts
- epitaph_input: integrated touch boost, sched-fork launch boost, thermal management
- Thermal and charging-aware performance tuner for Helio G88
- Thermal validation suite and CPU governor tuner scripts
- Auto-run build on push + Pollux-style error dump to Telegram

### Changed
- Replace KernelSU-Next & SUSFS with xxKSU No-Mount
- Transition to ThinLTO and inject ARMv8.2-A Cortex-A75 specific compiler optimization flags
- Optimize virtual memory page settings and block I/O scheduling for eMMC 5.1
- Optimize Helio G88 EAS cpuset locking, custom schedutil boosting, dynamic energy profiles
- Update repository name references in status badges

### Fixed
- Fix script injection vulnerabilities, build failures, docs consistency
- Fix account_group_exec_runtime redefinition errors for GKI 6.6 compatibility
- Fix SELinux permission warnings on manual tuner runs
- Fix vermagic -Werror unused-function in check_modinfo
- Fix vermagic bypass for out-of-tree WiFi modules in GKI 6.6
- Fix AnyKernel3 fallback and restore pershoot/next-susfs for SUSFS
- Resolve empty GITHUB_ENV in prepare script execution
- Export repo binary path to current shell in prepare_kernel_build
- Bypass same_magic vermagic check for stock vendor modules compatibility
- Revert compiler mixing, isolate bazel and custom toolchains with clean defconfig

---

## [v148] — 2026-05-28
### Fixed
- SUSFS build: migrated KSU source to pershoot/KernelSU-Next branch dev-susfs
- SUSFS build: added git commit after staging to make changes visible to Bazel
- SUSFS_INTEGRATED flag: now verified from actual patched source, not self-written Kconfig

### Added
- Epitaph Schedutil Performance: 3-profile runtime tuner (performance/balanced/battery)
- epitaph_tuner.sh v2.0: full logging to /data/local/tmp/epitaph_tuner.log
- WiFi fallback loader: auto-insmod if systemless loading fails

### Changed
- AnyKernel3: migrated from upstream clone to own fork (pinned)
- Rescue kernel: now ships as separate build_debug_bootimg.yml workflow

---

## [v73] — 2026-05-17
### Added
- Netfilter NAT IPv4 + IPv6 for stable hotspot
- Epitaph Tuner post-boot script (v1.0)
- PStore RAMoops at 0x4d010000
- ZRAM ZSTD multi-comp & KSM memory optimizations

### Fixed
- GPU thermal throttle bug via GED bypass in tuner

---

## [v72] — 2026-05-10
### Changed
- Migrated all inline Python scripts to standalone workflow_scripts/
- GKI Control Center UI: replaced checkboxes with dropdown menus
- Fixed KernelSU-Next version injection IndentationError

### Fixed
- SUSFS patch application — mandatory validation now exits on failure
- Removed unused Azure build server support to clean up workflows

---

## [v71] — 2026-05-05
### Changed
- Experimentally compiled using ZyClang toolchain (bootloop identified due to debug info stripping)

---

## [v70] — 2026-04-28
### Added
- Initial successful boot on Android 15 HyperOS 2.0 with GKI 6.6
- Basic KernelSU-Next integration
