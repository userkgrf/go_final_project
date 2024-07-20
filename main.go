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

	// Проверка наличия файла базы данных
	_, err = os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}

	// Открытие базы данных
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создание таблицы и индекса, если файл базы данных не найден
	if install {
		// Создание таблицы scheduler
		_, err = db.Exec(`
   CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    comment TEXT,
    repeat TEXT
   )
  `)
		if err != nil {
			log.Fatal(err)
		}

		// Создание индекса по полю date
		_, err = db.Exec(`CREATE INDEX date_index ON scheduler (date)`)
		if err != nil {
			log.Fatal(err)
		}
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
