package composer

import "context"

type PipelineBuilder[T any] struct {
	steps []ContextFunc[T, T]
}

func (p *PipelineBuilder[T]) Then(
	step ContextFunc[T, T],
) *PipelineBuilder[T] {
	p.steps = append(p.steps, step)
	return p
}

func (p *PipelineBuilder[T]) ThenAction(
	step Action[T],
) *PipelineBuilder[T] {
	ThrowIfNull(step)
	return p.Then(func(in T, ctx context.Context) (T, error) {
		step(in)
		return in, nil
	})
}

func (p *PipelineBuilder[T]) ThenPipeline(
	pipeline Pipeline[T],
) *PipelineBuilder[T] {
	ThrowIfNull(pipeline)
	return p.Then(func(in T, ctx context.Context) (T, error) {
		pipeline.RunGo(in, ctx)
		return in, nil
	})
}

func (p *PipelineBuilder[T]) Choose(
	decision Decision[T, T],
) *PipelineBuilder[T] {
	return p.Then(func(in T, ctx context.Context) (T, error) {
		result, runError := decision.RunContext(in, ctx)
		return result, runError
	})
}

func (p *PipelineBuilder[T]) When(
	predicate Func[T, bool],
	step ContextFunc[T, Task[T]],
) *PipelineBuilder[T] {
	ThrowIfNull(predicate)
	ThrowIfNull(step)

	return p.Then(func(in T, ctx context.Context) (T, error) {
		if !predicate(in) {
			return in, nil
		}
		result, stepError := step(in, ctx)
		return result.Value(), stepError
	})
}

func (p *PipelineBuilder[T]) WhenAction(
	predicate Func[T, bool],
	step Action[T],
) *PipelineBuilder[T] {
	ThrowIfNull(predicate)
	ThrowIfNull(step)

	return p.Then(func(in T, ctx context.Context) (T, error) {
		if !predicate(in) {
			return in, nil
		}

		step(in)
		return in, nil
	})
}

func (p *PipelineBuilder[T]) RepeatWhile(
	step ContextFunc[T, Task[bool]],
	maxIterations Func[T, int],
) *PipelineBuilder[T] {
	ThrowIfNull(maxIterations)
	ThrowIfNull(step)

	return p.Then(func(in T, ctx context.Context) (T, error) {
		var limit = maxIterations(in)
		if limit < 1 {
			panic("limit must be greater than or equal to 1")
		}

		for i := 0; i < limit; i++ {
			if ctx.Err() != nil {
				return in, ctx.Err()
			}

			_, err := step(in, ctx)
			if err != nil {
				var zero T
				return zero, err
			}
		}

		return in, nil
	})
}

func (p *PipelineBuilder[T]) WhenPipeline(
	predicate Func[T, bool],
	pipeline Pipeline[T],
) *PipelineBuilder[T] {
	ThrowIfNull(predicate)
	ThrowIfNull(pipeline)

	return p.When(predicate, func(in T, ctx context.Context) (Task[T], error) {
		pipeline.RunGo(in, ctx)
		return Task[T]{
			value: in,
		}, nil
	})
}

func (p *PipelineBuilder[T]) Build() *Pipeline[T] {
	return NewPipeline(p.steps...)
}
