package download

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var ts *httptest.Server

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "test")
	})
	mux.HandleFunc("/ua", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("User-Agent"))
	})
	ts = httptest.NewServer(mux)
}

func TestGet(t *testing.T) {
	resp, err := Get(ts.URL + "/test")
	if err != nil {
		t.Errorf("Error calling Get: %s", err)
	}
	defer resp.Body.Close()

	var body, n = make([]byte, 64), 0
	if n, err = resp.Body.Read(body); err != nil {
		t.Errorf("Error reading body: %s", err)
	}

	if !bytes.Equal(body[:n], []byte(`test`)) {
		t.Errorf("Invalid body: '%s'", body)
	}
}

func TestUserAgent(t *testing.T) {
	resp, err := Get(ts.URL + "/ua")
	if err != nil {
		t.Errorf("Error calling Get: %s", err)
	}
	defer resp.Body.Close()

	var body, n = make([]byte, 64), 0
	if n, err = resp.Body.Read(body); err != nil {
		t.Errorf("Error reading body: %s", err)
	}

	if !bytes.Equal(body[:n], []byte(UserAgent)) {
		t.Errorf("Invalid body: '%s'", body)
	}
}
