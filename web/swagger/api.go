package swagger

import (
	"github.com/altairsix/pkg/web"
	"github.com/savaki/swag/endpoint"
	"github.com/savaki/swag/swagger"
)

type optSummary struct {
	summary string
}

func Summary(in string) web.Option {
	return &optSummary{
		summary: in,
	}
}

type optDescription struct {
	description string
}

func Description(in string) web.Option {
	return &optDescription{
		description: in,
	}
}

type optIn struct {
	prototype   interface{}
	description string
	required    bool
}

func In(prototype interface{}, description string) web.Option {
	return &optIn{
		prototype:   prototype,
		description: description,
		required:    true,
	}
}

type optOut struct {
	code        int
	prototype   interface{}
	description string
}

func Out(code int, prototype interface{}, description string) web.Option {
	return &optOut{
		code:        code,
		prototype:   prototype,
		description: description,
	}
}

type optQuery struct {
	name        string
	typ         string
	description string
	required    bool
}

func Query(name, description string, required bool) web.Option {
	return &optQuery{
		name:        name,
		typ:         "string",
		description: description,
		required:    required,
	}
}

type optTags struct {
	tags []string
}

func Tags(tags ...string) web.Option {
	return &optTags{
		tags: tags,
	}
}

type API struct {
	Endpoints []*swagger.Endpoint
}

func New() *API {
	return &API{}
}

func (a *API) On(method, path string, webOpts ...web.Option) {
	summary := "Automatically generated"

	options := make([]endpoint.Option, 0, 8)
	for _, webOpt := range webOpts {
		switch v := webOpt.(type) {
		case *optSummary:
			summary = v.summary

		case *optDescription:
			options = append(options, endpoint.Description(v.description))

		case *optIn:
			options = append(options, endpoint.Body(v.prototype, v.description, v.required))

		case *optOut:
			options = append(options, endpoint.Response(v.code, v.prototype, v.description))

		case *optQuery:
			options = append(options, endpoint.Query(v.name, v.typ, v.description, v.required))

		case *optTags:
			options = append(options, endpoint.Tags(v.tags...))

		}
	}

	a.Endpoints = append(a.Endpoints, endpoint.New(method, FixPath(path), summary, options...))
}
