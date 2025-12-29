package cachedpath

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsArchive checks if a file is an archive (zip or tar.gz)
func IsArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".zip" {
		return true
	}
	if ext == ".gz" && strings.HasSuffix(strings.ToLower(path), ".tar.gz") {
		return true
	}
	if ext == ".tgz" {
		return true
	}
	return false
}

// ExtractArchive extracts a compressed file to a directory
func ExtractArchive(archivePath, destDir string) error {
	if err := EnsureDir(destDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(archivePath))

	if ext == ".zip" {
		return extractZip(archivePath, destDir)
	}

	if ext == ".gz" || ext == ".tgz" {
		return extractTarGz(archivePath, destDir)
	}

	return fmt.Errorf("unsupported archive format: %s", ext)
}

// extractZip extrai um arquivo ZIP
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		err := extractZipFile(f, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(f *zip.File, destDir string) error {
	filePath := filepath.Join(destDir, f.Name)

	// Previne path traversal
	if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(filePath, os.ModePerm)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := f.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// extractTarGz extrai um arquivo tar.gz
func extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		// Previne path traversal
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// ExtractSpecificFile extracts a specific file from an archive
func ExtractSpecificFile(archivePath, internalPath, destDir string) (string, error) {
	if err := EnsureDir(destDir); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(archivePath))

	if ext == ".zip" {
		return extractSpecificFromZip(archivePath, internalPath, destDir)
	}

	if ext == ".gz" || ext == ".tgz" {
		return extractSpecificFromTarGz(archivePath, internalPath, destDir)
	}

	return "", fmt.Errorf("unsupported archive format: %s", ext)
}

func extractSpecificFromZip(zipPath, internalPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == internalPath {
			destPath := filepath.Join(destDir, filepath.Base(internalPath))

			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return "", err
			}

			dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			defer dstFile.Close()

			srcFile, err := f.Open()
			if err != nil {
				return "", err
			}
			defer srcFile.Close()

			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return "", err
			}

			return destPath, nil
		}
	}

	return "", fmt.Errorf("file not found in archive: %s", internalPath)
}

func extractSpecificFromTarGz(tarGzPath, internalPath, destDir string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == internalPath && header.Typeflag == tar.TypeReg {
			destPath := filepath.Join(destDir, filepath.Base(internalPath))

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return "", err
			}

			outFile, err := os.Create(destPath)
			if err != nil {
				return "", err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return "", err
			}

			return destPath, nil
		}
	}

	return "", fmt.Errorf("file not found in archive: %s", internalPath)
}
