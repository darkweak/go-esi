package esi

import (
	"net/http"
	"regexp"
)

const comment = "comment"

var closeComment = regexp.MustCompile("/>")

type commentTag struct {
	*baseTag
}

// Input (e.g. comment text="This is a comment." />)
func (c *commentTag) process(b []byte, req *http.Request) ([]byte, int) {
	found := closeComment.FindIndex(b)
	if found == nil {
		return nil, len(b)
	}
	c.length = found[1]

	return []byte{}, len(b)
}
