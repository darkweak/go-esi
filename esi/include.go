package esi

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
)

const include = "include"

var (
	closeInclude = regexp.MustCompile("/>")
	srcAttribute = regexp.MustCompile(`src="?(.+?)"?( |/>)`)
	altAttribute = regexp.MustCompile(`alt="?(.+?)"?( |/>)`)
)

type includeTag struct {
	*baseTag
	src string
	alt string
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

func sanitizeURL(u string, reqUrl *url.URL) string {
	parsed, _ := url.Parse(u)
	return reqUrl.ResolveReference(parsed).String()
}

// Input (e.g. include src="https://domain.com/esi-include" alt="https://domain.com/alt-esi-include" />)
// With or without the alt
// With or without a space separator before the closing
// With or without the quotes around the src/alt value.
func (i *includeTag) Process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := closeInclude.FindIndex(b)

	if closeIdx == nil {
		return nil, len(b)
	}

	i.length = closeIdx[1]
	if e := i.loadAttributes(b[8:i.length]); e != nil {
		return nil, len(b)
	}

	rq, _ := http.NewRequest(http.MethodGet, sanitizeURL(i.src, req.URL), nil)
	client := &http.Client{}
	response, err := client.Do(rq)

	if err != nil || response.StatusCode >= 400 {
		rq, _ = http.NewRequest(http.MethodGet, sanitizeURL(i.alt, req.URL), nil)
		response, err = client.Do(rq)

		if err != nil || response.StatusCode >= 400 {
			return nil, len(b)
		}
	}

	defer response.Body.Close()
	x, _ := io.ReadAll(response.Body)
	b = Parse(x, req)

	return b, i.length
}

func (*includeTag) HasClose(b []byte) bool {
	return closeInclude.FindIndex(b) != nil
}

func (*includeTag) GetClosePosition(b []byte) int {
	if idx := closeInclude.FindIndex(b); idx != nil {
		return idx[1]
	}
	return 0
}
