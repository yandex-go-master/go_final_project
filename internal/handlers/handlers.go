package handlers

import (
	"net/http"
	"time"

	"github.com/yandex-go-master/go_final_project/internal/nextdate"
)

const DateFormat = "20060102"

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
