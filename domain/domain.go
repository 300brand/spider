package domain

import (
	"github.com/300brand/spider/download"
	"github.com/300brand/spider/page"
	"github.com/temoto/robotstxt-go"
	"net/url"
	"time"
)

type Domain struct {
	Name        string
	URL         string
	Exclude     []string      // Regex path exclusions from config
	StartPoints []string      // Paths to being when link-spidering completes
	Delay       time.Duration // Delay between GETs to domain
	robotRules  *robotstxt.Group
	url         *url.URL
}

func (d *Domain) CanDownload(p *page.Page) bool {
	if d.robotRules == nil {
		d.UpdateRobotRules()
	}
	return d.robotRules.Test(p.GetURL().Path)
}

func (d *Domain) GetURL() *url.URL {
	if d.url == nil {
		d.url, _ = url.Parse(d.URL)
	}
	return d.url
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
