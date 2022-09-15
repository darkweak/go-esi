package esi

import (
	"net/http"
)

type (
	tag interface {
		process([]byte, *http.Request) ([]byte, int)
	}

	baseTag struct {
		length int
	}
)

func newBaseTag() *baseTag {
	return &baseTag{length: 0}
}

func (b *baseTag) process(content []byte, _ *http.Request) ([]byte, int) {
	return []byte{}, len(content)
}
