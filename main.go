package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"main.go/repository"
	"net/http"
)

var repo *repository.Repository

func main() {
	port := "7540"
	webDir := "./web"

	var err error
	repo, err = repository.NewRepository("./scheduler.db")
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	server := chi.NewRouter()
	server.Mount("/", http.FileServer(http.Dir(webDir)))

	server.Get("/api/nextdate", handleNextDate())
	server.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleTaskGET(w, r)
		case http.MethodPost:
			handleTaskPOST(w, r)
		case http.MethodPut:
			handleTaskPUT(w, r)
		case http.MethodDelete:
			handleTaskDelete(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	server.HandleFunc("/api/task/done", handleTaskDone)
	server.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleTasksGET(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Printf("Server started successfully. Port: %s\n", port)
	err = http.ListenAndServe(":"+port, server)
	if err != nil {
		log.Fatal(err)
	}
}
