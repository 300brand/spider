package backend

import (
	"github.com/300brand/spider/page"
	"testing"
)

func TestSqlite(t *testing.T) {
	b, err := NewSqlite("/tmp/testdb")
	if err != nil {
		t.Fatalf("New: %s", err)
	}
	defer b.Close()
	testBackend(t, b)
}

func testBackend(t *testing.T, b Backend) {
	url := "http://google.com/news.html"

	p := new(page.Page)
	if err := b.GetPage(url, p); err != NotFound {
		t.Fatalf("GetPage: Expected NotFound error, got: %q", err)
	}
	if p != nil {
		t.Fatalf("Pre-Store Should Not Exist: %s", url)
	}

	p.URL = url
	if err := b.SavePage(p); err != nil {
		t.Fatalf("SavePage: %q", err)
	}

	p.URL = ""
	if err := b.GetPage(url, p); err != nil {
		t.Fatalf("GetPage: %q", err)
	}
	if p.URL != url {
		t.Fatalf("GetPage: URLs do not match %s != %s", p.URL, url)
	}
}
