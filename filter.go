package main

import (
	"regexp"
)

type Filter struct {
	CSSSelector string
	Accept      []*regexp.Regexp
	Reject      []*regexp.Regexp
}

var filters = map[string]Filter{
	"www.tmcnet.com": Filter{
		CSSSelector: "a[href]",
		Accept: []*regexp.Regexp{
			regexp.MustCompile(`^/voip/(departments|columns|features)/articles/`),
		},
		Reject: []*regexp.Regexp{},
	},
}
