package esi

import "regexp"

const (
	otherwise = "otherwise"
	try       = "try"
	vars      = "vars"
	when      = "when"
)

var (
	esi     = regexp.MustCompile("<esi:")
	tagname = regexp.MustCompile("^([a-z]+?)( |>)")

	// closeOtherwise = regexp.MustCompile("</esi:otherwise>")
	// closeTry       = regexp.MustCompile("</esi:try>")
	// closeVars      = regexp.MustCompile("</esi:vars>")
	// closeWhen      = regexp.MustCompile("</esi:when>")
)
