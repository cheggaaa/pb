package pb

import (
	"io"
)

// It's proxy, implement io.ReadWriteCloser
type ReadWriteCloser struct {
	io.ReadWriteCloser
	bar *ProgressBar
}

func (r *ReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = r.ReadWriteCloser.Read(p)
	r.bar.Add(n)
	return
}

func (r *ReadWriteCloser) Write(p []byte) (n int, err error) {
	n, err = r.ReadWriteCloser.Write(p)
	r.bar.Add(n)
	return
}

func (r *ReadWriteCloser) Close() error {
	return r.ReadWriteCloser.Close()
}
