package esi

import (
	"net/http"
	"regexp"
)

const remove = "remove"

var closeRemove = regexp.MustCompile("</esi:remove>")

type removeTag struct {
	*baseTag
}

func (r *removeTag) process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := closeRemove.FindIndex(b)
	if closeIdx == nil {
		return []byte{}, len(b)
	}
	r.length = closeIdx[1]

	return []byte{}, r.length
}
