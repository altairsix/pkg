package swagger

import (
	"regexp"
	"sort"
	"strings"

	"github.com/altairsix/pkg/web"
	"github.com/savaki/swag/endpoint"
)

var (
	pathRE = regexp.MustCompile(`(/:([^/]+))`)
)

func fixPath(path string) string {
	matches := pathRE.FindAllStringSubmatch(path, -1)

	// Sort matches by longest first to handle cases like /:a/:aa/:aaa
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i][0]) > len(matches[j][0])
	})

	for _, match := range matches {
		if len(match) == 3 {
			path = strings.Replace(path, match[1], "/{"+match[2]+"}", -1)
		}
	}
	return path
}

type pathParam struct {
	summary     string
	description string
	typ         string
	required    bool
}

type pathParamMap map[string]*pathParam

func (p pathParamMap) get(name string) *pathParam {
	item, ok := p[name]
	if !ok {
		description := "automatically generated"
		if len(name) > 0 {
			description = strings.ToUpper(name[0:1]) + name[1:] + "ID"
		}
		item = &pathParam{
			description: description,
			typ:         "string",
			required:    true,
		}
		p[name] = item
	}
	return item
}

type PathOption func(p pathParamMap)

func Path(name, description string, required bool) web.Option {
	return PathOption(func(p pathParamMap) {
		param := p.get(name)
		param.description = description
		param.required = required
	})
}

func pathOptions(path string, webOpts ...web.Option) []endpoint.Option {
	options := []endpoint.Option{}
	matches := pathRE.FindAllStringSubmatch(path, -1)

	params := pathParamMap{}
	for _, opt := range webOpts {
		if fn, ok := opt.(PathOption); ok {
			fn(params)
		}
	}

	for _, match := range matches {
		name := match[2]
		param := params.get(name)
		options = append(options, endpoint.Path(name, param.typ, param.description, param.required))
	}

	return options
}
