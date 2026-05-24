<div align="center">

# Epitaph Kernel

Epitaph Kernel adalah proyek pengembangan kernel kustom berbasis Generic Kernel Image (GKI 6.6) yang dirancang khusus untuk perangkat **Redmi 12 (fire)** yang menjalankan **HyperOS 2.0 (Android 15)**.

[![GitHub Stars](https://img.shields.io/github/stars/naidrahiqa/epitaph_kernel?style=for-the-badge&logo=github&color=yellow)](https://github.com/naidrahiqa/epitaph_kernel/stargazers)

![Android](https://img.shields.io/badge/Android-15-3DDC84?style=flat-square&logo=android)
![Xiaomi](https://img.shields.io/badge/Device-Redmi%2012%20(fire)-FF6900?style=flat-square&logo=xiaomi)
![Chipset](https://img.shields.io/badge/Chipset-MT6768-orange?style=flat-square)

Proyek ini bertujuan untuk menyediakan build kernel yang stabil dengan integrasi solusi root modern dan fitur keamanan tingkat kernel, dengan tetap mempertahankan kompatibilitas penuh terhadap modul vendor bawaan.

</div>

---

## 🛡️ Solusi Root & Fitur Kernel

*   **KernelSU-Next**: Metode root terintegrasi langsung di tingkat kernel untuk kontrol akses sistem yang lebih aman dan efisien.
*   **SUSFS (Pilihan Build)**: Patch keamanan tingkat kernel untuk menyembunyikan status modifikasi sistem, mount point kustom, serta keberadaan root dari deteksi aplikasi perbankan atau deteksi keamanan sensitif.
*   **Netfilter NAT**: Dukungan modul penunjang hotspot (`CONFIG_NF_NAT`, `CONFIG_IP_NF_NAT`, dan `CONFIG_NETFILTER_XT_TARGET_MASQUERADE`) untuk memulihkan fungsi berbagi koneksi internet.
*   **Optimasi Gaming & Low Latency**: 
    *   **HZ_1000**: Timer 1000Hz untuk input lag minimal dan responsivitas layar sentuh yang lebih cepat.
    *   **Full PREEMPT**: Kernel yang dapat di-interrupt secara instan untuk rendering game yang jauh lebih halus.
    *   **LRU-Gen & LZ4 ZRAM**: Manajemen memori RAM yang super cepat dan efisien demi mencegah lag saat multitasking berat.
*   **Optimasi Jaringan**: Menggunakan TCP BBR Congestion Control dan FQ (Fair Queueing) Packet Scheduler untuk transmisi data yang lebih stabil dan ping game yang lebih konsisten.

---

## 📥 Panduan Instalasi Bersih (Clean Installation)

### 1. Menonaktifkan Verifikasi Partisi (Hanya Sekali)
Sebelum memasang kernel kustom untuk pertama kalinya, Anda wajib menonaktifkan pemeriksaan veritas (AVB) pada partisi perangkat Anda.
Jalankan perangkat ke dalam **Fastboot Mode** lalu eksekusi perintah berikut melalui PC:
```bash
fastboot --disable-verity --disable-verification flash vbmeta vbmeta.img
fastboot --disable-verity --disable-verification flash vbmeta_system vbmeta_system.img
fastboot --disable-verity --disable-verification flash vbmeta_vendor vbmeta_vendor.img
```
*Catatan: Ekstrak file `.img` di atas dari Fastboot ROM stock HyperOS 2.0 resmi yang sesuai dengan versi firmware perangkat Anda.*

### 2. Pemasangan Kernel
Karena kompatibilitas rendering grafis pada Custom Recovery (TWRP/OrangeFox) di Android 15 untuk Redmi 12 masih sangat terbatas, pemasangan sangat direkomendasikan menggunakan aplikasi **Kernel Flasher**.

1.  Unduh arsip instalasi AnyKernel3 yang sesuai dari halaman rilis (misalnya: `Epitaph-v72-bazel-schedutil-false.zip`).
2.  Pasang aplikasi [Kernel Flasher Manager](https://github.com/capntrips/KernelFlasher/releases) pada perangkat Anda.
3.  Buka aplikasi, pilih arsip AnyKernel3 yang telah diunduh, lalu tekan opsi **Flash**.
4.  Setelah proses selesai dengan sukses, lakukan reboot sistem.
5.  Gunakan aplikasi manager pendukung seperti KernelSU-Next Manager untuk mengelola hak akses root.

---

## 👥 AI Alliance & Kontributor

Proyek ini dikembangkan secara kolaboratif oleh aliansi manusia dan kecerdasan buatan super epik:

| Kontributor | Peran & Deskripsi | Logo Badge |
| :--- | :--- | :--- |
| **Faqih Ardian Syah ([@naidrahiqa](https://github.com/naidrahiqa))** | *Lead Maintainer & Kernel Architect* (Otak & Jiwa Manusia Utama 🧠) | ![Human](https://img.shields.io/badge/Human-Developer-brightgreen?style=flat-square) |
| **Antigravity AI** | *Autonomous AI Co-Developer & CI/CD Debugger* (Flight Controller Utama 🚀) | ![Antigravity](https://img.shields.io/badge/Antigravity-AI-FF1493?style=flat-square&logo=google-gemini&logoColor=white) |
| **Gemini** (Google DeepMind) | *Architectural Reasoning & Decision Maker* (Pemikir Teoretis & Konseptual 🌌) | ![Gemini](https://img.shields.io/badge/Gemini-Google-8E75C2?style=flat-square&logo=googlegemini&logoColor=white) |
| **Claude** (Anthropic) | *Syntax Refiner & Code Aesthetics Specialist* (Pujangga Struktur & Perapi Tanda Baca ✍️) | ![Claude](https://img.shields.io/badge/Claude-Anthropic-D97706?style=flat-square&logo=claude&logoColor=white) |
| **DeepSeek** | *Algorithm Wizard & Performance Optimizer* (Penyihir Kode & Pemburu Lag ⚡) | ![DeepSeek](https://img.shields.io/badge/DeepSeek-AI-007AFF?style=flat-square&logo=deepseek&logoColor=white) |
| **Qwen** (Alibaba) | *Build System Helper & Sandbox Cracker* (Penjinak Bazel & Scripting Ninja 🐼) | ![Qwen](https://img.shields.io/badge/Qwen-Alibaba-4B0082?style=flat-square&logo=alibaba&logoColor=white) |

Terima kasih atas kolaborasi erat dan dedikasi luar biasa untuk membangun kernel Redmi 12 terbaik!
