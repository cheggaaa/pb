package pb

import (
	"io"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestPBProxyReader(t *testing.T) {
	bar := new(ProgressBar)
	if bar.GetBool(Bytes) {
		t.Errorf("By default bytes must be false")
	}

	testReader := new(testReaderWriterSeekerCloser)
	testReader.size = 1000000
	proxyReader := bar.NewProxyReader(testReader)

	if !bar.GetBool(Bytes) {
		t.Errorf("Bytes must be true after call NewProxyReader")
	}

	for i := 0; i < 10; i++ {
		// pick a random offset up to half the size of the Reader in either direction.
		offset := rand.Int63n(testReader.size) - (testReader.size / 2)
		expected :=  testReader.offset + offset
		if expected < 0 {
			expected = 0
		}
		if expected > testReader.size {
			expected = testReader.size
		}
		position, err := proxyReader.Seek(offset, io.SeekCurrent)
		if err != nil {
			t.Errorf("Proxy reader failed to seek: %v", err)
		}
		if position != testReader.offset {
			t.Errorf("Proxy reader offset doesn't match reported offset: %d vs %d", testReader.offset, position)
		}
		if position != expected {
			t.Errorf("Proxy reader returned unexpected position: %d instead of %d / %d", position, expected, testReader.size)
		}
		if proxyReader.bar.Current() != expected {
			t.Errorf("Proxy reader bar returned incorrect position: %d vs %d", proxyReader.bar.Current(), expected)
		}
	}
	offset, err := proxyReader.Seek(0, io.SeekStart)
	if err != nil || offset != 0 || proxyReader.bar.Current() != 0 {
		t.Errorf("Proxy reader failed to reset seek position: %d, %d, %v", offset, proxyReader.bar.Current(), err)
	}

	for i := 0; i < 10; i++ {
		buf := make([]byte, 10)
		n, e := proxyReader.Read(buf)
		if e != nil {
			t.Errorf("Proxy reader return err: %v", e)
		}
		if n != len(buf) {
			t.Errorf("Proxy reader return unexpected N: %d (wand %d)", n, len(buf))
		}
		for _, b := range buf {
			if b != 'f' {
				t.Errorf("Unexpected read value: %v (want %v)", b, 'f')
			}
		}
		if want := int64((i + 1) * len(buf)); bar.Current() != want {
			t.Errorf("Unexpected bar current value: %d (want %d)", bar.Current(), want)
		}
	}
	proxyReader.Close()
	if !testReader.closed {
		t.Errorf("Reader must be closed after call ProxyReader.Close")
	}
	proxyReader.Reader = nil
	proxyReader.Close()
}

func TestPBProxyWriter(t *testing.T) {
	bar := new(ProgressBar)
	if bar.GetBool(Bytes) {
		t.Errorf("By default bytes must be false")
	}

	testWriter := new(testReaderWriterSeekerCloser)
	proxyReader := bar.NewProxyWriter(testWriter)

	if !bar.GetBool(Bytes) {
		t.Errorf("Bytes must be true after call NewProxyReader")
	}

	for i := 0; i < 10; i++ {
		buf := make([]byte, 10)
		n, e := proxyReader.Write(buf)
		if e != nil {
			t.Errorf("Proxy reader return err: %v", e)
		}
		if n != len(buf) {
			t.Errorf("Proxy reader return unexpected N: %d (wand %d)", n, len(buf))
		}
		if want := int64((i + 1) * len(buf)); bar.Current() != want {
			t.Errorf("Unexpected bar current value: %d (want %d)", bar.Current(), want)
		}
	}
	proxyReader.Close()
	if !testWriter.closed {
		t.Errorf("Reader must be closed after call ProxyReader.Close")
	}
	proxyReader.Writer = nil
	proxyReader.Close()
}

type testReaderWriterSeekerCloser struct {
	size int64
	offset int64
	closed bool
	data   []byte
}

func (tr *testReaderWriterSeekerCloser) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'f'
	}
	tr.offset += int64(len(p))
	return len(p), nil
}

func (tr *testReaderWriterSeekerCloser) Seek(offset int64, whence int) (n int64, err error) {
	if whence == io.SeekStart {
		tr.offset = offset
	} else if whence == io.SeekEnd {
		tr.offset = tr.size - offset
	} else if whence == io.SeekCurrent {
		tr.offset += offset
	}

	if tr.offset >= tr.size {
		tr.offset = tr.size
	} else if tr.offset < 0 {
		tr.offset = 0
	}
	return tr.offset, err
}

func (tr *testReaderWriterSeekerCloser) Write(p []byte) (n int, err error) {
	tr.data = append(tr.data, p...)
	return len(p), nil
}

func (tr *testReaderWriterSeekerCloser) Close() (err error) {
	tr.closed = true
	return
}
