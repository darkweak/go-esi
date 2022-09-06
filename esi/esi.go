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
	case include:
		return &includeTag{
			baseTag: newBaseTag(),
		}
	case remove:
		return &removeTag{
			baseTag: newBaseTag(),
		}
	case otherwise:
	case try:
	case vars:
	case when:
	default:
		return nil
	}

	return nil
}

func Parse(b []byte, req *http.Request) []byte {
	pointer := 0

	for pointer < len(b) {
		next := b[pointer:]
		tagIdx := esi.FindIndex(next)

		if tagIdx == nil {
			break
		}

		esiPointer := tagIdx[1]
		t := findTagName(next[esiPointer:])

		res, p := t.process(next[esiPointer:], req)
		esiPointer += p

		b = append(b[:pointer], append(next[:tagIdx[0]], append(res, next[esiPointer:]...)...)...)
		pointer += len(res)
	}

	return b
}
