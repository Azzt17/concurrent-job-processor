# 🚀 Concurrent Job Processor

[![CI — Test, Race Detect & Security Audit](https://github.com/Azzt17/concurrent-job-processor/actions/workflows/ci.yml/badge.svg)](https://github.com/Azzt17/concurrent-job-processor/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)

Concurrent Job Processor adalah program _Command Line Interface_ (CLI) berkinerja tinggi yang dirancang untuk memproses ribuan tugas secara asinkron. Proyek ini mendemonstrasikan implementasi **Worker Pool Pattern**, manajemen memori yang ketat (_memory-safe_), dan praktik **DevSecOps (Shift-Left Security)**.

## ✨ Fitur Utama

- **Worker Pool Architecture (Fan-Out/Fan-In):** Mencegah _resource exhaustion_ dengan membatasi jumlah _goroutine_ yang berjalan secara bersamaan.
- **Race-Free State Aggregation:** Menggunakan `sync/atomic` untuk menjamin integritas data (thread-safe) tanpa _bottleneck_ dari `sync.Mutex`.
- **Graceful Shutdown:** Mendukung propagasi pembatalan (_cancellation_) ke seluruh unit kerja menggunakan `context.Context` dan menangkap sinyal OS (`SIGINT`/`SIGTERM`).
- **Zero Goroutine Leaks:** Orkestrasi saluran (_channel_) yang aman dipadukan dengan `sync.WaitGroup`.
- **DevSecOps Ready:** Terintegrasi penuh dengan GitHub Actions untuk menjalankan _Unit Tests_, _Race Detector_, _Staticcheck_ (Linter), _GoSec_ (SAST), dan _govulncheck_ secara otomatis pada setiap _push_.

## 🏗️ Arsitektur Sistem

Proyek ini memisahkan secara ketat antara **Input (Job)** dan **Output (Result)** untuk menjunjung tinggi prinsip _immutability_.

1. **Dispatcher** mengirimkan tugas ke dalam _buffered channel_.
2. Kumpulan **Workers** (_Goroutines_) mengambil tugas dari saluran, memprosesnya, dan mengirimkan hasil ke saluran hasil.
3. **Aggregator** mengumpulkan hasil secara aman menggunakan instruksi _atomic_ di tingkat CPU.
4. Jika pengguna menekan `Ctrl+C` atau batas _timeout_ tercapai, sinyal akan disebarkan dan seluruh _worker_ akan berhenti secara elegan tanpa merusak status memori.

## 🚀 Instalasi & Penggunaan

### Prasyarat

- [Go](https://go.dev/dl/) versi 1.25 atau lebih baru.

### Instalasi

Kloning repositori dan lakukan kompilasi _binary_:

```bash
git clone [https://github.com/Azzt17/concurrent-job-processor.git](https://github.com/Azzt17/concurrent-job-processor.git)
cd concurrent-job-processor
go build -o bin/processor ./cmd/processor/
```

### Cara Menjalankan (CLI)

Gunakan parameter flag untuk mengatur perilaku pemrosesan:

```bash
# Menjalankan dengan konfigurasi default (5 pekerja, 100 tugas)
./bin/processor

# Menjalankan dengan konfigurasi kustom
./bin/processor -workers 8 -jobs 1000 -timeout 15s -buffer 20
```

Daftar Flag yang Tersedia:

- -workers (int): Jumlah goroutine pekerja konkuren (default: 5).
- -jobs (int): Jumlah total tugas yang akan disimulasikan (default: 100).
- -timeout (duration): Batas waktu maksimum eksekusi sebelum dihentikan paksa (default: 30s).
- -buffer (int): Ukuran buffer saluran antrian tugas dan hasil (default: 10).

### Lapisan Keamanan (Shift-Left)

Repositori ini menerapkan pengujian otomatis multi-lapis untuk mendeteksi kerentanan di fase awal pengembangan:

1. Functional Testing: Memastikan seluruh pekerjaan dieksekusi dengan benar.
2. Concurrency Audit: Menggunakan go test -race untuk memblokir kode yang rentan terhadap Data Race.
3. Static Analysis: Menggunakan staticcheck untuk mencegah penggunaan anti-pattern dan efisiensi memori.
4. SAST (Static Application Security Testing): Terintegrasi dengan gosec untuk memindai pola kode berbahaya (misal: CVE, kriptografi lemah).
5. Dependency Audit: Menggunakan govulncheck untuk memastikan tidak ada kerentanan rantai pasok (supply chain attacks).

_Dikembangkan sebagai bagian dari eksplorasi DevSecOps dan arsitektur sistem konkuren._
