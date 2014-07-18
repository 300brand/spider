package main

import (
	"database/sql"

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
		Lock    *sql.Stmt
		NextSel *sql.Stmt
		NextUp  *sql.Stmt
		Save    *sql.Stmt
		Unlock  *sql.Stmt
	}
}

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
	if _, err = db.stmt.Lock.Exec(); err != nil {
		return
	}
	defer func() { _, err = db.stmt.Unlock.Exec() }()
	var intId uint64
	if err = db.stmt.NextSel.QueryRow().Scan(&intId, &url); err != nil {
		return
	}
	id = intId
	if _, err = db.stmt.NextUp.Exec(intId); err != nil {
		return
	}
	return
}

func (db *MySQL) Save(cmd *Command) (err error) {
	if _, err = db.stmt.Lock.Exec(); err != nil {
		return
	}
	defer func() { _, err = db.stmt.Unlock.Exec() }()
	title := "No Title Yet"
	_, err = db.stmt.Save.Exec(cmd.URL().String(), title, cmd.Id)
	return
}

func (db *MySQL) prebuild() (err error) {
	queries := []string{
		`CREATE TABLE ` + db.Table + ` (
			id      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			url     TEXT NOT NULL DEFAULT '',
			title   TEXT NOT NULL DEFAULT '',
			queue  	ENUM('HEAD', 'GET', 'PROCESSED'),
			requeue DATETIME NOT NULL,
			added   DATETIME NOT NULL,
			updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY (url)
		)`,
	}
	for _, query := range queries {
		if _, err = db.db.Exec(query); err != nil {
			return
		}
	}
	// Prepared statements for storage
	stmts := []struct {
		Target *sql.Stmt
		Query  string
	}{
		{
			db.stmt.Enqueue,
			`INSERT IGNORE INTO ` + db.Table + ` (url, added) VALUES (?, NOW())`,
		},
		{
			db.stmt.Lock,
			`LOCK TABLES ` + db.Table + ` WRITE`,
		},
		{
			db.stmt.NextSel,
			`SELECT id, url FROM ` + db.Table + ` WHERE requeue < NOW() AND queued = 0 ORDER BY id ASC LIMIT 1`,
		},
		{
			db.stmt.NextUp,
			`UPDATE ` + db.Table + ` SET requeue = NOW() + INTERVAL 30 MINUTE WHERE id = ?`,
		},
		{
			db.stmt.Save,
			`UPDATE ` + db.Table + ` SET queued = 0, url = ?, title = ? WHERE id = ?`,
		},
		{
			db.stmt.Unlock,
			`UNLOCK TABLES`,
		},
	}
	for _, stmt := range stmts {
		if stmt.Target, err = db.db.Prepare(stmt.Query); err != nil {
			return
		}
	}
	return
}
