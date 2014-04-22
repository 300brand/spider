package feed

import (
	"encoding/xml"
	"github.com/300brand/spider/page"
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
	domain := filepath.Base(r.URL.Path)
	key := r.FormValue("key")

	log.Printf("Domain:%s Key:%s", domain, key)

	pages := make([]*page.Page, 0, 100)
	if err := f.store.GetPages(domain, key, &pages); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rss := new(RSS)
	rss.Channel.Title = "RSS Feed for " + domain
	rss.Channel.Item = make([]Item, len(pages))
	for i := range pages {
		u := pages[i].GetURL()
		rss.Channel.Item[i] = Item{
			Guid:    u.String(),
			Link:    u.String(),
			PubDate: pages[i].FirstDownload,
			Source:  "CoverageSpider",
			// Title:   pages[i].Title,
		}
		// Knock down to whole seconds
		rss.Channel.Item[i].PubDate = rss.Channel.Item[i].PubDate.Add(time.Duration(-pages[i].FirstDownload.Nanosecond()))
	}

	enc := xml.NewEncoder(w)
	w.Write([]byte(xml.Header))
	enc.Indent("", "\t")
	enc.Encode(rss)
}
