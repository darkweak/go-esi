package esi

import (
	"bytes"
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

	vars = "vars"
)

var (
	interpretedVar   = regexp.MustCompile(`\$\((.+?)(\{(.+)\}(.+)?)?\)`)
	defaultExtractor = regexp.MustCompile(`\|('|")(.+?)('|")`)
	stringExtractor  = regexp.MustCompile(`('|")(.+)('|")`)

	varPrefix = []byte("$(")
	closeVars = []byte("</esi:vars>")
)

func parseVariables(b []byte, req *http.Request) string {
	interprets := interpretedVar.FindSubmatch(b)

	if interprets != nil {
		switch string(interprets[1]) {
		case httpAcceptLanguage:
			if strings.Contains(req.Header.Get("Accept-Language"), string(interprets[3])) {
				return "true"
			} else {
				return "false"
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

			return ""
		}
	} else {
		strs := stringExtractor.FindSubmatch(b)

		if len(strs) > 2 {
			return string(strs[2])
		}
	}

	return string(b)
}

type varsTag struct {
	*baseTag
}

// Input (e.g. comment text="This is a comment." />).
func (c *varsTag) Process(b []byte, req *http.Request) ([]byte, int) {
	closeStart, end := findClose(b, closeVars)
	if closeStart == -1 {
		return nil, len(b)
	}

	c.length = end

	// Skip the "vars>" remainder of the opening tag (the min builtin is unknown to Yaegi).
	contentStart := len(vars) + 1
	if contentStart > closeStart { //nolint:modernize // Yaegi (Traefik) does not support the min builtin.
		contentStart = closeStart
	}

	content := b[contentStart:closeStart]
	if !bytes.Contains(content, varPrefix) {
		return content, c.length
	}

	return interpretedVar.ReplaceAllFunc(content, func(b []byte) []byte {
		return []byte(parseVariables(b, req))
	}), c.length
}

func (*varsTag) HasClose(b []byte) bool {
	return bytes.Contains(b, closeVars)
}

func (*varsTag) GetClosePosition(b []byte) int {
	return closeEnd(b, closeVars)
}
