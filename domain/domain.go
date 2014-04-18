package domain

import (
	"errors"
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
	Include     []string      // Regex path inclusions from config
	Exclude     []string      // Regex path exclusions from config
	StartPoints []string      // Paths to being when link-spidering completes
	Delay       time.Duration // Delay between GETs to domain (15s)
	Redownload  time.Duration // Delay between re-downloading pages (3hr)
	domainName  string
	robotRules  *robotstxt.Group
	url         *url.URL
	reExclude   []*regexp.Regexp
	reInclude   []*regexp.Regexp
}

var (
	ErrTooSoon      = errors.New("Too soon to redownload")
	ErrRobot        = errors.New("Robots.txt denied access")
	ErrRegexExclude = errors.New("Excludes regexp matched")
	ErrRegexInclude = errors.New("Includes regexp did not match")
)

func FromURL(rawurl string) (name string) {
	u, _ := url.Parse(rawurl)
	if strings.HasPrefix(u.Host, "www.") {
		u.Host = u.Host[4:]
	}
	return u.Host
}

// Performs a few checks to determine if this page should be downloaded. Checks
// include:
// - Check if page last download is within the Redownload duration
// - Check if robots.txt blocks the page
// - Check if the page's URL is in the Exclude list
func (d *Domain) CanDownload(p *page.Page) (err error) {
	if p.LastDownload.After(time.Now().Add(-d.Redownload)) {
		return ErrTooSoon
	}

	// StartPoint check
	for i := range d.StartPoints {
		if d.StartPoints[i] == p.URL {
			return nil
		}
	}

	path := p.GetURL().Path

	if d.reInclude == nil || d.reExclude == nil {
		d.UpdateRegexpRules()
	}

	if len(d.Include) > 0 {
		include := false
		for i := range d.reInclude {
			if d.reInclude[i].MatchString(path) {
				include = true
			}
		}
		if !include {
			return ErrRegexInclude
		}
	}

	if d.robotRules == nil {
		d.UpdateRobotRules()
	}
	if !d.robotRules.Test(path) {
		return ErrRobot
	}

	for i := range d.reExclude {
		if d.reExclude[i].MatchString(path) {
			return ErrRegexExclude
		}
	}

	return nil
}

func (d *Domain) Domain() (domainName string) {
	if d.domainName == "" {
		d.domainName = FromURL(d.URL)
	}
	return d.domainName
}

func (d *Domain) GetURL() *url.URL {
	if d.url != nil {
		return d.url
	}
	d.url, _ = url.Parse(d.URL)
	if d.url.Path == "" {
		d.url.Path = "/"
	}
	return d.url
}

func (d *Domain) UpdateRegexpRules() {
	d.reExclude = d.buildRegexp(d.Exclude)
	d.reInclude = d.buildRegexp(d.Include)
}

func (d *Domain) buildRegexp(in []string) (out []*regexp.Regexp) {
	out = make([]*regexp.Regexp, len(in))
	for i := range in {
		out[i] = regexp.MustCompile(in[i])
	}
	return
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
