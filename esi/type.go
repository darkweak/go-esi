package esi

import (
	"io"
	"net/http"
	"sync"
)

var clientPool = &sync.Pool{
	New: func() any {
		return &http.Client{}
	},
}

type (
	tag interface {
		process([]byte, *http.Request) ([]byte, int)
	}

	baseTag struct {
		length int
	}
	includeTag struct {
		*baseTag
		src string
		alt string
	}
	removeTag struct {
		*baseTag
	}
	commentTag struct {
		*baseTag
	}
)

func newBaseTag() *baseTag {
	return &baseTag{}
}

func (b *baseTag) process(content []byte, _ *http.Request) ([]byte, int) {
	return []byte{}, len(content)
}

func (i *includeTag) loadAttributes(b []byte) error {
	src := srcAttribute.FindSubmatch(b)
	if src == nil {
		return errNotFound
	}
	i.src = string(src[1])

	alt := altAttribute.FindSubmatch(b)
	if alt != nil {
		i.alt = string(alt[1])
	}

	return nil
}

// Input (e.g. include src="https://domain.com/esi-include" alt="https://domain.com/alt-esi-include" />)
// With or without the alt
// With or without a space separator before the closing
// With or without the quotes around the src/alt value
func (i *includeTag) process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := closeInclude.FindIndex(b)

	if closeIdx == nil {
		return nil, len(b)
	}

	i.length = closeIdx[1]
	if e := i.loadAttributes(b[8:i.length]); e != nil {
		return nil, len(b)
	}

	rq, _ := http.NewRequest(http.MethodGet, i.src, nil)
	response, err := clientPool.Get().(*http.Client).Do(rq)
	if err != nil || response.StatusCode >= 400 {
		rq, _ = http.NewRequest(http.MethodGet, i.src, nil)
		response, err = clientPool.Get().(*http.Client).Do(rq)
		if err != nil || response.StatusCode >= 400 {
			return nil, len(b)
		}
	}

	x, _ := io.ReadAll(response.Body)
	b = Parse(x, req)

	return b, i.length
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

func (r *removeTag) process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := closeRemove.FindIndex(b)
	if closeIdx == nil {
		return []byte{}, len(b)
	}
	r.length = closeIdx[1]

	return []byte{}, r.length
}
