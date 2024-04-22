package dcontext

import (
	"context"
	"time"
)

type iNotifier interface {
	UpdTextNotify(text string) (err error)
	MakeProgressBar(percent float64) (err error)
}

type Context struct {
	n   iNotifier
	ctx context.Context
}

func NewDownloaderContext(ctx context.Context, notifier iNotifier) Context {
	return Context{
		ctx:   ctx,
		n:     notifier,
	}
}

func (c *Context) Notifier() iNotifier {
	return c.n
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Context) Err() error {
	return c.ctx.Err()
}

func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}
