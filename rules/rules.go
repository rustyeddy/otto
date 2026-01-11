package rules

import (
	"context"
	"sync"
)

// Rule is a named runnable unit of behavior.
type Rule interface {
	Name() string
	Run(ctx context.Context) error
}

// Runner executes a set of rules concurrently.
type Runner struct {
	rules []Rule
}

// NewRunner creates an empty Runner.
func NewRunner() *Runner { return &Runner{} }

// Add registers a rule to run.
func (r *Runner) Add(rule Rule) { r.rules = append(r.rules, rule) }

// Run starts all rules and returns the first fatal error (or ctx cancellation).
func (r *Runner) Run(ctx context.Context) error {
	errCh := make(chan error, len(r.rules))
	var wg sync.WaitGroup

	for _, rule := range r.rules {
		wg.Add(1)
		go func(rule Rule) {
			defer wg.Done()
			if err := rule.Run(ctx); err != nil {
				errCh <- err
			}
		}(rule)
	}

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// graceful
	}

	wg.Wait()
	return nil
}
