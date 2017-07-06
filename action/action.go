package action

import "context"

// Filter provides a wrapper around an action
type Filter func(Action) Action

// AndThen returns an action that executes the filter first and then the target Action
func (f Filter) AndThen(a Action) Action {
	return f(a)
}

// Action represents a cancelable unit of work
type Action func(ctx context.Context) error

// Do provides a helper method to make action invocations more legible
func (a Action) Do(ctx context.Context) error {
	return a(ctx)
}

// Use creates a new actions that uses all the specified filters
func (a Action) Use(filters ...Filter) Action {
	result := a
	for i := len(filters) - 1; i >= 0; i-- {
		result = filters[i].AndThen(result)
	}
	return result
}
