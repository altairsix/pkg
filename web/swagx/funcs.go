package swagx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"time"

	"github.com/altairsix/pkg/web"
	"github.com/pkg/errors"
	"github.com/savaki/swag/endpoint"
	"github.com/savaki/swag/swagger"
)

func configInput(b *endpoint.Builder, t reflect.Type) bool {
	count := t.NumIn()
	if count > 3 {
		b.Endpoint.Summary = "Func functions may accept at most 2 input arguments"
		return false
	}

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	if count <= 1 || (count == 2 && t.In(1).Implements(contextType)) {
		b.Endpoint.Method = http.MethodGet

	} else if count == 2 || (count == 3 && t.In(1).Implements(contextType)) {
		b.Endpoint.Method = http.MethodPost
		body := t.In(t.NumIn() - 1)
		endpoint.BodyType(body, "body", true).Apply(b)

	} else {
		b.Endpoint.Summary = "*ERROR* input arguments must either be (), (context.Context), (context.Context, T)"
		return false
	}

	return true
}

func configOutput(b *endpoint.Builder, t reflect.Type) {
	count := t.NumOut()
	if count > 2 {
		b.Endpoint.Summary = "Func functions may return at most 2 values"
		return
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if count == 0 || (count == 1 && t.Out(0).Implements(errorType)) {
		return
	}

	if count == 2 && !t.Out(1).Implements(errorType) {
		b.Endpoint.Summary = "*ERROR* return values must either be error or (error, T)"
		return
	}

	out := t.Out(0)
	endpoint.ResponseType(http.StatusOK, out, "success").Apply(b)
}

// Func accepts a function and generates a swagger definition for function.
//
// Functions should accept zero or one input parameters in additional to an
// optional context.Context parameter.  Functions should return either error
// or a type plus error.
func Func(fn interface{}) endpoint.Option {
	return func(b *endpoint.Builder) {
		t := reflect.TypeOf(fn)
		if !configInput(b, t) {
			return
		}

		configOutput(b, t)
	}
}

func Handler(receiver reflect.Value, method reflect.Method, timeout time.Duration) web.HandlerFunc {
	inCount := method.Type.NumIn()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	return func(c web.Context) error {
		// construct input arguments
		//
		in := make([]reflect.Value, 0, inCount)
		in = append(in, receiver)
		if inCount >= 2 && method.Type.In(1).Implements(contextType) {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			in = append(in, reflect.ValueOf(ctx))
		}
		if argType := method.Type.In(inCount - 1); inCount >= 2 && !argType.Implements(contextType) {
			for argType.Kind() == reflect.Ptr {
				argType = argType.Elem()
			}

			input := reflect.New(argType).Interface()
			if body := c.Request().Body; body != nil {
				if err := json.NewDecoder(c.Request().Body).Decode(input); err != nil {
					return errors.Wrap(err, "unable to decode json")
				}
			}

			inputValue := reflect.ValueOf(input)
			in = append(in, inputValue)
		}

		out := method.Func.Call(in)
		if len(out) == 0 {
			return nil
		}
		last := out[len(out)-1].Interface()
		if err, ok := last.(error); ok && err != nil {
			return err
		}

		first := out[0].Interface()
		if _, ok := first.(error); ok {
			return nil
		}

		c.Response().Header().Set("Content-Type", "application/json")
		c.Response().WriteHeader(http.StatusOK)
		return json.NewEncoder(c.Response()).Encode(first)
	}
}

func Endpoints(prefix string, receiver interface{}, timeout time.Duration, options ...endpoint.Option) ([]*swagger.Endpoint, error) {
	t := reflect.TypeOf(receiver)
	if t.Kind() != reflect.Struct && (t.Kind() == reflect.Ptr && t.Elem().Kind() != reflect.Struct) {
		return nil, fmt.Errorf("Bind only accepts struct types")
	}

	endpoints := []*swagger.Endpoint{}

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		path := filepath.Join(prefix, method.Name)

		opts := append([]endpoint.Option(nil), options...)
		opts = append(opts,
			Func(method.Func.Interface()),
			endpoint.Handler(Handler(reflect.ValueOf(receiver), method, timeout)),
		)
		e := endpoint.New(http.MethodPost, path, "Automatically generated", opts...)
		endpoints = append(endpoints, e)
	}

	return endpoints, nil
}
