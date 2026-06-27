# Epitaph Kernel — Development Roadmap

### Custom GKI Kernel for Redmi 12 (fire) — Android 15 HyperOS 2.0

---

```mermaid
graph TD
    classDef completed fill:#1b4d3e,stroke:#2ecc71,stroke-width:2px,color:#fff;
    classDef current fill:#2c3e50,stroke:#3498db,stroke-width:2px,color:#fff;
    classDef future fill:#1a1a1a,stroke:#555,stroke-width:1px,color:#888;

    P1["Phase 1: Foundation & Rescue"]:::completed
    P2["Phase 2: Power & Performance"]:::completed
    P3["Phase 3: Advanced Optimization"]:::current
    P4["Phase 4: Future Features"]:::future

    P1 --> P2
    P2 --> P3
    P3 --> P4
```

---

## Phase 1: Foundation & Recovery (Completed)

- [x] Unified multi-toolchain CI/CD pipeline (GitHub Actions)
- [x] KernelSU-Next & SUSFS 4 KSU integration
- [x] Xiaomi modular WiFi/Hotspot bypass (vermagic patching)
- [x] RAMoops (PStore) rescue subsystem
- [x] AnyKernel3 flashable ZIP packaging
- [x] Telegram build notification bot

## Phase 2: Power & Performance (Completed)

- [x] ZRAM ZSTD multi-stream compression
- [x] Epitaph Schedutil governor (3 profiles: performance/balanced/battery)
- [x] BFQ + Kyber I/O schedulers
- [x] BBR TCP congestion control + FQ queueing
- [x] MGLRU memory reclamation
- [x] HZ=300 scheduler tuning
- [x] eMMC 5.1 storage I/O latency tweaks
- [x] Epitaph Tuner post-boot script (GPU GED bypass, CPU uclamp)

## Phase 3: Advanced Optimization (Current)

- [ ] ThinLTO binary compression optimization
- [ ] Cortex-A75/A55 targeted compiler flags
- [ ] WireGuard VPN kernel driver
- [ ] Enhanced PStore log parser in Rescue Tool

## Phase 4: Future Features (Planned)

- [ ] LCM panel driver reverse engineering (if needed)
- [ ] Advanced thermal management profiles
- [ ] Custom SELinux policy modules

---

> **Status:** Phases 1-2 fully completed. Currently optimizing build pipeline and binary size in Phase 3.
