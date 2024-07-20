package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	//проверяем правила повторения на пустоту
	if repeat == "" {
		if date == "" {
			return now.Format("20060102"), nil
		}
		return "", nil
	}

	//Если date пустой, то используется сегодняшняя дата
	if date == "" {
		return now.Format("20060102"), nil
	}

	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	// вычисляем следующую дату на основе правила повторения
	switch repeat[0] {
	case 'd':
		if len(repeat) < 3 || repeat[1] != ' ' {
			return "", errors.New("Invalid 'd' rule format: " + repeat)
		}
		daysStr := repeat[2:]
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return "", err
		}
		if days < 1 || days > 400 {
			return "", errors.New("Invalid number of days: " + daysStr)
		}
		nextDate := parsedDate.AddDate(0, 0, days)
		if nextDate.Equal(now) {
			return now.Format("20060102"), nil
		}

		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}

		return nextDate.Format("20060102"), nil

	case 'y':
		nextDate := parsedDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil
	default:
		return "", errors.New("Unsupported repeat rule format: " + repeat)
	}
}

func handleNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'now' format: %s", err), http.StatusBadRequest)
		return
	}
	_, err = time.Parse("20060102", dateStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'date' format: %s", err), http.StatusBadRequest)
		return
	}
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calculating next date: %s", err), http.StatusBadRequest)
		return
	}
	fmt.Fprintln(w, nextDate)
}
