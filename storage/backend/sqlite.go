package backend

import (
	"database/sql"
	"github.com/300brand/spider/config"
	"github.com/300brand/spider/domain"
	"github.com/300brand/spider/page"
	"os"
	"path/filepath"
	"time"

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
	c.Domains = make([]domain.Domain, 0, 128)

	db, err := s.getDB("config")
	if err != nil {
		return
	}

	rows, err := db.Query(`SELECT domain, name, delay FROM domains`)
	if err != nil {
		return
	}

	var delay int64
	var subrows *sql.Rows
	var str string
	for rows.Next() {
		d := domain.Domain{
			Exclude:     make([]string, 0, 8),
			StartPoints: make([]string, 0, 8),
		}
		if err = rows.Scan(&d.URL, &d.Name, &delay); err != nil {
			return
		}
		d.Delay = time.Duration(delay)

		// Exclusion rules
		subrows, err = db.Query(`SELECT rule FROM excludes WHERE domain = ?`, d.URL)
		if err != nil {
			return
		}
		for subrows.Next() {
			if err = subrows.Scan(&str); err != nil {
				return
			}
			d.Exclude = append(d.Exclude, str)
		}
		if err = subrows.Err(); err != nil {
			return
		}

		// Start Points
		subrows, err = db.Query(`SELECT path FROM start_points WHERE domain = ?`, d.URL)
		if err != nil {
			return
		}
		for subrows.Next() {
			if err = subrows.Scan(&str); err != nil {
				return
			}
			d.StartPoints = append(d.StartPoints, str)
		}
		if err = subrows.Err(); err != nil {
			return
		}

		c.Domains = append(c.Domains, d)
	}

	return rows.Err()
}

func (s *Sqlite) GetPage(url string, p *page.Page) (err error) {
	p.URL = url
	db, err := s.getDB(p.Domain())
	if err != nil {
		return
	}

	var firstDownload, lastDownload, lastModified int64
	err = db.QueryRow(
		`
			SELECT
				url, first_download, last_download, last_modified, checksum
			FROM pages
			WHERE url = ?
			LIMIT 1
		`,
		url,
	).Scan(
		&p.URL,
		&firstDownload,
		&lastDownload,
		&lastModified,
		&p.Checksum,
	)
	if err == sql.ErrNoRows {
		p.URL = ""
		err = NotFound
	}
	p.FirstDownload = time.Unix(0, firstDownload)
	p.LastDownload = time.Unix(0, lastDownload)
	p.LastModified = time.Unix(0, lastModified)
	return
}

func (s *Sqlite) SaveConfig(c *config.Config) (err error) {
	db, err := s.getDB("config")
	if err != nil {
		return
	}

	// Mark everything for deletion
	if _, err = db.Exec(`UPDATE domains SET del = 1`); err != nil {
		return
	}

	// Update data
	exStmt, err := db.Prepare(`INSERT OR IGNORE INTO excludes (domain, rule) VALUES (?,?)`)
	if err != nil {
		return
	}
	defer exStmt.Close()

	spStmt, err := db.Prepare(`INSERT OR IGNORE INTO start_points (domain, path) VALUES (?,?)`)
	if err != nil {
		return
	}
	defer spStmt.Close()

	for _, d := range c.Domains {
		domain := d.GetURL().Scheme + "://" + d.GetURL().Host

		_, err = db.Exec(
			`INSERT OR REPLACE INTO domains
				(domain, name, delay, del)
			VALUES
				(?,      ?,    ?,     0  )
			`,
			domain,
			d.Name,
			d.Delay.Nanoseconds(),
		)
		if err != nil {
			return
		}

		// Domain Exclusion Rules
		if _, err = db.Exec(`DELETE FROM excludes WHERE domain = ?`, domain); err != nil {
			return
		}
		for _, ex := range d.Exclude {
			if _, err = exStmt.Exec(domain, ex); err != nil {
				return
			}
		}

		// Domain Start Points
		if _, err = db.Exec(`DELETE FROM start_points WHERE domain = ?`, domain); err != nil {
			return
		}
		for _, sp := range d.StartPoints {
			if _, err = spStmt.Exec(domain, sp); err != nil {
				return
			}
		}
	}

	// Delete previously marked rows
	if _, err = db.Exec(`DELETE FROM domains WHERE del = 1`); err != nil {
		return
	}

	return
}

func (s *Sqlite) SavePage(p *page.Page) (err error) {
	db, err := s.getDB(p.Domain())
	if err != nil {
		return
	}

	_, err = db.Exec(
		`
			INSERT OR REPLACE INTO pages
				(url, first_download, last_download, last_modified, checksum)
			VALUES
				(?,   ?,              ?,             ?,             ?   )
		`,
		p.URL,
		p.FirstDownload.UnixNano(),
		p.LastDownload.UnixNano(),
		time.Now().UnixNano(),
		p.Checksum,
	)
	return
}

func (s *Sqlite) getDB(name string) (db *sql.DB, err error) {
	if name == "" {
		return nil, NotFound
	}

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
			domain TEXT NOT NULL PRIMARY KEY,
			name   TEXT NOT NULL,
			delay  INTEGER,
			del    INTEGER DEFAULT 0
		)`,
		`CREATE TABLE excludes (
			domain TEXT NOT NULL,
			rule   TEXT NOT NULL,
			UNIQUE(domain, rule)
		)`,
		`CREATE TABLE start_points (
			domain TEXT NOT NULL,
			path   TEXT NOT NULL,
			UNIQUE(domain, path)
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
			first_download INTEGER NOT NULL,
			last_download  INTEGER NOT NULL,
			last_modified  INTEGER NOT NULL,
			checksum       INTEGER NOT NULL
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
