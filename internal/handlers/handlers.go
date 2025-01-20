package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandex-go-master/go_final_project/internal/database"
	"github.com/yandex-go-master/go_final_project/internal/nextdate"
)

const DateFormat = "20060102"
const webuiDateFormat = "02.01.2006"
const TasksLimit = 20
const tokenSecretKey = "abcdefghijklmnop"

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type UserPassword struct {
	Password string `json:"password"`
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

func RootTask() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			PostTask(w, r)
		case http.MethodGet:
			GetTask(w, r)
		case http.MethodPut:
			UpdateTask(w, r)
		case http.MethodDelete:
			DeleteTask(w, r)
		default:
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			log.Println("ERR: Method not allowed")
		}
	}
}

func PostTask(w http.ResponseWriter, r *http.Request) {
	var task Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		log.Println("ERR: invalid request payload:", err)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error": "Field 'title' is required"}`, http.StatusBadRequest)
		log.Println("ERR: field 'title' is required:", err)
		return
	}

	now := time.Now()
	var taskDate time.Time

	//log.Println("ADDTASK: now =", now)
	//log.Println("ADDTASK: task.Date =", task.Date)
	//log.Println("ADDTASK: task.Title =", task.Title)
	//log.Println("ADDTASK: task.Comment =", task.Comment)
	//log.Println("ADDTASK: task.Repeat =", task.Repeat)

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
		taskDate = now
	} else if task.Date == now.Format(DateFormat) {
		taskDate = now
	} else {
		taskDate, err = time.Parse(DateFormat, task.Date)
		if err != nil {
			http.Error(w, `{"error": "Invalid date format"}`, http.StatusBadRequest)
			log.Println("ERR: invalid date format:", err)
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
				log.Println("ERR: invalid next date:", err)
				return
			}
			task.Date = nextDate
		}
	}

	//log.Println("INSERTTASK: task.Date =", task.Date)
	//log.Println("INSERTTASK: task.Title =", task.Title)
	//log.Println("INSERTTASK: task.Comment =", task.Comment)
	//log.Println("INSERTTASK: task.Repeat =", task.Repeat)

	taskId, err := database.AddTask(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("ERR: invalid task Id:", err)
		return
	}

	response := map[string]string{"id": strconv.Itoa(taskId)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var rows *sql.Rows
	var err error

	search := r.FormValue("search")

	//log.Println("SEARCH: search =", search)

	if search != "" {
		_, err := time.Parse(webuiDateFormat, search)
		if err != nil {
			searchQuery := "%" + search + "%"
			rows, err = database.Db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?", searchQuery, searchQuery, TasksLimit)
			if err != nil {
				http.Error(w, `{"error": "Database select error"}`, http.StatusInternalServerError)
				log.Println(err)
				return
			}
		} else {
			searchDate, _ := time.Parse(webuiDateFormat, search)
			searchDate2 := searchDate.Format(DateFormat)
			//log.Println("SEARCH: searchDate =", searchDate)
			//log.Println("SEARCH: searchDate2 =", searchDate2)
			rows, err = database.Db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT ?", searchDate2, TasksLimit)
			if err != nil {
				http.Error(w, `{"error": "Database select error"}`, http.StatusInternalServerError)
				log.Println(err)
				return
			}
		}
	} else {
		rows, err = database.Db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?", TasksLimit)
		if err != nil {
			http.Error(w, `{"error": "Database select error"}`, http.StatusInternalServerError)
			log.Println(err)
			return
		}
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
		http.Error(w, `{"error": "Rows scan error"}`, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []Task{}
	}

	response := map[string][]Task{"tasks": tasks}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "json encode error"}`, http.StatusInternalServerError)
	}
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, `{"error":"Task id not set"}`, http.StatusBadRequest)
		return
	}

	var task Task
	query := "SELECT * FROM scheduler WHERE id = ?"
	err := database.Db.QueryRow(query, id).Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
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
		http.Error(w, `{"error": "json encode error"}`, http.StatusInternalServerError)
	}
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
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
			log.Println("ERR: invalid date format:", err)
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
				log.Println("ERR: invalid next date:", err)
				return
			}
			task.Date = nextDate
		}
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := database.Db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
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
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func DoneTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, `{"error":"Task id not set"}`, http.StatusBadRequest)
		return
	}

	var task Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	err := database.Db.QueryRow(query, id).Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database select error"}`, http.StatusInternalServerError)
		}
		return
	}

	now := time.Now()
	//log.Println("BEGIN: now =", now)
	//log.Println("CONTI: task.Id =", task.Id)
	//log.Println("CONTI: task.Date =", task.Date)
	//log.Println("CONTI: task.Title =", task.Title)
	//log.Println("CONTI: task.Comment =", task.Comment)
	//log.Println("CONTI: task.Repeat =", task.Repeat)
	if task.Repeat != "" {
		nextDate, err := nextdate.NextDate(now, task.Date, task.Repeat)
		//log.Println("CONTI: nextDate =", nextDate)
		if err != nil {
			http.Error(w, `{"error":"NextDate error"}`, http.StatusInternalServerError)
			log.Println("ERR: invalid next date:", err)
			return
		}
		res, err := database.Db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, id)
		if err != nil {
			http.Error(w, `{"error":"Task update error"}`, http.StatusInternalServerError)
			log.Println("ERR: task update error:", err)
			return
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, `{"error":"Task update error"}`, http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
			return
		}
	} else {
		_, err := database.Db.Exec("DELETE FROM scheduler WHERE id = ?", id)
		//log.Println("CONTI: delete task id =", id)
		if err != nil {
			http.Error(w, `{"error":"Task delete error"}`, http.StatusInternalServerError)
			log.Println("ERR: task delete error:", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, `{"error":"Nask id not set"}`, http.StatusBadRequest)
		return
	}

	res, err := database.Db.Exec("DELETE FROM scheduler WHERE id = $1", id)
	if err != nil {
		http.Error(w, `{"error":"Task not found"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, `{"error": "Task delete error"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	todoPass := os.Getenv("TODO_PASSWORD")
	if todoPass == "" {
		http.Error(w, `{"error":"Password not set"}`, http.StatusInternalServerError)
		return
	}

	var userPass UserPassword
	err := json.NewDecoder(r.Body).Decode(&userPass)
	if err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if userPass.Password != todoPass {
		http.Error(w, `{"error": "Invalid password"}`, http.StatusUnauthorized)
		return
	}

	hash := sha256.New()
	hash.Write([]byte(todoPass))
	todoPassHash := hex.EncodeToString(hash.Sum(nil))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"control": todoPassHash,
	})

	tokenSigned, err := token.SignedString([]byte(tokenSecretKey))
	if err != nil {
		http.Error(w, `{"error": "Toker create error"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": tokenSigned,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		todoPass := os.Getenv("TODO_PASSWORD")
		if len(todoPass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}
			tokenString := cookie.Value

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(tokenSecretKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			controlHash, ok := claims["control"].(string)
			if !ok {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			hash := sha256.New()
			hash.Write([]byte(todoPass))
			todoPassHash := hex.EncodeToString(hash.Sum(nil))

			//log.Println("claims =", claims)
			//log.Println("control =", controlHash)
			//log.Println("todoPassHash =", todoPassHash)

			if controlHash != todoPassHash {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
