package esi

import "regexp"

const (
	choose    = "choose"
	comment   = "comment"
	escape    = "escape"
	include   = "include"
	remove    = "remove"
	otherwise = "otherwise"
	try       = "try"
	vars      = "vars"
	when      = "when"
)

var (
	esi     = regexp.MustCompile("<esi:")
	tagname = regexp.MustCompile("^([a-z]+?)( |>)")

	// closeChoose    = regexp.MustCompile("</esi:choose>")
	closeComment = regexp.MustCompile("/>")
	closeInclude = regexp.MustCompile("/>")
	closeRemove  = regexp.MustCompile("</esi:remove>")
	// closeOtherwise = regexp.MustCompile("</esi:otherwise>")
	// closeTry       = regexp.MustCompile("</esi:try>")
	// closeVars      = regexp.MustCompile("</esi:vars>")
	// closeWhen      = regexp.MustCompile("</esi:when>")

	srcAttribute = regexp.MustCompile(`src="?(.+?)"?( |/>)`)
	altAttribute = regexp.MustCompile(`alt="?(.+?)"?( |/>)`)
)
