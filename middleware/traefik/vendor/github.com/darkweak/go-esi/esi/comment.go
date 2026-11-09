package esi

import (
	"bytes"
	"net/http"
)

const comment = "comment"

var closeComment = []byte("/>")

type commentTag struct {
	*baseTag
}

// commentEnd returns the end of the closing "/>" including the spaces and
// newlines following it, or 0.
func commentEnd(b []byte) int {
	end := closeEnd(b, closeComment)
	if end == 0 {
		return 0
	}

	for end < len(b) && isSpace(b[end]) {
		end++
	}

	return end
}

// Input (e.g. comment text="This is a comment." />).
func (c *commentTag) Process(b []byte, _ *http.Request) ([]byte, int) {
	end := commentEnd(b)
	if end == 0 {
		return nil, len(b)
	}

	return []byte{}, end
}

func (*commentTag) HasClose(b []byte) bool {
	return bytes.Contains(b, closeComment)
}

func (*commentTag) GetClosePosition(b []byte) int {
	return commentEnd(b)
}
