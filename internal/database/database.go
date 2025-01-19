package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var Db *sql.DB

func InitDb() {
	dbFileName := os.Getenv("TODO_DBFILE")
	if dbFileName == "" {
		dbFileName = "scheduler.db"
	}
	log.Printf("INFO: Database filename: %s", dbFileName)

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), dbFileName)
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
		log.Println("INFO: Database file not found. Creating new database...")
	}

	Db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	if install {
		_, err = Db.Exec("CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date VARCHAR(8) NOT NULL DEFAULT '', title VARCHAR(128) NOT NULL DEFAULT '', comment TEXT NOT NULL DEFAULT '', repeat VARCHAR(128) NOT NULL DEFAULT '')")
		if err != nil {
			log.Fatal(err)
		}
		_, err = Db.Exec("CREATE INDEX idx_scheduler_date ON scheduler (date)")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("INFO: Database created successfully")
	}

	log.Printf("INFO: Database file path: %s", dbFile)
}

func AddTask(date, title, comment, repeat string) (int, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	result, err := Db.Exec(query, date, title, comment, repeat)
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("database insert error: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("task id select error: %v", err)
	}

	return int(id), nil
}
