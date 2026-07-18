package sync

import (
	"context"
	"sync"
)

type Token struct{}

type ErrGroup struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
	err    error
}

func NewErrGroup(ctx context.Context) (*ErrGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &ErrGroup{cancel: cancel}, ctx
}

func (g *ErrGroup) Go(fn func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := fn(); err != nil {
			g.once.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *ErrGroup) Wait() error {
	g.wg.Wait()
	return g.err
}
