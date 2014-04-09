package storage

import (
	"github.com/300brand/logger"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/page"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

type Mongo struct {
	Database string
	Config   string
	Pages    string
	Exports  string
	session  *mgo.Session
	shard    bool
	sharded  bool
}

type mongoConfig struct {
	Id         string `bson:"_id"`
	LastUpdate time.Time
	Config     config.Config
}

type mongoExport struct {
	Id   bson.ObjectId `bson:"_id"`
	Key  string
	Len  int
	Time time.Time
}

type mongoPage struct {
	Id     bson.ObjectId `bson:",omitempty"`
	Url    string        `bson:"_id"`
	Domain string
	Page   page.Page
}

var _ Storage = new(Mongo)

func NewMongo(url string, shard bool) (m *Mongo, err error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return
	}
	m = &Mongo{
		Config:  "config",
		Pages:   "pages",
		Exports: "exports",
		session: s,
		shard:   shard,
	}
	if m.shard {
		err := m.session.DB("admin").Run(bson.M{
			"shardCollection": m.session.DB("").Name + ".config",
			"key":             bson.M{"_id": "hashed"},
		}, nil)
		if err != nil {
			logger.Error.Printf("sh.shardCollection(%s.config): %s", m.session.DB("").Name, err)
		}
		err = m.session.DB("admin").Run(bson.M{
			"shardCollection":  m.session.DB("").Name + ".exports",
			"key":              bson.M{"_id": "hashed"},
			"numInitialChunks": 16,
		}, nil)
		if err != nil {
			logger.Error.Printf("sh.shardCollection(%s.exports): %s", m.session.DB("").Name, err)
		}
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
	result := &mongoPage{
		Domain: p.Domain(),
	}
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

func (m *Mongo) GetPages(domain, key string, pages *[]*page.Page) (err error) {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB("").C(m.Exports)

	// Figure out how far back to go
	since := bson.ObjectIdHex("000000000000000000000000")
	if key != "" {
		v := new(mongoExport)
		if err = c.Find(bson.M{"key": key}).Sort("-_id").One(v); err != nil && err != mgo.ErrNotFound {
			return
		}
		if err != mgo.ErrNotFound {
			since = v.Id
		}
	}

	// Grab the goods
	s.DB("").C(m.cName(domain)).EnsureIndexKey("id")
	q := bson.M{
		"id": bson.M{
			"$gt": since,
		},
		"domain": domain,
	}
	mPages := make([]mongoPage, cap(*pages))
	if err = s.DB("").C(m.cName(domain)).Find(q).Limit(cap(mPages)).Sort("-id").All(&mPages); err != nil {
		logger.Error.Print(err)
		return
	}
	for i := range mPages {
		*pages = append(*pages, &mPages[i].Page)
	}

	// Set the flag back in the exports collection
	if len(mPages) == 0 {
		// ... unless nothing came out
		return
	}
	err = c.Insert(mongoExport{
		Id:   mPages[0].Id,
		Key:  key,
		Len:  len(mPages),
		Time: time.Now(),
	})
	return
}

func (m *Mongo) SavePage(p *page.Page) (err error) {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB("").C(m.cName(p.Domain()))

	u := p.GetURL().String()
	mp := mongoPage{
		Url:    u,
		Domain: p.Domain(),
		Page:   *p,
	}
	err = c.UpdateId(u, mp)
	if err == mgo.ErrNotFound {
		mp.Id = bson.NewObjectId()
		err = c.Insert(mp)
	}
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

func (m *Mongo) cName(domain string) (name string) {
	name = m.Pages + "_" + strings.Replace(domain, ".", "_", -1)
	if m.shard && !m.sharded {
		err := m.session.DB("admin").Run(bson.M{
			"shardCollection": m.session.DB("").Name + "." + name,
			"key":             bson.M{"_id": "hashed"},
		}, nil)
		if err != nil {
			logger.Error.Printf("sh.shardCollection(%s.%s): %s", m.session.DB("").Name, name, err)
		}
		m.sharded = true
	}
	return
}
