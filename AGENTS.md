# AGENTS.md — Epitaph Kernel Fire

Custom GKI 6.6 kernel for Xiaomi Redmi 12 (fire / MT6769 / Helio G88). Android 15 HyperOS 2.0.

---

## Team Registry

| Role | ID | Core Skill |
|------|-----|------------|
| Build Engineer | `builder` | `kernel-build-validator` |
| Patcher | `patcher` | `kernel-build-validator` |
| Performance Optimizer | `optimizer` | `kernel-build-validator` |
| Troubleshooter | `troubleshooter` | `kernel-build-validator` |

## Core Skill

### Kernel Build Validator Developer Skill
- **ID**: `kernel-build-validator`
- **File**: `SKILL.md`
- **Responsibility**: Validate defconfig, toolchain, kernel compilation, KSU/SUSFS integration, artifacts
- **Triggers**: Push to main/fire-*, PR to kernel/drivers/arch/, manual dispatch, daily schedule
- **Input Type**: git-context / directory
- **Output Type**: json
- **Runtime**: ~2400s (full build)
- **Owner**: builder, patcher, optimizer

## Role Skills

| Role | Skill File | Responsibility |
|------|-----------|----------------|
| Kernel Builder (GKI 6.6) | `../../.opencode/skills/kernel-builder-gki/SKILL.md` | Build GKI 6.6 kernel with Bazel/Kleaf; ThinLTO; produce boot.img/dtbo.img |
| Kernel Patcher (KSU/SUSFS) | `../../.opencode/skills/kernel-patcher-ksu/SKILL.md` | Integrate KernelSU and SUSFS into GKI 6.6; manage submodules; SELinux hooks |
| Performance Optimizer | `../../.opencode/skills/performance-optimizer/SKILL.md` | Tune EAS/ThinLTO/GPU for MT6769; cgroup v2; benchmark vs baseline |
| Security Auditor | `../../.opencode/skills/security-auditor/SKILL.md` | Audit GKI 6.6 SELinux/KSU; sparse static analysis; CVE scan; patch provenance |

## Workflow Definitions

### Workflow: Changelog Management
**ID**: `changelog`
**Trigger**: Before release, manual

**File**: `CHANGELOG.md`

**Format**: Keep a Changelog (Added/Changed/Fixed)

**Steps:**
1. Update `CHANGELOG.md` with curated important changes for new version
2. `generate_changelog.py` reads from `CHANGELOG.md` (fallback: flat git log)
3. Release body uses curated changelog

**Rules:**
- Only list important/significant changes, not every commit
- Write in clear English
- Use sections: `### Added`, `### Changed`, `### Fixed`
- One entry per release, newest on top

### Workflow: Kernel Build
**ID**: `kernel-build`
**Trigger**: Push to main, tag created

**Steps:**
1. **Toolchain Check** → `kernel-build-validator` (CHK-001) → on failure: halt
2. **Defconfig Validate** → `kernel-build-validator` (CHK-002) → on failure: halt (depends: step 1)
3. **Full Build** → `kernel-build-validator` (CHK-003) → on failure: halt, retry: 1 (depends: step 2)
4. **KSU + SUSFS Check** → `kernel-build-validator` (CHK-004, CHK-005) → on failure: halt (depends: step 3)
5. **Artifact Verify** → `kernel-build-validator` (CHK-006, CHK-007) → on failure: notify_only (depends: step 3)

**Est. Duration**: ~2700 seconds

### Workflow: Troubleshoot
**ID**: `troubleshoot`
**Trigger**: Build failure

**Steps:**
1. **Diagnose** → `kernel-build-validator` (CHK-001, CHK-002) → on failure: notify_only
2. **Fix + Rebuild** → `kernel-build-validator` (CHK-003) → on failure: halt

### Workflow: PR Review
**ID**: `pr-review`
**Trigger**: PR opened

**Steps:**
1. **Build** → `kernel-build-validator` (CHK-003) → on failure: halt
2. **KSU Check** → `kernel-build-validator` (CHK-004) → on failure: notify_only (parallel: step 1)

**Output Actions**: PR comment

## Critical Constraints

- **GKI 6.6** — NOT 4.19. Different KMI/patch approach.
- **Bazel/Kleaf** build system.
- **ThinLTO** enabled.
- **EAS** tuned for MT6769.
- **NoMount** — Kernel-based file injection framework (CONFIG_NOMOUNT=y, kernel 6.6 supported).
- **CHANGELOG.md** — Curated release notes. Use Keep a Changelog format (Added/Changed/Fixed).

## Notifications

| Channel | Trigger | Recipients |
|---------|---------|------------|
| GitHub Release | Build complete | builder |
| GitHub Commit Status | Build failure | builder |

## Project Context

```json
{
  "project_name": "Epitaph Kernel Fire",
  "project_type": "kernel",
  "primary_languages": ["C", "Shell", "Python"],
  "ci_cd_platform": "github-actions",
  "stage": "development"
}
```
