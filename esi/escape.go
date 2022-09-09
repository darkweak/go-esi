package esi

import (
	"net/http"
	"regexp"
)

const escape = "<!--esi"

var (
	escapeRg    = regexp.MustCompile("<!--esi")
	closeEscape = regexp.MustCompile("-->")
)

type escapeTag struct {
	*baseTag
}

func (e *escapeTag) process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := closeEscape.FindIndex(b)

	if closeIdx == nil {
		return nil, len(b)
	}

	e.length = closeIdx[1]
	b = b[:closeIdx[0]]

	return b, e.length
}
