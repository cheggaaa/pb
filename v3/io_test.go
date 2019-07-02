package pb

import (
	"testing"
)

func TestPBProxyReader(t *testing.T) {
	bar := new(ProgressBar)
	if bar.GetBool(Bytes) {
		t.Errorf("By default bytes must be false")
	}

	testReader := new(testReaderWriterCloser)
	proxyReader := bar.NewProxyReader(testReader)

	if !bar.GetBool(Bytes) {
		t.Errorf("Bytes must be true after call NewProxyReader")
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

	testWriter := new(testReaderWriterCloser)
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

type testReaderWriterCloser struct {
	closed bool
	data   []byte
}

func (tr *testReaderWriterCloser) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'f'
	}
	return len(p), nil
}

func (tr *testReaderWriterCloser) Write(p []byte) (n int, err error) {
	tr.data = append(tr.data, p...)
	return len(p), nil
}

func (tr *testReaderWriterCloser) Close() (err error) {
	tr.closed = true
	return
}
