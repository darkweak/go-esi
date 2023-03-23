package esi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

const include = "include"

var (
	closeInclude     = regexp.MustCompile("/>")
	srcAttribute     = regexp.MustCompile(`src="?(.+?)"?( |/>)`)
	altAttribute     = regexp.MustCompile(`alt="?(.+?)"?( |/>)`)
	onErrorAttribute = regexp.MustCompile(`onerror="?(.+?)"?( |/>)`)
)

// safe to pass to any origin.
var headersSafe = []string{
	"Accept",
	"Accept-Language",
}

// safe to pass only to same-origin (same scheme, same host, same port).
var headersUnsafe = []string{
	"Cookie",
	"Authorization",
}

type includeTag struct {
	*baseTag
	silent bool
	alt    string
	src    string
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

	onError := onErrorAttribute.FindSubmatch(b)
	if onError != nil {
		i.silent = string(onError[1]) == "continue"
	}

	return nil
}

func sanitizeURL(u string, reqURL *url.URL) string {
	parsed, _ := url.Parse(u)

	return reqURL.ResolveReference(parsed).String()
}

func addHeaders(headers []string, req *http.Request, rq *http.Request) {
	for _, h := range headers {
		v := req.Header.Get(h)
		if v != "" {
			rq.Header.Add(h, v)
		}
	}
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

	rq, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, sanitizeURL(i.src, req.URL), nil)
	addHeaders(headersSafe, req, rq)

	if rq.URL.Scheme == req.URL.Scheme && rq.URL.Host == req.URL.Host {
		addHeaders(headersUnsafe, req, rq)
	}

	client := &http.Client{}
	response, err := client.Do(rq)
	req = rq

	if (err != nil || response.StatusCode >= 400) && i.alt != "" {
		rq, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, sanitizeURL(i.alt, req.URL), nil)
		addHeaders(headersSafe, req, rq)

		if rq.URL.Scheme == req.URL.Scheme && rq.URL.Host == req.URL.Host {
			addHeaders(headersUnsafe, req, rq)
		}

		response, err = client.Do(rq)
		req = rq

		if !i.silent && (err != nil || response.StatusCode >= 400) {
			return nil, len(b)
		}
	}

	if response == nil {
		return nil, i.length
	}

	var buf bytes.Buffer

	defer response.Body.Close()
	_, _ = io.Copy(&buf, response.Body)

	b = Parse(buf.Bytes(), req)

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
