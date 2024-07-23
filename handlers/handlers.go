package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"final_project/repository"
	"final_project/taskRepRules"
)

type Handler struct {
	Repo *repository.Repository
}

type Response struct {
	Error string `json:"error,omitempty"`
}

const timeLayout = "20060102"
const maxTasksPerPage = 50

func HandleNextDate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nowStr := r.FormValue("now")
		dateStr := r.FormValue("date")
		repeat := r.FormValue("repeat")
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
		now, err := time.Parse(timeLayout, nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid 'now' format: %s", err), http.StatusBadRequest)
			return
		}
		_, err = time.Parse(timeLayout, dateStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid 'date' format: %s", err), http.StatusBadRequest)
			return
		}
		nextDate, err := taskRepRules.NextDate(now, dateStr, repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calculating next date: %s", err), http.StatusBadRequest)
			return
		}
		_, err = fmt.Fprintln(w, nextDate)
		if err != nil {
			log.Printf("Error writing response: %s", err.Error())
			return
		}

	}
}

func (h *Handler) HandleTaskPOST(w http.ResponseWriter, r *http.Request) {
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
		_, err = time.Parse(timeLayout, task.Date)
		if err != nil {
			sendErrorResponse(w, "Invalid 'date' format")
			return
		}
		parsedDate, _ := time.Parse(timeLayout, task.Date)
		if parsedDate.Before(time.Now()) {
			task.Date = time.Now().Format(timeLayout)
		}
	} else {
		task.Date = time.Now().Format(timeLayout)
	}
	if task.Repeat != "" {
		_, err = taskRepRules.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, err.Error())
			return
		}
	}

	id, err := h.Repo.InsertTask(&task)
	if err != nil {
		sendErrorResponse(w, "Error inserting task: "+err.Error())
		return
	}
	sendSuccessResponse(w, id)
}

func (h *Handler) HandleTasksGET(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error
	if dateStr != "" {
		date, err = time.Parse(timeLayout, dateStr)
		if err != nil {
			sendErrResponse(w, "Invalid date format: "+err.Error())
			return
		}
	}

	tasks, err := h.Repo.GetTasks(date, maxTasksPerPage)
	if err != nil {
		sendErrResponse(w, "Error getting tasks: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks}); err != nil {
		sendErrResponse(w, "Error encoding response: "+err.Error())
		return
	}
}

func (h *Handler) HandleTaskGET(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.Repo.GetTask(id)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		sendErrorResponse(w, "Error encoding response: "+err.Error())
		return
	}
}

func (h *Handler) HandlerMarkTaskDone(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "The task ID is not specified", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid format of the task ID", http.StatusBadRequest)
		return
	}
	err = h.Repo.MarkTaskDone(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK) // Возвращаем статутс 200 ОК с сообщением об успехе
	_, err = w.Write([]byte(`{"success": true, "message": "Task marked as done"}`))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) HandleTaskPUT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, "Error reading the request body: "+err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			sendErrorResponse(w, "Error closing the request body: "+err.Error())
			return
		}
	}(r.Body)

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
		parsedDate, err := time.Parse(timeLayout, task.Date)
		if err != nil {
			sendErrorResponse(w, "Invalid 'date' format")
			return
		}

		if parsedDate.Before(time.Now()) {
			if task.Repeat != "" {
				task.Date, err = taskRepRules.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					sendErrorResponse(w, err.Error())
					return
				}
			} else {
				task.Date = time.Now().Format(timeLayout)
			}
		}
	} else {
		task.Date = time.Now().Format(timeLayout)
	}

	if task.Repeat != "" {
		_, err = taskRepRules.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, err.Error())
			return
		}
	}

	rowsAffected, err := h.Repo.UpdateTask(&task)
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

func (h *Handler) HandleTaskDone(w http.ResponseWriter, r *http.Request) {
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

	err = h.Repo.MarkTaskDone(id)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{})
	if err != nil {
		sendErrorResponse(w, "Error encoding response: "+err.Error())
		return
	}
}

func (h *Handler) HandleTaskDelete(w http.ResponseWriter, r *http.Request) {
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
	err = h.Repo.DeleteTask(id)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{}); err != nil {
		sendErrorResponse(w, "Error encoding response: "+err.Error())
		return
	}
}

func sendSuccessResp(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	}); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func sendErrResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	errResponse := map[string]interface{}{"error": message}
	err := json.NewEncoder(w).Encode(errResponse)
	if err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, err string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)

	log.Printf("Error: %s", err)

	if err := json.NewEncoder(w).Encode(Response{Error: "Internal Server Error"}); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

func sendSuccessResponse(w http.ResponseWriter, id int64) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(map[string]interface{}{
		"success": true,
		"id":      id,
	})
	if err != nil {
		sendErrorResponse(w, "Error marshalling JSON response: "+err.Error())
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing JSON response: %s", err.Error())
	}
	fmt.Println("Sending response:", string(data))
}
