package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/yandex-go-master/go_final_project/internal/database"
	"github.com/yandex-go-master/go_final_project/internal/handlers"
	_ "modernc.org/sqlite"
)

func main() {
	log.Println("Starting up...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println(".env file loaded successfully")

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	log.Printf("Working port: %s", port)

	db := database.InitDb()
	defer db.Close()

	webDir := "./web"

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handlers.NextDate)
	http.HandleFunc("/api/task", handlers.RootTask(db))
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) { handlers.GetTasks(w, r, db) })
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) { handlers.DoneTask(w, r, db) })

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
