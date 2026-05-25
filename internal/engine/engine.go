package engine

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Azzt17/concurrent-job-processor/internal/job"
)

// Config menampung parameter konfigurasi untuk Engine.
type Config struct {
	NumWorkers int
	NumJobs    int
	BufferSize int
}

// Engine mengorkestrasi sistem Worker Pool.
type Engine struct {
	cfg Config
	wg  sync.WaitGroup

	// Variabel state di bawah ini SENGAJA ditulis tanpa perlindungan atomic
	// untuk mendemonstrasikan Race Condition.
	successCount   int64
	failureCount   int64
	cancelledCount int64
}

// New menginisialisasi Engine baru.
func New(cfg Config) *Engine {
	return &Engine{
		cfg: cfg,
	}
}

// ── 1. IMPLEMENTASI WORKER ───────────────────────────────────────────────

// worker bertugas mengambil job dari antrian, memprosesnya, dan mengirim hasil.
func (e *Engine) worker(ctx context.Context, workerID int, jobs <-chan job.Job, results chan<- job.Result) {
	defer e.wg.Done() // Wajib: Sinyal ke WaitGroup bahwa worker ini telah selesai bekerja

	for {
		select {
		case <-ctx.Done():
			// Skenario 1: Program dihentikan paksa (Ctrl+C atau timeout)
			return

		case j, ok := <-jobs:
			if !ok {
				// Skenario 2: Channel jobs ditutup oleh dispatcher (semua tugas selesai)
				return
			}

			// Proses tugas
			result := e.processJob(ctx, workerID, j)

			// Kirim hasil ke channel results.
			// Gunakan select lagi agar tidak blocking selamanya jika program tiba-tiba dibatalkan.
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// processJob mensimulasikan pekerjaan berat (CPU-bound / I/O-bound).
func (e *Engine) processJob(ctx context.Context, workerID int, j job.Job) job.Result {
	result := job.Result{
		JobID:     j.ID,
		WorkerID:  workerID,
		StartedAt: time.Now(),
	}

	simulateDuration := j.Simulate
	if simulateDuration == 0 {
		// Simulasi waktu pengerjaan acak 1-10 ms jika tidak ditentukan
		simulateDuration = time.Duration(rand.Intn(10)+1) * time.Millisecond
	}

	// Simulasi delay pengerjaan dengan respek terhadap Context cancellation
	select {
	case <-time.After(simulateDuration):
		// Lanjut normal
	case <-ctx.Done():
		result.FinishedAt = time.Now()
		result.Status = job.StatusCancelled
		result.Err = ctx.Err()
		return result
	}

	result.FinishedAt = time.Now()

	if j.ShouldFail {
		result.Status = job.StatusFailed
		result.Err = fmt.Errorf("job %d failed as instructed", j.ID)
		return result
	}

	result.Status = job.StatusSuccess
	result.Output = fmt.Sprintf("processed payload: %s", j.Payload)
	return result
}

// ── 2. IMPLEMENTASI DISPATCHER ───────────────────────────────────────────

// dispatcher membuat job dan mengisinya ke dalam channel.
func (e *Engine) dispatcher(ctx context.Context, jobs chan<- job.Job) {
	// PENTING: Dispatcher adalah satu-satunya entitas yang boleh menutup channel jobs
	defer close(jobs)

	for i := 0; i < e.cfg.NumJobs; i++ {
		j := job.Job{
			ID:      i + 1,
			Payload: fmt.Sprintf("data-payload-%d", i+1),
		}

		select {
		case jobs <- j:
		case <-ctx.Done():
			return // Berhenti memproduksi job jika dibatalkan
		}
	}
}

// ── 3. ORKESTRASI (RUN) & AGREGASI SEMENTARA ─────────────────────────────

// Run menjalankan seluruh pipeline dan memblokir hingga semua proses selesai.
func (e *Engine) Run(ctx context.Context) job.Stats {
	startTime := time.Now()

	bufferSize := e.cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = e.cfg.NumWorkers * 2 // Default aman
	}

	jobChan := make(chan job.Job, bufferSize)
	resultChan := make(chan job.Result, bufferSize)

	numWorkers := e.cfg.NumWorkers
	if numWorkers <= 0 {
		numWorkers = 1
	}

	// 1. Luncurkan para pekerja (Fan-Out)
	e.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go e.worker(ctx, i, jobChan, resultChan)
	}

	// 2. Goroutine pemantau: menutup resultChan saat semua worker selesai
	go func() {
		e.wg.Wait()
		close(resultChan)
	}()

	// 3. Luncurkan pendistribusi tugas
	go e.dispatcher(ctx, jobChan)

	// 4. Proses agregasi hasil (Fan-In)
	stats := e.aggregate(resultChan)
	stats.TotalDuration = time.Since(startTime)
	stats.TotalJobs = e.cfg.NumJobs

	return stats
}

// aggregate membaca semua hasil dari resultChan.
// PERINGATAN: Fungsi ini sengaja ditulis dengan operasi primitif
// untuk memancing Race Condition yang akan di analisis nanti.
func (e *Engine) aggregate(results <-chan job.Result) job.Stats {
	for result := range results {
		switch result.Status {
		case job.StatusSuccess:
			atomic.AddInt64(&e.successCount, 1)
		case job.StatusFailed:
			atomic.AddInt64(&e.failureCount, 1)
		case job.StatusCancelled:
			atomic.AddInt64(&e.cancelledCount, 1)
		}
	}

	return job.Stats{
		Succeeded: atomic.LoadInt64(&e.successCount),
		Failed:    atomic.LoadInt64(&e.failureCount),
		Cancelled: atomic.LoadInt64(&e.cancelledCount),
	}
}
