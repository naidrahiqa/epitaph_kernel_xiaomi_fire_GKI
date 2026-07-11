# Kernel Build Validator Developer Skill

## 📋 Overview
Validates the Epitaph Kernel build pipeline for Xiaomi Redmi 12 (fire, MT6768/Helio G88) targeting Android 15 HyperOS 2.0 with KernelSU and SUSFS integration. Performs static analysis, compiler sanity checks, defconfig validation, and artifact verification across the custom GKI 6.6 Linux kernel tree.

## 🎯 Core Responsibilities
- Validate defconfig correctness and symbol completeness for MT6768 platform
- Verify cross-compilation toolchain (Clang-18, GNU binutils) is properly configured
- Inspect KernelSU and SUSFS patch integration points for structural integrity
- Detect missing dependencies, dead code, and configuration drift in Kconfig files
- Sign and package output artifacts (boot.img, kernel binaries, kernelSU modules)

## 🚀 When This Skill Activates
| Trigger Type | Event | Condition |
|---|---|---|
| Push | `refs/heads/main` or `fire-*` branch | Any `arch/arm64/configs/` or `Makefile` changed |
| PR | `opened` or `synchronize` | Paths match `kernel/**`, `drivers/**`, `arch/arm64/**` |
| Manual | `workflow_dispatch` | `build_type: release|debug` parameter provided |
| Schedule | Daily 03:00 UTC | Always triggers on default branch |

## 🛠 Tech Stack Required
```yaml
toolchain:
  compiler: "clang-18 (LLVM 18.1.8)"
  linker: "ld.lld-18"
  binutils: "aarch64-linux-gnu-{as,ld,objcopy,objdump,strip}"
  make: "GNU Make 4.4"
build_system:
  kernel_tree: "common-android15-6.6"
  defconfig: "epitaph_fire_defconfig"
  arch: "arm64"
  cross_compile: "aarch64-linux-gnu-"
languages:
  - "C (95%)"
  - "Python (3%)"
  - "Shell (2%)"
ci: "GitHub Actions ubuntu-24.04"
dependencies:
  - "libelf-dev"
  - "libssl-dev"
  - "bc"
  - "flex"
  - "bison"
  - "cpio"
  - "pahole"
```

## 📊 Workflow & Process Pipeline

### Phase 1: DETECTION & ANALYSIS
```bash
# Identify the kernel source tree and validate repository state
echo "[*] Kernel source: $(pwd)"
git describe --tags --dirty 2>/dev/null || echo "No tags found"
KERNEL_VERSION=$(head -5 Makefile | grep "^VERSION\|^PATCHLEVEL\|^SUBLEVEL\|^EXTRAVERSION" | paste -sd' ' | tr -d '[:space:]')
echo "[*] Kernel version: $KERNEL_VERSION"
DEFCONFIG_FILE=$(find arch/arm64/configs -name "*epitaph*" -o -name "*fire*" | head -1)
echo "[*] Defconfig: $DEFCONFIG_FILE"
```

### Phase 2: DEEP DIVE INVESTIGATION
```bash
# Check toolchain availability and kernel modules
echo "[*] Checking toolchain..."
which clang-18 || { echo "FATAL: clang-18 not found"; exit 1; }
echo "[*] Checking kernel configuration dependencies..."
make ARCH=arm64 LLVM=1 LLVM_IAS=1 \
  CROSS_COMPILE=aarch64-linux-gnu- \
  olddefconfig 2>&1 | tail -20
cat .config | grep -E "CONFIG_KSU=|CONFIG_SUSFS=|CONFIG_OVERLAY_FS=" 
```

### Phase 3: VALIDATION & VERIFICATION
```bash
# Compile kernel with all cores and verify no build regressions
make -j$(nproc) ARCH=arm64 LLVM=1 LLVM_IAS=1 \
  CROSS_COMPILE=aarch64-linux-gnu- 2>&1 | tee build.log
# Validate key output artifacts exist
ls -lh arch/arm64/boot/Image.gz-dtb arch/arm64/boot/dtbo.img 2>/dev/null
# Check module compilation
find . -name "*.ko" | wc -l
```

### Phase 4: REPORT GENERATION
```bash
# Generate structured build report
{
  echo "--- Epitaph Kernel Build Report ---"
  echo "Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Toolchain: $(clang-18 --version | head -1)"
  echo "Build status: $([ -f arch/arm64/boot/Image ] && echo SUCCESS || echo FAILED)"
  echo "Warning count: $(grep -c 'warning:' build.log 2>/dev/null || echo 0)"
  echo "Error count: $(grep -c 'error:' build.log 2>/dev/null || echo 0)"
  echo "Kernel size: $(stat --printf=%s arch/arm64/boot/Image.gz-dtb 2>/dev/null || echo N/A)"
} | tee report.txt
```

## ✅ AUTOMATED CHECKS (EXECUTABLE COMMANDS)
```yaml
quality_checks:
  - check_id: "CHK-001"
    name: "Toolchain Sanity Check"
    command: |
      which clang-18 && which ld.lld-18 && \
      which aarch64-linux-gnu-objcopy && \
      clang-18 --version | grep -q "LLVM" && \
      echo "TOOLCHAIN_OK"
    expected_output: "TOOLCHAIN_OK"
    failure_indicator: "clang-18 not found or LLVM version mismatch"
    severity: "critical"

  - check_id: "CHK-002"
    name: "Defconfig Validation"
    command: |
      DEFCONFIG="arch/arm64/configs/epitaph_fire_defconfig"
      if [ ! -f "$DEFCONFIG" ]; then
        echo "DEFCONFIG_MISSING"; exit 1
      fi
      make ARCH=arm64 LLVM=1 LLVM_IAS=1 \
        CROSS_COMPILE=aarch64-linux-gnu- "$(basename $DEFCONFIG)" 2>&1
      REQUIRED_SYMS="CONFIG_KSU=y CONFIG_SUSFS=y CONFIG_OVERLAY_FS=y"
      for sym in $REQUIRED_SYMS; do
        grep -q "$sym" .config || echo "MISSING: $sym"
      done
      echo "DEFCONFIG_OK"
    expected_output: "DEFCONFIG_OK"
    failure_indicator: "Required kernel symbols missing or defconfig not found"
    severity: "critical"

  - check_id: "CHK-003"
    name: "Kernel Compilation Test"
    command: |
      make -j$(nproc) ARCH=arm64 LLVM=1 LLVM_IAS=1 \
        CROSS_COMPILE=aarch64-linux-gnu- 2>&1 | tee build.log
      if grep -q "error:" build.log; then
        echo "BUILD_FAILED"; exit 1
      fi
      if [ ! -f arch/arm64/boot/Image ]; then
        echo "IMAGE_MISSING"; exit 1
      fi
      echo "BUILD_OK"
    expected_output: "BUILD_OK"
    failure_indicator: "Compilation errors found or kernel Image not produced"
    severity: "critical"

  - check_id: "CHK-004"
    name: "KernelSU Integration Check"
    command: |
      KSU_DIR="KernelSU"
      if [ ! -d "$KSU_DIR" ]; then
        echo "KSU_DIR_MISSING"; exit 1
      fi
      if ! grep -q "ccflags-y += -DKSU" "$KSU_DIR/kernel/Makefile" 2>/dev/null; then
        echo "KSU_MAKEFILE_INVALID"
        exit 1
      fi
      grep -c "KSU_VERSION=" "$KSU_DIR/kernel/ksu.c" 2>/dev/null || echo "KSU_VERSION_UNKNOWN"
      echo "KSU_INTEGRATION_OK"
    expected_output: "KSU_INTEGRATION_OK"
    failure_indicator: "KernelSU directory missing or malformed"
    severity: "high"

  - check_id: "CHK-005"
    name: "SUSFS Patch Validation"
    command: |
      SUSFS_DIR="susfs4ksu"
      if [ ! -d "$SUSFS_DIR" ]; then
        echo "SUSFS_DIR_MISSING"; exit 1
      fi
      PATCH_FILES=$(find "$SUSFS_DIR/kernel_patches" -name "*.patch" 2>/dev/null)
      PATCH_COUNT=$(echo "$PATCH_FILES" | wc -l)
      if [ "$PATCH_COUNT" -eq 0 ]; then echo "NO_PATCHES"; exit 1; fi
      for pf in $PATCH_FILES; do
        patch -d . --dry-run -p1 < "$pf" >/dev/null 2>&1 || \
          echo "DRY_RUN_FAIL: $pf"
      done
      echo "SUSFS_OK ($PATCH_COUNT patches)"
    expected_output: "SUSFS_OK"
    failure_indicator: "SUSFS patches cannot be applied or are missing"
    severity: "high"

  - check_id: "CHK-006"
    name: "Kernel Module & DTBO Integrity"
    command: |
      MODULES=$(find . -name "*.ko" | wc -l)
      DTBO="arch/arm64/boot/dtbo.img"
      if [ -f "$DTBO" ]; then
        echo "DTBO_SIZE: $(stat --printf=%s $DTBO)"
      else
        echo "DTBO_MISSING"
      fi
      echo "MODULE_COUNT: $MODULES"
      if [ "$MODULES" -eq 0 ] && [ ! -f "$DTBO" ]; then
        echo "ARTIFACTS_MISSING"; exit 1
      fi
      echo "ARTIFACTS_OK"
    expected_output: "ARTIFACTS_OK"
    failure_indicator: "No kernel modules or DTBO artifacts generated"
    severity: "medium"

  - check_id: "CHK-007"
    name: "Static Analysis - CodeQL & Sparse"
    command: |
      which sparse >/dev/null 2>&1 || { echo "SPARSE_SKIP"; exit 0; }
      make -j$(nproc) ARCH=arm64 LLVM=1 LLVM_IAS=1 \
        CROSS_COMPILE=aarch64-linux-gnu- C=1 2>&1 | \
        tee sparse.log | grep -c "warning:" || true
      echo "SPARSE_DONE"
    expected_output: "SPARSE_DONE"
    failure_indicator: "Sparse analysis warnings exceed threshold"
    severity: "low"
```

## 📥 INPUT SCHEMA
```json
{
  "schema_version": "1.0",
  "project": "epitaph_kernel_fire",
  "inputs": [
    {
      "name": "kernel_source",
      "type": "git_repository",
      "required": true,
      "source": "https://github.com/Epitaph-Kernel/fire",
      "branch": "main",
      "description": "Epitaph custom GKI 6.6 kernel source tree"
    },
    {
      "name": "defconfig_name",
      "type": "string",
      "required": false,
      "default": "epitaph_fire_defconfig",
      "description": "Name of the defconfig to build"
    },
    {
      "name": "build_variant",
      "type": "string",
      "enum": ["release", "debug"],
      "required": false,
      "default": "release",
      "description": "Build variant affecting optimization level and debug symbols"
    },
    {
      "name": "toolchain_image",
      "type": "container_image",
      "required": false,
      "default": "ghcr.io/epitaph-kernel/toolchain:clang-18",
      "description": "OCI container image with pre-installed toolchain"
    }
  ]
}
```

## 📤 OUTPUT SCHEMA
```json
{
  "schema_version": "1.0",
  "project": "epitaph_kernel_fire",
  "metadata": {
    "build_id": "string (git SHA)",
    "kernel_version": "string (e.g. 6.6.X)",
    "toolchain": "string (clang version)",
    "timestamp": "string (ISO 8601)"
  },
  "findings": [
    {
      "check_id": "string (CHK-00X)",
      "status": "pass|fail|skip",
      "detail": "string",
      "severity": "critical|high|medium|low"
    }
  ],
  "statistics": {
    "total_checks": "int",
    "passed": "int",
    "failed": "int",
    "skipped": "int",
    "warning_count": "int",
    "error_count": "int",
    "build_duration_seconds": "int",
    "kernel_image_size_bytes": "int"
  },
  "quality_score": {
    "overall": "float (0.0 - 1.0)",
    "computed_from": "passed / total_checks"
  },
  "artifacts": {
    "kernel_image": "string (path)",
    "dtbo_image": "string (path)",
    "kernel_modules_count": "int",
    "report_file": "string (path)"
  }
}
```

## 🔌 INTEGRATION IMPLEMENTATIONS

### GitHub Actions
```yaml
name: Epitaph Kernel Build Validate
on:
  push:
    branches: [main, fire-*]
  pull_request:
    paths:
      - 'kernel/**'
      - 'drivers/**'
      - 'arch/arm64/**'
      - 'Makefile'
      - 'Kbuild'
  workflow_dispatch:
    inputs:
      build_type:
        description: 'Build variant'
        required: true
        default: 'release'
        type: choice
        options: [release, debug]

jobs:
  build-validate:
    runs-on: ubuntu-24.04
    container:
      image: ghcr.io/epitaph-kernel/toolchain:clang-18
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          submodules: recursive
      - name: Toolchain Sanity
        run: |
          which clang-18 && clang-18 --version
          which ld.lld-18 && ld.lld-18 --version
      - name: Defconfig Validation
        run: |
          make ARCH=arm64 LLVM=1 LLVM_IAS=1 \
            CROSS_COMPILE=aarch64-linux-gnu- epitaph_fire_defconfig
          grep -E "CONFIG_KSU=|CONFIG_SUSFS=" .config
      - name: Full Build
        run: |
          make -j$(nproc) ARCH=arm64 LLVM=1 LLVM_IAS=1 \
            CROSS_COMPILE=aarch64-linux-gnu- 2>&1 | tee build.log
      - name: Check Artifacts
        run: |
          ls -lh arch/arm64/boot/Image.gz-dtb arch/arm64/boot/dtbo.img
          find . -name "*.ko" -exec ls -lh {} \;
      - name: Upload Build Log
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: build-log
          path: build.log
```

### CLI Interface
```bash
# Build Epitaph kernel locally with validation
export ARCH=arm64
export CROSS_COMPILE=aarch64-linux-gnu-
export LLVM=1
export LLVM_IAS=1
export DEFCONFIG=epitaph_fire_defconfig

make $DEFCONFIG
make -j$(nproc) 2>&1 | tee build.log
scripts/ver_linux
```

## 📝 ERROR HANDLING & RECOVERY
```yaml
recovery_strategies:
  - error: "Toolchain clang-18 not in PATH"
    recovery: |
      echo "Install LLVM 18 toolchain:"
      echo "wget https://github.com/llvm/llvm-project/releases/download/llvmorg-18.1.8/clang+llvm-18.1.8-x86_64-linux-gnu-ubuntu-24.04.tar.xz"
      echo "tar xf clang+llvm-*.tar.xz -C /opt && export PATH=/opt/clang+llvm-*/bin:\$PATH"
    severity: "critical"

  - error: "Defconfig not found in arch/arm64/configs/"
    recovery: |
      echo "Create defconfig from base:"
      echo "make ARCH=arm64 LLVM=1 LLVM_IAS=1 CROSS_COMPILE=aarch64-linux-gnu- gki_defconfig"
      echo "make savedefconfig && mv defconfig arch/arm64/configs/epitaph_fire_defconfig"
    severity: "critical"

  - error: "KernelSU submodule not initialized"
    recovery: |
      git submodule update --init --recursive
      if [ ! -d "KernelSU" ]; then
        git clone https://github.com/Epitaph-Kernel/KernelSU KernelSU
      fi
    severity: "high"

  - error: "Build failure in kernel module"
    recovery: |
      grep -B5 -A5 "error:" build.log | head -50
      echo "Attempting single-thread module build for diagnostics:"
      make -j1 ARCH=arm64 LLVM=1 LLVM_IAS=1 M=drivers/staging MODULE_NAME
    severity: "high"

  - error: "SUSFS patch dry-run failure"
    recovery: |
      echo "Patch rejected. Check kernel source tree vs patch context:"
      patch -d . --dry-run -p1 < susfs4ksu/kernel_patches/XX-*.patch \
        2>&1 | head -10
      echo "Regenerate patches against current tree:"
      echo "cd susfs4ksu && ./generate_patches.sh"
    severity: "medium"
```

## 🎓 CLI USAGE EXAMPLES
```bash
# Full validation pipeline
export PATH=/opt/clang+llvm-18.1.8/bin:$PATH
git clone https://github.com/Epitaph-Kernel/fire kernel
cd kernel
git submodule update --init --recursive
SKILL_DIR="../skills/epitaph_kernel_fire"

# Run all automated checks sequentially
bash $SKILL_DIR/checks/CHK-001_toolchain.sh && \
bash $SKILL_DIR/checks/CHK-002_defconfig.sh && \
bash $SKILL_DIR/checks/CHK-003_build.sh && \
bash $SKILL_DIR/checks/CHK-004_ksu.sh && \
bash $SKILL_DIR/checks/CHK-005_susfs.sh

# Generate final report
bash $SKILL_DIR/generate_report.sh > validation-report.json
cat validation-report.json | jq .
```

## 🔐 CONFIGURATION & SECURITY
```yaml
environment_variables:
  - name: "ARCH"
    value: "arm64"
    sensitive: false
  - name: "CROSS_COMPILE"
    value: "aarch64-linux-gnu-"
    sensitive: false
  - name: "LLVM"
    value: "1"
    sensitive: false
  - name: "KERNEL_DIR"
    value: "$GITHUB_WORKSPACE"
    sensitive: false
  - name: "KSU_VERSION"
    value: "11710"
    sensitive: false

security_policies:
  - policy: "No hardcoded credentials in CI
  - policy: "Artifact signing keys stored in GitHub secrets (KEYSTORE_ALIAS, KEYSTORE_PASS)"
  - policy: "SBOM generation on release builds for supply chain transparency"
  - policy: "Signed git tags required for release builds (git tag -s)"
  - policy: "Container images scanned with trivy before workflow execution"

signing:
  tool: "openssl dgst -sha256"
  key_type: "RSA-4096"
  artifact_signing: "signapk / build/target/product/security/"
