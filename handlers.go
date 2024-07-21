package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main.go/repository"
	"net/http"
	"strconv"
	"time"
)

func handleNextDate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nowStr := r.FormValue("now")
		dateStr := r.FormValue("date")
		repeat := r.FormValue("repeat")
		now, err := time.Parse("20060102", nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid 'now' format: %s", err), http.StatusBadRequest)
			return
		}
		if nowStr == "" {
			http.Error(w, "Missing 'now' parameter", http.StatusBadRequest)
			return
		}
		if dateStr == "" {
			http.Error(w, "Missing 'date' parameter", http.StatusBadRequest)
			return
		}
		if repeat == "" {
			http.Error(w, "Missing 'repeat' parameter", http.StatusBadRequest)
			return
		}
		_, err = time.Parse("20060102", dateStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid 'date' format: %s", err), http.StatusBadRequest)
			return
		}
		nextDate, err := repository.NextDate(now, dateStr, repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calculating next date: %s", err), http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, nextDate)
	}
}

func handleTaskPOST(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var task repository.Task
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
		_, err = repository.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, err.Error())
			return
		}
	}

	id, err := repo.InsertTask(&task)
	if err != nil {
		sendErrorResponse(w, "Error inserting task: "+err.Error())
		return
	}
	sendSuccessResponse(w, id)
}

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

	tasks, err := repo.GetTasks(date, 50)
	if err != nil {
		sendErrResponse(w, "Error getting tasks: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func handleTaskGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		sendErrorResponse(w, "The ID is not specified")
		return
	}
	id, err := strconv.Atoi(idParam)
	if err != nil {
		sendErrorResponse(w, "Invalid ID format")
		return
	}

	task, err := repo.GetTask(id)
	if err != nil {
		sendErrorResponse(w, err.Error())
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

	var task repository.Task
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
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			sendErrorResponse(w, "Invalid 'date' format")
			return
		}

		if parsedDate.Before(time.Now()) {
			if task.Repeat != "" {
				task.Date, err = repository.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					sendErrorResponse(w, err.Error())
					return
				}
			} else {
				task.Date = time.Now().Format("20060102")
			}
		}
	} else {
		task.Date = time.Now().Format("20060102")
	}

	if task.Repeat != "" {
		_, err = repository.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, err.Error())
			return
		}
	}

	rowsAffected, err := repo.UpdateTask(&task)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}
	if rowsAffected == 0 {
		sendErrorResponse(w, "Task not found")
		return
	}
	sendSuccessResp(w)
}

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

	err = repo.MarkTaskDone(id)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
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

	err = repo.DeleteTask(id)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func sendSuccessResp(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func sendErrResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": message})
}

func sendErrorResponse(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(repository.Response{Error: err})
}

func sendSuccessResponse(w http.ResponseWriter, id int64) {
	w.Header().Set("Content-Type", "application/json")
	// Add logging to see the JSON output
	data, err := json.Marshal(map[string]interface{}{
		"success": true,
		"id":      id,
	})
	if err != nil {
		sendErrorResponse(w, "Error marshalling JSON response: "+err.Error())
		return
	}
	fmt.Println("Sending response:", string(data)) // Log the response data
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
	})
}
