package samplesite

import (
	"io/ioutil"
	"launchpad.net/gocheck"
	"net/http"
	"testing"
)

type SamplesiteSuite struct{}

var _ = gocheck.Suite(new(SamplesiteSuite))

func Test(t *testing.T) { gocheck.TestingT(t) }

func (s *SamplesiteSuite) TestRobotsTxt(c *gocheck.C) {
	resp, err := http.Get(URL + "/robots.txt")
	c.Assert(err, gocheck.IsNil)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.IsNil)
	c.Assert(string(data), gocheck.Equals, string(robotsBody))
}

func (s *SamplesiteSuite) TestIndex(c *gocheck.C) {
	resp, err := http.Get(URL + "/latest")
	c.Assert(err, gocheck.IsNil)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.IsNil)
	latestBody := `<!DOCTYPE html>
<html>
<head>
	<title>Latest</title>
</head>
<body>
	<header>
		<h1>Latest</h1>
		<nav>
			<ul>
				<li><a href="/">Home</a></li>
				<li><a href="/latest">Latest News</a></li>
				<li><a href="/contact">Contact Us</a></li>
			</ul>
		</nav>
	</header>

	<article>

		<p><a href="/article1">/article1</a></p>

		<p><a href="/article2">/article2</a></p>

	</article>

	<footer>
		<nav>
			<ul>
				<li><a href="/">Home</a></li>
				<li><a href="/latest">Latest News</a></li>
				<li><a href="/contact">Contact Us</a></li>
			</ul>
		</nav>
	</footer>
</body>
</html>`
	c.Assert(string(data), gocheck.Equals, latestBody)
}
