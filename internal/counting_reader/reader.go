package counting_reader

import (
	"errors"
	"io"
)

type iNotify interface {
	MakeProgressBar(percent float64) (err error)
}

type CountingReaderOpts struct {
	ByteLimit int64
	FileSize  float64
	Notifier  iNotify
}

type CountingReader struct {
	io.ReadCloser

	opts CountingReaderOpts

	bytesReaded int
	useNotifier bool
}

func NewCountingReader(reader io.ReadCloser, opts *CountingReaderOpts) *CountingReader {
	cr := CountingReader{
		ReadCloser: reader,
	}
	if opts.Notifier != nil {
		cr.useNotifier = true
		cr.opts = *opts
	}
	return &cr
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.bytesReaded += n

	if int64(r.bytesReaded) >= r.opts.ByteLimit {
		err = errors.New("size limit reached")
		return
	}
	if r.useNotifier {
		r.opts.Notifier.MakeProgressBar(float64(r.bytesReaded) / r.opts.FileSize * 100)
	}
	return n, err
}

func (r *CountingReader) Close() (err error) {
	return r.ReadCloser.Close()
}
