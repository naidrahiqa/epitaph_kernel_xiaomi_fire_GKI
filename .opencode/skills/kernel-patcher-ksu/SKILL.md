# Kernel Patcher (KSU/SUSFS) - Epitaph Kernel Fire

## Overview
Integrates KernelSU (KSU) and SUSFS into the GKI 6.6 kernel tree. Unlike 4.19's approach, GKI 6.6 uses KMI modules and different patch injection points. Manages `KernelSU/` submodule, SUSFS `kernel_patches/6.6/`, and ensures `CONFIG_KSU=y` and `CONFIG_SUSFS=y` in defconfig.

## Core Responsibilities
- Initialize and validate `KernelSU/` submodule (`git submodule update --init`)
- Apply SUSFS patches from `susfs4ksu/kernel_patches/6.6/` to GKI 6.6 tree
- Configure `CONFIG_KSU=y` via defconfig; verify `ccflags-y += -DKSU` in kernel Makefile
- Apply SELinux hook patches for SUSFS in `security/selinux/hooks.c`
- Ensure FUSE passthrough enabled (unlike 4.19 where it's disabled — GKI requires it)
- Run `patch --dry-run` for each patch; fall back to `wiggle` on fuzz failures

## When This Skill Activates
| Trigger | Event | Condition |
|---|---|---|
| Push | `refs/heads/main` or `fire-*` | Changes to `KernelSU/`, `susfs4ksu/`, `security/` |
| PR | `opened` | Paths match `KernelSU/kernel/*.c` or `susfs4ksu/kernel_patches/6.6/*.patch` |
| Manual | `workflow_dispatch` | `patch_scope: ksu|susfs|all` |

## Tech Stack
- **KernelSU**: `KernelSU/` submodule, `kernel/ksu.c`, `CONFIG_KSU_VERSION`
- **SUSFS**: `susfs4ksu/kernel_patches/6.6/*.patch`, `security/susfs4ksu/`
- **Build**: `make ARCH=arm64 LLVM=1`, `patch --dry-run`, `wiggle`
- **Diff**: GKI 6.6; KMI modules; NOT 4.19

## Automated Checks
```yaml
checks:
  - id: "PKC-001"
    name: "KernelSU Submodule Status"
    command: |
      [ -d KernelSU ] && grep -q "ccflags-y += -DKSU" KernelSU/kernel/Makefile && echo "KSU_SUBMODULE_OK"
    severity: "critical"
  - id: "PKC-002"
    name: "SUSFS Patch Dry-Run"
    command: |
      FAIL=0; for f in susfs4ksu/kernel_patches/6.6/*.patch; do
        patch --dry-run -p1 < "$f" >/dev/null 2>&1 || FAIL=$((FAIL+1))
      done; [ "$FAIL" -eq 0 ] && echo "SUSFS_PATCHES_OK"
    severity: "critical"
  - id: "PKC-003"
    name: "Defconfig KSU/SUSFS Symbols"
    command: |
      grep -q "CONFIG_KSU=y" .config && grep -q "CONFIG_SUSFS=y" .config && echo "KSU_SUSFS_SYMBOLS_OK"
    severity: "critical"
  - id: "PKC-004"
    name: "SELinux Hook Presence"
    command: |
      grep -c "susfs_" security/selinux/hooks.c 2>/dev/null | xargs
      [ "$(grep -c 'susfs_' security/selinux/hooks.c 2>/dev/null)" -ge 3 ] && echo "SELINUX_SUSFS_OK"
    severity: "high"
```

## Input/Output Schema
```json
{
  "inputs": [
    {"name": "patch_scope", "type": "string", "enum": ["ksu", "susfs", "all"]},
    {"name": "dry_run", "type": "boolean", "default": true},
    {"name": "submodule_init", "type": "boolean", "default": true}
  ],
  "outputs": {
    "ksu_version": "integer",
    "susfs_patches_applied": "integer",
    "susfs_patches_failed": "integer",
    "selinux_hooks": "integer",
    "rej_files": ["string"]
  }
}
```

## Error Recovery
- **KSU submodule missing**: Run `git submodule update --init --recursive`; verify remote URL in `.gitmodules`
- **SUSFS patch fails on 6.6**: Check patch context lines; regenerate from upstream `susfs4ksu` for GKI 6.6
- **KSU version symbol not found**: Verify `kernel/ksu.c` has `KSU_VERSION` define; update defconfig `CONFIG_KSU_VERSION`
- **FUSE disabled**: GKI requires FUSE; ensure `CONFIG_FUSE_FS=y` (NOT disabled like 4.19)
