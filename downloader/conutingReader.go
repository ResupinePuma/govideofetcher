package downloader

import (
	"errors"
	"io"
)

type CountingReaderOpts struct {
	ByteLimit int
	FileSize  float64
	Notifier  AbstractNotifier
}

type CountingReader struct {
	io.ReadCloser

	opts        *CountingReaderOpts
	bytesReaded int
	useNotifier bool
}

func NewCountingReader(reader io.ReadCloser, opts *CountingReaderOpts) *CountingReader {
	cr := CountingReader{
		ReadCloser: reader,
		opts:       opts,
	}

	if opts.Notifier != nil && opts.FileSize != 0 {
		cr.useNotifier = true
	}
	return &cr
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.bytesReaded += n

	if r.bytesReaded >= r.opts.ByteLimit {
		err = errors.New("size limit reached")
		return
	}
	if r.useNotifier {
		r.opts.Notifier.Count(float64(r.bytesReaded) / r.opts.FileSize * 100)
	}
	return n, err
}

func (r *CountingReader) Close() (err error) {
	return r.ReadCloser.Close()
}
