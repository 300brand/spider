package page

import (
	"hash/crc32"
	"net/url"
	"time"
)

type Page struct {
	URL           string
	Checksum      uint32
	FirstDownload time.Time
	LastDownload  time.Time
	LastModified  time.Time
	url           *url.URL
	data          []byte
}

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
	return p.GetURL().Host
}

func (p *Page) GetURL() (u *url.URL) {
	if p.url != nil {
		return p.url
	}
	u, _ = url.Parse(p.URL)
	return
}
