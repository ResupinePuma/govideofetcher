package dcontext

import (
	"context"
	"net/url"
	"videofetcher/internal/downloader/video"
)

type IDownloader interface {
	Download(ctx *Context) error
	Close() error
}

type iNotifier interface {
	UpdTextNotify(text string) (err error)
	//MakeProgressBar(percent float64) (err error)
	StartTicker(ctx context.Context) (err error)
}

type Context struct {
	context.Context
	n iNotifier

	u *url.URL

	results chan []video.Video
}

func NewDownloaderContext(ctx context.Context, notifier iNotifier) *Context {
	// start ticker here?
	return &Context{
		Context: ctx,
		n:       notifier,
		results: make(chan []video.Video),
	}
}

func (c *Context) GetUrl() *url.URL {
	return c.u
}

func (c *Context) SetUrl(u *url.URL) {
	c.u = u
}

// func (c *Context) SetProgress(percent float64) error {
// 	return c.n.MakeProgressBar(percent)
// }

func (c *Context) Notifier() iNotifier {
	return c.n
}

func (c *Context) Results() chan []video.Video {
	return c.results
}

func (c *Context) AddResult(v []video.Video) {
	if c.results == nil {
		c.results = make(chan []video.Video)
	}
	c.results <- v
}

func (c *Context) NextDownloader(d IDownloader) error {
	return d.Download(c)
}

// func (c *Context) Deadline() (deadline time.Time, ok bool) {
// 	return c.Context.Deadline()
// }

// func (c *Context) Done() <-chan struct{} {
// 	return c.ctx.Done()
// }

// func (c *Context) Err() error {
// 	return c.ctx.Err()
// }

// func (c *Context) Value(key any) any {
// 	return c.ctx.Value(key)
// }
