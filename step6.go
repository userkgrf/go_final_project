package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func handleTaskGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		sendErrorResponse(w, "The ID is specified")
		return
	}
	id, err := strconv.Atoi(idParam)
	if err != nil {
		sendErrorResponse(w, "Invalid ID format")
		return
	}
	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		sendErrorResponse(w, "Database connection error: "+err.Error())
		return
	}
	defer db.Close()

	var task Task
	row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "The task was not found")
			return
		}
		sendErrorResponse(w, "Error receiving task data: "+err.Error())
		return
	}
	json.NewEncoder(w).Encode(task)
}

func handleTaskPUT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, "Error reading the request body: "+err.Error())
		return
	}
	defer r.Body.Close()

	var task Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		sendErrorResponse(w, "Error decoding the JSON request: "+err.Error())
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
			if task.Repeat == "" {
				task.Date = time.Now().Format("20060102")
			} else {
				task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					sendErrorResponse(w, err.Error())
					return
				}
			}
		}
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
		sendErrorResponse(w, "Database connection error:"+err.Error())
		return
	}
	defer db.Close()

	result, err := db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		sendErrorResponse(w, "Task update error:"+err.Error())
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		sendErrorResponse(w, "Error getting the number of modified rows:"+err.Error())
		return
	}

	if rowsAffected == 0 {
		sendErrorResponse(w, "Task not found")
		return
	}

	sendSuccessResp(w)
}

func sendSuccessResp(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
