package main

import (
	"regexp"
	"time"
)

type Filter struct {
	Ident       string
	Start       string
	CSSSelector string
	Restart     time.Duration
	Accept      []*regexp.Regexp
	Reject      []*regexp.Regexp
}

var filters = map[string]Filter{
	"www.tmcnet.com": Filter{
		Ident:       "tmcnet",
		Start:       "http://www.tmcnet.com/voip/",
		CSSSelector: "a[href]",
		Restart:     time.Minute,
		Accept: []*regexp.Regexp{
			regexp.MustCompile(`^/voip/(departments|columns|features)/articles/`),
		},
		Reject: []*regexp.Regexp{},
	},
	// "www.ccweek.com": Filter{
	// 	Start:       "http://www.ccweek.com/",
	// 	CSSSelector: "a[href]",
	// 	Accept: []*regexp.Regexp{
	// 		regexp.MustCompile(`^/article-\d{4,}`),
	// 	},
	// 	Reject: []*regexp.Regexp{},
	// },
	// "www.pipelinepub.com": Filter{
	// 	Start:       "http://www.pipelinepub.com/",
	// 	CSSSelector: ".leftsidelinks a[href]",
	// 	Accept:      []*regexp.Regexp{},
	// 	Reject:      []*regexp.Regexp{},
	// },
	// "www.healthmgttech.com": Filter{
	// 	Start:       "http://www.healthmgttech.com/",
	// 	CSSSelector: "a[href]",
	// 	Accept: []*regexp.Regexp{
	// 		regexp.MustCompile(`^/articles/\d{6}/`),
	// 		regexp.MustCompile(`/news/.`),
	// 		regexp.MustCompile(`/blogs/.`),
	// 		regexp.MustCompile(`/online-only/.`),
	// 	},
	// 	Reject: []*regexp.Regexp{
	// 		regexp.MustCompile(`/news/all-news.php`),
	// 		regexp.MustCompile(`/articles/\d{6}/toc.php`),
	// 	},
	// },
}
