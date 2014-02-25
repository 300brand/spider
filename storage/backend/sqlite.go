package backend

import (
	"database/sql"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/page"
	"os"
	"path/filepath"

	_ "code.google.com/p/gosqlite/sqlite3"
)

type Sqlite struct {
	dir string
	dbs map[string]*sql.DB
}

var _ Backend = new(Sqlite)

func NewSqlite(dir string) (s *Sqlite, err error) {
	if err = os.MkdirAll(dir, 0755); err != nil {
		return
	}
	dbs, err := filepath.Glob(filepath.Join(dir, "*.sqlite3"))
	if err != nil {
		return
	}
	s = &Sqlite{
		dir: dir,
		dbs: make(map[string]*sql.DB, len(dbs)),
	}
	for _, db := range dbs {
		base := filepath.Base(db)
		domain := base[:len(base)-len(".sqlite3")]
		if s.dbs[domain], err = sql.Open("sqlite3", db); err != nil {
			return
		}
	}
	return
}

func (s *Sqlite) Close() error {
	for _, db := range s.dbs {
		db.Close()
	}
	return nil
}

func (s *Sqlite) GetConfig(c *config.Config) (err error) {
	return
}

func (s *Sqlite) GetPage(url string, p *page.Page) (err error) {
	return
}

func (s *Sqlite) SaveConfig(c *config.Config) (err error) {
	return
}

func (s *Sqlite) SavePage(p *page.Page) (err error) {
	return
}

func (s *Sqlite) getDB(name string) (db *sql.DB, err error) {
	db, ok := s.dbs[name]
	if ok {
		return
	}
	if name == "config" {
		return s.configDB()
	}
	return s.domainDB(name)
}

func (s *Sqlite) configDB() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", filepath.Join(s.dir, "config.sqlite3"))
	if err != nil {
		return
	}
	creates := []string{
		`CREATE TABLE domains (
			domain_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name      TEXT NOT NULL,
			url       TEXT NOT NULL,
			delay     INTEGER 
		)`,
		`CREATE TABLE excludes (
			exclude_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			domain_id  INTEGER NOT NULL,
			rule       TEXT NOT NULL
		)`,
		`CREATE TABLE start_points (
			start_id  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER NOT NULL,
			path      TEXT NOT NULL
		)`,
	}
	for _, create := range creates {
		if _, err = db.Exec(create); err != nil {
			return
		}
	}
	s.dbs["config"] = db
	return
}

func (s *Sqlite) domainDB(name string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", filepath.Join(s.dir, name+".sqlite3"))
	if err != nil {
		return
	}
	creates := []string{
		`CREATE TABLE pages (
			page_id        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			url            TEXT NOT NULL UNIQUE,
			path           TEXT NOT NULL,
			title          TEXT NOT NULL,
			first_download TEXT NOT NULL,
			last_download  TEXT NOT NULL,
			last_modified  TEXT NOT NULL,
			hash           BLOB NOT NULL
		)`,
	}
	for _, create := range creates {
		if _, err = db.Exec(create); err != nil {
			return
		}
	}
	s.dbs[name] = db
	return
}
