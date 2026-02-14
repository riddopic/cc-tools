package main

import "fmt"

// exitError carries a non-zero exit code through Cobra's error chain.
// Return this from RunE handlers instead of calling os.Exit() directly,
// which would bypass deferred cleanup.
type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}
