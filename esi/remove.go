package esi

import (
	"bytes"
	"net/http"
)

const remove = "remove"

var closeRemove = []byte("</esi:remove>")

type removeTag struct {
	*baseTag
}

func (r *removeTag) Process(b []byte, req *http.Request) ([]byte, int) {
	end := closeEnd(b, closeRemove)
	if end == 0 {
		return []byte{}, len(b)
	}

	r.length = end

	return []byte{}, r.length
}

func (*removeTag) HasClose(b []byte) bool {
	return bytes.Contains(b, closeRemove)
}

func (*removeTag) GetClosePosition(b []byte) int {
	return closeEnd(b, closeRemove)
}
