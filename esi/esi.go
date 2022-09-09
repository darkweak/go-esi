package esi

import (
	"net/http"
)

func findTagName(b []byte) tag {
	name := tagname.FindSubmatch(b)
	if name == nil {
		return nil
	}

	switch string(name[1]) {
	case comment:
		return &commentTag{
			baseTag: newBaseTag(),
		}
	case choose:
		return &chooseTag{
			baseTag: newBaseTag(),
		}
	case escape:
		return &escapeTag{
			baseTag: newBaseTag(),
		}
	case include:
		return &includeTag{
			baseTag: newBaseTag(),
		}
	case remove:
		return &removeTag{
			baseTag: newBaseTag(),
		}
	case try:
	case vars:
		return &varsTag{
			baseTag: newBaseTag(),
		}
	default:
		return nil
	}

	return nil
}

func Parse(b []byte, req *http.Request) []byte {
	pointer := 0

	for pointer < len(b) {
		var escapeTag bool

		next := b[pointer:]
		tagIdx := esi.FindIndex(next)

		if escIdx := escapeRg.FindIndex(next); escIdx != nil && (tagIdx == nil || escIdx[0] < tagIdx[0]) {
			tagIdx = escIdx
			tagIdx[1] = escIdx[0]
			escapeTag = true
		}

		if tagIdx == nil {
			break
		}

		esiPointer := tagIdx[1]
		t := findTagName(next[esiPointer:])

		if escapeTag {
			esiPointer += 7
		}

		res, p := t.process(next[esiPointer:], req)
		esiPointer += p

		b = append(b[:pointer], append(next[:tagIdx[0]], append(res, next[esiPointer:]...)...)...)
		pointer += len(res) + tagIdx[0]
	}

	return b
}
