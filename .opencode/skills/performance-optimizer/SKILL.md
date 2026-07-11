# Performance Optimizer - Epitaph Kernel Fire

## Overview
Tunes the GKI 6.6 kernel for MT6769/Helio G88 performance on Android 15 HyperOS 2.0. Focuses on EAS scheduler tuning, ThinLTO compilation flags, GPU/video driver optimization, and cgroup management for modern Android. Baseline-compares against 4.19 Pollux metrics.

## Core Responsibilities
- Tune EAS/CFS scheduler parameters for MT6769 big.LITTLE (Cortex-A75/A55 with different freq ranges than MT6768)
- Enable and verify ThinLTO passes for kernel module optimization
- Optimize MediaTek DRM/KMS display pipeline for HyperOS 2.0 UI compositing
- Configure cgroup v2 for Android 15 app governance (top-app, foreground, background)
- Benchmark with `perf bench`, `schedtool`, and `trace-cmd`
- Compare regression against stored `config/baseline.json` metrics

## When This Skill Activates
| Trigger | Event | Condition |
|---|---|---|
| Push | `refs/heads/opt-*` | Changes to `kernel/sched/`, `drivers/gpu/drm/mediatek/` |
| Manual | `workflow_dispatch` | `tune_target: sched|thinlto|gpu|memory` |
| Schedule | Weekly Saturday 03:00 UTC | Benchmark regression run |

## Tech Stack
- **Kernel**: GKI 6.6, EAS (schedutil), ThinLTO, cgroup v2
- **Profiling**: `perf stat`, `ftrace`, `trace-cmd`, `simpleperf`
- **Configs**: `CONFIG_SCHED_ENERGY=y`, `CONFIG_THINLTO=y`, `CONFIG_CGROUP_SCHED=y`
- **Different from 4.19**: GKI KMI constraints; ThinLTO instead of full LTO; FUSE passthrough enabled

## Automated Checks
```yaml
checks:
  - id: "OPT-001"
    name: "EAS Config Verification"
    command: |
      grep -E "CONFIG_SCHED_ENERGY|CONFIG_CPU_FREQ_GOV_SCHEDUTIL|ENERGY_MODEL" .config
    severity: "high"
  - id: "OPT-002"
    name: "ThinLTO Build Flag Check"
    command: |
      grep -q "CONFIG_THINLTO=y" .config && echo "THINLTO_ENABLED"
      grep "LTO_CLANG\|THINLTO" .config
    severity: "high"
  - id: "OPT-003"
    name: "Benchmark Execution"
    command: |
      [ -f scripts/bench_gki.sh ] && bash scripts/bench_gki.sh --quick 2>&1 | tee gki_bench.log
      echo "BENCH_DONE"
    severity: "medium"
  - id: "OPT-004"
    name: "Cgroup V2 Configuration"
    command: |
      grep -q "CONFIG_CGROUP=y" .config && grep -q "CONFIG_CGROUP_SCHED=y" .config && echo "CGROUP_V2_OK"
    severity: "medium"
```

## Input/Output Schema
```json
{
  "inputs": [
    {"name": "tune_target", "type": "string", "enum": ["sched", "thinlto", "gpu", "memory", "all"]},
    {"name": "baseline_ref", "type": "string", "default": "HEAD~1"}
  ],
  "outputs": {
    "sched_latency_us": "float",
    "gpu_fps": "float",
    "build_time_s": "integer",
    "kernel_image_size_bytes": "integer",
    "bench_delta_pct": "float"
  }
}
```

## Error Recovery
- **ThinLTO build OOM**: Reduce `-j` jobs; increase `vm.mmap_min_addr`; use `CONFIG_THINLTO=n` as fallback
- **EAS tune causes boot hang**: Boot with `schedutil` fallback governor; tune per-cluster DVFS
- **Benchmark regression >5%**: Use `git bisect` with `scripts/bisect-build.sh` to find culprit commit
- **MTK DRM failing**: Check `mediatek-drm` Kconfig deps; verify `CONFIG_DRM_MEDIATEK=y`
