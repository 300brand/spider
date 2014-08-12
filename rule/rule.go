package rule

import (
	"encoding/json"
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
	Restart  time.Duration
	Accept   []*regexp.Regexp
	Reject   []*regexp.Regexp
}

type LinkList struct {
	Accept []*url.URL
	Reject []*url.URL
	Ignore []*url.URL
}

type marshalRule struct {
	Ident       string
	Start       string
	CSSLinks    string
	CSSTitle    string
	RestartMins int
	Accept      []string
	Reject      []string
}

var (
	_ json.Marshaler   = new(Rule)
	_ json.Unmarshaler = new(Rule)
)

func (r *Rule) ExtractLinks(doc *goquery.Document, self *url.URL) (list LinkList, err error) {
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

func (r *Rule) ExtractTitle(doc *goquery.Document) (title string) {
	title = "No title tag"
	titleNode := doc.Find(r.CSSTitle)
	if titleNode.Length() > 0 {
		title = titleNode.First().Text()
	}
	return
}

func (r *Rule) MarshalJSON() (data []byte, err error) {
	mr := &marshalRule{
		Ident:       r.Ident,
		Start:       r.Start,
		CSSLinks:    r.CSSLinks,
		CSSTitle:    r.CSSTitle,
		RestartMins: int(r.Restart.Minutes()),
		Accept:      make([]string, len(r.Accept)),
		Reject:      make([]string, len(r.Reject)),
	}
	for i, re := range r.Accept {
		mr.Accept[i] = re.String()
	}
	for i, re := range r.Reject {
		mr.Reject[i] = re.String()
	}
	return json.Marshal(mr)
}

func (r *Rule) UnmarshalJSON(data []byte) (err error) {
	mr := new(marshalRule)
	if err = json.Unmarshal(data, mr); err != nil {
		return
	}

	if _, err = url.Parse(mr.Start); err != nil {
		return
	}

	r.Ident = mr.Ident
	r.Start = mr.Start
	r.CSSLinks = mr.CSSLinks
	r.CSSTitle = mr.CSSTitle
	r.Restart = time.Duration(mr.RestartMins) * time.Minute
	r.Accept = make([]*regexp.Regexp, len(mr.Accept))
	r.Reject = make([]*regexp.Regexp, len(mr.Reject))
	for i, expr := range mr.Accept {
		if r.Accept[i], err = regexp.Compile(expr); err != nil {
			return
		}
	}
	for i, expr := range mr.Reject {
		if r.Reject[i], err = regexp.Compile(expr); err != nil {
			return
		}
	}
	return
}
