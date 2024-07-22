package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"final_project/handlers"
	"final_project/repository"
)

func main() {
	port := "7540"
	webDir := "./web"
	var repo *repository.Repository
	var err error
	repo, err = repository.NewRepository("./scheduler.db")
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()
	handler := handlers.Handler{Repo: repo}

	server := chi.NewRouter()
	server.Mount("/", http.FileServer(http.Dir(webDir)))

	server.Get("/api/nextdate", handlers.HandleNextDate())
	server.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.HandleTaskGET(w, r)
		case http.MethodPost:
			handler.HandleTaskPOST(w, r)
		case http.MethodPut:
			handler.HandleTaskPUT(w, r)
		case http.MethodDelete:
			handler.HandleTaskDelete(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	server.HandleFunc("/api/task/done", handler.HandleTaskDone)
	server.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.HandleTasksGET(w, r)
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
