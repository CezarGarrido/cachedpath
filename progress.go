package cachedpath

import (
	"fmt"
	"io"
	"sync/atomic"
)

// ProgressDisplay is the interface for displaying progress
type ProgressDisplay interface {
	Start(total int64, description string)
	Update(written int64)
	Finish()
}

// SimpleProgress implements a simple progress bar
type SimpleProgress struct {
	total       int64
	written     int64
	description string
	quiet       bool
}

// NewSimpleProgress creates a new SimpleProgress
func NewSimpleProgress(quiet bool) *SimpleProgress {
	return &SimpleProgress{
		quiet: quiet,
	}
}

// Start starts the progress display
func (p *SimpleProgress) Start(total int64, description string) {
	p.total = total
	p.written = 0
	p.description = description

	if !p.quiet && total > 0 {
		fmt.Printf("Downloading %s: 0%%\n", description)
	}
}

// Update updates the progress
func (p *SimpleProgress) Update(written int64) {
	atomic.StoreInt64(&p.written, written)

	if !p.quiet && p.total > 0 {
		percentage := float64(written) / float64(p.total) * 100
		fmt.Printf("\rDownloading %s: %.1f%%", p.description, percentage)
	}
}

// Finish finishes the progress display
func (p *SimpleProgress) Finish() {
	if !p.quiet {
		fmt.Println("\nDownload complete!")
	}
}

// ProgressWriter is a writer that updates progress
type ProgressWriter struct {
	writer   io.Writer
	progress ProgressDisplay
	written  int64
}

// NewProgressWriter creates a new ProgressWriter
func NewProgressWriter(writer io.Writer, progress ProgressDisplay) *ProgressWriter {
	return &ProgressWriter{
		writer:   writer,
		progress: progress,
		written:  0,
	}
}

// Write implements io.Writer
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.written += int64(n)
		if pw.progress != nil {
			pw.progress.Update(pw.written)
		}
	}
	return n, err
}

// Written returns the total bytes written
func (pw *ProgressWriter) Written() int64 {
	return pw.written
}
