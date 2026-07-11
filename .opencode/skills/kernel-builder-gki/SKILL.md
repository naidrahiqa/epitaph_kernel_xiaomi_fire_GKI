# Kernel Builder (GKI 6.6) - Epitaph Kernel Fire

## Overview
Builds the custom GKI 6.6 kernel for Xiaomi Redmi 12 (fire, MT6769/Helio G88) targeting Android 15 HyperOS 2.0. Uses Bazel/Kleaf build system with Clang-18, ThinLTO enabled, and EAS tuned for MT6769. Validates defconfig, compiles kernel, and packages boot.img artifacts.

## Core Responsibilities
- Run Bazel/Kleaf build with `make ARCH=arm64 LLVM=1 LLVM_IAS=1 CROSS_COMPILE=aarch64-linux-gnu-`
- Validate `epitaph_fire_defconfig` for GKI 6.6 KMI compliance
- Verify ThinLTO cache (not 4.19's LTO) — check `build/thinlto-cache/` for module files
- Produce `Image.gz-dtb`, `dtbo.img`, and kernel modules
- Sign boot.img with `openssl dgst -sha256` and `signapk`

## When This Skill Activates
| Trigger | Event | Condition |
|---|---|---|
| Push | `refs/heads/main` or `fire-*` | Changes to `kernel/`, `arch/arm64/`, `Makefile` |
| PR | `opened` or `synchronize` | Paths match `build.config.*` or `Kbuild` |
| Schedule | Daily 03:00 UTC | Nightly build validation |
| Manual | `workflow_dispatch` | `build_type: release|debug` |

## Tech Stack
- **Build system**: Bazel/Kleaf, `make`, GNU Make 4.4
- **Toolchain**: Clang-18 (LLVM 18.1.8), `ld.lld-18`, `aarch64-linux-gnu-objcopy`
- **Key files**: `epitaph_fire_defconfig`, `build.config.gki.aarch64`, `arch/arm64/boot/`
- **Optimizations**: ThinLTO (NOT full LTO), `-O2`, LLVM_IAS=1

## Automated Checks
```yaml
checks:
  - id: "GBL-001"
    name: "GKI Toolchain Sanity"
    command: |
      which clang-18 && ld.lld-18 --version | grep -q "LLVM" && echo "TC_GKI_OK"
    severity: "critical"
  - id: "GBL-002"
    name: "GKI Defconfig Validation"
    command: |
      make ARCH=arm64 LLVM=1 LLVM_IAS=1 CROSS_COMPILE=aarch64-linux-gnu- epitaph_fire_defconfig
      grep -q "CONFIG_KSU=y" .config && grep -q "CONFIG_SUSFS=y" .config && echo "DEFCONFIG_GKI_OK"
    severity: "critical"
  - id: "GBL-003"
    name: "GKI Kernel Compilation"
    command: |
      make -j$(nproc) ARCH=arm64 LLVM=1 LLVM_IAS=1 CROSS_COMPILE=aarch64-linux-gnu- 2>&1 | tee build.log
      [ -f arch/arm64/boot/Image ] && echo "BUILD_GKI_OK"
    severity: "critical"
  - id: "GBL-004"
    name: "ThinLTO Cache Check"
    command: |
      [ -d build/thinlto-cache ] && MODS=$(find build/thinlto-cache -name '*.o' | wc -l)
      [ "$MODS" -ge 1 ] && echo "THINLTO_CACHE_OK ($MODS modules)" || echo "THINLTO cache empty"
    severity: "high"
```

## Input/Output Schema
```json
{
  "inputs": [
    {"name": "build_type", "type": "string", "enum": ["release", "debug"]},
    {"name": "defconfig", "type": "string", "default": "epitaph_fire_defconfig"},
    {"name": "container_image", "type": "string", "default": "ghcr.io/epitaph-kernel/toolchain:clang-18"}
  ],
  "outputs": {
    "kernel_image": "arch/arm64/boot/Image.gz-dtb",
    "dtbo_image": "arch/arm64/boot/dtbo.img",
    "modules_count": "integer",
    "build_log": "build.log"
  }
}
```

## Error Recovery
- **Bazel/Kleaf failure**: Check BUILD file syntax; verify `--config=release` flags match CI
- **ThinLTO OOM**: Reduce parallel jobs; ensure `vm.mmap_min_addr=65536` on build host
- **DTBO missing**: Verify MTK dtboimg.cfg in `arch/arm64/boot/dts/mediatek/`
- **KMI violation**: Check `android/abi_gki_aarch64.xml` for KMI symbol changes
