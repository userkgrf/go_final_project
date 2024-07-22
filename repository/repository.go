package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"final_project/taskRepRules"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title,omitempty" binding:"required"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) InsertTask(task *Task) (int64, error) {
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := r.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("error inserting task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting task ID: %w", err)
	}

	return id, nil
}

func (r *Repository) GetTasks(date time.Time, limit int) ([]Task, error) {
	var query string
	var args []interface{}
	if !date.IsZero() {
		query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?"
		args = []interface{}{date.Format("20060102"), limit}
	} else {
		query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?"
		args = []interface{}{limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var id int64
		var date, title, comment, repeat string
		err = rows.Scan(&id, &date, &title, &comment, &repeat)
		if err != nil {
			return nil, err
		}

		task := Task{
			ID:      fmt.Sprintf("%d", id), // Преобразование int64 в string
			Date:    date,
			Title:   title,
			Comment: comment,
			Repeat:  repeat,
		}

		tasks = append(tasks, task)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, nil
}

func (r *Repository) GetTask(id int) (*Task, error) {
	var task Task
	row := r.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("error receiving task data: %w", err)
	}
	return &task, nil
}

func (r *Repository) UpdateTask(task *Task) (int64, error) {
	result, err := r.db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return 0, fmt.Errorf("task update error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting the number of modified rows: %w", err)
	}

	if rowsAffected == 0 {
		return 0, fmt.Errorf("task not found")
	}

	return rowsAffected, nil
}

func (r *Repository) MarkTaskDone(id int64) error {
	var task Task
	err := r.db.QueryRow("SELECT * FROM scheduler WHERE id = ?", id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("task not found")
		}
		return fmt.Errorf("error receiving the task: %w", err)
	}

	if task.Repeat != "" {
		nextDate, err := taskRepRules.NextDate(time.Now(), task.Date, task.Repeat) // Assuming you have a NextDate function defined
		if err != nil {
			return fmt.Errorf("error in calculating the next date: %w", err)
		}
		task.Date = nextDate

		_, err = r.db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", task.Date, task.ID)
		if err != nil {
			return fmt.Errorf("error updating the task: %w", err)
		}
	} else {
		_, err = r.db.Exec("DELETE FROM scheduler WHERE id = ?", task.ID)
		if err != nil {
			return fmt.Errorf("error deleting a task: %w", err)
		}
	}
	return nil
}

func (r *Repository) DeleteTask(id int64) error {
	_, err := r.db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting a task: %w", err)
	}
	return nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}
