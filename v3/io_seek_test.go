package pb

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestPBProxyReadSeekerReplay(t *testing.T) {
	bar := New(6)
	proxy := bar.NewProxyReadSeeker(bytes.NewReader([]byte("abcdef")))
	var _ io.ReadSeeker = proxy
	if !bar.GetBool(Bytes) {
		t.Fatal("seekable proxy must enable byte units")
	}
	data := make([]byte, 4)
	if n, err := proxy.Read(data); n != 4 || err != nil || string(data) != "abcd" || bar.Current() != 4 {
		t.Fatalf("first read: %q, %d, %v, progress %d", data, n, err, bar.Current())
	}
	if pos, err := proxy.Seek(0, io.SeekStart); pos != 0 || err != nil || bar.Current() != 0 {
		t.Fatalf("rewind: %d, %v, progress %d", pos, err, bar.Current())
	}
	if n, err := io.Copy(io.Discard, proxy); n != 6 || err != nil || bar.Current() != 6 {
		t.Fatalf("replay: %d, %v, progress %d", n, err, bar.Current())
	}
	if bar.Total() != 6 {
		t.Fatalf("seeking changed total to %d", bar.Total())
	}
}

func TestPBProxyReadSeekerOriginsAndEOF(t *testing.T) {
	bar := New(6)
	proxy := bar.NewProxyReadSeeker(bytes.NewReader([]byte("abcdef")))
	for _, tc := range []struct {
		offset int64
		whence int
		want   int64
	}{{4, io.SeekStart, 4}, {-1, io.SeekCurrent, 3}, {-2, io.SeekEnd, 4}, {8, io.SeekStart, 8}} {
		pos, err := proxy.Seek(tc.offset, tc.whence)
		if pos != tc.want || err != nil || bar.Current() != tc.want {
			t.Fatalf("seek(%d,%d): %d, %v, progress %d", tc.offset, tc.whence, pos, err, bar.Current())
		}
	}
	if n, err := proxy.Read(make([]byte, 2)); n != 0 || err != io.EOF || bar.Current() != 8 {
		t.Fatalf("EOF: %d, %v, progress %d", n, err, bar.Current())
	}
	if _, err := proxy.Seek(-1, io.SeekStart); err == nil || bar.Current() != 8 {
		t.Fatalf("failed seek changed progress: %v, %d", err, bar.Current())
	}
}

func TestPBProxyReadSeekerAlignedInitialPosition(t *testing.T) {
	reader := bytes.NewReader([]byte("abcdef"))
	reader.Seek(3, io.SeekStart)
	bar := New(6).SetCurrent(3)
	proxy := bar.NewProxyReadSeeker(reader)
	if n, err := proxy.Read(make([]byte, 2)); n != 2 || err != nil || bar.Current() != 5 {
		t.Fatalf("aligned read: %d, %v, progress %d", n, err, bar.Current())
	}
}

func TestPBProxyReadSeekerErrors(t *testing.T) {
	readErr := errors.New("partial read")
	seekErr := errors.New("cannot seek")
	reader := &errorReadSeeker{readErr, seekErr}
	bar := New(6).SetCurrent(3)
	proxy := bar.NewProxyReadSeeker(reader)
	data := make([]byte, 2)
	if n, err := proxy.Read(data); n != 2 || err != readErr || string(data) != "xy" || bar.Current() != 5 {
		t.Fatalf("partial read: %q, %d, %v, progress %d", data, n, err, bar.Current())
	}
	if pos, err := proxy.Seek(0, io.SeekStart); pos != 41 || err != seekErr || bar.Current() != 5 {
		t.Fatalf("failed seek: %d, %v, progress %d", pos, err, bar.Current())
	}
}

func TestPBProxyReadSeekerClose(t *testing.T) {
	closeErr := errors.New("cannot close")
	reader := &closeReadSeeker{Reader: bytes.NewReader(nil), err: closeErr}
	bar := New(0)
	proxy := bar.NewProxyReadSeeker(reader)
	if err := proxy.Close(); err != closeErr || !reader.closed || !bar.finished {
		t.Fatalf("close: %v, reader closed %v, bar finished %v", err, reader.closed, bar.finished)
	}
	bar = New(0)
	if err := bar.NewProxyReadSeeker(bytes.NewReader(nil)).Close(); err != nil || !bar.finished {
		t.Fatalf("close without closer: %v, bar finished %v", err, bar.finished)
	}
}

func TestPBProxyReaderDoesNotExposeSeek(t *testing.T) {
	proxy := New(6).NewProxyReader(bytes.NewReader([]byte("abcdef")))
	if _, ok := interface{}(proxy).(io.Seeker); ok {
		t.Fatal("ordinary proxy unexpectedly implements io.Seeker")
	}
}

type errorReadSeeker struct{ readErr, seekErr error }

func (r *errorReadSeeker) Read(p []byte) (int, error)     { return copy(p, "xy"), r.readErr }
func (r *errorReadSeeker) Seek(int64, int) (int64, error) { return 41, r.seekErr }

type closeReadSeeker struct {
	*bytes.Reader
	closed bool
	err    error
}

func (r *closeReadSeeker) Close() error { r.closed = true; return r.err }
