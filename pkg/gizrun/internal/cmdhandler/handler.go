package cmdhandler

import (
	"context"
)

type Handler interface {
	Handle(context.Context, []string, []string) error
}

type HandleFunc func(context.Context, []string, []string) error

func (fn HandleFunc) Handle(ctx context.Context, args []string, flags []string) error {
	return fn(ctx, args, flags)
}
