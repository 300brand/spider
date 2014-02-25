package samplesite

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSite(t *testing.T) {
	resp, err := http.Get(URL + "/robots.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, robotsBody) {
		t.Fatalf("Body mismatch, got: %s", data)
	}
}
