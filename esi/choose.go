package esi

import (
	"fmt"
	"net/http"
	"regexp"
)

const (
	choose = "choose"
)

var (
	closeChoose   = regexp.MustCompile("</esi:choose>")
	testAttribute = regexp.MustCompile(`test="(.+?)" ?>`)
)

type chooseTag struct {
	*baseTag
}

func matchTestAttribute(b []byte) bool {
	fmt.Println(string(b))

	return false
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
	tagIdx := esi.FindIndex(b[:found[1]])

	if tagIdx == nil {
		return []byte{}, len(b)
	}

	name := tagname.FindSubmatch(b[tagIdx[1]:found[1]])
	if name == nil || string(name[1]) != "when" {
		return []byte{}, len(b)
	}

	testAttr := testAttribute.FindSubmatch(b[tagIdx[1]:found[1]])
	if testAttr == nil {
		return nil, len(b)
	}

	matchTestAttribute(testAttr[1])

	// fmt.Println(string(name[1]), string(b[tagIdx[1]:found[1]]))

	return []byte{}, len(b)
}
