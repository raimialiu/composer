package composer

import "context"

// DecisionBuilder collects ordered branches and a fallback for a decision.
type DecisionBuilder[TContext, TResult any] struct {
	branches []DecisionMapper[TContext, TResult]
	fallback ContextFunc[TContext, TResult]
}

func newDecisionBuilder[TContext, TResult any]() *DecisionBuilder[TContext, TResult] {
	return &DecisionBuilder[TContext, TResult]{}
}

func (d *DecisionBuilder[TContext, TResult]) When(
	predicate Func[TContext, bool],
	action Func[TContext, TResult],
) *DecisionBuilder[TContext, TResult] {
	if predicate == nil || action == nil {
		panic("Argument cannot be nil")
	}
	return d.WhenContext(predicate, func(in TContext, _ context.Context) (TResult, error) {
		return action(in), nil
	})
}

func (d *DecisionBuilder[TContext, TResult]) WhenContext(
	predicate Func[TContext, bool],
	action ContextFunc[TContext, TResult],
) *DecisionBuilder[TContext, TResult] {
	if predicate == nil || action == nil {
		panic("Argument cannot be nil")
	}
	d.branches = append(d.branches, DecisionMapper[TContext, TResult]{
		predicate: predicate,
		action:    action,
	})
	return d
}

func (d *DecisionBuilder[TContext, TResult]) Otherwise(
	fallback Func[TContext, TResult],
) *DecisionBuilder[TContext, TResult] {
	if fallback == nil {
		panic("Argument cannot be nil")
	}
	return d.OtherwiseContext(func(in TContext, _ context.Context) (TResult, error) {
		return fallback(in), nil
	})
}

func (d *DecisionBuilder[TContext, TResult]) OtherwiseContext(
	fallback ContextFunc[TContext, TResult],
) *DecisionBuilder[TContext, TResult] {
	if fallback == nil {
		panic("Argument cannot be nil")
	}
	if d.fallback != nil {
		panic("A decision can have only one fallback")
	}
	d.fallback = fallback
	return d
}

func (d *DecisionBuilder[TContext, TResult]) Build() Decision[TContext, TResult] {
	if d.fallback == nil {
		panic("A decision requires an Otherwise branch")
	}
	return NewDecision(d.branches, d.fallback)
}
