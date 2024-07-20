package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func handleTasksGET(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error
	if dateStr != "" {
		date, err = time.Parse("20060102", dateStr)
		if err != nil {
			sendErrResponse(w, "Invalid date format: "+err.Error())
			return
		}
	}

	db, err := sql.Open("sqlite3", "./scheduler.db")
	if err != nil {
		sendErrResponse(w, "Error connecting to database: "+err.Error())
		return
	}
	defer db.Close()

	var query string
	var args []interface{}
	if !date.IsZero() {
		query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?"
		args = []interface{}{date.Format("20060102"), 50}
	} else {
		query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?"
		args = []interface{}{50}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		sendErrResponse(w, "Error getting tasks: "+err.Error())
		return
	}
	defer rows.Close()

	var tasks []map[string]string
	for rows.Next() {
		var id int64
		var date, title, comment, repeat string
		err = rows.Scan(&id, &date, &title, &comment, &repeat)
		if err != nil {
			sendErrResponse(w, "Error scanning task: "+err.Error())
			return
		}
		task := map[string]string{
			"id":      strconv.FormatInt(id, 10),
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		}
		tasks = append(tasks, task)
	}
	if tasks == nil {
		tasks = []map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func sendErrResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": message})
}
