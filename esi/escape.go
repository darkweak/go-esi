package esi

import (
	"net/http"
	"regexp"
)

const escape = "<!--esi"

var (
	escapeRg    = regexp.MustCompile("<!--esi")
	closeEscape = regexp.MustCompile("((\n| +)+)?-->")
	startEscape = regexp.MustCompile("((\n| +)+)?")
)

type escapeTag struct {
	*baseTag
}

func (e *escapeTag) Process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := closeEscape.FindIndex(b)

	if closeIdx == nil {
		return nil, len(b)
	}

	startPosition := 0
	if startIdx := startEscape.FindIndex(b); startIdx != nil {
		startPosition = startIdx[1]
	}

	e.length = closeIdx[1]
	b = b[startPosition:closeIdx[0]]

	return b, e.length
}

func (*escapeTag) HasClose(b []byte) bool {
	return closeEscape.FindIndex(b) != nil
}

func (*escapeTag) GetClosePosition(b []byte) int {
	if idx := closeEscape.FindIndex(b); idx != nil {
		return idx[1]
	}

	return 0
}
