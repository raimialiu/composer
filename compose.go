package composer

type Compose[T, TResult any] struct {
	value T
}

func NewCompose[T, TResult any](value T) *Compose[T, TResult] {
	return &Compose[T, TResult]{
		value: value,
	}
}

func (c *Compose[T, TResult]) Create(
	factory Func[T, T],
) *ObjectComposer[T] {
	return NewObjectComposer(factory, c.value)
}

func (c *Compose[T, TResult]) Update(value T) *ObjectComposer[T] {
	ThrowIfNull(value)

	actionFunc := func(in T) T {
		// `in` is the argument passed to actionFunc when invoked
		// `value` is also accessible here via closure
		return in
	}

	return NewObjectComposer(actionFunc, c.value)
}

func (c *Compose[T, TResult]) Decision() *DecisionBuilder[T, TResult] {
	return newDecisionBuilder[T, TResult]()
}
