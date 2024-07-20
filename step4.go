package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty" binding:"required"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type Response struct {
	Error string `json:"error,omitempty"`
}

func handleTaskPOST(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		sendErrorResponse(w, "Error decoding JSON request: "+err.Error())
		return
	}
	if task.Title == "" {
		sendErrorResponse(w, "The task title is not specified")
		return
	}
	if task.Date != "" {
		_, err = time.Parse("20060102", task.Date)
		if err != nil {
			sendErrorResponse(w, "Invalid 'date' format")
			return
		}
		parsedDate, _ := time.Parse("20060102", task.Date)
		if parsedDate.Before(time.Now()) {
			//		if task.Repeat == "" {
			task.Date = time.Now().Format("20060102")
		} /*else {
			task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				sendErrorResponse(w, err.Error())
				return
			}
		}*/
	} else {
		task.Date = time.Now().Format("20060102")
	}
	if task.Repeat != "" {
		_, err = NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, err.Error())
			return
		}
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		sendErrorResponse(w, "Error connecting to database: "+err.Error())
		return
	}
	defer db.Close()

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		sendErrorResponse(w, "Error inserting task: "+err.Error())
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		sendErrorResponse(w, "Error getting task ID: "+err.Error())
		return
	}

	sendSuccessResponse(w, id)
}

func sendErrorResponse(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(Response{Error: err})
}

func sendSuccessResponse(w http.ResponseWriter, id int64) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
	})
}
