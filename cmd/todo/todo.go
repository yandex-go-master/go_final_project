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
	log.Println("INFO: starting up...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("ERR: cannot load .env file")
	}
	log.Println("INFO: .env file loaded successfully")

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	log.Printf("INFO: working port: %s", port)

	database.InitDb()
	defer database.Db.Close()

	webDir := "./web"

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handlers.NextDate)
	http.HandleFunc("/api/task", handlers.Auth(handlers.RootTask()))
	http.HandleFunc("/api/tasks", handlers.Auth(handlers.GetTasks))
	http.HandleFunc("/api/task/done", handlers.Auth(handlers.DoneTask))
	http.HandleFunc("/api/signin", handlers.SignIn)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
