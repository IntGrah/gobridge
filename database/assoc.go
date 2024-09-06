package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

var Assoc Database

type Database struct {
	sqlDB *sql.DB
}

type Association struct {
	DC  string
	WA  string
	JID string
}

func (db Database) FromDc(dc string) (*Association, error) {
	var assoc Association
	rows := db.sqlDB.QueryRow("SELECT dc, wa, jid FROM assoc WHERE dc = ?", dc)
	if err := rows.Scan(&assoc.DC, &assoc.WA, &assoc.JID); err != nil {
		return nil, err
	}
	return &assoc, nil
}

func (db Database) FromWa(wa string) (*Association, error) {
	var assoc Association
	rows := db.sqlDB.QueryRow("SELECT dc, wa, jid FROM assoc WHERE wa = ?", wa)
	if err := rows.Scan(&assoc.DC, &assoc.WA, &assoc.JID); err != nil {
		return nil, err
	}
	return &assoc, nil
}

func (db Database) EditWa(waOld, waNew string) error {
	_, err := db.sqlDB.Exec("UPDATE assoc SET wa = ? WHERE wa = ?", waNew, waOld)
	return err
}

func (db Database) EditDc(dcOld, dcNew string) error {
	_, err := db.sqlDB.Exec("UPDATE assoc SET dc = ? WHERE dc = ?", dcNew, dcOld)
	return err
}

func (db Database) Put(assoc Association) error {
	log.Println("Associating", assoc.DC, assoc.WA, assoc.JID)
	_, err := db.sqlDB.Exec("INSERT INTO assoc (dc, wa, jid) VALUES (?, ?, ?)", assoc.DC, assoc.WA, assoc.JID)
	if err != nil {
		log.Println("Failed to associate", assoc, err)
	}
	return err
}

func (db Database) Delete(assoc *Association) error {
	log.Println("Deleting", assoc)
	_, err := db.sqlDB.Exec("DELETE FROM assoc WHERE dc = ? OR wa = ? or jid = ?", assoc.DC, assoc.WA, assoc.JID)
	return err
}

func NewMySQL() Database {
	fmt.Println("Connecting to MySQL")
	cfg := mysql.Config{
		User:                 os.Getenv("MYSQL_USER"),
		Passwd:               os.Getenv("MYSQL_PASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("MYSQL_HOST"),
		DBName:               os.Getenv("MYSQL_DATABASE"),
		AllowNativePasswords: true,
	}
	mysqlDB, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return Database{sqlDB: mysqlDB}
}
