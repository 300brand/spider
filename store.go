package main

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

type Store interface {
	Enqueue(url string) (err error)
	Next() (id interface{}, url string, err error)
	Save(cmd *Command) (err error)
}

type MySQL struct {
	Table string
	db    *sql.DB
	stmt  struct {
		Enqueue *sql.Stmt
		NextSel *sql.Stmt
		NextUp  *sql.Stmt
		Save    *sql.Stmt
	}
}

var ErrNoNext = errors.New("Nothing next in queue")

func DialMySQL(dsn string, name string) (db *MySQL, err error) {
	db = new(MySQL)
	db.Table = name
	if db.db, err = sql.Open("mysql", dsn); err != nil {
		return
	}
	err = db.prebuild()
	return
}

func (db *MySQL) Close() (err error) {
	return db.Close()
}

func (db *MySQL) Enqueue(url string) (err error) {
	_, err = db.stmt.Enqueue.Exec(url)
	return
}

func (db *MySQL) Next() (id interface{}, url string, err error) {
	if _, err = db.db.Exec(`LOCK TABLES ` + db.Table + ` WRITE`); err != nil {
		return
	}
	defer db.db.Exec(`UNLOCK TABLES`)
	var intId uint64
	if err = db.stmt.NextSel.QueryRow().Scan(&intId, &url); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoNext
		}
		return
	}
	id = intId
	if _, err = db.stmt.NextUp.Exec(intId); err != nil {
		return
	}
	return
}

func (db *MySQL) Save(cmd *Command) (err error) {
	if _, err = db.db.Exec(`LOCK TABLES ` + db.Table + ` WRITE`); err != nil {
		return
	}
	defer func() { _, err = db.db.Exec(`UNLOCK TABLES`) }()
	title := "No Title Yet"
	_, err = db.stmt.Save.Exec(cmd.URL().String(), title, cmd.Id)
	return
}

func (db *MySQL) prebuild() (err error) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS ` + db.Table + ` (
			id      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			url     TEXT NOT NULL DEFAULT '',
			title   TEXT NOT NULL DEFAULT '',
			queue  	ENUM('QUEUED', 'PROCESSED') DEFAULT 'QUEUED',
			requeue DATETIME NOT NULL,
			added   DATETIME NOT NULL,
			updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY (url(255))
		)`,
	}
	for _, query := range queries {
		if _, err = db.db.Exec(query); err != nil {
			return
		}
	}
	// Prepared statements for storage
	if db.stmt.Enqueue, err = db.db.Prepare(`INSERT IGNORE INTO ` + db.Table + ` (url, added) VALUES (?, NOW())`); err != nil {
		return
	}
	if db.stmt.NextSel, err = db.db.Prepare(`SELECT id, url FROM ` + db.Table + ` WHERE requeue < NOW() AND queue = 'QUEUED' ORDER BY id ASC LIMIT 1`); err != nil {
		return
	}
	if db.stmt.NextUp, err = db.db.Prepare(`UPDATE ` + db.Table + ` SET requeue = NOW() + INTERVAL 30 MINUTE WHERE id = ?`); err != nil {
		return
	}
	if db.stmt.Save, err = db.db.Prepare(`UPDATE ` + db.Table + ` SET queue = 'PROCESSED', url = ?, title = ? WHERE id = ?`); err != nil {
		return
	}
	return
}
