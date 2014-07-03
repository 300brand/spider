package main

import (
	"github.com/PuerkitoBio/fetchbot"
	"net/url"
)

type Command struct {
	U     *url.URL
	M     string
	Depth int
}

var _ fetchbot.Command = new(Command)

func (c *Command) Method() string { return c.M }
func (c *Command) URL() *url.URL  { return c.U }
