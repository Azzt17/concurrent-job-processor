package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azzt17/concurrent-job-processor/internal/engine"
)

func main() {
	// ── 1. Parsing Konfigurasi via CLI ─────────────────────────────────
	numWorkers := flag.Int("workers", 5, "Jumlah worker goroutine (default: 5)")
	numJobs := flag.Int("jobs", 100, "Jumlah total job yang akan diproses (default: 100)")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout maksimum (default: 30s)")
	bufferSize := flag.Int("buffer", 10, "Ukuran buffer channel (default: 10)")
	flag.Parse()

	log.Printf("Starting processor: workers=%d, jobs=%d, timeout=%s\n",
		*numWorkers, *numJobs, *timeout)

	// ── 2. Setup Context dengan Batas Waktu ──────────────────────────────
	// Context ini adalah kunci utama penyebaran sinyal pembatalan ke seluruh engine
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// ── 3. Tangani Sinyal OS (Interupsi / Kill) ──────────────────────────
	sigChan := make(chan os.Signal, 1)
	// Kita menangkap interupsi keyboard (Ctrl+C) dan sinyal terminasi standar
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\n[SIGNAL] Menerima sinyal %s. Memulai graceful shutdown...\n", sig)
		// Memanggil cancel() akan menutup channel ctx.Done() di seluruh worker
		cancel()
	}()

	// ── 4. Eksekusi Engine Utama ─────────────────────────────────────────
	cfg := engine.Config{
		NumWorkers: *numWorkers,
		NumJobs:    *numJobs,
		BufferSize: *bufferSize,
	}

	e := engine.New(cfg)
	stats := e.Run(ctx)

	// ── 5. Pelaporan Hasil ───────────────────────────────────────────────
	fmt.Printf("\n═══════════════════════════════════════\n")
	fmt.Printf("  Execution Summary\n")
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Printf("  Total Jobs    : %d\n", stats.TotalJobs)
	fmt.Printf("  Succeeded     : %d\n", stats.Succeeded)
	fmt.Printf("  Failed        : %d\n", stats.Failed)
	fmt.Printf("  Cancelled     : %d\n", stats.Cancelled)
	fmt.Printf("  Total Duration: %s\n", stats.TotalDuration.Round(time.Millisecond))

	// Cegah pembagian dengan nol jika program selesai sangat cepat (< 1 detik)
	if stats.TotalDuration.Seconds() > 0 {
		throughput := float64(stats.Succeeded+stats.Failed) / stats.TotalDuration.Seconds()
		fmt.Printf("  Throughput    : %.1f jobs/sec\n", throughput)
	}
	fmt.Printf("═══════════════════════════════════════\n")

	// 6. Konvensi OS UNIX: Kode keluar 130 untuk terminasi via Ctrl+C
	if ctx.Err() == context.Canceled {
		os.Exit(130)
	}
}
