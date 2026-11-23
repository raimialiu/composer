package composer

import "context"

// ActionDecisionBuilder builds a first-match decision with no return value.
type ActionDecisionBuilder[TContext any] struct {
	inner DecisionBuilder[TContext, bool]
}

func NewActionDecisionBuilder[TContext any]() *ActionDecisionBuilder[TContext] {
	return &ActionDecisionBuilder[TContext]{}
}

func (d *ActionDecisionBuilder[TContext]) When(
	predicate Func[TContext, bool],
	action Action[TContext],
) *ActionDecisionBuilder[TContext] {
	if action == nil {
		panic("Argument cannot be nil")
	}
	d.inner.When(predicate, func(in TContext) bool {
		action(in)
		return true
	})
	return d
}

func (d *ActionDecisionBuilder[TContext]) WhenContext(
	predicate Func[TContext, bool],
	action ContextAction[TContext],
) *ActionDecisionBuilder[TContext] {
	if action == nil {
		panic("Argument cannot be nil")
	}
	d.inner.WhenContext(predicate, func(in TContext, ctx context.Context) (bool, error) {
		if err := action(in, ctx); err != nil {
			return false, err
		}
		return true, nil
	})
	return d
}

func (d *ActionDecisionBuilder[TContext]) Otherwise(
	action Action[TContext],
) *ActionDecisionBuilder[TContext] {
	if action == nil {
		panic("Argument cannot be nil")
	}
	d.inner.Otherwise(func(in TContext) bool {
		action(in)
		return true
	})
	return d
}

func (d *ActionDecisionBuilder[TContext]) OtherwiseContext(
	action ContextAction[TContext],
) *ActionDecisionBuilder[TContext] {
	if action == nil {
		panic("Argument cannot be nil")
	}
	d.inner.OtherwiseContext(func(in TContext, ctx context.Context) (bool, error) {
		if err := action(in, ctx); err != nil {
			return false, err
		}
		return true, nil
	})
	return d
}

func (d *ActionDecisionBuilder[TContext]) Build() ActionDecision[TContext] {
	return ActionDecision[TContext]{inner: d.inner.Build()}
}

// ActionDecision runs an action selected by its first matching predicate.
type ActionDecision[TContext any] struct {
	inner Decision[TContext, bool]
}

func (d ActionDecision[TContext]) Run(in TContext, ctx context.Context) {
	d.inner.Run(ctx)
}

func (d ActionDecision[TContext]) RunContext(in TContext, ctx context.Context) error {
	_, err := d.inner.RunContext(in, ctx)
	return err
}
