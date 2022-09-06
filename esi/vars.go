package esi

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	httpAcceptLanguage = "HTTP_ACCEPT_LANGUAGE"
	httpCookie         = "HTTP_COOKIE"
	httpHost           = "HTTP_HOST"
	httpReferrer       = "HTTP_REFERER"
	httpUserAgent      = "HTTP_USER_AGENT"
	httpQueryString    = "QUERY_STRING"
)

var (
	interpretedVar   = regexp.MustCompile(`\$\((.+?)(\{(.+)\}(.+)?)?\)`)
	defaultExtractor = regexp.MustCompile(`\|('|")(.+?)('|")`)
)

func parseVariables(b []byte, req *http.Request) string {
	interprets := interpretedVar.FindSubmatch(b)

	if interprets != nil {
		switch string(interprets[1]) {
		case httpAcceptLanguage:
			if strings.Contains(req.Header.Get("Accept-Language"), string(interprets[3])) {
				return "true"
			}
		case httpCookie:
			if c, e := req.Cookie(string(interprets[3])); e == nil && c.Value != "" {
				return c.Value
			}
		case httpHost:
			return req.Host
		case httpReferrer:
			return req.Referer()
		case httpUserAgent:
			return req.UserAgent()
		case httpQueryString:
			if q := req.URL.Query().Get(string(interprets[3])); q != "" {
				return q
			}
		}
		if len(interprets) > 3 {
			defaultValues := defaultExtractor.FindSubmatch(interprets[4])

			if len(defaultValues) > 2 {
				return string(defaultValues[2])
			}
		}
	}
	return string(b)
}
