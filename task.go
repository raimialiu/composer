package composer

type Task[T any] struct {
	value T
}

func (t *Task[T]) Value() T { return t.value }
