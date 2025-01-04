package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func InitDb() {
	dbFileName := os.Getenv("TODO_DBFILE")
	if dbFileName == "" {
		dbFileName = "scheduler.db"
	}
	log.Printf("database filename: %s", dbFileName)

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), dbFileName)
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if install {
		_, err = db.Exec("CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date VARCHAR(8) NOT NULL DEFAULT '', title VARCHAR(128) NOT NULL DEFAULT '', comment TEXT NOT NULL DEFAULT '', repeat VARCHAR(128) NOT NULL DEFAULT '')")
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("CREATE INDEX idx_scheduler_date ON scheduler (date)")
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(dbFile)
	fmt.Println(install)

	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX
}
