package main

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/raimialiu/composer"
)

type Student struct {
	Name   string
	Age    int
	Track  string
	Status string
}

type honorsAdapter struct{}

func (honorsAdapter) Apply(student Student) Student {
	student.Track = "honors"
	return student
}

func main() {
	student := objectExamples()
	decisionExamples(student)
	pipelineExamples(student)
}

func objectExamples() Student {
	fmt.Println("OBJECT COMPOSITION")

	seed := Student{Name: "Amina", Age: 19}
	student := composer.NewCompose[Student, string](seed).
		Create(func(in Student) Student {
			in.Status = "draft"
			return in
		}).
		With(func(in Student) Student {
			in.Name = "Amina Yusuf"
			return in
		}).
		Transform(func(in Student) Student {
			in.Status = "submitted"
			return in
		}).
		// This example uses a matching predicate; a false predicate currently yields zero Student.
		When(func(in Student) bool { return in.Age >= 18 }, func(in Student) Student {
			in.Status = "eligible"
			return in
		}).
		WhenWithAdapter(
			func(in Student) bool { return in.Status == "eligible" },
			func(in Student) Student { return in }, // Required argument; the adapter does the work.
			honorsAdapter{},
		).
		Build()
	fmt.Printf("created: %+v\n", student)

	// Update currently starts from the value passed to NewCompose.
	updated := composer.NewCompose[Student, Student](student).
		Update(student).
		With(func(in Student) Student {
			in.Status = "enrolled"
			return in
		}).
		Build()
	fmt.Printf("updated: %+v\n\n", updated)
	return updated
}

func decisionExamples(student Student) {
	fmt.Println("DECISIONS")
	ctx := context.Background()

	placement := composer.NewCompose[Student, string](student).
		Decision().
		When(func(in Student) bool { return in.Age < 18 }, func(Student) string {
			return "junior program"
		}).
		WhenContext(func(in Student) bool { return in.Track == "honors" },
			func(in Student, ctx context.Context) (string, error) {
				return in.Name + " -> honors program", ctx.Err()
			}).
		OtherwiseContext(func(in Student, ctx context.Context) (string, error) {
			return in.Name + " -> general program", ctx.Err()
		}).
		Build()

	for _, candidate := range []Student{
		student,
		{Name: "Tobi", Age: 16},
		{Name: "Lara", Age: 20},
	} {
		result, err := placement.RunContext(candidate, ctx)
		fmt.Printf("placement for %s: %q (error: %v)\n", candidate.Name, result, err)
	}

	action := composer.NewActionDecisionBuilder[Student]().
		WhenContext(func(in Student) bool { return in.Track == "honors" },
			func(in Student, ctx context.Context) error {
				fmt.Printf("action for %s: invite to honors orientation\n", in.Name)
				return ctx.Err()
			}).
		Otherwise(func(in Student) {
			fmt.Printf("action for %s: send general orientation\n", in.Name)
		}).
		Build()
	for _, candidate := range []Student{student, {Name: "Lara", Age: 20}} {
		if err := action.RunContext(candidate, ctx); err != nil {
			fmt.Println("action error:", err)
		}
	}
	fmt.Println()
}

func pipelineExamples(student Student) {
	fmt.Println("PIPELINE")
	ctx := context.Background()
	var mu sync.Mutex
	var wg sync.WaitGroup
	var events []string
	record := func(event string) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
		wg.Done()
	}

	// RunGo starts each step in its own goroutine. Count every observable callback
	// before running, then wait so this example does not exit before they finish.
	wg.Add(9)
	subpipeline := composer.NewPipeline[Student](func(in Student, _ context.Context) (Student, error) {
		record("nested pipeline: send welcome packet")
		return in, nil
	})
	route := composer.NewCompose[Student, Student](student).
		Decision().
		When(func(in Student) bool { return in.Track == "honors" }, func(in Student) Student {
			record("choose: honors route")
			return in
		}).
		Otherwise(func(in Student) Student {
			record("choose: general route")
			return in
		}).
		Build()

	pipeline := (&composer.PipelineBuilder[Student]{}).
		Then(func(in Student, _ context.Context) (Student, error) {
			record("then: validate " + in.Name)
			return in, nil
		}).
		ThenAction(func(in Student) {
			record("then action: save enrollment")
		}).
		When(func(in Student) bool { return in.Age >= 18 },
			func(in Student, _ context.Context) (composer.Task[Student], error) {
				record("when: adult check")
				// Task's value is private, so external packages can only return its zero value.
				return composer.Task[Student]{}, nil
			}).
		WhenAction(func(in Student) bool { return in.Age >= 18 }, func(in Student) {
			record("when action: adult onboarding")
		}).
		Choose(route).
		// RepeatWhile currently calls the step exactly maxIterations times.
		RepeatWhile(func(in Student, _ context.Context) (composer.Task[bool], error) {
			record("repeat while: check documents")
			return composer.Task[bool]{}, nil
		}, func(Student) int { return 2 }).
		ThenPipeline(*subpipeline).
		WhenPipeline(func(in Student) bool { return in.Track == "honors" }, *subpipeline).
		Build()

	pipeline.RunGo(student, ctx)
	wg.Wait()
	sort.Strings(events) // Goroutine completion order is not guaranteed.
	for _, event := range events {
		fmt.Println(event)
	}
}
