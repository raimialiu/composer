package composer

import (
	"context"
	"errors"
	"testing"
)

func TestDecisionBuilder(t *testing.T) {
	builder := StartDecisionBuilder[int, string]().
		When(func(n int) bool { return n > 0 }, func(int) string { return "positive" }).
		When(func(n int) bool { return n > 10 }, func(int) string { return "large" }).
		Otherwise(func(int) string { return "other" })
	decision := builder.Build()

	if got := decision.Run(12, context.Background()); got != "positive" {
		t.Fatalf("first matching branch: got %q", got)
	}
	if got := decision.Run(0, context.Background()); got != "other" {
		t.Fatalf("fallback: got %q", got)
	}

	builder.When(func(n int) bool { return n == 0 }, func(int) string { return "zero" })
	if got := decision.Run(0, context.Background()); got != "other" {
		t.Fatalf("built decision changed after builder reuse: got %q", got)
	}
}

func TestDecisionBuilderContext(t *testing.T) {
	wantErr := errors.New("action failed")
	ctx := context.Background()
	decision := NewDecisionBuilder[int, string]().
		WhenContext(func(n int) bool { return n > 0 }, func(n int, received context.Context) (string, error) {
			if received != ctx {
				t.Error("action received a different context")
			}
			return "", wantErr
		}).
		Otherwise(func(int) string { return "fallback" }).
		Build()

	if _, err := decision.RunContext(1, ctx); !errors.Is(err, wantErr) {
		t.Fatalf("action error: got %v, want %v", err, wantErr)
	}
	if got, err := decision.RunContext(0, ctx); err != nil || got != "fallback" {
		t.Fatalf("fallback: got %q, %v", got, err)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := decision.RunContext(1, canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled decision: got %v", err)
	}
}

func TestDecisionBuilderRequiresOneFallback(t *testing.T) {
	for name, build := range map[string]func(){
		"missing": func() { NewDecisionBuilder[int, int]().Build() },
		"duplicate": func() {
			NewDecisionBuilder[int, int]().Otherwise(func(int) int { return 1 }).Otherwise(func(int) int { return 2 })
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected a panic")
				}
			}()
			build()
		})
	}
}

func TestActionDecisionBuilder(t *testing.T) {
	var called string
	decision := StartActionDecisionBuilder[int]().
		When(func(n int) bool { return n > 0 }, func(int) { called = "branch" }).
		OtherwiseContext(func(_ int, _ context.Context) error {
			called = "fallback"
			return nil
		}).
		Build()

	if err := decision.RunContext(1, context.Background()); err != nil || called != "branch" {
		t.Fatalf("branch: called %q, error %v", called, err)
	}
	if err := decision.RunContext(0, context.Background()); err != nil || called != "fallback" {
		t.Fatalf("fallback: called %q, error %v", called, err)
	}

	wantErr := errors.New("action failed")
	failing := NewActionDecisionBuilder[int]().
		WhenContext(func(int) bool { return true }, func(int, context.Context) error { return wantErr }).
		Otherwise(func(int) {}).
		Build()
	if err := failing.RunContext(1, context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("action error: got %v, want %v", err, wantErr)
	}
}
