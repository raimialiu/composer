package composer

import "context"

type Decision[TContext any, TResult any] struct {
	branches []DecisionMapper[TContext, TResult]
	fallback ContextFunc[TContext, TResult]
	tContext *TContext
}

type DecisionMapper[TContext any, TResult any] struct {
	predicate Func[TContext, bool]
	action    ContextFunc[TContext, TResult]
}

func NewDecision[TContext any, TResult any](
	branches []DecisionMapper[TContext, TResult],
	fallback ContextFunc[TContext, TResult],
) Decision[TContext, TResult] {
	if fallback == nil {
		panic("A decision requires an Otherwise branch")
	}
	return Decision[TContext, TResult]{
		branches: append([]DecisionMapper[TContext, TResult](nil), branches...),
		fallback: fallback,
	}
}

func (d *Decision[TContext, TResult]) Apply(
		value ContextFunc[TContext, TResult],
	) TResult {
	d.branches = append(d.branches, DecisionMapper[TContext, TResult]{
		predicate: func(in TContext) bool {
			return true
		},
		action: value,
	})

	return d.Run(context.Background())
}

func (d *Decision[TContext, TResult]) Run(ctx context.Context) TResult {
	result, err := d.RunContext(*d.tContext, ctx)
	if err != nil {
		panic(err)
	}
	return result
}

// RunContext returns cancellation and action errors to the caller.
func (d *Decision[TContext, TResult]) RunContext(tContext TContext, ctx context.Context) (TResult, error) {
	var zero TResult
	for _, branch := range d.branches {
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		if branch.predicate(tContext) {
			if err := ctx.Err(); err != nil {
				return zero, err
			}
			return branch.action(tContext, ctx)
		}
	}

	if err := ctx.Err(); err != nil {
		return zero, err
	}
	return d.fallback(tContext, ctx)
}
