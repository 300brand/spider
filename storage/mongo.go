package storage

import (
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/page"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

type Mongo struct {
	session  *mgo.Session
	Database string
	Config   string
	Pages    string
}

type mongoConfig struct {
	Id         string `bson:"_id"`
	LastUpdate time.Time
	Config     config.Config
}

type mongoPage struct {
	Id   string `bson:"_id"`
	Page page.Page
}

var _ Storage = new(Mongo)

func NewMongo(url string) (m *Mongo, err error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return
	}
	m = &Mongo{
		session: s,
		Config:  "config",
		Pages:   "pages",
	}
	return
}

func (m *Mongo) Close() (err error) {
	m.session.Close()
	return
}

func (m *Mongo) GetPage(url string, p *page.Page) (err error) {
	s := m.session.Copy()
	defer s.Close()
	result := &mongoPage{}
	tmp := page.New(url)
	if err = s.DB("").C(m.cName(tmp.Domain())).FindId(url).One(result); err != nil {
		if err == mgo.ErrNotFound {
			err = ErrNotFound
		}
		return
	}
	*p = result.Page
	return
}

func (m *Mongo) SavePage(p *page.Page) (err error) {
	s := m.session.Copy()
	defer s.Close()
	mp := bson.M{
		"_id":  p.URL,
		"page": *p,
	}
	_, err = s.DB("").C(m.cName(p.Domain())).UpsertId(p.URL, mp)
	return
}

func (m *Mongo) GetConfig(c *config.Config) (err error) {
	s := m.session.Copy()
	defer s.Close()
	result := &mongoConfig{}
	if err = s.DB("").C(m.Config).FindId("config").One(result); err != nil && err != mgo.ErrNotFound {
		return
	}
	*c = result.Config
	return nil
}

func (m *Mongo) SaveConfig(c *config.Config) (err error) {
	s := m.session.Copy()
	defer s.Close()
	change := bson.M{
		"_id":        "config",
		"lastupdate": time.Now(),
		"config":     c,
	}
	_, err = s.DB("").C(m.Config).UpsertId("config", change)
	return
}

func (m *Mongo) cName(domain string) string {
	return m.Pages + "_" + strings.Replace(domain, ".", "_", -1)
}
