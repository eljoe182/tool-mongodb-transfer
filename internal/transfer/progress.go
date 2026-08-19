package transfer

import (
	"fmt"
	"io"
)

// Reporter reports how many documents have been processed out of the origin total.
type Reporter func(done, total int64)

// WriteProgress prints a single-line progress update to w.
func WriteProgress(w io.Writer, done, total int64) {
	if w == nil {
		return
	}
	percent := 100
	if total > 0 {
		percent = int(done * 100 / total)
		if percent > 100 {
			percent = 100
		}
	}
	fmt.Fprintf(w, "\r%d/%d (%d%%)", done, total, percent)
}

// LogDocumentError writes a per-document transfer failure.
func LogDocumentError(w io.Writer, id string, err error) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, "error transferring document %s: %v\n", id, err)
}
