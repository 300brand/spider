package main

import (
	"github.com/PuerkitoBio/fetchbot"
	"net/http"
)

type HTTPClient struct{}

var _ fetchbot.Doer = new(HTTPClient)

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept-Encoding", "identity")
	return http.DefaultClient.Do(req)
}
