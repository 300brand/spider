package domain

import (
	"github.com/300brand/spider/page"
	"github.com/300brand/spider/samplesite"
	"testing"
)

func TestRobotsTxt(t *testing.T) {
	d := &Domain{
		URL: samplesite.URL,
	}
	tests := map[string]bool{
		"/"
	}

	p := &page.Page{
		URL: samplesite.URL
	}
	d.CanDownload(p)
}
