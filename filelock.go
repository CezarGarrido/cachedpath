package cachedpath

import (
	"os"
	"syscall"
	"time"
)

// FileLock implementa um sistema de lock de arquivo para prevenir race conditions
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock cria um novo FileLock
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: path,
	}
}

// Lock acquires the file lock (with retry)
func (fl *FileLock) Lock() error {
	// Create lock file if it doesn't exist
	file, err := os.OpenFile(fl.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	fl.file = file

	// Try to acquire exclusive lock with retry
	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}

		// If lock is being used by another process, wait
		if err == syscall.EWOULDBLOCK {
			time.Sleep(1 * time.Second)
			continue
		}

		// Other error
		file.Close()
		return err
	}

	file.Close()
	return ErrLockFailed
}

// Unlock libera o lock do arquivo
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return nil
	}

	// Release the lock
	err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
	if err != nil {
		fl.file.Close()
		return err
	}

	// Close the file
	return fl.file.Close()
}

// WithLock executes a function with lock acquired
func WithLock(lockPath string, fn func() error) error {
	lock := NewFileLock(lockPath)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	return fn()
}
