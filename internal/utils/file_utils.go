package utils

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

// CreateTempDirForSDK creates a unique temporary directory for SDK generation.
// baseDir is the root directory under which temporary directories will be created.
// identifier is a unique string (e.g., SDK record ID) to make the directory name unique.
func CreateTempDirForSDK(baseDir string, identifier string) (string, error) {
	// Ensure the base temporary directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create base temp directory %s: %w", baseDir, err)
	}

	// Create a unique subdirectory within the base temporary directory
	tempDirPath, err := os.MkdirTemp(baseDir, fmt.Sprintf("sdk_gen_%s_*", identifier))
	if err != nil {
		return "", fmt.Errorf("failed to create specific temp directory for %s: %w", identifier, err)
	}
	return tempDirPath, nil
}

// Untar extracts a tar.gz stream to a target directory.
func Untar(gzipStream io.Reader, targetDir string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("Untar: failed to create gzip reader: %w", err)
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("Untar: failed to read next tar header: %w", err)
		}

		targetPath := filepath.Join(targetDir, header.Name)

		// Ensure the target path is within the target directory (security measure)
		if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("Untar: invalid tar header name %s, attempts to escape target directory", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("Untar: failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Create file
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return fmt.Errorf("Untar: failed to create parent directory for %s: %w", targetPath, err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("Untar: failed to create file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close() // Close file before returning error
				return fmt.Errorf("Untar: failed to copy data to file %s: %w", targetPath, err)
			}
			outFile.Close() // Close file successfully
		default:
			// Skip other types (symlinks, etc. for now for simplicity and security)
			// fmt.Printf("Untar: skipping type %c for file %s\n", header.Typeflag, header.Name)
		}
	}
	return nil
}

// ZipDirectory creates a zip archive from a source directory.
func ZipDirectory(sourceDir string, targetZipPath string) error {
	zipFile, err := os.Create(targetZipPath)
	if err != nil {
		return fmt.Errorf("ZipDirectory: failed to create zip file %s: %w", targetZipPath, err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// Ensure sourceDir is clean and absolute for reliable prefix stripping
	cleanSourceDir, err := filepath.Abs(filepath.Clean(sourceDir))
	if err != nil {
		return fmt.Errorf("ZipDirectory: failed to get absolute path for sourceDir %s: %w", sourceDir, err)
	}

	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("ZipDirectory: error accessing path %s: %w", filePath, err)
		}

		// Create a relative path for the file header
		relPath, err := filepath.Rel(cleanSourceDir, filePath)
		if err != nil {
			// This should ideally not happen if filePath is within sourceDir
			return fmt.Errorf("ZipDirectory: failed to get relative path for %s: %w", filePath, err)
		}

		// Ensure consistent path separators for zip headers (use '/')
		headerName := filepath.ToSlash(relPath)

		if info.IsDir() {
			// Add a trailing slash for directories to make them explicit in the zip archive
			// Some tools might rely on this.
			if !strings.HasSuffix(headerName, "/") {
				headerName += "/"
			}
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("ZipDirectory: failed to create zip header for %s: %w", filePath, err)
		}
		header.Name = headerName // Use the relative path

		if info.IsDir() {
			header.Name += "/"        // Ensure directory entries have a trailing slash
			header.Method = zip.Store // Directories don't need compression
		} else {
			header.Method = zip.Deflate // Use Deflate for files
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("ZipDirectory: failed to create entry in zip for %s: %w", headerName, err)
		}

		if !info.IsDir() {
			fileToZip, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("ZipDirectory: failed to open file %s for zipping: %w", filePath, err)
			}
			defer fileToZip.Close()

			if _, err := io.Copy(writer, fileToZip); err != nil {
				return fmt.Errorf("ZipDirectory: failed to copy data from %s to zip: %w", filePath, err)
			}
		}
		return nil
	})

	if err != nil {
		// If filepath.Walk encountered an error, remove the partially created zip file.
		archive.Close()
		zipFile.Close()
		os.Remove(targetZipPath)
		return fmt.Errorf("ZipDirectory: error walking through source directory %s: %w", sourceDir, err)
	}

	return nil
}
