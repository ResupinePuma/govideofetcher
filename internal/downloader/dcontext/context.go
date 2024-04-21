package dcontext

import "context"

type iNotifier interface {
	UpdTextNotify(text string) (err error)
	MakeProgressBar(percent float64) (err error)
}

type Context struct {
	ctx context.Context
	n   iNotifier
}

func NewContext(ctx context.Context, notifier iNotifier) Context {
	return Context{
		ctx: ctx,
		n:   notifier,
	}
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Notifier() iNotifier {
	return c.n
}
