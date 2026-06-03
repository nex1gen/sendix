package progress

import (
	"io"

	"github.com/schollz/progressbar/v3"
)

// Reader wraps an io.Reader with a progress bar.
type Reader struct {
	r io.Reader
	b *progressbar.ProgressBar
}

// NewReader creates a progress-tracking reader.
func NewReader(r io.Reader, size int64, description string) *Reader {
	bar := progressbar.DefaultBytes(size, description)
	return &Reader{r: r, b: bar}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.b.Add(n)
	return n, err
}

// Writer wraps an io.Writer with a progress bar.
type Writer struct {
	w io.Writer
	b *progressbar.ProgressBar
}

// NewWriter creates a progress-tracking writer.
func NewWriter(w io.Writer, size int64, description string) *Writer {
	bar := progressbar.DefaultBytes(size, description)
	return &Writer{w: w, b: bar}
}

func (pw *Writer) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.b.Add(n)
	return n, err
}
