package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	port := os.Getenv("TODO_PORT")
	if len(port) == 0 {
		port = "7540"
	}

	webDir := "./web"

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
