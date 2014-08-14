package main

import (
	"database/sql"
	"encoding/json"
	"github.com/300brand/spider/rule"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	db        *sql.DB
	lastCheck time.Time
	stmt      struct {
		Check *sql.Stmt
		Rules *sql.Stmt
	}
}

var configLastCheck time.Time

func DialConfig(dsn string) (db *Config, err error) {
	db = new(Config)
	if db.db, err = sql.Open("mysql", dsn); err != nil {
		return
	}
	err = db.prebuild()
	return
}

func (db *Config) Close() (err error) {
	return db.db.Close()
}

func (db *Config) Rules() (rules map[string]*rule.Rule, err error) {
	row := db.stmt.Check.QueryRow(configLastCheck)
	var count int
	if err = row.Scan(&count); err != nil || count == 0 {
		return
	}

	rows, err := db.stmt.Rules.Query()
	if err != nil {
		return
	}

	rules = make(map[string]*rule.Rule)
	for rows.Next() {
		var host string
		var data []byte
		if err = rows.Scan(&host, &data); err != nil {
			return
		}
		r := new(rule.Rule)
		if err = json.Unmarshal(data, r); err != nil {
			return
		}
		rules[host] = r
	}
	configLastCheck = time.Now()
	return
}

func (db *Config) prebuild() (err error) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS rules (
			id      BIGINT NOT NULL AUTO_INCREMENT,
			host    VARCHAR(128) NOT NULL DEFAULT '',
			json    TEXT NOT NULL,
			updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY (host)
		)`,
	}
	for _, query := range queries {
		if _, err = db.db.Exec(query); err != nil {
			return
		}
	}
	// Prepared statements for storage
	if db.stmt.Check, err = db.db.Prepare(`SELECT COUNT(*) FROM rules WHERE updated > ?`); err != nil {
		return
	}
	if db.stmt.Rules, err = db.db.Prepare(`SELECT host, json FROM rules`); err != nil {
		return
	}

	return
}
