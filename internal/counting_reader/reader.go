package counting_reader

import (
	"errors"
	"io"
)

const blockSize = 500 * 1024

type CountingReaderOpts struct {
	ByteLimit int64
	FileSize  float64
}

type CountingReader struct {
	io.ReadCloser

	opts CountingReaderOpts

	bytesReaded int
	blocknum    int
}

func NewCountingReader(reader io.ReadCloser, opts *CountingReaderOpts) *CountingReader {
	cr := CountingReader{
		ReadCloser: reader,
	}
	cr.opts = *opts
	return &cr
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	if err != nil {
		return
	}
	r.bytesReaded += n

	if int64(r.bytesReaded) >= r.opts.ByteLimit {
		err = errors.New("size limit reached")
		return
	}

	return n, err
}

func (r *CountingReader) Close() (err error) {
	return r.ReadCloser.Close()
}
