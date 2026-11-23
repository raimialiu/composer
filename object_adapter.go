package composer

type ObjectAdapter[T any] interface {
	Apply(value T) T
}
