package download

import (
	"net/http"
	"net/http/cookiejar"
)

const (
	BotName   = "GoSpiderBot"
	BotURL    = "http://github.com/300brand/spider"
	UserAgent = BotName + " (+" + BotURL + ")"
)

var client = new(http.Client)

func init() {
	// Turns out cookiejar.New() returns a nil error
	client.Jar, _ = cookiejar.New(nil)
}

func Do(req *http.Request) (resp *http.Response, err error) {
	req.Header.Add("User-Agent", UserAgent)
	return client.Do(req)
}

func Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	return Do(req)
}
