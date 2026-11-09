package esi

import (
	"bytes"
	"net/http"
)

const escape = "<!--esi"

var closeEscape = []byte("-->")

type escapeTag struct {
	*baseTag
}

func (e *escapeTag) Process(b []byte, req *http.Request) ([]byte, int) {
	closeStart, end := findClose(b, closeEscape)
	if closeStart == -1 {
		return nil, len(b)
	}

	startPosition := 0
	for startPosition < closeStart && isSpace(b[startPosition]) {
		startPosition++
	}

	e.length = end

	return b[startPosition:closeStart], e.length
}

func (*escapeTag) HasClose(b []byte) bool {
	return bytes.Contains(b, closeEscape)
}

func (*escapeTag) GetClosePosition(b []byte) int {
	return closeEnd(b, closeEscape)
}
