# Security Auditor - Epitaph Kernel Fire

## Overview
Performs security audit of the GKI 6.6 kernel with KernelSU and SUSFS integration. Validates SELinux policies for 6.6's LSM framework, checks KSU module isolation in GKI KMI context, scans for MediaTek vendor CVE patterns, and verifies patch provenance for all external patches.

## Core Responsibilities
- Audit SELinux policy compatibility: 6.6 LSM hooks differ from 4.19; verify SUSFS hooks match new API
- Validate KernelSU module loading via `kernel/ksu.c` in GKI context
- Scan `drivers/mediatek/` and `drivers/gpu/mediatek/` for untrusted DMA or user copy patterns
- Verify all patches in `patches/6.6/` have `Signed-off-by`, `From`, and `Date` headers
- Generate CVE audit report with `scripts/cve_scan.py` against NVD database
- Produce SBOM for all kernel modules and external sources

## When This Skill Activates
| Trigger | Event | Condition |
|---|---|---|
| Push | `refs/heads/main` | Changes to `security/`, `KernelSU/`, `kernel/ksu.c` |
| PR | `opened` | Paths match `security/selinux/**` or `patches/6.6/**` |
| Schedule | Weekly Monday 04:00 UTC | CVE database sync and scan |
| Manual | `workflow_dispatch` | `audit_scope: full|selinux|ksu|cve` |

## Tech Stack
- **Security**: SELinux (6.6 LSM), SUSFS (security/susfs4ksu/), KernelSU (kernel/ksu.c)
- **Tools**: `grep`, `checkpatch.pl`, `sparse`, `smatch`
- **Config keys**: `CONFIG_SECURITY_SELINUX`, `CONFIG_KSU`, `CONFIG_SUSFS`, `CONFIG_LSM`
- **Key files**: `security/selinux/hooks.c`, `kernel/ksu.c`, `security/susfs4ksu/Kconfig`

## Automated Checks
```yaml
checks:
  - id: "SEC-001"
    name: "SELinux LSM Hook Audit"
    command: |
      grep -c "susfs_" security/selinux/hooks.c 2>/dev/null || echo "0"
      [ "$(grep -c 'susfs_' security/selinux/hooks.c)" -ge 3 ] && echo "LSM_HOOKS_OK"
    severity: "critical"
  - id: "SEC-002"
    name: "KSU GKI Compliance"
    command: |
      grep -q "KSU_VERSION" kernel/ksu.c && grep -q "ksu_handle_exec" kernel/ksu.c
      grep -q "ccflags-y += -DKSU" KernelSU/kernel/Makefile && echo "KSU_GKI_OK"
    severity: "critical"
  - id: "SEC-003"
    name: "Patch Header Validation"
    command: |
      FAIL=0; for f in patches/6.6/*.patch; do
        grep -q "^Signed-off-by:" "$f" || FAIL=$((FAIL+1))
      done; [ "$FAIL" -eq 0 ] && echo "ALL_PATCHES_SIGNED_OK"
    severity: "high"
  - id: "SEC-004"
    name: "Sparse Static Analysis"
    command: |
      which sparse >/dev/null 2>&1 && make -j$(nproc) ARCH=arm64 LLVM=1 C=1 2>&1 | grep -c "warning:" | xargs
      echo "SPARSE_DONE"
    severity: "medium"
```

## Input/Output Schema
```json
{
  "inputs": [
    {"name": "audit_scope", "type": "string", "enum": ["full", "selinux", "ksu", "cve"]},
    {"name": "cve_db_path", "type": "string", "default": "/var/lib/cve-db"}
  ],
  "outputs": {
    "selinux_hooks": "integer",
    "ksu_compliance": "pass|fail",
    "patches_signed": "integer",
    "patches_unsigned": "integer",
    "cve_findings": ["object"],
    "sparse_warnings": "integer"
  }
}
```

## Error Recovery
- **GKI 6.6 LSM API changed**: Check `include/linux/lsm_hooks.h` for updated hook signatures; regenerate SUSFS patches
- **KSU module fails GKI signature check**: Verify `CONFIG_MODULE_SIG=y` and sign module with GKI key
- **Sparse warnings exceed threshold**: Filter known GKI false positives; fix genuine issues in vendor drivers
