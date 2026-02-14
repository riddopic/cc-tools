package compact

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// logFileName is the name of the compaction log file.
const logFileName = "compaction-log.txt"

// LogCompaction appends a timestamped compaction entry to the log file
// in logDir.
func LogCompaction(logDir string) error {
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, logFileName)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open compaction log: %w", err)
	}
	defer f.Close()

	entry := fmt.Sprintf("[%s] compaction triggered\n",
		time.Now().Format("2006-01-02 15:04:05"))

	if _, writeErr := f.WriteString(entry); writeErr != nil {
		return fmt.Errorf("write compaction log entry: %w", writeErr)
	}

	return nil
}
