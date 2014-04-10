package page

import (
	"bytes"
	"errors"
	"github.com/300brand/logger"
	"github.com/300brand/spider/download"
	"github.com/PuerkitoBio/goquery"
	"hash/crc32"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

type Page struct {
	URL           string
	Title         string
	Checksum      uint32
	FirstDownload time.Time
	LastDownload  time.Time
	LastModified  time.Time
	url           *url.URL
	data          []byte
}

var ErrNotModified = errors.New("Not modified")

func New(rawurl string) (p *Page) {
	p = &Page{
		URL: rawurl,
	}
	return
}

func (p *Page) GetChecksum() uint32 {
	return crc32.ChecksumIEEE(p.data)
}

func (p *Page) Domain() string {
	d := p.GetURL().Host
	if strings.HasPrefix(d, "www.") {
		d = d[4:]
	}
	return d
}

func (p *Page) Download() (err error) {
	now := time.Now()
	resp, err := download.Get(p.URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	logger.Debug.Printf("%s Content-Type: %s", p.URL, contentType)

	if p.data, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	p.LastDownload = now
	if p.FirstDownload.IsZero() {
		p.FirstDownload = now
	}

	if sum := p.GetChecksum(); sum != p.Checksum {
		p.LastModified = now
		p.Checksum = sum
		return
	}

	return ErrNotModified
}

func (p *Page) GetURL() (u *url.URL) {
	if p.url != nil {
		return p.url
	}
	u, _ = url.Parse(p.URL)
	if u.Path == "" {
		u.Path = "/"
	}
	return
}

func (p *Page) Links() (links []string, err error) {
	d, err := goquery.NewDocumentFromReader(bytes.NewReader(p.data))
	if err != nil {
		return
	}

	base := p.GetURL()
	sel := d.Find("a[href]")
	links = make([]string, 0, sel.Length())
	sel.Each(func(i int, s *goquery.Selection) {
		// TODO add check for target attr
		refStr, exists := s.Attr("href")
		if !exists {
			return
		}
		ref, err := url.Parse(refStr)
		if err != nil {
			return
		}
		links = append(links, base.ResolveReference(ref).String())
	})
	return
}

func (p *Page) SetTitle() (err error) {
	d, err := goquery.NewDocumentFromReader(bytes.NewReader(p.data))
	if err != nil {
		return
	}

	p.Title = d.Find("title").First().Text()
	return
}
