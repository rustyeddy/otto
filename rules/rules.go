package rules

import (
	"context"
	"sync"
)

type Rule interface {
	Name() string
	Run(ctx context.Context) error
}

type Runner struct {
	rules []Rule
}

func NewRunner() *Runner { return &Runner{} }

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
