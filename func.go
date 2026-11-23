package composer

import "context"

type Func[T, O any] func(in T) O

type Action[T any] func(obj T)

type ContextFunc[T, O any] func(in T, ctx context.Context) (O, error)

type ContextAction[T any] func(in T, ctx context.Context) error
