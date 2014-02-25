package page

import (
	"net/url"
	"time"
)

type Page struct {
	URL           string
	FirstDownload time.Time
	LastDownload  time.Time
	LastModified  time.Time
	url           *url.URL
}

func New(rawurl string) (p *Page) {
	p = &Page{
		URL: rawurl,
	}
	return
}

func (p *Page) Domain() string {
	return p.GetURL().Host
}

func (p *Page) GetURL() (u *url.URL) {
	if p.url != nil {
		return p.url
	}
	u, _ = url.Parse(p.URL)
	return
}
