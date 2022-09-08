package esi

import (
	"net/http"
	"regexp"
)

const (
	choose    = "choose"
	otherwise = "otherwise"
	when      = "when"
)

var (
	closeChoose   = regexp.MustCompile("</esi:choose>")
	whenRg        = regexp.MustCompile(`(?s)<esi:when test="(.+?)">(.+?)</esi:when>`)
	otherwiseRg   = regexp.MustCompile(`(?s)<esi:otherwise>(.+?)</esi:otherwise>`)
	testAttribute = regexp.MustCompile(`test="(.+?)" ?>`)
)

type chooseTag struct {
	*baseTag
}

// Input (e.g.
//  <esi:choose>
//    <esi:when test="$(HTTP_COOKIE{group})=='Advanced'">
//        <esi:include src="http://www.example.com/advanced.html"/>
//    </esi:when>
//    <esi:when test="$(HTTP_COOKIE{group})=='Basic User'">
//        <esi:include src="http://www.example.com/basic.html"/>
//    </esi:when>
//    <esi:otherwise>
//        <esi:include src="http://www.example.com/new_user.html"/>
//    </esi:otherwise>
//  </esi:choose>
//)
func (c *chooseTag) process(b []byte, req *http.Request) ([]byte, int) {
	found := closeChoose.FindIndex(b)
	if found == nil {
		return nil, len(b)
	}
	c.length = found[1]

	// first when esi tag
	tagIdxs := whenRg.FindAllSubmatch(b, -1)
	var res []byte
	for _, v := range tagIdxs {
		if validateTest(v[1], req) {
			res = Parse(v[2], req)
			break
		}
	}

	tagIdx := otherwiseRg.FindSubmatch(b)
	if tagIdx != nil {
		res = Parse(tagIdx[1], req)
	}

	return res, len(b)
}
