package page

import (
	"time"
)

type Page struct {
	Url           string
	FirstDownload time.Time
	LastDownload  time.Time
	LastModified  time.Time
}
