package composer

import "context"

type Pipeline[T any] struct {
	steps []ContextFunc[T, T]
}

func NewPipeline[T any](steps ...ContextFunc[T, T]) *Pipeline[T] {
	return &Pipeline[T]{
		steps: steps,
	}
}

func (p *Pipeline[T]) RunGo(tContext T, ctx context.Context) {
	for _, step := range p.steps {
		go func() {
			_, err := step(tContext, ctx)
			if err != nil {
				panic(err)
			}
		}()
	}
}
