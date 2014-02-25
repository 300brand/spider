package domain

import (
	"github.com/300brand/spider/download"
	"net/url"
)

type Domain struct {
	Name        string
	URL         string
	Exclude     []string // Regex path exclusions from config
	StartPoints []string // Paths to being when link-spidering completes
	RobotRules  []string // Disallow: rules for us and all bots
	url         *url.URL
}

func (d *Domain) GetURL() *url.URL {
	if d.url == nil {
		d.url, _ = url.Parse(d.URL)
	}
	return d.url
}

func (d Domain) UpdateRobotRules() (err error) {
	robotstxt := d.GetURL().ResolveReference(&url.URL{Path: "/robots.txt"}).String()
	resp, err := download.Get(robotstxt)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}
