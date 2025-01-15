package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/yandex-go-master/go_final_project/internal/database"
	"github.com/yandex-go-master/go_final_project/internal/nextdate"
)

const DateFormat = "20060102"

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

func NextDate(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	currentDate, err := time.Parse(DateFormat, now)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	nextDate, err := nextdate.NextDate(currentDate, date, repeat)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

func RootTask(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			PostTask(w, r, db)
		case http.MethodGet:
			GetTask(w, r, db)
		case http.MethodPut:
			UpdateTask(w, r, db)
		//case http.MethodDelete:
		//	handleDeleteTask(w, r, db)
		default:
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			log.Println("Method not allowed")
		}
	}
}

func PostTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		log.Println("Invalid request payload:", err)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error": "Field 'title' is required"}`, http.StatusBadRequest)
		log.Println("Field 'title' is required:", err)
		return
	}

	now := time.Now()
	var taskDate time.Time

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	} else if task.Date == now.Format(DateFormat) {
		taskDate = now
	} else {
		taskDate, err = time.Parse(DateFormat, task.Date)
		if err != nil {
			http.Error(w, `{"error": "Invalid date format"}`, http.StatusBadRequest)
			log.Println("Invalid date format:", err)
			return
		}
	}

	if taskDate.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(DateFormat)
		} else {
			nextDate, err := nextdate.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error": "Invalid next date"}`, http.StatusInternalServerError)
				log.Println("Invalid next date:", err)
				return
			}
			task.Date = nextDate
		}
	}

	taskId, err := database.AddTask(db, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Invalid task Id:", err)
		return
	}

	response := map[string]string{"id": strconv.Itoa(taskId)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetTasks(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	TasksLimit := 20

	rows, err := db.Query("SELECT * FROM scheduler ORDER BY date ASC LIMIT ?", TasksLimit)
	if err != nil {
		http.Error(w, `{"error": "Database select error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			http.Error(w, `{"error": "Rows scan error"}`, http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error": "meta error"}`, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []Task{}
	}

	response := map[string][]Task{"tasks": tasks}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "meta error"}`, http.StatusInternalServerError)
	}
}

func GetTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	var task Task
	query := "SELECT * FROM scheduler WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			//http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		http.Error(w, `{"error": "meta error"}`, http.StatusInternalServerError)
	}
}

func UpdateTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		log.Println("Invalid request payload:", err)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error": "Field 'title' is required"}`, http.StatusBadRequest)
		log.Println("Field 'title' is required:", err)
		return
	}

	now := time.Now()
	var taskDate time.Time

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	} else if task.Date == now.Format(DateFormat) {
		taskDate = now
	} else {
		taskDate, err = time.Parse(DateFormat, task.Date)
		if err != nil {
			http.Error(w, `{"error": "Invalid date format"}`, http.StatusBadRequest)
			log.Println("Invalid date format:", err)
			return
		}
	}

	if taskDate.Before(now) {
		if task.Repeat == "" {
			task.Date = now.Format(DateFormat)
		} else {
			nextDate, err := nextdate.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error": "Invalid next date"}`, http.StatusInternalServerError)
				log.Println("Invalid next date:", err)
				return
			}
			task.Date = nextDate
		}
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	if err != nil {
		http.Error(w, `{"error": "Task update error"}`, http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, `{"error": "Task update error"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}

	// Устанавливаем заголовок Content-Type и код состояния
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
