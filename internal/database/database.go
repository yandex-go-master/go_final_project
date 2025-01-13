package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func InitDb() (db *sql.DB) {
	dbFileName := os.Getenv("TODO_DBFILE")
	if dbFileName == "" {
		dbFileName = "scheduler.db"
	}
	log.Printf("Database filename: %s", dbFileName)

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

	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}

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

	fmt.Println(dbFile)  /////////////////////////////////////////////////////
	fmt.Println(install) /////////////////////////////////////////////////////

	return db
}

func AddTask(db *sql.DB, date, title, comment, repeat string) (int, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	log.Println("insert", date, title, comment, repeat)

	result, err := db.Exec(query, date, title, comment, repeat)
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("database insert error: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("task id select error: %v", err)
	}

	log.Println(id)
	return int(id), nil
}
