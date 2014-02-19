package download

import (
	"net/http"
	"net/http/cookiejar"
)

const UserAgent = "GoSpiderBot (+http://github.com/300brand/spider)"

var client = new(http.Client)

func init() {
	// Turns out cookiejar.New() returns a nil error
	client.Jar, _ = cookiejar.New(nil)
}

func Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}
