package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := "7540"

	webDir := "./web" // Определяем путь к папке с фронтендом
	//	fmt.Println("Web directory:", webDir)

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "./scheduler.db")
	// Открываем базу данных
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Создаем таблицу и индекс, если они не существуют
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT,
            title TEXT,
            comment TEXT,
            repeat TEXT
        );
        CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
    `)

	if err != nil {
		log.Fatal(err)
	}

	server := chi.NewRouter()
	server.Mount("/", http.FileServer(http.Dir(webDir)))

	// Обрабатываем все запросы
	server.Get("/api/nextdate", handleNextDate)
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
	// Обработка GET-запросов к /api/tasks
	server.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleTasksGET(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Запуск сервера
	fmt.Printf("Server started successfully. Port: %s\n", port)
	err = http.ListenAndServe(":"+port, server)
	if err != nil {
		log.Fatal(err)
	}
}
