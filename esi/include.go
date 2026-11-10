package esi

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

const (
	include             = "include"
	maxIdleConnsPerHost = 64
)

var (
	closeInclude     = []byte("/>")
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

// client is shared by every include so connections are pooled and reused.
// Includes run in parallel, so keep more idle connections per host than the
// default (2) to avoid reconnecting on every page.
var client = &http.Client{Transport: newTransport()}

func newTransport() http.RoundTripper {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultTransport
	}

	t = t.Clone()
	t.MaxIdleConnsPerHost = maxIdleConnsPerHost

	return t
}

func fetch(src string, req *http.Request) (*http.Request, *http.Response, error) {
	//nolint:gosec // G704: fetching the include source is the purpose of the tag.
	rq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, sanitizeURL(src, req.URL), nil)
	if err != nil {
		return nil, nil, err
	}

	addHeaders(headersSafe, req, rq)

	if rq.URL.Scheme == req.URL.Scheme && rq.URL.Host == req.URL.Host {
		addHeaders(headersUnsafe, req, rq)
	}

	response, err := client.Do(rq) //nolint:gosec // G704: same as above.

	return rq, response, err
}

func failed(response *http.Response, err error) bool {
	return err != nil || response.StatusCode >= http.StatusBadRequest
}

// Input (e.g. include src="https://domain.com/esi-include" alt="https://domain.com/alt-esi-include" />)
// With or without the alt
// With or without a space separator before the closing
// With or without the quotes around the src/alt value.
func (i *includeTag) Process(b []byte, req *http.Request) ([]byte, int) {
	closeIdx := bytes.Index(b, closeInclude)
	if closeIdx == -1 {
		return nil, len(b)
	}

	i.length = closeIdx + len(closeInclude)
	if e := i.loadAttributes(b[8:i.length]); e != nil {
		return nil, len(b)
	}

	rq, response, err := fetch(i.src, req)
	if rq == nil {
		return nil, i.length
	}

	if failed(response, err) && i.alt != "" {
		if response != nil {
			_ = response.Body.Close()
		}

		var altRq *http.Request

		altRq, response, err = fetch(i.alt, rq)
		if altRq == nil {
			return nil, i.length
		}

		rq = altRq

		if !i.silent && failed(response, err) {
			if response != nil {
				_ = response.Body.Close()
			}

			return nil, len(b)
		}
	}

	if response == nil {
		return nil, i.length
	}

	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()

	if err != nil {
		return nil, i.length
	}

	return Parse(body, rq), i.length
}

func (*includeTag) HasClose(b []byte) bool {
	return bytes.Contains(b, closeInclude)
}

func (*includeTag) GetClosePosition(b []byte) int {
	if idx := bytes.Index(b, closeInclude); idx != -1 {
		return idx + len(closeInclude)
	}

	return 0
}
