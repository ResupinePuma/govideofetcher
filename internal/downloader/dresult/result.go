package dresult

import (
	"context"
	"io"
	"os"
)

type DownloadResult struct {
	io.ReadCloser

	ctx     context.Context
	cancel  context.CancelFunc
	TmpPath string
}

func NewDownloaderResult(c context.Context) *DownloadResult {
	ctx, cancel := context.WithCancel(c)
	return &DownloadResult{ctx: ctx, cancel: cancel}
}

func (d *DownloadResult) Read(p []byte) (n int, err error) {
	n, err = d.ReadCloser.Read(p)
	if err != nil {
		d.cancel()
		return
	}
	return
}

func (d *DownloadResult) Close() error {
	d.cancel()
	if d.ReadCloser != nil {
		d.ReadCloser.Close()
	}
	os.RemoveAll(d.TmpPath)
	return nil
}

func (d *DownloadResult) Context() context.Context {
	return d.ctx
}
