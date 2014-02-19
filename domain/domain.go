package domain

import (
	"net/http"
)

type Domain struct {
	Name string
	Url  string
}

func (d Domain) ReadRobotsTxt() (err error) {

	return
}
