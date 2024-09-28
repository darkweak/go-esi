package esi

import (
	"net/http"
	"slices"
	"sync"
)

func findTagName(b []byte) Tag {
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

func HasOpenedTags(b []byte) bool {
	return esi.FindIndex(b) != nil || escapeRg.FindIndex(b) != nil
}

func ReadToTag(next []byte, pointer int) (startTagPosition, esiPointer int, t Tag) {
	var isEscapeTag bool

	tagIdx := esi.FindIndex(next)

	if escIdx := escapeRg.FindIndex(next); escIdx != nil && (tagIdx == nil || escIdx[0] < tagIdx[0]) {
		tagIdx = escIdx
		tagIdx[1] = escIdx[0]
		isEscapeTag = true
	}

	if tagIdx == nil {
		return len(next), 0, nil
	}

	esiPointer = tagIdx[1]
	startTagPosition = tagIdx[0]
	t = findTagName(next[esiPointer:])

	if isEscapeTag {
		esiPointer += 7
	}

	return
}

func Parse(b []byte, req *http.Request) []byte {
	pointer := 0
	includes := make(map[int]Tag)

	for pointer < len(b) {
		next := b[pointer:]
		tagIdx := esi.FindIndex(next)

		if escIdx := escapeRg.FindIndex(next); escIdx != nil && (tagIdx == nil || escIdx[0] < tagIdx[0]) {
			tagIdx = escIdx
			tagIdx[1] = escIdx[0]
		}

		if tagIdx == nil {
			break
		}

		t := findTagName(next[tagIdx[1]:])

		if t != nil {
			includes[pointer+tagIdx[0]] = t
		}

		switch t.(type) {
		case *escapeTag:
			pointer += tagIdx[1] + 7
		default:
			pointer += tagIdx[1]
		}
	}

	if len(includes) == 0 {
		return b
	}

	return processEsiTags(b, req, includes)
}

func processEsiTags(b []byte, req *http.Request, includes map[int]Tag) []byte {
	type Response struct {
		Start int
		End   int
		Res   []byte
		Tag   Tag
	}

	var wg sync.WaitGroup

	responses := make(chan Response, len(includes))

	for start, t := range includes {
		wg.Add(1)

		go func(s int, tag Tag) {
			defer wg.Done()

			data, length := tag.Process(b[s:], req)

			responses <- Response{
				Start: s,
				End:   s + length,
				Res:   data,
				Tag:   tag,
			}
		}(start, t)
	}

	wg.Wait()
	close(responses)

	// we need to replace from the bottom to top
	sorted := make([]Response, 0, len(responses))
	for r := range responses {
		sorted = append(sorted, r)
	}

	slices.SortFunc(sorted, func(a, b Response) int {
		return b.Start - a.Start
	})

	for _, r := range sorted {
		b = append(b[:r.Start], append(r.Res, b[r.End:]...)...)
	}

	return b
}
