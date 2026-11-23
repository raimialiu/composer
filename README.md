# composer

`composer` is a Go package for building values, selecting the first matching decision branch, and launching a set of context-aware pipeline steps. The module path is `github.com/raimialiu/composer` and the module currently declares Go 1.27.

The complete, runnable student enrollment example is in [`examples/main.go`](examples/main.go). From the repository root, run:

```sh
go run ./examples
```

## Quick start

Use this import path in a Go program after adding the module as a dependency:

```sh
go get github.com/raimialiu/composer
```

For a program inside this repository, the same import resolves to the local module:

```go
package main

import (
	"context"
	"fmt"

	"github.com/raimialiu/composer"
)

type Student struct {
	Name   string
	Age    int
	Status string
}

func main() {
	seed := Student{Name: "Amina", Age: 19}
	student := composer.NewCompose[Student, string](seed).
		Create(func(in Student) Student {
			in.Status = "draft"
			return in
		}).
		With(func(in Student) Student {
			in.Status = "eligible"
			return in
		}).
		Build()

	placement := composer.NewCompose[Student, string](student).
		Decision().
		When(func(in Student) bool { return in.Age >= 18 },
			func(Student) string { return "adult program" }).
		Otherwise(func(Student) string { return "junior program" }).
		Build()

	label, err := placement.RunContext(student, context.Background())
	fmt.Printf("%+v | %s | %v\n", student, label, err)
}
```

Output:

```text
{Name:Amina Age:19 Status:eligible} | adult program | <nil>
```

## More usage examples

### Update a value

```go
updated := composer.NewCompose[Student, Student](student).
	Update(student).
	With(func(in Student) Student {
		in.Status = "enrolled"
		return in
	}).
	Build()
```

Pass the same value to `NewCompose` and `Update` for now: `Update` checks its argument for nil but builds from the value supplied to `NewCompose`.

### Run an action decision

```go
action := composer.NewActionDecisionBuilder[Student]().
	When(func(in Student) bool { return in.Age >= 18 }, func(in Student) {
		fmt.Println("Invite", in.Name, "to adult orientation")
	}).
	Otherwise(func(in Student) {
		fmt.Println("Invite", in.Name, "to junior orientation")
	}).
	Build()

if err := action.RunContext(student, context.Background()); err != nil {
	fmt.Println("action error:", err)
}
```

The first matching action runs. `WhenContext` and `OtherwiseContext` accept actions that receive a `context.Context` and return an error.

### Launch pipeline steps

```go
done := make(chan string, 2)
pipeline := (&composer.PipelineBuilder[Student]{}).
	Then(func(in Student, ctx context.Context) (Student, error) {
		done <- "checked " + in.Name
		return in, nil
	}).
	WhenAction(func(in Student) bool { return in.Age >= 18 }, func(in Student) {
		done <- "adult onboarding for " + in.Name
	}).
	Build()

pipeline.RunGo(student, context.Background())
fmt.Println(<-done)
fmt.Println(<-done)
```

This example assumes `student.Age >= 18`, so both callbacks send a message. `RunGo` starts each step in a separate goroutine and returns immediately. The two messages can arrive in either order. Every step receives the original `student`; returned values are not passed to another step. See [`examples/main.go`](examples/main.go) for `Choose`, nested pipelines, `RepeatWhile`, and a `sync.WaitGroup` used to wait for completion.

## API reference

### Function types

| Type | Signature | Purpose |
| --- | --- | --- |
| `Func[T, O]` | `func(T) O` | Synchronous transformation or predicate. |
| `Action[T]` | `func(T)` | Synchronous side effect. |
| `ContextFunc[T, O]` | `func(T, context.Context) (O, error)` | Context-aware transformation. |
| `ContextAction[T]` | `func(T, context.Context) error` | Context-aware side effect. |

### Object composition

| API | Behavior |
| --- | --- |
| `NewCompose[T, TResult](value T) *Compose[T, TResult]` | Holds the input value. `TResult` is used by `Decision`; object operations still return `T`. |
| `(*Compose).Create(factory Func[T, T]) *ObjectComposer[T]` | Starts an object composer whose factory receives the held value. |
| `(*Compose).Update(value T) *ObjectComposer[T]` | Starts with an identity factory over the held value. The `value` argument is nil-checked but otherwise unused. |
| `NewObjectComposer(factory Func[T, T], value T) *ObjectComposer[T]` | Creates an object composer directly. |
| `(*ObjectComposer).With(change Func[T, T]) *ObjectComposer[T]` | Appends a change. |
| `(*ObjectComposer).Transform(change Func[T, T]) *ObjectComposer[T]` | Appends a change; currently behaves like `With`. |
| `(*ObjectComposer).When(predicate Func[T, bool], change Func[T, T]) *ObjectComposer[T]` | Applies `change` when true. When false, the current implementation returns the zero value of `T`. |
| `(*ObjectComposer).WhenWithAdapter(predicate Func[T, bool], change Func[T, T], adapter ObjectAdapter[T]) *ObjectComposer[T]` | Calls `adapter.Apply` when true and keeps the input when false. `change` is required but currently not called. |
| `(*ObjectComposer).Build() T` | Calls the factory, then each appended change in order. |
| `ObjectAdapter[T]` | Interface with `Apply(value T) T`. |

### Value decisions

Create a builder with `NewCompose[T, TResult](value).Decision()`. A decision requires one `Otherwise` or `OtherwiseContext` fallback before `Build`.

| API | Behavior |
| --- | --- |
| `(*DecisionBuilder).When(predicate, action)` | Adds a synchronous branch. |
| `(*DecisionBuilder).WhenContext(predicate, action)` | Adds a context-aware branch that can return an error. |
| `(*DecisionBuilder).Otherwise(fallback)` | Sets the synchronous fallback. |
| `(*DecisionBuilder).OtherwiseContext(fallback)` | Sets the context-aware fallback. |
| `(*DecisionBuilder).Build() Decision[T, TResult]` | Copies the branches into a decision. |
| `(*Decision).RunContext(input T, ctx context.Context) (TResult, error)` | Checks branches in insertion order, runs the first match, or runs the fallback. Returns cancellation and action errors. |
| `NewDecision(branches, fallback)` | Creates a decision directly. `DecisionMapper` has private fields, so callers outside this package cannot populate branch entries directly. |

`Decision` also exposes `Run(ctx)` and `Apply(action)`. In the current implementation, both use an internal context value that the public constructors do not set; calling them on a normally built decision panics. Use `RunContext`.

### Action decisions

Create an action builder with `NewActionDecisionBuilder[T]()`. It has `When`, `WhenContext`, `Otherwise`, `OtherwiseContext`, and `Build` methods analogous to value decisions. `Build` returns `ActionDecision[T]`; call `RunContext(input, ctx) error` to execute it. Its `Run(input, ctx)` method currently delegates to `Decision.Run` and therefore panics for a normally built decision.

### Pipelines and tasks

Use `&composer.PipelineBuilder[T]{}` to start a builder, or `NewPipeline[T](steps...)` to create a pipeline directly. `Build()` returns `*Pipeline[T]`.

| Builder method | Behavior |
| --- | --- |
| `Then(step ContextFunc[T, T])` | Adds a context-aware step. |
| `ThenAction(step Action[T])` | Adds a synchronous action. |
| `ThenPipeline(pipeline Pipeline[T])` | Launches another pipeline as a step. |
| `Choose(decision Decision[T, T])` | Runs a value decision as a step. |
| `When(predicate Func[T, bool], step ContextFunc[T, Task[T]])` | Calls the step only when the predicate matches; reads the returned task's value. |
| `WhenAction(predicate Func[T, bool], step Action[T])` | Calls the action only when the predicate matches. |
| `WhenPipeline(predicate Func[T, bool], pipeline Pipeline[T])` | Launches the nested pipeline only when the predicate matches. |
| `RepeatWhile(step ContextFunc[T, Task[bool]], maxIterations Func[T, int])` | Calls the step exactly `maxIterations(input)` times unless an error or cancellation occurs. The returned bool is currently ignored. Limits below 1 panic. |

`(*Pipeline).RunGo(input, ctx)` launches every step independently in a goroutine, returns no result, and does not wait. Each step receives the same input. A step error causes a panic in its goroutine. `Task[T]` exposes `Value() T`, but its value field is private and there is no public constructor, so code outside `composer` can only create a zero-value task.

### Other exports

`ThrowIfNull(value any)` panics only when the interface itself is nil; a typed nil inside an interface passes. `GetFuncArg()` is currently an empty function.

## Current test state

`go run ./examples` works. `go test ./...` currently fails to compile because `decision_builder_test.go` refers to `StartDecisionBuilder`, `NewDecisionBuilder`, and `StartActionDecisionBuilder`, which are not defined in the package.
