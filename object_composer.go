package composer

type ObjectComposer[T any] struct {
	factory Func[T, T]
	changes []Func[T, T]
	value   T
}

func NewObjectComposer[T any](
	factory Func[T, T],
	value T,
) *ObjectComposer[T] {
	return &ObjectComposer[T]{
		factory: factory,
		changes: []Func[T, T]{},
		value:   value,
	}
}

func (o *ObjectComposer[T]) With(
	change Func[T, T],
) *ObjectComposer[T] {
	ThrowIfNull(change)
	o.changes = append(o.changes, func(in T) T {
		return change(in)
	})

	return o
}

func (o *ObjectComposer[T]) Transform(
	change Func[T, T],
) *ObjectComposer[T] {
	ThrowIfNull(change)
	o.changes = append(o.changes, change)
	return o
}

func (o *ObjectComposer[T]) When(
	predicate Func[T, bool],
	change Func[T, T],
) *ObjectComposer[T] {
	if predicate == nil || change == nil {
		panic("Argument cannot be nil")
	}

	return o.With(func(in T) T {
		if predicate(in) {
			return change(in)
		}

		var zero T
		return zero
	})
}

func (o *ObjectComposer[T]) WhenWithAdapter(
	predicate Func[T, bool],
	change Func[T, T],
	adapter ObjectAdapter[T],
) *ObjectComposer[T] {
	if predicate == nil || change == nil {
		panic("Argument cannot be nil")
	}

	return o.With(func(in T) T {
		if predicate(in) {
			return adapter.Apply(in)
		}

		return in
	})
}

func (o *ObjectComposer[T]) Build() T {
	value := o.factory(o.value)
	for _, change := range o.changes {
		value = change(value)
		ThrowIfNull(value)
	}

	return value
}
