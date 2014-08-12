package rule

import (
	"github.com/300brand/logger"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"regexp"
	"time"
)

type Rule struct {
	Ident    string
	Start    string
	CSSLinks string
	CSSTitle string

	Restart time.Duration
	Accept  []*regexp.Regexp
	Reject  []*regexp.Regexp
}

type LinkList struct {
	Accept []*url.URL
	Reject []*url.URL
	Ignore []*url.URL
}

func (r Rule) ExtractLinks(doc *goquery.Document, self *url.URL) (list LinkList, err error) {
	var selection = doc.Find(r.CSSLinks)
	// Preallocate space for URLs
	list.Accept = make([]*url.URL, 0, selection.Length())
	list.Reject = make([]*url.URL, 0, selection.Length())
	list.Ignore = make([]*url.URL, 0, selection.Length())
	// Process each item
	selection.Each(func(i int, s *goquery.Selection) {
		var u *url.URL
		var bucket = 'I'
		var reUsed string

		// Assign to proper list
		defer func() {
			if u == nil {
				return
			}
			switch bucket {
			case 'A':
				list.Accept = append(list.Accept, u)
				logger.Trace.Printf("[%s] ACCEPT %s with %s", r.Ident, u, reUsed)
			case 'I':
				list.Ignore = append(list.Ignore, u)
				logger.Trace.Printf("[%s] IGNORE %s", r.Ident, u)
			case 'R':
				list.Reject = append(list.Reject, u)
				logger.Trace.Printf("[%s] REJECT %s with %s", r.Ident, u, reUsed)
			}
		}()

		// Resolve address
		val, _ := s.Attr("href")
		if u, err = self.Parse(val); err != nil {
			logger.Error.Printf("[%s] Resolve URL %s - %s", r.Ident, val, err)
			return
		}

		// Reject
		for _, re := range r.Reject {
			if re.MatchString(u.Path) {
				bucket, reUsed = 'R', re.String()
				return
			}
		}

		// Accept - if none, accept all
		if len(r.Accept) == 0 {
			bucket, reUsed = 'A', "*"
			return
		}

		// Accept - only accept matching
		for _, re := range r.Accept {
			if re.MatchString(u.Path) {
				bucket, reUsed = 'A', re.String()
				return
			}
		}
	})
	return
}

func (r Rule) ExtractTitle(doc *goquery.Document) (title string) {
	title = "No title tag"
	titleNode := doc.Find(r.CSSTitle)
	if titleNode.Length() > 0 {
		title = titleNode.First().Text()
	}
	return
}
