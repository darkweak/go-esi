package esi

import (
	"net/http"
)

type (
	Tag interface {
		Process(body []byte, request *http.Request) ([]byte, int)
		HasClose(body []byte) bool
		GetClosePosition(body []byte) int
	}

	baseTag struct {
		length int
	}
)

func newBaseTag() *baseTag {
	return &baseTag{length: 0}
}

func (b *baseTag) Process(content []byte, _ *http.Request) ([]byte, int) {
	return []byte{}, len(content)
}
