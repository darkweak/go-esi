package esi

import (
	"bytes"
	"net/http"
)

var (
	esiPrefix     = []byte("<esi:")
	escapePrefix  = []byte("<!--esi")
	includePrefix = []byte("<esi:include")
)

type tagKind uint8

const (
	kindNone    tagKind = iota // no opening sequence found.
	kindESI                    // "<esi:" opening sequence.
	kindEscape                 // "<!--esi" opening sequence.
	kindPartial                // the buffer ends with an incomplete opening sequence.
)

// sink receives the output of scan in document order.
type sink interface {
	// static receives bytes that must be written as-is. The slice may alias
	// the scanned buffer and must not be retained after the call returns.
	static(b []byte)
	// tag receives a recognized tag and its body (the bytes between the
	// opening sequence and the end of the closing sequence). The body may
	// alias the scanned buffer and must not be retained after the call
	// returns.
	tag(t Tag, body []byte)
}

// nextTagStart returns the position of the first ESI opening sequence in b.
// When final is false and b ends with an incomplete opening sequence, its
// position is returned with kindPartial so the caller can wait for more data.
func nextTagStart(b []byte, final bool) (int, tagKind) {
	for i := 0; i < len(b); i++ {
		j := bytes.IndexByte(b[i:], '<')
		if j < 0 {
			break
		}

		i += j
		rest := b[i:]

		switch {
		case bytes.HasPrefix(rest, esiPrefix):
			return i, kindESI
		case bytes.HasPrefix(rest, escapePrefix):
			return i, kindEscape
		case !final && len(rest) < len(escapePrefix) &&
			(bytes.HasPrefix(esiPrefix, rest) || bytes.HasPrefix(escapePrefix, rest)):
			return i, kindPartial
		}
	}

	return -1, kindNone
}

// readTagName returns the tag name at the beginning of b and whether a
// non-name byte terminates it (i.e. the name cannot grow with more data).
func readTagName(b []byte) ([]byte, bool) {
	n := 0
	for n < len(b) && b[n] >= 'a' && b[n] <= 'z' {
		n++
	}

	return b[:n], n < len(b)
}

func newTag(name []byte) Tag {
	switch string(name) {
	case comment:
		return &commentTag{baseTag: newBaseTag()}
	case choose:
		return &chooseTag{baseTag: newBaseTag()}
	case escape:
		return &escapeTag{baseTag: newBaseTag()}
	case include:
		return &includeTag{baseTag: newBaseTag()}
	case remove:
		return &removeTag{baseTag: newBaseTag()}
	case vars:
		return &varsTag{baseTag: newBaseTag()}
	default:
		return nil
	}
}

// needsLookahead reports whether the closing sequence of t may extend beyond
// its first match (e.g. trailing whitespace), in which case a match ending at
// the very end of a non-final buffer must wait for more data.
func needsLookahead(t Tag) bool {
	_, ok := t.(*commentTag)

	return ok
}

// isAsync reports whether processing t may block on I/O and is therefore
// worth running concurrently.
func isAsync(t Tag, body []byte) bool {
	switch t.(type) {
	case *includeTag:
		return true
	case *chooseTag:
		return bytes.Contains(body, includePrefix)
	default:
		return false
	}
}

// scan walks b, sending static content and complete tags to s in document
// order, and returns the number of bytes consumed. When final is false it
// stops before the first construct that may still be incomplete so the caller
// can retry once more data is available.
func scan(b []byte, final bool, s sink) int {
	pos := 0

	for pos < len(b) {
		rest := b[pos:]

		i, kind := nextTagStart(rest, final)
		if kind == kindNone {
			s.static(rest)

			return len(b)
		}

		if i > 0 {
			s.static(rest[:i])
		}

		pos += i
		if kind == kindPartial {
			return pos
		}

		rest = rest[i:]

		var (
			t      Tag
			prefix int
		)

		if kind == kindEscape {
			t, prefix = &escapeTag{baseTag: newBaseTag()}, len(escapePrefix)
		} else {
			name, complete := readTagName(rest[len(esiPrefix):])
			if !complete && !final {
				return pos
			}

			if t = newTag(name); t == nil {
				// Unsupported tag: emit the opening sequence verbatim.
				s.static(rest[:len(esiPrefix)])
				pos += len(esiPrefix)

				continue
			}

			prefix = len(esiPrefix)
		}

		end := t.GetClosePosition(rest)
		if end == 0 || (!final && end == len(rest) && needsLookahead(t)) {
			if !final {
				return pos
			}

			// Unclosed tag at the end of the document: drop the remainder.
			return len(b)
		}

		s.tag(t, rest[prefix:end])
		pos += end
	}

	return pos
}

// process runs t on body. A panic in a tag implementation is contained so it
// cannot take down the whole process when running in its own goroutine.
func process(t Tag, body []byte, req *http.Request) (out []byte) {
	defer func() {
		if recover() != nil {
			out = nil
		}
	}()

	out, _ = t.Process(body, req)

	return out
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\n'
}

// findClose locates the first occurrence of lit in b, extended backwards over
// the spaces and newlines preceding it. It is the byte-level equivalent of the
// leftmost match of `((\n| +)+)?lit`, without the regexp engine cost.
func findClose(b, lit []byte) (start, end int) {
	i := bytes.Index(b, lit)
	if i == -1 {
		return -1, -1
	}

	start = i
	for start > 0 && isSpace(b[start-1]) {
		start--
	}

	return start, i + len(lit)
}

// closeEnd returns the end of the first occurrence of lit in b, or 0.
func closeEnd(b, lit []byte) int {
	if i := bytes.Index(b, lit); i != -1 {
		return i + len(lit)
	}

	return 0
}
