package swagger

import (
	"regexp"
	"sort"
	"strings"
)

var (
	pathRE = regexp.MustCompile(`(/:([^/]+))`)
)

func FixPath(path string) string {
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
