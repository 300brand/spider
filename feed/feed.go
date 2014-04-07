package feed

import (
	"encoding/xml"
	"github.com/300brand/spider/storage"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type Feed struct {
	store storage.Storage
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Item  []Item `xml:"item"`  // Optional. Stories within the feed
	Title string `xml:"title"` // Required. Defines the title of the channel
}

type Item struct {
	Guid    string    `xml:"guid"`    // Optional. Defines a unique identifier for the item
	Link    string    `xml:"link"`    // Required. Defines the hyperlink to the item
	PubDate time.Time `xml:"pubDate"` // Optional. Defines the last-publication date for the item
	Source  string    `xml:"source"`  // Optional. Specifies a third-party source for the item
	Title   string    `xml:"title"`   // Required. Defines the title of the item
}

var _ http.Handler = new(Feed)

func New(store storage.Storage) *Feed {
	return &Feed{
		store: store,
	}
}

func (f *Feed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := filepath.Base(r.RequestURI)
	key := r.FormValue("key")
	//pages := f.store.GetEntries(domain, key)
}
