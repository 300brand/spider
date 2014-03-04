package domain

import (
	"fmt"
	"github.com/300brand/spider/download"
	"github.com/300brand/spider/page"
	"github.com/temoto/robotstxt-go"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Domain struct {
	Name        string
	URL         string
	Exclude     []string      // Regex path exclusions from config
	StartPoints []string      // Paths to being when link-spidering completes
	Delay       time.Duration // Delay between GETs to domain
	domainName  string
	robotRules  *robotstxt.Group
	url         *url.URL
	reExclude   []*regexp.Regexp
}

func FromURL(rawurl string) (name string) {
	u, _ := url.Parse(rawurl)
	if strings.HasPrefix(u.Host, "www.") {
		u.Host = u.Host[4:]
	}
	return u.Host
}

func (d *Domain) CanDownload(p *page.Page) bool {
	path := p.GetURL().Path

	if d.robotRules == nil {
		d.UpdateRobotRules()
	}
	if !d.robotRules.Test(path) {
		return false
	}

	if d.reExclude == nil {
		d.UpdateRegexpRules()
	}
	for i := range d.reExclude {
		fmt.Printf("Testing %s against %s\n", path, d.Exclude[i])
		if d.reExclude[i].MatchString(path) {
			return false
		}
	}

	return true
}

func (d *Domain) Domain() (domainName string) {
	if d.domainName == "" {
		d.domainName = FromURL(d.URL)
	}
	return d.domainName
}

func (d *Domain) GetURL() *url.URL {
	if d.url == nil {
		d.url, _ = url.Parse(d.URL)
	}
	return d.url
}

func (d *Domain) UpdateRegexpRules() {
	d.reExclude = make([]*regexp.Regexp, len(d.Exclude))
	for i := range d.Exclude {
		d.reExclude[i] = regexp.MustCompile(d.Exclude[i])
	}
}

func (d *Domain) UpdateRobotRules() {
	var robots *robotstxt.RobotsData
	u := d.GetURL().ResolveReference(&url.URL{Path: "/robots.txt"}).String()
	resp, err := download.Get(u)
	if err != nil {
		goto AllowAll
	}
	defer resp.Body.Close()
	robots, err = robotstxt.FromResponse(resp)
	if err != nil {
		goto AllowAll
	}
	d.robotRules = robots.FindGroup(download.BotName)
	return

AllowAll:
	robots = &robotstxt.RobotsData{}
	d.robotRules = robots.FindGroup(download.BotName)
}
