package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func handleTaskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendErrorResponse(w, "The task ID is not specified")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendErrorResponse(w, "Invalid format of the task ID")
		return
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		sendErrorResponse(w, "Error connecting to database: "+err.Error())
		return
	}
	defer db.Close()

	var task Task
	err = db.QueryRow("SELECT * FROM scheduler WHERE id = ?", id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "Task not found")
		} else {
			sendErrorResponse(w, "Error receiving the task: "+err.Error())
		}
		return
	}

	if task.Repeat != "" {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, "Error in calculating the next date: "+err.Error())
			return
		}
		task.Date = nextDate

		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", task.Date, task.ID)
		if err != nil {
			sendErrorResponse(w, "Error updating the task: "+err.Error())
			return
		}
	} else {
		_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", task.ID)
		if err != nil {
			sendErrorResponse(w, "Error deleting a task: "+err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendErrorResponse(w, "The task ID is not specified")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendErrorResponse(w, "Invalid format of the task ID")
		return
	}
	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		sendErrorResponse(w, "Error connecting to database: "+err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		sendErrorResponse(w, "Error deleting a task:"+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
