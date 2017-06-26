package natz

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var (
	urlRE = regexp.MustCompile(`(\S+)://((\S+):(\S+)@)?(\S+):(\d+)`)
)

type Config struct {
	Url      string
	Username string
	Password string
}

func (c Config) String() string {
	return c.Url
}

func Url(c Config) string {
	urls := []string{}

	for _, segment := range strings.Split(c.Url, ",") {
		match := urlRE.FindStringSubmatch(strings.TrimSpace(segment))
		if len(match) != 7 {
			log.Fatalln("invalid nats url")
		}

		proto, username, password, host, port := match[1], match[3], match[4], match[5], match[6]
		if c.Username != "" {
			username = c.Username
		}
		if c.Password != "" {
			password = c.Password
		}

		credentials := ""
		if username != "" && password != "" {
			credentials = fmt.Sprintf("%s:%s@", username, password)
		}

		urls = append(urls, fmt.Sprintf("%s://%s%s:%s", proto, credentials, host, port))
	}

	return strings.Join(urls, ", ")
}

func Subject(env, service, subject string) string {
	return env + "." + service + "." + subject
}
