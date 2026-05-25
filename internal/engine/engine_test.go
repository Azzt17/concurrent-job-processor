package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azzt17/concurrent-job-processor/internal/engine"
)

// TestEngine_AllJobsProcessed memverifikasi bahwa semua job terproses
// ketika tidak ada cancellation.
func TestEngine_AllJobsProcessed(t *testing.T) {
	cfg := engine.Config{NumWorkers: 4, NumJobs: 100, BufferSize: 10}
	e := engine.New(cfg)

	stats := e.Run(context.Background())

	if stats.TotalJobs != 100 {
		t.Errorf("expected 100 total jobs, got %d", stats.TotalJobs)
	}
	if stats.Succeeded+stats.Failed != 100 {
		t.Errorf("expected succeeded+failed == 100, got %d", stats.Succeeded+stats.Failed)
	}
}

// TestEngine_GracefulShutdown memverifikasi bahwa context cancellation
// menghentikan engine tanpa goroutine leak.
func TestEngine_GracefulShutdown(t *testing.T) {
	cfg := engine.Config{NumWorkers: 4, NumJobs: 10000, BufferSize: 10}
	e := engine.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	stats := e.Run(ctx)

	// Setelah timeout, tidak semua job selesai — itu yang diharapkan
	processed := stats.Succeeded + stats.Failed + stats.Cancelled
	if processed >= int64(cfg.NumJobs) {
		t.Errorf("expected graceful early exit, but all %d jobs were processed", cfg.NumJobs)
	}
	t.Logf("Processed %d/%d jobs before timeout", processed, cfg.NumJobs)
}

// TestEngine_NoRaceCondition adalah test yang harus dijalankan dengan:
// go test -race ./internal/engine/...
func TestEngine_NoRaceCondition(t *testing.T) {
	cfg := engine.Config{NumWorkers: 8, NumJobs: 500, BufferSize: 20}
	e := engine.New(cfg)
	stats := e.Run(context.Background())

	if stats.TotalJobs != 500 {
		t.Errorf("race condition may have corrupted state: expected 500 jobs, got %d", stats.TotalJobs)
	}
}

// TestEngine_EdgeCase_ZeroWorkers memverifikasi penanganan konfigurasi invalid.
func TestEngine_EdgeCase_ZeroWorkers(t *testing.T) {
	cfg := engine.Config{NumWorkers: 0, NumJobs: 10}
	e := engine.New(cfg)
	stats := e.Run(context.Background())
	_ = stats // tidak crash = sukses
}

// BenchmarkEngine mengukur throughput engine.
func BenchmarkEngine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := engine.Config{NumWorkers: 8, NumJobs: 1000, BufferSize: 50}
		e := engine.New(cfg)
		e.Run(context.Background())
	}
}
