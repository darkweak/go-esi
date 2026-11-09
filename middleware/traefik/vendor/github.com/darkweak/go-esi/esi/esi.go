package esi

import (
	"bytes"
	"net/http"
	"sync"
)

func findTagName(b []byte) Tag {
	if bytes.HasPrefix(b, escapePrefix) {
		return &escapeTag{baseTag: newBaseTag()}
	}

	name, _ := readTagName(b)

	return newTag(name)
}

// HasOpenedTags reports whether b contains an ESI opening sequence.
func HasOpenedTags(b []byte) bool {
	_, kind := nextTagStart(b, true)

	return kind != kindNone
}

func CanProcess(b []byte) bool {
	if tag := findTagName(b); tag != nil {
		return tag.HasClose(b)
	}

	return false
}

// ReadToTag returns the position of the next ESI tag in next, the position
// right after its opening sequence and the matching Tag.
//
// Deprecated: use Parse or NewStream which handle the whole tag lifecycle.
func ReadToTag(next []byte, _ int) (startTagPosition, esiPointer int, t Tag) {
	i, kind := nextTagStart(next, true)
	if kind == kindNone {
		return len(next), 0, nil
	}

	if kind == kindEscape {
		return i, i + len(escapePrefix), &escapeTag{baseTag: newBaseTag()}
	}

	esiPointer = i + len(esiPrefix)

	return i, esiPointer, findTagName(next[esiPointer:])
}

// asyncPart is a tag processed concurrently whose output fills a slot of the
// final document.
type asyncPart struct {
	t    Tag
	body []byte
	out  []byte
}

type part struct {
	static []byte
	async  *asyncPart
}

// collector is the sink used by Parse. It runs I/O-bound tags concurrently,
// except the last one which is processed on the calling goroutine since it
// would wait anyway.
type collector struct {
	req     *http.Request
	parts   []part
	pending *asyncPart
	wg      sync.WaitGroup
}

func (c *collector) static(b []byte) {
	if len(b) > 0 {
		c.parts = append(c.parts, part{static: b})
	}
}

func (c *collector) tag(t Tag, body []byte) {
	if !isAsync(t, body) {
		c.static(process(t, body, c.req))

		return
	}

	a := &asyncPart{t: t, body: body}
	c.parts = append(c.parts, part{async: a})

	if p := c.pending; p != nil {
		c.wg.Go(func() { p.out = process(p.t, p.body, c.req) })
	}

	c.pending = a
}

// Parse processes the ESI tags of b and returns the resulting document.
// Independent include tags are fetched in parallel. b is not modified but the
// result may alias it.
func Parse(b []byte, req *http.Request) []byte {
	if !HasOpenedTags(b) {
		return b
	}

	c := collector{req: req}
	scan(b, true, &c)

	if p := c.pending; p != nil {
		p.out = process(p.t, p.body, req)
	}

	c.wg.Wait()

	size := 0
	for _, p := range c.parts {
		if p.async != nil {
			size += len(p.async.out)
		} else {
			size += len(p.static)
		}
	}

	out := make([]byte, 0, size)
	for _, p := range c.parts {
		if p.async != nil {
			out = append(out, p.async.out...)
		} else {
			out = append(out, p.static...)
		}
	}

	return out
}
