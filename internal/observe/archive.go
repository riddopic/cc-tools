package observe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// bytesPerMegabyte is the number of bytes in one megabyte.
const bytesPerMegabyte = 1024 * 1024

// archiveTimestampFormat is the Go time layout used for rotated file names.
const archiveTimestampFormat = "20060102-150405"

// RotateIfNeeded checks file size and rotates to a timestamped archive if over limit.
// The rotated file is renamed from observations.jsonl to observations-{timestamp}.jsonl.
func RotateIfNeeded(filePath string, maxSizeMB int) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("stat observations file: %w", err)
	}

	maxBytes := int64(maxSizeMB) * bytesPerMegabyte
	if info.Size() < maxBytes {
		return nil
	}

	archivePath := buildArchivePath(filePath)

	if renameErr := os.Rename(filePath, archivePath); renameErr != nil {
		return fmt.Errorf("rename observations file: %w", renameErr)
	}

	return nil
}

func buildArchivePath(filePath string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), ext)
	timestamp := time.Now().Format(archiveTimestampFormat)

	return filepath.Join(dir, base+"-"+timestamp+ext)
}
