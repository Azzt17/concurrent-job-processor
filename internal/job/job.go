package job

import (
	"time"
)

// JobStatus merepresentasikan hasil akhir sebuah job.
// Kita menggunakan custom type (int) dan iota sebagai enum.
// Ini jauh lebih aman dan hemat memori dibandingkan menggunakan string murni.
type JobStatus int

const (
	StatusSuccess   JobStatus = iota // bernilai 0
	StatusFailed                     // bernilai 1
	StatusCancelled                  // bernilai 2 (dibatalkan sebelum/saat diproses)
)

func (s JobStatus) String() string {
	switch s {
	case StatusSuccess:
		return "SUCCESS"
	case StatusFailed:
		return "FAILED"
	case StatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// Job merepresentasikan satu unit pekerjaan yang masuk ke antrian (Input).
type Job struct {
	ID         int           // Identifikasi unik job
	Payload    string        // Data yang akan diproses
	Simulate   time.Duration // Simulasi waktu pengerjaan (untuk testing)
	ShouldFail bool          // Jika true, worker dipaksa mengembalikan error
}

// Result merepresentasikan hasil pengerjaan dari seorang Worker (Output).
type Result struct {
	JobID      int
	Status     JobStatus
	Output     string
	Err        error
	StartedAt  time.Time
	FinishedAt time.Time
	WorkerID   int // ID Worker mana yang mengeksekusi job ini (untuk tracing)
}

// Duration mengembalikan total waktu pengerjaan satu job spesifik.
func (r Result) Duration() time.Duration {
	return r.FinishedAt.Sub(r.StartedAt)
}

// Stats merepresentasikan ringkasan agregasi seluruh job pool.
type Stats struct {
	TotalJobs     int
	Succeeded     int64
	Failed        int64
	Cancelled     int64
	TotalDuration time.Duration
}
