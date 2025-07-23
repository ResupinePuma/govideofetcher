package dcontext

import (
	"context"
	"net/url"
	"videofetcher/internal/downloader/media"
	"videofetcher/internal/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
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

	u    *url.URL
	lang string

	results chan []media.Media
}

func NewDownloaderContext(ctx context.Context, notifier iNotifier) *Context {
	// start ticker here?
	return &Context{
		Context: ctx,
		n:       notifier,
		results: make(chan []media.Media),
	}
}

func NewTracerContext(ctx *Context, name string) (*Context, trace.Span) {
	tracer := otel.Tracer(telemetry.ServiceName)
	tctx, span := tracer.Start(ctx.Context, name)
	nctx := Context{
		Context: tctx,
		n:       ctx.n,
		u:       ctx.u,
		lang:    ctx.lang,
		results: ctx.results,
	}
	return &nctx, span
}

func (c *Context) GetUrl() *url.URL {
	return c.u
}

func (c *Context) SetUrl(u *url.URL) {
	c.u = u
}

func (c *Context) GetLang() string {
	return c.lang
}

func (c *Context) SetLang(l string) {
	c.lang = l
}

// func (c *Context) SetProgress(percent float64) error {
// 	return c.n.MakeProgressBar(percent)
// }

func (c *Context) Notifier() iNotifier {
	return c.n
}

func (c *Context) Results() chan []media.Media {
	return c.results
}

func (c *Context) AddResult(v []media.Media) {
	if c.results == nil {
		c.results = make(chan []media.Media)
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
